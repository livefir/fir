package fir

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestResetActionHandler tests the ResetActionHandler implementation
func TestResetActionHandler(t *testing.T) {
	handler := &ResetActionHandler{}

	// Test basic properties
	require.Equal(t, "reset", handler.Name())
	require.Equal(t, 35, handler.Precedence())

	// Test translation
	tests := []struct {
		name     string
		value    string
		expected string
		wantErr  bool
	}{
		{
			name:     "Basic event",
			value:    "create-chirp",
			expected: `@fir:create-chirp:ok.nohtml="$el.reset()"`,
			wantErr:  false,
		},
		{
			name:     "Event with state",
			value:    "submit:ok",
			expected: `@fir:submit:ok.nohtml="$el.reset()"`,
			wantErr:  false,
		},
		{
			name:     "Event with modifier",
			value:    "submit.debounce",
			expected: `@fir:submit:ok.debounce.nohtml="$el.reset()"`,
			wantErr:  false,
		},
		{
			name:     "Multiple events",
			value:    "create:ok,update:done",
			expected: `@fir:[create:ok,update:done].nohtml="$el.reset()"`,
			wantErr:  false,
		},
		{
			name:     "Event with target (ignored)",
			value:    "submit->myForm",
			expected: `@fir:submit:ok.nohtml="$el.reset()"`,
			wantErr:  false,
		},
		{
			name:     "Event with action target (ignored)",
			value:    "submit=>doSubmit",
			expected: `@fir:submit:ok.nohtml="$el.reset()"`,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := ActionInfo{
				ActionName: "reset",
				Value:      tt.value,
			}

			result, err := handler.Translate(info, map[string]string{})

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestToggleDisabledActionHandler tests the ToggleDisabledActionHandler implementation
func TestToggleDisabledActionHandler(t *testing.T) {
	handler := &ToggleDisabledActionHandler{}

	// Test basic properties
	require.Equal(t, "toggle-disabled", handler.Name())
	require.Equal(t, 34, handler.Precedence())

	// Test translation
	tests := []struct {
		name     string
		value    string
		expected string
		wantErr  bool
	}{
		{
			name:     "Basic event",
			value:    "submit",
			expected: `@fir:submit:ok.nohtml="$fir.toggleDisabled()"`,
			wantErr:  false,
		},
		{
			name:     "Event with state",
			value:    "save:pending",
			expected: `@fir:save:pending.nohtml="$fir.toggleDisabled()"`,
			wantErr:  false,
		},
		{
			name:     "Event with modifier",
			value:    "submit.debounce",
			expected: `@fir:submit:ok.debounce.nohtml="$fir.toggleDisabled()"`,
			wantErr:  false,
		},
		{
			name:     "Multiple events",
			value:    "save:pending,save:ok",
			expected: `@fir:[save:pending,save:ok].nohtml="$fir.toggleDisabled()"`,
			wantErr:  false,
		},
		{
			name:     "Event with target (ignored)",
			value:    "submit->myForm",
			expected: `@fir:submit:ok.nohtml="$fir.toggleDisabled()"`,
			wantErr:  false,
		},
		{
			name:     "Event with action target (ignored)",
			value:    "submit=>doSubmit",
			expected: `@fir:submit:ok.nohtml="$fir.toggleDisabled()"`,
			wantErr:  false,
		},
		{
			name:     "Complex multi-state scenario",
			value:    "save:pending,save:ok,save:error",
			expected: `@fir:[save:pending,save:ok,save:error].nohtml="$fir.toggleDisabled()"`,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := ActionInfo{
				ActionName: "toggle-disabled",
				Value:      tt.value,
			}

			result, err := handler.Translate(info, map[string]string{})

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestTriggerActionHandler tests the TriggerActionHandler implementation
func TestTriggerActionHandler(t *testing.T) {
	handler := &TriggerActionHandler{}

	// Test basic properties
	require.Equal(t, "runjs", handler.Name())
	require.Equal(t, 32, handler.Precedence())

	// Test translation
	tests := []struct {
		name       string
		info       ActionInfo
		actionsMap map[string]string
		expected   string
		wantErr    bool
	}{
		{
			name: "Basic trigger with single event",
			info: ActionInfo{
				ActionName: "trigger",
				Params:     []string{"resetForm"},
				Value:      "submit",
			},
			actionsMap: map[string]string{
				"resetForm": "$fir.resetForm()",
			},
			expected: `@fir:submit:ok.nohtml="$fir.resetForm()"`,
			wantErr:  false,
		},
		{
			name: "Trigger with multiple events",
			info: ActionInfo{
				ActionName: "trigger",
				Params:     []string{"resetForm"},
				Value:      "inc,dec",
			},
			actionsMap: map[string]string{
				"resetForm": "$fir.resetForm()",
			},
			expected: `@fir:[inc:ok,dec:ok].nohtml="$fir.resetForm()"`,
			wantErr:  false,
		},
		{
			name: "Trigger with events having states",
			info: ActionInfo{
				ActionName: "trigger",
				Params:     []string{"clearData"},
				Value:      "save:ok,load:error",
			},
			actionsMap: map[string]string{
				"clearData": "$fir.clearData()",
			},
			expected: `@fir:[save:ok,load:error].nohtml="$fir.clearData()"`,
			wantErr:  false,
		},
		{
			name: "Trigger with events having modifiers (nohtml is added)",
			info: ActionInfo{
				ActionName: "trigger",
				Params:     []string{"updateForm"},
				Value:      "submit.debounce",
			},
			actionsMap: map[string]string{
				"updateForm": "$fir.updateForm()",
			},
			expected: `@fir:submit:ok.debounce.nohtml="$fir.updateForm()"`,
			wantErr:  false,
		},
		{
			name: "Trigger with mixed events and modifiers",
			info: ActionInfo{
				ActionName: "trigger",
				Params:     []string{"processData"},
				Value:      "create:ok.debounce,delete:error.throttle",
			},
			actionsMap: map[string]string{
				"processData": "$fir.processData()",
			},
			expected: `@fir:[create:ok,delete:error].debounce.nohtml.throttle="$fir.processData()"`,
			wantErr:  false,
		},
		{
			name: "Trigger with complex action value",
			info: ActionInfo{
				ActionName: "trigger",
				Params:     []string{"validateAndSubmit"},
				Value:      "submit",
			},
			actionsMap: map[string]string{
				"validateAndSubmit": "validate() && $fir.submit()",
			},
			expected: `@fir:submit:ok.nohtml="validate() && $fir.submit()"`,
			wantErr:  false,
		},
		{
			name: "Trigger with event targets (ignored)",
			info: ActionInfo{
				ActionName: "trigger",
				Params:     []string{"resetForm"},
				Value:      "submit->myForm",
			},
			actionsMap: map[string]string{
				"resetForm": "$fir.resetForm()",
			},
			expected: `@fir:submit:ok.nohtml="$fir.resetForm()"`,
			wantErr:  false,
		},
		{
			name: "Trigger with event action targets (ignored)",
			info: ActionInfo{
				ActionName: "trigger",
				Params:     []string{"resetForm"},
				Value:      "submit=>doSubmit",
			},
			actionsMap: map[string]string{
				"resetForm": "$fir.resetForm()",
			},
			expected: `@fir:submit:ok.nohtml="$fir.resetForm()"`,
			wantErr:  false,
		},
		// Error cases
		{
			name: "Error: No parameters",
			info: ActionInfo{
				ActionName: "trigger",
				Params:     []string{},
				Value:      "submit",
			},
			actionsMap: map[string]string{},
			wantErr:    true,
		},
		{
			name: "Error: Too many parameters",
			info: ActionInfo{
				ActionName: "trigger",
				Params:     []string{"resetForm", "extraParam"},
				Value:      "submit",
			},
			actionsMap: map[string]string{},
			wantErr:    true,
		},
		{
			name: "Error: Empty parameter",
			info: ActionInfo{
				ActionName: "trigger",
				Params:     []string{""},
				Value:      "submit",
			},
			actionsMap: map[string]string{},
			wantErr:    true,
		},
		{
			name: "Error: Whitespace-only parameter",
			info: ActionInfo{
				ActionName: "trigger",
				Params:     []string{"   "},
				Value:      "submit",
			},
			actionsMap: map[string]string{},
			wantErr:    true,
		},
		{
			name: "Error: Action not found in map",
			info: ActionInfo{
				ActionName: "trigger",
				Params:     []string{"missingAction"},
				Value:      "submit",
			},
			actionsMap: map[string]string{
				"existingAction": "$fir.existing()",
			},
			wantErr: true,
		},
		{
			name: "Error: Empty action value",
			info: ActionInfo{
				ActionName: "trigger",
				Params:     []string{"emptyAction"},
				Value:      "submit",
			},
			actionsMap: map[string]string{
				"emptyAction": "",
			},
			wantErr: true,
		},
		{
			name: "Error: Whitespace-only action value",
			info: ActionInfo{
				ActionName: "trigger",
				Params:     []string{"whitespaceAction"},
				Value:      "submit",
			},
			actionsMap: map[string]string{
				"whitespaceAction": "   ",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handler.Translate(tt.info, tt.actionsMap)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestToggleClassActionHandler tests the ToggleClassActionHandler implementation
func TestToggleClassActionHandler(t *testing.T) {
	handler := &ToggleClassActionHandler{}

	// Test basic properties
	require.Equal(t, "toggleClass", handler.Name())
	require.Equal(t, 33, handler.Precedence())

	// Test translation
	tests := []struct {
		name     string
		params   []string
		value    string
		expected string
		wantErr  bool
	}{
		{
			name:     "Single class",
			params:   []string{"is-loading"},
			value:    "submit",
			expected: `@fir:submit:ok.nohtml="$fir.toggleClass('is-loading')"`,
			wantErr:  false,
		},
		{
			name:     "Multiple classes",
			params:   []string{"is-loading", "is-active"},
			value:    "save",
			expected: `@fir:save:ok.nohtml="$fir.toggleClass('is-loading','is-active')"`,
			wantErr:  false,
		},
		{
			name:     "Event with state",
			params:   []string{"is-disabled"},
			value:    "save:pending",
			expected: `@fir:save:pending.nohtml="$fir.toggleClass('is-disabled')"`,
			wantErr:  false,
		},
		{
			name:     "Multiple events",
			params:   []string{"is-loading"},
			value:    "save:pending,save:ok",
			expected: `@fir:[save:pending,save:ok].nohtml="$fir.toggleClass('is-loading')"`,
			wantErr:  false,
		},
		{
			name:     "Complex multi-state scenario",
			params:   []string{"is-loading"},
			value:    "save:pending,save:ok,save:error",
			expected: `@fir:[save:pending,save:ok,save:error].nohtml="$fir.toggleClass('is-loading')"`,
			wantErr:  false,
		},
		{
			name:     "Event with modifier",
			params:   []string{"is-loading"},
			value:    "submit.debounce",
			expected: `@fir:submit:ok.debounce.nohtml="$fir.toggleClass('is-loading')"`,
			wantErr:  false,
		},
		{
			name:     "Event with target (ignored)",
			params:   []string{"is-loading"},
			value:    "submit->myForm",
			expected: `@fir:submit:ok.nohtml="$fir.toggleClass('is-loading')"`,
			wantErr:  false,
		},
		{
			name:     "Error: No class names",
			params:   []string{},
			value:    "submit",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := ActionInfo{
				ActionName: "toggleClass",
				Params:     tt.params,
				Value:      tt.value,
			}

			result, err := handler.Translate(info, map[string]string{})

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestActionsConflict tests the actionsConflict function
func TestActionsConflict(t *testing.T) {
	tests := []struct {
		name     string
		action1  string
		value1   string
		action2  string
		value2   string
		expected bool
	}{
		// Refresh conflicts with remove/remove-parent
		{
			name:     "refresh conflicts with remove",
			action1:  "refresh",
			value1:   "query:ok",
			action2:  "remove",
			value2:   "delete:ok",
			expected: true,
		},
		{
			name:     "refresh conflicts with remove-parent",
			action1:  "refresh",
			value1:   "query:ok",
			action2:  "remove-parent",
			value2:   "delete:ok",
			expected: true,
		},
		{
			name:     "remove conflicts with refresh",
			action1:  "remove",
			value1:   "delete:ok",
			action2:  "refresh",
			value2:   "query:ok",
			expected: true,
		},
		{
			name:     "remove-parent conflicts with refresh",
			action1:  "remove-parent",
			value1:   "delete:ok",
			action2:  "refresh",
			value2:   "query:ok",
			expected: true,
		},

		// Remove and remove-parent conflict with each other
		{
			name:     "remove conflicts with remove-parent",
			action1:  "remove",
			value1:   "delete:ok",
			action2:  "remove-parent",
			value2:   "delete:ok",
			expected: true,
		},
		{
			name:     "remove-parent conflicts with remove",
			action1:  "remove-parent",
			value1:   "delete:ok",
			action2:  "remove",
			value2:   "delete:ok",
			expected: true,
		},

		// Same events - these should conflict
		{
			name:     "append conflicts with prepend on same event",
			action1:  "append",
			value1:   "create:ok",
			action2:  "prepend",
			value2:   "create:ok",
			expected: true,
		},
		{
			name:     "replace conflicts with append on same event",
			action1:  "refresh", // refresh is the replace action
			value1:   "update:ok",
			action2:  "append",
			value2:   "update:ok",
			expected: true,
		},

		// Different events - these should not conflict except for DOM manipulation actions
		{
			name:     "append conflicts with prepend even on different events",
			action1:  "append",
			value1:   "create:ok",
			action2:  "prepend",
			value2:   "update:ok",
			expected: true, // DOM manipulation actions always conflict due to precedence
		},
		{
			name:     "refresh doesn't conflict with append on different events",
			action1:  "refresh",
			value1:   "update:ok",
			action2:  "append",
			value2:   "create:ok",
			expected: false,
		},

		// Multiple events - should conflict if any overlap
		{
			name:     "multiple events with overlap should conflict",
			action1:  "append",
			value1:   "create:ok,update:ok",
			action2:  "prepend",
			value2:   "update:ok,delete:ok",
			expected: true,
		},
		{
			name:     "multiple DOM manipulation events always conflict",
			action1:  "append",
			value1:   "create:ok,save:ok",
			action2:  "prepend",
			value2:   "update:ok,delete:ok",
			expected: true, // DOM manipulation actions always conflict due to precedence
		},

		// Mixed single and multiple events
		{
			name:     "single event conflicts with multiple containing it",
			action1:  "append",
			value1:   "create:ok",
			action2:  "prepend",
			value2:   "create:ok,update:ok,delete:ok",
			expected: true,
		},

		// Complex event expressions
		{
			name:     "complex events with modifiers should conflict on same base event",
			action1:  "append",
			value1:   "create:ok.debounce",
			action2:  "prepend",
			value2:   "create:ok.throttle",
			expected: true,
		},

		// Actions that can coexist
		{
			name:     "refresh and append on different events don't conflict",
			action1:  "refresh",
			value1:   "query:ok",
			action2:  "append",
			value2:   "create:ok",
			expected: false,
		},
		{
			name:     "reset and toggle-disabled don't conflict",
			action1:  "reset",
			value1:   "submit:ok",
			action2:  "toggle-disabled",
			value2:   "submit:pending",
			expected: false,
		},

		// Edge cases
		{
			name:     "same action on same event should conflict",
			action1:  "append",
			value1:   "create:ok",
			action2:  "append",
			value2:   "create:ok",
			expected: true,
		},
		{
			name:     "same action on different events should not conflict",
			action1:  "append",
			value1:   "create:ok",
			action2:  "append",
			value2:   "update:ok",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create collectedAction structs for testing
			action1 := collectedAction{
				Handler: &RefreshActionHandler{}, // Use a dummy handler
				Info: ActionInfo{
					ActionName: tt.action1,
					Value:      tt.value1,
				},
			}
			action2 := collectedAction{
				Handler: &RefreshActionHandler{}, // Use a dummy handler
				Info: ActionInfo{
					ActionName: tt.action2,
					Value:      tt.value2,
				},
			}

			result := actionsConflict(action1, action2)
			require.Equal(t, tt.expected, result, "actionsConflict(%s, %s, %s, %s) = %v, expected %v",
				tt.action1, tt.value1, tt.action2, tt.value2, result, tt.expected)
		})
	}
}

// TestParseEventExpression tests the parseEventExpression function
func TestParseEventExpression(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		expected []string
	}{
		{
			name:     "single event",
			expr:     "create:ok",
			expected: []string{"create:ok"},
		},
		{
			name:     "multiple events",
			expr:     "create:ok,update:ok,delete:ok",
			expected: []string{"create:ok", "update:ok", "delete:ok"},
		},
		{
			name:     "events with spaces",
			expr:     "create:ok, update:ok , delete:ok",
			expected: []string{"create:ok", "update:ok", "delete:ok"},
		},
		{
			name:     "event with modifiers",
			expr:     "create:ok.debounce",
			expected: []string{"create:ok"},
		},
		{
			name:     "multiple events with modifiers",
			expr:     "create:ok.debounce,update:ok.throttle",
			expected: []string{"create:ok", "update:ok"},
		},
		{
			name:     "event with target (should be ignored)",
			expr:     "create:ok->myTarget",
			expected: []string{"create:ok"},
		},
		{
			name:     "event with action target (should be ignored)",
			expr:     "create:ok=>myAction",
			expected: []string{"create:ok"},
		},
		{
			name:     "complex mixed expression",
			expr:     "create:ok.debounce->target, update:ok.throttle=>action, delete:error",
			expected: []string{"create:ok", "update:ok", "delete:error"},
		},
		{
			name:     "empty expression",
			expr:     "",
			expected: []string{},
		},
		{
			name:     "expression with only spaces",
			expr:     "   ",
			expected: []string{},
		},
		{
			name:     "expression with empty segments",
			expr:     "create:ok,,update:ok",
			expected: []string{"create:ok", "update:ok"},
		},
		{
			name:     "single event without state (defaults to :ok)",
			expr:     "create",
			expected: []string{"create:ok"},
		},
		{
			name:     "multiple events without state",
			expr:     "create,update,delete",
			expected: []string{"create:ok", "update:ok", "delete:ok"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseEventExpression(tt.expr)
			require.Equal(t, tt.expected, result, "parseEventExpression(%s) = %v, expected %v",
				tt.expr, result, tt.expected)
		})
	}
}

// TestHasCommonEvents tests the hasCommonEvents function
func TestHasCommonEvents(t *testing.T) {
	tests := []struct {
		name     string
		events1  []string
		events2  []string
		expected bool
	}{
		{
			name:     "identical single events",
			events1:  []string{"create:ok"},
			events2:  []string{"create:ok"},
			expected: true,
		},
		{
			name:     "different single events",
			events1:  []string{"create:ok"},
			events2:  []string{"update:ok"},
			expected: false,
		},
		{
			name:     "one event in common",
			events1:  []string{"create:ok", "update:ok"},
			events2:  []string{"update:ok", "delete:ok"},
			expected: true,
		},
		{
			name:     "no events in common",
			events1:  []string{"create:ok", "save:ok"},
			events2:  []string{"update:ok", "delete:ok"},
			expected: false,
		},
		{
			name:     "all events in common",
			events1:  []string{"create:ok", "update:ok"},
			events2:  []string{"create:ok", "update:ok"},
			expected: true,
		},
		{
			name:     "subset relationship",
			events1:  []string{"create:ok"},
			events2:  []string{"create:ok", "update:ok", "delete:ok"},
			expected: true,
		},
		{
			name:     "empty first list",
			events1:  []string{},
			events2:  []string{"create:ok", "update:ok"},
			expected: false,
		},
		{
			name:     "empty second list",
			events1:  []string{"create:ok", "update:ok"},
			events2:  []string{},
			expected: false,
		},
		{
			name:     "both lists empty",
			events1:  []string{},
			events2:  []string{},
			expected: false,
		},
		{
			name:     "case sensitivity test",
			events1:  []string{"Create:OK"},
			events2:  []string{"create:ok"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasCommonEvents(tt.events1, tt.events2)
			require.Equal(t, tt.expected, result, "hasCommonEvents(%v, %v) = %v, expected %v",
				tt.events1, tt.events2, result, tt.expected)
		})
	}
}

// TestActionsConflictIntegration tests the integration of conflict resolution with processRenderAttributes
func TestActionsConflictIntegration(t *testing.T) {
	tests := []struct {
		name        string
		action1     string
		value1      string
		action2     string
		value2      string
		description string
	}{
		{
			name:        "non-conflicting actions should both be processed",
			action1:     "refresh",
			value1:      "query:ok",
			action2:     "append",
			value2:      "create:ok",
			description: "refresh and append on different events should not conflict",
		},
		{
			name:        "conflicting actions should conflict",
			action1:     "refresh",
			value1:      "query:ok",
			action2:     "remove",
			value2:      "query:ok",
			description: "refresh and remove on same event should conflict",
		},
		{
			name:        "multiple non-conflicting actions",
			action1:     "reset",
			value1:      "submit:ok",
			action2:     "toggle-disabled",
			value2:      "submit:pending",
			description: "reset and toggle-disabled on different event states should not conflict",
		},
		{
			name:        "complex conflicting scenario",
			action1:     "append",
			value1:      "create:ok",
			action2:     "prepend",
			value2:      "create:ok",
			description: "append and prepend conflict on same event",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create collectedAction structs
			action1 := collectedAction{
				Handler: &RefreshActionHandler{},
				Info: ActionInfo{
					ActionName: tt.action1,
					Value:      tt.value1,
				},
			}
			action2 := collectedAction{
				Handler: &RefreshActionHandler{},
				Info: ActionInfo{
					ActionName: tt.action2,
					Value:      tt.value2,
				},
			}

			// Test the conflict resolution
			result := actionsConflict(action1, action2)

			// Verify parseEventExpression works
			events1 := parseEventExpression(tt.value1)
			events2 := parseEventExpression(tt.value2)
			require.NotNil(t, events1)
			require.NotNil(t, events2)

			// Verify hasCommonEvents works
			hasCommon := hasCommonEvents(events1, events2)

			// Log for debugging
			t.Logf("Action1: %s=%s, Action2: %s=%s", tt.action1, tt.value1, tt.action2, tt.value2)
			t.Logf("Events1: %v, Events2: %v", events1, events2)
			t.Logf("HasCommonEvents: %v, ActionsConflict: %v", hasCommon, result)
		})
	}
}
