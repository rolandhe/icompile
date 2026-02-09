package golang

import (
	"fmt"
	"icomplie/common"
	"icomplie/internal/transfer"
	"os"
	"path/filepath"
	"strings"
)

// RenderField 定义字段结构体
type RenderField struct {
	Name         string
	Type         string
	Tag          string
	Required     bool
	StructName   string
	Remark       string // 注释
	IsList       bool   // 是否是list
	SourceType   string // 原类型
	ExtendStruct string // 继承的结构体
}

// FunctionCallField 定义函数调用字段结构体
type FunctionCallField struct {
	Method       string
	InputType    string
	ReturnType   string
	RelativePath string
	NonLogin     bool
	BizHandler   string
	Products     string
}

// InterfaceField 定义接口字段结构体
type InterfaceField struct {
	Name       string
	Sep        string
	InputName  string
	InputType  string
	ReturnType string
}

// InterfaceAchieveField 定义接口实现字段结构体
type InterfaceAchieveField struct {
	InterfaceName  string
	FuncDefinition string
	ReturnType     string
}

// ControllerHeaderData 控制器文件头数据
type ControllerHeaderData struct {
	PackageName  string
	PkgPath      string
	ExtraImports []string
}

// StructsHeaderData 结构体文件头数据
type StructsHeaderData struct {
	PackageName string
	Imports     string
}

// ImplHeaderData 实现文件头数据
type ImplHeaderData struct {
	PackageName  string
	PkgPath      string
	ExtraImports []string
}

// StructRenderData 结构体渲染数据
type StructRenderData struct {
	Name         string
	Fields       []*RenderField
	ExtendStruct string
	ExtendPoint  bool
}

// InterfaceRenderData 接口渲染数据
type InterfaceRenderData struct {
	Name    string
	Methods []string // 已渲染的方法声明
}

// BindFunctionData Bind函数数据
type BindFunctionData struct {
	ServiceName      string
	ServiceNameLower string
	RootUrl          string
	Handlers         []string // 已渲染的处理器
}

// ImplStructData 实现结构体数据
type ImplStructData struct {
	Name string
}

// 非模板常量
const (
	controllerName       = "Controller"
	stdServiceSuffix     = "Std"
	stdServiceImplSuffix = "Impl"
)

const (
	outputGoFileStructSuffix       = "_structs.go"
	outputGoFileInterfaceSuffix    = "_controller.go"
	outputGoFileImplSuffix         = "_controller_impl.go"
	paramStructsPrefix             = "structs."
	commonModelIDReqStruct         = "commons.IdRequest"
	commonModelIDListReqStruct     = "commons.IdListReq"
	commonModelStringReqStruct     = "commons.StringReq"
	commonModelStringListReqStruct = "commons.StringListReq"
)

const (
	voidType = "*commons.Void"
)

func ParseFieldType(tp *transfer.FieldType) (string, error) {
	if tp.IsBasic {
		return parseBasic(tp.TypeName), nil
	}
	if tp.IsStruct {
		return "*" + tp.TypeName, nil
	}
	if tp.IsList {
		vt, err := ParseFieldType(tp.ValueType)
		if err != nil {
			return "", err
		}
		return "[]" + vt, nil
	}
	if tp.IsMap {
		vt, err := ParseFieldType(tp.ValueType)
		if err != nil {
			return "", err
		}
		return "map[" + parseBasic(tp.TypeName) + "]" + vt, nil
	}

	return "", fmt.Errorf("invalid field type: %#v", tp)
}

func parseBasic(typeName string) string {
	switch typeName {
	case TYPE_BOOL:
		return "bool"
	case TYPE_BYTE:
		return "byte"
	case TYPE_I8:
		return "int8"
	case TYPE_I16:
		return "int16"
	case TYPE_I32:
		return "int32"
	case TYPE_I64:
		return "int64"
	case TYPE_FLOAT:
		return "float32"
	case TYPE_DOUBLE:
		return "float64"
	case TYPE_STRING:
		return "string"
	default:
	}
	return ""
}

func save(pkgName string, filename string, content string) error {
	if err := common.CreateDir(pkgName); err != nil {
		return err
	}
	filePath := filepath.Join(pkgName, filename)
	// 创建新的代码文件
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer file.Close()

	// 将生成的代码写入文件
	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	if err := common.FormatGoFile(filePath); err != nil {
		return fmt.Errorf("failed to format file %s: %w", filePath, err)
	}
	return nil
}

func getInterfacePostMethodInputParamType(postParam transfer.PostParam, collectMap map[string]struct{}) (string, error) {
	r := ""
	if postParam.IsEmpty {
		return "", fmt.Errorf("post param is empty: %#v", postParam)
	}
	if postParam.StructName == "" {
		return "", fmt.Errorf("post param type error: %#v", postParam)
	}

	if postParam.IsList {
		r += "[]" + r
	}

	if postParam.StructName == commonModelIDReqStruct ||
		postParam.StructName == commonModelIDListReqStruct ||
		postParam.StructName == commonModelStringReqStruct ||
		postParam.StructName == commonModelStringListReqStruct {
		r += postParam.StructName
		return r, nil
	}

	if strings.Contains(postParam.StructName, ".") {
		items := strings.Split(postParam.StructName, ".")
		collectMap[items[0]] = struct{}{}
		r += postParam.StructName
		return r, nil
	}
	r += paramStructsPrefix + postParam.StructName

	return r, nil
}

func getInterfaceMethodReturnType(returnType transfer.MethodReturnType, collectMap map[string]struct{}) (string, error) {
	r := ""
	if returnType.IsList {
		r = "[]" + r
	}

	if returnType.IsPager && !returnType.IsStruct {
		return "", fmt.Errorf("pageable requires struct return type")
	}

	structNameFunc := func(name string) string {
		if strings.Contains(name, ".") {
			items := strings.Split(name, ".")
			collectMap[items[0]] = struct{}{}
			return name
		}
		return paramStructsPrefix + returnType.TypeName
	}

	if returnType.IsStruct {
		if returnType.IsPager {
			r = fmt.Sprintf("*commons.PageList[%s]", structNameFunc(returnType.TypeName))
		} else {
			r += "*" + structNameFunc(returnType.TypeName)
		}
	} else {
		r += parseBasic(returnType.TypeName)
	}
	if returnType.IsVoid {
		r = voidType
	}

	return r, nil
}
