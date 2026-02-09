package typescript

import (
	"icomplie/internal/transfer"
	"icomplie/pkg/codegen"
	"icomplie/pkg/codegen/typescript/client"
)

// Platform constants
const (
	PlatformBrowser = "browser"
	PlatformMiniApp = "miniapp"
)

// Generator generates TypeScript code from IDL definitions
type Generator struct{}

// NewGenerator creates a new TypeScript generator
func NewGenerator() *Generator {
	return &Generator{}
}

// Name returns the generator name
func (g *Generator) Name() string {
	return "typescript"
}

// Generate generates TypeScript code from the definition
func (g *Generator) Generate(ctx *codegen.Context, def *transfer.Definition) (*codegen.Result, error) {
	result := &codegen.Result{
		GeneratedFiles: make([]string, 0),
		Warnings:       make([]string, 0),
	}

	// Determine platform (default to browser)
	platform := ctx.Platform
	if platform == "" {
		platform = PlatformBrowser
	}

	// Generate client code
	if err := client.RenderClient(ctx.OutputDir, def.Namespace, def, platform); err != nil {
		return nil, err
	}

	// Add generated files to result
	result.GeneratedFiles = append(result.GeneratedFiles,
		"types.ts",
		"httpClient.ts",
	)

	if platform == PlatformMiniApp {
		result.GeneratedFiles = append(result.GeneratedFiles, "wxClient.ts")
	} else {
		result.GeneratedFiles = append(result.GeneratedFiles, "axiosClient.ts")
	}

	for _, svc := range def.Services {
		result.GeneratedFiles = append(result.GeneratedFiles, svc.Name+"Client.ts")
	}

	return result, nil
}
