package fir

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandlerChainPublicAPIIntegration implements Step 3.1 of the migration guide:
// Integration tests using public APIs instead of internal methods
func TestHandlerChainPublicAPIIntegration(t *testing.T) {
	tests := []struct {
		name          string
		method        string
		path          string
		contentType   string
		headers       map[string]string
		body          string
		routeOptions  func() RouteOptions
		expectSuccess bool
		description   string
	}{
		{
			name:        "WebSocket upgrade request (expected to fail in test)",
			method:      "GET",
			path:        "/ws",
			contentType: "text/html",
			headers: map[string]string{
				"Upgrade":               "websocket",
				"Connection":            "Upgrade",
				"Sec-WebSocket-Key":     "dGhlIHNhbXBsZSBub25jZQ==",
				"Sec-WebSocket-Version": "13",
			},
			routeOptions: func() RouteOptions {
				return RouteOptions{
					ID("websocket-route"),
					Content("<div>WebSocket content</div>"),
				}
			},
			expectSuccess: false, // WebSocket fails in test environment (no hijacker)
			description:   "WebSocket requests attempt WebSocketHandler but fail in test environment",
		},
		{
			name:        "JSON event submission",
			method:      "POST",
			path:        "/",
			contentType: "application/json",
			body:        `{"event_id": "test-event", "params": {"data": "test"}}`,
			routeOptions: func() RouteOptions {
				return RouteOptions{
					ID("json-route"),
					Content("<div>JSON content</div>"),
					OnEvent("test-event", func(ctx RouteContext) error {
						// Event handler implementation
						return nil
					}),
				}
			},
			expectSuccess: true,
			description:   "JSON events should be handled by JSONEventHandler in handler chain",
		},
		{
			name:        "Form submission",
			method:      "POST",
			path:        "/",
			contentType: "application/x-www-form-urlencoded",
			body:        "event_id=form-submit&data=test",
			routeOptions: func() RouteOptions {
				return RouteOptions{
					ID("form-route"),
					Content("<form>Form content</form>"),
					OnEvent("form-submit", func(ctx RouteContext) error {
						// Form handler implementation
						return nil
					}),
				}
			},
			expectSuccess: true,
			description:   "Form submissions should be handled by FormHandler in handler chain",
		},
		{
			name:        "GET request with onLoad",
			method:      "GET",
			path:        "/",
			contentType: "text/html",
			routeOptions: func() RouteOptions {
				return RouteOptions{
					ID("get-with-onload"),
					Content("<div>GET content</div>"),
					OnLoad(func(ctx RouteContext) error {
						// OnLoad handler implementation
						return nil
					}),
				}
			},
			expectSuccess: true,
			description:   "GET requests with onLoad should be handled by GetHandler in handler chain",
		},
		{
			name:        "GET request without onLoad (legacy fallback)",
			method:      "GET",
			path:        "/",
			contentType: "text/html",
			routeOptions: func() RouteOptions {
				return RouteOptions{
					ID("get-legacy"),
					Content("<div>Legacy content</div>"),
					// No OnLoad - should fall back to legacy ServeHTTP
				}
			},
			expectSuccess: true,
			description:   "GET requests without onLoad should fall back to legacy ServeHTTP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create controller using public API
			ctrl := NewController("test-integration")

			// Create route handler using public RouteFunc API
			handler := ctrl.RouteFunc(tt.routeOptions)

			// Create HTTP request
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", tt.contentType)

			// Add additional headers if provided
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			resp := httptest.NewRecorder()

			// Execute request using public ServeHTTP API
			handler(resp, req)

			// Verify response
			if tt.expectSuccess {
				// Allow both success and redirects for realistic behavior
				assert.True(t, resp.Code < 400 || resp.Code == 302, "Should not return client/server error: %s", tt.description)

				// For non-redirect responses, verify response body is generated
				body := resp.Body.String()
				if resp.Code != 302 {
					assert.NotEmpty(t, body, "Should generate response body for non-redirect: %s", tt.description)
				}
			} else {
				// Expecting failure (like WebSocket in test environment)
				assert.True(t, resp.Code >= 400, "Should return error status: %s", tt.description)
			}

			t.Logf("Test '%s' completed - Status: %d, Description: %s", tt.name, resp.Code, tt.description)
		})
	}
}

