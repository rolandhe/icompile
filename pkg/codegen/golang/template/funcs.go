package template

import (
	"icomplie/common"
	"strings"
	"text/template"
)

// DefaultFuncMap provides common template functions
var DefaultFuncMap = template.FuncMap{
	"formatVariable": common.FormatVariable,
	"capitalize":     common.Capitalize,
	"lower":          strings.ToLower,
	"upper":          strings.ToUpper,
	"join":           strings.Join,
	"gt":             gt,
	"len":            length,
}

func gt(a, b int) bool {
	return a > b
}

func length(s string) int {
	return len(s)
}
