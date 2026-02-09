package golang

import (
	"icomplie/internal/transfer"
	"icomplie/pkg/codegen/golang/client"
)

func genClient(def *transfer.Definition, rootDir, idlFileName string) error {
	return client.RenderClient(rootDir, def, idlFileName)
}
