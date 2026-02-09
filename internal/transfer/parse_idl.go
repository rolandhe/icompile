package transfer

import (
	"github.com/antlr4-go/antlr/v4"
	"icomplie/internal/parser"
	"log"
	"path/filepath"
)

func ParseIdl(fileName string) (*Definition, error) {
	absPath, err := filepath.Abs(fileName)
	if err != nil {
		return nil, err
	}
	input, err := antlr.NewFileStream(fileName)
	if err != nil {
		return nil, err
	}

	lexer := parser.NewServiceLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, 0)
	p := parser.NewServiceParser(stream)
	def := &Definition{
		IdlFilePath:     absPath,
		GoStructImports: make(map[string][]*ImportStructDefine),
	}
	p.AddErrorListener(antlr.NewDiagnosticErrorListener(true))
	p.AddErrorListener(newAcceptErrorListener(def))

	tree := p.Document()

	if def.err != nil {
		log.Printf("%v", def.err)
		return nil, def.err
	}

	antlr.ParseTreeWalkerDefault.Walk(newWalkListener(def), tree)
	if def.err != nil {
		log.Printf("%v", def.err)
		return nil, def.err
	}
	return def, nil
}
