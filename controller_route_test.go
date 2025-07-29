package fir

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRoute provides a test implementation of the Route interface
type TestRoute struct {
	options RouteOptions
}

func (tr *TestRoute) Options() RouteOptions {
	return tr.options
}

// TestController_Route tests the Route method of the controller
func TestController_Route(t *testing.T) {
	t.Run("Route with valid RouteOptions", func(t *testing.T) {
		ctrl := NewController("test")

		testRoute := &TestRoute{
			options: RouteOptions{
				ID("test-route"),
				Content("Hello Test Route"),
				OnLoad(func(ctx RouteContext) error {
					return ctx.KV("message", "loaded")
				}),
			},
		}

		handler := ctrl.Route(testRoute)
		assert.NotNil(t, handler)

		// Test the handler
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Hello Test Route")
	})

	t.Run("Route with Layout and Content", func(t *testing.T) {
		ctrl := NewController("test")

		testRoute := &TestRoute{
			options: RouteOptions{
				ID("layout-route"),
				Content("Test Title: {{.title}}"),
				OnLoad(func(ctx RouteContext) error {
					return ctx.KV("title", "Success")
				}),
			},
		}

		handler := ctrl.Route(testRoute)
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Test Title: Success")
	})

	t.Run("Route with OnEvent handler", func(t *testing.T) {
		ctrl := NewController("test")

		testRoute := &TestRoute{
			options: RouteOptions{
				ID("event-route"),
				Content(`<div @fir:test:ok="$fir.replace()">{{.count}}</div>`),
				OnLoad(func(ctx RouteContext) error {
					return ctx.KV("count", 0)
				}),
				OnEvent("test", func(ctx RouteContext) error {
					return ctx.KV("count", 42)
				}),
			},
		}

		handler := ctrl.Route(testRoute)

		// Test initial load
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "0")
	})

	t.Run("Route with multiple options", func(t *testing.T) {
		ctrl := NewController("test")

		testRoute := &TestRoute{
			options: RouteOptions{
				ID("multi-route"),
				Content("{{.data}}"),
				Extensions(".html", ".tmpl"),
				OnLoad(func(ctx RouteContext) error {
					return ctx.Data(map[string]any{
						"data": "multi-option test",
					})
				}),
			},
		}

		handler := ctrl.Route(testRoute)
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "multi-option test")
	})

	t.Run("Route with empty options", func(t *testing.T) {
		ctrl := NewController("test")

		testRoute := &TestRoute{
			options: RouteOptions{},
		}

		handler := ctrl.Route(testRoute)
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		handler(w, req)

		// Should use default content
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Hello Fir App!")
	})

	t.Run("Route with nil options", func(t *testing.T) {
		ctrl := NewController("test")

		testRoute := &TestRoute{
			options: nil,
		}

		handler := ctrl.Route(testRoute)
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		handler(w, req)

		// Should handle nil options gracefully
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Route registration in controller", func(t *testing.T) {
		ctrl := NewController("test")

		testRoute1 := &TestRoute{
			options: RouteOptions{
				ID("route1"),
				Content("Route 1"),
			},
		}

		testRoute2 := &TestRoute{
			options: RouteOptions{
				ID("route2"),
				Content("Route 2"),
			},
		}

		handler1 := ctrl.Route(testRoute1)
		handler2 := ctrl.Route(testRoute2)

		assert.NotNil(t, handler1)
		assert.NotNil(t, handler2)

		// Verify both routes work independently
		req1 := httptest.NewRequest("GET", "/", nil)
		w1 := httptest.NewRecorder()
		handler1(w1, req1)
		assert.Contains(t, w1.Body.String(), "Route 1")

		req2 := httptest.NewRequest("GET", "/", nil)
		w2 := httptest.NewRecorder()
		handler2(w2, req2)
		assert.Contains(t, w2.Body.String(), "Route 2")
	})

	t.Run("Route with custom FuncMap", func(t *testing.T) {
		ctrl := NewController("test")

		testRoute := &TestRoute{
			options: RouteOptions{
				ID("funcmap-route"),
				Content("{{upper .message}}"),
				FuncMap(map[string]any{
					"upper": func(s string) string {
						return "UPPER: " + s
					},
				}),
				OnLoad(func(ctx RouteContext) error {
					return ctx.KV("message", "test")
				}),
			},
		}

		handler := ctrl.Route(testRoute)
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "UPPER: test")
	})
}

// TestController_Route_ErrorHandling tests error scenarios in Route method
func TestController_Route_ErrorHandling(t *testing.T) {
	t.Run("Route with invalid template syntax", func(t *testing.T) {
		ctrl := NewController("test")

		testRoute := &TestRoute{
			options: RouteOptions{
				ID("invalid-route"),
				Content("{{invalid template syntax"),
			},
		}

		handler := ctrl.Route(testRoute)
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		handler(w, req)

		// Should return error status due to template parsing failure
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "route creation failed")
	})

	t.Run("Route with invalid layout template", func(t *testing.T) {
		ctrl := NewController("test")

		testRoute := &TestRoute{
			options: RouteOptions{
				ID("invalid-layout-route"),
				Layout("{{invalid layout"),
				Content("Valid content"),
			},
		}

		handler := ctrl.Route(testRoute)
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		handler(w, req)

		// Should return error status due to layout parsing failure
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "route creation failed")
	})

	t.Run("Route with error template access", func(t *testing.T) {
		ctrl := NewController("test")

		testRoute := &TestRoute{
			options: RouteOptions{
				ID("test-route"),
				Content("Hello Test Route"),
				ErrorContent("Error occurred"),
				OnEvent("test-error", func(ctx RouteContext) error {
					// Return an error to trigger error template rendering
					return ctx.FieldError("username", errors.New("required"))
				}),
			},
		}

		// Create the route which will parse both normal and error templates
		handler := ctrl.Route(testRoute)
		assert.NotNil(t, handler)

		// Test normal request first
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Hello Test Route")
	})
}
