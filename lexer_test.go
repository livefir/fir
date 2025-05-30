package fir

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require" // Import testify/require
)

func TestLexer(t *testing.T) {
	parser, err := getRenderExpressionParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	tests := []struct {
		name      string
		input     string
		expected  []string
		expectErr bool
	}{
		// Group 1: Basic Event Parsing
		{
			name:  "Single Event without State, Template, or Action",
			input: "create",
			expected: []string{
				"EventExpression: {Name:create State: Modifiers:[]}",
			},
		},
		{
			name:  "Single Event with Template Target",
			input: "create->todo",
			expected: []string{
				"EventExpression: {Name:create State: Modifiers:[]}",
				"Template Target: todo",
			},
		},
		{
			name:  "Multiple Events without States",
			input: "create,delete=>replace",
			expected: []string{
				"EventExpression: {Name:create State: Modifiers:[]}",
				"EventExpression: {Name:delete State: Modifiers:[]}",
				"Action Target: replace",
			},
		},

		// Group 2: Events with States and Templates
		{
			name:  "Event with State and Template",
			input: "create:ok->todo",
			expected: []string{
				"EventExpression: {Name:create State::ok Modifiers:[]}",
				"Template Target: todo",
			},
		},
		{
			name:  "Event with State: pending", // New test
			input: "create:pending",
			expected: []string{
				"EventExpression: {Name:create State::pending Modifiers:[]}",
			},
		},
		{
			name:  "Event with State: done", // New test
			input: "create:done",
			expected: []string{
				"EventExpression: {Name:create State::done Modifiers:[]}",
			},
		},

		// Group 3: Events with Modifiers
		{
			name:  "Event with Modifier",
			input: "create.nohtml",
			expected: []string{
				"EventExpression: {Name:create State: Modifiers:[.nohtml]}",
			},
		},
		{
			name:  "Event with State and Modifier",
			input: "create:ok.nohtml",
			expected: []string{
				"EventExpression: {Name:create State::ok Modifiers:[.nohtml]}",
			},
		},
		{
			name:  "Event with Modifier and Template Target",
			input: "create.nohtml->todo",
			expected: []string{
				"EventExpression: {Name:create State: Modifiers:[.nohtml]}",
				"Template Target: todo",
			},
		},

		// Group 4: Complex Inputs
		{
			name:  "Complex Mixed Input",
			input: "create:ok->todo,delete:error=>replace;update:pending->done=>archive",
			expected: []string{
				"EventExpression: {Name:create State::ok Modifiers:[]}",
				"Template Target: todo",
				"EventExpression: {Name:delete State::error Modifiers:[]}",
				"Action Target: replace",
				"EventExpression: {Name:update State::pending Modifiers:[]}",
				"Template Target: done",
				"Action Target: archive",
			},
		},

		// Group 5: Invalid Inputs
		{
			name:      "Invalid Modifier after Template",
			input:     "create:ok->todo.nohtml",
			expectErr: true,
		},
		{
			name:      "Event with Modifier and Invalid State",
			input:     "create:invalid.nohtml",
			expectErr: true,
		},
		{
			name:      "Event with Modifier and Invalid Target",
			input:     "create.nohtml->123",
			expectErr: true,
		},
		{
			name:      "Event with Modifier and Empty Target",
			input:     "create.nohtml->",
			expectErr: true,
		},
		{
			name:      "Event with Modifier and Multiple Actions",
			input:     "create.nohtml=>replace=>append",
			expectErr: true,
		},
		{
			name:      "Event with Modifier and Invalid Characters in Modifier",
			input:     "create.no_html",
			expectErr: true,
		},
		{
			name:      "Event with Modifier and Multiple States",
			input:     "create:ok:error.nohtml",
			expectErr: true,
		},

		// Group 7: Modifiers with Complex Scenarios
		{
			name:      "Event with Modifier and Special Characters in Target",
			input:     "create.nohtml->todo@123",
			expectErr: true,
		},
		{
			name:  "Event with Modifier and Valid State but No Targets",
			input: "create:ok.nohtml",
			expected: []string{
				"EventExpression: {Name:create State::ok Modifiers:[.nohtml]}",
			},
		},

		// Group 8: Complex Mixed Inputs
		{
			name:  "Complex Input with Multiple Modifiers and Targets",
			input: "create:ok.nohtml->todo,delete:error.nocache=>replace;update:pending->done=>archive",
			expected: []string{
				"EventExpression: {Name:create State::ok Modifiers:[.nohtml]}",
				"Template Target: todo",
				"EventExpression: {Name:delete State::error Modifiers:[.nocache]}",
				"Action Target: replace",
				"EventExpression: {Name:update State::pending Modifiers:[]}",
				"Template Target: done",
				"Action Target: archive",
			},
		},

		// Group 9: Edge Cases
		{
			name:      "Event with Only Modifier",
			input:     ".nohtml",
			expectErr: true,
		},
		{ // New test case for trailing semicolon
			name:  "Multiple Expressions with Trailing Semicolon",
			input: "create:ok->todo;delete:error=>replace;",
			expected: []string{
				"EventExpression: {Name:create State::ok Modifiers:[]}",
				"Template Target: todo",
				"EventExpression: {Name:delete State::error Modifiers:[]}",
				"Action Target: replace",
			},
			expectErr: false, // Should parse successfully now
		},
		{
			name:      "Event with Only State",
			input:     ":ok",
			expectErr: true,
		},
		{
			name:      "Event with Only Target",
			input:     "->todo",
			expectErr: true,
		},
		{
			name:      "Event with Only Action",
			input:     "=>replace",
			expectErr: true,
		},
		{
			name:      "Event with Empty Modifier",
			input:     "create.",
			expectErr: true,
		},
		{
			name:      "Event with Invalid Characters in Name",
			input:     "cre@te:ok.nohtml",
			expectErr: true,
		},
		{
			name:  "Event with Modifier and Action Target Only", // New test
			input: "create.mod=>doAction",
			expected: []string{
				"EventExpression: {Name:create State: Modifiers:[.mod]}",
				"Action Target: doAction",
			},
		},
		{
			name:  "Event with Modifier and Template Target",
			input: "create.mod->doTemplate",
			expected: []string{
				"EventExpression: {Name:create State: Modifiers:[.mod]}",
				"Template Target: doTemplate",
			},
		},
		{
			name:  "Single Event with Action Target Only", // New test
			input: "create=>doAction",
			expected: []string{
				"EventExpression: {Name:create State: Modifiers:[]}",
				"Action Target: doAction",
			},
		},
		{
			name:  "Multiple Events without States",
			input: "create,delete=>replace",
			expected: []string{
				"EventExpression: {Name:create State: Modifiers:[]}",
				"EventExpression: {Name:delete State: Modifiers:[]}",
				"Action Target: replace",
			},
		},
		{
			name:  "Comma-Separated Bindings without Target", // New test
			input: "event1:ok, event2.mod",
			expected: []string{
				"EventExpression: {Name:event1 State::ok Modifiers:[]}",
				"EventExpression: {Name:event2 State: Modifiers:[.mod]}",
			},
		},
		{
			name:  "Identifiers with Numbers and Underscores", // New test
			input: "event_1->templateA=>action123_B",
			expected: []string{
				"EventExpression: {Name:event_1 State: Modifiers:[]}",
				"Template Target: templateA",
				"Action Target: action123_B",
			},
		},
		{
			name:  "Identifiers with Hyphens (like mutation-observer)", // New test for hyphenated identifiers
			input: "mutation-observer:ok->template-name=>action-handler",
			expected: []string{
				"EventExpression: {Name:mutation-observer State::ok Modifiers:[]}",
				"Template Target: template-name",
				"Action Target: action-handler",
			},
		},

		// Group 10: Fir Action Rule Tests
		{
			name:  "Single Event with Fir Action",
			input: "create => $fir.X()",
			expected: []string{
				"EventExpression: {Name:create State: Modifiers:[]}",
				"Action Target: $fir.X()",
			},
		},
		{
			name:  "Event with Template and Fir Action",
			input: "update -> myTemplate => $fir.Y()",
			expected: []string{
				"EventExpression: {Name:update State: Modifiers:[]}",
				"Template Target: myTemplate",
				"Action Target: $fir.Y()",
			},
		},
		{
			name:  "Multiple Bindings with Fir Actions",
			input: "load:ok -> data, save => $fir.Z(); submit => $fir.A()",
			expected: []string{
				"EventExpression: {Name:load State::ok Modifiers:[]}",
				"Template Target: data",
				"EventExpression: {Name:save State: Modifiers:[]}",
				"Action Target: $fir.Z()",
				"EventExpression: {Name:submit State: Modifiers:[]}",
				"Action Target: $fir.A()",
			},
		},
		{
			name:      "Invalid: Modifier after Fir Action",
			input:     "create => $fir.X().mod", // Modifier after action is no longer allowed
			expectErr: true,
		},
		{
			name:      "Invalid: Modifier after Standard Action",
			input:     "update => myAction.mod", // Modifier after action is no longer allowed
			expectErr: true,
		},
		{
			name:      "Invalid: Fir Action with incorrect format",
			input:     "create => $fir.1()", // Digit instead of letter
			expectErr: true,
		},
		{
			name:      "Invalid: Fir Action with incorrect format 2",
			input:     "create => $fir.X", // Missing parentheses
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := parseRenderExpression(parser, tt.input)
			if tt.expectErr {
				if err == nil {
					t.Fatalf("Expected an error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("Failed to parse input: %v", err)
			}

			var output []string
			for _, expr := range parsed.Expressions {
				for _, binding := range expr.Bindings {
					for _, eventExpr := range binding.Eventexpressions {
						// Format modifiers for output
						modsStr := "[]"
						if len(eventExpr.Modifiers) > 0 {
							modsStr = fmt.Sprintf("[%s]", strings.Join(eventExpr.Modifiers, " "))
						}
						output = append(output, fmt.Sprintf(
							"EventExpression: {Name:%s State:%s Modifiers:%s}", // Changed Modifier to Modifiers
							eventExpr.Name,
							eventExpr.State,
							modsStr, // Use formatted modifiers string
						))
					}
					if binding.Target != nil {
						if binding.Target.Template != "" {
							output = append(output, fmt.Sprintf("Template Target: %s", binding.Target.Template))
						}
						if binding.Target.Action != "" {
							output = append(output, fmt.Sprintf("Action Target: %s", binding.Target.Action))
						}
					}
				}
			}

			if len(output) != len(tt.expected) {
				t.Fatalf("Expected %d outputs, got %d", len(tt.expected), len(output))
			}

			for i, expected := range tt.expected {
				if output[i] != expected {
					t.Errorf("Expected output[%d] to be %q, got %q", i, expected, output[i])
				}
			}
		})
	}
}

func TestParseActionExpression(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedMap map[string][]string
		wantErr     bool
		errContains string // Optional: check for specific error message content
	}{
		// --- Valid Cases ---
		{
			name:        "Valid: Action with multiple parameters (brackets)",
			input:       "x-fir-toggleClass:[loading,visible-state]",
			expectedMap: map[string][]string{"toggleClass": {"loading", "visible-state"}},
			wantErr:     false,
		},
		{
			name:        "Valid: Action with single parameter (brackets)",
			input:       "x-fir-addClass:[active]",
			expectedMap: map[string][]string{"addClass": {"active"}},
			wantErr:     false,
		},
		{
			name:        "Valid: Action with single parameter (no brackets)", // New test case
			input:       "x-fir-addClass:active",
			expectedMap: map[string][]string{"addClass": {"active"}},
			wantErr:     false,
		},
		{
			name:        "Valid: Action with no parameters (no colon/brackets)",
			input:       "x-fir-removeClass",
			expectedMap: map[string][]string{"removeClass": {}}, // Expect empty slice
			wantErr:     false,
		},
		{
			name:        "Valid: Action with empty parameters list (brackets)",
			input:       "x-fir-setValue:[]",
			expectedMap: map[string][]string{"setValue": {}}, // Expect empty slice
			wantErr:     false,
		},
		{
			name:        "Valid: Whitespace around expression (brackets)",
			input:       "  x-fir-addClass:[  active  ]  ",
			expectedMap: map[string][]string{"addClass": {"active"}},
			wantErr:     false,
		},
		{
			name:        "Valid: Whitespace around expression (no brackets)", // New test case
			input:       "  x-fir-addClass : active  ",
			expectedMap: map[string][]string{"addClass": {"active"}},
			wantErr:     false,
		},
		{
			name:        "Valid: Whitespace within parameters (brackets)",
			input:       "x-fir-toggleClass : [ loading , visible ]",
			expectedMap: map[string][]string{"toggleClass": {"loading", "visible"}},
			wantErr:     false,
		},
		{
			name:        "Valid: Parameters with underscores and hyphens (brackets)",
			input:       "x-fir-addClass:[is_active, data-state-loading]",
			expectedMap: map[string][]string{"addClass": {"is_active", "data-state-loading"}},
			wantErr:     false,
		},
		{
			name:        "Valid: Parameter with underscores and hyphens (no brackets)", // New test case
			input:       "x-fir-addClass:is_active-state",
			expectedMap: map[string][]string{"addClass": {"is_active-state"}},
			wantErr:     false,
		},
		{
			name:        "Valid: Long action name",
			input:       "x-fir-long_action-name:param1",
			expectedMap: map[string][]string{"long_action-name": {"param1"}},
			wantErr:     false,
		},
		{
			name:        "Valid: Long parameter name (no brackets)",
			input:       "x-fir-setValue:this_is_a_very_long_parameter_name-123",
			expectedMap: map[string][]string{"setValue": {"this_is_a_very_long_parameter_name-123"}},
			wantErr:     false,
		},
		{
			name:        "Valid: Unknown action name (parser perspective)",
			input:       "x-fir-unknownAction:param",
			expectedMap: map[string][]string{"unknownAction": {"param"}},
			wantErr:     false,
		},

		// --- Error Cases ---
		{
			name:        "Error: Invalid prefix",
			input:       "xfir-toggleClass:loading", // Still invalid prefix
			wantErr:     true,
			errContains: "invalid prefix for action key",
		},
		{
			name:        "Error: Empty input string",
			input:       "",
			wantErr:     true,
			errContains: "action key cannot be empty",
		},
		{
			name:        "Error: Malformed parameters - missing closing bracket",
			input:       "x-fir-toggleClass:[loading",
			wantErr:     true,
			errContains: "unexpected token \"<EOF>\"", // Parser error
		},
		{
			name:        "Error: Malformed parameters - missing opening bracket",
			input:       "x-fir-toggleClass:loading]", // Still error, needs IDENT or '[' after ':'
			wantErr:     true,
			errContains: "unexpected token \"]\"", // Corrected: Parser sees 'loading' then unexpected ']'
		},
		{
			name:        "Error: Malformed parameters - invalid character in bracketed list",
			input:       "x-fir-toggleClass:[loa@ding]",
			wantErr:     true,
			errContains: "lexer: invalid input text", // Lexer error
		},
		{
			name:        "Error: Malformed parameters - invalid character in single param", // New test case
			input:       "x-fir-toggleClass:loa@ding",
			wantErr:     true,
			errContains: "lexer: invalid input text", // Lexer error
		},
		{
			name:        "Error: Missing action name",
			input:       "x-fir-:[param]",
			wantErr:     true,
			errContains: "unexpected token \":\"", // Parser error
		},
		{
			name:        "Error: Only whitespace input",
			input:       "   ",
			wantErr:     true,
			errContains: "action key cannot be empty",
		},
		{
			name:        "Error: Only prefix",
			input:       "x-fir-",
			wantErr:     true,
			errContains: "unexpected token \"<EOF>\"",
		},
		{
			name:        "Error: Prefix and action, missing params after colon",
			input:       "x-fir-toggleClass:", // Colon but nothing after
			wantErr:     true,
			errContains: "unexpected token \"<EOF>\"", // Expects IDENT or '[' after ':'
		},
		{
			name:        "Error: Prefix and action, missing params after colon with whitespace", // New test
			input:       "x-fir-toggleClass:   ",                                                // Colon followed by whitespace
			wantErr:     true,
			errContains: "unexpected token \"<EOF>\"", // Expects IDENT or '[' after ':'
		},
		{
			name:        "Error: Prefix, action, colon, only opening bracket",
			input:       "x-fir-toggleClass:[",
			wantErr:     true,
			errContains: "unexpected token \"<EOF>\"",
		},
		{
			name:        "Error: Prefix, action, colon, only closing bracket",
			input:       "x-fir-toggleClass:]",
			wantErr:     true,
			errContains: "unexpected token \"]\"", // Parser error, expects IDENT or '[' after ':'
		},
		{
			name:    "Error: Trailing comma in parameters",
			input:   "x-fir-toggleClass:[loading,]",
			wantErr: true,
			// Participle error might vary slightly, but it should fail on the ']' after ','
			errContains: "unexpected token \",\"", // Corrected: Parser expects IDENT after comma, gets ']'
		},
		{
			name:        "Error: Leading comma in parameters",
			input:       "x-fir-toggleClass:[,loading]",
			wantErr:     true,
			errContains: "unexpected token \",\"",
		},
		{
			name:        "Error: Empty parameter name between commas",
			input:       "x-fir-toggleClass:[loading,,visible]",
			wantErr:     true,
			errContains: "unexpected token \",\"",
		},
		{
			name:    "Error: Multiple parameters without brackets", // New test case
			input:   "x-fir-toggleClass:loading,visible",
			wantErr: true,
			// The parser expects EOF after the first IDENT ('loading') if no brackets are used
			errContains: "unexpected token \",\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actionName, params, err := parseActionExpression(tt.input)

			if tt.wantErr {
				require.Error(t, err, "Expected an error but got none")
				if tt.errContains != "" {
					require.ErrorContains(t, err, tt.errContains, "Error message mismatch")
				}
			} else {
				require.NoError(t, err, "Got unexpected error")
				gotMap := make(map[string][]string)
				if actionName != "" {
					// Ensure params is never nil, should be empty slice if no params found
					if params == nil {
						params = []string{}
					}
					gotMap[actionName] = params
				}
				require.Equal(t, tt.expectedMap, gotMap, "Parsed map mismatch")
			}
		})
	}
}

