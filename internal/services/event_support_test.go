package services

import (
	"context"
	"fmt"
	"testing"

	firHttp "github.com/livefir/fir/internal/http"
	"github.com/livefir/fir/pubsub"
)

func TestDefaultEventValidator_ValidateEvent(t *testing.T) {
	validator := NewDefaultEventValidator()

	tests := []struct {
		name    string
		req     EventRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid event",
			req: EventRequest{
				ID:           "test-event",
				Context:      context.Background(),
				RequestModel: &firHttp.RequestModel{},
				Params:       make(map[string]interface{}),
			},
			wantErr: false,
		},
		{
			name: "missing event ID",
			req: EventRequest{
				Context:      context.Background(),
				RequestModel: &firHttp.RequestModel{},
				Params:       make(map[string]interface{}),
			},
			wantErr: true,
			errMsg:  "event ID is required",
		},
		{
			name: "missing context",
			req: EventRequest{
				ID:           "test-event",
				RequestModel: &firHttp.RequestModel{},
				Params:       make(map[string]interface{}),
			},
			wantErr: true,
			errMsg:  "context is required",
		},
		{
			name: "missing request model",
			req: EventRequest{
				ID:      "test-event",
				Context: context.Background(),
				Params:  make(map[string]interface{}),
			},
			wantErr: true,
			errMsg:  "request model is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateEvent(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEvent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("ValidateEvent() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestDefaultEventValidator_ValidateParams(t *testing.T) {
	validator := NewDefaultEventValidator()
	validator.SetRequiredFields("test-event", []string{"field1", "field2"})

	tests := []struct {
		name    string
		eventID string
		params  map[string]interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid params",
			eventID: "test-event",
			params: map[string]interface{}{
				"field1": "value1",
				"field2": "value2",
			},
			wantErr: false,
		},
		{
			name:    "missing required field",
			eventID: "test-event",
			params: map[string]interface{}{
				"field1": "value1",
			},
			wantErr: true,
			errMsg:  "required field 'field2' is missing for event 'test-event'",
		},
		{
			name:    "no validation rules",
			eventID: "unknown-event",
			params:  map[string]interface{}{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateParams(tt.eventID, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("ValidateParams() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestDefaultEventValidator_SetRequiredFields(t *testing.T) {
	validator := NewDefaultEventValidator()
	
	fields := []string{"field1", "field2"}
	validator.SetRequiredFields("test-event", fields)

	// Test that the fields were set correctly
	params := map[string]interface{}{
		"field1": "value1",
		// missing field2
	}

	err := validator.ValidateParams("test-event", params)
	if err == nil {
		t.Error("Expected validation error for missing field2")
	}
}

func TestDefaultEventLogger(t *testing.T) {
	// Test with debug enabled
	logger := NewDefaultEventLogger(true)
	req := EventRequest{
		ID:        "test-event",
		SessionID: "session-123",
		Context:   context.Background(),
	}

	// These methods should not panic and should work with debug mode
	logger.LogEventStart(context.Background(), req)
	
	resp := &EventResponse{
		StatusCode:   200,
		Events:       make([]firHttp.DOMEvent, 1),
		PubSubEvents: make([]pubsub.Event, 2),
	}
	logger.LogEventSuccess(context.Background(), req, resp)
	
	logger.LogEventError(context.Background(), req, fmt.Errorf("test error"))

	// Test with debug disabled
	logger = NewDefaultEventLogger(false)
	logger.LogEventStart(context.Background(), req)
	logger.LogEventSuccess(context.Background(), req, resp)
	logger.LogEventError(context.Background(), req, fmt.Errorf("test error"))
}

func TestRouteEventHandler(t *testing.T) {
	handlerCalled := false
	handler := func(ctx RouteContext) error {
		handlerCalled = true
		// Verify the context has the expected event
		event := ctx.Event()
		if event.ID != "test-event" {
			t.Errorf("Expected event ID 'test-event', got '%s'", event.ID)
		}
		if event.SessionID != "session-123" {
			t.Errorf("Expected session ID 'session-123', got '%s'", event.SessionID)
		}
		return nil
	}

	eventHandler := NewRouteEventHandler("test-event", handler)

	target := "button"
	elementKey := "key123"
	req := EventRequest{
		ID:         "test-event",
		Target:     &target,
		ElementKey: &elementKey,
		SessionID:  "session-123",
		Context:    context.Background(),
		Params:     make(map[string]interface{}),
	}

	resp, err := eventHandler.Handle(context.Background(), req)
	if err != nil {
		t.Errorf("Handle() error = %v", err)
	}

	if !handlerCalled {
		t.Error("Handler was not called")
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	if eventHandler.GetEventID() != "test-event" {
		t.Errorf("Expected event ID 'test-event', got '%s'", eventHandler.GetEventID())
	}
}

func TestRouteEventHandler_Error(t *testing.T) {
	expectedErr := fmt.Errorf("handler error")
	handler := func(ctx RouteContext) error {
		return expectedErr
	}

	eventHandler := NewRouteEventHandler("test-event", handler)

	req := EventRequest{
		ID:        "test-event",
		SessionID: "session-123",
		Context:   context.Background(),
		Params:    make(map[string]interface{}),
	}

	resp, err := eventHandler.Handle(context.Background(), req)
	if err != expectedErr {
		t.Errorf("Handle() error = %v, want %v", err, expectedErr)
	}

	if resp != nil {
		t.Error("Expected nil response on error")
	}
}

func TestEventResponseBuilder(t *testing.T) {
	builder := NewEventResponseBuilder()

	event := firHttp.DOMEvent{
		ID:     "test-event",
		Type:   "update",
		Target: "div",
	}

	pubsubEvent := pubsub.Event{
		ID: stringPtr("pubsub-event"),
	}

	resp := builder.
		WithStatus(201).
		WithHeader("Content-Type", "application/json").
		WithBody([]byte("test body")).
		WithHTML("<p>test</p>").
		WithEvent(event).
		WithPubSubEvent(pubsubEvent).
		WithRedirect("/test", 302).
		WithError("field1", "error message").
		Build()

	if resp.StatusCode != 201 {
		t.Errorf("Expected status code 201, got %d", resp.StatusCode)
	}

	if resp.Headers["Content-Type"] != "text/html; charset=utf-8" {
		t.Errorf("Expected Content-Type to be overridden by WithHTML")
	}

	if string(resp.Body) != "<p>test</p>" {
		t.Errorf("Expected body '<p>test</p>', got '%s'", string(resp.Body))
	}

	if len(resp.Events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(resp.Events))
	}

	if len(resp.PubSubEvents) != 1 {
		t.Errorf("Expected 1 pubsub event, got %d", len(resp.PubSubEvents))
	}

	if resp.Redirect == nil {
		t.Error("Expected redirect to be set")
	} else {
		if resp.Redirect.URL != "/test" {
			t.Errorf("Expected redirect URL '/test', got '%s'", resp.Redirect.URL)
		}
		if resp.Redirect.StatusCode != 302 {
			t.Errorf("Expected redirect status 302, got %d", resp.Redirect.StatusCode)
		}
	}

	if resp.Errors["field1"] != "error message" {
		t.Errorf("Expected error message 'error message', got '%v'", resp.Errors["field1"])
	}
}

func TestEventRequestBuilder(t *testing.T) {
	builder := NewEventRequestBuilder()

	target := "button"
	elementKey := "key123"
	requestModel := &firHttp.RequestModel{}

	req := builder.
		WithID("test-event").
		WithTarget(target).
		WithElementKey(elementKey).
		WithSessionID("session-123").
		WithContext(context.Background()).
		WithParam("param1", "value1").
		WithParam("param2", 42).
		WithRequestModel(requestModel).
		Build()

	if req.ID != "test-event" {
		t.Errorf("Expected ID 'test-event', got '%s'", req.ID)
	}

	if req.Target == nil || *req.Target != target {
		t.Errorf("Expected target '%s', got %v", target, req.Target)
	}

	if req.ElementKey == nil || *req.ElementKey != elementKey {
		t.Errorf("Expected element key '%s', got %v", elementKey, req.ElementKey)
	}

	if req.SessionID != "session-123" {
		t.Errorf("Expected session ID 'session-123', got '%s'", req.SessionID)
	}

	if req.Context == nil {
		t.Error("Expected context to be set")
	}

	if req.Params["param1"] != "value1" {
		t.Errorf("Expected param1 'value1', got '%v'", req.Params["param1"])
	}

	if req.Params["param2"] != 42 {
		t.Errorf("Expected param2 42, got '%v'", req.Params["param2"])
	}

	if req.RequestModel != requestModel {
		t.Error("Expected request model to be set")
	}
}

func TestMockRouteContext(t *testing.T) {
	event := Event{
		ID:         "test-event",
		SessionID:  "session-123",
		Target:     stringPtr("button"),
		ElementKey: stringPtr("key123"),
	}

	ctx := &mockRouteContext{
		event:  event,
		params: map[string]interface{}{"key": "value"},
	}

	resultEvent := ctx.Event()
	if resultEvent.ID != event.ID {
		t.Errorf("Expected event ID '%s', got '%s'", event.ID, resultEvent.ID)
	}

	if resultEvent.SessionID != event.SessionID {
		t.Errorf("Expected session ID '%s', got '%s'", event.SessionID, resultEvent.SessionID)
	}

	// Test Bind method (currently a no-op)
	err := ctx.Bind(&struct{}{})
	if err != nil {
		t.Errorf("Bind() error = %v", err)
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
