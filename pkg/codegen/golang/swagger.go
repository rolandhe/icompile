package golang

import (
	"encoding/json"
	"fmt"
	"github.com/go-openapi/spec"
	"icomplie/common"
	"icomplie/internal/transfer"
	"os"
	"strings"
)

/*
	基础 schema 结构体：
		spec.SchemaProps{
			Type:       []string{"object"},
			Properties: schema,
		},
*/

const (
	swaggerVersion = "2.0"
	title          = "接口文档"
	desc           = "接口描述"
	version        = "v1.0.0"
	//swaggerFile    = "swagger.json"
)

const (
	pagerPageNo     = "pageNo"
	pagerPageSize   = "pageSize"
	pagerTotalCount = "totalCount"
	pagerTotalPages = "totalPages"
	pagerList       = "list"
)

const (
	returnSchemaCode    = "code"
	returnSchemaMessage = "errMsg"
	//returnSchemaSuccess = "success"
	returnSchemaData = "data"
)

const (
	swaggerObject = "object"
	swaggerArray  = "array"
)

func generateSwaggerFile(def *transfer.Definition, outFileName string) error {
	if len(def.Services) == 0 {
		return nil
	}
	//initExtend(def.Structs)

	infoProp := spec.InfoProps{
		Title:       fmt.Sprintf("%s%s", def.Namespace, title),
		Description: fmt.Sprintf("%s%s", def.Namespace, desc),
		Version:     version,
	}
	info := spec.Info{
		InfoProps: infoProp,
	}
	paths, err := getApiPaths(def)
	if err != nil {
		return err
	}
	swaggerProps := spec.SwaggerProps{
		Swagger:             swaggerVersion,
		Definitions:         spec.Definitions{},
		SecurityDefinitions: spec.SecurityDefinitions{},
		Info:                &info,
		Paths:               paths,
	}

	return saveSwagger(swaggerProps, outFileName)
}

func makeCommonResponses(def *transfer.Definition, returnType transfer.MethodReturnType) *spec.Responses {
	var schema spec.SchemaProps
	if returnType.IsList {
		schema = getReturnSchemaArray(def, returnType)
	} else {
		schema = getReturnSchemaObject(def, returnType)
	}
	responses := spec.Responses{}
	response200 := spec.Response{
		ResponseProps: spec.ResponseProps{
			Description: "成功",
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: []string{swaggerObject},
					//Title: "名字",
					Properties: map[string]spec.Schema{
						//returnSchemaSuccess: {
						//	SchemaProps: spec.SchemaProps{
						//		Type: []string{RESTFUL_STRING},
						//	},
						//},
						returnSchemaCode: {
							SchemaProps: spec.SchemaProps{
								Type: []string{RESTFUL_INTEGER},
							},
						},
						returnSchemaMessage: {
							SchemaProps: spec.SchemaProps{
								Type: []string{RESTFUL_STRING},
							},
						},
						returnSchemaData: {
							SchemaProps: schema,
						},
					},
				},
			},
		},
	}
	responses.ResponsesProps = spec.ResponsesProps{
		StatusCodeResponses: map[int]spec.Response{
			RESTFUL_RESPONSE_OK: response200,
		},
	}
	return &responses
}

// 将Swagger JSON保存到文件
func saveSwagger(swaggerProps spec.SwaggerProps, filename string) error {
	// 将Swagger文档对象转换为JSON
	swaggerJSON, err := json.MarshalIndent(swaggerProps, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal Swagger JSON: %w", err)
	}

	err = os.WriteFile(filename, swaggerJSON, 0644)
	if err != nil {
		return fmt.Errorf("failed to write Swagger file %s: %w", filename, err)
	}
	return nil
}

func getApiPaths(def *transfer.Definition) (*spec.Paths, error) {
	paths := map[string]spec.PathItem{}

	for _, service := range def.Services {
		if err := getPostPaths(def, paths, service); err != nil {
			return nil, err
		}
		if err := getGetPaths(def, paths, service); err != nil {
			return nil, err
		}
		if err := getPutPaths(def, paths, service); err != nil {
			return nil, err
		}
	}

	return &spec.Paths{
		Paths: paths,
	}, nil
}

