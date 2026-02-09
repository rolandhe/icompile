package golang

import (
	"fmt"
	"icomplie/common"
	"icomplie/internal/transfer"
	gtpl "icomplie/pkg/codegen/golang/template"
	"os"
	"path/filepath"
	"strings"
)

func genServer(def *transfer.Definition, rootDir, serviceIdlFileName, pp string, onlyStruct bool) error {
	if err := generateGoStructFile(def, rootDir, serviceIdlFileName, onlyStruct); err != nil {
		return err
	}
	if onlyStruct {
		return nil
	}
	return generateGoServiceFile(def, rootDir, serviceIdlFileName, pp+"/"+def.Namespace)
}

// backupFile 备份文件
// 如果 .bak 存在，使用 .bak.1, .bak.2 递增
func backupFile(filePath string) error {
	if !common.IsExist(filePath) {
		return nil
	}

	bakPath := filePath + ".bak"
	if !common.IsExist(bakPath) {
		return os.Rename(filePath, bakPath)
	}

	// 找到可用的备份编号
	for i := 1; ; i++ {
		numberedBak := fmt.Sprintf("%s.bak.%d", filePath, i)
		if !common.IsExist(numberedBak) {
			return os.Rename(filePath, numberedBak)
		}
	}
}

func generateGoServiceFile(def *transfer.Definition, dirPath, serviceName, pp string) error {
	goFileName := fmt.Sprintf("%s%s", serviceName, outputGoFileInterfaceSuffix)
	goAchieveFileName := fmt.Sprintf("%s%s", serviceName, outputGoFileImplSuffix)
	if len(def.Services) == 0 {
		return nil
	}

	// 删除 controller 文件（如果存在）
	controllerPath := filepath.Join(dirPath, goFileName)
	if common.IsExist(controllerPath) {
		os.Remove(controllerPath)
	}

	// 备份 controller_impl 文件（如果存在）
	implPath := filepath.Join(dirPath, goAchieveFileName)
	if err := backupFile(implPath); err != nil {
		return fmt.Errorf("failed to backup impl file: %w", err)
	}

	services := def.Services
	namespace := def.Namespace
	reg := getRegistry()

	var result string
	var resultAchieve string

	collectImportMap := make(map[string]struct{})

	for _, service := range services {
		interfaceResult, funcCallResult, funcs, err := parseInterface(service, collectImportMap, reg)
		if err != nil {
			return err
		}

		// Render bind function
		bindFuncData := &BindFunctionData{
			ServiceName:      service.Name,
			ServiceNameLower: common.FormatVariable(service.Name, false),
			RootUrl:          service.RootUrl,
			Handlers:         strings.Split(funcCallResult, "\n"),
		}
		bindFunc, err := renderBindFunction(reg, bindFuncData)
		if err != nil {
			return err
		}
		result += bindFunc + "\n"

		// Render impl struct
		implStructData := &ImplStructData{
			Name: common.FormatVariable(service.Name, false) + controllerName + stdServiceImplSuffix,
		}
		implStruct, err := renderImplStruct(reg, implStructData)
		if err != nil {
			return err
		}
		result += implStruct + "\n\n"

		result += interfaceResult + "\n\n"

		// Generate impl methods
		implResult, err := generateGoServiceImplFile(funcs, service.Name, reg)
		if err != nil {
			return err
		}
		resultAchieve += implResult
	}

	// Build extra imports
	extraImports, err := buildExtraImports(collectImportMap, def)
	if err != nil {
		return err
	}

	// Render controller header
	controllerHeaderData := &ControllerHeaderData{
		PackageName:  namespace,
		PkgPath:      pp,
		ExtraImports: extraImports,
	}
	controllerHeader, err := renderControllerHeader(reg, controllerHeaderData)
	if err != nil {
		return err
	}

	// Render impl header
	implHeaderData := &ImplHeaderData{
		PackageName:  namespace,
		PkgPath:      pp,
		ExtraImports: extraImports,
	}
	implHeader, err := renderImplHeader(reg, implHeaderData)
	if err != nil {
		return err
	}

	result = controllerHeader + result
	resultAchieve = implHeader + resultAchieve

	if err := save(dirPath, goFileName, result); err != nil {
		return err
	}

	if err := save(dirPath, goAchieveFileName, resultAchieve); err != nil {
		return err
	}
	return nil
}

