package client

import (
	"bytes"
	"fmt"
	"icomplie/common"
	"icomplie/internal/transfer"
	tstpl "icomplie/pkg/codegen/typescript/template"
	"os"
	"path/filepath"
	"strings"
)

// Platform constants
const (
	PlatformBrowser = "browser"
	PlatformMiniApp = "miniapp"
)

// RenderClient generates TypeScript client code files
func RenderClient(outputDir, namespace string, def *transfer.Definition, platform string) error {
	clientDir := filepath.Join(outputDir, namespace)
	if err := os.MkdirAll(clientDir, 0755); err != nil {
		return fmt.Errorf("failed to create client directory: %w", err)
	}

	reg := tstpl.GetDefaultRegistry()

	// Select templates based on platform
	var tplTypes, tplHTTPClient, tplImplClient, tplServiceClient string
	var implClientFileName string

	if platform == PlatformMiniApp {
		tplTypes = tstpl.TplMiniAppTypes
		tplHTTPClient = tstpl.TplMiniAppHTTPClient
		tplImplClient = tstpl.TplMiniAppWxClient
		tplServiceClient = tstpl.TplMiniAppServiceClient
		implClientFileName = "wxClient.ts"
	} else {
		// Default to browser
		tplTypes = tstpl.TplBrowserTypes
		tplHTTPClient = tstpl.TplBrowserHTTPClient
		tplImplClient = tstpl.TplBrowserAxiosClient
		tplServiceClient = tstpl.TplBrowserServiceClient
		implClientFileName = "axiosClient.ts"
	}

	// 1. Generate types.ts
	typesData := convertStructsToTypesRender(def.Structs)
	if err := renderTemplateToFile(reg, tplTypes, typesData, filepath.Join(clientDir, "types.ts")); err != nil {
		return err
	}

	// 2. Generate httpClient.ts (interface)
	if err := renderTemplateToFile(reg, tplHTTPClient, &EmptyRender{}, filepath.Join(clientDir, "httpClient.ts")); err != nil {
		return err
	}

	// 3. Generate implementation (axiosClient.ts or wxClient.ts)
	if err := renderTemplateToFile(reg, tplImplClient, &EmptyRender{}, filepath.Join(clientDir, implClientFileName)); err != nil {
		return err
	}

	// 4. Generate service clients
	for _, svc := range def.Services {
		clientData := convertServiceToClientRender(svc)
		fileName := common.FormatVariable(svc.Name, false) + "Client.ts"
		if err := renderTemplateToFile(reg, tplServiceClient, clientData, filepath.Join(clientDir, fileName)); err != nil {
			return err
		}
	}

	return nil
}

func convertStructsToTypesRender(structs []*transfer.StructDefine) *TypesRender {
	render := &TypesRender{
		Structs: make([]*StructRender, 0, len(structs)),
	}

	for _, st := range structs {
		sr := &StructRender{
			Name:    st.Name,
			Extends: st.Extends,
			Fields:  make([]*FieldRender, 0, len(st.Fields)),
		}

		for _, field := range st.Fields {
			tsType := getTypeScriptType(field.Tp)
			optional := field.ReqDefine != "required"
			sr.Fields = append(sr.Fields, &FieldRender{
				Name:     common.FormatVariable(field.Name, false),
				Type:     tsType,
				Optional: optional,
			})
		}

		render.Structs = append(render.Structs, sr)
	}

	return render
}

func convertServiceToClientRender(svc *transfer.ServiceDefine) *ServiceClientRender {
	render := &ServiceClientRender{
		ServiceName: common.Capitalize(svc.Name),
		Methods:     make([]*MethodRender, 0),
	}

	// Collect import types
	importTypes := make(map[string]struct{})

	// Process methods in order
	for _, method := range svc.Methods {
		var mr *MethodRender
		switch method.HTTPMethod {
		case transfer.HTTPMethodPost:
			mr = convertPostMethodFromMethod(svc.RootUrl, method, importTypes)
		case transfer.HTTPMethodGet:
			mr = convertGetMethodFromMethod(svc.RootUrl, method, importTypes)
		case transfer.HTTPMethodPut:
			mr = convertPutMethodFromMethod(svc.RootUrl, method, importTypes)
		}
		if mr != nil {
			render.Methods = append(render.Methods, mr)
		}
	}

	// Build import types string
	var types []string
	for t := range importTypes {
		types = append(types, t)
	}
	render.ImportTypes = strings.Join(types, ", ")

	return render
}

func convertPostMethodFromMethod(rootUrl string, m *transfer.Method, importTypes map[string]struct{}) *MethodRender {
	returnType := getTSReturnType(m.MethodReturnType, importTypes)

	mr := &MethodRender{
		MethodName:  common.FormatVariable(m.Name, false),
		Description: m.Description,
		HTTPMethod:  "POST",
		FullPath:    joinURL(rootUrl, m.Url),
		ReturnType:  returnType,
	}

	if m.PostParams == nil || m.PostParams.IsEmpty {
		mr.ParamsSignature = ""
		mr.BodyParam = "undefined"
	} else {
		paramType := getTSTypeName(m.PostParams.StructName, importTypes)
		paramName := common.FormatVariable(m.PostParams.ParamName, false)
		if paramName == "" {
			paramName = "req"
		}
		mr.ParamsSignature = paramName + ": " + paramType
		mr.BodyParam = paramName
	}

	return mr
}

