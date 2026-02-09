package java

import (
	"fmt"
	"icomplie/common"
	"icomplie/internal/transfer"
	"icomplie/pkg/codegen"
	"icomplie/pkg/codegen/java/client"
	"os"
	"path/filepath"
)

// Target constants
const (
	TargetServer = "server"
	TargetClient = "client"
	TargetAll    = "all"
)

// Generator generates Java code from IDL definitions
type Generator struct{}

// NewGenerator creates a new Java generator
func NewGenerator() *Generator {
	return &Generator{}
}

// Name returns the generator name
func (g *Generator) Name() string {
	return "java"
}

// Generate generates Java code from the definition
func (g *Generator) Generate(ctx *codegen.Context, def *transfer.Definition) (*codegen.Result, error) {
	result := &codegen.Result{
		GeneratedFiles: make([]string, 0),
		Warnings:       make([]string, 0),
	}

	// Create output directory structure
	baseDir := filepath.Join(ctx.OutputDir, def.Namespace)

	target := ctx.Target
	if target == "" {
		target = TargetAll
	}

	// Generate server code
	if target == TargetServer || target == TargetAll {
		files, err := g.generateServer(ctx, def, baseDir)
		if err != nil {
			return nil, err
		}
		result.GeneratedFiles = append(result.GeneratedFiles, files...)
	}

	// Generate client code
	if target == TargetClient || target == TargetAll {
		files, err := g.generateClient(ctx, def, baseDir)
		if err != nil {
			return nil, err
		}
		result.GeneratedFiles = append(result.GeneratedFiles, files...)
	}

	return result, nil
}

func (g *Generator) generateServer(ctx *codegen.Context, def *transfer.Definition, baseDir string) ([]string, error) {
	var files []string

	pojoDir := filepath.Join(baseDir, "model")
	controllerDir := filepath.Join(baseDir, "controller")

	if err := os.MkdirAll(pojoDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create model directory: %w", err)
	}
	if err := os.MkdirAll(controllerDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create controller directory: %w", err)
	}

	// Generate POJOs
	for _, st := range def.Structs {
		file, err := generatePOJO(pojoDir, ctx.PackagePath, def.Namespace, st)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	// Generate Controllers (skip if onlyStruct)
	if !ctx.OnlyStruct {
		for _, svc := range def.Services {
			// 备份已存在的Controller文件
			className := common.Capitalize(svc.Name) + "Controller"
			controllerFile := filepath.Join(controllerDir, className+".java")
			if err := backupFile(controllerFile); err != nil {
				return nil, fmt.Errorf("failed to backup controller file: %w", err)
			}

			file, err := generateController(controllerDir, ctx.PackagePath, def.Namespace, svc)
			if err != nil {
				return nil, err
			}
			files = append(files, file)
		}
	}

	return files, nil
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

func (g *Generator) generateClient(ctx *codegen.Context, def *transfer.Definition, baseDir string) ([]string, error) {
	if ctx.OnlyStruct {
		return nil, nil
	}

	if len(def.Services) == 0 {
		return nil, nil
	}

	if err := client.RenderClient(baseDir, ctx.PackagePath, def.Namespace, def.Services); err != nil {
		return nil, err
	}

	// Return generated file paths
	clientDir := filepath.Join(baseDir, "client")
	var files []string

	// Add HttpClient interface
	files = append(files, filepath.Join(clientDir, "HttpClient.java"))
	// Add ApacheHttpClient implementation
	files = append(files, filepath.Join(clientDir, "ApacheHttpClient.java"))

	// Add service clients
	for _, svc := range def.Services {
		files = append(files, filepath.Join(clientDir, svc.Name+"Client.java"))
	}

	return files, nil
}
