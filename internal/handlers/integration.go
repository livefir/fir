package handlers

import (
	"net/http"

	firHttp "github.com/livefir/fir/internal/http"
	"github.com/livefir/fir/internal/routeservices"
)

// RouteHandlerIntegration provides integration between the new handler chain
// and the existing route system
type RouteHandlerIntegration struct {
	handlerChain HandlerChain
}

// NewRouteHandlerIntegration creates a new integration layer
func NewRouteHandlerIntegration(handlerChain HandlerChain) *RouteHandlerIntegration {
	return &RouteHandlerIntegration{
		handlerChain: handlerChain,
	}
}

// HandleRequest processes an HTTP request through the handler chain
func (i *RouteHandlerIntegration) HandleRequest(w http.ResponseWriter, r *http.Request) bool {
	// Convert http.Request to RequestModel
	requestModel, err := i.adaptRequest(r)
	if err != nil {
		http.Error(w, "Failed to adapt request", http.StatusInternalServerError)
		return false
	}

	// Process through handler chain
	responseModel, err := i.handlerChain.Handle(r.Context(), requestModel)
	if err != nil {
		// If no handler can process this request, return false to let legacy system handle it
		return false
	}

	// Convert ResponseModel back to http.Response
	err = i.adaptResponse(w, responseModel)
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return false
	}

	return true
}

// CanHandleRequest checks if the handler chain can handle this request without processing it
func (i *RouteHandlerIntegration) CanHandleRequest(r *http.Request) bool {
	requestModel, err := i.adaptRequest(r)
	if err != nil {
		return false
	}

	// Check if any handler in the chain supports this request
	for _, handler := range i.handlerChain.GetHandlers() {
		if handler.SupportsRequest(requestModel) {
			return true
		}
	}

	return false
}

// adaptRequest converts http.Request to RequestModel
func (i *RouteHandlerIntegration) adaptRequest(r *http.Request) (*firHttp.RequestModel, error) {
	// Use the existing request adapter or create inline conversion
	return &firHttp.RequestModel{
		Method:      r.Method,
		URL:         r.URL,
		Proto:       r.Proto,
		Header:      r.Header,
		Body:        r.Body,
		Host:        r.Host,
		RemoteAddr:  r.RemoteAddr,
		RequestURI:  r.RequestURI,
		Context:     r.Context(),
		QueryParams: r.URL.Query(),
	}, nil
}

// adaptResponse converts ResponseModel to http.Response
func (i *RouteHandlerIntegration) adaptResponse(w http.ResponseWriter, resp *firHttp.ResponseModel) error {
	// Set headers
	for key, value := range resp.Headers {
		w.Header().Set(key, value)
	}

	// Set status code
	w.WriteHeader(resp.StatusCode)

	// Write body
	if len(resp.Body) > 0 {
		_, err := w.Write(resp.Body)
		return err
	}

	return nil
}

// SetupDefaultHandlerChain creates a default handler chain with common handlers
func SetupDefaultHandlerChain(services *routeservices.RouteServices) HandlerChain {
	// Create logger and metrics implementations
	logger := &defaultHandlerLogger{}
	metrics := &defaultHandlerMetrics{}

	// Create handler chain
	chain := NewPriorityHandlerChain(logger, metrics)

	// Add handlers in priority order (lower number = higher priority)

	// 1. WebSocket handler (highest priority)
	if services.EventService != nil && services.ResponseBuilder != nil {
		wsHandler := NewWebSocketHandler(services.EventService, services.ResponseBuilder)
		chain.AddHandlerWithConfig(wsHandler, HandlerConfig{
			Name:     wsHandler.HandlerName(), // Use actual handler name
			Priority: 5,
			Enabled:  true,
		})
	}

	// 2. JSON event handler
	if services.EventService != nil && services.RenderService != nil && services.ResponseBuilder != nil {
		jsonHandler := NewJSONEventHandler(
			services.EventService,
			services.RenderService,
			services.ResponseBuilder,
			nil, // validator optional
		)
		chain.AddHandlerWithConfig(jsonHandler, HandlerConfig{
			Name:     jsonHandler.HandlerName(), // Use actual handler name
			Priority: 10,
			Enabled:  true,
		})
	}

	// 3. Form handler
	if services.EventService != nil && services.RenderService != nil && services.ResponseBuilder != nil {
		formHandler := NewFormHandler(
			services.EventService,
			services.RenderService,
			services.ResponseBuilder,
			nil, // validator optional
		)
		chain.AddHandlerWithConfig(formHandler, HandlerConfig{
			Name:     formHandler.HandlerName(), // Use actual handler name
			Priority: 20,
			Enabled:  true,
		})
	}

	// 4. POC handler - Proof of concept handler with no service dependencies
	// This ensures the chain always has at least one handler for testing
	// HIGH PRIORITY: Must come before GET handler to handle /poc requests specifically
	pocHandler := NewPOCHandler()
	chain.AddHandlerWithConfig(pocHandler, HandlerConfig{
		Name:     pocHandler.HandlerName(),
		Priority: 40, // Higher priority than GET handler to handle /poc specifically
		Enabled:  true,
	})

	// 5. GET handler (lowest priority) - Re-enabled after session management fixes
	if services.RenderService != nil && services.TemplateService != nil && services.ResponseBuilder != nil {
		getHandler := NewGetHandler(
			services.RenderService,
			services.TemplateService,
			services.ResponseBuilder,
			services.EventService, // Added for onLoad support
		)
		chain.AddHandlerWithConfig(getHandler, HandlerConfig{
			Name:     getHandler.HandlerName(), // Use actual handler name
			Priority: 50,
			Enabled:  true,
		})
	}

	return chain
}

// Default implementations for logger and metrics

type defaultHandlerLogger struct{}

func (l *defaultHandlerLogger) LogRequest(handlerName string, req *firHttp.RequestModel) {
	// Could integrate with existing Fir logging
}

func (l *defaultHandlerLogger) LogResponse(handlerName string, resp *firHttp.ResponseModel, duration int64) {
	// Could integrate with existing Fir logging
}

func (l *defaultHandlerLogger) LogError(handlerName string, err error, req *firHttp.RequestModel) {
	// Could integrate with existing Fir logging
}

func (l *defaultHandlerLogger) LogHandlerSelection(selectedHandler string, req *firHttp.RequestModel) {
	// Could integrate with existing Fir logging
}

type defaultHandlerMetrics struct{}

func (m *defaultHandlerMetrics) RecordRequest(handlerName string, method string) {
	// Could integrate with existing Fir metrics
}

func (m *defaultHandlerMetrics) RecordResponse(handlerName string, statusCode int, duration int64) {
	// Could integrate with existing Fir metrics
}

func (m *defaultHandlerMetrics) RecordError(handlerName string, err error) {
	// Could integrate with existing Fir metrics
}
