package handlers

import (
	"context"
	"net/url"
	"testing"

	"github.com/livefir/fir/internal/http"
)

// mustParseURL is a helper function for tests
func mustParseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return u
}

// TestPOCHandler_SupportsRequest tests that POCHandler only supports GET /poc requests
func TestPOCHandler_SupportsRequest(t *testing.T) {
	handler := NewPOCHandler()

	tests := []struct {
		name     string
		request  *http.RequestModel
		expected bool
	}{
		{
			name: "supports GET /poc",
			request: &http.RequestModel{
				Method: "GET",
				URL:    mustParseURL("/poc"),
			},
			expected: true,
		},
		{
			name: "does not support POST /poc",
			request: &http.RequestModel{
				Method: "POST",
				URL:    mustParseURL("/poc"),
			},
			expected: false,
		},
		{
			name: "does not support GET /other",
			request: &http.RequestModel{
				Method: "GET",
				URL:    mustParseURL("/other"),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.SupportsRequest(tt.request)
			if result != tt.expected {
				t.Errorf("SupportsRequest() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestPOCHandler_Handle tests that POCHandler returns "POC Working" for supported requests
func TestPOCHandler_Handle(t *testing.T) {
	handler := NewPOCHandler()
	ctx := context.Background()

	request := &http.RequestModel{
		Method: "GET",
		URL:    mustParseURL("/poc"),
	}

	response, err := handler.Handle(ctx, request)
	if err != nil {
		t.Fatalf("Handle() returned error: %v", err)
	}

	if response == nil {
		t.Fatal("Handle() returned nil response")
	}

	if response.StatusCode != 200 {
		t.Errorf("Handle() returned status code %d, expected 200", response.StatusCode)
	}

	expectedBody := "POC Working"
	actualBody := string(response.Body)
	if actualBody != expectedBody {
		t.Errorf("Handle() returned body %q, expected %q", actualBody, expectedBody)
	}
}

// TestPOCHandler_HandlerName tests that POCHandler returns correct name
func TestPOCHandler_HandlerName(t *testing.T) {
	handler := NewPOCHandler()
	expected := "POCHandler"
	actual := handler.HandlerName()

	if actual != expected {
		t.Errorf("HandlerName() = %q, expected %q", actual, expected)
	}
}

// TestPOCHandler_UnsupportedRequest tests that POCHandler returns error for unsupported requests
func TestPOCHandler_UnsupportedRequest(t *testing.T) {
	handler := NewPOCHandler()
	ctx := context.Background()

	request := &http.RequestModel{
		Method: "POST",
		URL:    mustParseURL("/poc"),
	}

	response, err := handler.Handle(ctx, request)
	if err == nil {
		t.Fatal("Handle() should return error for unsupported request")
	}

	if response != nil {
		t.Error("Handle() should return nil response for unsupported request")
	}

	expectedErrorMsg := "request not supported by POCHandler"
	if err.Error() != expectedErrorMsg {
		t.Errorf("Handle() returned error %q, expected %q", err.Error(), expectedErrorMsg)
	}
}
