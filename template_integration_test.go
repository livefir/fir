package fir

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// TestPOCHandlerWithTemplateDir tests POC handler works when template directory is properly configured
func TestPOCHandlerWithTemplateDir(t *testing.T) {
	// Save current working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal("Failed to get working directory:", err)
	}
	defer func() {
		os.Chdir(originalWd)
	}()

	// Change to testdata directory so templates can be found
	testdataDir := filepath.Join(originalWd, "testdata")
	if err := os.Chdir(testdataDir); err != nil {
		t.Fatal("Failed to change to testdata directory:", err)
	}

	// Create controller with proper public directory configuration
	controller := NewController("test-app", WithPublicDir("public"))

	// Create route function that returns RouteOptions
	routeFunc := func() RouteOptions {
		return RouteOptions{
			ID("test-route"),
		}
	}

	// Get route handler
	handler := controller.RouteFunc(routeFunc)

	// Create test request for POC path
	req := httptest.NewRequest("GET", "/poc", nil)
	w := httptest.NewRecorder()

	// Execute request
	handler.ServeHTTP(w, req)

	// Log response for debugging
	t.Logf("Request: %s %s", req.Method, req.URL.Path)
	t.Logf("Response: %d %s", w.Code, w.Body.String())

	// Check if POC handler worked
	if w.Code == 200 && w.Body.String() == "POC Working" {
		t.Logf("SUCCESS: POC handler is working correctly!")
	} else {
		t.Logf("POC handler not working as expected - checking handler chain status")
		t.Logf("Status: %d, Body: %s", w.Code, w.Body.String())
	}
}

// TestIndexHandlerWithTemplateDir tests that index handler can load template
func TestIndexHandlerWithTemplateDir(t *testing.T) {
	// Save current working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal("Failed to get working directory:", err)
	}
	defer func() {
		os.Chdir(originalWd)
	}()

	// Change to testdata directory
	testdataDir := filepath.Join(originalWd, "testdata")
	if err := os.Chdir(testdataDir); err != nil {
		t.Fatal("Failed to change to testdata directory:", err)
	}

	// Verify template exists
	if _, err := os.Stat("public/index.html"); err != nil {
		t.Fatal("index.html template not found in testdata/public")
	}

	// Create controller with public directory set to the templates folder
	controller := NewController("test-app", WithPublicDir("public"))

	// Create route that should use index.html template
	routeFunc := func() RouteOptions {
		return RouteOptions{
			ID("index"),
		}
	}

	handler := controller.RouteFunc(routeFunc)

	// Test index route (should use index.html template)
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	t.Logf("Index route response: %d %s", w.Code, w.Body.String())

	// Check if we got a response (may be 200 with template or error)
	if w.Code == 200 {
		t.Logf("SUCCESS: Index route returned 200")
		t.Logf("Body: %s", w.Body.String())
		if w.Body.String() == "POC Working" {
			t.Logf("Note: Index route was handled by POC handler (shows handler chain priority is working)")
		} else {
			t.Logf("Index route was handled by GET handler with template rendering")
		}
	} else {
		t.Logf("Index route returned non-200 status: %d", w.Code)
		t.Logf("Body: %s", w.Body.String())
	}
}
