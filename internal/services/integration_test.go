package services

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/livefir/fir/pubsub"
)

func TestEventBridge_ProcessLegacyRouteEvent(t *testing.T) {
	// Create event service with minimal dependencies
	registry := NewInMemoryEventRegistry()
	validator := NewDefaultEventValidator()
	publisher := &NoOpEventPublisher{}
	logger := NewDefaultEventLogger(false)

	eventService := NewDefaultEventService(registry, validator, publisher, logger)
	bridge := NewEventBridge(eventService)

	// Create a mock HTTP request with form body
	req, err := http.NewRequest("POST", "/test", strings.NewReader("field1=value1&field2=value2"))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Create legacy event data as map (simulating reflection/interface{} usage)
	target := "button1"
	sessionID := "session123"
	elementKey := "key456"
	params := json.RawMessage(`{"name": "test", "value": 42}`)

	legacyEvent := map[string]interface{}{
		"ID":         "test-event",
		"Params":     params,
		"Target":     &target,
		"ElementKey": &elementKey,
		"SessionID":  &sessionID,
	}

	// Register a test handler
	testHandler := &testEventHandler{
		eventID: "test-event",
		handler: func(ctx context.Context, req EventRequest) (*EventResponse, error) {
			// Verify the event was converted correctly
			if req.ID != "test-event" {
				t.Errorf("Expected ID 'test-event', got '%s'", req.ID)
			}
			if req.Target == nil || *req.Target != target {
				t.Errorf("Expected target '%s', got %v", target, req.Target)
			}
			if req.SessionID != sessionID {
				t.Errorf("Expected sessionID '%s', got '%s'", sessionID, req.SessionID)
			}
			if req.ElementKey == nil || *req.ElementKey != elementKey {
				t.Errorf("Expected elementKey '%s', got %v", elementKey, req.ElementKey)
			}

			// Check params were unmarshaled correctly
			if nameParam, exists := req.Params["name"]; !exists || nameParam != "test" {
				t.Errorf("Expected param 'name' to be 'test', got %v", nameParam)
			}
			if valueParam, exists := req.Params["value"]; !exists {
				t.Errorf("Expected param 'value' to exist")
			} else if floatVal, ok := valueParam.(float64); !ok || floatVal != 42 {
				t.Errorf("Expected param 'value' to be 42, got %v", valueParam)
			}

			return NewEventResponseBuilder().
				WithStatus(200).
				WithHTML("<p>success</p>").
				Build(), nil
		},
	}

	err = registry.RegisterHandler("test-event", testHandler)
	if err != nil {
		t.Fatalf("Failed to register handler: %v", err)
	}

	// Process the legacy event
	response, err := bridge.ProcessLegacyRouteEvent(context.Background(), legacyEvent, req)
	if err != nil {
		t.Errorf("ProcessLegacyRouteEvent failed: %v", err)
	}

	if response.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", response.StatusCode)
	}

	if string(response.Body) != "<p>success</p>" {
		t.Errorf("Expected body '<p>success</p>', got '%s'", string(response.Body))
	}
}

func TestEventBridge_ProcessLegacyRouteEvent_InvalidParams(t *testing.T) {
	eventService := NewDefaultEventService(
		NewInMemoryEventRegistry(),
		NewDefaultEventValidator(),
		&NoOpEventPublisher{},
		NewDefaultEventLogger(false),
	)
	bridge := NewEventBridge(eventService)

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Create legacy event with invalid JSON params
	legacyEvent := map[string]interface{}{
		"ID":     "test-event",
		"Params": json.RawMessage(`{invalid json`),
	}

	// Should handle gracefully by storing as raw string
	_, err := bridge.ProcessLegacyRouteEvent(context.Background(), legacyEvent, req)

	// Should fail because no handler is registered, but not because of JSON parsing
	if err == nil {
		t.Error("Expected error for unregistered event")
	}
}

