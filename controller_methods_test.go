package fir

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestController_RouteFunc(t *testing.T) {
	// Test Controller.RouteFunc method (this is part of the interface)
	ctrl := NewController("test-route")

	// Create a simple route function
	routeFunc := func() RouteOptions {
		return RouteOptions{
			Content("<div>Test Route</div>"),
		}
	}

	// Test RouteFunc method
	handler := ctrl.RouteFunc(routeFunc)
	assert.NotNil(t, handler)

	// The handler should be callable as http.HandlerFunc
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	// Should not panic and should handle the request
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestController_MultipleRouteFuncs(t *testing.T) {
	// Test creating multiple route functions with the same controller
	ctrl := NewController("multi-route")

	routeFunc1 := func() RouteOptions {
		return RouteOptions{
			ID("route1"),
			Content("<div>Route 1</div>"),
		}
	}

	routeFunc2 := func() RouteOptions {
		return RouteOptions{
			ID("route2"),
			Content("<div>Route 2</div>"),
		}
	}

	handler1 := ctrl.RouteFunc(routeFunc1)
	handler2 := ctrl.RouteFunc(routeFunc2)

	assert.NotNil(t, handler1)
	assert.NotNil(t, handler2)

	// Verify they're different instances by testing their behavior
	req := httptest.NewRequest("GET", "/", nil)

	w1 := httptest.NewRecorder()
	handler1(w1, req)

	w2 := httptest.NewRecorder()
	handler2(w2, req)

	// Both should work without panicking
	assert.Equal(t, http.StatusOK, w1.Code)
	assert.Equal(t, http.StatusOK, w2.Code)
}

func TestController_RouteFuncWithOptions(t *testing.T) {
	// Test Controller.RouteFunc with various route options
	ctrl := NewController("test-options",
		WithSessionName("test-session"),
		EnableDebugLog(),
	)

	routeFunc := func() RouteOptions {
		return RouteOptions{
			ID("test-route"),
			Content("<div>Test Content</div>"),
			Extensions(".html", ".gohtml"),
		}
	}

	handler := ctrl.RouteFunc(routeFunc)
	assert.NotNil(t, handler)

	// Test that the handler can be called
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	// Should handle the request successfully
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestController_RouteFuncWithEvents(t *testing.T) {
	// Test RouteFunc with event handlers
	ctrl := NewController("test-events")

	routeFunc := func() RouteOptions {
		return RouteOptions{
			ID("event-route"),
			Content("<div>{{ .message }}</div>"),
			OnLoad(func(ctx RouteContext) error {
				return ctx.KV("message", "Loaded")
			}),
			OnEvent("click", func(ctx RouteContext) error {
				return ctx.KV("message", "Clicked")
			}),
		}
	}

	handler := ctrl.RouteFunc(routeFunc)
	assert.NotNil(t, handler)

	// Test GET request (OnLoad)
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Loaded")
}

func TestController_EmptyRouteFunc(t *testing.T) {
	// Test with empty route function
	ctrl := NewController("empty-route")

	emptyRouteFunc := func() RouteOptions {
		return RouteOptions{}
	}

	handler := ctrl.RouteFunc(emptyRouteFunc)
	assert.NotNil(t, handler)

	// Should still work with empty options
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	// Should not panic
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestController_RouteFuncWithDifferentControllerOptions(t *testing.T) {
	// Test RouteFunc with different controller configurations
	testCases := []struct {
		name string
		ctrl Controller
	}{
		{
			name: "minimal controller",
			ctrl: NewController("minimal"),
		},
		{
			name: "controller with session",
			ctrl: NewController("session", WithSessionName("custom-session")),
		},
		{
			name: "controller with debug",
			ctrl: NewController("debug", EnableDebugLog()),
		},
		{
			name: "development controller",
			ctrl: NewController("dev", DevelopmentMode(true)),
		},
	}

	routeFunc := func() RouteOptions {
		return RouteOptions{
			Content("<div>Test</div>"),
		}
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := tc.ctrl.RouteFunc(routeFunc)
			assert.NotNil(t, handler)

			// Test that handler works
			req := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()
			handler(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}
