package services

import (
	"context"
	"fmt"

	firHttp "github.com/livefir/fir/internal/http"
	"github.com/livefir/fir/pubsub"
)

// DefaultEventValidator implements EventValidator
type DefaultEventValidator struct {
	requiredFields map[string][]string // eventID -> required fields
}

// NewDefaultEventValidator creates a new DefaultEventValidator
func NewDefaultEventValidator() *DefaultEventValidator {
	return &DefaultEventValidator{
		requiredFields: make(map[string][]string),
	}
}

// ValidateEvent validates an event request
func (v *DefaultEventValidator) ValidateEvent(req EventRequest) error {
	// Basic validation
	if req.ID == "" {
		return fmt.Errorf("event ID is required")
	}

	if req.Context == nil {
		return fmt.Errorf("context is required")
	}

	if req.RequestModel == nil {
		return fmt.Errorf("request model is required")
	}

	// Validate event-specific parameters
	return v.ValidateParams(req.ID, req.Params)
}

// ValidateParams validates parameters for a specific event
func (v *DefaultEventValidator) ValidateParams(eventID string, params map[string]interface{}) error {
	requiredFields, exists := v.requiredFields[eventID]
	if !exists {
		// No specific validation rules for this event
		return nil
	}

	for _, field := range requiredFields {
		if _, exists := params[field]; !exists {
			return fmt.Errorf("required field '%s' is missing for event '%s'", field, eventID)
		}
	}

	return nil
}

// SetRequiredFields sets required fields for an event
func (v *DefaultEventValidator) SetRequiredFields(eventID string, fields []string) {
	v.requiredFields[eventID] = fields
}

// DefaultEventLogger implements EventLogger
type DefaultEventLogger struct {
	debug bool
}

// NewDefaultEventLogger creates a new DefaultEventLogger
func NewDefaultEventLogger(debug bool) *DefaultEventLogger {
	return &DefaultEventLogger{
		debug: debug,
	}
}

// LogEventStart logs the start of event processing
func (l *DefaultEventLogger) LogEventStart(ctx context.Context, req EventRequest) {
	if l.debug {
		fmt.Printf("[EVENT_START] ID: %s, SessionID: %s\n", req.ID, req.SessionID)
	}
}

// LogEventSuccess logs successful event processing
func (l *DefaultEventLogger) LogEventSuccess(ctx context.Context, req EventRequest, resp *EventResponse) {
	if l.debug {
		eventsCount := len(resp.Events)
		pubsubCount := len(resp.PubSubEvents)
		fmt.Printf("[EVENT_SUCCESS] ID: %s, Status: %d, Events: %d, PubSub: %d\n", 
			req.ID, resp.StatusCode, eventsCount, pubsubCount)
	}
}

// LogEventError logs event processing errors
func (l *DefaultEventLogger) LogEventError(ctx context.Context, req EventRequest, err error) {
	if l.debug {
		fmt.Printf("[EVENT_ERROR] ID: %s, Error: %v\n", req.ID, err)
	}
}

// RouteEventHandler adapts a legacy route event handler to the new interface
type RouteEventHandler struct {
	eventID string
	handler func(ctx RouteContext) error
}

// RouteContext represents the legacy route context interface
type RouteContext interface {
	Event() Event
	Bind(v any) error
	// Add other methods as needed
}

// Event represents the legacy event structure
type Event struct {
	ID         string
	Target     *string
	ElementKey *string
	SessionID  string
}

// NewRouteEventHandler creates a new RouteEventHandler
func NewRouteEventHandler(eventID string, handler func(ctx RouteContext) error) *RouteEventHandler {
	return &RouteEventHandler{
		eventID: eventID,
		handler: handler,
	}
}

// Handle processes the event using the legacy handler
func (h *RouteEventHandler) Handle(ctx context.Context, req EventRequest) (*EventResponse, error) {
	// Create a mock route context from the event request
	routeCtx := &mockRouteContext{
		event: Event{
			ID:         req.ID,
			Target:     req.Target,
			ElementKey: req.ElementKey,
			SessionID:  req.SessionID,
		},
		params: req.Params,
	}

	// Call the legacy handler
	err := h.handler(routeCtx)
	if err != nil {
		return nil, err
	}

	// Create a basic success response
	return &EventResponse{
		StatusCode:   200,
		Headers:      make(map[string]string),
		Events:       make([]firHttp.DOMEvent, 0),
		PubSubEvents: make([]pubsub.Event, 0),
	}, nil
}

