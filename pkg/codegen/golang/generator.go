package golang

import (
	"icomplie/common"
	"icomplie/internal/transfer"
	gtpl "icomplie/pkg/codegen/golang/template"
	"path/filepath"
)

// Target constants
const (
	TargetServer = "server"
	TargetClient = "client"
	TargetAll    = "all"
)

// getRegistry returns the template registry
func getRegistry() *gtpl.Registry {
	return gtpl.GetDefaultRegistry()
}

func Main(def *transfer.Definition, outDir, srcServiceFile, pp string, onlyStruct bool, onlySwagger bool, target string) error {
	inputIdl := srcServiceFile

	// 仅生成swagger文档
	if onlySwagger {
		return generateSwaggerFile(def, inputIdl+".json")
	}

	rootDir := filepath.Join(outDir, def.Namespace)
	serviceIdlFileName := common.GetFileNameFromPath(getServiceFileName(srcServiceFile))

	// Generate server code
	if target == TargetServer || target == TargetAll {
		if err := genServer(def, rootDir, serviceIdlFileName, pp, onlyStruct); err != nil {
			return err
		}
	}

	// Generate client code
	if target == TargetClient || target == TargetAll {
		if err := genClient(def, rootDir, serviceIdlFileName); err != nil {
			return err
		}
	}

	if onlyStruct {
		return nil
	}

	// Generate swagger only for server
	if target == TargetServer || target == TargetAll {
		return generateSwaggerFile(def, inputIdl+".json")
	}

	return nil
}

func getServiceFileName(srcServiceFile string) string {
	serviceName := srcServiceFile[:len(srcServiceFile)-len(common.ServiceFileSuffix)]
	return serviceName
}
