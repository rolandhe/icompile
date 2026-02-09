package template

import "embed"

//go:embed templates/java/server/*.tmpl templates/java/client/*.tmpl
var EmbeddedTemplates embed.FS
