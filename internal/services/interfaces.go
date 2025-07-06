package services

import (
	"context"

	firHttp "github.com/livefir/fir/internal/http"
	"github.com/livefir/fir/pubsub"
)

// EventRequest represents a request to process an event
type EventRequest struct {
	// Event information
	ID         string
	Target     *string
	ElementKey *string
	SessionID  string

	// Request context and data
	Context context.Context
	Params  map[string]interface{}

	// Request metadata
	RequestModel *firHttp.RequestModel
}

// EventResponse represents the result of event processing
type EventResponse struct {
	// Response data
	StatusCode int
	Headers    map[string]string
	Body       []byte

	// DOM events to send to client
	Events []firHttp.DOMEvent

	// Redirect information
	Redirect *firHttp.RedirectInfo

	// PubSub events to publish
	PubSubEvents []pubsub.Event

	// Error information
	Errors map[string]interface{}
}

// EventProcessor processes individual events
type EventProcessor interface {
	ProcessEvent(ctx context.Context, req EventRequest) (*EventResponse, error)
}

// EventRegistry manages event handlers
type EventRegistry interface {
	GetHandler(eventID string) (EventHandler, bool)
	RegisterHandler(eventID string, handler EventHandler) error
	ListHandlers() []string
}

// EventHandler represents a handler for a specific event
type EventHandler interface {
	Handle(ctx context.Context, req EventRequest) (*EventResponse, error)
	GetEventID() string
}

// EventValidator validates event requests
type EventValidator interface {
	ValidateEvent(req EventRequest) error
	ValidateParams(eventID string, params map[string]interface{}) error
}

// EventTransformer transforms event data
type EventTransformer interface {
	TransformRequest(req *EventRequest) error
	TransformResponse(resp *EventResponse) error
}

// EventLogger logs event processing
type EventLogger interface {
	LogEventStart(ctx context.Context, req EventRequest)
	LogEventSuccess(ctx context.Context, req EventRequest, resp *EventResponse)
	LogEventError(ctx context.Context, req EventRequest, err error)
}

// EventPublisher publishes events to subscribers
type EventPublisher interface {
	PublishEvent(event pubsub.Event) error
	PublishEvents(events []pubsub.Event) error
}

// EventService orchestrates event processing
type EventService interface {
	ProcessEvent(ctx context.Context, req EventRequest) (*EventResponse, error)
	RegisterHandler(eventID string, handler EventHandler) error
	GetEventMetrics() EventMetrics
}

// EventMetrics contains event processing metrics
type EventMetrics struct {
	TotalEvents    int64
	SuccessfulEvents int64
	FailedEvents   int64
	AverageLatency float64
}

// EventError represents an event processing error
type EventError struct {
	EventID string
	Code    string
	Message string
	Details map[string]interface{}
	Cause   error
}

func (e *EventError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

// NewEventError creates a new EventError
func NewEventError(eventID, code, message string) *EventError {
	return &EventError{
		EventID: eventID,
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// WithCause adds a cause to the EventError
func (e *EventError) WithCause(cause error) *EventError {
	e.Cause = cause
	return e
}

// WithDetail adds a detail to the EventError
func (e *EventError) WithDetail(key string, value interface{}) *EventError {
	e.Details[key] = value
	return e
}