func buildExtraImports(collectImportMap map[string]struct{}, def *transfer.Definition) ([]string, error) {
	var imports []string
	for k := range collectImportMap {
		impDefine := def.GoStructImports[k]
		if len(impDefine) == 0 {
			return nil, fmt.Errorf("type '%s' has not been imported", k)
		}
		if impDefine[0].NeedAlias() {
			imports = append(imports, fmt.Sprintf("%s \"%s\"", impDefine[0].Alias, impDefine[0].PackageName))
		} else {
			imports = append(imports, fmt.Sprintf("\"%s\"", impDefine[0].PackageName))
		}
	}
	return imports, nil
}

func renderControllerHeader(reg *gtpl.Registry, data *ControllerHeaderData) (string, error) {
	tmpl, err := reg.Get(gtpl.TplControllerHeader)
	if err != nil {
		return "", fmt.Errorf("failed to get controller header template: %w", err)
	}
	var builder strings.Builder
	if err := tmpl.Execute(&builder, data); err != nil {
		return "", fmt.Errorf("failed to execute controller header template: %w", err)
	}
	return builder.String(), nil
}

func renderImplHeader(reg *gtpl.Registry, data *ImplHeaderData) (string, error) {
	tmpl, err := reg.Get(gtpl.TplImplHeader)
	if err != nil {
		return "", fmt.Errorf("failed to get impl header template: %w", err)
	}
	var builder strings.Builder
	if err := tmpl.Execute(&builder, data); err != nil {
		return "", fmt.Errorf("failed to execute impl header template: %w", err)
	}
	return builder.String(), nil
}

func renderBindFunction(reg *gtpl.Registry, data *BindFunctionData) (string, error) {
	tmpl, err := reg.Get(gtpl.TplBindFunction)
	if err != nil {
		return "", fmt.Errorf("failed to get bind function template: %w", err)
	}
	var builder strings.Builder
	if err := tmpl.Execute(&builder, data); err != nil {
		return "", fmt.Errorf("failed to execute bind function template: %w", err)
	}
	return builder.String(), nil
}

func renderImplStruct(reg *gtpl.Registry, data *ImplStructData) (string, error) {
	tmpl, err := reg.Get(gtpl.TplImplStruct)
	if err != nil {
		return "", fmt.Errorf("failed to get impl struct template: %w", err)
	}
	var builder strings.Builder
	if err := tmpl.Execute(&builder, data); err != nil {
		return "", fmt.Errorf("failed to execute impl struct template: %w", err)
	}
	return builder.String(), nil
}

func parseInterface(service *transfer.ServiceDefine, collectMap map[string]struct{}, reg *gtpl.Registry) (string, string, []string, error) {
	var funcs []string
	var funcCallResult strings.Builder

	// 按 IDL 定义顺序遍历所有方法
	for _, method := range service.Methods {
		switch method.HTTPMethod {
		case transfer.HTTPMethodPost:
			funcStr, handlerStr, err := parseSinglePostMethod(method, collectMap, reg)
			if err != nil {
				return "", "", nil, err
			}
			funcs = append(funcs, funcStr)
			funcCallResult.WriteString(handlerStr + "\n")

		case transfer.HTTPMethodGet:
			funcStr, handlerStr, err := parseSingleGetMethod(method, collectMap, reg)
			if err != nil {
				return "", "", nil, err
			}
			funcs = append(funcs, funcStr)
			funcCallResult.WriteString(handlerStr + "\n")

		case transfer.HTTPMethodPut:
			funcStr, handlerStr, err := parseSinglePutMethod(method, collectMap, reg)
			if err != nil {
				return "", "", nil, err
			}
			funcs = append(funcs, funcStr)
			funcCallResult.WriteString(handlerStr + "\n")
		}
	}

	// Render interface
	interfaceData := &InterfaceRenderData{
		Name:    service.Name + controllerName + stdServiceSuffix,
		Methods: funcs,
	}
	interfaceResult, err := renderInterface(reg, interfaceData)
	if err != nil {
		return "", "", nil, err
	}

	return interfaceResult, funcCallResult.String(), funcs, nil
}

