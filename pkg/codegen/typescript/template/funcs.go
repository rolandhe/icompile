package template

import (
	"strings"
	"text/template"
)

// DefaultFuncMap provides common template functions for TypeScript code generation
var DefaultFuncMap = template.FuncMap{
	"lower":      strings.ToLower,
	"upper":      strings.ToUpper,
	"join":       strings.Join,
	"trimSuffix": strings.TrimSuffix,
	"trimPrefix": strings.TrimPrefix,
	"contains":   strings.Contains,
	"hasPrefix":  strings.HasPrefix,
	"hasSuffix":  strings.HasSuffix,
}
