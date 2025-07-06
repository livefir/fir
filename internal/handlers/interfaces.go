package handlers

import (
	"context"

	firHttp "github.com/livefir/fir/internal/http"
)

// RequestHandler represents a handler that can process HTTP requests
type RequestHandler interface {
	// Handle processes a request and returns a response
	Handle(ctx context.Context, req *firHttp.RequestModel) (*firHttp.ResponseModel, error)
	
	// SupportsRequest determines if this handler can process the given request
	SupportsRequest(req *firHttp.RequestModel) bool
	
	// HandlerName returns a unique name for this handler (for logging/debugging)
	HandlerName() string
}

// HandlerChain manages a sequence of request handlers
type HandlerChain interface {
	// Handle processes a request through the handler chain
	Handle(ctx context.Context, req *firHttp.RequestModel) (*firHttp.ResponseModel, error)
	
	// AddHandler adds a handler to the chain
	AddHandler(handler RequestHandler)
	
	// RemoveHandler removes a handler from the chain by name
	RemoveHandler(handlerName string) bool
	
	// GetHandlers returns all handlers in the chain
	GetHandlers() []RequestHandler
	
	// ClearHandlers removes all handlers from the chain
	ClearHandlers()
}

// HandlerConfig contains configuration for handlers
type HandlerConfig struct {
	// Handler identification
	Name     string
	Priority int // Lower numbers = higher priority
	
	// Request matching configuration
	Methods     []string // HTTP methods this handler supports
	ContentType string   // Content-Type this handler supports
	PathPattern string   // Path pattern this handler supports (optional)
	
	// Handler behavior
	Enabled bool // Whether this handler is enabled
}

// HandlerMiddleware represents middleware that can wrap handlers
type HandlerMiddleware interface {
	// Wrap wraps a handler with middleware functionality
	Wrap(handler RequestHandler) RequestHandler
	
	// MiddlewareName returns the name of this middleware
	MiddlewareName() string
}

// HandlerMetrics provides metrics collection for handlers
type HandlerMetrics interface {
	// RecordRequest records a request being processed
	RecordRequest(handlerName string, method string)
	
	// RecordResponse records a response being returned
	RecordResponse(handlerName string, statusCode int, duration int64)
	
	// RecordError records an error during request processing
	RecordError(handlerName string, err error)
}

// HandlerLogger provides logging for handlers
type HandlerLogger interface {
	// LogRequest logs an incoming request
	LogRequest(handlerName string, req *firHttp.RequestModel)
	
	// LogResponse logs an outgoing response
	LogResponse(handlerName string, resp *firHttp.ResponseModel, duration int64)
	
	// LogError logs an error during request processing
	LogError(handlerName string, err error, req *firHttp.RequestModel)
	
	// LogHandlerSelection logs which handler was selected for a request
	LogHandlerSelection(selectedHandler string, req *firHttp.RequestModel)
}

// HandlerContext provides shared context for handlers
type HandlerContext struct {
	// Request information
	RequestID string
	SessionID string
	UserID    string
	
	// Handler chain information
	ChainPosition int
	TotalHandlers int
	
	// Timing information
	StartTime int64
	
	// Shared data between handlers
	SharedData map[string]interface{}
}

// ErrorHandler handles errors that occur during request processing
type ErrorHandler interface {
	// HandleError processes an error and returns an appropriate response
	HandleError(ctx context.Context, err error, req *firHttp.RequestModel) (*firHttp.ResponseModel, error)
	
	// CanHandleError determines if this error handler can process the given error
	CanHandleError(err error) bool
}

// ValidationHandler validates requests before they are processed
type ValidationHandler interface {
	// ValidateRequest validates an incoming request
	ValidateRequest(ctx context.Context, req *firHttp.RequestModel) error
	
	// GetValidationRules returns the validation rules this handler enforces
	GetValidationRules() []ValidationRule
}

// ValidationRule represents a single validation rule
type ValidationRule struct {
	Name        string
	Description string
	Required    bool
	Validator   func(req *firHttp.RequestModel) error
}

// HandlerFactory creates handlers with dependencies injected
type HandlerFactory interface {
	// CreateHandler creates a handler by name with the given configuration
	CreateHandler(name string, config HandlerConfig) (RequestHandler, error)
	
	// RegisterHandlerType registers a new handler type that can be created
	RegisterHandlerType(name string, factory func(config HandlerConfig) (RequestHandler, error))
	
	// GetAvailableHandlers returns the names of all registered handler types
	GetAvailableHandlers() []string
}
