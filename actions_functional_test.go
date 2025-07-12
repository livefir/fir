package fir

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestActionsFunctional_WorkflowScenarios tests real-world action workflows and chaining
func TestActionsFunctional_WorkflowScenarios(t *testing.T) {
	tests := []struct {
		name            string
		template        string
		workflowType    string
		expectedActions []string
		description     string
	}{
		{
			name: "form submission workflow",
			template: `
				<form id="user-form" 
					  x-fir-toggle-disabled="submit:pending" 
					  x-fir-reset="submit:ok"
					  x-fir-toggleClass:error="submit:error">
					<input type="text" name="username" required />
					<button type="submit" 
						    x-fir-toggleClass:is-loading="submit:pending"
							x-fir-dispatch:[form-validated]="submit:ok">
						Submit
					</button>
				</form>
			`,
			workflowType: "form_submission",
			expectedActions: []string{
				`@fir:submit:pending="$fir.toggleDisabled()"`,
				`@fir:submit:ok="$el.reset()"`,
				`@fir:submit:error="$fir.toggleClass('error')"`,
				`@fir:submit:pending="$fir.toggleClass('is-loading')"`,
				`@fir:submit:ok="$dispatch('form-validated')"`,
			},
			description: "Complex form with loading states, validation, and notifications",
		},
		{
			name: "dynamic list management workflow",
			template: `
				<div id="todo-container">
					<button x-fir-append:todo-item="add-todo"
						    x-fir-toggleClass:success="add-todo:ok"
							x-fir-dispatch:[list-updated]="add-todo:ok">
						Add Todo
					</button>
					<ul id="todo-list" 
						x-fir-refresh="refresh-list"
						x-fir-toggleClass:empty="refresh-list:ok">
					</ul>
				</div>
			`,
			workflowType: "list_management",
			expectedActions: []string{
				`@fir:add-todo:ok::todo-item="$fir.appendEl()"`,
				`@fir:add-todo:ok="$fir.toggleClass('success')"`,
				`@fir:add-todo:ok="$dispatch('list-updated')"`,
				`@fir:refresh-list:ok="$fir.replace()"`,
				`@fir:refresh-list:ok="$fir.toggleClass('empty')"`,
			},
			description: "Todo list with dynamic addition, refresh, and state management",
		},
		{
			name: "modal and navigation workflow",
			template: `
				<div id="modal-overlay" 
					 x-fir-remove="close-modal"
					 x-fir-toggleClass:visible="show-modal">
					<div class="modal-content">
						<button x-fir-dispatch:[modal-closed]="close-modal"
								x-fir-remove-parent="close-modal">
							Close
						</button>
						<form x-fir-dispatch:[form-submitted]="submit"
							  x-fir-remove-parent="submit:ok">
							<input type="text" name="data" />
							<button type="submit">Save</button>
						</form>
					</div>
				</div>
			`,
			workflowType: "modal_navigation",
			expectedActions: []string{
				`@fir:close-modal:ok="$fir.removeEl()"`,
				`@fir:show-modal:ok="$fir.toggleClass('visible')"`,
				`@fir:close-modal:ok="$dispatch('modal-closed')"`,
				`@fir:close-modal:ok="$fir.removeParentEl()"`,
				`@fir:submit:ok="$dispatch('form-submitted')"`,
				`@fir:submit:ok="$fir.removeParentEl()"`,
			},
			description: "Modal with form submission and cleanup actions",
		},
		{
			name: "progressive disclosure workflow",
			template: `
				<div id="disclosure-container">
					<button x-fir-toggleClass:expanded="toggle-details"
						    x-fir-dispatch:[section-toggled]="toggle-details">
						Toggle Details
					</button>
					<div class="details-section" 
						 x-fir-refresh="load-details"
						 x-fir-toggleClass:loading="load-details:pending">
						<div x-fir-append:detail-item="load-details:ok">
							Details will appear here
						</div>
					</div>
				</div>
			`,
			workflowType: "progressive_disclosure",
			expectedActions: []string{
				`@fir:toggle-details:ok="$fir.toggleClass('expanded')"`,
				`@fir:toggle-details:ok="$dispatch('section-toggled')"`,
				`@fir:load-details:ok="$fir.replace()"`,
				`@fir:load-details:pending="$fir.toggleClass('loading')"`,
				`@fir:load-details:ok::detail-item="$fir.appendEl()"`,
			},
			description: "Progressive disclosure with lazy loading and state transitions",
		},
		{
			name: "notification system workflow",
			template: `
				<div id="notification-area">
					<div class="notification" 
						 x-fir-prepend:alert="show-notification"
						 x-fir-toggleClass:visible="show-notification"
						 x-fir-remove="hide-notification"
						 x-fir-dispatch:[notification-shown]="show-notification">
						<button x-fir-dispatch:[notification-dismissed]="hide-notification">
							Dismiss
						</button>
					</div>
				</div>
			`,
			workflowType: "notification_system",
			expectedActions: []string{
				`@fir:show-notification:ok::alert="$fir.prependEl()"`,
				`@fir:show-notification:ok="$fir.toggleClass('visible')"`,
				`@fir:hide-notification:ok="$fir.removeEl()"`,
				`@fir:show-notification:ok="$dispatch('notification-shown')"`,
				`@fir:hide-notification:ok="$dispatch('notification-dismissed')"`,
			},
			description: "Notification system with show/hide animations and tracking",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := NewController("test-functional-" + tt.workflowType)

			handler := ctrl.RouteFunc(func() RouteOptions {
				return RouteOptions{
					ID("workflow-route"),
					Content(tt.template),
				}
			})

			req := httptest.NewRequest("GET", "/", nil)
			resp := httptest.NewRecorder()
			handler(resp, req)

			require.Equal(t, http.StatusOK, resp.Code, "Workflow should render successfully")

			html := resp.Body.String()

			// Verify all expected actions are present
			for _, expectedAction := range tt.expectedActions {
				require.Contains(t, html, expectedAction,
					"Expected action missing in %s workflow: %s", tt.workflowType, expectedAction)
			}

			// Verify no x-fir-* attributes remain in the output
			require.NotContains(t, html, "x-fir-", "All x-fir attributes should be translated")

			t.Logf("Workflow '%s' successfully rendered %d actions", tt.workflowType, len(tt.expectedActions))
		})
	}
}