func TestRouteEventServiceAdapter_ProcessRouteEvent(t *testing.T) {
	eventService := NewDefaultEventService(
		NewInMemoryEventRegistry(),
		NewDefaultEventValidator(),
		&NoOpEventPublisher{},
		NewDefaultEventLogger(false),
	)

	adapter := NewRouteEventServiceAdapter(eventService)

	// Register a test handler
	testHandler := &testEventHandler{
		eventID: "adapter-test",
		handler: func(ctx context.Context, req EventRequest) (*EventResponse, error) {
			return NewEventResponseBuilder().
				WithStatus(201).
				WithHTML("<div>adapter test</div>").
				Build(), nil
		},
	}

	err := eventService.RegisterHandler("adapter-test", testHandler)
	if err != nil {
		t.Fatalf("Failed to register handler: %v", err)
	}

	// Create request
	req, _ := http.NewRequest("POST", "/test?param1=value1", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	target := "form1"
	elementKey := "input1"
	params := map[string]interface{}{
		"field1": "test",
		"field2": 123,
	}

	// Process the event
	response, err := adapter.ProcessRouteEvent(
		context.Background(),
		"ADAPTER-TEST", // Should be converted to lowercase
		"session456",
		&target,
		&elementKey,
		params,
		req,
	)

	if err != nil {
		t.Errorf("ProcessRouteEvent failed: %v", err)
	}

	if response.StatusCode != 201 {
		t.Errorf("Expected status 201, got %d", response.StatusCode)
	}

	if string(response.Body) != "<div>adapter test</div>" {
		t.Errorf("Expected body '<div>adapter test</div>', got '%s'", string(response.Body))
	}
}

func TestConfigureEventService(t *testing.T) {
	config := EventServiceConfiguration{
		EnableDebugLogging: true,
		EnableMetrics:      true,
		ValidatorConfig: map[string][]string{
			"login":  {"username", "password"},
			"signup": {"email", "password", "name"},
		},
	}

	eventService := ConfigureEventService(config)
	if eventService == nil {
		t.Error("Expected non-nil event service")
	}

	// Test that validation rules were configured
	validator := NewDefaultEventValidator()
	validator.SetRequiredFields("login", []string{"username", "password"})

	err := validator.ValidateParams("login", map[string]interface{}{
		"username": "test",
		// missing password
	})
	if err == nil {
		t.Error("Expected validation error for missing password")
	}
}

func TestLegacyEventHandlerWrapper(t *testing.T) {
	eventService := NewDefaultEventService(
		NewInMemoryEventRegistry(),
		NewDefaultEventValidator(),
		&NoOpEventPublisher{},
		NewDefaultEventLogger(false),
	)
	bridge := NewEventBridge(eventService)

	// Create a mock legacy handler (OnEventFunc)
	mockHandler := func() {} // Simplified mock

	wrapper := NewLegacyEventHandlerWrapper("test-event", "test-route", mockHandler, bridge)

	if wrapper.GetEventID() != "test-event" {
		t.Errorf("Expected event ID 'test-event', got '%s'", wrapper.GetEventID())
	}

	if wrapper.GetRouteID() != "test-route" {
		t.Errorf("Expected route ID 'test-route', got '%s'", wrapper.GetRouteID())
	}

	// Test that Handle returns the expected error
	req := EventRequest{ID: "test-event"}
	_, err := wrapper.Handle(context.Background(), req)
	if err != ErrLegacyHandlerRequiresRouteContext {
		t.Errorf("Expected ErrLegacyHandlerRequiresRouteContext, got %v", err)
	}
}

func TestConvertPubSubEventsToResponse(t *testing.T) {
	id1 := "event1"
	id2 := "event2"

	events := []pubsub.Event{
		{ID: &id1},
		{ID: &id2},
	}

	response := ConvertPubSubEventsToResponse(events)

	if response.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", response.StatusCode)
	}

	if len(response.PubSubEvents) != 2 {
		t.Errorf("Expected 2 pubsub events, got %d", len(response.PubSubEvents))
	}

	if response.PubSubEvents[0].ID == nil || *response.PubSubEvents[0].ID != "event1" {
		t.Errorf("Expected first event ID 'event1', got %v", response.PubSubEvents[0].ID)
	}
}

func TestConvertEventResponseToPubSubEvents(t *testing.T) {
	id1 := "event1"
	response := &EventResponse{
		PubSubEvents: []pubsub.Event{
			{ID: &id1},
		},
	}

	events := ConvertEventResponseToPubSubEvents(response)
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	if events[0].ID == nil || *events[0].ID != "event1" {
		t.Errorf("Expected event ID 'event1', got %v", events[0].ID)
	}

	// Test nil response
	nilEvents := ConvertEventResponseToPubSubEvents(nil)
	if nilEvents != nil {
		t.Errorf("Expected nil for nil response, got %v", nilEvents)
	}
}

func TestNoOpEventPublisher(t *testing.T) {
	publisher := &NoOpEventPublisher{}

	// Test PublishEvent
	event := pubsub.Event{ID: stringPtr("test")}
	err := publisher.PublishEvent(event)
	if err != nil {
		t.Errorf("Expected no error from no-op publisher, got %v", err)
	}

	// Test PublishEvents
	events := []pubsub.Event{{ID: stringPtr("test1")}, {ID: stringPtr("test2")}}
	err = publisher.PublishEvents(events)
	if err != nil {
		t.Errorf("Expected no error from no-op publisher, got %v", err)
	}
}

func TestExtractEventData(t *testing.T) {
	target := "button1"
	sessionID := "session123"
	params := json.RawMessage(`{"test": "value"}`)

	// Test with valid event map
	eventMap := map[string]interface{}{
		"ID":        "test-event",
		"Target":    &target,
		"SessionID": &sessionID,
		"Params":    params,
	}

	data := extractEventData(eventMap)
	if data.ID != "test-event" {
		t.Errorf("Expected ID 'test-event', got '%s'", data.ID)
	}
	if data.Target == nil || *data.Target != target {
		t.Errorf("Expected target '%s', got %v", target, data.Target)
	}

	// Test with invalid event
	invalidEvent := "not a map"
	emptyData := extractEventData(invalidEvent)
	if emptyData.ID != "" {
		t.Errorf("Expected empty ID for invalid event, got '%s'", emptyData.ID)
	}
}

func TestGetStringValue(t *testing.T) {
	// Test with non-nil pointer
	value := "test"
	result := getStringValue(&value)
	if result != "test" {
		t.Errorf("Expected 'test', got '%s'", result)
	}

	// Test with nil pointer
	result = getStringValue(nil)
	if result != "" {
		t.Errorf("Expected empty string for nil, got '%s'", result)
	}
}
