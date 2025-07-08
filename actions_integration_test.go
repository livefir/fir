package fir

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestActionsIntegration_Refresh tests refresh action via HTTP request/response cycles
func TestActionsIntegration_Refresh(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		expectedInHTML []string
	}{
		{
			name: "basic refresh on event",
			template: `<div id="content" x-fir-refresh="update">Original content</div>`,
			expectedInHTML: []string{`@fir:update:ok="$fir.replace()"`},
		},
		{
			name: "refresh with event state",
			template: `<div id="content" x-fir-refresh="load:done">Loading...</div>`,
			expectedInHTML: []string{`@fir:load:done="$fir.replace()"`},
		},
		{
			name: "refresh with multiple events",
			template: `<div id="content" x-fir-refresh="create:ok,update:ok">Dynamic content</div>`,
			expectedInHTML: []string{`@fir:create:ok="$fir.replace()"`, `@fir:update:ok="$fir.replace()"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create controller and route
			ctrl := NewController("test-actions-refresh")
			
			handler := ctrl.RouteFunc(func() RouteOptions {
				return RouteOptions{
					ID("test-route"),
					Content(tt.template),
				}
			})

			// Perform HTTP request
			req := httptest.NewRequest("GET", "/", nil)
			resp := httptest.NewRecorder()
			handler(resp, req)

			// Verify HTTP response
			require.Equal(t, http.StatusOK, resp.Code)
			
			// Verify the translated action is in the HTML
			html := resp.Body.String()
			for _, expected := range tt.expectedInHTML {
				require.Contains(t, html, expected, "Expected action not found in HTML")
			}

			// Verify x-fir-refresh attribute is removed
			require.NotContains(t, html, "x-fir-refresh", "Original x-fir attribute should be removed")
		})
	}
}

// TestActionsIntegration_Remove tests remove action via HTTP request/response cycles
func TestActionsIntegration_Remove(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		expectedInHTML []string
	}{
		{
			name:           "basic remove on event",
			template:       `<button x-fir-remove="delete">Delete Item</button>`,
			expectedInHTML: []string{`@fir:delete:ok="$fir.removeEl()"`},
		},
		{
			name:           "remove with complex event",
			template:       `<div x-fir-remove="cleanup:done,timeout:error">Temporary content</div>`,
			expectedInHTML: []string{`@fir:cleanup:done="$fir.removeEl()"`, `@fir:timeout:error="$fir.removeEl()"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := NewController("test-actions-remove")
			
			handler := ctrl.RouteFunc(func() RouteOptions {
				return RouteOptions{
					ID("test-route"),
					Content(tt.template),
				}
			})

			req := httptest.NewRequest("GET", "/", nil)
			resp := httptest.NewRecorder()
			handler(resp, req)

			require.Equal(t, http.StatusOK, resp.Code)
			
			html := resp.Body.String()
			for _, expected := range tt.expectedInHTML {
				require.Contains(t, html, expected)
			}
			require.NotContains(t, html, "x-fir-remove")
		})
	}
}

// TestActionsIntegration_Append tests append action via HTTP request/response cycles
func TestActionsIntegration_Append(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		expectedInHTML []string
	}{
		{
			name:           "append with template parameter",
			template:       `<div x-fir-append:todo="create"><ul id="todo-list"></ul></div>`,
			expectedInHTML: []string{`@fir:create:ok::todo="$fir.appendEl()"`},
		},
		{
			name:           "append without template parameter",
			template:       `<div x-fir-append="create"><ul id="items"></ul></div>`,
			expectedInHTML: []string{`@fir:create:ok="$fir.appendEl()"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := NewController("test-actions-append")
			
			handler := ctrl.RouteFunc(func() RouteOptions {
				return RouteOptions{
					ID("test-route"),
					Content(tt.template),
				}
			})

			req := httptest.NewRequest("GET", "/", nil)
			resp := httptest.NewRecorder()
			handler(resp, req)

			require.Equal(t, http.StatusOK, resp.Code)
			
			html := resp.Body.String()
			for _, expected := range tt.expectedInHTML {
				require.Contains(t, html, expected)
			}
			require.NotContains(t, html, "x-fir-append")
		})
	}
}

// TestActionsIntegration_Prepend tests prepend action via HTTP request/response cycles
func TestActionsIntegration_Prepend(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		expectedInHTML []string
	}{
		{
			name:           "prepend with template parameter",
			template:       `<div x-fir-prepend:notification="alert"><div id="alerts"></div></div>`,
			expectedInHTML: []string{`@fir:alert:ok::notification="$fir.prependEl()"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := NewController("test-actions-prepend")
			
			handler := ctrl.RouteFunc(func() RouteOptions {
				return RouteOptions{
					ID("test-route"),
					Content(tt.template),
				}
			})

			req := httptest.NewRequest("GET", "/", nil)
			resp := httptest.NewRecorder()
			handler(resp, req)

			require.Equal(t, http.StatusOK, resp.Code)
			
			html := resp.Body.String()
			for _, expected := range tt.expectedInHTML {
				require.Contains(t, html, expected)
			}
			require.NotContains(t, html, "x-fir-prepend")
		})
	}
}

