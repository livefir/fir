package fir

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestActionHandlers tests the basic action handler implementations
func TestActionHandlers(t *testing.T) {
	t.Run("RefreshActionHandler", func(t *testing.T) {
		handler := &RefreshActionHandler{}
		require.Equal(t, "refresh", handler.Name())
		require.Equal(t, 20, handler.Precedence())

		info := ActionInfo{
			ActionName: "refresh",
			Value:      "update",
		}
		result, err := handler.Translate(info, nil)
		require.NoError(t, err)
		require.Equal(t, `@fir:update:ok="$fir.replace()"`, result)
	})

	t.Run("RemoveActionHandler", func(t *testing.T) {
		handler := &RemoveActionHandler{}
		require.Equal(t, "remove", handler.Name())
		require.Equal(t, 30, handler.Precedence())

		info := ActionInfo{
			ActionName: "remove",
			Value:      "delete",
		}
		result, err := handler.Translate(info, nil)
		require.NoError(t, err)
		require.Equal(t, `@fir:delete:ok="$fir.removeEl()"`, result)
	})

	t.Run("AppendActionHandler", func(t *testing.T) {
		handler := &AppendActionHandler{}
		require.Equal(t, "append", handler.Name())
		require.Equal(t, 50, handler.Precedence())

		info := ActionInfo{
			ActionName: "append",
			Params:     []string{"todo"},
			Value:      "create",
		}
		result, err := handler.Translate(info, nil)
		require.NoError(t, err)
		require.Equal(t, `@fir:create:ok::todo="$fir.appendEl()"`, result)
	})

	t.Run("PrependActionHandler", func(t *testing.T) {
		handler := &PrependActionHandler{}
		require.Equal(t, "prepend", handler.Name())
		require.Equal(t, 60, handler.Precedence())

		info := ActionInfo{
			ActionName: "prepend",
			Params:     []string{"header"},
			Value:      "create",
		}
		result, err := handler.Translate(info, nil)
		require.NoError(t, err)
		require.Equal(t, `@fir:create:ok::header="$fir.prependEl()"`, result)
	})

	t.Run("RemoveParentActionHandler", func(t *testing.T) {
		handler := &RemoveParentActionHandler{}
		require.Equal(t, "remove-parent", handler.Name())
		require.Equal(t, 40, handler.Precedence())

		info := ActionInfo{
			ActionName: "remove-parent",
			Value:      "delete",
		}
		result, err := handler.Translate(info, nil)
		require.NoError(t, err)
		require.Equal(t, `@fir:delete:ok="$fir.removeParentEl()"`, result)
	})

	t.Run("ResetActionHandler", func(t *testing.T) {
		handler := &ResetActionHandler{}
		require.Equal(t, "reset", handler.Name())
		require.Equal(t, 35, handler.Precedence())

		info := ActionInfo{
			ActionName: "reset",
			Value:      "submit",
		}
		result, err := handler.Translate(info, map[string]string{})
		require.NoError(t, err)
		require.Equal(t, `@fir:submit:ok="$el.reset()"`, result)
	})

	t.Run("ToggleDisabledActionHandler", func(t *testing.T) {
		handler := &ToggleDisabledActionHandler{}
		require.Equal(t, "toggle-disabled", handler.Name())
		require.Equal(t, 34, handler.Precedence())

		info := ActionInfo{
			ActionName: "toggle-disabled",
			Value:      "submit",
		}
		result, err := handler.Translate(info, map[string]string{})
		require.NoError(t, err)
		require.Equal(t, `@fir:submit:ok="$fir.toggleDisabled()"`, result)
	})

	t.Run("ToggleClassActionHandler", func(t *testing.T) {
		handler := &ToggleClassActionHandler{}
		require.Equal(t, "toggleClass", handler.Name())
		require.Equal(t, 33, handler.Precedence())

		info := ActionInfo{
			ActionName: "toggleClass",
			Params:     []string{"is-loading"},
			Value:      "submit",
		}
		result, err := handler.Translate(info, map[string]string{})
		require.NoError(t, err)
		require.Equal(t, `@fir:submit:ok="$fir.toggleClass('is-loading')"`, result)
	})

	t.Run("TriggerActionHandler", func(t *testing.T) {
		handler := &TriggerActionHandler{}
		require.Equal(t, "runjs", handler.Name())
		require.Equal(t, 32, handler.Precedence())

		info := ActionInfo{
			ActionName: "trigger",
			Params:     []string{"resetForm"},
			Value:      "submit",
		}
		actionsMap := map[string]string{
			"resetForm": "$fir.resetForm()",
		}
		result, err := handler.Translate(info, actionsMap)
		require.NoError(t, err)
		require.Equal(t, `@fir:submit:ok="$fir.resetForm()"`, result)
	})

	t.Run("DispatchActionHandler", func(t *testing.T) {
		handler := &DispatchActionHandler{}
		require.Equal(t, "dispatch", handler.Name())
		require.Equal(t, 33, handler.Precedence())

		info := ActionInfo{
			ActionName: "dispatch",
			Params:     []string{"modal-close"},
			Value:      "click",
		}
		result, err := handler.Translate(info, nil)
		require.NoError(t, err)
		require.Equal(t, `@fir:click:ok="$dispatch('modal-close')"`, result)
	})

	t.Run("ActionPrefixHandler", func(t *testing.T) {
		handler := &ActionPrefixHandler{}
		require.Equal(t, "js", handler.Name())
		require.Equal(t, 100, handler.Precedence())

		info := ActionInfo{
			ActionName: "js",
			Value:      "click",
		}
		result, err := handler.Translate(info, nil)
		require.NoError(t, err)
		require.Empty(t, result) // Prefix handler returns empty result
	})
}

