package template

import "embed"

//go:embed templates/typescript/browser/*.tmpl templates/typescript/miniapp/*.tmpl
var EmbeddedTemplates embed.FS
