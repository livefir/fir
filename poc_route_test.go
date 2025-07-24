package fir

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPOCRoute(t *testing.T) {
	controller := NewController("test", DevelopmentMode(true))

	// Create a simple route that should fallback to legacy for non-POC paths
	simpleRoute := func() RouteOptions {
		return RouteOptions{
			ID("test-route"),
			Content("<html><body><h1>Test Route</h1></body></html>"),
		}
	}

	server := httptest.NewServer(controller.RouteFunc(simpleRoute))
	defer server.Close()

	// Test POC path - should be handled by handler chain
	resp := getPageWithSession(t, server.URL+"/poc")
	t.Logf("POC route - Status: %d", resp.statusCode)
	t.Logf("POC route - Body: %s", resp.body)

	if resp.statusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.statusCode)
	}

	expectedBody := "POC Working"
	if resp.body != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, resp.body)
	}
}

func TestLegacyFallback(t *testing.T) {
	controller := NewController("test", DevelopmentMode(true))

	// Create a simple route
	simpleRoute := func() RouteOptions {
		return RouteOptions{
			ID("test-route"),
			Content("<html><body><h1>Legacy Fallback Works</h1></body></html>"),
		}
	}

	server := httptest.NewServer(controller.RouteFunc(simpleRoute))
	defer server.Close()

	// Test root path - should fallback to legacy
	resp := getPageWithSession(t, server.URL+"/")
	t.Logf("Legacy route - Status: %d", resp.statusCode)
	t.Logf("Legacy route - Body: %s", resp.body)

	if resp.statusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.statusCode)
	}

	if len(resp.body) == 0 {
		t.Error("Expected response body from legacy fallback, got empty")
	}

	if !strings.Contains(resp.body, "Legacy Fallback Works") {
		t.Errorf("Expected body to contain 'Legacy Fallback Works', got %q", resp.body)
	}
}

// Note: removed custom contains function, using strings.Contains instead