// TestActionRegistry tests action handler registration
func TestActionRegistry(t *testing.T) {
	// Save original registry
	originalRegistry := make(map[string]ActionHandler)
	for k, v := range actionRegistry {
		originalRegistry[k] = v
	}

	// Clear registry for test
	actionRegistry = make(map[string]ActionHandler)

	// Test registration
	handler := &RefreshActionHandler{}
	RegisterActionHandler(handler)

	// Test that handler was registered
	registered, exists := actionRegistry["refresh"]
	require.True(t, exists)
	require.Equal(t, handler, registered)

	// Test duplicate registration panics
	require.Panics(t, func() {
		RegisterActionHandler(&RefreshActionHandler{})
	})

	// Restore original registry
	actionRegistry = originalRegistry
}

// TestSortActionsByPrecedence tests the precedence sorting
func TestSortActionsByPrecedence(t *testing.T) {
	actions := []collectedAction{
		{Handler: &PrependActionHandler{}, Info: ActionInfo{ActionName: "prepend"}},   // precedence 60
		{Handler: &RefreshActionHandler{}, Info: ActionInfo{ActionName: "refresh"}},   // precedence 20
		{Handler: &AppendActionHandler{}, Info: ActionInfo{ActionName: "append"}},     // precedence 50
		{Handler: &RemoveActionHandler{}, Info: ActionInfo{ActionName: "remove"}},     // precedence 30
	}

	sortActionsByPrecedence(actions)

	// Should be sorted by precedence (lowest first)
	require.Equal(t, "refresh", actions[0].Info.ActionName)   // 20
	require.Equal(t, "remove", actions[1].Info.ActionName)    // 30
	require.Equal(t, "append", actions[2].Info.ActionName)    // 50
	require.Equal(t, "prepend", actions[3].Info.ActionName)   // 60
}

// TestParseTranslatedString tests the helper function for parsing translated strings
func TestParseTranslatedString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Single attribute",
			input:    `@fir:click:ok="$fir.refresh()"`,
			expected: []string{`@fir:click:ok="$fir.refresh()"`},
		},
		{
			name:  "Multiple attributes",
			input: "@fir:click:ok=\"$fir.refresh()\"\n@fir:submit:ok=\"$fir.submit()\"",
			expected: []string{
				`@fir:click:ok="$fir.refresh()"`,
				`@fir:submit:ok="$fir.submit()"`,
			},
		},
		{
			name:     "Empty input",
			input:    "",
			expected: []string{},
		},
		{
			name:     "Whitespace only",
			input:    "   \n   \n   ",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTranslatedString(tt.input)
			require.Len(t, result, len(tt.expected))
			for i, expected := range tt.expected {
				require.Equal(t, expected, result[i].Key+`="`+result[i].Val+`"`)
			}
		})
	}
}