func getPostPaths(def *transfer.Definition, paths map[string]spec.PathItem, service *transfer.ServiceDefine) error {
	for _, postApi := range service.Posts {
		url := fmt.Sprintf("%s/%s", service.RootUrl, postApi.Url)
		var description string
		if strings.Contains(postApi.Description, "@") {
			description = postApi.Description[:strings.Index(postApi.Description, "@")]
		} else {
			description = postApi.Description
		}

		params, err := getPostParameters(postApi, def)
		if err != nil {
			return err
		}

		paths[url] = spec.PathItem{
			PathItemProps: spec.PathItemProps{
				Post: &spec.Operation{
					OperationProps: spec.OperationProps{
						ID:          postApi.Name,
						Summary:     description,
						Description: description,
						Consumes:    []string{RESTFUL_CONTENT_TYPE_JSON},
						Produces:    []string{RESTFUL_CONTENT_TYPE_JSON},
						Parameters:  params,
						Responses:   makeCommonResponses(def, postApi.MethodReturnType),
					},
				},
			},
		}
	}
	return nil
}

func getGetPaths(def *transfer.Definition, paths map[string]spec.PathItem, service *transfer.ServiceDefine) error {
	for _, getApi := range service.Gets {
		url := fmt.Sprintf("%s/%s", service.RootUrl, getApi.Url)
		var description string
		if strings.Contains(getApi.Description, "@") {
			description = getApi.Description[:strings.Index(getApi.Description, "@")]
		} else {
			description = getApi.Description
		}

		params, err := getGetParameters(getApi, def)
		if err != nil {
			return err
		}

		paths[url] = spec.PathItem{
			PathItemProps: spec.PathItemProps{
				Get: &spec.Operation{
					OperationProps: spec.OperationProps{
						ID:          getApi.Name,
						Summary:     description,
						Description: description,
						Consumes:    []string{RESTFUL_CONTENT_TYPE_JSON},
						Parameters:  params,
						Responses:   makeCommonResponses(def, getApi.MethodReturnType),
					},
				},
			},
		}
	}
	return nil
}

func getPutPaths(def *transfer.Definition, paths map[string]spec.PathItem, service *transfer.ServiceDefine) error {
	for _, putApi := range service.Puts {
		url := fmt.Sprintf("%s/%s", service.RootUrl, putApi.Url)
		var description string
		if strings.Contains(putApi.Description, "@") {
			description = putApi.Description[:strings.Index(putApi.Description, "@")]
		} else {
			description = putApi.Description
		}

		params, err := getPutParameters(putApi, def)
		if err != nil {
			return err
		}

		paths[url] = spec.PathItem{
			PathItemProps: spec.PathItemProps{
				Put: &spec.Operation{
					OperationProps: spec.OperationProps{
						ID:          putApi.Name,
						Summary:     description,
						Description: description,
						Consumes:    []string{RESTFUL_CONTENT_TYPE_JSON},
						Produces:    []string{RESTFUL_CONTENT_TYPE_JSON},
						Parameters:  params,
						Responses:   makeCommonResponses(def, putApi.MethodReturnType),
					},
				},
			},
		}
	}
	return nil
}

func getPutParameters(putApi *transfer.PutMethod, def *transfer.Definition) ([]spec.Parameter, error) {
	properties := make(map[string]spec.Schema)
	var required []string

	if putApi.Params.StructName == "commons.IdRequest" {
		properties["id"] = spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type: []string{"integer"},
			},
		}
		p := *spec.BodyParam(RESTFUL_BODY_KEY, &spec.Schema{
			SchemaProps: spec.SchemaProps{
				Properties: properties,
				Required:   []string{"id"},
			},
		})
		p.Required = true
		return []spec.Parameter{p}, nil
	}

	if putApi.Params.StructName == "commons.IdListReq" {
		properties["idList"] = spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type: []string{"array"},
				Items: &spec.SchemaOrArray{
					Schema: &spec.Schema{
						SchemaProps: spec.SchemaProps{
							Type: []string{"integer"},
						},
					},
				},
			},
		}
		p := *spec.BodyParam(RESTFUL_BODY_KEY, &spec.Schema{
			SchemaProps: spec.SchemaProps{
				Properties: properties,
				Required:   []string{"idList"},
			},
		})
		p.Required = true
		return []spec.Parameter{p}, nil
	}

	if putApi.Params.StructName != "" {
		var err error
		required, err = genStructSchema(properties, def, putApi.Params.StructName)
		if err != nil {
			return nil, err
		}
	}

	parameter := []spec.Parameter{
		*spec.BodyParam(RESTFUL_BODY_KEY, &spec.Schema{
			SchemaProps: spec.SchemaProps{
				Properties: properties,
				Required:   required,
			},
		}),
	}
	return parameter, nil
}

