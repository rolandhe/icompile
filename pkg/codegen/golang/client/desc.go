package client

import (
	"errors"
	"fmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"icomplie/internal/transfer"
	"net/url"
	"strings"
)

// Type mapping from IDL types to Go types
var typeMapping = map[string]string{
	"i8":     "int8",
	"i16":    "int16",
	"i32":    "int32",
	"i64":    "int64",
	"string": "string",
	"double": "float64",
	"float":  "float32",
	"byte":   "byte",
	"bool":   "bool",
}

// Convert expression for basic types
var convertExprMapping = map[string]string{
	"int8":    "strconv.FormatInt(int64(%s), 10)",
	"int16":   "strconv.FormatInt(int64(%s), 10)",
	"int32":   "strconv.FormatInt(int64(%s), 10)",
	"int64":   "strconv.FormatInt(%s, 10)",
	"float32": "strconv.FormatFloat(float64(%s), 'f', -1, 32)",
	"float64": "strconv.FormatFloat(%s, 'f', -1, 64)",
	"byte":    "strconv.FormatInt(int64(%s), 10)",
	"bool":    "strconv.FormatBool(%s)",
	"string":  "%s",
}

// ServiceClientRender holds data for rendering service client
type ServiceClientRender struct {
	Namespace   string
	ServiceName string
	BasePath    string
	Imports     []string
	Methods     []*MethodRender
}

// MethodRender holds data for rendering a single method
type MethodRender struct {
	Name            string
	Description     string
	HTTPMethod      string
	FullPath        string
	ParamsSignature string // e.g., ", req *OrderRequest" or ", orderId int64"
	ReturnType      string // e.g., "*Result[int64]"
	ResultType      string // e.g., "Result[int64]" (without pointer)
	BodyParam       string // e.g., "req" for POST body
	HasParams       bool
	QueryParams     []*QueryParam
}

// QueryParam holds data for a query parameter
type QueryParam struct {
	ParamName   string
	GoType      string
	IsList      bool
	ConvertExpr string // e.g., "strconv.FormatInt(orderId, 10)"
}

// Convert converts a Definition to ServiceClientRender
func Convert(def *transfer.Definition) (*ServiceClientRender, error) {
	if len(def.Services) == 0 {
		return nil, errors.New("no service defined")
	}
	svc := def.Services[0]

	render := &ServiceClientRender{
		Namespace:   def.Namespace,
		ServiceName: toPublicName(svc.Name),
		BasePath:    svc.RootUrl,
		Imports:     []string{},
		Methods:     []*MethodRender{},
	}

	// Collect imports
	importSet := make(map[string]struct{})

	// Process methods in IDL order
	for _, method := range svc.Methods {
		var methodRender *MethodRender
		var err error

		switch method.HTTPMethod {
		case transfer.HTTPMethodPost:
			methodRender, err = convertPostMethod(svc.RootUrl, method, importSet)
		case transfer.HTTPMethodGet:
			methodRender, err = convertGetMethod(svc.RootUrl, method, importSet)
		case transfer.HTTPMethodPut:
			methodRender, err = convertPutMethod(svc.RootUrl, method, importSet)
		default:
			continue
		}

		if err != nil {
			return nil, err
		}
		render.Methods = append(render.Methods, methodRender)
	}

	// Build imports list
	for imp := range importSet {
		render.Imports = append(render.Imports, imp)
	}

	return render, nil
}

func convertPostMethod(rootUrl string, method *transfer.Method, imports map[string]struct{}) (*MethodRender, error) {
	if method.PostParams == nil || method.PostParams.IsEmpty {
		return nil, fmt.Errorf("POST method %s must have parameters", method.Name)
	}

	fullPath, err := url.JoinPath(rootUrl, method.Url)
	if err != nil {
		return nil, err
	}

	returnType, resultType, err := toReturnTypes(&method.MethodReturnType, imports)
	if err != nil {
		return nil, err
	}

	paramType := method.PostParams.StructName
	if strings.Contains(paramType, ".") {
		// External type, add import
		parts := strings.Split(paramType, ".")
		imports[fmt.Sprintf(`"%s"`, parts[0])] = struct{}{}
	}

	return &MethodRender{
		Name:            toPublicName(method.Name),
		Description:     method.Description,
		HTTPMethod:      "POST",
		FullPath:        fullPath,
		ParamsSignature: fmt.Sprintf(", %s *%s", method.PostParams.ParamName, paramType),
		ReturnType:      returnType,
		ResultType:      resultType,
		BodyParam:       method.PostParams.ParamName,
		HasParams:       false,
		QueryParams:     nil,
	}, nil
}

