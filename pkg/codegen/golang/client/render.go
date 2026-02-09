package client

import (
	"github.com/iancoleman/strcase"
	"icomplie/common"
	"icomplie/internal/transfer"
	gtpl "icomplie/pkg/codegen/golang/template"
	"log"
	"os"
	"path/filepath"
)

// RenderClient generates client code files
func RenderClient(targetFolder string, def *transfer.Definition, idlFileName string) error {
	render, err := Convert(def)
	if err != nil {
		return err
	}

	// Create client directory
	targetFolderClient := filepath.Join(targetFolder, def.Namespace+"_client")
	if err = os.MkdirAll(targetFolderClient, os.ModePerm); err != nil {
		log.Println(err)
		return err
	}

	reg := gtpl.GetDefaultRegistry()

	// 1. Generate http_client.go (interface)
	httpClientPath := filepath.Join(targetFolderClient, "http_client.go")
	if err := renderTemplateToFile(reg, gtpl.TplHTTPClient, render, httpClientPath); err != nil {
		return err
	}

	// 2. Generate default_http_client.go (default implementation)
	defaultClientPath := filepath.Join(targetFolderClient, "default_http_client.go")
	if err := renderTemplateToFile(reg, gtpl.TplDefaultHTTPClient, render, defaultClientPath); err != nil {
		return err
	}

	// 3. Generate service client
	serviceClientPath := filepath.Join(targetFolderClient, strcase.ToSnake(render.ServiceName)+"_client.go")
	if err := renderTemplateToFile(reg, gtpl.TplServiceClient, render, serviceClientPath); err != nil {
		return err
	}

	// 4. Copy structs file with updated package name
	if err := copyStructsFile(targetFolder, targetFolderClient, idlFileName, def.Namespace); err != nil {
		// Log but don't fail - structs file might not exist
		log.Printf("Warning: could not copy structs file: %v", err)
	}

	return nil
}

func renderTemplateToFile(reg *gtpl.Registry, templateName string, data any, outputPath string) error {
	tmpl, err := reg.Get(templateName)
	if err != nil {
		return err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	if err = tmpl.Execute(file, data); err != nil {
		return err
	}

	return common.FormatGoFile(outputPath)
}

func copyStructsFile(targetFolder, targetFolderClient, idlFileName, namespace string) error {
	structsFile := filepath.Join(targetFolderClient, idlFileName+"_structs.go")
	oldStructFile := filepath.Join(targetFolder, "structs", idlFileName+"_structs.go")

	data, err := os.ReadFile(oldStructFile)
	if err != nil {
		return err
	}

	// Replace package name
	newPackageName := "package " + namespace + "_client"
	newData := replacePackageName(data, newPackageName)

	return os.WriteFile(structsFile, newData, os.ModePerm)
}

func replacePackageName(data []byte, newPackageName string) []byte {
	// Simple replacement of "package structs" with new package name
	oldPackage := []byte("package structs")
	newPackage := []byte(newPackageName)

	result := make([]byte, 0, len(data))
	i := 0
	for i < len(data) {
		if i+len(oldPackage) <= len(data) {
			match := true
			for j := 0; j < len(oldPackage); j++ {
				if data[i+j] != oldPackage[j] {
					match = false
					break
				}
			}
			if match {
				result = append(result, newPackage...)
				i += len(oldPackage)
				continue
			}
		}
		result = append(result, data[i])
		i++
	}
	return result
}
