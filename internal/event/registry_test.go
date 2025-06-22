package event

import (
	"fmt"
	"sync"
	"testing"
)

// Mock OnEventFunc for testing
type mockOnEventFunc func(ctx interface{}) error

func mockHandler() mockOnEventFunc {
	return func(ctx interface{}) error {
		return nil
	}
}

func TestNewEventRegistry(t *testing.T) {
	registry := NewEventRegistry()
	if registry == nil {
		t.Fatal("NewEventRegistry() returned nil")
	}

	// Verify empty registry
	allEvents := registry.GetAllEvents()
	if len(allEvents) != 0 {
		t.Errorf("Expected empty registry, got %d routes", len(allEvents))
	}
}

func TestEventRegistry_Register(t *testing.T) {
	registry := NewEventRegistry()

	tests := []struct {
		name    string
		routeID string
		eventID string
		handler interface{}
		wantErr bool
	}{
		{
			name:    "valid registration",
			routeID: "route1",
			eventID: "event1",
			handler: mockHandler(),
			wantErr: false,
		},
		{
			name:    "empty routeID",
			routeID: "",
			eventID: "event1",
			handler: mockHandler(),
			wantErr: true,
		},
		{
			name:    "empty eventID",
			routeID: "route1",
			eventID: "",
			handler: mockHandler(),
			wantErr: true,
		},
		{
			name:    "nil handler",
			routeID: "route1",
			eventID: "event1",
			handler: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.Register(tt.routeID, tt.eventID, tt.handler)
			if (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEventRegistry_Get(t *testing.T) {
	registry := NewEventRegistry()
	handler := mockHandler()

	// Register test event
	err := registry.Register("route1", "event1", handler)
	if err != nil {
		t.Fatalf("Failed to register event: %v", err)
	}

	tests := []struct {
		name       string
		routeID    string
		eventID    string
		wantExists bool
	}{
		{
			name:       "existing event",
			routeID:    "route1",
			eventID:    "event1",
			wantExists: true,
		},
		{
			name:       "non-existing route",
			routeID:    "route2",
			eventID:    "event1",
			wantExists: false,
		},
		{
			name:       "non-existing event",
			routeID:    "route1",
			eventID:    "event2",
			wantExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHandler, exists := registry.Get(tt.routeID, tt.eventID)
			if exists != tt.wantExists {
				t.Errorf("Get() exists = %v, want %v", exists, tt.wantExists)
			}
			if tt.wantExists && gotHandler == nil {
				t.Error("Get() returned nil handler for existing event")
			}
		})
	}
}

func TestEventRegistry_GetRouteEvents(t *testing.T) {
	registry := NewEventRegistry()
	handler1 := mockHandler()
	handler2 := mockHandler()

	// Register test events
	registry.Register("route1", "event1", handler1)
	registry.Register("route1", "event2", handler2)
	registry.Register("route2", "event1", handler1)

	tests := []struct {
		name           string
		routeID        string
		expectedEvents int
	}{
		{
			name:           "route with events",
			routeID:        "route1",
			expectedEvents: 2,
		},
		{
			name:           "route with one event",
			routeID:        "route2",
			expectedEvents: 1,
		},
		{
			name:           "non-existing route",
			routeID:        "route3",
			expectedEvents: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events := registry.GetRouteEvents(tt.routeID)
			if len(events) != tt.expectedEvents {
				t.Errorf("GetRouteEvents() returned %d events, want %d", len(events), tt.expectedEvents)
			}
		})
	}
}

func TestEventRegistry_GetAllEvents(t *testing.T) {
	registry := NewEventRegistry()
	handler := mockHandler()

	// Register test events
	registry.Register("route1", "event1", handler)
	registry.Register("route1", "event2", handler)
	registry.Register("route2", "event1", handler)

	allEvents := registry.GetAllEvents()

	if len(allEvents) != 2 {
		t.Errorf("GetAllEvents() returned %d routes, want 2", len(allEvents))
	}

	if len(allEvents["route1"]) != 2 {
		t.Errorf("GetAllEvents() route1 has %d events, want 2", len(allEvents["route1"]))
	}

	if len(allEvents["route2"]) != 1 {
		t.Errorf("GetAllEvents() route2 has %d events, want 1", len(allEvents["route2"]))
	}
}

func TestEventRegistry_Remove(t *testing.T) {
	registry := NewEventRegistry()
	handler := mockHandler()

	// Register test events
	registry.Register("route1", "event1", handler)
	registry.Register("route1", "event2", handler)

	tests := []struct {
		name         string
		routeID      string
		eventID      string
		wantRemoved  bool
		expectEvents int // Expected events remaining for route1
	}{
		{
			name:         "remove existing event",
			routeID:      "route1",
			eventID:      "event1",
			wantRemoved:  true,
			expectEvents: 1,
		},
		{
			name:         "remove non-existing event",
			routeID:      "route1",
			eventID:      "event3",
			wantRemoved:  false,
			expectEvents: 1,
		},
		{
			name:         "remove from non-existing route",
			routeID:      "route2",
			eventID:      "event1",
			wantRemoved:  false,
			expectEvents: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			removed := registry.Remove(tt.routeID, tt.eventID)
			if removed != tt.wantRemoved {
				t.Errorf("Remove() = %v, want %v", removed, tt.wantRemoved)
			}

			// Check remaining events for route1
			events := registry.GetRouteEvents("route1")
			if len(events) != tt.expectEvents {
				t.Errorf("After Remove(), route1 has %d events, want %d", len(events), tt.expectEvents)
			}
		})
	}
}

func TestEventRegistry_RemoveRoute(t *testing.T) {
	registry := NewEventRegistry()
	handler := mockHandler()

	// Register test events
	registry.Register("route1", "event1", handler)
	registry.Register("route1", "event2", handler)
	registry.Register("route2", "event1", handler)

	tests := []struct {
		name        string
		routeID     string
		wantRemoved bool
	}{
		{
			name:        "remove existing route",
			routeID:     "route1",
			wantRemoved: true,
		},
		{
			name:        "remove non-existing route",
			routeID:     "route3",
			wantRemoved: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			removed := registry.RemoveRoute(tt.routeID)
			if removed != tt.wantRemoved {
				t.Errorf("RemoveRoute() = %v, want %v", removed, tt.wantRemoved)
			}
		})
	}

	// Verify route1 is removed but route2 remains
	allEvents := registry.GetAllEvents()
	if _, exists := allEvents["route1"]; exists {
		t.Error("route1 should be removed")
	}
	if _, exists := allEvents["route2"]; !exists {
		t.Error("route2 should still exist")
	}
}

