package client

// TypesRender holds data for rendering types template
type TypesRender struct {
	Structs []*StructRender
}

// StructRender holds data for rendering a struct
type StructRender struct {
	Name    string
	Extends string
	Fields  []*FieldRender
}

// FieldRender holds data for rendering a field
type FieldRender struct {
	Name     string
	Type     string
	Optional bool
}

// ServiceClientRender holds data for rendering service client template
type ServiceClientRender struct {
	ServiceName string
	ImportTypes string
	Methods     []*MethodRender
}

// MethodRender holds data for rendering a single method
type MethodRender struct {
	MethodName        string
	Description       string
	HTTPMethod        string
	FullPath          string
	ParamsSignature   string
	ReturnType        string
	BodyParam         string
	HasQueryParams    bool
	QueryParamsObject string
}

// EmptyRender is used for templates that don't need data
type EmptyRender struct{}