func renderInterface(reg *gtpl.Registry, data *InterfaceRenderData) (string, error) {
	tmpl, err := reg.Get(gtpl.TplInterface)
	if err != nil {
		return "", fmt.Errorf("failed to get interface template: %w", err)
	}
	var builder strings.Builder
	if err := tmpl.Execute(&builder, data); err != nil {
		return "", fmt.Errorf("failed to execute interface template: %w", err)
	}
	return builder.String(), nil
}

// parseSinglePostMethod 处理单个 POST 方法
func parseSinglePostMethod(method *transfer.Method, collectMap map[string]struct{}, reg *gtpl.Registry) (string, string, error) {
	tp, err := getInterfacePostMethodInputParamType(*method.PostParams, collectMap)
	if err != nil {
		return "", "", err
	}
	v := common.FormatVariable(method.PostParams.ParamName, false)
	returnType, err := getInterfaceMethodReturnType(method.MethodReturnType, collectMap)
	if err != nil {
		return "", "", err
	}
	interfaceDataFields := InterfaceField{
		Name:       method.Name,
		Sep:        ",",
		InputName:  v,
		InputType:  tp,
		ReturnType: returnType,
	}

	// Render interface method
	interfaceStr, err := renderInterfaceMethod(reg, interfaceDataFields)
	if err != nil {
		return "", "", err
	}

	// Render request handler
	functionCallField := FunctionCallField{
		Method:       HTTPMethodPost,
		InputType:    interfaceDataFields.InputType,
		ReturnType:   interfaceDataFields.ReturnType,
		RelativePath: method.Url,
		NonLogin:     method.NotLogin,
		BizHandler:   fmt.Sprintf("svc.%s", method.Name),
		Products:     method.Products,
	}
	handlerStr, err := renderRequestHandler(reg, functionCallField)
	if err != nil {
		return "", "", err
	}

	return interfaceStr, handlerStr, nil
}

// parseSingleGetMethod 处理单个 GET 方法
func parseSingleGetMethod(method *transfer.Method, collectMap map[string]struct{}, reg *gtpl.Registry) (string, string, error) {
	tp, v := getInterfaceGetMethodInputParamTypeFromMethod(method, collectMap)
	returnType, err := getInterfaceMethodReturnType(method.MethodReturnType, collectMap)
	if err != nil {
		return "", "", err
	}
	interfaceDataFields := InterfaceField{
		Name:       method.Name,
		Sep:        ",",
		InputName:  common.FormatVariable(v, false),
		ReturnType: returnType,
	}
	if tp != "" {
		interfaceDataFields.InputType = tp
	} else if v == "" {
		interfaceDataFields.InputType = "commons.Void"
		interfaceDataFields.InputName = "t"
	}

	// Render interface method
	interfaceStr, err := renderInterfaceMethod(reg, interfaceDataFields)
	if err != nil {
		return "", "", err
	}

	// Render request handler
	functionCallField := FunctionCallField{
		Method:       HTTPMethodGet,
		InputType:    interfaceDataFields.InputType,
		ReturnType:   interfaceDataFields.ReturnType,
		RelativePath: method.Url,
		NonLogin:     method.NotLogin,
		BizHandler:   fmt.Sprintf("svc.%s", method.Name),
		Products:     method.Products,
	}
	handlerStr, err := renderRequestHandler(reg, functionCallField)
	if err != nil {
		return "", "", err
	}

	return interfaceStr, handlerStr, nil
}

