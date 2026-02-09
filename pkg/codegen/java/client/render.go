package client

import (
	"bytes"
	"fmt"
	"icomplie/common"
	"icomplie/internal/transfer"
	jtpl "icomplie/pkg/codegen/java/template"
	"os"
	"path/filepath"
	"strings"
)

// RenderClient generates Java client code files
func RenderClient(outputDir, packagePath, namespace string, services []*transfer.ServiceDefine) error {
	clientDir := filepath.Join(outputDir, "client")
	if err := os.MkdirAll(clientDir, 0755); err != nil {
		return fmt.Errorf("failed to create client directory: %w", err)
	}

	reg := jtpl.GetDefaultRegistry()

	// 1. Generate HttpClient interface
	httpClientData := &HTTPClientRender{
		PackagePath: packagePath,
		Namespace:   namespace,
	}
	if err := renderTemplateToFile(reg, jtpl.TplHTTPClient, httpClientData, filepath.Join(clientDir, "HttpClient.java")); err != nil {
		return err
	}

	// 2. Generate ApacheHttpClient implementation
	if err := renderTemplateToFile(reg, jtpl.TplApacheHTTPClient, httpClientData, filepath.Join(clientDir, "ApacheHttpClient.java")); err != nil {
		return err
	}

	// 3. Generate service clients
	for _, svc := range services {
		clientData := convertServiceToClientRender(packagePath, namespace, svc)
		fileName := common.Capitalize(svc.Name) + "Client.java"
		if err := renderTemplateToFile(reg, jtpl.TplServiceClient, clientData, filepath.Join(clientDir, fileName)); err != nil {
			return err
		}
	}

	return nil
}

func convertServiceToClientRender(packagePath, namespace string, svc *transfer.ServiceDefine) *ServiceClientRender {
	render := &ServiceClientRender{
		PackagePath: packagePath,
		Namespace:   namespace,
		ServiceName: common.Capitalize(svc.Name),
		Methods:     make([]*MethodRender, 0),
	}

	// Process methods in order
	for _, method := range svc.Methods {
		var mr *MethodRender
		switch method.HTTPMethod {
		case transfer.HTTPMethodPost:
			mr = convertPostMethodFromMethod(svc.RootUrl, method)
		case transfer.HTTPMethodGet:
			mr = convertGetMethodFromMethod(svc.RootUrl, method)
		case transfer.HTTPMethodPut:
			mr = convertPutMethodFromMethod(svc.RootUrl, method)
		}
		if mr != nil {
			render.Methods = append(render.Methods, mr)
		}
	}

	return render
}

func convertPostMethodFromMethod(rootUrl string, m *transfer.Method) *MethodRender {
	returnType := getJavaReturnType(m.MethodReturnType)
	resultClass := getJavaResultClass(m.MethodReturnType)

	mr := &MethodRender{
		MethodName:  common.FormatVariable(m.Name, false),
		Description: m.Description,
		HTTPMethod:  "POST",
		FullPath:    joinURL(rootUrl, m.Url),
		ReturnType:  returnType,
		ResultClass: resultClass,
	}

	if m.PostParams == nil || m.PostParams.IsEmpty {
		mr.ParamsSignature = ""
		mr.BodyParam = "null"
	} else {
		paramType := m.PostParams.StructName
		paramName := common.FormatVariable(m.PostParams.ParamName, false)
		if paramName == "" {
			paramName = "request"
		}
		mr.ParamsSignature = paramType + " " + paramName
		mr.BodyParam = paramName
	}

	return mr
}

