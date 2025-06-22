package fir

import (
	"testing"
)

func TestEventRegistry_Integration(t *testing.T) {
	// Create a controller with EventRegistry
	controller := NewController("test-registry")

	// Verify EventRegistry is initialized
	registry := controller.GetEventRegistry()
	if registry == nil {
		t.Fatal("EventRegistry should be initialized")
	}

	// Test empty registry
	allEvents := registry.GetAllEvents()
	if len(allEvents) != 0 {
		t.Errorf("Expected empty registry, got %d routes", len(allEvents))
	}

	// Create a route with events
	routeOpts := func() RouteOptions {
		return RouteOptions{
			ID("test-route"),
			Content("Hello World"),
			OnEvent("click", func(ctx RouteContext) error {
				return nil
			}),
			OnEvent("load", func(ctx RouteContext) error {
				return nil
			}),
			OnLoad(func(ctx RouteContext) error {
				return nil
			}),
		}
	}

	// Register the route
	_ = controller.RouteFunc(routeOpts)

	// Verify events are registered in the EventRegistry
	allEvents = registry.GetAllEvents()
	if len(allEvents) != 1 {
		t.Errorf("Expected 1 route in registry, got %d", len(allEvents))
	}

	routeEvents, exists := allEvents["test-route"]
	if !exists {
		t.Error("test-route should exist in registry")
	}

	// Should have click, load events plus onLoad registered as "load"
	expectedEvents := []string{"click", "load"}
	for _, eventID := range expectedEvents {
		if _, exists := routeEvents[eventID]; !exists {
			t.Errorf("Event %s should be registered", eventID)
		}
	}

	// Test introspection methods
	routes := registry.ListRoutes()
	if len(routes) != 1 || routes[0] != "test-route" {
		t.Errorf("Expected routes [test-route], got %v", routes)
	}

	eventIDs := registry.ListEventIDs("test-route")
	if len(eventIDs) < 2 {
		t.Errorf("Expected at least 2 events for test-route, got %d", len(eventIDs))
	}

	// Test event retrieval
	clickHandler, exists := registry.Get("test-route", "click")
	if !exists {
		t.Error("click event should be retrievable")
	}
	if clickHandler == nil {
		t.Error("click handler should not be nil")
	}
}

func TestEventRegistry_RouteIsolation(t *testing.T) {
	controller := NewController("test-isolation")
	registry := controller.GetEventRegistry()

	// Create two routes with same event names
	route1Opts := func() RouteOptions {
		return RouteOptions{
			ID("route1"),
			Content("Route 1"),
			OnEvent("action", func(ctx RouteContext) error {
				return nil
			}),
		}
	}

	route2Opts := func() RouteOptions {
		return RouteOptions{
			ID("route2"),
			Content("Route 2"),
			OnEvent("action", func(ctx RouteContext) error {
				return nil
			}),
		}
	}

	// Register both routes
	_ = controller.RouteFunc(route1Opts)
	_ = controller.RouteFunc(route2Opts)

	// Verify events are isolated by route
	route1Handler, exists1 := registry.Get("route1", "action")
	route2Handler, exists2 := registry.Get("route2", "action")

	if !exists1 || !exists2 {
		t.Error("Both routes should have action events")
	}

	if route1Handler == nil || route2Handler == nil {
		t.Error("Both handlers should be non-nil")
	}

	// Verify routes are independent
	allEvents := registry.GetAllEvents()
	if len(allEvents) != 2 {
		t.Errorf("Expected 2 routes, got %d", len(allEvents))
	}

	// Test removal doesn't affect other routes
	removed := registry.Remove("route1", "action")
	if !removed {
		t.Error("Should be able to remove route1 action")
	}

	// route2 should still have its event
	_, stillExists := registry.Get("route2", "action")
	if !stillExists {
		t.Error("route2 action should still exist after removing route1 action")
	}

	// route1 should no longer have the event
	_, shouldNotExist := registry.Get("route1", "action")
	if shouldNotExist {
		t.Error("route1 action should not exist after removal")
	}
}
