package fir

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
		expected string
		wantErr  bool
	}{
		// --- Existing Basic Cases (Updated for Defaults) ---
		{
			name:     "single event, no target (default state, default action)", // Updated name
			input:    "click",
			expected: `@fir:click:ok="$fir.replace()"`, // Default state :ok and default action added
			wantErr:  false,
		},
		{
			name:     "single event with state, no target (default action)", // Updated name
			input:    "click:ok",
			expected: `@fir:click:ok="$fir.replace()"`, // Default action added
			wantErr:  false,
		},
		{
			name:     "single event, template target (default state, default action)",
			input:    "submit->myform",
			expected: `@fir:submit:ok::myform="$fir.replace()"`, // Default state :ok and action added
			wantErr:  false,
		},
		{
			name:     "single event, action target (default state)",
			input:    "submit=>doSubmit",
			expected: `@fir:submit:ok="doSubmit"`, // Default state :ok added
			wantErr:  false,
		},
		{
			name:     "single event, template and action target (default state)",
			input:    "submit->myform=>doSubmit",
			expected: `@fir:submit:ok::myform="doSubmit"`, // Default state :ok added
			wantErr:  false,
		},
		{
			name:     "multiple events, template and action target",
			input:    "create:ok,update:ok->mytemplate=>myfunction",
			expected: `@fir:[create:ok,update:ok]::mytemplate="myfunction"`,
			wantErr:  false,
		},
		{
			name:     "multiple events, no state, template and action target (default state)",
			input:    "create,update->mytemplate=>myfunction",
			expected: `@fir:[create:ok,update:ok]::mytemplate="myfunction"`, // Default state :ok added for both
			wantErr:  false,
		},
		{
			name:     "multiple events, no target (default action)", // Updated name
			input:    "create:ok,update:ok",
			expected: `@fir:[create:ok,update:ok]="$fir.replace()"`, // Default action added
			wantErr:  false,
		},
		{
			name:     "multiple events, no state, no target (default state, default action)", // Updated name
			input:    "create,update",
			expected: `@fir:[create:ok,update:ok]="$fir.replace()"`, // Default state :ok and default action added
			wantErr:  false,
		},
		// --- Modifiers (Updated for Defaults) ---
		{
			name:     "Single Event with Modifier, No Target (default state, default action)", // Updated name
			input:    "create.nohtml",
			expected: `@fir:create:ok.nohtml="$fir.replace()"`, // Default state :ok and default action added
			wantErr:  false,
		},
		{
			name:     "Single Event with State and Modifier, No Target (default action)", // Updated name
			input:    "create:ok.nohtml",
			expected: `@fir:create:ok.nohtml="$fir.replace()"`, // Default action added
			wantErr:  false,
		},
		{
			name:     "Single Event with Modifier and Template Target (default state, default action)",
			input:    "create.nohtml->todo",
			expected: `@fir:create:ok::todo.nohtml="$fir.replace()"`, // Default state :ok and action added
			wantErr:  false,
		},
		{
			name:     "Single Event with Modifier and Action Target (default state)",
			input:    "create.mod=>doAction",
			expected: `@fir:create:ok.mod="doAction"`, // Default state :ok added
			wantErr:  false,
		},
		{
			name:     "Single Event with Modifier, Template, and Action Target (default state)",
			input:    "create.mod->view=>doAction",
			expected: `@fir:create:ok::view.mod="doAction"`, // Default state :ok added
			wantErr:  false,
		},
		{
			name:     "Grouped Events with Different Modifiers, Template and Action Target",
			input:    "create:ok.debounce,update:error.nohtml->template=>myaction",
			expected: `@fir:[create:ok,update:error]::template.debounce.nohtml="myaction"`,
			wantErr:  false,
		},
		{
			name:     "Grouped Events with Different Modifiers, No Target (default action)", // Updated name
			input:    "create:ok.debounce,update:error.nohtml",
			expected: `@fir:[create:ok,update:error].debounce.nohtml="$fir.replace()"`, // Default action added
			wantErr:  false,
		},
		{
			name:     "Grouped Events with Different Modifiers, Template Target Only (default action)",
			input:    "create:ok.debounce,update:error.nohtml->template",
			expected: `@fir:[create:ok,update:error]::template.debounce.nohtml="$fir.replace()"`, // Default action added
			wantErr:  false,
		},
		{
			name:     "Grouped Events with Different Modifiers, Action Target Only",
			input:    "create:ok.debounce,update:error.nohtml=>myaction",
			expected: `@fir:[create:ok,update:error].debounce.nohtml="myaction"`,
			wantErr:  false,
		},
		{
			name:     "Grouped Events, no state, Template Target Only (default state, default action)",
			input:    "create.debounce,update.nohtml->template",
			expected: `@fir:[create:ok,update:ok]::template.debounce.nohtml="$fir.replace()"`, // Default state :ok and action added
			wantErr:  false,
		},
		// --- Grouped Bindings (Comma) / Separate Expressions (Semicolon) (Updated for Defaults) ---
		{
			name:     "Grouped Bindings (comma) with Modifiers and Targets - generates SINGLE line",
			input:    "create:ok.debounce,delete:error.nohtml->todo=>replace",
			expected: `@fir:[create:ok,delete:error]::todo.debounce.nohtml="replace"`,
			wantErr:  false,
		},
		{
			name:     "Grouped Bindings (comma) No Target - generates SINGLE line (default action)", // Updated name
			input:    "create:ok.debounce,delete:error.nohtml",
			expected: `@fir:[create:ok,delete:error].debounce.nohtml="$fir.replace()"`, // Default action added
			wantErr:  false,
		},
		{
			name:     "Grouped Bindings (comma), no state, No Target (default state, default action)", // Updated name
			input:    "create.debounce,delete.nohtml",
			expected: `@fir:[create:ok,delete:ok].debounce.nohtml="$fir.replace()"`, // Default state :ok and default action added
			wantErr:  false,
		},
		{
			name:     "Grouped Bindings (comma), Template Only (default action)",
			input:    "create:ok.debounce,delete:error.nohtml->todo",
			expected: `@fir:[create:ok,delete:error]::todo.debounce.nohtml="$fir.replace()"`, // Default action added
			wantErr:  false,
		},
		{
			name:     "Grouped Bindings (comma), Action Only",
			input:    "create:ok.debounce,delete:error.nohtml=>replace",
			expected: `@fir:[create:ok,delete:error].debounce.nohtml="replace"`,
			wantErr:  false,
		},
		{
			name:     "Multiple Expressions (semicolon) with Modifiers (default action)",
			input:    "create:ok.debounce->todo;delete:error.nohtml=>replace",
			expected: "@fir:create:ok::todo.debounce=\"$fir.replace()\"\n@fir:delete:error.nohtml=\"replace\"", // Default action added to first part
			wantErr:  false,
		},
		{
			name:     "Multiple Expressions (semicolon), no state (default state, default action)",
			input:    "create.debounce->todo;delete.nohtml=>replace",
			expected: "@fir:create:ok::todo.debounce=\"$fir.replace()\"\n@fir:delete:ok.nohtml=\"replace\"", // Default state :ok and action added
			wantErr:  false,
		},
		{
			name:     "Complex Mix (comma and semicolon) with Modifiers",
			input:    "create:ok.nohtml,delete:error->todo=>replace.mod;update:pending.debounce->done=>archive",
			expected: "@fir:[create:ok,delete:error]::todo.nohtml=\"replace.mod\"\n@fir:update:pending::done.debounce=\"archive\"",
			wantErr:  false,
		},
		{
			name:     "Complex Mix (comma and semicolon), no state/action (default state, default action)",
			input:    "create.nohtml,delete->todo;update.debounce->done",
			expected: "@fir:[create:ok,delete:ok]::todo.nohtml=\"$fir.replace()\"\n@fir:update:ok::done.debounce=\"$fir.replace()\"", // Defaults added
			wantErr:  false,
		},
		// --- Error Cases Inspired by lexer_test.go ---
		{
			name:    "Invalid State",
			input:   "create:invalid.nohtml",
			wantErr: true,
		},
		{
			name:    "Invalid Target Name (numeric)",
			input:   "create.nohtml->123",
			wantErr: true,
		},
		{
			name:    "Event with Only Modifier",
			input:   ".nohtml",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Assuming TranslateRenderExpression internally gets/uses the correct parser
			got, err := TranslateRenderExpression(tt.input)
			if tt.wantErr {
				require.Error(t, err, "Expected an error but got none for input: %s", tt.input)
			} else {
				require.NoError(t, err, "Got unexpected error for input: %s", tt.input)
				// Use require.Equal for better diffs
				require.Equal(t, tt.expected, got, "Mismatch for input: %s", tt.input)
			}
		})
	}
}
