package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResponseAdapter_EdgeCases(t *testing.T) {
	t.Run("double write protection", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		adapter := NewResponseAdapter(recorder)

		// First write should succeed
		err := adapter.WriteHTML("<h1>First</h1>")
		if err != nil {
			t.Errorf("First write should succeed: %v", err)
		}

		// Second write should fail
		err = adapter.WriteHTML("<h1>Second</h1>")
		if err == nil {
			t.Error("Second write should fail with double write protection")
		}
	})

	t.Run("status code handling", func(t *testing.T) {
		tests := []struct {
			name           string
			statusCode     int
			expectedStatus int
		}{
			{"explicit 201", 201, 201},
			{"explicit 404", 404, 404},
			{"explicit 500", 500, 500},
			{"zero status (default 200)", 0, 200},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				recorder := httptest.NewRecorder()
				adapter := NewResponseAdapter(recorder)

				response := ResponseModel{
					StatusCode: tt.statusCode,
					Body:       []byte("test"),
				}

				err := adapter.WriteResponse(response)
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				if recorder.Code != tt.expectedStatus {
					t.Errorf("Expected status %d, got %d", tt.expectedStatus, recorder.Code)
				}
			})
		}
	})

	t.Run("multiple headers", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		adapter := NewResponseAdapter(recorder)

		response := ResponseModel{
			Headers: map[string]string{
				"X-Custom-1": "value1",
				"X-Custom-2": "value2",
				"X-Custom-3": "value3",
			},
			Body: []byte("test"),
		}

		err := adapter.WriteResponse(response)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		for key, expectedValue := range response.Headers {
			actualValue := recorder.Header().Get(key)
			if actualValue != expectedValue {
				t.Errorf("Expected header %s to be %s, got %s", key, expectedValue, actualValue)
			}
		}
	})

	t.Run("empty response", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		adapter := NewResponseAdapter(recorder)

		response := ResponseModel{}

		err := adapter.WriteResponse(response)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if recorder.Code != 200 {
			t.Errorf("Expected default status 200, got %d", recorder.Code)
		}

		if recorder.Body.Len() != 0 {
			t.Errorf("Expected empty body, got %s", recorder.Body.String())
		}
	})
}

func TestResponseAdapter_JSON(t *testing.T) {
	t.Run("valid JSON object", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		adapter := NewResponseAdapter(recorder)

		data := map[string]interface{}{
			"message": "success",
			"count":   42,
		}

		err := adapter.WriteJSON(data)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if recorder.Header().Get("Content-Type") != "application/json" {
			t.Error("Expected Content-Type to be application/json")
		}

		body := recorder.Body.String()
		if !strings.Contains(body, "success") || !strings.Contains(body, "42") {
			t.Errorf("Expected JSON body to contain data, got %s", body)
		}
	})

	t.Run("invalid JSON data", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		adapter := NewResponseAdapter(recorder)

		// channels cannot be marshaled to JSON
		invalidData := make(chan int)

		err := adapter.WriteJSON(invalidData)
		if err == nil {
			t.Error("Expected error when marshaling invalid JSON data")
		}
	})
}

func TestResponseBuilder_EdgeCases(t *testing.T) {
	t.Run("method chaining", func(t *testing.T) {
		response := NewResponseBuilder().
			WithStatus(201).
			WithHeader("X-Test", "value").
			WithHeader("X-Another", "another").
			WithHTML("<div>Test</div>").
			Build()

		if response.StatusCode != 201 {
			t.Errorf("Expected status 201, got %d", response.StatusCode)
		}

		if len(response.Headers) != 3 { // Content-Type + 2 custom headers
			t.Errorf("Expected 3 headers, got %d", len(response.Headers))
		}

		if response.Headers["Content-Type"] != "text/html; charset=utf-8" {
			t.Error("Expected HTML content type to be set")
		}
	})

	t.Run("JSON with invalid data", func(t *testing.T) {
		response := NewResponseBuilder().
			WithJSON(make(chan int)). // invalid JSON data
			Build()

		if response.StatusCode != http.StatusInternalServerError {
			t.Errorf("Expected status 500 for invalid JSON, got %d", response.StatusCode)
		}

		body := string(response.Body)
		if !strings.Contains(body, "error") {
			t.Errorf("Expected error message in body, got %s", body)
		}
	})

	t.Run("redirect builder", func(t *testing.T) {
		response := NewResponseBuilder().
			WithRedirect("/new-location", 301).
			Build()

		if response.Redirect == nil {
			t.Error("Expected redirect to be set")
		}

		if response.Redirect.URL != "/new-location" {
			t.Errorf("Expected redirect URL '/new-location', got %s", response.Redirect.URL)
		}

		if response.Redirect.StatusCode != 301 {
			t.Errorf("Expected redirect status 301, got %d", response.Redirect.StatusCode)
		}
	})

	t.Run("DOM events", func(t *testing.T) {
		event1 := DOMEvent{ID: "event1", Type: "update", HTML: "<div>1</div>"}
		event2 := DOMEvent{ID: "event2", Type: "replace", HTML: "<div>2</div>"}

		response := NewResponseBuilder().
			WithEvent(event1).
			WithEvent(event2).
			Build()

		if len(response.Events) != 2 {
			t.Errorf("Expected 2 events, got %d", len(response.Events))
		}

		if response.Events[0].ID != "event1" {
			t.Errorf("Expected first event ID 'event1', got %s", response.Events[0].ID)
		}

		if response.Events[1].ID != "event2" {
			t.Errorf("Expected second event ID 'event2', got %s", response.Events[1].ID)
		}
	})
}

