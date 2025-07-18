package actions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestActionHandlerRegistry tests the internal registry functionality
func TestActionHandlerRegistry(t *testing.T) {
	// Save original registry
	originalRegistry := make(map[string]ActionHandler)
	for k, v := range Registry {
		originalRegistry[k] = v
	}

	// Clear registry for test
	Registry = make(map[string]ActionHandler)

	// Test registration
	handler := &RefreshActionHandler{}
	RegisterHandler(handler)

	// Test that handler was registered
	registered, exists := Registry["refresh"]
	require.True(t, exists)
	require.Equal(t, handler, registered)

	// Test duplicate registration panics
	require.Panics(t, func() {
		RegisterHandler(&RefreshActionHandler{})
	})

	// Restore original registry
	Registry = originalRegistry
}

// TestActionHandlerInterface tests the action handler interface implementation
func TestActionHandlerInterface(t *testing.T) {
	testCases := []struct {
		name      string
		handler   ActionHandler
		info      ActionInfo
		actions   map[string]string
		expectErr bool
	}{
		{
			name:    "RefreshActionHandler",
			handler: &RefreshActionHandler{},
			info: ActionInfo{
				ActionName: "refresh",
				Value:      "update",
			},
			actions:   nil,
			expectErr: false,
		},
		{
			name:    "RemoveActionHandler",
			handler: &RemoveActionHandler{},
			info: ActionInfo{
				ActionName: "remove",
				Value:      "delete",
			},
			actions:   nil,
			expectErr: false,
		},
		{
			name:    "ToggleClassActionHandler with valid params",
			handler: &ToggleClassActionHandler{},
			info: ActionInfo{
				ActionName: "toggleClass",
				Params:     []string{"visible", "hidden"},
				Value:      "toggle",
			},
			actions:   nil,
			expectErr: false,
		},
		{
			name:    "ToggleClassActionHandler with no params",
			handler: &ToggleClassActionHandler{},
			info: ActionInfo{
				ActionName: "toggleClass",
				Params:     []string{},
				Value:      "toggle",
			},
			actions:   nil,
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.handler.Translate(tc.info, tc.actions)
			
			if tc.expectErr {
				require.Error(t, err)
				require.Empty(t, result)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, result)
			}
		})
	}
}

// TestCollectedActionSorting tests the CollectedAction sorting functionality
func TestCollectedActionSorting(t *testing.T) {
	actions := []CollectedAction{
		{Handler: &PrependActionHandler{}, Info: ActionInfo{ActionName: "prepend"}}, // precedence 60
		{Handler: &RefreshActionHandler{}, Info: ActionInfo{ActionName: "refresh"}}, // precedence 20
		{Handler: &AppendActionHandler{}, Info: ActionInfo{ActionName: "append"}},   // precedence 50
		{Handler: &RemoveActionHandler{}, Info: ActionInfo{ActionName: "remove"}},   // precedence 30
	}

	SortActionsByPrecedence(actions)

	// Should be sorted by precedence (lowest first)
	require.Equal(t, "refresh", actions[0].Info.ActionName) // 20
	require.Equal(t, "remove", actions[1].Info.ActionName)  // 30
	require.Equal(t, "append", actions[2].Info.ActionName)  // 50
	require.Equal(t, "prepend", actions[3].Info.ActionName) // 60
}

// TestParseTranslatedStringInternal tests the internal ParseTranslatedString function
func TestParseTranslatedStringInternal(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Single attribute",
			input:    `@fir:update:ok="$fir.replace()"`,
			expected: []string{`@fir:update:ok="$fir.replace()"`},
		},
		{
			name: "Multiple attributes",
			input: `@fir:update:ok="$fir.replace()"
@fir:update:error="console.error('Failed')"`,
			expected: []string{
				`@fir:update:ok="$fir.replace()"`,
				`@fir:update:error="console.error('Failed')"`,
			},
		},
		{
			name:     "Empty input",
			input:    "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseTranslatedString(tt.input)
			require.Len(t, result, len(tt.expected))
			for i, expected := range tt.expected {
				require.Equal(t, expected, result[i].Key+`="`+result[i].Val+`"`)
			}
		})
	}
}