// TestHandlerChainCoveragePublicAPI verifies handler chain coverage diagnostics through public API
func TestHandlerChainCoveragePublicAPI(t *testing.T) {
	t.Run("Handler chain coverage verification", func(t *testing.T) {
		// Create controller
		ctrl := NewController("test-coverage")

		// Create route with comprehensive event handling
		handler := ctrl.RouteFunc(func() RouteOptions {
			return RouteOptions{
				ID("coverage-test"),
				Content("<div>Coverage test content</div>"),
				OnEvent("test-event", func(ctx RouteContext) error {
					return nil
				}),
				OnLoad(func(ctx RouteContext) error {
					return nil
				}),
			}
		})

		// Test different request types to verify coverage diagnostics
		testRequests := []struct {
			name        string
			method      string
			contentType string
			body        string
			headers     map[string]string
		}{
			{
				name:        "WebSocket upgrade",
				method:      "GET",
				contentType: "text/html",
				headers: map[string]string{
					"Upgrade":    "websocket",
					"Connection": "Upgrade",
				},
			},
			{
				name:        "JSON event",
				method:      "POST",
				contentType: "application/json",
				body:        `{"event_id": "test-event"}`,
			},
			{
				name:        "Form submission",
				method:      "POST",
				contentType: "application/x-www-form-urlencoded",
				body:        "event_id=test-event",
			},
			{
				name:        "GET with onLoad",
				method:      "GET",
				contentType: "text/html",
			},
		}

		for _, tr := range testRequests {
			t.Run(tr.name, func(t *testing.T) {
				req := httptest.NewRequest(tr.method, "/", strings.NewReader(tr.body))
				req.Header.Set("Content-Type", tr.contentType)

				for key, value := range tr.headers {
					req.Header.Set(key, value)
				}

				resp := httptest.NewRecorder()
				handler(resp, req)

				// Verify that coverage diagnostics work (no internal server errors)
				require.True(t, resp.Code < 500, "Coverage diagnostics should not cause errors")

				// Note: The actual coverage diagnostics are logged internally
				// This test verifies the system remains stable during coverage checking
				t.Logf("Coverage test for %s: Status %d", tr.name, resp.Code)
			})
		}
	})
}

// TestHandlerChainFallbackPublicAPI verifies graceful fallback behavior through public API
func TestHandlerChainFallbackPublicAPI(t *testing.T) {
	t.Run("Graceful fallback scenarios", func(t *testing.T) {
		// Create controller
		ctrl := NewController("test-fallback")

		// Create route without modern handlers (should trigger legacy fallback)
		handler := ctrl.RouteFunc(func() RouteOptions {
			return RouteOptions{
				ID("fallback-test"),
				Content("<div>Fallback content</div>"),
				// Intentionally no OnEvent or OnLoad handlers
				// This should trigger fallback to legacy ServeHTTP
			}
		})

		// Test requests that should fall back to legacy behavior
		fallbackRequests := []struct {
			name        string
			method      string
			contentType string
			body        string
		}{
			{
				name:        "POST without event handlers",
				method:      "POST",
				contentType: "application/json",
				body:        `{"event_id": "unknown-event"}`,
			},
			{
				name:        "GET without onLoad",
				method:      "GET",
				contentType: "text/html",
			},
			{
				name:        "Form POST without handlers",
				method:      "POST",
				contentType: "application/x-www-form-urlencoded",
				body:        "event_id=unknown-event",
			},
		}

		for _, fr := range fallbackRequests {
			t.Run(fr.name, func(t *testing.T) {
				req := httptest.NewRequest(fr.method, "/", strings.NewReader(fr.body))
				req.Header.Set("Content-Type", fr.contentType)

				resp := httptest.NewRecorder()
				handler(resp, req)

				// Verify graceful fallback (no crashes, generates response)
				assert.True(t, resp.Code < 500, "Fallback should not cause server errors")

				body := resp.Body.String()
				assert.NotEmpty(t, body, "Fallback should generate response body")

				t.Logf("Fallback test for %s: Status %d", fr.name, resp.Code)
			})
		}
	})
}
