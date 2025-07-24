package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	firHttp "github.com/livefir/fir/internal/http"
	"github.com/livefir/fir/internal/services"
)

// TestHTTPResponseWriting_JSONEventResponse tests that JSON event responses are written correctly to HTTP
func TestHTTPResponseWriting_JSONEventResponse(t *testing.T) {
	// Create a mock response builder that returns a specific response
	mockBuilder := &mockResponseBuilder{
		buildEventResponse: &firHttp.ResponseModel{
			StatusCode: 200,
			Headers: map[string]string{
				"Content-Type":    "application/json",
				"X-Custom-Header": "test-value",
			},
			Body: []byte(`{"status": "success", "message": "Event processed"}`),
		},
	}

	// Create services and handler
	services := createMockServices()
	services.ResponseBuilder = mockBuilder
	jsonHandler := NewJSONEventHandler(
		services.EventService,
		services.RenderService,
		mockBuilder,
		nil,
	)

	// Create integration and setup chain
	chain := NewPriorityHandlerChain(&defaultHandlerLogger{}, &defaultHandlerMetrics{})
	chain.AddHandlerWithConfig(jsonHandler, HandlerConfig{
		Name:     jsonHandler.HandlerName(),
		Priority: 10,
		Enabled:  true,
	})
	integration := NewRouteHandlerIntegration(chain)

	// Create test request
	req := httptest.NewRequest("POST", "/events", strings.NewReader(`{"id": "test", "data": "value"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-FIR-MODE", "event")

	// Create response recorder
	w := httptest.NewRecorder()

	// Process the request
	handled := integration.HandleRequest(w, req)
	if !handled {
		t.Fatal("Expected integration to handle the request")
	}

	// Verify status code
	if w.Code != 200 {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}

	// Verify headers are set correctly
	if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}

	if customHeader := w.Header().Get("X-Custom-Header"); customHeader != "test-value" {
		t.Errorf("Expected X-Custom-Header 'test-value', got '%s'", customHeader)
	}

	// Verify body is written correctly
	expectedBody := `{"status": "success", "message": "Event processed"}`
	actualBody := w.Body.String()
	if actualBody != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, actualBody)
	}
}

// TestHTTPResponseWriting_TemplateResponse tests that template responses are written correctly to HTTP
func TestHTTPResponseWriting_TemplateResponse(t *testing.T) {
	// Create a mock response builder that returns a template response
	mockBuilder := &mockResponseBuilder{
		buildTemplateResponse: &firHttp.ResponseModel{
			StatusCode: 200,
			Headers: map[string]string{
				"Content-Type":  "text/html; charset=utf-8",
				"Cache-Control": "no-cache",
			},
			Body: []byte(`<html><body><h1>Test Page</h1></body></html>`),
		},
	}

	// Create services and handler
	services := createMockServices()
	services.ResponseBuilder = mockBuilder
	getHandler := NewGetHandler(
		services.RenderService,
		services.TemplateService,
		mockBuilder,
		services.EventService,
	)

	// Create integration and setup chain
	chain := NewPriorityHandlerChain(&defaultHandlerLogger{}, &defaultHandlerMetrics{})
	chain.AddHandlerWithConfig(getHandler, HandlerConfig{
		Name:     getHandler.HandlerName(),
		Priority: 50,
		Enabled:  true,
	})
	integration := NewRouteHandlerIntegration(chain)

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)

	// Create response recorder
	w := httptest.NewRecorder()

	// Process the request
	handled := integration.HandleRequest(w, req)
	if !handled {
		t.Fatal("Expected integration to handle the request")
	}

	// Verify status code
	if w.Code != 200 {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}

	// Verify headers are set correctly
	if contentType := w.Header().Get("Content-Type"); contentType != "text/html; charset=utf-8" {
		t.Errorf("Expected Content-Type 'text/html; charset=utf-8', got '%s'", contentType)
	}

	if cacheControl := w.Header().Get("Cache-Control"); cacheControl != "no-cache" {
		t.Errorf("Expected Cache-Control 'no-cache', got '%s'", cacheControl)
	}

	// Verify body is written correctly
	expectedBody := `<html><body><h1>Test Page</h1></body></html>`
	actualBody := w.Body.String()
	if actualBody != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, actualBody)
	}
}

// TestHTTPResponseWriting_ErrorResponse tests that error responses are written correctly to HTTP
func TestHTTPResponseWriting_ErrorResponse(t *testing.T) {
	// Create a mock response builder that returns an error response
	mockBuilder := &mockResponseBuilder{
		buildErrorResponse: &firHttp.ResponseModel{
			StatusCode: 400,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: []byte(`{"error": "Bad Request", "message": "Invalid request format"}`),
		},
	}

	// Create services and handler
	services := createMockServices()
	services.ResponseBuilder = mockBuilder
	jsonHandler := NewJSONEventHandler(
		services.EventService,
		services.RenderService,
		mockBuilder,
		nil,
	)

	// Create integration and setup chain
	chain := NewPriorityHandlerChain(&defaultHandlerLogger{}, &defaultHandlerMetrics{})
	chain.AddHandlerWithConfig(jsonHandler, HandlerConfig{
		Name:     jsonHandler.HandlerName(),
		Priority: 10,
		Enabled:  true,
	})
	integration := NewRouteHandlerIntegration(chain)

	// Create test request with invalid JSON to trigger error
	req := httptest.NewRequest("POST", "/events", strings.NewReader(`{invalid json`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-FIR-MODE", "event")

	// Create response recorder
	w := httptest.NewRecorder()

	// Process the request
	handled := integration.HandleRequest(w, req)
	if !handled {
		t.Fatal("Expected integration to handle the request")
	}

	// Verify status code
	if w.Code != 400 {
		t.Errorf("Expected status code 400, got %d", w.Code)
	}

	// Verify headers are set correctly
	if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}

	// Verify body is written correctly
	expectedBody := `{"error": "Bad Request", "message": "Invalid request format"}`
	actualBody := w.Body.String()
	if actualBody != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, actualBody)
	}
}

// TestHTTPResponseWriting_RedirectResponse tests that redirect responses are written correctly to HTTP
func TestHTTPResponseWriting_RedirectResponse(t *testing.T) {
	// Create a mock response builder that returns a redirect response
	mockBuilder := &mockResponseBuilder{
		buildRedirectResponse: &firHttp.ResponseModel{
			StatusCode: 302,
			Headers: map[string]string{
				"Location": "/success",
			},
			Body: []byte{}, // Redirect responses typically have no body
		},
	}

	// Create services and handler
	services := createMockServices()
	services.ResponseBuilder = mockBuilder

	// Create form handler - but we need to make sure the response builder is used for redirect
	formHandler := NewFormHandler(
		services.EventService,
		services.RenderService,
		mockBuilder, // Use the mockBuilder directly, not services.ResponseBuilder
		nil,
	)

	// Create integration and setup chain
	chain := NewPriorityHandlerChain(&defaultHandlerLogger{}, &defaultHandlerMetrics{})
	chain.AddHandlerWithConfig(formHandler, HandlerConfig{
		Name:     formHandler.HandlerName(),
		Priority: 20,
		Enabled:  true,
	})
	integration := NewRouteHandlerIntegration(chain)

	// Create test request for form submission
	req := httptest.NewRequest("POST", "/form", strings.NewReader("_redirect=/success"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Create response recorder
	w := httptest.NewRecorder()

	// Process the request
	handled := integration.HandleRequest(w, req)
	if !handled {
		t.Fatal("Expected integration to handle the request")
	}

	// Verify status code
	if w.Code != 302 {
		t.Errorf("Expected status code 302, got %d", w.Code)
	}

	// Verify Location header is set correctly
	if location := w.Header().Get("Location"); location != "/success" {
		t.Errorf("Expected Location '/success', got '%s'", location)
	}

	// Verify body is empty (redirects typically have no body)
	actualBody := w.Body.String()
	if actualBody != "" {
		t.Errorf("Expected empty body, got %q", actualBody)
	}
}

// createMockServices creates a full set of mock services for testing
func createMockServices() *mockServices {
	return &mockServices{
		EventService: &mockEventService{
			processEventResponse: &services.EventResponse{
				StatusCode: 200,
				Body:       []byte("event processed"),
			},
		},
		RenderService: &mockRenderService{
			renderTemplateResponse: &services.RenderResult{
				HTML: []byte("<html>test</html>"),
			},
		},
		TemplateService: &mockTemplateService{},
		ResponseBuilder: &mockResponseBuilder{},
	}
}

// mockServices holds all mock services for easy creation
type mockServices struct {
	EventService    services.EventService
	RenderService   services.RenderService
	TemplateService services.TemplateService
	ResponseBuilder services.ResponseBuilder
}
