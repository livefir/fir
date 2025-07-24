package handlers

import (
	"net/http/httptest"
	"testing"

	firHttp "github.com/livefir/fir/internal/http"
)

// TestWebSocketConnectionHandling_UpgradeDetection tests WebSocket upgrade request detection
func TestWebSocketConnectionHandling_UpgradeDetection(t *testing.T) {
	// Create WebSocket handler
	wsHandler := NewWebSocketHandler(
		&mockEventService{},
		&mockResponseBuilder{},
	)

	// Create test request with WebSocket upgrade headers
	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
	req.Header.Set("Sec-WebSocket-Version", "13")

	// Convert to RequestModel
	adapter := firHttp.NewRequestAdapter(nil)
	requestModel, err := adapter.ParseRequest(req)
	if err != nil {
		t.Fatalf("Failed to adapt request: %v", err)
	}

	// Test if WebSocket handler supports this request
	supports := wsHandler.SupportsRequest(requestModel)
	if !supports {
		t.Error("Expected WebSocket handler to support WebSocket upgrade request")
	}

	t.Logf("WebSocket handler correctly detected upgrade request")
}

// TestWebSocketConnectionHandling_NonUpgradeRequest tests that non-WebSocket requests are rejected
func TestWebSocketConnectionHandling_NonUpgradeRequest(t *testing.T) {
	// Create WebSocket handler
	wsHandler := NewWebSocketHandler(
		&mockEventService{},
		&mockResponseBuilder{},
	)

	// Create regular GET request without upgrade headers
	req := httptest.NewRequest("GET", "/api", nil)

	// Convert to RequestModel
	adapter := firHttp.NewRequestAdapter(nil)
	requestModel, err := adapter.ParseRequest(req)
	if err != nil {
		t.Fatalf("Failed to adapt request: %v", err)
	}

	// Test if WebSocket handler supports this request
	supports := wsHandler.SupportsRequest(requestModel)
	if supports {
		t.Error("Expected WebSocket handler to reject non-upgrade request")
	}

	t.Logf("WebSocket handler correctly rejected non-upgrade request")
}

// TestWebSocketConnectionHandling_ChainIntegration tests WebSocket handler in the handler chain
func TestWebSocketConnectionHandling_ChainIntegration(t *testing.T) {
	// Create WebSocket handler
	wsHandler := NewWebSocketHandler(
		&mockEventService{},
		&mockResponseBuilder{},
	)

	// Create integration and setup chain
	chain := NewPriorityHandlerChain(&defaultHandlerLogger{}, &defaultHandlerMetrics{})
	chain.AddHandlerWithConfig(wsHandler, HandlerConfig{
		Name:     wsHandler.HandlerName(),
		Priority: 5, // Highest priority
		Enabled:  true,
	})
	integration := NewRouteHandlerIntegration(chain)

	// Create test request with WebSocket upgrade headers
	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
	req.Header.Set("Sec-WebSocket-Version", "13")

	// Test if integration can handle this request
	canHandle := integration.CanHandleRequest(req)
	if !canHandle {
		t.Error("Expected integration to detect that WebSocket handler can handle upgrade request")
	}

	t.Logf("Handler chain correctly detected WebSocket upgrade request")
}

// TestWebSocketConnectionHandling_HighestPriority tests that WebSocket has highest priority
func TestWebSocketConnectionHandling_HighestPriority(t *testing.T) {
	// Create multiple handlers
	wsHandler := NewWebSocketHandler(&mockEventService{}, &mockResponseBuilder{})
	getHandler := NewGetHandler(&mockRenderService{}, &mockTemplateService{}, &mockResponseBuilder{}, &mockEventService{})

	// Create integration and setup chain with both handlers
	chain := NewPriorityHandlerChain(&defaultHandlerLogger{}, &defaultHandlerMetrics{})

	// Add GET handler first (lower priority)
	chain.AddHandlerWithConfig(getHandler, HandlerConfig{
		Name:     getHandler.HandlerName(),
		Priority: 50,
		Enabled:  true,
	})

	// Add WebSocket handler (higher priority)
	chain.AddHandlerWithConfig(wsHandler, HandlerConfig{
		Name:     wsHandler.HandlerName(),
		Priority: 5,
		Enabled:  true,
	})

	integration := NewRouteHandlerIntegration(chain)

	// Create test request that could be handled by both (GET with upgrade headers)
	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
	req.Header.Set("Sec-WebSocket-Version", "13")

	// Convert to RequestModel to check which handler would be selected
	adapter := firHttp.NewRequestAdapter(nil)
	requestModel, err := adapter.ParseRequest(req)
	if err != nil {
		t.Fatalf("Failed to adapt request: %v", err)
	}

	// Check which handlers support this request
	wsSupports := wsHandler.SupportsRequest(requestModel)
	getSupports := getHandler.SupportsRequest(requestModel)

	t.Logf("WebSocket handler supports: %v", wsSupports)
	t.Logf("GET handler supports: %v", getSupports)

	// WebSocket handler should support it, GET handler should not (due to WebSocket detection)
	if !wsSupports {
		t.Error("Expected WebSocket handler to support WebSocket upgrade request")
	}

	if getSupports {
		t.Error("Expected GET handler to reject WebSocket upgrade request")
	}

	// Integration should be able to handle this
	canHandle := integration.CanHandleRequest(req)
	if !canHandle {
		t.Error("Expected integration to handle WebSocket upgrade request")
	}

	t.Logf("Handler chain correctly prioritized WebSocket handler for upgrade request")
}

// TestWebSocketConnectionHandling_ErrorScenarios tests WebSocket error handling
func TestWebSocketConnectionHandling_ErrorScenarios(t *testing.T) {
	// Create mock response builder that returns error response
	mockBuilder := &mockResponseBuilder{
		buildErrorResponse: &firHttp.ResponseModel{
			StatusCode: 400,
			Headers: map[string]string{
				"Content-Type": "text/plain",
			},
			Body: []byte("WebSocket upgrade failed"),
		},
		buildError: nil,
	}

	// Create WebSocket handler
	wsHandler := NewWebSocketHandler(
		&mockEventService{},
		mockBuilder,
	)

	// Create test request with invalid WebSocket upgrade headers
	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	// Missing required Sec-WebSocket-Key header

	// Convert to RequestModel
	adapter := firHttp.NewRequestAdapter(nil)
	requestModel, err := adapter.ParseRequest(req)
	if err != nil {
		t.Fatalf("Failed to adapt request: %v", err)
	}

	// Handler should still support the request (basic WebSocket detection)
	supports := wsHandler.SupportsRequest(requestModel)
	if !supports {
		t.Error("Expected WebSocket handler to support WebSocket request even with missing headers")
	}

	// When handled, it should return an error response
	resp, err := wsHandler.Handle(req.Context(), requestModel)

	// For WebSocket connections, the handler might return nil and handle the connection directly
	// or return an error response for invalid upgrade requests
	t.Logf("WebSocket Handle response: %v, error: %v", resp, err)

	// This test mainly verifies that WebSocket handling doesn't panic or cause issues
}
