package firattr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Assume Expressions, Binding, Eventexpression, Target structs are defined correctly
// matching the grammar, including Target using "=>" for Action.
// Assume getRenderExpressionParser() and parseRenderExpression() exist and work correctly.

func TestTranslateRenderExpression(t *testing.T) {

	tests := []struct {
		name     string
		input    string
		actions  map[string]string // Optional map for action replacement
		expected string
		wantErr  bool
	}{
		// --- Existing Basic Cases (Updated for Defaults) ---
		{
			name:     "single event, no target (default state, default action)",
			input:    "click",
			expected: `@fir:click:ok="$fir.replace()"`,
			wantErr:  false,
		},
		{
			name:     "single event with state, no target (default action)",
			input:    "click:ok",
			expected: `@fir:click:ok="$fir.replace()"`,
			wantErr:  false,
		},
		{
			name:     "single event, template target (default state, default action)",
			input:    "submit->myform",
			expected: `@fir:submit:ok::myform="$fir.replace()"`,
			wantErr:  false,
		},
		{
			name:     "single event, action target (default state)",
			input:    "submit=>doSubmit",
			expected: `@fir:submit:ok="doSubmit"`, // No map, action name used directly
			wantErr:  false,
		},
		{
			name:     "single event, template and action target (default state)",
			input:    "submit->myform=>doSubmit",
			expected: `@fir:submit:ok::myform="doSubmit"`, // No map, action name used directly
			wantErr:  false,
		},
		{
			name:     "multiple events, template and action target",
			input:    "create:ok,update:ok->mytemplate=>myfunction",
			expected: `@fir:[create:ok,update:ok]::mytemplate="myfunction"`, // No map, action name used directly
			wantErr:  false,
		},
		{
			name:     "multiple events, no state, template and action target (default state)",
			input:    "create,update->mytemplate=>myfunction",
			expected: `@fir:[create:ok,update:ok]::mytemplate="myfunction"`, // No map, action name used directly
			wantErr:  false,
		},
		{
			name:     "multiple events, no target (default action)",
			input:    "create:ok,update:ok",
			expected: `@fir:[create:ok,update:ok]="$fir.replace()"`,
			wantErr:  false,
		},
		{
			name:     "multiple events, no state, no target (default state, default action)",
			input:    "create,update",
			expected: `@fir:[create:ok,update:ok]="$fir.replace()"`,
			wantErr:  false,
		},
		// --- Modifiers (Updated for Defaults) ---
		{
			name:     "Single Event with Modifier, No Target (default state, default action)",
			input:    "create",
			expected: `@fir:create:ok="$fir.replace()"`,
			wantErr:  false,
		},
		{
			name:     "Single Event with State and Modifier, No Target (default action)",
			input:    "create:ok",
			expected: `@fir:create:ok="$fir.replace()"`,
			wantErr:  false,
		},
		{
			name:     "Single Event with Modifier and Template Target (default state, default action)",
			input:    "create->todo",
			expected: `@fir:create:ok::todo="$fir.replace()"`,
			wantErr:  false,
		},
		{
			name:     "Single Event with Modifier and Action Target (default state)",
			input:    "create.mod=>doAction",
			expected: `@fir:create:ok.mod="doAction"`, // No map, action name used directly
			wantErr:  false,
		},
		{
			name:     "Single Event with Modifier, Template, and Action Target (default state)",
			input:    "create.mod->view=>doAction",
			expected: `@fir:create:ok::view.mod="doAction"`, // No map, action name used directly
			wantErr:  false,
		},
		{
			name:     "Grouped Events with Different Modifiers, Template and Action Target",
			input:    "create:ok.debounce,update:error->template=>myaction",
			expected: `@fir:[create:ok,update:error]::template.debounce="myaction"`, // No map, action name used directly
			wantErr:  false,
		},
		{
			name:     "Grouped Events with Different Modifiers, No Target (default action)",
			input:    "create:ok.debounce,update:error",
			expected: `@fir:[create:ok,update:error].debounce="$fir.replace()"`,
			wantErr:  false,
		},
		{
			name:     "Grouped Events with Different Modifiers, Template Target Only (default action)",
			input:    "create:ok.debounce,update:error->template",
			expected: `@fir:[create:ok,update:error]::template.debounce="$fir.replace()"`,
			wantErr:  false,
		},
		{
			name:     "Grouped Events with Different Modifiers, Action Target Only",
			input:    "create:ok.debounce,update:error=>myaction",
			expected: `@fir:[create:ok,update:error].debounce="myaction"`, // No map, action name used directly
			wantErr:  false,
		},
		{
			name:     "Grouped Events, no state, Template Target Only (default state, default action)",
			input:    "create.debounce,update->template",
			expected: `@fir:[create:ok,update:ok]::template.debounce="$fir.replace()"`,
			wantErr:  false,
		},
		{
			name:     "Grouped Events with Various Valid States, Template Target Only (default action)",
			input:    "save:pending,load:done,check:error->form",
			expected: `@fir:[save:pending,load:done,check:error]::form="$fir.replace()"`,
			wantErr:  false,
		},
		// --- Grouped Bindings (Comma) / Separate Expressions (Semicolon) (Updated for Defaults) ---
		{
			name:     "Grouped Bindings (comma) with Modifiers and Targets - generates SINGLE line",
			input:    "create:ok.debounce,delete:error->todo=>replace",
			expected: `@fir:[create:ok,delete:error]::todo.debounce="replace"`, // No map, action name used directly
			wantErr:  false,
		},
		{
			name:     "Grouped Bindings (comma) No Target - generates SINGLE line (default action)",
			input:    "create:ok.debounce,delete:error",
			expected: `@fir:[create:ok,delete:error].debounce="$fir.replace()"`,
			wantErr:  false,
		},
		{
			name:     "Grouped Bindings (comma), no state, No Target (default state, default action)",
			input:    "create.debounce,delete",
			expected: `@fir:[create:ok,delete:ok].debounce="$fir.replace()"`,
			wantErr:  false,
		},
		{
			name:     "Grouped Bindings (comma), Template Only (default action)",
			input:    "create:ok.debounce,delete:error->todo",
			expected: `@fir:[create:ok,delete:error]::todo.debounce="$fir.replace()"`,
			wantErr:  false,
		},
		{
			name:     "Grouped Bindings (comma), Action Only",
			input:    "create:ok.debounce,delete:error=>replace",
			expected: `@fir:[create:ok,delete:error].debounce="replace"`, // No map, action name used directly
			wantErr:  false,
		},
		{
			name:     "Multiple Expressions (semicolon) with Modifiers (default action)",
			input:    "create:ok.debounce->todo;delete:error=>replace",
			expected: "@fir:create:ok::todo.debounce=\"$fir.replace()\"\n@fir:delete:error=\"replace\"", // No map, action name used directly
			wantErr:  false,
		},
		{
			name:     "Multiple Expressions (semicolon), no state (default state, default action)",
			input:    "create.debounce->todo;delete=>replace",
			expected: "@fir:create:ok::todo.debounce=\"$fir.replace()\"\n@fir:delete:ok=\"replace\"", // No map, action name used directly
			wantErr:  false,
		},
		{
			name: "Complex Mix (comma and semicolon) with Modifiers",
			// Removed .mod from replace
			input: "create:ok,delete:error->todo=>replace;update:pending.debounce->done=>archive",
			// Removed .mod from replace in expected output
			expected: "@fir:[create:ok,delete:error]::todo=\"replace\"\n@fir:update:pending::done.debounce=\"archive\"",
			wantErr:  false,
		},
		{
			name:     "Complex Mix (comma and semicolon), no state/action (default state, default action)",
			input:    "create,delete->todo;update.debounce->done",
			expected: "@fir:[create:ok,delete:ok]::todo=\"$fir.replace()\"\n@fir:update:ok::done.debounce=\"$fir.replace()\"",
			wantErr:  false,
		},

		// --- Action Map Tests ---
		{
			name:     "Single event, action target, action in map",
			input:    "submit=>doSubmit",
			actions:  map[string]string{"doSubmit": "replacedAction()"},
			expected: `@fir:submit:ok="replacedAction()"`, // Action replaced from map
			wantErr:  false,
		},
		{
			name:     "Single event, action target, action NOT in map",
			input:    "submit=>doSubmit",
			actions:  map[string]string{"anotherAction": "value"},
			expected: `@fir:submit:ok="doSubmit"`, // Action not found, original used
			wantErr:  false,
		},
		{
			name:     "Single event, action target, empty map provided",
			input:    "submit=>doSubmit",
			actions:  map[string]string{},
			expected: `@fir:submit:ok="doSubmit"`, // Empty map, original used
			wantErr:  false,
		},
		{
			name:     "Single event, action target, nil map provided",
			input:    "submit=>doSubmit",
			actions:  nil,
			expected: `@fir:submit:ok="doSubmit"`, // Nil map, original used (same as no map)
			wantErr:  false,
		},
		{
			name:     "Multiple events, action target, action in map",
			input:    "create,update=>myFunc",
			actions:  map[string]string{"myFunc": "handleMultiple()"},
			expected: `@fir:[create:ok,update:ok]="handleMultiple()"`, // Action replaced
			wantErr:  false,
		},
		{
			name:     "Multiple expressions, different actions, one in map",
			input:    "save=>saveData;load=>loadData",
			actions:  map[string]string{"saveData": "saveInternal()"},
			expected: "@fir:save:ok=\"saveInternal()\"\n@fir:load:ok=\"loadData\"", // Only saveData replaced
			wantErr:  false,
		},
		{
			name:     "Multiple expressions, different actions, both in map",
			input:    "save=>saveData;load=>loadData",
			actions:  map[string]string{"saveData": "saveInternal()", "loadData": "loadInternal()"},
			expected: "@fir:save:ok=\"saveInternal()\"\n@fir:load:ok=\"loadInternal()\"", // Both replaced
			wantErr:  false,
		},
		{
			name: "Complex mix, actions in map",
			// Removed .mod from replace
			input: "create:ok,delete:error->todo=>replace;update:pending.debounce->done=>archive",
			// Removed .mod from replace key in map and expected output
			actions:  map[string]string{"replace": "doReplace()", "archive": "doArchive()"},
			expected: "@fir:[create:ok,delete:error]::todo=\"doReplace()\"\n@fir:update:pending::done.debounce=\"doArchive()\"",
			wantErr:  false,
		},
		{
			name:     "Action name looks like default but is in map",
			input:    "click=>$fir.replace()",                            // This is now VALID input
			actions:  map[string]string{"someAction": "customReplace()"}, // Map is irrelevant for $fir actions
			expected: `@fir:click:ok="$fir.replace()"`,                   // Correct expected output
			wantErr:  false,                                              // Updated: This input is now valid
		},

		// --- Error Cases Inspired by lexer_test.go ---
		{
			name:    "Invalid State",
			input:   "create:invalid",
			wantErr: true,
		},
		{
			name:    "Invalid Target Name (numeric)",
			input:   "create->123",
			wantErr: true,
		},
		{
			name:    "Event with Only Modifier",
			input:   "",
			wantErr: true,
		},
		{
			name:    "Event with Only State",
			input:   ":ok",
			wantErr: true,
		},
		{
			name:    "Event with Only Target", // Assuming parser requires an event name
			input:   "->todo",
			wantErr: true,
		},
		{
			name:    "Event with Only Action", // Assuming parser requires an event name
			input:   "=>replace",
			wantErr: true,
		},
		{
			name:    "Event with Empty Modifier",
			input:   "create.",
			wantErr: true,
		},
		{
			name:     "Single Event with Modifier and Action Target Only (default state)",
			input:    "update.debounce=>doUpdate",
			expected: `@fir:update:ok.debounce="doUpdate"`,
			wantErr:  false,
		},
		{
			name:    "Empty Input",
			input:   "",
			wantErr: true, // Expect error for empty input
		},
		{
			name:     "Multiple Expressions with Trailing Semicolon",
			input:    "event1->tmpl1;event2=>act2;",
			expected: "@fir:event1:ok::tmpl1=\"$fir.replace()\"\n@fir:event2:ok=\"act2\"", // Expect trailing semicolon to be ignored
			wantErr:  false,                                                               // Updated: Should no longer error with the modified grammar
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			var err error
			// Call TranslateRenderExpression with the map if it's provided
			if tt.actions != nil {
				got, err = TranslateRenderExpression(tt.input, tt.actions)
			} else {
				got, err = TranslateRenderExpression(tt.input) // Call without map
			}

			if tt.wantErr {
				require.Error(t, err, "Expected an error but got none for input: %s", tt.input)
			} else {
				require.NoError(t, err, "Got unexpected error for input: %s", tt.input)
				require.Equal(t, tt.expected, got, "Mismatch for input: %s", tt.input)
			}
		})
	}
}

