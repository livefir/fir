package fir

import (
	"net/http/httptest"
	"reflect"
	"testing"

	firHttp "github.com/livefir/fir/internal/http"
	"github.com/livefir/fir/internal/services"
)

// TestDetailedHandlerChainDebug provides detailed debugging of handler chain flow
func TestDetailedHandlerChainDebug(t *testing.T) {
	// Create a real controller
	controller := NewController("test")

	// Create a simple route
	simpleRoute := func() RouteOptions {
		return RouteOptions{
			ID("poc-debug"),
			Content("POC Content"),
		}
	}

	// Create the route and handler through the same path as the failing test
	routeHandler := controller.RouteFunc(simpleRoute)

	// Now use reflection or other means to inspect the route object
	// Since we can't access the route directly, let's try a different approach:
	// Override the canHandlerChainHandle function temporarily

	// Test the actual request
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/poc", nil)

	t.Logf("=== Testing GET /poc request ===")

	// Create the same RequestModel that would be used in canHandlerChainHandle
	requestModel := &firHttp.RequestModel{
		Method: req.Method,
		URL:    req.URL,
		Header: req.Header,
	}

	t.Logf("RequestModel: Method=%s, URL=%s", requestModel.Method, requestModel.URL.Path)

	// Since we can't access route services directly through Controller interface,
	// let's focus on the actual request behavior

	// Now execute the actual request
	routeHandler.ServeHTTP(w, req)

	t.Logf("=== Response Analysis ===")
	t.Logf("Status: %d", w.Code)
	t.Logf("Body: %s", w.Body.String())
	t.Logf("Headers: %v", w.Header())

	// Analyze the result
	if w.Code == 200 && w.Body.String() == "POC Working" {
		t.Log("✅ SUCCESS: Handler chain worked - POC handler responded")
	} else if w.Code == 500 && len(w.Body.String()) > 0 {
		t.Logf("❌ FAILED: Legacy fallback with error: %s", w.Body.String())
	} else {
		t.Logf("❓ UNKNOWN: Unexpected response - Status: %d, Body: %s", w.Code, w.Body.String())
	}
}

// TestServiceComparison compares services between unit test environment and controller environment
func TestServiceComparison(t *testing.T) {
	t.Logf("=== Controller Environment ===")
	// We can't directly access controller's route services, but we can test the behavior
	controller := NewController("test")
	
	simpleRoute := func() RouteOptions {
		return RouteOptions{
			ID("comparison"),
			Content("Test Content"),
		}
	}
	
	handler := controller.RouteFunc(simpleRoute)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/poc", nil)
	handler.ServeHTTP(w, req)
	
	t.Logf("Controller environment - Status: %d", w.Code)
	if w.Code == 200 && w.Body.String() == "POC Working" {
		t.Log("✅ Controller: Handler chain working")
	} else {
		t.Log("❌ Controller: Handler chain failing")
	}

	t.Logf("=== Unit Test Environment (Isolated Services) ===")
	// Recreate the same setup as in our unit tests
	serviceFactory := services.NewServiceFactory(false)
	renderService, templateService, responseBuilder, _, _ := serviceFactory.CreateRenderServices()
	
	t.Logf("Unit Test - RenderService: %v (type: %s)", 
		renderService != nil, 
		reflect.TypeOf(renderService))
	t.Logf("Unit Test - TemplateService: %v (type: %s)", 
		templateService != nil, 
		reflect.TypeOf(templateService))
	t.Logf("Unit Test - ResponseBuilder: %v (type: %s)", 
		responseBuilder != nil, 
		reflect.TypeOf(responseBuilder))

	// Both should have the same services, so the handler chain should work the same
}
