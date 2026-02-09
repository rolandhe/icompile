package client

// ServiceClientRender holds data for rendering service client template
type ServiceClientRender struct {
	PackagePath string
	Namespace   string
	ServiceName string
	Methods     []*MethodRender
}

// MethodRender holds data for rendering a single method
type MethodRender struct {
	MethodName      string
	Description     string
	HTTPMethod      string
	FullPath        string
	ParamsSignature string
	ReturnType      string
	ResultClass     string
	BodyParam       string
	HasQueryParams  bool
	QueryParams     []*QueryParam
}

// QueryParam represents a query parameter
type QueryParam struct {
	Name    string
	VarName string
}

// POJORender holds data for rendering POJO template
type POJORender struct {
	PackagePath string
	Namespace   string
	ClassName   string
	Extends     string
	HasExtends  bool
	HasList     bool
	HasMap      bool
	Fields      []*FieldRender
}

// FieldRender holds data for rendering a field
type FieldRender struct {
	Type string
	Name string
}

// ControllerRender holds data for rendering controller template
type ControllerRender struct {
	PackagePath    string
	Namespace      string
	ClassName      string
	RootUrl        string
	HasRequestBody bool
	HasRequestParam bool
	HasList        bool
	Methods        []*ControllerMethodRender
}

// ControllerMethodRender holds data for rendering a controller method
type ControllerMethodRender struct {
	Content string
}

// ControllerMethodData holds data for rendering controller_method template
type ControllerMethodData struct {
	HTTPMethod      string
	Url             string
	ReturnType      string
	MethodName      string
	ParamsSignature string
}

// HTTPClientRender holds data for rendering http_client template
type HTTPClientRender struct {
	PackagePath string
	Namespace   string
}