// TestActionsIntegration_Reset tests reset action via HTTP request/response cycles
func TestActionsIntegration_Reset(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		expectedInHTML []string
	}{
		{
			name:           "reset form on submit",
			template:       `<form x-fir-reset="submit:ok"><input type="text" name="name" /><button type="submit">Submit</button></form>`,
			expectedInHTML: []string{`@fir:submit:ok="$el.reset()"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := NewController("test-actions-reset")
			
			handler := ctrl.RouteFunc(func() RouteOptions {
				return RouteOptions{
					ID("test-route"),
					Content(tt.template),
				}
			})

			req := httptest.NewRequest("GET", "/", nil)
			resp := httptest.NewRecorder()
			handler(resp, req)

			require.Equal(t, http.StatusOK, resp.Code)
			
			html := resp.Body.String()
			for _, expected := range tt.expectedInHTML {
				require.Contains(t, html, expected)
			}
			require.NotContains(t, html, "x-fir-reset")
		})
	}
}

// TestActionsIntegration_ToggleDisabled tests toggle-disabled action via HTTP request/response cycles
func TestActionsIntegration_ToggleDisabled(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		expectedInHTML []string
	}{
		{
			name:           "toggle disabled on save",
			template:       `<button x-fir-toggle-disabled="save:pending,save:ok">Save Data</button>`,
			expectedInHTML: []string{`@fir:save:pending="$fir.toggleDisabled()"`, `@fir:save:ok="$fir.toggleDisabled()"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := NewController("test-actions-toggle-disabled")
			
			handler := ctrl.RouteFunc(func() RouteOptions {
				return RouteOptions{
					ID("test-route"),
					Content(tt.template),
				}
			})

			req := httptest.NewRequest("GET", "/", nil)
			resp := httptest.NewRecorder()
			handler(resp, req)

			require.Equal(t, http.StatusOK, resp.Code)
			
			html := resp.Body.String()
			for _, expected := range tt.expectedInHTML {
				require.Contains(t, html, expected)
			}
			require.NotContains(t, html, "x-fir-toggle-disabled")
		})
	}
}

// TestActionsIntegration_ToggleClass tests toggleClass action via HTTP request/response cycles
func TestActionsIntegration_ToggleClass(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		expectedInHTML []string
	}{
		{
			name:           "toggle single class",
			template:       `<div x-fir-toggleClass:is-loading="submit">Content</div>`,
			expectedInHTML: []string{`@fir:submit:ok="$fir.toggleClass('is-loading')"`},
		},
		{
			name:           "toggle multiple classes",
			template:       `<div x-fir-toggleClass:[is-loading,is-active]="process">Content</div>`,
			expectedInHTML: []string{`@fir:process:ok="$fir.toggleClass('is-loading','is-active')"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := NewController("test-actions-toggle-class")
			
			handler := ctrl.RouteFunc(func() RouteOptions {
				return RouteOptions{
					ID("test-route"),
					Content(tt.template),
				}
			})

			req := httptest.NewRequest("GET", "/", nil)
			resp := httptest.NewRecorder()
			handler(resp, req)

			require.Equal(t, http.StatusOK, resp.Code)
			
			html := resp.Body.String()
			for _, expected := range tt.expectedInHTML {
				require.Contains(t, html, expected)
			}
			require.NotContains(t, html, "x-fir-toggleClass")
		})
	}
}

// TestActionsIntegration_Dispatch tests dispatch action via HTTP request/response cycles
func TestActionsIntegration_Dispatch(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		expectedInHTML []string
	}{
		{
			name:           "dispatch single event",
			template:       `<button x-fir-dispatch:[modal-close]="click">Close Modal</button>`,
			expectedInHTML: []string{`@fir:click:ok="$dispatch('modal-close')"`},
		},
		{
			name:           "dispatch multiple events",
			template:       `<button x-fir-dispatch:[toggle-sidebar,update-nav]="click">Toggle Sidebar</button>`,
			expectedInHTML: []string{`@fir:click:ok="$dispatch('toggle-sidebar','update-nav')"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := NewController("test-actions-dispatch")
			
			handler := ctrl.RouteFunc(func() RouteOptions {
				return RouteOptions{
					ID("test-route"),
					Content(tt.template),
				}
			})

			req := httptest.NewRequest("GET", "/", nil)
			resp := httptest.NewRecorder()
			handler(resp, req)

			require.Equal(t, http.StatusOK, resp.Code)
			
			html := resp.Body.String()
			for _, expected := range tt.expectedInHTML {
				require.Contains(t, html, expected)
			}
			require.NotContains(t, html, "x-fir-dispatch")
		})
	}
}

