package handlers

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	firHttp "github.com/livefir/fir/internal/http"
	"github.com/livefir/fir/internal/services"
)

// selectiveMockResponseBuilder allows different behaviors for different response types
type selectiveMockResponseBuilder struct {
	errorResponse *firHttp.ResponseModel
	templateError error
}

func (m *selectiveMockResponseBuilder) BuildEventResponse(result *services.EventResponse, request *firHttp.RequestModel) (*firHttp.ResponseModel, error) {
	return nil, errors.New("not implemented")
}

func (m *selectiveMockResponseBuilder) BuildTemplateResponse(render *services.RenderResult, statusCode int) (*firHttp.ResponseModel, error) {
	return nil, m.templateError
}

func (m *selectiveMockResponseBuilder) BuildErrorResponse(err error, statusCode int) (*firHttp.ResponseModel, error) {
	return m.errorResponse, nil
}

func (m *selectiveMockResponseBuilder) BuildRedirectResponse(url string, statusCode int) (*firHttp.ResponseModel, error) {
	return nil, errors.New("not implemented")
}

// TestErrorResponseHandling_MalformedJSON tests 400 Bad Request for malformed JSON
func TestErrorResponseHandling_MalformedJSON(t *testing.T) {
	// Create mock response builder that returns proper error response
	mockBuilder := &mockResponseBuilder{
		buildErrorResponse: &firHttp.ResponseModel{
			StatusCode: 400,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: []byte(`{"error": "Bad Request", "message": "Malformed JSON"}`),
		},
		buildError: nil, // No error from response builder
	}

	// Create services
	services := createMockServices()
	services.ResponseBuilder = mockBuilder

	// Create JSON handler
	jsonHandler := NewJSONEventHandler(
		services.EventService,
		services.RenderService,
		mockBuilder, // Use the specific mock builder
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

	// Create test request with malformed JSON
	req := httptest.NewRequest("POST", "/events", strings.NewReader(`{invalid json`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-FIR-MODE", "event")

	// Create response recorder
	w := httptest.NewRecorder()

	// Process the request
	handled := integration.HandleRequest(w, req)
	if !handled {
		t.Fatal("Expected integration to handle the request (should return error response)")
	}

	// Verify 400 Bad Request status
	if w.Code != 400 {
		t.Errorf("Expected status code 400 for malformed JSON, got %d", w.Code)
	}

	// Verify error response has proper content type
	if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json' for error response, got '%s'", contentType)
	}

	// Verify response body contains error information
	body := w.Body.String()
	if body == "" {
		t.Error("Expected error response to have body")
	}
	t.Logf("Error response body: %s", body)
}

// TestErrorResponseHandling_EventProcessingFailure tests 500 Internal Server Error for processing failures
func TestErrorResponseHandling_EventProcessingFailure(t *testing.T) {
	// Create mock event service that returns error
	mockEventService := &mockEventService{
		processEventError: errors.New("event processing failed"),
	}

	// Create mock response builder that returns 500 error response
	mockBuilder := &mockResponseBuilder{
		buildErrorResponse: &firHttp.ResponseModel{
			StatusCode: 500,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: []byte(`{"error": "Internal Server Error", "message": "Event processing failed"}`),
		},
		buildError: nil, // No error from response builder
	}

	// Create JSON handler with failing event service
	jsonHandler := NewJSONEventHandler(
		mockEventService,
		&mockRenderService{},
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

	// Create test request with valid JSON
	req := httptest.NewRequest("POST", "/events", strings.NewReader(`{"id": "test", "data": "value"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-FIR-MODE", "event")

	// Create response recorder
	w := httptest.NewRecorder()

	// Process the request - this should NOT trigger legacy fallback
	handled := integration.HandleRequest(w, req)
	if !handled {
		t.Fatal("Expected integration to handle the request (should return error response, not fallback)")
	}

	// Verify 500 Internal Server Error status
	if w.Code != 500 {
		t.Errorf("Expected status code 500 for processing failure, got %d", w.Code)
	}

	// Verify error response format
	if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json' for error response, got '%s'", contentType)
	}

	// Verify response body contains error information
	expectedBody := `{"error": "Internal Server Error", "message": "Event processing failed"}`
	actualBody := w.Body.String()
	if actualBody != expectedBody {
		t.Errorf("Expected error response body %q, got %q", expectedBody, actualBody)
	}
}

// TestErrorResponseHandling_NoLegacyFallback tests that errors don't trigger legacy fallback
func TestErrorResponseHandling_NoLegacyFallback(t *testing.T) {
	// Create mock event service that returns error
	mockEventService := &mockEventService{
		processEventError: errors.New("service failure"),
	}

	// Create mock response builder - but return nil to simulate what happens when handler returns error
	mockBuilder := &mockResponseBuilder{
		buildErrorResponse: nil, // This will cause handler to return error
		buildError:         errors.New("response building failed"),
	}

	// Create JSON handler with failing services
	jsonHandler := NewJSONEventHandler(
		mockEventService,
		&mockRenderService{},
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

	// Process the request - this should fail but NOT trigger legacy fallback
	// Instead, it should return false to indicate handler chain couldn't process it
	handled := integration.HandleRequest(w, req)
	if handled {
		t.Fatal("Expected integration to fail to handle the request (should return false, not trigger legacy fallback)")
	}

	// In this test case, the integration layer couldn't handle it,
	// which would normally trigger legacy fallback in the route layer
	// But the important thing is that the handler chain attempted to handle it
	// and failed gracefully rather than panicking

	t.Log("Handler chain failed gracefully without triggering internal legacy fallback")
}

// TestErrorResponseHandling_ValidationErrors tests 400 Bad Request for validation failures
func TestErrorResponseHandling_ValidationErrors(t *testing.T) {
	// Create mock validator that returns validation error
	mockValidator := &mockEventValidator{
		validateError: errors.New("validation failed: required field missing"),
	}

	// Create mock response builder for validation errors
	mockBuilder := &mockResponseBuilder{
		buildErrorResponse: &firHttp.ResponseModel{
			StatusCode: 400,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: []byte(`{"error": "Bad Request", "message": "Validation failed: required field missing"}`),
		},
		buildError: nil, // No error from response builder
	}

	// Create JSON handler with validator
	jsonHandler := NewJSONEventHandler(
		&mockEventService{},
		&mockRenderService{},
		mockBuilder,
		mockValidator,
	)

	// Create integration and setup chain
	chain := NewPriorityHandlerChain(&defaultHandlerLogger{}, &defaultHandlerMetrics{})
	chain.AddHandlerWithConfig(jsonHandler, HandlerConfig{
		Name:     jsonHandler.HandlerName(),
		Priority: 10,
		Enabled:  true,
	})
	integration := NewRouteHandlerIntegration(chain)

	// Create test request with valid JSON but invalid event data
	req := httptest.NewRequest("POST", "/events", strings.NewReader(`{"id": "test", "data": "value"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-FIR-MODE", "event")

	// Create response recorder
	w := httptest.NewRecorder()

	// Process the request
	handled := integration.HandleRequest(w, req)
	if !handled {
		t.Fatal("Expected integration to handle the request (should return validation error response)")
	}

	// Verify 400 Bad Request status
	if w.Code != 400 {
		t.Errorf("Expected status code 400 for validation failure, got %d", w.Code)
	}

	// Verify error response format
	if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json' for error response, got '%s'", contentType)
	}

	// Verify response body contains validation error information
	expectedBody := `{"error": "Bad Request", "message": "Validation failed: required field missing"}`
	actualBody := w.Body.String()
	if actualBody != expectedBody {
		t.Errorf("Expected validation error response body %q, got %q", expectedBody, actualBody)
	}
}

// TestErrorResponseHandling_FormDataErrors tests error handling for form data processing
func TestErrorResponseHandling_FormDataErrors(t *testing.T) {
	// Create mock response builder that fails only for template responses
	mockBuilder := &selectiveMockResponseBuilder{
		errorResponse: &firHttp.ResponseModel{
			StatusCode: 500,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: []byte(`{"error": "Internal Server Error", "message": "Failed to build form submit response: template response builder failed"}`),
		},
		templateError: errors.New("template response builder failed"),
	}

	// Create mock event service that fails to force fallback to template response
	mockEvent := &mockEventService{
		processEventError: errors.New("event service unavailable"),
	}

	// Create form handler
	formHandler := NewFormHandler(
		mockEvent,
		&mockRenderService{},
		mockBuilder,
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

	// Create test request with empty form data (will cause determineFormAction to fail)
	req := httptest.NewRequest("POST", "/form", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Create response recorder
	w := httptest.NewRecorder()

	// Process the request
	handled := integration.HandleRequest(w, req)
	if !handled {
		t.Fatal("Expected integration to handle the request (should return error response)")
	}

	// Verify 500 Internal Server Error status (template response builder failed)
	if w.Code != 500 {
		t.Errorf("Expected status code 500 for template response failure, got %d", w.Code)
	}

	// Verify error response format
	if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json' for error response, got '%s'", contentType)
	}

	// Verify response body contains error information
	expectedBody := `{"error": "Internal Server Error", "message": "Failed to build form submit response: template response builder failed"}`
	actualBody := w.Body.String()
	if actualBody != expectedBody {
		t.Errorf("Expected form error response body %q, got %q", expectedBody, actualBody)
	}
}
