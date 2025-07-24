package fir

import (
	"testing"

	"github.com/livefir/fir/internal/routeservices"
)

// TestRouteChainDisable_TemporaryDisableIsActive tests the current state where chain gets disabled
func TestRouteChainDisable_TemporaryDisableIsActive(t *testing.T) {
	// This test documents the current behavior where the temporary disable is active
	// It should FAIL once we remove the temporary disable (which is what we want)

	// Create route services without EventService (the condition that triggers disable)
	services := &routeservices.RouteServices{
		// EventService: nil - this triggers the temporary disable
		Options: &routeservices.Options{
			DisableTemplateCache: false,
			DisableWebsocket:     false,
		},
	}

	// Create a new route to test the actual route-level behavior
	routeOpt := &routeOpt{
		id:      "test-route",
		content: "<div>test</div>",
	}

	route, err := newRoute(services, routeOpt)
	if err != nil {
		t.Fatalf("Failed to create route: %v", err)
	}

	// Check if the handler chain was disabled by the temporary disable logic
	if route.handlerChain != nil {
		t.Error("TEMPORARY DISABLE REMOVED: Handler chain should be nil when EventService is nil (current behavior)")
		t.Error("This test failure means the temporary disable has been removed, which is GOOD!")
		t.Error("Please update this test to expect handlerChain != nil and rename to TestRouteChainEnable_...")
	} else {
		t.Log("CURRENT STATE: Handler chain is disabled when EventService is nil")
		t.Log("This test will fail once we remove the temporary disable")
	}
}

// TestRouteChainDisable_ExpectedBehaviorAfterFix tests what should happen after we remove temporary disable
func TestRouteChainDisable_ExpectedBehaviorAfterFix(t *testing.T) {
	// This test defines what we expect AFTER removing the temporary disable
	// It should FAIL now and PASS after we remove the disable

	// Create route services without EventService
	services := &routeservices.RouteServices{
		// EventService: nil - should NOT disable chain after fix
		Options: &routeservices.Options{
			DisableTemplateCache: false,
			DisableWebsocket:     false,
		},
	}

	// Create a new route
	routeOpt := &routeOpt{
		id:      "test-route",
		content: "<div>test</div>",
	}

	route, err := newRoute(services, routeOpt)
	if err != nil {
		t.Fatalf("Failed to create route: %v", err)
	}

	// After removing temporary disable, chain should exist
	if route.handlerChain == nil {
		t.Error("Handler chain should NOT be nil after removing temporary disable")
		t.Error("Chain should exist even when EventService is nil because POC handler has no dependencies")
	} else {
		t.Log("SUCCESS: Handler chain exists even when EventService is nil")

		// Verify the chain has our POC handler
		handlers := route.handlerChain.GetHandlers()
		if len(handlers) == 0 {
			t.Error("Handler chain is empty - POC handler should be registered")
		} else {
			t.Logf("Handler chain has %d handlers", len(handlers))
			for i, handler := range handlers {
				t.Logf("  Handler %d: %s", i, handler.HandlerName())
			}
		}
	}
}