func convertGetMethodFromMethod(rootUrl string, m *transfer.Method, importTypes map[string]struct{}) *MethodRender {
	returnType := getTSReturnType(m.MethodReturnType, importTypes)

	mr := &MethodRender{
		MethodName:  common.FormatVariable(m.Name, false),
		Description: m.Description,
		HTTPMethod:  "GET",
		FullPath:    joinURL(rootUrl, m.Url),
		ReturnType:  returnType,
	}

	if m.GetParams == nil || m.GetParams.IsEmpty {
		mr.ParamsSignature = ""
		mr.HasQueryParams = false
	} else if len(m.GetParams.BasicParams) > 0 {
		// Basic params as query parameters (preferred over struct for client)
		var params []string
		var queryParts []string
		for _, bp := range m.GetParams.BasicParams {
			tsType := getBasicTSType(bp.TypeName)
			if bp.IsList {
				tsType = tsType + "[]"
			}
			paramName := common.FormatVariable(bp.ParamName, false)
			params = append(params, paramName+": "+tsType)
			queryParts = append(queryParts, paramName)
		}
		mr.ParamsSignature = strings.Join(params, ", ")
		mr.HasQueryParams = len(queryParts) > 0
		mr.QueryParamsObject = strings.Join(queryParts, ", ")
	} else if m.GetParams.IsSingleStruct {
		paramType := getTSTypeName(m.GetParams.StructName, importTypes)
		paramName := common.FormatVariable(m.GetParams.StructParamName, false)
		if paramName == "" {
			paramName = "req"
		}
		mr.ParamsSignature = paramName + ": " + paramType
		mr.HasQueryParams = false
	}

	return mr
}

func convertPutMethodFromMethod(rootUrl string, m *transfer.Method, importTypes map[string]struct{}) *MethodRender {
	returnType := getTSReturnType(m.MethodReturnType, importTypes)

	mr := &MethodRender{
		MethodName:  common.FormatVariable(m.Name, false),
		Description: m.Description,
		HTTPMethod:  "PUT",
		FullPath:    joinURL(rootUrl, m.Url),
		ReturnType:  returnType,
	}

	if m.PostParams == nil || m.PostParams.IsEmpty {
		mr.ParamsSignature = ""
		mr.BodyParam = "undefined"
	} else {
		paramType := getTSTypeName(m.PostParams.StructName, importTypes)
		paramName := common.FormatVariable(m.PostParams.ParamName, false)
		if paramName == "" {
			paramName = "req"
		}
		mr.ParamsSignature = paramName + ": " + paramType
		mr.BodyParam = paramName
	}

	return mr
}

func getTypeScriptType(ft *transfer.FieldType) string {
	if ft == nil {
		return "any"
	}

	if ft.IsBasic {
		return getBasicTSType(ft.TypeName)
	}

	if ft.IsStruct {
		return ft.TypeName
	}

	if ft.IsList {
		if ft.ValueType == nil {
			return "any[]"
		}
		innerType := getTypeScriptType(ft.ValueType)
		return innerType + "[]"
	}

	if ft.IsMap {
		keyType := getBasicTSType(ft.TypeName)
		valueType := "any"
		if ft.ValueType != nil {
			valueType = getTypeScriptType(ft.ValueType)
		}
		return fmt.Sprintf("Record<%s, %s>", keyType, valueType)
	}

	return "any"
}

func getBasicTSType(typeName string) string {
	switch typeName {
	case "bool":
		return "boolean"
	case "byte", "i8", "i16", "i32", "i64", "float", "double":
		return "number"
	case "string":
		return "string"
	default:
		return "any"
	}
}

func getTSReturnType(rt transfer.MethodReturnType, importTypes map[string]struct{}) string {
	if rt.IsVoid {
		return "void"
	}

	var baseType string
	if rt.IsStruct {
		baseType = rt.TypeName
		// Add to import types (only local types, not external)
		if !strings.Contains(rt.TypeName, ".") {
			importTypes[rt.TypeName] = struct{}{}
		}
	} else {
		baseType = getBasicTSType(rt.TypeName)
	}

	if rt.IsList {
		return baseType + "[]"
	}

	return baseType
}

func getTSTypeName(structName string, importTypes map[string]struct{}) string {
	// Handle external types (e.g., order_share.AgentOrder)
	if strings.Contains(structName, ".") {
		// For now, just use the type name as-is
		// In a more complete implementation, we'd handle imports properly
		return structName
	}

	// Add to import types
	importTypes[structName] = struct{}{}
	return structName
}

func joinURL(rootUrl, path string) string {
	if rootUrl == "" {
		return "/" + strings.TrimPrefix(path, "/")
	}
	rootUrl = strings.TrimSuffix(rootUrl, "/")
	path = strings.TrimPrefix(path, "/")
	return rootUrl + "/" + path
}

func renderTemplateToFile(reg *tstpl.Registry, templateName string, data any, outputPath string) error {
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