// TestActionsIntegration_MultipleActions tests multiple actions on the same element
func TestActionsIntegration_MultipleActions(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		expectedInHTML []string
		shouldError    bool
	}{
		{
			name:     "compatible actions on same element",
			template: `<form x-fir-reset="submit:ok" x-fir-toggle-disabled="submit:pending"><input type="text" name="data" /><button type="submit">Submit</button></form>`,
			expectedInHTML: []string{
				`@fir:submit:ok="$el.reset()"`,
				`@fir:submit:pending="$fir.toggleDisabled()"`,
			},
			shouldError: false,
		},
		{
			name:     "compatible actions on different events",
			template: `<div x-fir-refresh="update" x-fir-append:item="create"><div id="content">Content</div></div>`,
			expectedInHTML: []string{
				`@fir:update:ok="$fir.replace()"`,
				`@fir:create:ok::item="$fir.appendEl()"`,
			},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := NewController("test-actions-multiple")
			
			handler := ctrl.RouteFunc(func() RouteOptions {
				return RouteOptions{
					ID("test-route"),
					Content(tt.template),
				}
			})

			req := httptest.NewRequest("GET", "/", nil)
			resp := httptest.NewRecorder()
			handler(resp, req)

			if tt.shouldError {
				// Expect a 500 or error response for conflicting actions
				require.NotEqual(t, http.StatusOK, resp.Code)
			} else {
				require.Equal(t, http.StatusOK, resp.Code)
				
				html := resp.Body.String()
				for _, expected := range tt.expectedInHTML {
					require.Contains(t, html, expected)
				}
			}
		})
	}
}

// TestActionsIntegration_EventStateHandling tests how different event states are processed
func TestActionsIntegration_EventStateHandling(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		expectedInHTML []string
	}{
		{
			name:           "event without state defaults to :ok",
			template:       `<div x-fir-refresh="update">Content</div>`,
			expectedInHTML: []string{`@fir:update:ok="$fir.replace()"`},
		},
		{
			name:           "event with explicit state",
			template:       `<div x-fir-refresh="update:done">Content</div>`,
			expectedInHTML: []string{`@fir:update:done="$fir.replace()"`},
		},
		{
			name:           "multiple events with mixed states",
			template:       `<div x-fir-refresh="save:pending,save:ok,save:error">Content</div>`,
			expectedInHTML: []string{`@fir:save:pending="$fir.replace()"`, `@fir:save:ok="$fir.replace()"`, `@fir:save:error="$fir.replace()"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := NewController("test-actions-event-states")
			
			handler := ctrl.RouteFunc(func() RouteOptions {
				return RouteOptions{
					ID("test-route"),
					Content(tt.template),
				}
			})

			req := httptest.NewRequest("GET", "/", nil)
			resp := httptest.NewRecorder()
			handler(resp, req)

			require.Equal(t, http.StatusOK, resp.Code)
			
			html := resp.Body.String()
			for _, expected := range tt.expectedInHTML {
				require.Contains(t, html, expected)
			}
		})
	}
}

// TestActionsIntegration_ParameterValidation tests parameter validation in HTTP context
func TestActionsIntegration_ParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		shouldError bool
	}{
		{
			name:        "valid toggleClass parameters",
			template:    `<div x-fir-toggleClass:[is-loading,is-active]="submit">Content</div>`,
			shouldError: false,
		},
		{
			name:        "invalid toggleClass - no parameters",
			template:    `<div x-fir-toggleClass="submit">Content</div>`,
			shouldError: true,
		},
		{
			name:        "valid dispatch parameters",
			template:    `<button x-fir-dispatch:[modal-close]="click">Close</button>`,
			shouldError: false,
		},
		{
			name:        "invalid dispatch - no parameters",
			template:    `<button x-fir-dispatch="click">Close</button>`,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := NewController("test-actions-validation")
			
			handler := ctrl.RouteFunc(func() RouteOptions {
				return RouteOptions{
					ID("test-route"),
					Content(tt.template),
				}
			})

			req := httptest.NewRequest("GET", "/", nil)
			resp := httptest.NewRecorder()
			handler(resp, req)

			if tt.shouldError {
				require.NotEqual(t, http.StatusOK, resp.Code)
			} else {
				require.Equal(t, http.StatusOK, resp.Code)
			}
		})
	}
}
