package fir

import (
	"testing"

	routeFactory "github.com/livefir/fir/internal/route"
	"github.com/livefir/fir/internal/routeservices"
)

func TestHandlerChainDebug(t *testing.T) {
	// Create minimal route services like a real controller would
	services := &routeservices.RouteServices{
		Options: &routeservices.Options{
			DisableTemplateCache: false,
			DisableWebsocket:     false,
		},
	}

	t.Logf("EventService: %v", services.EventService)
	t.Logf("RenderService: %v", services.RenderService)
	t.Logf("TemplateService: %v", services.TemplateService)
	t.Logf("ResponseBuilder: %v", services.ResponseBuilder)

	// Test handler chain setup
	factory := routeFactory.NewRouteServiceFactory(services)
	chain := factory.CreateHandlerChain()

	if chain == nil {
		t.Log("Handler chain is nil")
	} else {
		handlers := chain.GetHandlers()
		t.Logf("Handler chain has %d handlers:", len(handlers))
		for i, handler := range handlers {
			t.Logf("  Handler %d: %s", i, handler.HandlerName())
		}
	}
}
