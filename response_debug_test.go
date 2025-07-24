package fir

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResponseWritingDebug(t *testing.T) {
	// First, let's test with temporary disable ENABLED (current state)
	t.Run("WithTemporaryDisable", func(t *testing.T) {
		controller := NewController("test", DevelopmentMode(true))

		simpleRoute := func() RouteOptions {
			return RouteOptions{
				ID("test-route"),
				Content("<html><body><h1>Hello World</h1></body></html>"),
			}
		}

		server := httptest.NewServer(controller.RouteFunc(simpleRoute))
		defer server.Close()

		// Test normal route
		resp := getPageWithSession(t, server.URL+"/")
		t.Logf("WITH DISABLE - Status: %d, Body length: %d", resp.statusCode, len(resp.body))
		if len(resp.body) == 0 {
			t.Error("Body should not be empty with temporary disable")
		}

		// Test POC route (should fallback to legacy and try to find poc.html)
		resp2 := getPageWithSession(t, server.URL+"/poc")
		t.Logf("WITH DISABLE - POC Status: %d, Body: %s", resp2.statusCode, resp2.body)
	})

	// Now test with temporary disable REMOVED
	t.Run("WithoutTemporaryDisable", func(t *testing.T) {
		// We'll need to temporarily remove the disable for this test
		// For now, let's just document what we expect to happen
		t.Skip("Will implement after we remove the temporary disable")
	})
}

func TestPOCHandlerDirectly(t *testing.T) {
	// Test POC handler directly to make sure it works
	t.Run("DirectHandlerTest", func(t *testing.T) {
		controller := NewController("test", DevelopmentMode(true))

		// Create a route that will definitely have a handler chain
		simpleRoute := func() RouteOptions {
			return RouteOptions{
				ID("test-route"),
				Content("<html><body><h1>Test</h1></body></html>"),
			}
		}

		// We need to access the route object to test the handler chain directly
		// This is tricky because createRouteHandler is private
		// Let's create a server and inspect the behavior
		server := httptest.NewServer(controller.RouteFunc(simpleRoute))
		defer server.Close()

		// Test what happens when we hit /poc
		resp := getPageWithSession(t, server.URL+"/poc")
		t.Logf("POC route response - Status: %d", resp.statusCode)
		t.Logf("POC route response - Body: %s", resp.body)
		t.Logf("POC route response - Body length: %d", len(resp.body))

		// Analyze the response to understand what's happening
		if resp.statusCode == 404 {
			t.Log("404 means legacy is trying to find poc.html template")
		} else if resp.statusCode == 500 {
			t.Log("500 means there's an error in processing")
			if strings.Contains(resp.body, "template") {
				t.Log("Error is template-related")
			}
		} else if resp.statusCode == 200 {
			if resp.body == "POC Working" {
				t.Log("SUCCESS: Handler chain is working!")
			} else {
				t.Log("Handler chain might be disabled, legacy template system responded")
			}
		}
	})
}
