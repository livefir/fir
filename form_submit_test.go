package fir

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHandlePostFormResult tests the handlePostFormResult function coverage
func TestHandlePostFormResult(t *testing.T) {
	t.Run("successful form submission with redirect", func(t *testing.T) {
		ctrl := NewController("test")

		// Create a route with a form event handler
		testRoute := func() RouteOptions {
			return RouteOptions{
				ID("form-test"),
				Content(`<form method="post"><input name="username"></form>`),
				OnEvent("submit", func(ctx RouteContext) error {
					// Return nil for successful submission
					return nil
				}),
			}
		}

		handler := ctrl.RouteFunc(testRoute)

		// Create form data
		formData := url.Values{}
		formData.Set("username", "testuser")

		// Create POST request with form data
		req := httptest.NewRequest("POST", "/?event=submit", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		handler(w, req)

		// Should redirect on successful form submission
		assert.Equal(t, http.StatusFound, w.Code)
		assert.Equal(t, "/", w.Header().Get("Location"))
	})

	t.Run("form submission with error", func(t *testing.T) {
		ctrl := NewController("test")

		// Create a route with a form event handler that returns an error
		testRoute := func() RouteOptions {
			return RouteOptions{
				ID("form-test"),
				Content(`<form method="post"><input name="username"></form>`),
				OnEvent("submit", func(ctx RouteContext) error {
					// Return a field error
					return ctx.FieldError("username", errors.New("required"))
				}),
			}
		}

		handler := ctrl.RouteFunc(testRoute)

		// Create form data
		formData := url.Values{}
		formData.Set("username", "")

		// Create POST request with form data
		req := httptest.NewRequest("POST", "/?event=submit", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		handler(w, req)

		// Should render the page with errors (not redirect)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "form method")
	})

	t.Run("form submission returning routeData", func(t *testing.T) {
		ctrl := NewController("test")

		// Create a route with a form event handler that returns routeData
		testRoute := func() RouteOptions {
			return RouteOptions{
				ID("form-test"),
				Content(`<form method="post"><input name="username"></form>`),
				OnEvent("submit", func(ctx RouteContext) error {
					// Return routeData for redirect
					return ctx.Data(map[string]any{"success": true})
				}),
			}
		}

		handler := ctrl.RouteFunc(testRoute)

		// Create form data
		formData := url.Values{}
		formData.Set("username", "testuser")

		// Create POST request with form data
		req := httptest.NewRequest("POST", "/?event=submit", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		handler(w, req)

		// Should redirect when returning routeData
		assert.Equal(t, http.StatusFound, w.Code)
		assert.Equal(t, "/", w.Header().Get("Location"))
	})
}
