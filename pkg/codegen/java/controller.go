package java

import (
	"bytes"
	"fmt"
	"icomplie/common"
	"icomplie/internal/transfer"
	"icomplie/pkg/codegen/java/client"
	jtpl "icomplie/pkg/codegen/java/template"
	"os"
	"path/filepath"
)

// generateController generates a Spring MVC Controller for a service
func generateController(outputDir, packagePath, namespace string, svc *transfer.ServiceDefine) (string, error) {
	className := common.Capitalize(svc.Name) + "Controller"
	fileName := filepath.Join(outputDir, className+".java")

	reg := jtpl.GetDefaultRegistry()

	render := &client.ControllerRender{
		PackagePath:     packagePath,
		Namespace:       namespace,
		ClassName:       className,
		RootUrl:         svc.RootUrl,
		HasRequestBody:  hasRequestBody(svc),
		HasRequestParam: hasRequestParam(svc),
		HasList:         hasListReturnType(svc),
		Methods:         make([]*client.ControllerMethodRender, 0),
	}

	// Generate methods in order
	for _, method := range svc.Methods {
		var methodContent string
		var err error

		switch method.HTTPMethod {
		case transfer.HTTPMethodPost:
			methodContent, err = generatePostMethodContentFromMethod(reg, method)
		case transfer.HTTPMethodGet:
			methodContent, err = generateGetMethodContentFromMethod(reg, method)
		case transfer.HTTPMethodPut:
			methodContent, err = generatePutMethodContentFromMethod(reg, method)
		}

		if err != nil {
			return "", err
		}

		render.Methods = append(render.Methods, &client.ControllerMethodRender{
			Content: methodContent,
		})
	}

	tmpl, err := reg.Get(jtpl.TplController)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, render); err != nil {
		return "", fmt.Errorf("failed to execute controller template: %w", err)
	}

	if err := os.WriteFile(fileName, buf.Bytes(), 0644); err != nil {
		return "", fmt.Errorf("failed to write controller file %s: %w", fileName, err)
	}

	return fileName, nil
}

func generatePostMethodContentFromMethod(reg *jtpl.Registry, m *transfer.Method) (string, error) {
	methodName := common.FormatVariable(m.Name, false)
	returnType := getMethodReturnType(m.MethodReturnType)

	var paramsSignature string
	if m.PostParams == nil || m.PostParams.IsEmpty {
		paramsSignature = ""
	} else {
		paramType := m.PostParams.StructName
		paramName := common.FormatVariable(m.PostParams.ParamName, false)
		if paramName == "" {
			paramName = "request"
		}
		paramsSignature = fmt.Sprintf("@RequestBody %s %s", paramType, paramName)
	}

	data := &client.ControllerMethodData{
		HTTPMethod:      "POST",
		Url:             m.Url,
		ReturnType:      returnType,
		MethodName:      methodName,
		ParamsSignature: paramsSignature,
	}

	return renderMethodTemplate(reg, data)
}

func generateGetMethodContentFromMethod(reg *jtpl.Registry, m *transfer.Method) (string, error) {
	methodName := common.FormatVariable(m.Name, false)
	returnType := getMethodReturnType(m.MethodReturnType)

	var paramsSignature string
	if m.GetParams == nil || m.GetParams.IsEmpty {
		paramsSignature = ""
	} else if m.GetParams.IsSingleStruct {
		paramType := m.GetParams.StructName
		paramName := common.FormatVariable(m.GetParams.StructParamName, false)
		if paramName == "" {
			paramName = "request"
		}
		paramsSignature = fmt.Sprintf("%s %s", paramType, paramName)
	}

	data := &client.ControllerMethodData{
		HTTPMethod:      "GET",
		Url:             m.Url,
		ReturnType:      returnType,
		MethodName:      methodName,
		ParamsSignature: paramsSignature,
	}

	return renderMethodTemplate(reg, data)
}

func generatePutMethodContentFromMethod(reg *jtpl.Registry, m *transfer.Method) (string, error) {
	methodName := common.FormatVariable(m.Name, false)
	returnType := getMethodReturnType(m.MethodReturnType)

	var paramsSignature string
	if m.PostParams == nil || m.PostParams.IsEmpty {
		paramsSignature = ""
	} else {
		paramType := m.PostParams.StructName
		paramName := common.FormatVariable(m.PostParams.ParamName, false)
		if paramName == "" {
			paramName = "request"
		}
		paramsSignature = fmt.Sprintf("@RequestBody %s %s", paramType, paramName)
	}

	data := &client.ControllerMethodData{
		HTTPMethod:      "PUT",
		Url:             m.Url,
		ReturnType:      returnType,
		MethodName:      methodName,
		ParamsSignature: paramsSignature,
	}

	return renderMethodTemplate(reg, data)
}

func renderMethodTemplate(reg *jtpl.Registry, data *client.ControllerMethodData) (string, error) {
	tmpl, err := reg.Get(jtpl.TplControllerMethod)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute controller method template: %w", err)
	}

	return buf.String(), nil
}

// getMethodReturnType converts IDL return type to Java type
func getMethodReturnType(rt transfer.MethodReturnType) string {
	if rt.IsVoid {
		return "void"
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

// hasRequestBody checks if any method uses request body
func hasRequestBody(svc *transfer.ServiceDefine) bool {
	for _, m := range svc.Posts {
		if !m.Params.IsEmpty {
			return true
		}
	}
	for _, m := range svc.Puts {
		if !m.Params.IsEmpty {
			return true
		}
	}
	return false
}

// hasRequestParam checks if any GET method uses request params
func hasRequestParam(svc *transfer.ServiceDefine) bool {
	for _, m := range svc.Gets {
		if !m.Params.IsEmpty && !m.Params.IsSingleStruct {
			return true
		}
	}
	return false
}

// hasListReturnType checks if any method returns a list
func hasListReturnType(svc *transfer.ServiceDefine) bool {
	for _, m := range svc.Posts {
		if m.MethodReturnType.IsList {
			return true
		}
	}
	for _, m := range svc.Gets {
		if m.MethodReturnType.IsList {
			return true
		}
	}
	for _, m := range svc.Puts {
		if m.MethodReturnType.IsList {
			return true
		}
	}
	return false
}
