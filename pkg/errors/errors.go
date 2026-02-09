package errors

import (
	"errors"
	"fmt"
)

// CompileError represents a compilation error with location information
type CompileError struct {
	File    string
	Line    int
	Column  int
	Message string
}

func (e *CompileError) Error() string {
	if e.File != "" {
		return fmt.Sprintf("%s:%d:%d: %s", e.File, e.Line, e.Column, e.Message)
	}
	if e.Line > 0 {
		return fmt.Sprintf("line %d:%d: %s", e.Line, e.Column, e.Message)
	}
	return e.Message
}

// NewCompileError creates a new compile error
func NewCompileError(file string, line, column int, message string) *CompileError {
	return &CompileError{
		File:    file,
		Line:    line,
		Column:  column,
		Message: message,
	}
}

// ParseError represents a parsing error
type ParseError struct {
	File    string
	Message string
	Cause   error
}

func (e *ParseError) Error() string {
	if e.File != "" {
		return fmt.Sprintf("parse error in %s: %s", e.File, e.Message)
	}
	return fmt.Sprintf("parse error: %s", e.Message)
}

func (e *ParseError) Unwrap() error {
	return e.Cause
}

// NewParseError creates a new parse error
func NewParseError(file, message string, cause error) *ParseError {
	return &ParseError{
		File:    file,
		Message: message,
		Cause:   cause,
	}
}

// GenerateError represents a code generation error
type GenerateError struct {
	Phase   string
	Message string
	Cause   error
}

func (e *GenerateError) Error() string {
	if e.Phase != "" {
		return fmt.Sprintf("generate error [%s]: %s", e.Phase, e.Message)
	}
	return fmt.Sprintf("generate error: %s", e.Message)
}

func (e *GenerateError) Unwrap() error {
	return e.Cause
}

// NewGenerateError creates a new generate error
func NewGenerateError(phase, message string, cause error) *GenerateError {
	return &GenerateError{
		Phase:   phase,
		Message: message,
		Cause:   cause,
	}
}

// ValidationError represents a semantic validation error
type ValidationError struct {
	Type    string
	Name    string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Name != "" {
		return fmt.Sprintf("validation error [%s '%s']: %s", e.Type, e.Name, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(typ, name, message string) *ValidationError {
	return &ValidationError{
		Type:    typ,
		Name:    name,
		Message: message,
	}
}

// Common error variables
var (
	ErrDuplicateImport    = errors.New("duplicate import path")
	ErrDuplicateStruct    = errors.New("duplicate struct definition")
	ErrDuplicateService   = errors.New("duplicate service definition")
	ErrDuplicateMethod    = errors.New("duplicate method name")
	ErrUndefinedType      = errors.New("undefined type reference")
	ErrCyclicDependency   = errors.New("cyclic dependency detected")
	ErrInvalidMethod      = errors.New("invalid method definition")
	ErrMissingImport      = errors.New("missing import for type")
	ErrInvalidFieldType   = errors.New("invalid field type")
	ErrInvalidReturnType  = errors.New("invalid return type")
	ErrFileOperation      = errors.New("file operation failed")
	ErrTemplateExecution  = errors.New("template execution failed")
)

// Wrap wraps an error with additional context
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// Wrapf wraps an error with formatted message
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}

// Is reports whether any error in err's chain matches target
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's chain that matches target
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// New creates a new error with the given message
func New(message string) error {
	return errors.New(message)
}

// Errorf creates a new formatted error
func Errorf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}