func convertGetMethodFromMethod(rootUrl string, m *transfer.Method) *MethodRender {
	returnType := getJavaReturnType(m.MethodReturnType)
	resultClass := getJavaResultClass(m.MethodReturnType)

	mr := &MethodRender{
		MethodName:  common.FormatVariable(m.Name, false),
		Description: m.Description,
		HTTPMethod:  "GET",
		FullPath:    joinURL(rootUrl, m.Url),
		ReturnType:  returnType,
		ResultClass: resultClass,
	}

	if m.GetParams == nil || m.GetParams.IsEmpty {
		mr.ParamsSignature = ""
		mr.HasQueryParams = false
	} else if len(m.GetParams.BasicParams) > 0 {
		// Basic params as query parameters (preferred over struct for client)
		var params []string
		var queryParams []*QueryParam
		for _, bp := range m.GetParams.BasicParams {
			javaType := getBasicJavaType(bp.TypeName)
			if bp.IsList {
				javaType = fmt.Sprintf("List<%s>", boxJavaType(javaType))
			}
			paramName := common.FormatVariable(bp.ParamName, false)
			params = append(params, javaType+" "+paramName)
			queryParams = append(queryParams, &QueryParam{
				Name:    bp.ParamName,
				VarName: paramName,
			})
		}
		mr.ParamsSignature = strings.Join(params, ", ")
		mr.HasQueryParams = len(queryParams) > 0
		mr.QueryParams = queryParams
	} else if m.GetParams.IsSingleStruct {
		// For struct params, pass as single object
		paramType := m.GetParams.StructName
		paramName := common.FormatVariable(m.GetParams.StructParamName, false)
		if paramName == "" {
			paramName = "request"
		}
		mr.ParamsSignature = paramType + " " + paramName
		mr.HasQueryParams = false
	}

	return mr
}

func convertPutMethodFromMethod(rootUrl string, m *transfer.Method) *MethodRender {
	returnType := getJavaReturnType(m.MethodReturnType)
	resultClass := getJavaResultClass(m.MethodReturnType)

	mr := &MethodRender{
		MethodName:  common.FormatVariable(m.Name, false),
		Description: m.Description,
		HTTPMethod:  "PUT",
		FullPath:    joinURL(rootUrl, m.Url),
		ReturnType:  returnType,
		ResultClass: resultClass,
	}

	if m.PostParams == nil || m.PostParams.IsEmpty {
		mr.ParamsSignature = ""
		mr.BodyParam = "null"
	} else {
		paramType := m.PostParams.StructName
		paramName := common.FormatVariable(m.PostParams.ParamName, false)
		if paramName == "" {
			paramName = "request"
		}
		mr.ParamsSignature = paramType + " " + paramName
		mr.BodyParam = paramName
	}

	return mr
}

func getJavaReturnType(rt transfer.MethodReturnType) string {
	if rt.IsVoid {
		return "Void"
	}

	var baseType string
	if rt.IsStruct {
		baseType = rt.TypeName
	} else {
		baseType = getBasicJavaType(rt.TypeName)
		baseType = boxJavaType(baseType)
	}

	if rt.IsList {
		return fmt.Sprintf("List<%s>", baseType)
	}

	return baseType
}

func getJavaResultClass(rt transfer.MethodReturnType) string {
	if rt.IsVoid {
		return "Void"
	}

	if rt.IsStruct {
		if rt.IsList {
			return "List" // Will need type token for generics
		}
		return rt.TypeName
	}

	baseType := getBasicJavaType(rt.TypeName)
	baseType = boxJavaType(baseType)

	if rt.IsList {
		return "List"
	}

	return baseType
}

func getBasicJavaType(typeName string) string {
	switch typeName {
	case "bool":
		return "boolean"
	case "byte", "i8":
		return "byte"
	case "i16":
		return "short"
	case "i32":
		return "int"
	case "i64":
		return "long"
	case "float":
		return "float"
	case "double":
		return "double"
	case "string":
		return "String"
	default:
		return "Object"
	}
}

func boxJavaType(javaType string) string {
	switch javaType {
	case "boolean":
		return "Boolean"
	case "byte":
		return "Byte"
	case "short":
		return "Short"
	case "int":
		return "Integer"
	case "long":
		return "Long"
	case "float":
		return "Float"
	case "double":
		return "Double"
	default:
		return javaType
	}
}

func renderTemplateToFile(reg *jtpl.Registry, templateName string, data any, outputPath string) error {
	tmpl, err := reg.Get(templateName)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template %s: %w", templateName, err)
	}

	return os.WriteFile(outputPath, buf.Bytes(), 0644)
}

// joinURL joins root URL and path with proper slash handling
func joinURL(rootUrl, path string) string {
	if rootUrl == "" {
		return "/" + strings.TrimPrefix(path, "/")
	}
	rootUrl = strings.TrimSuffix(rootUrl, "/")
	path = strings.TrimPrefix(path, "/")
	return rootUrl + "/" + path
}
