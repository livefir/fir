package handlers

import (
	"context"
	"net/http/httptest"
	"testing"

	firHttp "github.com/livefir/fir/internal/http"
)

// TestChainIntegration_AddHandler tests adding POC handler to chain
func TestChainIntegration_AddHandler(t *testing.T) {
	// Create handler chain with minimal dependencies
	chain := NewDefaultHandlerChain(nil, nil)
	pocHandler := NewPOCHandler()

	// Add POC handler to chain
	chain.AddHandler(pocHandler)

	// Verify handler was added
	handlers := chain.GetHandlers()
	if len(handlers) != 1 {
		t.Fatalf("Expected 1 handler in chain, got %d", len(handlers))
	}

	if handlers[0].HandlerName() != "POCHandler" {
		t.Errorf("Expected handler name 'POCHandler', got %q", handlers[0].HandlerName())
	}
}

// TestChainIntegration_ProcessRequest tests processing request through chain
func TestChainIntegration_ProcessRequest(t *testing.T) {
	// Create handler chain with POC handler
	chain := NewDefaultHandlerChain(nil, nil)
	pocHandler := NewPOCHandler()
	chain.AddHandler(pocHandler)

	// Create request that POC handler should support
	request := &firHttp.RequestModel{
		Method: "GET",
		URL:    mustParseURL("/poc"),
	}

	// Process request through chain
	ctx := context.Background()
	response, err := chain.Handle(ctx, request)

	// Verify successful processing
	if err != nil {
		t.Fatalf("Chain.Handle() returned error: %v", err)
	}

	if response == nil {
		t.Fatal("Chain.Handle() returned nil response")
	}

	if response.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", response.StatusCode)
	}

	expectedBody := "POC Working"
	actualBody := string(response.Body)
	if actualBody != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, actualBody)
	}
}

// TestChainIntegration_UnsupportedRequest tests chain behavior with unsupported requests
func TestChainIntegration_UnsupportedRequest(t *testing.T) {
	// Create handler chain with POC handler
	chain := NewDefaultHandlerChain(nil, nil)
	pocHandler := NewPOCHandler()
	chain.AddHandler(pocHandler)

	// Create request that POC handler does NOT support
	request := &firHttp.RequestModel{
		Method: "POST",
		URL:    mustParseURL("/other"),
	}

	// Process request through chain
	ctx := context.Background()
	response, err := chain.Handle(ctx, request)

	// Verify chain returns error for unsupported request
	if err == nil {
		t.Fatal("Chain.Handle() should return error for unsupported request")
	}

	if response != nil {
		t.Error("Chain.Handle() should return nil response for unsupported request")
	}

	expectedErrorMsg := "no handler found for request: POST /other"
	if err.Error() != expectedErrorMsg {
		t.Errorf("Expected error %q, got %q", expectedErrorMsg, err.Error())
	}
}

// TestChainIntegration_NoFallback tests that successful chain handling doesn't trigger fallback
func TestChainIntegration_NoFallback(t *testing.T) {
	// This test verifies the key integration point: when handler chain succeeds,
	// no legacy fallback should be triggered

	// Create handler chain with POC handler
	chain := NewDefaultHandlerChain(nil, nil)
	pocHandler := NewPOCHandler()
	chain.AddHandler(pocHandler)

	// Create HTTP request and response writer
	req := httptest.NewRequest("GET", "/poc", nil)
	w := httptest.NewRecorder()

	// Create request model using the HTTP adapter (like route.go does)
	pair, err := firHttp.NewRequestResponsePair(w, req, nil)
	if err != nil {
		t.Fatalf("Failed to create request/response pair: %v", err)
	}

	// Test canHandlerChainHandle equivalent logic
	chainHandlers := chain.GetHandlers()
	if len(chainHandlers) == 0 {
		t.Fatal("Chain should have handlers configured")
	}

	// Verify at least one handler supports the request
	var foundSupportingHandler bool
	for _, handler := range chainHandlers {
		if handler.SupportsRequest(pair.Request) {
			foundSupportingHandler = true
			break
		}
	}

	if !foundSupportingHandler {
		t.Fatal("No handler in chain supports GET /poc request - fallback would be triggered")
	}

	// Process request through chain
	ctx := context.Background()
	response, err := chain.Handle(ctx, pair.Request)

	// Verify successful processing (no fallback needed)
	if err != nil {
		t.Fatalf("Chain processing failed, would trigger fallback: %v", err)
	}

	if response == nil {
		t.Fatal("Chain returned nil response, would trigger fallback")
	}

	// Verify response is what we expect from POC handler
	if response.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", response.StatusCode)
	}

	expectedBody := "POC Working"
	actualBody := string(response.Body)
	if actualBody != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, actualBody)
	}
}
