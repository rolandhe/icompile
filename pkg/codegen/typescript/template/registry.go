package template

import (
	"fmt"
	"sync"
	"text/template"
)

// Template name constants for browser
const (
	// Browser templates
	TplBrowserTypes         = "browser/types"
	TplBrowserHTTPClient    = "browser/http_client"
	TplBrowserAxiosClient   = "browser/axios_client"
	TplBrowserServiceClient = "browser/service_client"

	// MiniApp templates
	TplMiniAppTypes         = "miniapp/types"
	TplMiniAppHTTPClient    = "miniapp/http_client"
	TplMiniAppWxClient      = "miniapp/wx_client"
	TplMiniAppServiceClient = "miniapp/service_client"
)

// Registry manages template loading and caching
type Registry struct {
	loader TemplateLoader
	cache  map[string]*template.Template
	mu     sync.RWMutex
}

// NewRegistry creates a new template registry
func NewRegistry(loader TemplateLoader) *Registry {
	return &Registry{
		loader: loader,
		cache:  make(map[string]*template.Template),
	}
}

// Get retrieves a template by name, using cache if available
func (r *Registry) Get(name string) (*template.Template, error) {
	r.mu.RLock()
	if tmpl, ok := r.cache[name]; ok {
		r.mu.RUnlock()
		return tmpl, nil
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock
	if tmpl, ok := r.cache[name]; ok {
		return tmpl, nil
	}

	tmpl, err := r.loader.Load(name)
	if err != nil {
		return nil, fmt.Errorf("failed to load template %s: %w", name, err)
	}

	r.cache[name] = tmpl
	return tmpl, nil
}

// Exists checks if a template exists
func (r *Registry) Exists(name string) bool {
	return r.loader.Exists(name)
}

// MustGet retrieves a template by name, panics on error
func (r *Registry) MustGet(name string) *template.Template {
	tmpl, err := r.Get(name)
	if err != nil {
		panic(err)
	}
	return tmpl
}

// ClearCache clears the template cache
func (r *Registry) ClearCache() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cache = make(map[string]*template.Template)
}

// DefaultRegistry is the global default registry
var (
	defaultRegistry     *Registry
	defaultRegistryOnce sync.Once
)

// GetDefaultRegistry returns the default registry with embedded templates
func GetDefaultRegistry() *Registry {
	defaultRegistryOnce.Do(func() {
		loader := NewLoader(EmbeddedTemplates, "typescript", DefaultFuncMap)
		defaultRegistry = NewRegistry(loader)
	})
	return defaultRegistry
}
