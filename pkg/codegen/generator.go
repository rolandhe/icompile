package codegen

import (
	"icomplie/internal/transfer"
)

// Context holds the generation context
type Context struct {
	InputFile   string
	OutputDir   string
	PackagePath string
	OnlyStruct  bool
	Language    string
	Target      string
	Platform    string // browser, miniapp (for TypeScript)
}

// Result holds the generation result
type Result struct {
	GeneratedFiles []string
	Warnings       []string
}

// Generator is the interface for code generators
type Generator interface {
	// Name returns the generator name (e.g., "go", "java")
	Name() string

	// Generate generates code from the definition
	Generate(ctx *Context, def *transfer.Definition) (*Result, error)
}

// Registry holds registered generators
type Registry struct {
	generators map[string]Generator
}

// NewRegistry creates a new generator registry
func NewRegistry() *Registry {
	return &Registry{
		generators: make(map[string]Generator),
	}
}

// Register registers a generator
func (r *Registry) Register(gen Generator) {
	r.generators[gen.Name()] = gen
}

// Get returns a generator by name
func (r *Registry) Get(name string) (Generator, bool) {
	gen, ok := r.generators[name]
	return gen, ok
}

// List returns all registered generator names
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.generators))
	for name := range r.generators {
		names = append(names, name)
	}
	return names
}
