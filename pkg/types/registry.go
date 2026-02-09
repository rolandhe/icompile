package types

import (
	"fmt"
	"sync"
)

// Registry is a type registry that manages type mappings
type Registry struct {
	mu    sync.RWMutex
	types map[string]Type
}

// DefaultRegistry is the global default type registry
var DefaultRegistry = NewRegistry()

// NewRegistry creates a new type registry with built-in types
func NewRegistry() *Registry {
	r := &Registry{
		types: make(map[string]Type),
	}
	r.registerBuiltinTypes()
	return r
}

// registerBuiltinTypes registers all built-in IDL types
func (r *Registry) registerBuiltinTypes() {
	// Boolean
	r.Register(&BasicType{
		name:        "bool",
		goType:      "bool",
		javaType:    "Boolean",
		swaggerType: "boolean",
	})

	// Byte
	r.Register(&BasicType{
		name:        "byte",
		goType:      "byte",
		javaType:    "Byte",
		swaggerType: "integer",
	})

	// i8 (signed 8-bit integer)
	r.Register(&BasicType{
		name:        "i8",
		goType:      "int8",
		javaType:    "Byte",
		swaggerType: "integer",
	})

	// i16 (signed 16-bit integer)
	r.Register(&BasicType{
		name:        "i16",
		goType:      "int16",
		javaType:    "Short",
		swaggerType: "integer",
	})

	// i32 (signed 32-bit integer)
	r.Register(&BasicType{
		name:        "i32",
		goType:      "int32",
		javaType:    "Integer",
		swaggerType: "integer",
	})

	// i64 (signed 64-bit integer)
	r.Register(&BasicType{
		name:          "i64",
		goType:        "int64",
		javaType:      "Long",
		swaggerType:   "integer",
		swaggerFormat: "int64",
	})

	// float (32-bit floating point)
	r.Register(&BasicType{
		name:          "float",
		goType:        "float32",
		javaType:      "Float",
		swaggerType:   "number",
		swaggerFormat: "float",
	})

	// double (64-bit floating point)
	r.Register(&BasicType{
		name:          "double",
		goType:        "float64",
		javaType:      "Double",
		swaggerType:   "number",
		swaggerFormat: "double",
	})

	// string
	r.Register(&BasicType{
		name:        "string",
		goType:      "string",
		javaType:    "String",
		swaggerType: "string",
	})
}

// Register registers a type in the registry
func (r *Registry) Register(t Type) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.types[t.Name()] = t
}

// Get retrieves a type by name
func (r *Registry) Get(name string) (Type, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.types[name]
	return t, ok
}

// GetOrCreate retrieves a type by name, or creates a struct type if not found
func (r *Registry) GetOrCreate(name string) Type {
	if t, ok := r.Get(name); ok {
		return t
	}
	return NewStructType(name)
}

// MustGet retrieves a type by name, panics if not found
func (r *Registry) MustGet(name string) Type {
	t, ok := r.Get(name)
	if !ok {
		panic(fmt.Sprintf("type %s not found in registry", name))
	}
	return t
}

// IsBasicType checks if a type name is a basic type
func (r *Registry) IsBasicType(name string) bool {
	t, ok := r.Get(name)
	if !ok {
		return false
	}
	return t.IsBasic()
}

// CreateListType creates a list type with the given element type
func (r *Registry) CreateListType(elementType Type) *ListType {
	return &ListType{ElementType: elementType}
}

// CreateMapType creates a map type with the given key and value types
func (r *Registry) CreateMapType(keyType, valueType Type) *MapType {
	return &MapType{KeyType: keyType, ValueType: valueType}
}

// Global convenience functions using DefaultRegistry

// GetType retrieves a type from the default registry
func GetType(name string) (Type, bool) {
	return DefaultRegistry.Get(name)
}

// GetOrCreateType retrieves or creates a type from the default registry
func GetOrCreateType(name string) Type {
	return DefaultRegistry.GetOrCreate(name)
}

// IsBasic checks if a type name is a basic type in the default registry
func IsBasic(name string) bool {
	return DefaultRegistry.IsBasicType(name)
}

// ToGoType converts an IDL type name to Go type
func ToGoType(name string) string {
	t := DefaultRegistry.GetOrCreate(name)
	return t.GoType()
}

// ToJavaType converts an IDL type name to Java type
func ToJavaType(name string) string {
	t := DefaultRegistry.GetOrCreate(name)
	return t.JavaType()
}

// ToSwaggerType converts an IDL type name to Swagger type
func ToSwaggerType(name string) (string, string) {
	t := DefaultRegistry.GetOrCreate(name)
	return t.SwaggerType(), t.SwaggerFormat()
}
