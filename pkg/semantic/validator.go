package semantic

import (
	"icomplie/internal/transfer"
	"icomplie/pkg/errors"
)

// Validator performs semantic validation on a Definition
type Validator struct {
	def    *transfer.Definition
	errors []error
}

// NewValidator creates a new semantic validator
func NewValidator(def *transfer.Definition) *Validator {
	return &Validator{
		def:    def,
		errors: make([]error, 0),
	}
}

// Validate performs all semantic validations and returns any errors found
func (v *Validator) Validate() []error {
	v.errors = make([]error, 0)

	// Check for duplicate definitions
	v.checkDuplicateStructs()
	v.checkDuplicateServices()
	v.checkDuplicateMethods()

	// Check for undefined references
	v.checkUndefinedReferences()

	// Check for circular inheritance
	v.checkCircularInheritance()

	// Check for URL path conflicts
	v.checkURLConflicts()

	return v.errors
}

// addError adds a validation error
func (v *Validator) addError(errType, name, message string) {
	v.errors = append(v.errors, errors.NewValidationError(errType, name, message))
}

// checkDuplicateStructs checks for duplicate struct definitions
func (v *Validator) checkDuplicateStructs() {
	seen := make(map[string]bool)
	for _, st := range v.def.Structs {
		if seen[st.Name] {
			v.addError("struct", st.Name, "duplicate struct definition")
		}
		seen[st.Name] = true
	}
}

// checkDuplicateServices checks for duplicate service definitions
func (v *Validator) checkDuplicateServices() {
	seen := make(map[string]bool)
	for _, svc := range v.def.Services {
		if seen[svc.Name] {
			v.addError("service", svc.Name, "duplicate service definition")
		}
		seen[svc.Name] = true
	}
}

// checkDuplicateMethods checks for duplicate method names within services
func (v *Validator) checkDuplicateMethods() {
	for _, svc := range v.def.Services {
		seen := make(map[string]bool)

		for _, m := range svc.Posts {
			if seen[m.Name] {
				v.addError("method", svc.Name+"."+m.Name, "duplicate method definition")
			}
			seen[m.Name] = true
		}

		for _, m := range svc.Gets {
			if seen[m.Name] {
				v.addError("method", svc.Name+"."+m.Name, "duplicate method definition")
			}
			seen[m.Name] = true
		}

		for _, m := range svc.Puts {
			if seen[m.Name] {
				v.addError("method", svc.Name+"."+m.Name, "duplicate method definition")
			}
			seen[m.Name] = true
		}
	}
}

// checkUndefinedReferences checks for references to undefined types
func (v *Validator) checkUndefinedReferences() {
	// Check struct field types
	for _, st := range v.def.Structs {
		// Check extends reference
		if st.Extends != "" {
			if _, found := v.def.GetStruct(st.Extends); found == nil {
				v.addError("reference", st.Name, "extends undefined struct: "+st.Extends)
			}
		}

		// Check field types
		for _, field := range st.Fields {
			v.checkFieldType(st.Name, field.Tp)
		}
	}

	// Check method parameter and return types
	for _, svc := range v.def.Services {
		for _, m := range svc.Posts {
			v.checkMethodTypes(svc.Name+"."+m.Name, m.Params.StructName, m.MethodReturnType)
		}
		for _, m := range svc.Gets {
			if m.Params.IsSingleStruct {
				v.checkStructReference(svc.Name+"."+m.Name, m.Params.StructName)
			}
			v.checkReturnType(svc.Name+"."+m.Name, m.MethodReturnType)
		}
		for _, m := range svc.Puts {
			v.checkMethodTypes(svc.Name+"."+m.Name, m.Params.StructName, m.MethodReturnType)
		}
	}
}

// checkFieldType checks if a field type reference is valid
func (v *Validator) checkFieldType(context string, ft *transfer.FieldType) {
	if ft == nil {
		return
	}

	if ft.IsStruct {
		v.checkStructReference(context, ft.TypeName)
	}

	if ft.ValueType != nil {
		v.checkFieldType(context, ft.ValueType)
	}
}

// checkStructReference checks if a struct reference is valid
func (v *Validator) checkStructReference(context, structName string) {
	if structName == "" {
		return
	}

	// Skip common types that are defined externally
	if isExternalType(structName) {
		return
	}

	if st, _ := v.def.GetStruct(structName); st == nil {
		v.addError("reference", context, "references undefined struct: "+structName)
	}
}

// checkMethodTypes checks method parameter and return types
func (v *Validator) checkMethodTypes(context, paramStruct string, returnType transfer.MethodReturnType) {
	v.checkStructReference(context, paramStruct)
	v.checkReturnType(context, returnType)
}

// checkReturnType checks if a return type reference is valid
func (v *Validator) checkReturnType(context string, rt transfer.MethodReturnType) {
	if rt.IsVoid {
		return
	}
	if rt.IsStruct {
		v.checkStructReference(context, rt.TypeName)
	}
}

// checkCircularInheritance checks for circular inheritance in structs
func (v *Validator) checkCircularInheritance() {
	for _, st := range v.def.Structs {
		if st.Extends == "" {
			continue
		}

		visited := make(map[string]bool)
		current := st
		for current != nil && current.Extends != "" {
			if visited[current.Name] {
				v.addError("inheritance", st.Name, "circular inheritance detected")
				break
			}
			visited[current.Name] = true
			parent, _ := v.def.GetStruct(current.Extends)
			current = parent
		}
	}
}

// checkURLConflicts checks for URL path conflicts within services
func (v *Validator) checkURLConflicts() {
	for _, svc := range v.def.Services {
		urlMethods := make(map[string]map[string]string) // url -> method -> name

		addURL := func(url, httpMethod, name string) {
			if urlMethods[url] == nil {
				urlMethods[url] = make(map[string]string)
			}
			if existing, ok := urlMethods[url][httpMethod]; ok {
				v.addError("url", svc.Name, "URL conflict: "+url+" "+httpMethod+" used by both "+existing+" and "+name)
			}
			urlMethods[url][httpMethod] = name
		}

		for _, m := range svc.Posts {
			addURL(m.Url, "POST", m.Name)
		}
		for _, m := range svc.Gets {
			addURL(m.Url, "GET", m.Name)
		}
		for _, m := range svc.Puts {
			addURL(m.Url, "PUT", m.Name)
		}
	}
}

// isExternalType checks if a type is defined externally (e.g., commons.IdRequest)
func isExternalType(typeName string) bool {
	// Types with dots are from imported packages
	for _, c := range typeName {
		if c == '.' {
			return true
		}
	}
	return false
}
