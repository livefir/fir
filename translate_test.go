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
			expected: `@fir:[create:ok,update:ok]`,
			wantErr:  false,
		},
		// --- Inspired by lexer_test.go ---
		{
			name:     "Event with Modifier (modifier ignored in output)",
			input:    "create.nohtml",
			expected: "@fir:create",
			wantErr:  false,
		},
		{
			name:     "Event with State and Modifier (modifier ignored)",
			input:    "create:ok.nohtml",
			expected: "@fir:create:ok",
			wantErr:  false,
		},
		{
			name:     "Event with Modifier and Template Target (modifier ignored)",
			input:    "create.nohtml->todo",
			expected: "@fir:create::todo",
			wantErr:  false,
		},
		{
			name:     "Event with Action Target with Modifier",
			input:    "create=>doAction.mod", // Modifier on action is part of action name
			expected: `@fir:create="doAction.mod"`,
			wantErr:  false,
		},
		{
			name:     "Event with Template and Action Target with Modifier",
			input:    "create->view=>doAction.mod",
			expected: `@fir:create::view="doAction.mod"`,
			wantErr:  false,
		},
		{
			name:     "Multiple Bindings (comma) - generates multiple lines",
			input:    "create:ok->todo,delete:error=>replace",
			expected: "@fir:create:ok::todo\n@fir:delete:error=\"replace\"",
			wantErr:  false,
		},
		{
			name:     "Multiple Expressions (semicolon) - generates multiple lines",
			input:    "create:ok->todo;delete:error=>replace",
			expected: "@fir:create:ok::todo\n@fir:delete:error=\"replace\"",
			wantErr:  false,
		},
		{
			name:     "Complex Mix (comma and semicolon) - generates multiple lines",
			input:    "create:ok.nohtml->todo,delete:error=>replace.mod;update:pending->done=>archive",
			expected: "@fir:create:ok::todo\n@fir:delete:error=\"replace.mod\"\n@fir:update:pending::done=\"archive\"", // Note: .nohtml ignored, .mod included
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
