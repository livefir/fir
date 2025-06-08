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
