package handlers

import (
	"testing"

	firHttp "github.com/livefir/fir/internal/http"
	"github.com/livefir/fir/internal/routeservices"
)

// TestDefaultChainRegistration_POCHandler tests that POC handler is registered in default setup
func TestDefaultChainRegistration_POCHandler(t *testing.T) {
	// Create minimal route services for default chain setup
	services := &routeservices.RouteServices{
		// Add minimal services required for POC handler
		// POC handler has no service dependencies, so this should work
	}

	// Create default handler chain using the standard setup
	chain := SetupDefaultHandlerChain(services)

	if chain == nil {
		t.Fatal("SetupDefaultHandlerChain() returned nil chain")
	}

	// Get all handlers from the chain
	handlers := chain.GetHandlers()

	// Check if POC handler is registered
	var foundPOCHandler bool
	for _, handler := range handlers {
		if handler.HandlerName() == "POCHandler" {
			foundPOCHandler = true
			break
		}
	}

	if !foundPOCHandler {
		t.Error("POCHandler not found in default handler chain setup")
		t.Logf("Registered handlers: %d", len(handlers))
		for i, handler := range handlers {
			t.Logf("  Handler %d: %s", i, handler.HandlerName())
		}
	}
}

// TestDefaultChainRegistration_ChainNotEmpty tests that default chain has at least one handler
func TestDefaultChainRegistration_ChainNotEmpty(t *testing.T) {
	// Create minimal route services
	services := &routeservices.RouteServices{}

	// Create default handler chain
	chain := SetupDefaultHandlerChain(services)

	if chain == nil {
		t.Fatal("SetupDefaultHandlerChain() returned nil chain")
	}

	// Chain should have at least one handler (our POC handler at minimum)
	handlers := chain.GetHandlers()
	if len(handlers) == 0 {
		t.Error("Default handler chain is empty - this would cause all requests to fall back to legacy system")
	}
}

// TestDefaultChainRegistration_POCHandlerSupportsRequest tests POC handler in default chain responds correctly
func TestDefaultChainRegistration_POCHandlerSupportsRequest(t *testing.T) {
	// Create minimal route services
	services := &routeservices.RouteServices{}

	// Create default handler chain
	chain := SetupDefaultHandlerChain(services)

	if chain == nil {
		t.Fatal("SetupDefaultHandlerChain() returned nil chain")
	}

	// Test that chain can handle POC request
	request := &firHttp.RequestModel{
		Method: "GET",
		URL:    mustParseURL("/poc"),
	}

	// Check if any handler in chain supports this request
	handlers := chain.GetHandlers()
	var canHandle bool
	for _, handler := range handlers {
		if handler.SupportsRequest(request) {
			canHandle = true
			if handler.HandlerName() == "POCHandler" {
				// Found our POC handler and it supports the request
				return
			}
		}
	}

	if !canHandle {
		t.Error("Default handler chain cannot handle GET /poc request - would trigger legacy fallback")
		t.Logf("Registered handlers: %d", len(handlers))
		for i, handler := range handlers {
			t.Logf("  Handler %d: %s (supports request: %v)", i, handler.HandlerName(), handler.SupportsRequest(request))
		}
	}
}
