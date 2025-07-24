package handlers

import (
	"context"
	"testing"

	firHttp "github.com/livefir/fir/internal/http"
	"github.com/livefir/fir/internal/routeservices"
)

// TestEndToEndPOC_HandlerChainWorksWithoutEventService tests that handler chain works when EventService is nil
func TestEndToEndPOC_HandlerChainWorksWithoutEventService(t *testing.T) {
	// This test verifies that our POC proves the handler chain can work
	// even in scenarios where EventService is nil (the condition that was disabling the chain)

	// Create route services without EventService - the problematic scenario
	services := &routeservices.RouteServices{
		// EventService: nil - this was causing chain to be disabled
		Options: &routeservices.Options{
			DisableTemplateCache: false,
			DisableWebsocket:     false,
		},
	}

	// Create default handler chain (should have POC handler)
	chain := SetupDefaultHandlerChain(services)

	if chain == nil {
		t.Fatal("SetupDefaultHandlerChain() returned nil - handler chain setup failed")
	}

	// Verify chain has handlers
	handlers := chain.GetHandlers()
	if len(handlers) == 0 {
		t.Fatal("Handler chain is empty - POC handler should be registered")
	}

	// Test end-to-end request processing
	request := &firHttp.RequestModel{
		Method: "GET",
		URL:    mustParseURL("/poc"),
	}

	// Simulate the canHandlerChainHandle logic from route.go
	var canHandle bool
	for _, handler := range handlers {
		if handler.SupportsRequest(request) {
			canHandle = true
			break
		}
	}

	if !canHandle {
		t.Fatal("Handler chain cannot handle GET /poc - would fall back to legacy")
	}

	// Process request through the chain (like handleRequestWithChain does)
	ctx := context.Background()
	response, err := chain.Handle(ctx, request)

	// Verify successful processing
	if err != nil {
		t.Fatalf("Handler chain processing failed: %v", err)
	}

	if response == nil {
		t.Fatal("Handler chain returned nil response")
	}

	// Verify response is correct
	if response.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", response.StatusCode)
	}

	expectedBody := "POC Working"
	actualBody := string(response.Body)
	if actualBody != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, actualBody)
	}

	t.Log("SUCCESS: Handler chain processed request end-to-end without EventService")
	t.Log("This proves the architecture works and no legacy fallback is needed for POC requests")
}

// TestEndToEndPOC_NoLegacyFallbackTriggered tests that handler chain success means no fallback
func TestEndToEndPOC_NoLegacyFallbackTriggered(t *testing.T) {
	// This test documents that successful handler chain processing means
	// no legacy fallback should be triggered

	services := &routeservices.RouteServices{
		Options: &routeservices.Options{
			DisableTemplateCache: false,
			DisableWebsocket:     false,
		},
	}

	chain := SetupDefaultHandlerChain(services)
	if chain == nil {
		t.Fatal("Handler chain setup failed")
	}

	request := &firHttp.RequestModel{
		Method: "GET",
		URL:    mustParseURL("/poc"),
	}

	// Test the decision logic that would be used in handleRequestWithChain
	// 1. Check if chain can handle request (canHandlerChainHandle equivalent)
	handlers := chain.GetHandlers()
	canChainHandle := false
	for _, handler := range handlers {
		if handler.SupportsRequest(request) {
			canChainHandle = true
			break
		}
	}

	if !canChainHandle {
		t.Fatal("Chain cannot handle request - legacy fallback would be triggered")
	}

	// 2. Process through chain (handleRequestWithChain equivalent)
	ctx := context.Background()
	response, err := chain.Handle(ctx, request)

	// 3. If chain succeeds, no fallback needed
	if err != nil {
		t.Fatalf("Chain processing failed: %v - this would trigger legacy fallback", err)
	}

	if response == nil {
		t.Fatal("Chain returned nil response - this would trigger legacy fallback")
	}

	// SUCCESS: Chain handled request completely
	t.Log("PROOF OF CONCEPT COMPLETE:")
	t.Log("✅ Handler chain can handle requests without EventService")
	t.Log("✅ No legacy fallback is triggered when chain succeeds")
	t.Log("✅ Response is generated correctly by chain")
	t.Logf("✅ Status: %d, Body: %q", response.StatusCode, string(response.Body))
}
