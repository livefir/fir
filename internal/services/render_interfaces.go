package services

import (
	"context"
	"html/template"

	firHttp "github.com/livefir/fir/internal/http"
	"github.com/livefir/fir/pubsub"
)

// TemplateType represents the type of template to render
type TemplateType int

const (
	// StandardTemplate is the normal content template
	StandardTemplate TemplateType = iota
	// ErrorTemplate is the error content template
	ErrorTemplate
	// EventTemplate is a template for specific event responses
	EventTemplate
)

// RenderContext contains all the information needed to render a template
type RenderContext struct {
	// Basic rendering information
	RouteID      string
	TemplateType TemplateType
	Data         interface{}
	
	// Template configuration
	TemplatePath    string
	LayoutPath      string
	PartialPaths    []string
	
	// Template engine configuration
	FuncMap          template.FuncMap
	Extensions       []string
	CacheDisabled    bool
	
	// Request context for template functions
	RequestModel     *firHttp.RequestModel
	SessionData      map[string]interface{}
	
	// Error handling
	IsError          bool
	ErrorData        map[string]interface{}
	
	// Context for cancellation and timeouts
	Context          context.Context
}

// RenderResult contains the result of template rendering
type RenderResult struct {
	// Rendered content
	HTML []byte
	
	// DOM events generated during rendering
	Events []firHttp.DOMEvent
	
	// Metadata about the rendering
	TemplateUsed    string
	RenderDuration  int64 // in milliseconds
	CacheHit        bool
}

// ErrorContext contains information for rendering error templates
type ErrorContext struct {
	// Error information
	Error        error
	StatusCode   int
	ErrorData    map[string]interface{}
	
	// Template configuration (inherits from RenderContext)
	RenderContext
}

// TemplateConfig represents the configuration for a template
type TemplateConfig struct {
	// Template paths
	ContentPath         string
	LayoutPath          string
	PartialPaths        []string
	
	// Template settings
	Extensions          []string
	FuncMap             template.FuncMap
	CacheDisabled       bool
	
	// Content names for layout templates
	LayoutContentName   string
	
	// Route identification
	RouteID             string
}

// RenderService is the main interface for template rendering operations
type RenderService interface {
	// RenderTemplate renders a template with the given context
	RenderTemplate(ctx RenderContext) (*RenderResult, error)
	
	// RenderError renders an error template
	RenderError(ctx ErrorContext) (*RenderResult, error)
	
	// RenderEvents converts PubSub events to DOM events
	RenderEvents(events []pubsub.Event, routeID string) ([]firHttp.DOMEvent, error)
}

// TemplateService manages template loading, caching, and parsing
type TemplateService interface {
	// LoadTemplate loads a template with the given configuration
	LoadTemplate(config TemplateConfig) (*template.Template, error)
	
	// ParseTemplate parses template content with partials and layout
	ParseTemplate(content, layout string, partials []string, funcMap template.FuncMap) (*template.Template, error)
	
	// GetTemplate retrieves a cached template or loads it if not cached
	GetTemplate(routeID string, templateType TemplateType) (*template.Template, error)
	
	// ClearCache clears the template cache
	ClearCache() error
	
	// SetCacheEnabled enables or disables template caching
	SetCacheEnabled(enabled bool)
}

// TemplateEngine abstracts different template engines (Go templates, etc.)
type TemplateEngine interface {
	// ParseFiles parses template files into a template
	ParseFiles(files ...string) (TemplateHandle, error)
	
	// ParseContent parses template content string into a template
	ParseContent(content string) (TemplateHandle, error)
	
	// Execute executes a template with data and writes to a buffer
	Execute(tmpl TemplateHandle, data interface{}) ([]byte, error)
	
	// AddFuncMap adds function map to the template engine
	AddFuncMap(funcMap template.FuncMap)
}

// TemplateHandle represents a handle to a parsed template
type TemplateHandle interface {
	// Execute executes the template with data
	Execute(data interface{}) ([]byte, error)
	
	// Clone creates a copy of the template
	Clone() (TemplateHandle, error)
}

// ResponseBuilder builds HTTP responses from render results
type ResponseBuilder interface {
	// BuildEventResponse builds a response from an event processing result
	BuildEventResponse(result *EventResponse, request *firHttp.RequestModel) (*firHttp.ResponseModel, error)
	
	// BuildTemplateResponse builds a response from a template render result
	BuildTemplateResponse(render *RenderResult, statusCode int) (*firHttp.ResponseModel, error)
	
	// BuildErrorResponse builds an error response
	BuildErrorResponse(err error, statusCode int) (*firHttp.ResponseModel, error)
	
	// BuildRedirectResponse builds a redirect response
	BuildRedirectResponse(url string, statusCode int) (*firHttp.ResponseModel, error)
}

// TemplateCache provides caching for parsed templates
type TemplateCache interface {
	// Get retrieves a template from cache
	Get(key string) (TemplateHandle, bool)
	
	// Set stores a template in cache
	Set(key string, template TemplateHandle)
	
	// Delete removes a template from cache
	Delete(key string)
	
	// Clear clears all cached templates
	Clear()
	
	// Stats returns cache statistics
	Stats() TemplateCacheStats
}

// TemplateCacheStats provides statistics about template cache usage
type TemplateCacheStats struct {
	Hits        int64
	Misses      int64
	Entries     int
	HitRatio    float64
}
