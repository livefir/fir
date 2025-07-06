package fir

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/goccy/go-json"

	firHttp "github.com/livefir/fir/internal/http"
	"github.com/livefir/fir/internal/services"
	"github.com/livefir/fir/pubsub"
)

// MockEventService for testing
type MockEventService struct {
	processFunc func(ctx context.Context, req services.EventRequest) (*services.EventResponse, error)
	metrics     services.EventMetrics
}

func (m *MockEventService) ProcessEvent(ctx context.Context, req services.EventRequest) (*services.EventResponse, error) {
	if m.processFunc != nil {
		return m.processFunc(ctx, req)
	}
	return &services.EventResponse{
		StatusCode: http.StatusOK,
		Headers:    make(map[string]string),
		Body:       []byte("test response"),
		Events:     []firHttp.DOMEvent{},
		PubSubEvents: []pubsub.Event{},
	}, nil
}

func (m *MockEventService) RegisterHandler(eventID string, handler services.EventHandler) error {
	return nil
}

func (m *MockEventService) GetEventMetrics() services.EventMetrics {
	return m.metrics
}

func TestRouteEventProcessor_ProcessEvent(t *testing.T) {
	// Create a mock event service
	mockService := &MockEventService{}
	
	// Create a simple route
	route := createTestRoute("test-route")

	processor := NewRouteEventProcessor(mockService, route)

	// Create test event
	sessionID := "test-session"
	target := "test-target"
	elementKey := "test-element"
	
	event := Event{
		ID:         "test-event",
		Target:     &target,
		ElementKey: &elementKey,
		SessionID:  &sessionID,
		IsForm:     false,
		Params:     []byte(`{"test": "data"}`),
	}

	// Create test HTTP request
	req := httptest.NewRequest("POST", "/test", strings.NewReader("test body"))
	w := httptest.NewRecorder()

	// Process the event
	response, err := processor.ProcessEvent(context.Background(), event, req, w)

	// Verify results
	if err != nil {
		t.Fatalf("ProcessEvent failed: %v", err)
	}

	if response == nil {
		t.Fatal("Response is nil")
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, response.StatusCode)
	}

	if string(response.Body) != "test response" {
		t.Errorf("Expected body 'test response', got '%s'", string(response.Body))
	}
}

func TestRouteEventProcessor_ConvertEventParams(t *testing.T) {
	sessionID := "test-session"
	event := Event{
		ID:        "test-event",
		SessionID: &sessionID,
		IsForm:    true,
		Params:    []byte(`{"test": "data"}`),
	}

	params := convertEventParams(event)

	if params["is_form"] != true {
		t.Errorf("Expected is_form to be true, got %v", params["is_form"])
	}

	rawParams, ok := params["raw"].([]byte)
	if !ok {
		// json.RawMessage is an alias for []byte, so try that
		if jsonRaw, isJsonRaw := params["raw"].(json.RawMessage); isJsonRaw {
			rawParams = []byte(jsonRaw)
		} else {
			t.Fatalf("Expected raw params to be []byte or json.RawMessage, got %T", params["raw"])
		}
	}

	if string(rawParams) != `{"test": "data"}` {
		t.Errorf("Expected raw params to be '{\"test\": \"data\"}', got '%s'", string(rawParams))
	}
}

func TestLegacyEventHandler_ProcessEvent(t *testing.T) {
	// Create a legacy handler function
	handlerCalled := false
	legacyHandler := func(ctx RouteContext) error {
		handlerCalled = true
		
		// Verify the event was properly converted
		if ctx.event.ID != "test-event" {
			t.Errorf("Expected event ID 'test-event', got '%s'", ctx.event.ID)
		}
		
		if ctx.event.SessionID == nil || *ctx.event.SessionID != "test-session" {
			t.Errorf("Expected session ID 'test-session', got %v", ctx.event.SessionID)
		}
		
		return nil
	}

	// Create a simple route
	route := createTestRoute("test-route")

	handler := NewLegacyEventHandler(legacyHandler, route)

	// Create test event request
	eventReq := services.EventRequest{
		ID:        "test-event",
		SessionID: "test-session",
		Context:   context.Background(),
		Params: map[string]interface{}{
			"raw":     []byte(`{"test": "data"}`),
			"is_form": false,
		},
		RequestModel: &firHttp.RequestModel{
			Method: "POST",
			URL:    mustParseURL("http://test.com/test"),
			Header: make(http.Header),
			Host:   "test.com",
		},
	}

	// Process the event
	response, err := handler.ProcessEvent(context.Background(), eventReq)

	// Verify results
	if err != nil {
		t.Fatalf("ProcessEvent failed: %v", err)
	}

	if !handlerCalled {
		t.Error("Legacy handler was not called")
	}

	if response == nil {
		t.Fatal("Response is nil")
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, response.StatusCode)
	}
}

func TestLegacyEventHandler_ProcessEvent_WithError(t *testing.T) {
	// Create a legacy handler that returns an error
	legacyHandler := func(ctx RouteContext) error {
		return errors.New("test error")
	}

	// Create a simple route
	route := createTestRoute("test-route")

	handler := NewLegacyEventHandler(legacyHandler, route)

	// Create test event request
	eventReq := services.EventRequest{
		ID:        "test-event",
		SessionID: "test-session",
		Context:   context.Background(),
		Params: map[string]interface{}{
			"raw":     []byte(`{"test": "data"}`),
			"is_form": false,
		},
		RequestModel: &firHttp.RequestModel{
			Method: "POST",
			URL:    mustParseURL("http://test.com/test"),
			Header: make(http.Header),
			Host:   "test.com",
		},
	}

	// Process the event
	response, err := handler.ProcessEvent(context.Background(), eventReq)

	// Verify results
	if err != nil {
		t.Fatalf("ProcessEvent failed: %v", err)
	}

	if response == nil {
		t.Fatal("Response is nil")
	}

	if response.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, response.StatusCode)
	}

	if response.Errors == nil {
		t.Error("Expected errors to be set")
	} else if response.Errors["error"] != "test error" {
		t.Errorf("Expected error 'test error', got '%v'", response.Errors["error"])
	}
}

// Helper functions for tests
func mustParseURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}

// createTestRoute creates a test route with the given ID
func createTestRoute(id string) *route {
	return &route{
		routeOpt: routeOpt{
			id: id,
		},
	}
}
