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
		// --- Existing Basic Cases ---
		{
			name:     "single event, no target",
			input:    "click",
			expected: "@fir:click",
			wantErr:  false,
		},
		{
			name:     "single event with state, no target",
			input:    "click:ok",
			expected: "@fir:click:ok",
			wantErr:  false,
		},
		{
			name:     "single event, template target",
			input:    "submit->myform",
			expected: "@fir:submit::myform",
			wantErr:  false,
		},
		{
			name:     "single event, action target", // Added for completeness
			input:    "submit=>doSubmit",
			expected: `@fir:submit="doSubmit"`,
			wantErr:  false,
		},
		{
			name:     "single event, template and action target (arrow)",
			input:    "submit->myform=>doSubmit", // Using => for action
			expected: `@fir:submit::myform="doSubmit"`,
			wantErr:  false,
		},
		{
			name:     "multiple events, template and action target (arrow)",
			input:    "create:ok,update:ok->mytemplate=>myfunction", // Using => for action
			expected: `@fir:[create:ok,update:ok]::mytemplate="myfunction"`,
			wantErr:  false,
		},
		{
			name:     "multiple events, no state, template and action target (arrow)",
			input:    "create,update->mytemplate=>myfunction", // Using => for action
			expected: `@fir:[create,update]::mytemplate="myfunction"`,
			wantErr:  false,
		},
		{
			name:     "multiple events, no target",
			input:    "create:ok,update:ok",
			expected: `@fir:[create:ok,update:ok]`, // Brackets for multiple events
			wantErr:  false,
		},
		// --- Modifiers ---
		{
			name:     "Single Event with Modifier, No Target",
			input:    "create.nohtml",
			expected: "@fir:create.nohtml", // No brackets for single event
			wantErr:  false,
		},
		{
			name:     "Single Event with State and Modifier, No Target",
			input:    "create:ok.nohtml",
			expected: "@fir:create:ok.nohtml", // No brackets for single event
			wantErr:  false,
		},
		{
			name:     "Single Event with Modifier and Template Target",
			input:    "create.nohtml->todo",
			expected: "@fir:create::todo.nohtml", // No brackets for single event
			wantErr:  false,
		},
		{
			name:     "Single Event with Modifier and Action Target",
			input:    "create.mod=>doAction",
			expected: `@fir:create.mod="doAction"`, // No brackets for single event
			wantErr:  false,
		},
		{
			name:     "Single Event with Modifier, Template, and Action Target",
			input:    "create.mod->view=>doAction",
			expected: `@fir:create::view.mod="doAction"`, // No brackets for single event
			wantErr:  false,
		},
		{
			name:     "Grouped Events with Different Modifiers, Template and Action Target", // User prompt example
			input:    "create:ok.debounce,update:error.nohtml->template=>myaction",
			expected: `@fir:[create:ok,update:error]::template.debounce.nohtml="myaction"`, // Brackets for grouped events
			wantErr:  false,
		},
		{
			name:     "Grouped Events with Different Modifiers, No Target",
			input:    "create:ok.debounce,update:error.nohtml",
			expected: `@fir:[create:ok,update:error].debounce.nohtml`, // Brackets for grouped events
			wantErr:  false,
		},
		{
			name:     "Grouped Events with Different Modifiers, Template Target Only",
			input:    "create:ok.debounce,update:error.nohtml->template",
			expected: `@fir:[create:ok,update:error]::template.debounce.nohtml`, // Brackets for grouped events
			wantErr:  false,
		},
		{
			name:     "Grouped Events with Different Modifiers, Action Target Only",
			input:    "create:ok.debounce,update:error.nohtml=>myaction",
			expected: `@fir:[create:ok,update:error].debounce.nohtml="myaction"`, // Brackets for grouped events
			wantErr:  false,
		},
		// --- Multiple Bindings (Separate Lines) ---
		{
			name:     "Multiple Bindings (comma) with Modifiers - generates multiple lines",
			input:    "create:ok.debounce->todo,delete:error.nohtml=>replace",               // Note: comma separates BINDINGS here
			expected: "@fir:create:ok::todo.debounce\n@fir:delete:error.nohtml=\"replace\"", // No brackets for single events in separate bindings
			wantErr:  false,
		},
		{
			name:     "Multiple Expressions (semicolon) with Modifiers - generates multiple lines",
			input:    "create:ok.debounce->todo;delete:error.nohtml=>replace",
			expected: "@fir:create:ok::todo.debounce\n@fir:delete:error.nohtml=\"replace\"", // No brackets for single events in separate bindings
			wantErr:  false,
		},
		{
			name: "Complex Mix (comma and semicolon) with Modifiers - generates multiple lines",
			// Binding 1: create:ok.nohtml->todo
			// Binding 2: delete:error=>replace.mod (action name contains .mod)
			// Binding 3: update:pending.debounce->done=>archive
			input:    "create:ok.nohtml->todo,delete:error=>replace.mod;update:pending.debounce->done=>archive",
			expected: "@fir:create:ok::todo.nohtml\n@fir:delete:error=\"replace.mod\"\n@fir:update:pending::done.debounce=\"archive\"", // No brackets for single events in separate bindings
			wantErr:  false,
		},
		// --- Error Cases Inspired by lexer_test.go ---
		{
			name:    "Invalid State",
			input:   "create:invalid.nohtml", // Parser should reject 'invalid' state
			wantErr: true,
		},
		{
			name:    "Invalid Target Name (numeric)",
			input:   "create.nohtml->123", // Parser should reject '123' as Ident
			wantErr: true,
		},
		{
			name:    "Event with Only Modifier",
			input:   ".nohtml", // Parser expects Ident first
			wantErr: true,
		},
		{
			name:    "Event with Only State",
			input:   ":ok", // Parser expects Ident first
			wantErr: true,
		},
		{
			name:    "Event with Only Target",
			input:   "->todo", // Parser expects event expression first
			wantErr: true,
		},
		{
			name:    "Event with Only Action",
			input:   "=>replace", // Parser expects event expression first
			wantErr: true,
		},
		{
			name:    "Event with Empty Modifier",
			input:   "create.", // Parser expects modifier name after '.'
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
				require.Equal(t, tt.expected, got, "Mismatch for input: %s", tt.input)
			}
		})
	}
}
