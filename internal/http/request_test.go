package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestAdapter_ParseRequest(t *testing.T) {
	tests := []struct {
		name               string
		setupRequest       func() *http.Request
		pathParamExtractor func(*http.Request) map[string]string
		expectError        bool
		validateModel      func(*testing.T, *RequestModel)
	}{
		{
			name: "basic GET request",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/test?param=value", nil)
				req.Header.Set("User-Agent", "test-agent")
				return req
			},
			validateModel: func(t *testing.T, model *RequestModel) {
				if model.Method != "GET" {
					t.Errorf("Expected method GET, got %s", model.Method)
				}
				if model.URL.Path != "/test" {
					t.Errorf("Expected path /test, got %s", model.URL.Path)
				}
				if model.QueryParams.Get("param") != "value" {
					t.Errorf("Expected query param 'param' to be 'value', got %s", model.QueryParams.Get("param"))
				}
				if model.Header.Get("User-Agent") != "test-agent" {
					t.Errorf("Expected User-Agent header to be preserved")
				}
			},
		},
		{
			name: "POST request with form data",
			setupRequest: func() *http.Request {
				body := strings.NewReader("name=test&value=123")
				req := httptest.NewRequest("POST", "/submit", body)
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				return req
			},
			validateModel: func(t *testing.T, model *RequestModel) {
				if model.Method != "POST" {
					t.Errorf("Expected method POST, got %s", model.Method)
				}
				if model.Form.Get("name") != "test" {
					t.Errorf("Expected form param 'name' to be 'test', got %s", model.Form.Get("name"))
				}
				if model.Form.Get("value") != "123" {
					t.Errorf("Expected form param 'value' to be '123', got %s", model.Form.Get("value"))
				}
			},
		},
		{
			name: "request with path parameters",
			setupRequest: func() *http.Request {
				return httptest.NewRequest("GET", "/users/123", nil)
			},
			pathParamExtractor: func(r *http.Request) map[string]string {
				return map[string]string{"id": "123"}
			},
			validateModel: func(t *testing.T, model *RequestModel) {
				if model.PathParams["id"] != "123" {
					t.Errorf("Expected path param 'id' to be '123', got %s", model.PathParams["id"])
				}
			},
		},
		{
			name: "WebSocket upgrade request",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/ws", nil)
				req.Header.Set("Connection", "Upgrade")
				req.Header.Set("Upgrade", "websocket")
				return req
			},
			validateModel: func(t *testing.T, model *RequestModel) {
				if !model.IsWebSocket {
					t.Error("Expected IsWebSocket to be true for WebSocket upgrade request")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewRequestAdapter(tt.pathParamExtractor)
			req := tt.setupRequest()

			model, err := adapter.ParseRequest(req)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if tt.validateModel != nil && model != nil {
				tt.validateModel(t, model)
			}
		})
	}
}