// TestActionsFunctional_EdgeCaseWorkflows tests edge cases and complex combinations
func TestActionsFunctional_EdgeCaseWorkflows(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		description string
		shouldError bool
	}{
		{
			name: "deeply nested action inheritance",
			template: `
				<div x-fir-refresh="update">
					<div x-fir-append:item="create">
						<div x-fir-toggleClass:active="select">
							<button x-fir-dispatch:[action-triggered]="click">
								Nested Action
							</button>
						</div>
					</div>
				</div>
			`,
			description: "Multiple levels of nested elements with different actions",
			shouldError: false,
		},
		{
			name: "action overloading scenario",
			template: `
				<div x-fir-refresh="event1,event2"
					 x-fir-append:template="event3"
					 x-fir-toggleClass:[class1,class2]="event4,event5"
					 x-fir-dispatch:[msg1,msg2]="event6">
					Complex element
				</div>
			`,
			description: "Single element with multiple actions on different events",
			shouldError: false,
		},
		{
			name: "conditional action chains",
			template: `
				<div x-fir-toggleClass:loading="start:pending"
					 x-fir-toggleClass:success="start:ok"
					 x-fir-toggleClass:error="start:error"
					 x-fir-refresh="start:ok"
					 x-fir-dispatch:[status-changed]="start:pending,start:ok,start:error">
					State-driven element
				</div>
			`,
			description: "Element with conditional actions based on event states",
			shouldError: false,
		},

		{
			name: "invalid parameter combinations",
			template: `
				<div x-fir-toggleClass="submit"
					 x-fir-dispatch="click">
					Invalid parameters
				</div>
			`,
			description: "Should fail due to missing required parameters",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := NewController("test-edge-case")

			handler := ctrl.RouteFunc(func() RouteOptions {
				return RouteOptions{
					ID("edge-case-route"),
					Content(tt.template),
				}
			})

			req := httptest.NewRequest("GET", "/", nil)
			resp := httptest.NewRecorder()
			handler(resp, req)

			if tt.shouldError {
				require.NotEqual(t, http.StatusOK, resp.Code,
					"Expected error for: %s", tt.description)
			} else {
				require.Equal(t, http.StatusOK, resp.Code,
					"Unexpected error for: %s", tt.description)

				html := resp.Body.String()
				require.NotContains(t, html, "x-fir-",
					"All x-fir attributes should be translated for: %s", tt.description)
			}
		})
	}
}

