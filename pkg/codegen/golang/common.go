package golang

import (
	"fmt"
	"icomplie/common"
	"icomplie/internal/transfer"
)

const (
	HTTPMethodPost = "Post"
	HTTPMethodGet  = "Get"
	HTTPMethodPut  = "Put"
)

const (
	Required        = "required"
	Remark          = "remark"
	BindingRequired = "binding:\"required\""
)

const (
	TYPE_BOOL   = "bool"
	TYPE_BYTE   = "byte"
	TYPE_I8     = "i8"
	TYPE_I16    = "i16"
	TYPE_I32    = "i32"
	TYPE_I64    = "i64"
	TYPE_FLOAT  = "float"
	TYPE_DOUBLE = "double"
	TYPE_STRING = "string"
)

const (
	RESTFUL_RESPONSE_OK       = 200
	RESTFUL_CONTENT_TYPE_JSON = "application/json"
	RESTFUL_BODY_KEY          = "body"
	RESTFUL_INTEGER           = "integer"
	RESTFUL_STRING            = "string"
	RESTFUL_BOOLEAN           = "boolean"
)

func converseStructFieldsWithFormTag(structFields []*transfer.Field, lm *LinkedSetString, hasFormTag bool) []*RenderField {
	var fields []*RenderField
	for _, f := range structFields {
		fieldType, err := ParseFieldType(f.Tp)
		if err != nil {
			// For backward compatibility, use empty string on error
			fieldType = ""
		}
		field := RenderField{
			Name:   common.Capitalize(f.Name),
			Type:   fieldType,
			Tag:    generateTags(f.ReqDefine, f.Name, hasFormTag, f.Annotations),
			IsList: f.Tp.IsList,
			//StructName:
		}

		if f.ReqDefine == Required {
			field.Required = true
		}

		if f.Tp.IsStruct {
			field.StructName = f.Tp.TypeName
			lm.add(field.StructName)
		} else if f.Tp.IsList {
			if f.Tp.ValueType.IsStruct {
				field.StructName = f.Tp.ValueType.TypeName
				lm.add(field.StructName)
			}
			field.SourceType = f.Tp.ValueType.TypeName
		}

		field.Remark = getRemarkValue(f.Annotations)

		fields = append(fields, &field)
	}

	return fields
}

// generateTags generates struct field tags
// hasFormTag: if true, adds form tag alongside json tag
func generateTags(reqDefine, s string, hasFormTag bool, annotations []*transfer.Annotation) string {
	sTag := ""
	tag := common.FormatVariable(s, false)
	sTag += fmt.Sprintf("json:\"%s\"", tag)

	if hasFormTag {
		sTag += fmt.Sprintf(" form:\"%s\"", tag)
	}
	//if reqDefine == Required {
	//	sTag += " " + BindingRequired
	//}

	for _, an := range annotations {
		if an.Key == "remark" {
			continue
		}
		sTag += " " + an.Key + ":" + an.Value
	}
	return fmt.Sprintf("`%s`", sTag)
}

func getRemarkValue(annotation []*transfer.Annotation) string {
	for _, obj := range annotation {
		if obj.Key == Remark {
			return obj.Value
		}
	}

	return ""
}
