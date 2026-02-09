package cmd

import (
	"flag"
	"fmt"
	"icomplie/common"
	"icomplie/internal/transfer"
	"icomplie/pkg/codegen"
	"icomplie/pkg/codegen/golang"
	"icomplie/pkg/codegen/java"
	"icomplie/pkg/codegen/typescript"
	"icomplie/pkg/semantic"
	"os"
)

func InitParams() *Base {
	// 定义参数选项
	filePath := flag.String("i", "", "输入文件路径")
	outputDir := flag.String("o", "", "输出目录路径")
	pkgPath := flag.String("pp", "", "structs包路径")
	onlyStruct := flag.Bool("onlyStruct", false, "仅仅输出结构体")
	onlySwagger := flag.Bool("onlySwagger", false, "仅仅输出swagger文档")
	lang := flag.String("lang", "go", "目标语言: go (默认), java, typescript")
	target := flag.String("target", "server", "生成目标: server (默认), client, all")
	platform := flag.String("platform", "browser", "平台 (仅TypeScript): browser (默认), miniapp")

	// 解析命令行参数
	flag.Parse()
	if flag.NFlag() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	return &Base{
		FilePath:    *filePath,
		OutputDir:   common.FormatDir(*outputDir),
		StructsPkg:  *pkgPath,
		OnlyStruct:  *onlyStruct,
		OnlySwagger: *onlySwagger,
		Language:    *lang,
		Target:      *target,
		Platform:    *platform,
	}
}

type Base struct {
	FilePath    string
	OutputDir   string
	StructsPkg  string
	OnlyStruct  bool
	OnlySwagger bool
	Language    string
	Target      string
	Platform    string
}

func Run(cmd *Base) error {
	def, err := transfer.ParseIdl(cmd.FilePath)
	if err != nil {
		return fmt.Errorf("failed to parse IDL file: %w", err)
	}

	// Run semantic validation
	validator := semantic.NewValidator(def)
	if validationErrors := validator.Validate(); len(validationErrors) > 0 {
		for _, verr := range validationErrors {
			fmt.Fprintf(os.Stderr, "Warning: %s\n", verr.Error())
		}
	}

	// Use Java generator if language is java
	if cmd.Language == "java" {
		ctx := &codegen.Context{
			InputFile:   cmd.FilePath,
			OutputDir:   cmd.OutputDir,
			PackagePath: cmd.StructsPkg,
			OnlyStruct:  cmd.OnlyStruct,
			Language:    "java",
			Target:      cmd.Target,
		}
		gen := java.NewGenerator()
		_, err := gen.Generate(ctx, def)
		if err != nil {
			return fmt.Errorf("failed to generate Java code: %w", err)
		}
		return nil
	}

	// Use TypeScript generator if language is typescript
	if cmd.Language == "typescript" || cmd.Language == "ts" {
		ctx := &codegen.Context{
			InputFile:   cmd.FilePath,
			OutputDir:   cmd.OutputDir,
			PackagePath: cmd.StructsPkg,
			OnlyStruct:  cmd.OnlyStruct,
			Language:    "typescript",
			Target:      cmd.Target,
			Platform:    cmd.Platform,
		}
		gen := typescript.NewGenerator()
		_, err := gen.Generate(ctx, def)
		if err != nil {
			return fmt.Errorf("failed to generate TypeScript code: %w", err)
		}
		return nil
	}

	// Default: use Go generator (existing logic)
	if err := golang.Main(def, cmd.OutputDir, cmd.FilePath, cmd.StructsPkg, cmd.OnlyStruct, cmd.OnlySwagger, cmd.Target); err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}
	return nil
}
