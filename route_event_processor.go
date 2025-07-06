package fir

import (
	"context"
	"fmt"
	"net/http"

	firHttp "github.com/livefir/fir/internal/http"
	"github.com/livefir/fir/internal/services"
	"github.com/livefir/fir/pubsub"
)

// RouteEventProcessor integrates the new event service with route event handling
type RouteEventProcessor struct {
	eventService services.EventService
	route        *route
}

// NewRouteEventProcessor creates a new route event processor
func NewRouteEventProcessor(eventService services.EventService, route *route) *RouteEventProcessor {
	return &RouteEventProcessor{
		eventService: eventService,
		route:        route,
	}
}

// ProcessEvent processes an event using the new service layer and returns a response
func (p *RouteEventProcessor) ProcessEvent(ctx context.Context, event Event, r *http.Request, w http.ResponseWriter) (*services.EventResponse, error) {
	// Create adapter for request parsing
	adapter := firHttp.NewStandardHTTPAdapter(w, r, nil)
	
	// Parse request to our abstraction
	requestModel, err := adapter.ParseRequest(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}

	// Convert session ID from pointer to string
	sessionID := ""
	if event.SessionID != nil {
		sessionID = *event.SessionID
	}

	// Build event request
	eventReq := services.EventRequest{
		ID:         event.ID,
		Target:     event.Target,
		ElementKey: event.ElementKey,
		SessionID:  sessionID,
		Context:    ctx,
		Params:     convertEventParams(event),
		RequestModel: requestModel,
	}

	// Process event through service layer
	response, err := p.eventService.ProcessEvent(ctx, eventReq)
	if err != nil {
		return nil, fmt.Errorf("failed to process event: %w", err)
	}

	return response, nil
}

// convertEventParams converts the raw event params to a map
func convertEventParams(event Event) map[string]interface{} {
	params := make(map[string]interface{})
	if event.Params != nil {
		// For now, we'll store the raw params as bytes
		// The service layer can decode them as needed
		params["raw"] = event.Params
		params["is_form"] = event.IsForm
	}
	return params
}

// LegacyEventHandler wraps the existing OnEventFunc to work with the new service layer
type LegacyEventHandler struct {
	handler OnEventFunc
	route   *route
}

// NewLegacyEventHandler creates a new legacy event handler wrapper
func NewLegacyEventHandler(handler OnEventFunc, route *route) *LegacyEventHandler {
	return &LegacyEventHandler{
		handler: handler,
		route:   route,
	}
}

// ProcessEvent implements the EventProcessor interface for legacy handlers
func (h *LegacyEventHandler) ProcessEvent(ctx context.Context, req services.EventRequest) (*services.EventResponse, error) {
	// Convert session ID from string to pointer
	var sessionIDPtr *string
	if req.SessionID != "" {
		sessionIDPtr = &req.SessionID
	}

	// Convert back to legacy Event structure
	event := Event{
		ID:         req.ID,
		Target:     req.Target,
		ElementKey: req.ElementKey,
		SessionID:  sessionIDPtr,
		IsForm:     getBoolParam(req.Params, "is_form"),
		Params:     getBytesParam(req.Params, "raw"),
	}

	// For now, create a minimal HTTP request from the request model
	// In a real implementation, we might need more sophisticated conversion
	httpReq := &http.Request{
		Method: req.RequestModel.Method,
		URL:    req.RequestModel.URL,
		Header: req.RequestModel.Header,
		Body:   req.RequestModel.Body,
		Host:   req.RequestModel.Host,
		RemoteAddr: req.RequestModel.RemoteAddr,
		RequestURI: req.RequestModel.RequestURI,
	}
	httpReq = httpReq.WithContext(req.Context)

	// Create a simple response writer for capturing output
	httpResp := &simpleResponseWriter{
		statusCode: http.StatusOK,
		headers:    make(http.Header),
	}

	// Create route context
	routeCtx := RouteContext{
		event:    event,
		request:  httpReq,
		response: httpResp,
		route:    h.route,
	}

	// Call the legacy handler
	err := h.handler(routeCtx)
	
	// Convert the result to our new response format
	response := &services.EventResponse{
		StatusCode: httpResp.statusCode,
		Headers:    make(map[string]string),
		Events:     []firHttp.DOMEvent{},
		PubSubEvents: []pubsub.Event{},
	}

	// Copy headers
	for k, v := range httpResp.headers {
		if len(v) > 0 {
			response.Headers[k] = v[0]
		}
	}

	// Handle errors by converting them to appropriate response format
	if err != nil {
		response.StatusCode = http.StatusInternalServerError
		response.Errors = map[string]interface{}{
			"error": err.Error(),
		}
	}

	return response, nil
}

// simpleResponseWriter is a minimal implementation of http.ResponseWriter for testing
type simpleResponseWriter struct {
	statusCode int
	headers    http.Header
	body       []byte
}

func (w *simpleResponseWriter) Header() http.Header {
	return w.headers
}

func (w *simpleResponseWriter) Write(data []byte) (int, error) {
	w.body = append(w.body, data...)
	return len(data), nil
}

func (w *simpleResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

// Helper functions to extract parameters
func getBoolParam(params map[string]interface{}, key string) bool {
	if val, ok := params[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

func getBytesParam(params map[string]interface{}, key string) []byte {
	if val, ok := params[key]; ok {
		if b, ok := val.([]byte); ok {
			return b
		}
	}
	return nil
}