// TestTranslateEventExpression tests the TranslateEventExpression function
func TestTranslateEventExpression(t *testing.T) {
	// Define different action types to test
	actionTypes := map[string]string{
		"refresh":       "$fir.replace()",
		"remove":        "$fir.removeEl()",
		"remove-parent": "$fir.removeParentEl()",
		"append":        "$fir.appendEl()",
		"prepend":       "$fir.prependEl()",
	}

	tests := []struct {
		name                string
		input               string
		actionType          string   // Add actionType to the test case struct
		templateValue       string   // Add templateValue for testing
		additionalModifiers []string // Add optional additional modifiers
		expected            string
		wantErr             bool
	}{
		// --- Existing Test cases (adapted) ---
		{
			name:       "refresh: single event, no target",
			input:      "click",
			actionType: "refresh",
			expected:   `@fir:click:ok="$fir.replace()"`,
			wantErr:    false,
		},
		{
			name:          "refresh: single event, with templateValue",
			input:         "click",
			actionType:    "refresh",
			templateValue: "myTemplate",
			expected:      `@fir:click:ok::myTemplate="$fir.replace()"`,
			wantErr:       false,
		},
		{
			name:       "refresh: single event, template target (ignored)",
			input:      "submit->myform",
			actionType: "refresh",
			expected:   `@fir:submit:ok="$fir.replace()"`, // TranslateEventExpression ignores targets in input
			wantErr:    false,
		},
		{
			name:          "refresh: single event, template target (ignored), with templateValue",
			input:         "submit->myform",
			actionType:    "refresh",
			templateValue: "overrideTmpl",
			expected:      `@fir:submit:ok::overrideTmpl="$fir.replace()"`, // templateValue overrides ignored target
			wantErr:       false,
		},
		{
			name:       "refresh: multiple expressions (semicolon) with targets (ignored)",
			input:      "create:ok.debounce->todo;delete:error=>replace",
			actionType: "refresh",
			expected:   "@fir:create:ok.debounce=\"$fir.replace()\"\n@fir:delete:error=\"$fir.replace()\"",
			wantErr:    false,
		},
		{
			name:          "refresh: multiple expressions (semicolon), with templateValue",
			input:         "create:ok.debounce->todo;delete:error=>replace",
			actionType:    "refresh",
			templateValue: "commonTmpl",
			expected:      "@fir:create:ok::commonTmpl.debounce=\"$fir.replace()\"\n@fir:delete:error::commonTmpl=\"$fir.replace()\"",
			wantErr:       false,
		},
		{
			name:       "remove: multiple events (comma) with modifier",
			input:      "clear:ok,reset:done.mod",
			actionType: "remove",
			expected:   `@fir:[clear:ok,reset:done].mod="$fir.removeEl()"`,
			wantErr:    false,
		},
		{
			name:          "remove: multiple events (comma) with modifier and templateValue",
			input:         "clear:ok,reset:done.mod",
			actionType:    "remove",
			templateValue: "listTmpl",
			expected:      `@fir:[clear:ok,reset:done]::listTmpl.mod="$fir.removeEl()"`,
			wantErr:       false,
		},
		// --- New Tests for Additional Modifiers ---
		{
			name:                "refresh: single event, no input modifier, with additional modifiers",
			input:               "click",
			actionType:          "refresh",
			additionalModifiers: []string{"debounce", "throttle"},
			expected:            `@fir:click:ok.debounce.throttle="$fir.replace()"`, // Sorted
			wantErr:             false,
		},
		{
			name:                "refresh: single event, with input modifier, with additional modifiers (distinct)",
			input:               "click.self",
			actionType:          "refresh",
			additionalModifiers: []string{"debounce", "throttle"},
			expected:            `@fir:click:ok.debounce.self.throttle="$fir.replace()"`, // Sorted
			wantErr:             false,
		},
		{
			name:                "refresh: single event, with input modifier, with additional modifiers (overlap)",
			input:               "click.self.debounce", // Reverted to original input with multiple modifiers
			actionType:          "refresh",
			additionalModifiers: []string{"debounce", "throttle", "self"},                // Duplicates ignored
			expected:            `@fir:click:ok.debounce.self.throttle="$fir.replace()"`, // Sorted unique
			wantErr:             false,
		},
		{
			name:                "refresh: multiple events (comma), with input modifiers, with additional modifiers",
			input:               "click:ok.self,submit:error.once",
			actionType:          "refresh",
			additionalModifiers: []string{"debounce", "throttle", "once"},
			expected:            `@fir:[click:ok,submit:error].debounce.once.self.throttle="$fir.replace()"`, // Merged, sorted unique
			wantErr:             false,
		},
		{
			name:                "refresh: multiple expressions (semicolon), with additional modifiers",
			input:               "click.self;submit.once",
			actionType:          "refresh",
			additionalModifiers: []string{"debounce", "throttle"},
			expected:            "@fir:click:ok.debounce.self.throttle=\"$fir.replace()\"\n@fir:submit:ok.debounce.once.throttle=\"$fir.replace()\"", // Applied to both, sorted
			wantErr:             false,
		},
		{
			name:                "append: single event, with template, with additional modifiers",
			input:               "add.fast",
			actionType:          "append",
			templateValue:       "newItem",
			additionalModifiers: []string{"debounce"},
			expected:            `@fir:add:ok::newItem.debounce.fast="$fir.appendEl()"`, // Sorted
			wantErr:             false,
		},
		{
			name:                "remove-parent: multiple events (comma), with template, with additional modifiers",
			input:               "dismiss:ok.now,hide:done",
			actionType:          "remove-parent",
			templateValue:       "modalTmpl",
			additionalModifiers: []string{"delay", "now"},
			expected:            `@fir:[dismiss:ok,hide:done]::modalTmpl.delay.now="$fir.removeParentEl()"`, // Merged, sorted unique
			wantErr:             false,
		},
		{
			name:       "refresh: multiple input modifiers",
			input:      "click.debounce.throttle.self",
			actionType: "refresh",
			expected:   `@fir:click:ok.debounce.self.throttle="$fir.replace()"`, // Sorted
			wantErr:    false,
		},
		{
			name:                "refresh: multiple input modifiers merged with additional",
			input:               "click.debounce.self",
			actionType:          "refresh",
			additionalModifiers: []string{"throttle", "once"},
			expected:            `@fir:click:ok.debounce.once.self.throttle="$fir.replace()"`, // Merged and sorted
			wantErr:             false,
		},
		// --- Error Cases (actionType doesn't matter here) ---
		{
			name:       "error: Invalid State",
			input:      "create:invalid",
			actionType: "refresh", // actionType is irrelevant for parse errors
			wantErr:    true,
		},
		{
			name:       "error: Empty Input",
			input:      "",
			actionType: "refresh",
			wantErr:    true,
		},
		// ... other error cases ...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get the expected action value based on actionType
			actionValue, ok := actionTypes[tt.actionType]
			if !ok {
				t.Fatalf("Invalid actionType '%s' in test case '%s'", tt.actionType, tt.name)
			}

			var got string
			var err error

			// Call TranslateEventExpression with optional template and additional modifiers
			got, err = TranslateEventExpression(tt.input, actionValue, tt.templateValue, tt.additionalModifiers...)

			if tt.wantErr {
				require.Error(t, err, "Expected an error but got none for input: %s, actionType: %s, templateValue: '%s', additionalModifiers: %v", tt.input, tt.actionType, tt.templateValue, tt.additionalModifiers)
			} else {
				require.NoError(t, err, "Got unexpected error for input: %s, actionType: %s, templateValue: '%s', additionalModifiers: %v", tt.input, tt.actionType, tt.templateValue, tt.additionalModifiers)
				require.Equal(t, tt.expected, got, "Mismatch for input: %s, actionType: %s, templateValue: '%s', additionalModifiers: %v", tt.input, tt.actionType, tt.templateValue, tt.additionalModifiers)
			}
		})
	}
}
