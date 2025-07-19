package fir

import (
	"testing"
)

func TestPhase5HandlerChainDebug(t *testing.T) {
	// Create a controller with development mode to check if services are created
	controller := NewController("debug-handler-chain", DevelopmentMode(true))

	// Use the same doubler route as the analysis test
	routeFunc := func() RouteOptions {
		return RouteOptions{
			ID("test-route"),
			Content(`<div>{{ .num }}</div>`),
			OnLoad(func(ctx RouteContext) error {
				return ctx.KV("num", 0)
			}),
		}
	}

	// Create the route handler 
	handler := controller.RouteFunc(routeFunc)
	if handler == nil {
		t.Fatalf("Failed to create route handler")
	}

	// The key insight is that if our debug logs show "handler chain cannot handle request type",
	// it means one of two things:
	// 1. Handler chain exists but has no handlers
	// 2. Handler chain exists but handlers don't support the request type
	
	t.Logf("Handler creation successful - the debug logs from our earlier test show the real issue")
	t.Logf("Based on debug output: 'handler chain cannot handle request type'")
	t.Logf("This means either:")
	t.Logf("  1. No handlers in chain (services missing)")
	t.Logf("  2. Handlers exist but don't support GET/, POST, etc.")
	
	// The analysis test already showed us that ALL request types are falling back to legacy
	// This means the handler chain either has no handlers, or the handlers aren't recognizing these request patterns
	t.Logf("From analysis test results:")
	t.Logf("  - GET /: legacy fallback")
	t.Logf("  - POST JSON events: legacy fallback") 
	t.Logf("  - POST forms: legacy fallback")
	t.Logf("  - WebSocket upgrades: legacy fallback")
	t.Logf("This suggests the handler chain is not properly configured with handlers")
}
