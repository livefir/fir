package fir

import (
	"testing"

	firHttp "github.com/livefir/fir/internal/http"
	routeFactory "github.com/livefir/fir/internal/route"
	"github.com/livefir/fir/internal/routeservices"
)

func TestControllerHandlerSetup(t *testing.T) {
	// Create a real controller to see what services it creates
	controller := NewController("test", DevelopmentMode(true))

	// We need to access the internal controller to see the services
	// Since we can't access private methods directly, let's create a route and inspect it
	
	simpleRoute := func() RouteOptions {
		return RouteOptions{
			ID("test-route"),
			Content("<html><body><h1>Test</h1></body></html>"),
		}
	}

	// Create the route handler function
	_ = controller.RouteFunc(simpleRoute)
	
	// The handlerFunc wraps the actual route, but we can't easily inspect it
	// Let's try a different approach: check what happens when we create route services directly
	
	t.Log("Testing what happens in route creation...")
	
	// This is tricky because createRouteServices is private
	// Let's test the observable behavior instead
	
	t.Log("We need to find another way to inspect the handler chain setup")
}

func TestDirectServiceCreation(t *testing.T) {
	// Test creating services like the controller does
	// We'll create minimal services and see what handlers get registered
	
	// First, let's see what the handler setup looks like with nil services
	services := &routeservices.RouteServices{
		Options: &routeservices.Options{
			DisableTemplateCache: false,
			DisableWebsocket:     false,
		},
	}

	t.Logf("Service state:")
	t.Logf("  EventService: %v", services.EventService)
	t.Logf("  RenderService: %v", services.RenderService)
	t.Logf("  TemplateService: %v", services.TemplateService)
	t.Logf("  ResponseBuilder: %v", services.ResponseBuilder)

	// Create handler chain
	factory := routeFactory.NewRouteServiceFactory(services)
	chain := factory.CreateHandlerChain()

	if chain == nil {
		t.Fatal("Handler chain is nil")
	}

	handlers := chain.GetHandlers()
	t.Logf("Handlers registered: %d", len(handlers))
	for i, handler := range handlers {
		t.Logf("  Handler %d: %s", i, handler.HandlerName())
	}

	// Test if POC handler supports /poc
	request := &firHttp.RequestModel{
		Method: "GET",
		URL:    mustParseURL("/poc"),
	}

	for _, handler := range handlers {
		supports := handler.SupportsRequest(request)
		t.Logf("Handler %s supports GET /poc: %v", handler.HandlerName(), supports)
	}
}
