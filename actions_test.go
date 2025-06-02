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