func TestRequestAdapter_ParseEventData(t *testing.T) {
	tests := []struct {
		name          string
		setupRequest  func() *http.Request
		expectedEvent EventData
	}{
		{
			name: "event from query parameters",
			setupRequest: func() *http.Request {
				return httptest.NewRequest("GET", "/test?event=click&target=button1", nil)
			},
			expectedEvent: EventData{
				ID:     "click",
				Target: stringPtr("button1"),
			},
		},
		{
			name: "event from form data",
			setupRequest: func() *http.Request {
				body := strings.NewReader("event=submit&element_key=form1&name=value")
				req := httptest.NewRequest("POST", "/test", body)
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				return req
			},
			expectedEvent: EventData{
				ID:         "submit",
				ElementKey: stringPtr("form1"),
				Params: map[string]interface{}{
					"event":       "submit",
					"element_key": "form1",
					"name":        "value",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewRequestAdapter(nil)
			req := tt.setupRequest()

			eventData, err := adapter.ParseEventData(req)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if eventData.ID != tt.expectedEvent.ID {
				t.Errorf("Expected event ID %s, got %s", tt.expectedEvent.ID, eventData.ID)
			}

			if (eventData.Target == nil) != (tt.expectedEvent.Target == nil) {
				t.Errorf("Target pointer mismatch")
			} else if eventData.Target != nil && *eventData.Target != *tt.expectedEvent.Target {
				t.Errorf("Expected target %s, got %s", *tt.expectedEvent.Target, *eventData.Target)
			}

			if (eventData.ElementKey == nil) != (tt.expectedEvent.ElementKey == nil) {
				t.Errorf("ElementKey pointer mismatch")
			} else if eventData.ElementKey != nil && *eventData.ElementKey != *tt.expectedEvent.ElementKey {
				t.Errorf("Expected element key %s, got %s", *tt.expectedEvent.ElementKey, *eventData.ElementKey)
			}
		})
	}
}

func TestResponseAdapter_WriteResponse(t *testing.T) {
	tests := []struct {
		name           string
		response       ResponseModel
		expectedStatus int
		expectedHeader string
		expectedBody   string
	}{
		{
			name: "basic HTML response",
			response: ResponseModel{
				StatusCode: 200,
				Headers:    map[string]string{"Content-Type": "text/html"},
				Body:       []byte("<h1>Hello</h1>"),
			},
			expectedStatus: 200,
			expectedHeader: "text/html",
			expectedBody:   "<h1>Hello</h1>",
		},
		{
			name: "JSON response with events",
			response: ResponseModel{
				StatusCode: 200,
				Events: []DOMEvent{
					{ID: "test", Type: "update", HTML: "<div>Updated</div>"},
				},
			},
			expectedStatus: 200,
			expectedHeader: "application/json",
			expectedBody:   `[{"id":"test","type":"update","html":"\u003cdiv\u003eUpdated\u003c/div\u003e"}]`,
		},
		{
			name: "redirect response",
			response: ResponseModel{
				Redirect: &RedirectInfo{
					URL:        "/new-location",
					StatusCode: 302,
				},
			},
			expectedStatus: 302,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			adapter := NewResponseAdapter(recorder)

			err := adapter.WriteResponse(tt.response)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if recorder.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, recorder.Code)
			}

			if tt.expectedHeader != "" {
				contentType := recorder.Header().Get("Content-Type")
				if !strings.Contains(contentType, tt.expectedHeader) {
					t.Errorf("Expected header to contain %s, got %s", tt.expectedHeader, contentType)
				}
			}

			if tt.expectedBody != "" {
				body := recorder.Body.String()
				if body != tt.expectedBody {
					t.Errorf("Expected body %s, got %s", tt.expectedBody, body)
				}
			}

			// Special handling for redirect test
			if tt.response.Redirect != nil {
				location := recorder.Header().Get("Location")
				if location != tt.response.Redirect.URL {
					t.Errorf("Expected Location header %s, got %s", tt.response.Redirect.URL, location)
				}
			}
		})
	}
}

func TestResponseAdapter_WriteError(t *testing.T) {
	recorder := httptest.NewRecorder()
	adapter := NewResponseAdapter(recorder)

	err := adapter.WriteError(http.StatusBadRequest, "Test error")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "Test error") {
		t.Errorf("Expected body to contain 'Test error', got %s", body)
	}
}

func TestResponseBuilder(t *testing.T) {
	builder := NewResponseBuilder()
	response := builder.
		WithStatus(201).
		WithHeader("X-Custom", "test").
		WithHTML("<h1>Created</h1>").
		Build()

	if response.StatusCode != 201 {
		t.Errorf("Expected status 201, got %d", response.StatusCode)
	}

	if response.Headers["X-Custom"] != "test" {
		t.Errorf("Expected custom header 'test', got %s", response.Headers["X-Custom"])
	}

	if response.Headers["Content-Type"] != "text/html; charset=utf-8" {
		t.Errorf("Expected HTML content type to be set")
	}

	if string(response.Body) != "<h1>Created</h1>" {
		t.Errorf("Expected body '<h1>Created</h1>', got %s", string(response.Body))
	}
}

func TestStandardHTTPAdapter(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?param=value", nil)
	recorder := httptest.NewRecorder()

	adapter := NewStandardHTTPAdapter(recorder, req, func(r *http.Request) map[string]string {
		return map[string]string{"test": "param"}
	})

	// Test request parsing
	model, err := adapter.ParseRequest(req)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if model.Method != "GET" {
		t.Errorf("Expected method GET, got %s", model.Method)
	}

	// Test response writing
	response := ResponseModel{
		StatusCode: 200,
		Body:       []byte("Test response"),
	}

	err = adapter.WriteResponse(response)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if recorder.Code != 200 {
		t.Errorf("Expected status 200, got %d", recorder.Code)
	}
}

func TestHandleHTTPRequest(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	recorder := httptest.NewRecorder()

	handlerCalled := false
	HandleHTTPRequest(recorder, req, nil, func(pair *RequestResponsePair) error {
		handlerCalled = true
		
		// Verify request was parsed correctly
		if pair.Request.Method != "GET" {
			t.Errorf("Expected method GET, got %s", pair.Request.Method)
		}

		// Write a response
		return pair.Response.WriteHTML("<h1>Success</h1>")
	})

	if !handlerCalled {
		t.Error("Handler was not called")
	}

	if recorder.Code != 200 {
		t.Errorf("Expected status 200, got %d", recorder.Code)
	}

	body := recorder.Body.String()
	if body != "<h1>Success</h1>" {
		t.Errorf("Expected body '<h1>Success</h1>', got %s", body)
	}
}

// Helper function for tests
func stringPtr(s string) *string {
	return &s
}