// parseSinglePutMethod 处理单个 PUT 方法
func parseSinglePutMethod(method *transfer.Method, collectMap map[string]struct{}, reg *gtpl.Registry) (string, string, error) {
	tp, err := getInterfacePostMethodInputParamType(*method.PostParams, collectMap)
	if err != nil {
		return "", "", err
	}
	v := common.FormatVariable(method.PostParams.ParamName, false)
	returnType, err := getInterfaceMethodReturnType(method.MethodReturnType, collectMap)
	if err != nil {
		return "", "", err
	}
	interfaceDataFields := InterfaceField{
		Name:       method.Name,
		Sep:        ",",
		InputName:  v,
		InputType:  tp,
		ReturnType: returnType,
	}

	// Render interface method
	interfaceStr, err := renderInterfaceMethod(reg, interfaceDataFields)
	if err != nil {
		return "", "", err
	}

	// Render request handler
	functionCallField := FunctionCallField{
		Method:       HTTPMethodPut,
		InputType:    interfaceDataFields.InputType,
		ReturnType:   interfaceDataFields.ReturnType,
		RelativePath: method.Url,
		NonLogin:     method.NotLogin,
		BizHandler:   fmt.Sprintf("svc.%s", method.Name),
		Products:     method.Products,
	}
	handlerStr, err := renderRequestHandler(reg, functionCallField)
	if err != nil {
		return "", "", err
	}

	return interfaceStr, handlerStr, nil
}

// getInterfaceGetMethodInputParamTypeFromMethod 从 Method 获取 GET 方法输入参数类型
func getInterfaceGetMethodInputParamTypeFromMethod(method *transfer.Method, collectMap map[string]struct{}) (string, string) {
	if method.GetParams == nil || method.GetParams.IsEmpty {
		return "", ""
	}

	if method.GetParams.IsSingleStruct {
		result, param := method.GetParams.StructName, method.GetParams.StructParamName
		if result == commonModelIDReqStruct ||
			result == commonModelIDListReqStruct ||
			result == commonModelStringReqStruct ||
			result == commonModelStringListReqStruct {
			return result, param
		}

		if strings.Contains(result, ".") {
			items := strings.Split(result, ".")
			collectMap[items[0]] = struct{}{}
			return result, param
		}
		return paramStructsPrefix + result, param
	}

	return "", ""
}

func generateGoServiceImplFile(funcs []string, implName string, reg *gtpl.Registry) (string, error) {
	var result strings.Builder

	for _, f := range funcs {
		fNew := strings.ReplaceAll(f, "\n\t", "")
		interfaceAchieveField := InterfaceAchieveField{
			InterfaceName:  fmt.Sprintf("%s%s%s", common.FormatVariable(implName, false), controllerName, stdServiceImplSuffix),
			FuncDefinition: strings.ReplaceAll(fNew, "\n", ""),
		}

		tmpl, err := reg.Get(gtpl.TplImplMethod)
		if err != nil {
			return "", fmt.Errorf("failed to get impl method template: %w", err)
		}

		err = tmpl.Execute(&result, interfaceAchieveField)
		if err != nil {
			return "", fmt.Errorf("failed to execute impl method template: %w", err)
		}
	}
	return result.String(), nil
}

func renderInterfaceMethod(reg *gtpl.Registry, data InterfaceField) (string, error) {
	tmpl, err := reg.Get(gtpl.TplInterfaceMethod)
	if err != nil {
		return "", fmt.Errorf("failed to get interface method template: %w", err)
	}
	var builder strings.Builder
	if err := tmpl.Execute(&builder, data); err != nil {
		return "", fmt.Errorf("failed to execute interface method template: %w", err)
	}
	return builder.String(), nil
}

func renderRequestHandler(reg *gtpl.Registry, data FunctionCallField) (string, error) {
	tmpl, err := reg.Get(gtpl.TplRequestHandler)
	if err != nil {
		return "", fmt.Errorf("failed to get request handler template: %w", err)
	}
	var builder strings.Builder
	if err := tmpl.Execute(&builder, data); err != nil {
		return "", fmt.Errorf("failed to execute request handler template: %w", err)
	}
	return builder.String(), nil
}