func TestEventRegistry_ListRoutes(t *testing.T) {
	registry := NewEventRegistry()
	handler := mockHandler()

	// Start with empty registry
	routes := registry.ListRoutes()
	if len(routes) != 0 {
		t.Errorf("ListRoutes() on empty registry returned %d routes, want 0", len(routes))
	}

	// Register events for different routes
	registry.Register("route1", "event1", handler)
	registry.Register("route2", "event1", handler)
	registry.Register("route1", "event2", handler) // Same route, different event

	routes = registry.ListRoutes()
	if len(routes) != 2 {
		t.Errorf("ListRoutes() returned %d routes, want 2", len(routes))
	}

	// Check that both routes are present (order doesn't matter)
	routeMap := make(map[string]bool)
	for _, route := range routes {
		routeMap[route] = true
	}

	if !routeMap["route1"] || !routeMap["route2"] {
		t.Errorf("ListRoutes() missing expected routes, got %v", routes)
	}
}

func TestEventRegistry_ListEventIDs(t *testing.T) {
	registry := NewEventRegistry()
	handler := mockHandler()

	// Register test events
	registry.Register("route1", "event1", handler)
	registry.Register("route1", "event2", handler)
	registry.Register("route2", "event1", handler)

	tests := []struct {
		name           string
		routeID        string
		expectedEvents []string
	}{
		{
			name:           "route with multiple events",
			routeID:        "route1",
			expectedEvents: []string{"event1", "event2"},
		},
		{
			name:           "route with single event",
			routeID:        "route2",
			expectedEvents: []string{"event1"},
		},
		{
			name:           "non-existing route",
			routeID:        "route3",
			expectedEvents: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventIDs := registry.ListEventIDs(tt.routeID)
			if len(eventIDs) != len(tt.expectedEvents) {
				t.Errorf("ListEventIDs() returned %d events, want %d", len(eventIDs), len(tt.expectedEvents))
				return
			}

			// Check that all expected events are present (order doesn't matter)
			eventMap := make(map[string]bool)
			for _, eventID := range eventIDs {
				eventMap[eventID] = true
			}

			for _, expectedEvent := range tt.expectedEvents {
				if !eventMap[expectedEvent] {
					t.Errorf("ListEventIDs() missing expected event %s", expectedEvent)
				}
			}
		})
	}
}

// TestEventRegistry_ConcurrentAccess tests thread safety
func TestEventRegistry_ConcurrentAccess(t *testing.T) {
	registry := NewEventRegistry()
	handler := mockHandler()

	const numGoroutines = 10
	const numOperationsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 3) // 3 types of operations

	// Concurrent Register operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperationsPerGoroutine; j++ {
				registry.Register(fmt.Sprintf("route%d", id), fmt.Sprintf("event%d", j), handler)
			}
		}(i)
	}

	// Concurrent Get operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperationsPerGoroutine; j++ {
				registry.Get(fmt.Sprintf("route%d", id), fmt.Sprintf("event%d", j))
			}
		}(i)
	}

	// Concurrent introspection operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperationsPerGoroutine; j++ {
				registry.GetAllEvents()
				registry.ListRoutes()
				registry.ListEventIDs(fmt.Sprintf("route%d", id))
			}
		}(i)
	}

	wg.Wait()

	// Verify final state
	allEvents := registry.GetAllEvents()
	if len(allEvents) != numGoroutines {
		t.Errorf("Expected %d routes after concurrent operations, got %d", numGoroutines, len(allEvents))
	}

	for i := 0; i < numGoroutines; i++ {
		routeID := fmt.Sprintf("route%d", i)
		events := registry.GetRouteEvents(routeID)
		if len(events) != numOperationsPerGoroutine {
			t.Errorf("Route %s should have %d events, got %d", routeID, numOperationsPerGoroutine, len(events))
		}
	}
}

// TestEventRegistry_CleanupEmptyRoutes tests that empty route maps are cleaned up
func TestEventRegistry_CleanupEmptyRoutes(t *testing.T) {
	registry := NewEventRegistry()
	handler := mockHandler()

	// Register and then remove all events for a route
	registry.Register("route1", "event1", handler)
	registry.Register("route1", "event2", handler)

	// Verify route exists
	allEvents := registry.GetAllEvents()
	if len(allEvents) != 1 {
		t.Errorf("Expected 1 route, got %d", len(allEvents))
	}

	// Remove all events
	registry.Remove("route1", "event1")
	registry.Remove("route1", "event2")

	// Verify route is cleaned up
	allEvents = registry.GetAllEvents()
	if len(allEvents) != 0 {
		t.Errorf("Expected empty registry after removing all events, got %d routes", len(allEvents))
	}
}