// GetEventID returns the event ID this handler processes
func (h *RouteEventHandler) GetEventID() string {
	return h.eventID
}

// mockRouteContext implements RouteContext for legacy compatibility
type mockRouteContext struct {
	event  Event
	params map[string]interface{}
}

func (m *mockRouteContext) Event() Event {
	return m.event
}

func (m *mockRouteContext) Bind(v any) error {
	// Simple binding implementation - could be enhanced
	return nil
}

// EventResponseBuilder helps build EventResponse objects
type EventResponseBuilder struct {
	response *EventResponse
}

// NewEventResponseBuilder creates a new EventResponseBuilder
func NewEventResponseBuilder() *EventResponseBuilder {
	return &EventResponseBuilder{
		response: &EventResponse{
			StatusCode:   200,
			Headers:      make(map[string]string),
			Events:       make([]firHttp.DOMEvent, 0),
			PubSubEvents: make([]pubsub.Event, 0),
			Errors:       make(map[string]interface{}),
		},
	}
}

// WithStatus sets the response status code
func (b *EventResponseBuilder) WithStatus(code int) *EventResponseBuilder {
	b.response.StatusCode = code
	return b
}

// WithHeader adds a header to the response
func (b *EventResponseBuilder) WithHeader(key, value string) *EventResponseBuilder {
	b.response.Headers[key] = value
	return b
}

// WithBody sets the response body
func (b *EventResponseBuilder) WithBody(body []byte) *EventResponseBuilder {
	b.response.Body = body
	return b
}

// WithHTML sets the response body as HTML
func (b *EventResponseBuilder) WithHTML(html string) *EventResponseBuilder {
	b.response.Body = []byte(html)
	b.response.Headers["Content-Type"] = "text/html; charset=utf-8"
	return b
}

// WithEvent adds a DOM event to the response
func (b *EventResponseBuilder) WithEvent(event firHttp.DOMEvent) *EventResponseBuilder {
	b.response.Events = append(b.response.Events, event)
	return b
}

// WithPubSubEvent adds a PubSub event to the response
func (b *EventResponseBuilder) WithPubSubEvent(event pubsub.Event) *EventResponseBuilder {
	b.response.PubSubEvents = append(b.response.PubSubEvents, event)
	return b
}

// WithRedirect sets up a redirect response
func (b *EventResponseBuilder) WithRedirect(url string, statusCode int) *EventResponseBuilder {
	b.response.Redirect = &firHttp.RedirectInfo{
		URL:        url,
		StatusCode: statusCode,
	}
	return b
}

// WithError adds an error to the response
func (b *EventResponseBuilder) WithError(key string, value interface{}) *EventResponseBuilder {
	b.response.Errors[key] = value
	return b
}

// Build returns the constructed EventResponse
func (b *EventResponseBuilder) Build() *EventResponse {
	return b.response
}

// EventRequestBuilder helps build EventRequest objects
type EventRequestBuilder struct {
	request *EventRequest
}

// NewEventRequestBuilder creates a new EventRequestBuilder
func NewEventRequestBuilder() *EventRequestBuilder {
	return &EventRequestBuilder{
		request: &EventRequest{
			Params: make(map[string]interface{}),
		},
	}
}

// WithID sets the event ID
func (b *EventRequestBuilder) WithID(id string) *EventRequestBuilder {
	b.request.ID = id
	return b
}

// WithTarget sets the event target
func (b *EventRequestBuilder) WithTarget(target string) *EventRequestBuilder {
	b.request.Target = &target
	return b
}

// WithElementKey sets the element key
func (b *EventRequestBuilder) WithElementKey(key string) *EventRequestBuilder {
	b.request.ElementKey = &key
	return b
}

// WithSessionID sets the session ID
func (b *EventRequestBuilder) WithSessionID(sessionID string) *EventRequestBuilder {
	b.request.SessionID = sessionID
	return b
}

// WithContext sets the context
func (b *EventRequestBuilder) WithContext(ctx context.Context) *EventRequestBuilder {
	b.request.Context = ctx
	return b
}

// WithParam adds a parameter
func (b *EventRequestBuilder) WithParam(key string, value interface{}) *EventRequestBuilder {
	b.request.Params[key] = value
	return b
}

// WithRequestModel sets the request model
func (b *EventRequestBuilder) WithRequestModel(model *firHttp.RequestModel) *EventRequestBuilder {
	b.request.RequestModel = model
	return b
}

// Build returns the constructed EventRequest
func (b *EventRequestBuilder) Build() *EventRequest {
	return b.request
}
