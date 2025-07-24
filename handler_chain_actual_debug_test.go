package fir

import (
	"net/http/httptest"
	"testing"
)

// TestHandlerChainInActualController tests the handler chain behavior in a real controller setup
func TestHandlerChainInActualController(t *testing.T) {
	// Create a real controller with default setup
	controller := NewController("test")

	// Create a simple route that should trigger the handler chain
	simpleRoute := func() RouteOptions {
		return RouteOptions{
			ID("poc"),
			Content("POC Content"),
		}
	}

	// Get the HTTP handler from RouteFunc
	handler := controller.RouteFunc(simpleRoute)

	// Test with actual HTTP request to /poc
	t.Logf("Testing actual HTTP request to /poc")
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/poc", nil)
	handler.ServeHTTP(w, req)

	t.Logf("Response Status: %d", w.Code)
	t.Logf("Response Body: %s", w.Body.String())
	t.Logf("Response Headers: %v", w.Header())

	// Check if it's working as expected
	if w.Code == 200 && w.Body.String() == "POC Working" {
		t.Log("✅ Handler chain is working correctly - POC handler responded")
	} else if w.Code == 500 {
		t.Log("❌ Handler chain failed - falling back to legacy with template error")
	} else {
		t.Logf("❓ Unexpected response - Status: %d, Body: %s", w.Code, w.Body.String())
	}
}
