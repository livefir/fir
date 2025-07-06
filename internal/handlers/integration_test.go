package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	firHttp "github.com/livefir/fir/internal/http"
	"github.com/livefir/fir/internal/routeservices"
	"github.com/livefir/fir/internal/services"
)

func TestRouteHandlerIntegration_HandleRequest(t *testing.T) {
	// Create mock services
	mockEventService := &mockEventService{
		processEventResponse: &services.EventResponse{
			StatusCode: 200,
			Body:       []byte("event processed"),
		},
	}
	mockRenderService := &mockRenderService{
		renderTemplateResponse: &services.RenderResult{
			HTML: []byte("<html>test</html>"),
		},
	}
	mockTemplateService := &mockTemplateService{}
	mockResponseBuilder := &mockResponseBuilder{
		buildEventResponse: &firHttp.ResponseModel{
			StatusCode: 200,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       []byte("event response"),
		},
		buildTemplateResponse: &firHttp.ResponseModel{
			StatusCode: 200,
			Headers:    map[string]string{"Content-Type": "text/html"},
			Body:       []byte("<html>test</html>"),
		},
	}

	// Create RouteServices
	routeServices := &routeservices.RouteServices{
		EventService:    mockEventService,
		RenderService:   mockRenderService,
		TemplateService: mockTemplateService,
		ResponseBuilder: mockResponseBuilder,
	}

	// Setup handler chain
	handlerChain := SetupDefaultHandlerChain(routeServices)
	integration := NewRouteHandlerIntegration(handlerChain)

	tests := []struct {
		name           string
		method         string
		url            string
		contentType    string
		body           string
		expectedStatus int
		shouldHandle   bool
	}{
		{
			name:           "handles JSON event request",
			method:         "POST",
			url:            "/events",
			contentType:    "application/json",
			body:           `{"id": "test", "data": "value"}`,
			expectedStatus: 200,
			shouldHandle:   true,
		},
		{
			name:           "handles GET request",
			method:         "GET",
			url:            "/test",
			contentType:    "",
			body:           "",
			expectedStatus: 200,
			shouldHandle:   true,
		},
		{
			name:           "handles form POST request",
			method:         "POST",
			url:            "/form",
			contentType:    "application/x-www-form-urlencoded",
			body:           "name=test&value=123",
			expectedStatus: 200,
			shouldHandle:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test request
			var bodyReader *strings.Reader
			if tt.body != "" {
				bodyReader = strings.NewReader(tt.body)
			} else {
				bodyReader = strings.NewReader("") // Use empty string instead of nil
			}

			req := httptest.NewRequest(tt.method, tt.url, bodyReader)
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Test if integration can handle the request
			canHandle := integration.CanHandleRequest(req)
			if canHandle != tt.shouldHandle {
				t.Errorf("CanHandleRequest() = %v, expected %v", canHandle, tt.shouldHandle)
			}

			if !tt.shouldHandle {
				return // Skip actual handling test
			}

			// Handle the request
			handled := integration.HandleRequest(w, req)
			if !handled {
				t.Errorf("HandleRequest() = false, expected true")
			}

			// Check response status
			if w.Code != tt.expectedStatus {
				t.Errorf("response status = %d, expected %d", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestSetupDefaultHandlerChain(t *testing.T) {
	// Create minimal RouteServices
	routeServices := &routeservices.RouteServices{
		EventService:    &mockEventService{},
		RenderService:   &mockRenderService{},
		TemplateService: &mockTemplateService{},
		ResponseBuilder: &mockResponseBuilder{},
	}

	// Setup handler chain
	chain := SetupDefaultHandlerChain(routeServices)

	// Verify handlers were added
	handlers := chain.GetHandlers()
	if len(handlers) == 0 {
		t.Error("expected handlers to be added to chain")
	}

	// Check that we have the expected handler types
	handlerNames := make(map[string]bool)
	for _, handler := range handlers {
		handlerNames[handler.HandlerName()] = true
	}

	expectedHandlers := []string{"websocket-handler", "json-event-handler", "form-handler", "get-handler"}
	for _, expected := range expectedHandlers {
		if !handlerNames[expected] {
			t.Errorf("expected handler %s not found in chain", expected)
		}
	}
}

func TestRouteHandlerIntegration_CanHandleRequest(t *testing.T) {
	// Create handler chain with limited handlers
	logger := &defaultHandlerLogger{}
	metrics := &defaultHandlerMetrics{}
	chain := NewPriorityHandlerChain(logger, metrics)

	// Add only a JSON event handler
	mockEventService := &mockEventService{}
	mockRenderService := &mockRenderService{}
	mockResponseBuilder := &mockResponseBuilder{}

	jsonHandler := NewJSONEventHandler(
		mockEventService,
		mockRenderService,
		mockResponseBuilder,
		nil,
	)

	chain.AddHandlerWithConfig(jsonHandler, HandlerConfig{
		Name:     jsonHandler.HandlerName(), // Use actual handler name
		Priority: 10,
		Enabled:  true,
	})

	integration := NewRouteHandlerIntegration(chain)

	tests := []struct {
		name         string
		method       string
		url          string
		contentType  string
		expectedCan  bool
	}{
		{
			name:        "can handle JSON POST",
			method:      "POST",
			url:         "/test",
			contentType: "application/json",
			expectedCan: true,
		},
		{
			name:        "cannot handle GET",
			method:      "GET", 
			url:         "/test",
			contentType: "",
			expectedCan: false,
		},
		{
			name:        "cannot handle form POST",
			method:      "POST",
			url:         "/test",
			contentType: "application/x-www-form-urlencoded",
			expectedCan: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, strings.NewReader(""))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			canHandle := integration.CanHandleRequest(req)
			if canHandle != tt.expectedCan {
				t.Errorf("CanHandleRequest() = %v, expected %v", canHandle, tt.expectedCan)
			}
		})
	}
}