func getPostParameters(postApi *transfer.PostMethod, def *transfer.Definition) ([]spec.Parameter, error) {
	properties := make(map[string]spec.Schema)
	var required []string

	if postApi.Params.StructName == "commons.IdRequest" {
		properties["id"] = spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type: []string{"integer"},
			},
		}
		p := *spec.BodyParam(RESTFUL_BODY_KEY, &spec.Schema{
			SchemaProps: spec.SchemaProps{
				Properties: properties,
				Required:   []string{"id"},
			},
		})
		p.Required = true
		return []spec.Parameter{p}, nil
	}

	if postApi.Params.StructName == "commons.IdListReq" {
		properties["idList"] = spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type: []string{"array"}, // 设置为数组类型
				Items: &spec.SchemaOrArray{
					Schema: &spec.Schema{
						SchemaProps: spec.SchemaProps{
							Type: []string{"integer"}, // 数组的项是整数
						},
					},
				},
			},
		}
		p := *spec.BodyParam(RESTFUL_BODY_KEY, &spec.Schema{
			SchemaProps: spec.SchemaProps{
				Properties: properties,
				Required:   []string{"idList"},
			},
		})
		p.Required = true
		return []spec.Parameter{p}, nil
	}

	if postApi.Params.StructName != "" {
		var err error
		required, err = genStructSchema(properties, def, postApi.Params.StructName)
		if err != nil {
			return nil, err
		}
	}

	parameter := []spec.Parameter{
		*spec.BodyParam(RESTFUL_BODY_KEY, &spec.Schema{
			SchemaProps: spec.SchemaProps{
				Properties: properties,
				Required:   required,
			},
		}),
	}
	return parameter, nil
}

func genStructSchema(properties map[string]spec.Schema, def *transfer.Definition, name string) ([]string, error) {
	st, myDef := def.GetStruct(name)
	if st == nil {
		return nil, fmt.Errorf("struct %s not found", name)
	}
	return forOneStructSchema(properties, st, myDef)
}

type forSchemaField struct {
	field *transfer.Field
	myDef *transfer.Definition
}

func (fsf *forSchemaField) getRemark() string {
	for _, anno := range fsf.field.Annotations {
		if anno.Key == "remark" {
			return anno.Value
		}
	}
	return ""
}

func (fsf *forSchemaField) getTypeName() (string, error) {
	if fsf.field.Tp.IsList {
		return fsf.field.Tp.ValueType.TypeName, nil
	}
	if fsf.field.Tp.IsStruct || fsf.field.Tp.IsBasic {
		return fsf.field.Tp.TypeName, nil
	}

	return "", fmt.Errorf("unsupported type in field %s", fsf.field.Name)
}

func forOneStructSchema(properties map[string]spec.Schema, st *transfer.StructDefine, def *transfer.Definition) ([]string, error) {
	lmap := newLinkedMapContainer[*forSchemaField]()
	//if st.Extends != "" {
	//	extSt, myDef := def.GetStruct(st.Extends)
	//	if extSt == nil {
	//		panic(fmt.Sprintf("struct %s not found", st.Extends))
	//	}
	//	extractStructFields(extSt, myDef, lmap)
	//}
	if err := extractStructFields(st, def, lmap); err != nil {
		return nil, err
	}

	var required []string

	for _, name := range lmap.getList() {
		sst := lmap.getValue(name)
		nameFmt := common.FormatVariable(sst.field.Name, false)
		schema, err := genSchema(sst)
		if err != nil {
			return nil, err
		}
		properties[nameFmt] = schema
		if sst.field.ReqDefine == Required {
			required = append(required, name)
		}
	}

	return required, nil
}

