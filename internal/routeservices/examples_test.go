package routeservices

import (
	"testing"

	"github.com/livefir/fir/internal/event"
	"github.com/livefir/fir/pubsub"
)

// Additional examples demonstrating RouteServices testing patterns

// Example: Testing RouteServices with different configurations
func TestRouteServicesExampleConfigurations(t *testing.T) {
	// Test with minimal configuration
	minimalServices := NewRouteServices(
		event.NewEventRegistry(),
		pubsub.NewInmem(),
		&mockRenderer{},
		&Options{AppName: "minimal"},
	)

	if err := minimalServices.ValidateServices(); err != nil {
		t.Errorf("Minimal services should validate: %v", err)
	}

	// Test with full configuration
	fullOptions := &Options{
		AppName:              "full-test",
		DisableTemplateCache: false,
		DisableWebsocket:     false,
		DevelopmentMode:      true,
		DebugLog:             true,
		ReadFile: func(filename string) (string, []byte, error) {
			return filename, []byte("content"), nil
		},
		ExistFile: func(filename string) bool {
			return true
		},
	}

	fullServices := NewRouteServices(
		event.NewEventRegistry(),
		pubsub.NewInmem(),
		&mockRenderer{},
		fullOptions,
	)

	if err := fullServices.ValidateServices(); err != nil {
		t.Errorf("Full services should validate: %v", err)
	}

	// Test that options are correctly set
	if fullServices.Options.AppName != "full-test" {
		t.Error("App name should be set correctly")
	}

	if !fullServices.Options.DevelopmentMode {
		t.Error("Development mode should be enabled")
	}
}

// Example: Testing event registry patterns
func TestEventRegistryExamplePatterns(t *testing.T) {
	services := NewRouteServices(
		event.NewEventRegistry(),
		pubsub.NewInmem(),
		&mockRenderer{},
		&Options{AppName: "event-test"},
	)

	// Register multiple events for different routes
	routes := []string{"route1", "route2", "route3"}
	events := []string{"click", "submit", "load"}

	for _, routeID := range routes {
		for _, eventID := range events {
			handler := func(data interface{}) {
				// Event handler
			}

			err := services.EventRegistry.Register(routeID, eventID, handler)
			if err != nil {
				t.Fatalf("Failed to register %s:%s - %v", routeID, eventID, err)
			}
		}
	}

	// Verify all events are registered
	allEvents := services.EventRegistry.GetAllEvents()
	if len(allEvents) != 3 {
		t.Errorf("Expected 3 routes, got %d", len(allEvents))
	}

	for _, routeID := range routes {
		routeEvents := services.EventRegistry.GetRouteEvents(routeID)
		if len(routeEvents) != 3 {
			t.Errorf("Expected 3 events for %s, got %d", routeID, len(routeEvents))
		}
	}

	// Test removing events
	removed := services.EventRegistry.Remove("route1", "click")
	if !removed {
		t.Error("Should have removed the event")
	}

	// Verify removal
	_, found := services.EventRegistry.Get("route1", "click")
	if found {
		t.Error("Event should have been removed")
	}

	// Test removing entire route
	removed = services.EventRegistry.RemoveRoute("route2")
	if !removed {
		t.Error("Should have removed the route")
	}

	allEvents = services.EventRegistry.GetAllEvents()
	if len(allEvents) != 2 {
		t.Errorf("Expected 2 routes after removal, got %d", len(allEvents))
	}
}