func TestRequestResponsePair(t *testing.T) {
	t.Run("successful creation and usage", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/test?query=value", strings.NewReader("form=data"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		recorder := httptest.NewRecorder()

		pathExtractor := func(r *http.Request) map[string]string {
			return map[string]string{"id": "123"}
		}

		pair, err := NewRequestResponsePair(recorder, req, pathExtractor)
		if err != nil {
			t.Errorf("Unexpected error creating pair: %v", err)
		}

		// Test request parsing
		if pair.Request.Method != "POST" {
			t.Errorf("Expected method POST, got %s", pair.Request.Method)
		}

		if pair.Request.QueryParams.Get("query") != "value" {
			t.Errorf("Expected query param 'query' to be 'value', got %s", pair.Request.QueryParams.Get("query"))
		}

		if pair.Request.PathParams["id"] != "123" {
			t.Errorf("Expected path param 'id' to be '123', got %s", pair.Request.PathParams["id"])
		}

		// Test response writing
		err = pair.Response.WriteHTML("<h1>Success</h1>")
		if err != nil {
			t.Errorf("Unexpected error writing response: %v", err)
		}

		if recorder.Code != 200 {
			t.Errorf("Expected status 200, got %d", recorder.Code)
		}
	})

	t.Run("malformed request handling", func(t *testing.T) {
		// Create a request with an extremely long URL that might cause issues
		longPath := strings.Repeat("x", 10000)
		req := httptest.NewRequest("GET", "/"+longPath, nil)
		recorder := httptest.NewRecorder()

		// This should succeed since URL parsing is quite robust
		pair, err := NewRequestResponsePair(recorder, req, nil)
		if err != nil {
			t.Logf("Got expected error for extreme request: %v", err)
		} else {
			// If it succeeds, that's also fine - just verify it parsed correctly
			if pair.Request.URL.Path != "/"+longPath {
				t.Errorf("URL not parsed correctly")
			}
		}
	})
}

func TestHandleHTTPRequest_ErrorHandling(t *testing.T) {
	t.Run("handler returns error", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		recorder := httptest.NewRecorder()

		HandleHTTPRequest(recorder, req, nil, func(pair *RequestResponsePair) error {
			return errors.New("test error")
		})

		if recorder.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500, got %d", recorder.Code)
		}

		body := recorder.Body.String()
		if !strings.Contains(body, "test error") {
			t.Errorf("Expected error message in body, got %s", body)
		}
	})

	t.Run("handler writes response then returns error", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		recorder := httptest.NewRecorder()

		HandleHTTPRequest(recorder, req, nil, func(pair *RequestResponsePair) error {
			// Write response first
			pair.Response.WriteHTML("<h1>Success</h1>")
			// Then return error - should not overwrite response
			return errors.New("error after write")
		})

		if recorder.Code != 200 {
			t.Errorf("Expected status 200 (from successful write), got %d", recorder.Code)
		}

		body := recorder.Body.String()
		if body != "<h1>Success</h1>" {
			t.Errorf("Expected original response body, got %s", body)
		}
	})

	t.Run("successful request handling", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		recorder := httptest.NewRecorder()

		HandleHTTPRequest(recorder, req, nil, func(pair *RequestResponsePair) error {
			return pair.Response.WriteHTML("<h1>Success</h1>")
		})

		if recorder.Code != 200 {
			t.Errorf("Expected status 200, got %d", recorder.Code)
		}

		body := recorder.Body.String()
		if body != "<h1>Success</h1>" {
			t.Errorf("Expected success body, got %s", body)
		}
	})
}
