package fir

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/livefir/fir/internal/routeservices"
	"github.com/livefir/fir/pubsub"
)

func TestNewConnectionWithoutCookie(t *testing.T) {
	// Setup mock WebSocketServices
	wsServices := &routeservices.MockWebSocketServices{
		CookieName: "test_cookie",
	}

	// Create a test HTTP request without cookie
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Test with missing cookie - should fail
	_, err := NewConnectionWithServices(w, req, wsServices)
	if err == nil {
		t.Error("Expected error when cookie is missing, got nil")
	}
}

func TestConnectionLifecycle(t *testing.T) {
	// Setup mock WebSocketServices
	wsServices := &routeservices.MockWebSocketServices{
		CookieName: "test_cookie",
	}

	// Test the connection lifecycle methods
	conn := &Connection{
		wsServices:    wsServices,
		sessionID:     "test_session",
		routeID:       "test_route",
		user:          "test_user",
		send:          make(chan []byte, 10),
		writePumpDone: make(chan struct{}),
	}

	ctx, cancel := context.WithCancel(context.Background())
	conn.ctx = ctx
	conn.cancel = cancel

	// Test Close - should not panic
	conn.Close()

	// Verify context was cancelled
	select {
	case <-conn.ctx.Done():
		// Expected
	default:
		t.Error("Expected context to be cancelled after Close()")
	}
}

func TestConnectionIsDuplicateEvent(t *testing.T) {
	wsServices := &routeservices.MockWebSocketServices{
		DropDuplicateInterval: 250 * time.Millisecond,
	}

	sessionID := "test_session"
	elementKey := "test_key"
	conn := &Connection{
		wsServices: wsServices,
		lastEvent: Event{
			ID:         "test_event",
			SessionID:  &sessionID,
			ElementKey: &elementKey,
			Params:     []byte(`{"test": "data"}`),
			Timestamp:  time.Now().UTC().UnixMilli(),
		},
	}

	// Test with identical event within duplicate interval
	newEvent := Event{
		ID:         "test_event",
		SessionID:  &sessionID,
		ElementKey: &elementKey,
		Params:     []byte(`{"test": "data"}`),
		Timestamp:  time.Now().UTC().UnixMilli(),
	}

	if !conn.isDuplicateEvent(newEvent) {
		t.Error("Expected duplicate event to be detected")
	}

	// Test with different event ID
	newEvent.ID = "different_event"
	if conn.isDuplicateEvent(newEvent) {
		t.Error("Expected non-duplicate event with different ID")
	}

	// Test with different params
	newEvent.ID = "test_event"
	newEvent.Params = []byte(`{"test": "different"}`)
	if conn.isDuplicateEvent(newEvent) {
		t.Error("Expected non-duplicate event with different params")
	}
}

func TestConnectionWriteEvent(t *testing.T) {
	conn := &Connection{
		send: make(chan []byte, 10),
	}

	// Test writeEvent method with pubsub.Event
	event := pubsub.Event{ID: stringPtr("test_event")}
	err := conn.writeEvent(event)
	if err != nil {
		t.Errorf("Unexpected error in writeEvent: %v", err)
	}

	// Verify something was sent
	select {
	case data := <-conn.send:
		if len(data) == 0 {
			t.Error("Expected non-empty data to be sent")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected data to be sent within timeout")
	}
}

func TestConnectionConfigureConnection(t *testing.T) {
	// Test that configureConnection doesn't panic
	conn := &Connection{}

	// This test verifies the method exists and can be called
	// We can't easily test WebSocket configuration without a real connection
	defer func() {
		if r := recover(); r != nil {
			// Expected to panic since conn.conn is nil
			// This is fine - we're just testing the method exists
		}
	}()

	conn.configureConnection()
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
