package fir

import (
	"net/http/httptest"
	"testing"
)

func TestHandlerChainDebugFlow(t *testing.T) {
	// Enable debug logging to see what's happening
	controller := NewController("test", DevelopmentMode(true))

	simpleRoute := func() RouteOptions {
		return RouteOptions{
			ID("debug-route"),
			Content("<html><body><h1>Debug Test</h1></body></html>"),
		}
	}

	server := httptest.NewServer(controller.RouteFunc(simpleRoute))
	defer server.Close()

	t.Log("Testing /poc request with handler chain enabled...")
	
	// Test POC path - this should trigger our debug logs
	resp := getPageWithSession(t, server.URL+"/poc")
	
	t.Logf("Response Status: %d", resp.statusCode)
	t.Logf("Response Body: %s", resp.body)
	t.Logf("Response Body Length: %d", len(resp.body))
	
	// Analyze the result
	if resp.statusCode == 500 && resp.body != "POC Working" {
		t.Log("DIAGNOSIS: Handler chain is not processing the request")
		t.Log("This means either:")
		t.Log("  1. canHandlerChainHandle() returned false")
		t.Log("  2. Handler chain returned an error")
		t.Log("  3. Response writing failed")
	} else if resp.body == "POC Working" {
		t.Log("SUCCESS: Handler chain is working correctly!")
	}
}

func TestNormalRouteStillWorks(t *testing.T) {
	// Test that normal routes still work
	controller := NewController("test", DevelopmentMode(true))

	simpleRoute := func() RouteOptions {
		return RouteOptions{
			ID("normal-route"),
			Content("<html><body><h1>Normal Route</h1></body></html>"),
		}
	}

	server := httptest.NewServer(controller.RouteFunc(simpleRoute))
	defer server.Close()

	t.Log("Testing normal / request...")
	
	resp := getPageWithSession(t, server.URL+"/")
	
	t.Logf("Normal route - Status: %d", resp.statusCode)
	t.Logf("Normal route - Body: %s", resp.body)
	t.Logf("Normal route - Body Length: %d", len(resp.body))
	
	if len(resp.body) == 0 {
		t.Error("Normal route should return content, but got empty body")
	}
}