func convertGetMethod(rootUrl string, method *transfer.Method, imports map[string]struct{}) (*MethodRender, error) {
	fullPath, err := url.JoinPath(rootUrl, method.Url)
	if err != nil {
		return nil, err
	}

	returnType, resultType, err := toReturnTypes(&method.MethodReturnType, imports)
	if err != nil {
		return nil, err
	}

	render := &MethodRender{
		Name:        toPublicName(method.Name),
		Description: method.Description,
		HTTPMethod:  "GET",
		FullPath:    fullPath,
		ReturnType:  returnType,
		ResultType:  resultType,
		HasParams:   false,
		QueryParams: []*QueryParam{},
	}

	if method.GetParams == nil || method.GetParams.IsEmpty {
		render.ParamsSignature = ""
		return render, nil
	}

	// Basic params: use per-field query params (preferred over struct for client)
	if len(method.GetParams.BasicParams) > 0 {
		var sigParts []string
		for _, bp := range method.GetParams.BasicParams {
			goType := typeMapping[bp.TypeName]
			if goType == "" {
				return nil, fmt.Errorf("unsupported type: %s", bp.TypeName)
			}

			typeStr := goType
			if bp.IsList {
				typeStr = "[]" + goType
			}
			sigParts = append(sigParts, fmt.Sprintf("%s %s", bp.ParamName, typeStr))

			convertExpr := convertExprMapping[goType]
			if bp.IsList {
				convertExpr = fmt.Sprintf(convertExpr, "v")
			} else {
				convertExpr = fmt.Sprintf(convertExpr, bp.ParamName)
			}

			render.QueryParams = append(render.QueryParams, &QueryParam{
				ParamName:   bp.ParamName,
				GoType:      goType,
				IsList:      bp.IsList,
				ConvertExpr: convertExpr,
			})
		}

		if len(sigParts) > 0 {
			render.ParamsSignature = ", " + strings.Join(sigParts, ", ")
			render.HasParams = true
		}

		return render, nil
	}

	if method.GetParams.IsSingleStruct {
		paramType := method.GetParams.StructName
		if strings.Contains(paramType, ".") {
			parts := strings.Split(paramType, ".")
			imports[fmt.Sprintf(`"%s"`, parts[0])] = struct{}{}
		}
		render.ParamsSignature = fmt.Sprintf(", %s *%s", method.GetParams.StructParamName, paramType)
		render.HasParams = true
		render.QueryParams = []*QueryParam{{
			ParamName:   method.GetParams.StructParamName,
			IsList:      false,
			ConvertExpr: "",
		}}
		return render, nil
	}

	return render, nil
}

func convertPutMethod(rootUrl string, method *transfer.Method, imports map[string]struct{}) (*MethodRender, error) {
	if method.PostParams == nil || method.PostParams.IsEmpty {
		return nil, fmt.Errorf("PUT method %s must have parameters", method.Name)
	}

	fullPath, err := url.JoinPath(rootUrl, method.Url)
	if err != nil {
		return nil, err
	}

	returnType, resultType, err := toReturnTypes(&method.MethodReturnType, imports)
	if err != nil {
		return nil, err
	}

	paramType := method.PostParams.StructName
	if strings.Contains(paramType, ".") {
		parts := strings.Split(paramType, ".")
		imports[fmt.Sprintf(`"%s"`, parts[0])] = struct{}{}
	}

	return &MethodRender{
		Name:            toPublicName(method.Name),
		Description:     method.Description,
		HTTPMethod:      "PUT",
		FullPath:        fullPath,
		ParamsSignature: fmt.Sprintf(", %s *%s", method.PostParams.ParamName, paramType),
		ReturnType:      returnType,
		ResultType:      resultType,
		BodyParam:       method.PostParams.ParamName,
		HasParams:       false,
		QueryParams:     nil,
	}, nil
}

func toReturnTypes(rt *transfer.MethodReturnType, imports map[string]struct{}) (string, string, error) {
	if rt.IsVoid {
		return "*Result[any]", "Result[any]", nil
	}

	var innerType string
	if rt.IsStruct {
		if strings.Contains(rt.TypeName, ".") {
			parts := strings.Split(rt.TypeName, ".")
			imports[fmt.Sprintf(`"%s"`, parts[0])] = struct{}{}
		}

		if rt.IsPager {
			innerType = fmt.Sprintf("*PageList[%s]", rt.TypeName)
		} else if rt.IsList {
			innerType = fmt.Sprintf("[]*%s", rt.TypeName)
		} else {
			innerType = "*" + rt.TypeName
		}
	} else {
		goType := typeMapping[rt.TypeName]
		if goType == "" {
			return "", "", fmt.Errorf("unsupported return type: %s", rt.TypeName)
		}
		if rt.IsList {
			innerType = "[]" + goType
		} else {
			innerType = goType
		}
	}

	return fmt.Sprintf("*Result[%s]", innerType), fmt.Sprintf("Result[%s]", innerType), nil
}

func toPublicName(name string) string {
	return cases.Title(language.AmericanEnglish, cases.NoLower).String(name)
}
