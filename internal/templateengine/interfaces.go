package templateengine

import (
	"errors"
	"html/template"
	"io"
)

// Common errors for template engine operations
var (
	ErrNotImplemented       = errors.New("not implemented")
	ErrInvalidTemplate      = errors.New("invalid template")
	ErrInvalidEventTemplate = errors.New("invalid event template")
	ErrTemplateNotFound     = errors.New("template not found")
	ErrTemplateParseFailed  = errors.New("template parse failed")

	// Event template related errors
	ErrEventNotFound       = errors.New("event not found")
	ErrEventStateNotFound  = errors.New("event state not found")
	ErrInvalidEventID      = errors.New("invalid event ID")
	ErrInvalidEventState   = errors.New("invalid event state")
	ErrInvalidTemplateName = errors.New("invalid template name")
)

// TemplateEngine defines the interface for template parsing, rendering, and management.
// This abstraction allows for different template engine implementations while maintaining
// consistent behavior across the Fir framework.
type TemplateEngine interface {
	// Template loading and parsing
	LoadTemplate(config TemplateConfig) (Template, error)
	LoadErrorTemplate(config TemplateConfig) (Template, error)

	// Template loading with context (for function map injection)
	LoadTemplateWithContext(config TemplateConfig, ctx TemplateContext) (Template, error)
	LoadErrorTemplateWithContext(config TemplateConfig, ctx TemplateContext) (Template, error)

	// Template rendering
	Render(template Template, data interface{}, w io.Writer) error
	RenderWithContext(template Template, ctx TemplateContext, data interface{}, w io.Writer) error

	// Event template handling
	ExtractEventTemplates(template Template) (EventTemplateMap, error)
	RenderEventTemplate(template Template, eventID string, state string, data interface{}) (string, error)

	// Template caching and management
	CacheTemplate(id string, template Template)
	GetCachedTemplate(id string) (Template, bool)
	ClearCache()
}

// Template wraps template functionality and provides a consistent interface
// across different template implementations.
type Template interface {
	// Core template execution
	Execute(wr io.Writer, data interface{}) error
	ExecuteTemplate(wr io.Writer, name string, data interface{}) error

	// Template metadata
	Name() string
	Templates() []Template

	// Template manipulation
	Clone() (Template, error)
	Funcs(funcMap template.FuncMap) Template

	// Fir-specific functionality
	Lookup(name string) Template
}

// TemplateCache provides caching functionality for templates to improve performance.
type TemplateCache interface {
	// Cache operations
	Set(key string, template Template) error
	Get(key string) (Template, bool)
	Delete(key string) bool
	Clear()

	// Cache statistics and management
	Size() int
	Keys() []string
}

// EventTemplateMap represents the mapping of event IDs to their template states.
// This mirrors the existing eventTemplates type but provides a cleaner interface.
type EventTemplateMap map[string]EventTemplateState

// EventTemplateState represents the different states (ok, error, pending, done)
// for a particular event and their associated template names.
type EventTemplateState map[string]struct{}

// TemplateContext provides context information needed for template rendering,
// including route context, error information, and function maps.
type TemplateContext struct {
	// Route context for accessing request/response information
	RouteContext interface{} // Will be RouteContext from main package

	// Error context for error templates
	Errors map[string]interface{}

	// Function map for template functions
	FuncMap template.FuncMap

	// Additional context data
	Data map[string]interface{}
}

// FuncMapBuilder creates function maps for template rendering based on context.
type FuncMapBuilder interface {
	BuildFuncMap(ctx TemplateContext) template.FuncMap
}

// TemplateLoader handles the loading and parsing of template files.
type TemplateLoader interface {
	// Load templates from various sources
	LoadFromFile(path string, config TemplateConfig) (Template, error)
	LoadFromString(content string, config TemplateConfig) (Template, error)
	LoadFromBytes(content []byte, config TemplateConfig) (Template, error)

	// Load partial templates
	LoadPartials(paths []string, config TemplateConfig) ([]Template, error)
}

// TemplateValidator validates template configuration and content.
type TemplateValidator interface {
	ValidateConfig(config TemplateConfig) error
	ValidateTemplate(template Template) error
	ValidateEventTemplates(eventTemplates EventTemplateMap) error
}

// Default implementation using Go's html/template package
// Implementation moved to go_template_engine.go

// DefaultFuncMapBuilder provides the default function map building functionality
type DefaultFuncMapBuilder struct{}

// BuildFuncMap implements FuncMapBuilder interface
func (b *DefaultFuncMapBuilder) BuildFuncMap(ctx TemplateContext) template.FuncMap {
	// Return existing function map or empty map as fallback
	if ctx.FuncMap != nil {
		return ctx.FuncMap
	}
	return template.FuncMap{}
}

// FileTemplateLoader provides file-based template loading functionality
type FileTemplateLoader struct{}

// LoadFromFile implements TemplateLoader interface
func (l *FileTemplateLoader) LoadFromFile(path string, config TemplateConfig) (Template, error) {
	// Use config's ReadFile function if available, otherwise use default file reading
	var content []byte
	var err error

	if config.ReadFile != nil {
		_, fileContent, readErr := config.ReadFile(path)
		if readErr != nil {
			return nil, readErr
		}
		content = fileContent
	} else {
		// Default file reading implementation would go here
		// For now, return not implemented as this requires file system integration
		return nil, ErrNotImplemented
	}

	// Parse the content as a template
	tmpl, err := template.New(path).Funcs(config.FuncMap).Parse(string(content))
	if err != nil {
		return nil, ErrTemplateParseFailed
	}

	return NewGoTemplate(tmpl), nil
}

// LoadFromString implements TemplateLoader interface
func (l *FileTemplateLoader) LoadFromString(content string, config TemplateConfig) (Template, error) {
	// Parse the content as a template
	tmpl, err := template.New("string").Funcs(config.FuncMap).Parse(content)
	if err != nil {
		return nil, ErrTemplateParseFailed
	}

	return NewGoTemplate(tmpl), nil
}

// LoadFromBytes implements TemplateLoader interface
func (l *FileTemplateLoader) LoadFromBytes(content []byte, config TemplateConfig) (Template, error) {
	// Parse the content as a template
	tmpl, err := template.New("bytes").Funcs(config.FuncMap).Parse(string(content))
	if err != nil {
		return nil, ErrTemplateParseFailed
	}

	return NewGoTemplate(tmpl), nil
}

// LoadPartials implements TemplateLoader interface
func (l *FileTemplateLoader) LoadPartials(paths []string, config TemplateConfig) ([]Template, error) {
	if len(paths) == 0 {
		return []Template{}, nil
	}

	templates := make([]Template, 0, len(paths))

	for _, path := range paths {
		tmpl, err := l.LoadFromFile(path, config)
		if err != nil {
			return nil, err
		}
		templates = append(templates, tmpl)
	}

	return templates, nil
}

// DefaultTemplateValidator provides basic template validation
type DefaultTemplateValidator struct{}

// ValidateConfig implements TemplateValidator interface
func (v *DefaultTemplateValidator) ValidateConfig(config TemplateConfig) error {
	return config.Validate()
}

// ValidateTemplate implements TemplateValidator interface
func (v *DefaultTemplateValidator) ValidateTemplate(template Template) error {
	if template == nil {
		return ErrInvalidTemplate
	}
	return nil
}

// ValidateEventTemplates implements TemplateValidator interface
func (v *DefaultTemplateValidator) ValidateEventTemplates(eventTemplates EventTemplateMap) error {
	if eventTemplates == nil {
		return ErrInvalidEventTemplate
	}
	return nil
}
