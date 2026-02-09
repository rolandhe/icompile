package template

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"text/template"
)

// TemplateLoader defines the interface for loading templates
type TemplateLoader interface {
	Load(name string) (*template.Template, error)
	Exists(name string) bool
}

// Loader implements TemplateLoader using embed.FS
type Loader struct {
	fs       embed.FS
	basePath string
	funcMap  template.FuncMap
}

// NewLoader creates a new template loader
func NewLoader(fs embed.FS, language string, funcMap template.FuncMap) *Loader {
	return &Loader{
		fs:       fs,
		basePath: filepath.Join("templates", language),
		funcMap:  funcMap,
	}
}

// Load loads a template by name
func (l *Loader) Load(name string) (*template.Template, error) {
	path := filepath.Join(l.basePath, name+".tmpl")

	content, err := l.fs.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read template %s: %w", name, err)
	}

	tmpl, err := template.New(name).Funcs(l.funcMap).Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
	}

	return tmpl, nil
}

// Exists checks if a template exists
func (l *Loader) Exists(name string) bool {
	path := filepath.Join(l.basePath, name+".tmpl")
	_, err := fs.Stat(l.fs, path)
	return err == nil
}