// TestActionsFunctional_PerformanceScenarios tests action processing under load
func TestActionsFunctional_PerformanceScenarios(t *testing.T) {
	tests := []struct {
		name                string
		templateGenerator   func(int) string
		elementCount        int
		expectedActionCount int
		description         string
	}{
		{
			name: "many elements with single actions",
			templateGenerator: func(count int) string {
				html := "<div>"
				for i := 0; i < count; i++ {
					html += `<div x-fir-refresh="update">Element ` + string(rune(i)) + `</div>`
				}
				html += "</div>"
				return html
			},
			elementCount:        100,
			expectedActionCount: 100,
			description:         "100 elements each with refresh action",
		},
		{
			name: "few elements with many actions each",
			templateGenerator: func(count int) string {
				html := "<div>"
				for i := 0; i < count; i++ {
					html += `<div x-fir-refresh="update" x-fir-toggleClass:active="select" x-fir-dispatch:[msg]="click">Complex Element</div>`
				}
				html += "</div>"
				return html
			},
			elementCount:        20,
			expectedActionCount: 60, // 20 elements * 3 actions each
			description:         "20 elements each with 3 different actions",
		},
		{
			name: "nested structure with cascading actions",
			templateGenerator: func(depth int) string {
				html := ""
				for i := 0; i < depth; i++ {
					html += `<div x-fir-refresh="level` + fmt.Sprintf("%d", i) + `">`
				}
				html += "Deeply nested content"
				for i := 0; i < depth; i++ {
					html += "</div>"
				}
				return html
			},
			elementCount:        10, // depth levels
			expectedActionCount: 10,
			description:         "10 levels of nested elements with different refresh events",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := tt.templateGenerator(tt.elementCount)

			ctrl := NewController("test-performance")

			handler := ctrl.RouteFunc(func() RouteOptions {
				return RouteOptions{
					ID("performance-route"),
					Content(template),
				}
			})

			req := httptest.NewRequest("GET", "/", nil)
			resp := httptest.NewRecorder()
			handler(resp, req)

			require.Equal(t, http.StatusOK, resp.Code, "Performance test should succeed")

			html := resp.Body.String()

			// Count @fir: occurrences as a proxy for action count
			actionCount := 0
			for i := 0; i < len(html); i++ {
				if i+5 <= len(html) && html[i:i+5] == "@fir:" {
					actionCount++
				}
			}

			require.Equal(t, tt.expectedActionCount, actionCount,
				"Expected %d actions but found %d in %s",
				tt.expectedActionCount, actionCount, tt.description)

			// Verify no x-fir-* attributes remain
			require.NotContains(t, html, "x-fir-", "All x-fir attributes should be processed")

			t.Logf("Performance test '%s' processed %d actions successfully",
				tt.name, actionCount)
		})
	}
}

