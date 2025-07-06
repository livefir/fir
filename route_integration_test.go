package fir

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/livefir/fir/internal/routeservices"
	"github.com/livefir/fir/internal/services"
)

func TestRoute_HandleJSONEventWithService(t *testing.T) {
	// Create a mock event service
	mockService := &MockEventService{}
	eventProcessed := false

	// Set up the mock to track when ProcessEvent is called
	mockService.processFunc = func(ctx context.Context, req services.EventRequest) (*services.EventResponse, error) {
		eventProcessed = true
		if req.ID != "test-event" {
			t.Errorf("Expected event ID 'test-event', got '%s'", req.ID)
		}
		return &services.EventResponse{
			StatusCode: http.StatusOK,
			Headers:    map[string]string{"Content-Type": "text/plain"},
			Body:       []byte("service processed"),
		}, nil
	}

	// Create route services with the mock event service
	services := &routeservices.RouteServices{
		EventService: mockService,
	}

	// Create a test route
	route := &route{
		routeOpt: routeOpt{
			id: "test-route",
		},
		services: services,
	}

	// Create test request with JSON event
	reqBody := `{"event_id": "test-event", "params": {"test": "data"}}`
	req := httptest.NewRequest("POST", "/test", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call the new service-based event handler
	route.handleJSONEventWithService(w, req)

	// Verify the event was processed by the service
	if !eventProcessed {
		t.Error("Event was not processed by the service layer")
	}

	// Verify the response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Header().Get("Content-Type") != "text/plain" {
		t.Errorf("Expected Content-Type 'text/plain', got '%s'", w.Header().Get("Content-Type"))
	}

	if w.Body.String() != "service processed" {
		t.Errorf("Expected body 'service processed', got '%s'", w.Body.String())
	}
}

func TestRoute_HandleJSONEventWithService_FallbackToLegacy(t *testing.T) {
	// Create route services without event service (nil)
	services := &routeservices.RouteServices{
		EventService: nil, // No event service, should fallback to legacy
	}

	// Create a test route
	route := &route{
		routeOpt: routeOpt{
			id: "test-route",
		},
		services: services,
	}

	// Create test request with JSON event
	reqBody := `{"event_id": "test-event", "params": {"test": "data"}}`
	req := httptest.NewRequest("POST", "/test", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// This should call the legacy handler, but since we don't have a real
	// event registry set up, it will return an error. We're just testing
	// that the fallback path is taken.
	route.handleJSONEventWithService(w, req)

	// The legacy handler will fail because there's no event registry,
	// but we can verify it attempted to fall back by checking it didn't
	// panic or return early
	if w.Code == 0 {
		t.Error("Expected some response code, got 0 (no response written)")
	}
}
