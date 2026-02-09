package java

import (
	"bytes"
	"fmt"
	"icomplie/common"
	"icomplie/internal/transfer"
	"icomplie/pkg/codegen/java/client"
	jtpl "icomplie/pkg/codegen/java/template"
	"icomplie/pkg/types"
	"os"
	"path/filepath"
)

// generatePOJO generates a Java POJO class for a struct
func generatePOJO(outputDir, packagePath, namespace string, st *transfer.StructDefine) (string, error) {
	className := st.Name
	fileName := filepath.Join(outputDir, className+".java")

	reg := jtpl.GetDefaultRegistry()

	render := &client.POJORender{
		PackagePath: packagePath,
		Namespace:   namespace,
		ClassName:   className,
		Extends:     st.Extends,
		HasExtends:  st.Extends != "",
		HasList:     hasListField(st.Fields),
		HasMap:      hasMapField(st.Fields),
		Fields:      make([]*client.FieldRender, 0, len(st.Fields)),
	}

	for _, field := range st.Fields {
		javaType, err := getJavaType(field.Tp)
		if err != nil {
			return "", fmt.Errorf("failed to get Java type for field %s: %w", field.Name, err)
		}
		fieldName := common.FormatVariable(field.Name, false)
		render.Fields = append(render.Fields, &client.FieldRender{
			Type: javaType,
			Name: fieldName,
		})
	}

	tmpl, err := reg.Get(jtpl.TplPOJO)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, render); err != nil {
		return "", fmt.Errorf("failed to execute POJO template: %w", err)
	}

	if err := os.WriteFile(fileName, buf.Bytes(), 0644); err != nil {
		return "", fmt.Errorf("failed to write POJO file %s: %w", fileName, err)
	}

	return fileName, nil
}

// getJavaType converts an IDL field type to Java type
func getJavaType(ft *transfer.FieldType) (string, error) {
	if ft == nil {
		return "Object", nil
	}

	if ft.IsBasic {
		return getBasicJavaType(ft.TypeName), nil
	}

	if ft.IsStruct {
		return ft.TypeName, nil
	}

	if ft.IsList {
		if ft.ValueType == nil {
			return "List<Object>", nil
		}
		innerType, err := getJavaType(ft.ValueType)
		if err != nil {
			return "", err
		}
		// Box primitive types for generics
		innerType = boxJavaType(innerType)
		return fmt.Sprintf("List<%s>", innerType), nil
	}

	if ft.IsMap {
		keyType := getBasicJavaType(ft.TypeName)
		keyType = boxJavaType(keyType)
		valueType := "Object"
		if ft.ValueType != nil {
			var err error
			valueType, err = getJavaType(ft.ValueType)
			if err != nil {
				return "", err
			}
			valueType = boxJavaType(valueType)
		}
		return fmt.Sprintf("Map<%s, %s>", keyType, valueType), nil
	}

	return "Object", nil
}

// getBasicJavaType converts an IDL basic type to Java type
func getBasicJavaType(typeName string) string {
	registry := types.NewRegistry()
	if t, ok := registry.Get(typeName); ok {
		return t.JavaType()
	}

	// Fallback mappings
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

// boxJavaType converts primitive types to their boxed equivalents
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

// hasListField checks if any field is a list type
func hasListField(fields []*transfer.Field) bool {
	for _, f := range fields {
		if f.Tp != nil && f.Tp.IsList {
			return true
		}
	}
	return false
}

// hasMapField checks if any field is a map type
func hasMapField(fields []*transfer.Field) bool {
	for _, f := range fields {
		if f.Tp != nil && f.Tp.IsMap {
			return true
		}
	}
	return false
}
