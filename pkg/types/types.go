package types

// Type represents an IDL type
type Type interface {
	// Name returns the IDL type name
	Name() string
	// GoType returns the Go type representation
	GoType() string
	// JavaType returns the Java type representation
	JavaType() string
	// SwaggerType returns the Swagger/OpenAPI type
	SwaggerType() string
	// SwaggerFormat returns the Swagger/OpenAPI format (optional)
	SwaggerFormat() string
	// IsBasic returns true if this is a basic/primitive type
	IsBasic() bool
}

// BasicType represents a primitive IDL type
type BasicType struct {
	name          string
	goType        string
	javaType      string
	swaggerType   string
	swaggerFormat string
}

func (t *BasicType) Name() string          { return t.name }
func (t *BasicType) GoType() string        { return t.goType }
func (t *BasicType) JavaType() string      { return t.javaType }
func (t *BasicType) SwaggerType() string   { return t.swaggerType }
func (t *BasicType) SwaggerFormat() string { return t.swaggerFormat }
func (t *BasicType) IsBasic() bool         { return true }

// ListType represents a list/array type
type ListType struct {
	ElementType Type
}

func (t *ListType) Name() string          { return "list<" + t.ElementType.Name() + ">" }
func (t *ListType) GoType() string        { return "[]" + t.ElementType.GoType() }
func (t *ListType) JavaType() string      { return "List<" + t.ElementType.JavaType() + ">" }
func (t *ListType) SwaggerType() string   { return "array" }
func (t *ListType) SwaggerFormat() string { return "" }
func (t *ListType) IsBasic() bool         { return false }

// MapType represents a map type
type MapType struct {
	KeyType   Type
	ValueType Type
}

func (t *MapType) Name() string          { return "map<" + t.KeyType.Name() + "," + t.ValueType.Name() + ">" }
func (t *MapType) GoType() string        { return "map[" + t.KeyType.GoType() + "]" + t.ValueType.GoType() }
func (t *MapType) JavaType() string      { return "Map<" + t.KeyType.JavaType() + ", " + t.ValueType.JavaType() + ">" }
func (t *MapType) SwaggerType() string   { return "object" }
func (t *MapType) SwaggerFormat() string { return "" }
func (t *MapType) IsBasic() bool         { return false }

// StructType represents a struct/object type
type StructType struct {
	name string
}

func NewStructType(name string) *StructType {
	return &StructType{name: name}
}

func (t *StructType) Name() string          { return t.name }
func (t *StructType) GoType() string        { return "*" + t.name }
func (t *StructType) JavaType() string      { return t.name }
func (t *StructType) SwaggerType() string   { return "object" }
func (t *StructType) SwaggerFormat() string { return "" }
func (t *StructType) IsBasic() bool         { return false }

// VoidType represents a void return type
type VoidType struct{}

func (t *VoidType) Name() string          { return "void" }
func (t *VoidType) GoType() string        { return "*commons.Void" }
func (t *VoidType) JavaType() string      { return "Void" }
func (t *VoidType) SwaggerType() string   { return "" }
func (t *VoidType) SwaggerFormat() string { return "" }
func (t *VoidType) IsBasic() bool         { return false }