//func extractStructFields(st *transfer.StructDefine, def *transfer.Definition, lmap *linkedMapContainer[*forSchemaField]) {
//	for _, field := range st.Fields {
//		lmap.add(field.Name, &forSchemaField{
//			field: field,
//			myDef: def,
//		})
//	}
//}

func extractStructFields(st *transfer.StructDefine, def *transfer.Definition, lmap *linkedMapContainer[*forSchemaField]) error {
	if st.Extends != "" {
		extSt, myDef := def.GetStruct(st.Extends)
		if extSt == nil {
			return fmt.Errorf("struct %s not found", st.Extends)
		}
		if err := extractStructFields(extSt, myDef, lmap); err != nil {
			return err
		}
	}
	for _, field := range st.Fields {
		lmap.add(field.Name, &forSchemaField{
			field: field,
			myDef: def,
		})
	}
	return nil
}

func getReturnSchemaObject(def *transfer.Definition, returnType transfer.MethodReturnType) spec.SchemaProps {
	var result spec.SchemaProps
	properties := make(map[string]spec.Schema)
	if returnType.IsVoid {
		return result
	}

	if returnType.IsPager { // 分页按照数组处理
		result.Type = []string{swaggerObject}
		returnType.IsPager = false
		result.Properties = getPageListSchema()
		result.Properties[pagerList] = spec.Schema{
			SchemaProps: getReturnSchemaArray(def, returnType),
		}
		return result
	}

	if returnType.IsStruct {
		required, _ := genStructSchema(properties, def, returnType.TypeName)
		result.Required = required
		result.Type = []string{swaggerObject}
	} else {
		if returnType.TypeName != "" {
			result.Type = []string{convertTp2Swagger(parseBasic(returnType.TypeName))}
		} else {
			result.Type = []string{}
		}
	}

	result.Properties = properties
	return result
}

func getReturnSchemaArray(def *transfer.Definition, returnType transfer.MethodReturnType) spec.SchemaProps {
	//returnType.IsVoid = false
	//returnType.IsList = false
	properties := getReturnSchemaObject(def, returnType)
	return spec.SchemaProps{
		Type: []string{swaggerArray},
		Items: &spec.SchemaOrArray{
			Schema: &spec.Schema{
				SchemaProps: properties,
			},
		},
	}
}

func getGetParameters(getApi *transfer.GetMethod, def *transfer.Definition) ([]spec.Parameter, error) {
	if getApi.Params.StructName == "commons.IdRequest" {
		p := *spec.QueryParam("id")
		p.Type = "integer"
		p.Required = true
		return []spec.Parameter{p}, nil
	}

	var parameter []spec.Parameter

	if getApi.Params.IsSingleStruct {
		st, myDef := def.GetStruct(getApi.Params.StructName)
		if st == nil {
			return nil, fmt.Errorf("struct does not exist: %s", getApi.Params.StructName)
		}
		lmap := newLinkedMapContainer[*forSchemaField]()
		if err := extractStructFields(st, myDef, lmap); err != nil {
			return nil, err
		}
		parameter = []spec.Parameter{}
		for _, name := range lmap.getList() {
			f := lmap.getValue(name)
			if !f.field.Tp.IsBasic {
				return nil, fmt.Errorf("field %s is not basic type", f.field.Name)
			}
			nameFmt := common.FormatVariable(name, false)

			p := *spec.QueryParam(nameFmt)
			p.Type = convertTp2Swagger(parseBasic(f.field.Tp.TypeName))
			if f.field.ReqDefine == Required {
				p.Required = true
			}
			p.Description = f.getRemark()

			parameter = append(parameter, p)
		}
	} else { // 单个参数
		var (
			tp, v    string
			required string
		)

		parameter = []spec.Parameter{}
		for _, basicParam := range getApi.Params.BasicParams {
			tp, v = basicParam.TypeName, basicParam.ParamName
			required = basicParam.ReqDefine
			name := common.FormatVariable(v, false)
			p := *spec.QueryParam(name)
			p.Type = convertTp2Swagger(parseBasic(tp))
			if required == Required {
				p.Required = true
			}
			p.Description = getRemarkValue(basicParam.Annotations)
			parameter = append(parameter, p)
		}

	}

	return parameter, nil
}

