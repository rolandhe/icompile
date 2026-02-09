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

// generateGoStructFile generates the Go struct file
func generateGoStructFile(def *transfer.Definition, dirPath, serviceIdlFileName string, onlyStruct bool) error {
	structs := def.Structs
	if len(structs) == 0 {
		return nil
	}

	var structResult []string

	lm := NewLinkedSetString()
	reg := getRegistry()

	for _, structV := range structs {
		s, err := renderStruct(structV, lm, reg)
		if err != nil {
			return err
		}
		structResult = append(structResult, s)
	}

	var packageName string
	var targetPath string
	var goFileName string

	if onlyStruct {
		packageName = def.Namespace
		targetPath = dirPath
		p := filepath.Dir(dirPath)
		if strings.HasSuffix(p, string(filepath.Separator)+def.Namespace) {
			targetPath = filepath.Dir(dirPath)
		}
		goFileName = fmt.Sprintf("%s%s", serviceIdlFileName, ".go")
	} else {
		packageName = "structs"
		targetPath = filepath.Join(dirPath, "structs")
		goFileName = fmt.Sprintf("%s%s", serviceIdlFileName, outputGoFileStructSuffix)
		// 删除 structs 目录后重建
		if common.IsExist(targetPath) {
			os.RemoveAll(targetPath)
		}
	}

	// Build imports string
	var importsBuilder strings.Builder
	guard := map[string]struct{}{}
	for _, depend := range lm.getList() {
		if !strings.Contains(depend, ".") {
			continue
		}
		items := strings.Split(depend, ".")
		alias := items[0]
		if _, ok := guard[alias]; ok {
			continue
		}
		im, ok := def.GoStructImports[alias]
		if !ok || len(im) == 0 {
			return fmt.Errorf("dependency '%s' not found", alias)
		}
		if im[0].NeedAlias() {
			importsBuilder.WriteString(fmt.Sprintf("import %s \"%s\"\n", alias, im[0].PackageName))
		} else {
			importsBuilder.WriteString(fmt.Sprintf("import \"%s\"\n", im[0].PackageName))
		}
		guard[alias] = struct{}{}
	}

	// Render header using template
	headerData := &StructsHeaderData{
		PackageName: packageName,
		Imports:     importsBuilder.String(),
	}

	var headerBuilder strings.Builder
	headerTmpl, err := reg.Get(gtpl.TplStructsHeader)
	if err != nil {
		return fmt.Errorf("failed to get structs header template: %w", err)
	}
	if err := headerTmpl.Execute(&headerBuilder, headerData); err != nil {
		return fmt.Errorf("failed to execute structs header template: %w", err)
	}

	result := headerBuilder.String() + strings.Join(structResult, "\n\n")

	return save(targetPath, goFileName, result)
}

// renderStruct renders a single struct definition
func renderStruct(structV *transfer.StructDefine, lm *LinkedSetString, reg *gtpl.Registry) (string, error) {
	// Use HasFormTag to determine if form tags should be generated
	fields := converseStructFieldsWithFormTag(structV.Fields, lm, structV.HasFormTag)

	tmpl, err := reg.Get(gtpl.TplStruct)
	if err != nil {
		return "", fmt.Errorf("failed to get struct template: %w", err)
	}

	data := &StructRenderData{
		Name:         structV.Name,
		Fields:       fields,
		ExtendStruct: structV.Extends,
		ExtendPoint:  structV.Point,
	}
	if structV.Extends != "" {
		lm.add(structV.Extends)
	}

	var builder strings.Builder
	err = tmpl.Execute(&builder, data)
	if err != nil {
		return "", fmt.Errorf("failed to execute struct template for %s: %w", structV.Name, err)
	}
	return builder.String(), nil
}