func TestParseActionExpressionMutationObserver(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedMap map[string][]string
		wantErr     bool
	}{
		{
			name:        "Valid: x-fir-mutation-observer with no parameters",
			input:       "x-fir-mutation-observer",
			expectedMap: map[string][]string{"mutation-observer": {}},
			wantErr:     false,
		},
		{
			name:        "Valid: x-fir-mutation-observer with single parameter",
			input:       "x-fir-mutation-observer:childList",
			expectedMap: map[string][]string{"mutation-observer": {"childList"}},
			wantErr:     false,
		},
		{
			name:        "Valid: x-fir-mutation-observer with multiple parameters",
			input:       "x-fir-mutation-observer:[childList,attributes,subtree]",
			expectedMap: map[string][]string{"mutation-observer": {"childList", "attributes", "subtree"}},
			wantErr:     false,
		},
		{
			name:        "Valid: x-fir-mutation-observer with hyphenated parameters",
			input:       "x-fir-mutation-observer:[child-list,attribute-old-value]",
			expectedMap: map[string][]string{"mutation-observer": {"child-list", "attribute-old-value"}},
			wantErr:     false,
		},
		{
			name:        "Valid: x-fir-mutation-observer with whitespace",
			input:       "  x-fir-mutation-observer : [ childList , attributes ]  ",
			expectedMap: map[string][]string{"mutation-observer": {"childList", "attributes"}},
			wantErr:     false,
		},
		{
			name:        "Valid: x-fir-mutation-observer with complex modifier names",
			input:       "x-fir-mutation-observer:[child-list,attribute-old-value,character-data-old-value]",
			expectedMap: map[string][]string{"mutation-observer": {"child-list", "attribute-old-value", "character-data-old-value"}},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actionName, params, err := parseActionExpression(tt.input)

			if tt.wantErr {
				require.Error(t, err, "Expected an error but got none")
			} else {
				require.NoError(t, err, "Got unexpected error")
				gotMap := make(map[string][]string)
				if actionName != "" {
					// Ensure params is never nil, should be empty slice if no params found
					if params == nil {
						params = []string{}
					}
					gotMap[actionName] = params
				}
				require.Equal(t, tt.expectedMap, gotMap, "Parsed map mismatch")
			}
		})
	}
}