func convertTp2Swagger(tp string) string {
	switch tp {
	case "int":
		return "integer"
	case "int8":
		return "integer"
	case "int16":
		return "integer"
	case "int32":
		return "integer"
	case "int64":
		return "integer"
	case "bool":
		return "boolean"
	case "float32":
		return "number"
	case "float64":
		return "number"
	case "string":
		return "string"
	}

	return ""
	//panic(errors.New("unsupported type " + tp))
}

func genSchema(sst *forSchemaField) (spec.Schema, error) {
	if sst.field.Tp.IsList {
		result := spec.Schema{}
		result.Type = []string{swaggerArray}
		if sst.field.Tp.ValueType.IsStruct {
			structDef, myDef := sst.myDef.GetStruct(sst.field.Tp.ValueType.TypeName)

			if structDef == nil {
				return spec.Schema{}, fmt.Errorf("struct does not exist: %s", sst.field.Tp.ValueType.TypeName)
			}
			// 结构体数组
			sp, err := createSchemaForStruct(myDef, structDef.Name, "")
			if err != nil {
				return spec.Schema{}, err
			}
			result.Items = &spec.SchemaOrArray{
				Schema: &sp,
			}
		} else { // 基本类型数组
			typeName, err := sst.getTypeName()
			if err != nil {
				return spec.Schema{}, err
			}
			result.Items = &spec.SchemaOrArray{
				Schema: &spec.Schema{
					SchemaProps: spec.SchemaProps{
						Type:        []string{convertTp2Swagger(parseBasic(typeName))},
						Description: sst.getRemark(),
					},
				},
			}
		}
		return result, nil
	}

	if !sst.field.Tp.IsStruct {
		return spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type:        []string{convertTp2Swagger(parseBasic(sst.field.Tp.TypeName))},
				Description: sst.getRemark(),
			},
		}, nil
	}
	typeName, err := sst.getTypeName()
	if err != nil {
		return spec.Schema{}, err
	}
	return createSchemaForStruct(sst.myDef, typeName, sst.getRemark())
}

func createSchemaForStruct(myDef *transfer.Definition, structName, remark string) (spec.Schema, error) {
	properties := make(map[string]spec.Schema)
	required, err := genStructSchema(properties, myDef, structName)
	if err != nil {
		return spec.Schema{}, err
	}
	result := spec.Schema{}

	//if inList {
	//	result.Type = []string{swaggerArray}
	//	result.Items = &spec.SchemaOrArray{
	//		Schema: &spec.Schema{
	//			SchemaProps: spec.SchemaProps{
	//				Type:        []string{swaggerObject},
	//				Description: sst.getRemark(),
	//				Properties:  properties,
	//				Required:    required,
	//			},
	//		},
	//	}
	//	return result
	//} else {
	result.Type = []string{swaggerObject}
	result.Description = remark
	result.Properties = properties
	result.Required = required
	//}

	return result, nil
}

func getPageListSchema() map[string]spec.Schema {
	return map[string]spec.Schema{
		pagerPageNo: {
			SchemaProps: spec.SchemaProps{
				Type:        []string{RESTFUL_INTEGER},
				Description: "页码",
			},
		},
		pagerPageSize: {
			SchemaProps: spec.SchemaProps{
				Type:        []string{RESTFUL_INTEGER},
				Description: "页面大小",
			},
		},
		pagerTotalCount: {
			SchemaProps: spec.SchemaProps{
				Type:        []string{RESTFUL_INTEGER},
				Description: "总数",
			},
		},
		pagerTotalPages: {
			SchemaProps: spec.SchemaProps{
				Type:        []string{RESTFUL_INTEGER},
				Description: "总页数",
			},
		},
	}
}
