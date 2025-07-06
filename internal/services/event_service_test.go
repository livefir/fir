package services

import (
	"context"
	"fmt"
	"testing"

	firHttp "github.com/livefir/fir/internal/http"
	"github.com/livefir/fir/pubsub"
)

// mockEventPublisher for testing
type mockEventPublisher struct {
	published []pubsub.Event
}

func (m *mockEventPublisher) PublishEvent(event pubsub.Event) error {
	m.published = append(m.published, event)
	return nil
}

func (m *mockEventPublisher) PublishEvents(events []pubsub.Event) error {
	m.published = append(m.published, events...)
	return nil
}

// testEventHandler implements EventHandler for testing
type testEventHandler struct {
	eventID string
	handler func(ctx context.Context, req EventRequest) (*EventResponse, error)
}

func (h *testEventHandler) Handle(ctx context.Context, req EventRequest) (*EventResponse, error) {
	return h.handler(ctx, req)
}

func (h *testEventHandler) GetEventID() string {
	return h.eventID
}

func TestDefaultEventService_ProcessEvent(t *testing.T) {
	registry := NewInMemoryEventRegistry()
	validator := NewDefaultEventValidator()
	publisher := &mockEventPublisher{}
	logger := NewDefaultEventLogger(false) // Disable debug for tests

	service := NewDefaultEventService(registry, validator, publisher, logger)

	// Register a test handler
	handlerCalled := false
	handler := &testEventHandler{
		eventID: "test-event",
		handler: func(ctx context.Context, req EventRequest) (*EventResponse, error) {
			handlerCalled = true
			return NewEventResponseBuilder().
				WithStatus(200).
				WithHTML("<p>test response</p>").
				Build(), nil
		},
	}

	err := registry.RegisterHandler("test-event", handler)
	if err != nil {
		t.Fatalf("Failed to register handler: %v", err)
	}

	// Test processing an event
	req := EventRequest{
		ID:           "test-event",
		SessionID:    "session-123",
		Context:      context.Background(),
		RequestModel: &firHttp.RequestModel{},
		Params:       make(map[string]interface{}),
	}

	resp, err := service.ProcessEvent(context.Background(), req)
	if err != nil {
		t.Errorf("ProcessEvent() error = %v", err)
	}

	if !handlerCalled {
		t.Error("Handler was not called")
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	if string(resp.Body) != "<p>test response</p>" {
		t.Errorf("Expected body '<p>test response</p>', got '%s'", string(resp.Body))
	}
}

func TestDefaultEventService_ProcessEvent_HandlerNotFound(t *testing.T) {
	registry := NewInMemoryEventRegistry()
	validator := NewDefaultEventValidator()
	publisher := &mockEventPublisher{}
	logger := NewDefaultEventLogger(false)

	service := NewDefaultEventService(registry, validator, publisher, logger)

	req := EventRequest{
		ID:           "unknown-event",
		SessionID:    "session-123",
		Context:      context.Background(),
		RequestModel: &firHttp.RequestModel{},
		Params:       make(map[string]interface{}),
	}

	_, err := service.ProcessEvent(context.Background(), req)
	if err == nil {
		t.Error("Expected error for unknown event")
	}

	// Check that it's an EventError
	eventErr, ok := err.(*EventError)
	if !ok {
		t.Errorf("Expected EventError, got %T", err)
	} else {
		if eventErr.EventID != "unknown-event" {
			t.Errorf("Expected EventID 'unknown-event', got '%s'", eventErr.EventID)
		}
		if eventErr.Code != "HANDLER_NOT_FOUND" {
			t.Errorf("Expected Code 'HANDLER_NOT_FOUND', got '%s'", eventErr.Code)
		}
	}
}

func TestDefaultEventService_ProcessEvent_ValidationError(t *testing.T) {
	registry := NewInMemoryEventRegistry()
	validator := NewDefaultEventValidator()
	publisher := &mockEventPublisher{}
	logger := NewDefaultEventLogger(false)

	service := NewDefaultEventService(registry, validator, publisher, logger)

	// Register a handler
	handler := &testEventHandler{
		eventID: "test-event",
		handler: func(ctx context.Context, req EventRequest) (*EventResponse, error) {
			return NewEventResponseBuilder().Build(), nil
		},
	}
	registry.RegisterHandler("test-event", handler)

	// Create invalid request (missing context)
	req := EventRequest{
		ID:           "test-event",
		SessionID:    "session-123",
		RequestModel: &firHttp.RequestModel{},
		Params:       make(map[string]interface{}),
		// Context is nil - should fail validation
	}

	_, err := service.ProcessEvent(context.Background(), req)
	if err == nil {
		t.Error("Expected validation error")
	}

	// Check that it's an EventError
	eventErr, ok := err.(*EventError)
	if !ok {
		t.Errorf("Expected EventError, got %T", err)
	} else {
		if eventErr.Code != "VALIDATION_ERROR" {
			t.Errorf("Expected Code 'VALIDATION_ERROR', got '%s'", eventErr.Code)
		}
	}
}

func TestDefaultEventService_ProcessEvent_HandlerError(t *testing.T) {
	registry := NewInMemoryEventRegistry()
	validator := NewDefaultEventValidator()
	publisher := &mockEventPublisher{}
	logger := NewDefaultEventLogger(false)

	service := NewDefaultEventService(registry, validator, publisher, logger)

	// Register a handler that returns an error
	expectedErr := fmt.Errorf("handler error")
	handler := &testEventHandler{
		eventID: "test-event",
		handler: func(ctx context.Context, req EventRequest) (*EventResponse, error) {
			return nil, expectedErr
		},
	}

	registry.RegisterHandler("test-event", handler)

	req := EventRequest{
		ID:           "test-event",
		SessionID:    "session-123",
		Context:      context.Background(),
		RequestModel: &firHttp.RequestModel{},
		Params:       make(map[string]interface{}),
	}

	_, err := service.ProcessEvent(context.Background(), req)
	if err == nil {
		t.Error("Expected error from handler")
	}

	// Check that it's an EventError
	eventErr, ok := err.(*EventError)
	if !ok {
		t.Errorf("Expected EventError, got %T", err)
	} else {
		if eventErr.Code != "PROCESSING_ERROR" {
			t.Errorf("Expected Code 'PROCESSING_ERROR', got '%s'", eventErr.Code)
		}
	}
}

func TestInMemoryEventRegistry_RegisterHandler(t *testing.T) {
	registry := NewInMemoryEventRegistry()

	handler := &testEventHandler{
		eventID: "test-event",
		handler: func(ctx context.Context, req EventRequest) (*EventResponse, error) {
			return nil, nil
		},
	}

	err := registry.RegisterHandler("test-event", handler)
	if err != nil {
		t.Errorf("RegisterHandler() error = %v", err)
	}

	// Test registering with empty ID
	err = registry.RegisterHandler("", handler)
	if err == nil {
		t.Error("Expected error when registering with empty event ID")
	}

	// Test registering nil handler
	err = registry.RegisterHandler("test-event", nil)
	if err == nil {
		t.Error("Expected error when registering nil handler")
	}
}

func TestInMemoryEventRegistry_GetHandler(t *testing.T) {
	registry := NewInMemoryEventRegistry()

	handler := &testEventHandler{
		eventID: "test-event",
		handler: func(ctx context.Context, req EventRequest) (*EventResponse, error) {
			return nil, nil
		},
	}

	// Test getting non-existent handler
	_, exists := registry.GetHandler("unknown-event")
	if exists {
		t.Error("Expected false for unknown event")
	}

	// Register and get handler
	registry.RegisterHandler("test-event", handler)

	retrievedHandler, exists := registry.GetHandler("test-event")
	if !exists {
		t.Error("Expected true for registered event")
	}

	if retrievedHandler == nil {
		t.Error("Expected non-nil handler")
	}

	if retrievedHandler.GetEventID() != "test-event" {
		t.Errorf("Expected event ID 'test-event', got '%s'", retrievedHandler.GetEventID())
	}
}

func TestInMemoryEventRegistry_ListHandlers(t *testing.T) {
	registry := NewInMemoryEventRegistry()

	// Initially empty
	eventIDs := registry.ListHandlers()
	if len(eventIDs) != 0 {
		t.Errorf("Expected 0 event IDs, got %d", len(eventIDs))
	}

	// Register some handlers
	handler1 := &testEventHandler{eventID: "event1"}
	handler2 := &testEventHandler{eventID: "event2"}
	handler3 := &testEventHandler{eventID: "event3"}

	registry.RegisterHandler("event1", handler1)
	registry.RegisterHandler("event2", handler2)
	registry.RegisterHandler("event3", handler3)

	eventIDs = registry.ListHandlers()
	if len(eventIDs) != 3 {
		t.Errorf("Expected 3 event IDs, got %d", len(eventIDs))
	}

	// Check that all expected IDs are present
	expectedIDs := map[string]bool{
		"event1": false,
		"event2": false,
		"event3": false,
	}

	for _, id := range eventIDs {
		if _, exists := expectedIDs[id]; exists {
			expectedIDs[id] = true
		} else {
			t.Errorf("Unexpected event ID: %s", id)
		}
	}

	for id, found := range expectedIDs {
		if !found {
			t.Errorf("Expected event ID '%s' not found", id)
		}
	}
}

func TestEventServiceMetrics(t *testing.T) {
	registry := NewInMemoryEventRegistry()
	validator := NewDefaultEventValidator()
	publisher := &mockEventPublisher{}
	logger := NewDefaultEventLogger(false)

	service := NewDefaultEventService(registry, validator, publisher, logger)

	// Register a successful handler
	successHandler := &testEventHandler{
		eventID: "success-event",
		handler: func(ctx context.Context, req EventRequest) (*EventResponse, error) {
			return NewEventResponseBuilder().Build(), nil
		},
	}
	registry.RegisterHandler("success-event", successHandler)

	// Register a failing handler
	failHandler := &testEventHandler{
		eventID: "fail-event",
		handler: func(ctx context.Context, req EventRequest) (*EventResponse, error) {
			return nil, fmt.Errorf("handler error")
		},
	}
	registry.RegisterHandler("fail-event", failHandler)

	req := EventRequest{
		Context:      context.Background(),
		RequestModel: &firHttp.RequestModel{},
		Params:       make(map[string]interface{}),
	}

	// Test successful event
	req.ID = "success-event"
	_, err := service.ProcessEvent(context.Background(), req)
	if err != nil {
		t.Errorf("Expected no error for successful event, got %v", err)
	}

	// Test failing event
	req.ID = "fail-event"
	_, err = service.ProcessEvent(context.Background(), req)
	if err == nil {
		t.Error("Expected error for failing event")
	}

	// Check metrics
	metrics := service.GetEventMetrics()
	if metrics.TotalEvents != 2 {
		t.Errorf("Expected TotalEvents to be 2, got %d", metrics.TotalEvents)
	}
	if metrics.SuccessfulEvents != 1 {
		t.Errorf("Expected SuccessfulEvents to be 1, got %d", metrics.SuccessfulEvents)
	}
	if metrics.FailedEvents != 1 {
		t.Errorf("Expected FailedEvents to be 1, got %d", metrics.FailedEvents)
	}
	if metrics.AverageLatency <= 0 {
		t.Errorf("Expected positive average latency, got %f", metrics.AverageLatency)
	}
}
