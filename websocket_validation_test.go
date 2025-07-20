package fir

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/livefir/fir/internal/handlers"
	firHttp "github.com/livefir/fir/internal/http"
)

// TestWebSocketValidation specifically tests the validation logic
func TestWebSocketValidation(t *testing.T) {
	factory := &MockServiceFactory{}
	services := factory.CreateTestRouteServices()
	handler := handlers.NewWebSocketHandler(
		services.EventService,
		services.ResponseBuilder,
	)

	req := &firHttp.RequestModel{
		Method: http.MethodGet,
		URL:    &url.URL{Path: "/ws"},
		Header: http.Header{
			"Upgrade":               {"websocket"},
			"Connection":            {"upgrade"},
			"Sec-Websocket-Version": {"13"},                       // Use canonical form
			"Sec-Websocket-Key":     {"dGhlIHNhbXBsZSBub25jZQ=="}, // Use canonical form
		},
	}

	// Test individual header values
	fmt.Printf("Raw headers: %v\n", req.Header)
	fmt.Printf("Upgrade header: '%s'\n", req.Header.Get("Upgrade"))
	fmt.Printf("Connection header: '%s'\n", req.Header.Get("Connection"))
	fmt.Printf("WebSocket-Version header: '%s'\n", req.Header.Get("Sec-WebSocket-Version"))
	fmt.Printf("WebSocket-Key header: '%s'\n", req.Header.Get("Sec-WebSocket-Key"))

	// Try accessing directly (using canonical form)
	fmt.Printf("Direct access to Sec-WebSocket-Version: %v\n", req.Header[http.CanonicalHeaderKey("Sec-WebSocket-Version")])
	fmt.Printf("Direct access to Sec-WebSocket-Key: %v\n", req.Header[http.CanonicalHeaderKey("Sec-WebSocket-Key")])

	// Try with different case variations
	fmt.Printf("WebSocket-Version (titlecase): '%s'\n", req.Header.Get("Sec-Websocket-Version"))
	fmt.Printf("WebSocket-Key (titlecase): '%s'\n", req.Header.Get("Sec-Websocket-Key"))

	// Test if isWebSocketUpgrade works
	isUpgrade := func(req *firHttp.RequestModel) bool {
		upgrade := req.Header.Get("Upgrade")
		connection := req.Header.Get("Connection")
		return upgrade == "websocket" && connection == "upgrade"
	}

	fmt.Printf("isWebSocketUpgrade: %v\n", isUpgrade(req))

	// Test SupportsRequest
	fmt.Printf("SupportsRequest: %v\n", handler.SupportsRequest(req))

	// Test empty headers to see what happens
	emptyReq := &firHttp.RequestModel{
		Method: http.MethodGet,
		URL:    &url.URL{Path: "/ws"},
		Header: http.Header{},
	}

	fmt.Printf("Empty headers SupportsRequest: %v\n", handler.SupportsRequest(emptyReq))
}
