package template

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"
	"text/template"
)

// TemplateLoader loads templates from embedded filesystem
type TemplateLoader interface {
	// Load loads a single template by name
	Load(name string) (*template.Template, error)
	// Exists checks if a template exists
	Exists(name string) bool
}

// embeddedLoader loads templates from embedded filesystem
type embeddedLoader struct {
	fs       embed.FS
	basePath string
	funcMap  template.FuncMap
}

// NewLoader creates a new template loader for embedded templates
func NewLoader(embeddedFS embed.FS, language string, funcMap template.FuncMap) TemplateLoader {
	if funcMap == nil {
		funcMap = DefaultFuncMap
	}
	if language == "" {
		language = "go"
	}

	return &embeddedLoader{
		fs:       embeddedFS,
		basePath: "templates/" + language,
		funcMap:  funcMap,
	}
}

// Load loads a template by name from embedded filesystem
func (l *embeddedLoader) Load(name string) (*template.Template, error) {
	path := l.basePath + "/" + name + ".tmpl"
	content, err := l.fs.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded template %s: %w", name, err)
	}

	tmpl, err := template.New(name).Funcs(l.funcMap).Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
	}

	return tmpl, nil
}

// Exists checks if a template exists in embedded filesystem
func (l *embeddedLoader) Exists(name string) bool {
	path := l.basePath + "/" + name + ".tmpl"
	_, err := l.fs.ReadFile(path)
	return err == nil
}

// ListEmbeddedTemplates lists all available templates from embedded filesystem
func ListEmbeddedTemplates(embeddedFS embed.FS, language string) ([]string, error) {
	var templates []string
	basePath := "templates/" + language

	err := fs.WalkDir(embeddedFS, basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".tmpl") {
			// Remove base path and .tmpl suffix
			name := strings.TrimPrefix(path, basePath+"/")
			name = strings.TrimSuffix(name, ".tmpl")
			templates = append(templates, name)
		}
		return nil
	})

	return templates, err
}
