package handlers

import (
	"context"
	"testing"

	firHttp "github.com/livefir/fir/internal/http"
	"github.com/livefir/fir/internal/routeservices"
)

// TestRouteChainIntegration_TemporaryDisableRemoved tests that handler chain is not disabled when EventService is nil
func TestRouteChainIntegration_TemporaryDisableRemoved(t *testing.T) {
	// Create route services WITHOUT EventService (nil) to test the temporary disable condition
	services := &routeservices.RouteServices{
		// EventService: nil - this is the condition that currently disables the chain
		// Other services also nil to test minimal scenario
	}

	// Create default handler chain - this should succeed and have our POC handler
	chain := SetupDefaultHandlerChain(services)

	if chain == nil {
		t.Fatal("SetupDefaultHandlerChain() returned nil chain")
	}

	// Verify POC handler is in the chain
	handlers := chain.GetHandlers()
	if len(handlers) == 0 {
		t.Fatal("Handler chain is empty - POC handler should be registered")
	}

	var foundPOCHandler bool
	for _, handler := range handlers {
		if handler.HandlerName() == "POCHandler" {
			foundPOCHandler = true
			break
		}
	}

	if !foundPOCHandler {
		t.Error("POCHandler not found in chain, but it should be there regardless of EventService being nil")
		t.Logf("Registered handlers: %d", len(handlers))
		for i, handler := range handlers {
			t.Logf("  Handler %d: %s", i, handler.HandlerName())
		}
	}
}

// TestRouteChainIntegration_ChainCanHandleRequest tests that the chain can handle requests even without EventService
func TestRouteChainIntegration_ChainCanHandleRequest(t *testing.T) {
	// Create route services without EventService to simulate the condition that currently disables chain
	services := &routeservices.RouteServices{
		// EventService: nil - the problematic condition
	}

	// Create default handler chain
	chain := SetupDefaultHandlerChain(services)

	if chain == nil {
		t.Fatal("SetupDefaultHandlerChain() returned nil chain")
	}

	// Test that chain can handle a POC request
	request := &firHttp.RequestModel{
		Method: "GET",
		URL:    mustParseURL("/poc"),
	}

	// This simulates the canHandlerChainHandle() logic from route.go
	chainHandlers := chain.GetHandlers()
	if len(chainHandlers) == 0 {
		t.Fatal("Chain has no handlers - canHandlerChainHandle() would return false")
	}

	var canHandle bool
	for _, handler := range chainHandlers {
		if handler.SupportsRequest(request) {
			canHandle = true
			break
		}
	}

	if !canHandle {
		t.Error("Chain cannot handle GET /poc request - this would trigger legacy fallback")
		t.Logf("Registered handlers: %d", len(chainHandlers))
		for i, handler := range chainHandlers {
			t.Logf("  Handler %d: %s (supports request: %v)", i, handler.HandlerName(), handler.SupportsRequest(request))
		}
	}
}

// TestRouteChainIntegration_NoLegacyFallbackNeeded tests that with POC handler, no legacy fallback should occur
func TestRouteChainIntegration_NoLegacyFallbackNeeded(t *testing.T) {
	// Create route services representing current production state (EventService = nil)
	services := &routeservices.RouteServices{
		// This represents the current problematic scenario where EventService is nil
		// and the chain gets disabled, forcing legacy fallback
	}

	// Create default handler chain
	chain := SetupDefaultHandlerChain(services)

	if chain == nil {
		t.Fatal("SetupDefaultHandlerChain() returned nil chain")
	}

	// Test the full request processing pipeline
	request := &firHttp.RequestModel{
		Method: "GET",
		URL:    mustParseURL("/poc"),
	}

	// Process request through chain (like handleRequestWithChain would do)
	ctx := context.Background()
	response, err := chain.Handle(ctx, request)

	// Verify successful processing without fallback
	if err != nil {
		t.Fatalf("Chain processing failed: %v - this would trigger legacy fallback", err)
	}

	if response == nil {
		t.Fatal("Chain returned nil response - this would trigger legacy fallback")
	}

	// Verify it's the POC response
	if response.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", response.StatusCode)
	}

	expectedBody := "POC Working"
	actualBody := string(response.Body)
	if actualBody != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, actualBody)
	}

	// If we get here, it means NO LEGACY FALLBACK was needed
	t.Logf("SUCCESS: Handler chain processed request without legacy fallback")
}
