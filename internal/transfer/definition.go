package transfer

import (
	"fmt"
	"strings"
)

// HTTPMethod represents an HTTP method type
type HTTPMethod string

const (
	HTTPMethodGet  HTTPMethod = "GET"
	HTTPMethodPost HTTPMethod = "POST"
	HTTPMethodPut  HTTPMethod = "PUT"
)

type Definition struct {
	IdlFilePath     string
	Namespace       string
	GoStructImports map[string][]*ImportStructDefine
	Structs         []*StructDefine
	Services        []*ServiceDefine
	err             error
}

func (def *Definition) acceptImport(is *ImportStructDefine) error {
	has := def.GoStructImports[is.Alias]
	for _, iidl := range has {
		if iidl.IdlPath == is.IdlPath {
			return fmt.Errorf("duplicate import path: %s", is.IdlPath)
		}
	}
	has = append(has, is)
	def.GoStructImports[is.Alias] = has
	return nil
}

func (def *Definition) GetStruct(name string) (*StructDefine, *Definition) {
	if !strings.Contains(name, ".") {
		for _, st := range def.Structs {
			if st.Name == name {
				return st, def
			}
		}
		return nil, nil
	}
	items := strings.Split(name, ".")
	importStructs := def.GoStructImports[items[0]]
	for _, is := range importStructs {
		for _, st := range is.Def.Structs {
			if st.Name == items[1] {
				return st, is.Def
			}
		}
	}
	return nil, nil
}

type ImportStructDefine struct {
	Alias       string
	PackageName string
	IdlPath     string
	Def         *Definition
}

func (st *ImportStructDefine) NeedAlias() bool {
	if st.Alias == st.PackageName {
		return false
	}
	if strings.HasSuffix(st.PackageName, "/"+st.Alias) {
		return false
	}
	return true
}

func (is *ImportStructDefine) parseIdl() error {
	var err error
	is.Def, err = ParseIdl(is.IdlPath)
	if err != nil {
		return fmt.Errorf("failed to parse imported IDL %s: %w", is.IdlPath, err)
	}
	return nil
}

func (def *Definition) addStruct(st *StructDefine) {
	def.Structs = append(def.Structs, st)
}

func (def *Definition) addService(svc *ServiceDefine) {
	def.Services = append(def.Services, svc)
}

type StructDefine struct {
	Name       string
	Fields     []*Field
	Extends    string
	Point      bool
	HasFormTag bool // New: indicates if struct should generate form tags
}

func (st *StructDefine) addField(field *Field) {
	st.Fields = append(st.Fields, field)
}

type Field struct {
	ReqDefine   string
	Tp          *FieldType
	Name        string
	Annotations []*Annotation
}

func (f *Field) addAnnotation(anno *Annotation) {
	f.Annotations = append(f.Annotations, anno)
}

type FieldType struct {
	TypeName string
	IsStruct bool
	IsBasic  bool
	IsList   bool
	IsMap    bool

	ValueType *FieldType
}

type Annotation struct {
	Key   string
	Value string
}

type ServiceDefine struct {
	Name    string
	RootUrl string
	Methods []*Method // Unified method list preserving order
	Posts   []*PostMethod
	Gets    []*GetMethod
	Puts    []*PutMethod // New: PUT methods
}

func (svc *ServiceDefine) addGet(g *GetMethod) {
	svc.Gets = append(svc.Gets, g)
	svc.Methods = append(svc.Methods, &Method{
		HTTPMethod: HTTPMethodGet,
		BaseMethod: g.BaseMethod,
		GetParams:  &g.Params,
	})
}

func (svc *ServiceDefine) addPost(p *PostMethod) {
	svc.Posts = append(svc.Posts, p)
	svc.Methods = append(svc.Methods, &Method{
		HTTPMethod: HTTPMethodPost,
		BaseMethod: p.BaseMethod,
		PostParams: &p.Params,
	})
}

func (svc *ServiceDefine) addPut(p *PutMethod) {
	svc.Puts = append(svc.Puts, p)
	svc.Methods = append(svc.Methods, &Method{
		HTTPMethod: HTTPMethodPut,
		BaseMethod: p.BaseMethod,
		PostParams: &p.Params, // PUT uses same params as POST
	})
}

// Method is a unified method representation that preserves order
type Method struct {
	HTTPMethod HTTPMethod
	BaseMethod
	PostParams *PostParam
	GetParams  *GetParam
}

type BaseMethod struct {
	Name        string
	Url         string
	NotLogin    bool
	Description string
	Products    string
	MethodReturnType
}

type PostMethod struct {
	BaseMethod
	Params PostParam
}

// PutMethod represents a PUT HTTP method (same structure as POST)
type PutMethod struct {
	BaseMethod
	Params PostParam
}

type PostParam struct {
	IsEmpty    bool
	IsList     bool
	StructName string
	ParamName  string
}

type MethodReturnType struct {
	IsList   bool
	IsStruct bool
	IsVoid   bool
	TypeName string
	IsPager  bool
}

type GetMethod struct {
	BaseMethod
	Params GetParam
}
type GetParam struct {
	IsEmpty         bool
	IsSingleStruct  bool
	StructName      string
	StructParamName string

	BasicParams []*BasicGetParam
}

func (gp *GetParam) addBasicParams(bp *BasicGetParam) {
	gp.BasicParams = append(gp.BasicParams, bp)
}

type BasicGetParam struct {
	ReqDefine   string
	IsList      bool
	TypeName    string
	ParamName   string
	Annotations []*Annotation
}

func (bp *BasicGetParam) addAnnotation(anno *Annotation) {
	bp.Annotations = append(bp.Annotations, anno)
}

// packGetBasicParamsToStruct converts a GET method's multiple basic params into a StructDefine,
// and modifies the GetParam to reference that struct (IsSingleStruct=true).
// Returns the generated StructDefine to be added to Definition.Structs.
func packGetBasicParamsToStruct(methodName string, getParam *GetParam) *StructDefine {
	structName := Capitalize(methodName) + "Req"
	st := &StructDefine{
		Name:       structName,
		HasFormTag: true,
	}
	for _, bp := range getParam.BasicParams {
		ft := &FieldType{
			IsBasic:  true,
			TypeName: bp.TypeName,
		}
		if bp.IsList {
			ft = &FieldType{
				IsList: true,
				ValueType: &FieldType{
					IsBasic:  true,
					TypeName: bp.TypeName,
				},
			}
		}
		field := &Field{
			ReqDefine:   bp.ReqDefine,
			Tp:          ft,
			Name:        bp.ParamName,
			Annotations: bp.Annotations,
		}
		st.addField(field)
	}

	// Modify GetParam to be a single struct reference
	// Keep BasicParams for client generators that need per-field query param info
	getParam.IsSingleStruct = true
	getParam.StructName = structName
	getParam.StructParamName = methodName

	return st
}

// Capitalize converts first letter to upper case and handles underscore-separated words
func Capitalize(s string) string {
	if s == "" {
		return s
	}
	words := strings.Split(s, "_")
	for i := range words {
		if len(words[i]) > 0 {
			words[i] = strings.ToUpper(words[i][:1]) + words[i][1:]
		}
	}
	return strings.Join(words, "")
}