// TestActionsFunctional_StateTransitions tests complex state-based action workflows
func TestActionsFunctional_StateTransitions(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		stateScenario  string
		expectedStates []string
		description    string
	}{
		{
			name: "loading to success transition",
			template: `
				<button x-fir-toggleClass:loading="submit:pending"
						x-fir-toggleClass:success="submit:ok"
						x-fir-toggleClass:error="submit:error"
						x-fir-toggle-disabled="disable,enable">
					Submit Data
				</button>
			`,
			stateScenario: "submit_workflow",
			expectedStates: []string{
				`@fir:submit:pending="$fir.toggleClass('loading')"`,
				`@fir:submit:ok="$fir.toggleClass('success')"`,
				`@fir:submit:error="$fir.toggleClass('error')"`,
				`@fir:disable:ok="$fir.toggleDisabled()"`,
				`@fir:enable:ok="$fir.toggleDisabled()"`,
			},
			description: "Button with loading, success, and error state transitions",
		},
		{
			name: "multi-step form progression",
			template: `
				<div id="form-step-1" 
					 x-fir-toggleClass:hidden="next-step"
					 x-fir-dispatch:[step-changed]="next-step">
					<button x-fir-dispatch:[validate-step1]="next-step">Next</button>
				</div>
				<div id="form-step-2" 
					 x-fir-toggleClass:visible="next-step"
					 x-fir-refresh="load-step2">
					<button x-fir-dispatch:[form-complete]="finish">Complete</button>
				</div>
			`,
			stateScenario: "multi_step_form",
			expectedStates: []string{
				`@fir:next-step:ok="$fir.toggleClass('hidden')"`,
				`@fir:next-step:ok="$dispatch('step-changed')"`,
				`@fir:next-step:ok="$dispatch('validate-step1')"`,
				`@fir:next-step:ok="$fir.toggleClass('visible')"`,
				`@fir:load-step2:ok="$fir.replace()"`,
				`@fir:finish:ok="$dispatch('form-complete')"`,
			},
			description: "Multi-step form with progressive disclosure and validation",
		},
		{
			name: "async operation with retry logic",
			template: `
				<div x-fir-toggleClass:trying="attempt:pending"
					 x-fir-toggleClass:failed="attempt:error"
					 x-fir-append:retry-button="attempt:error"
					 x-fir-refresh="attempt:ok"
					 x-fir-dispatch:[operation-complete]="attempt:ok"
					 x-fir-dispatch:[retry-needed]="attempt:error">
					<button x-fir-dispatch:[start-operation]="click">Start Operation</button>
				</div>
			`,
			stateScenario: "async_with_retry",
			expectedStates: []string{
				`@fir:attempt:pending="$fir.toggleClass('trying')"`,
				`@fir:attempt:error="$fir.toggleClass('failed')"`,
				`@fir:attempt:error::retry-button="$fir.appendEl()"`,
				`@fir:attempt:ok="$fir.replace()"`,
				`@fir:attempt:ok="$dispatch('operation-complete')"`,
				`@fir:attempt:error="$dispatch('retry-needed')"`,
				`@fir:click:ok="$dispatch('start-operation')"`,
			},
			description: "Async operation with pending, success, error, and retry states",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := NewController("test-state-transitions")

			handler := ctrl.RouteFunc(func() RouteOptions {
				return RouteOptions{
					ID("state-route"),
					Content(tt.template),
				}
			})

			req := httptest.NewRequest("GET", "/", nil)
			resp := httptest.NewRecorder()
			handler(resp, req)

			require.Equal(t, http.StatusOK, resp.Code, "State transition test should succeed")

			html := resp.Body.String()

			// Verify all expected state transitions are present
			for _, expectedState := range tt.expectedStates {
				require.Contains(t, html, expectedState,
					"Expected state transition missing in %s: %s", tt.stateScenario, expectedState)
			}

			// Verify no x-fir-* attributes remain
			require.NotContains(t, html, "x-fir-", "All x-fir attributes should be translated")

			t.Logf("State scenario '%s' implemented %d state transitions",
				tt.stateScenario, len(tt.expectedStates))
		})
	}
}
