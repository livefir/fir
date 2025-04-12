package fir

import (
	"fmt"
	"testing"
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
				"EventExpression: {Name:create State: Modifier:}",
			},
		},
		{
			name:  "Single Event without State",
			input: "create->todo",
			expected: []string{
				"EventExpression: {Name:create State: Modifier:}",
				"Template Target: todo",
			},
		},
		{
			name:  "Multiple Events without States",
			input: "create,delete=>replace",
			expected: []string{
				"EventExpression: {Name:create State: Modifier:}",
				"EventExpression: {Name:delete State: Modifier:}",
				"Action Target: replace",
			},
		},

		// Group 2: Events with States and Templates
		{
			name:  "Event with State and Template",
			input: "create:ok->todo",
			expected: []string{
				"EventExpression: {Name:create State::ok Modifier:}",
				"Template Target: todo",
			},
		},
		{
			name:  "Single Event with State",
			input: "create:ok->todo",
			expected: []string{
				"EventExpression: {Name:create State::ok Modifier:}",
				"Template Target: todo",
			},
		},

		// Group 3: Events with Modifiers
		{
			name:  "Event with Modifier",
			input: "create.nohtml",
			expected: []string{
				"EventExpression: {Name:create State: Modifier:.nohtml}",
			},
		},
		{
			name:  "Event with State and Modifier",
			input: "create:ok.nohtml",
			expected: []string{
				"EventExpression: {Name:create State::ok Modifier:.nohtml}",
			},
		},
		{
			name:  "Event with Modifier and Template Target",
			input: "create.nohtml->todo",
			expected: []string{
				"EventExpression: {Name:create State: Modifier:.nohtml}",
				"Template Target: todo",
			},
		},

		// Group 4: Complex Inputs
		{
			name:  "Complex Mixed Input",
			input: "create:ok->todo,delete:error=>replace;update:pending->done=>archive",
			expected: []string{
				"EventExpression: {Name:create State::ok Modifier:}",
				"Template Target: todo",
				"EventExpression: {Name:delete State::error Modifier:}",
				"Action Target: replace",
				"EventExpression: {Name:update State::pending Modifier:}",
				"Template Target: done",
				"Action Target: archive",
			},
		},
		{
			name:  "Multiple Events with Modifiers, States, and Targets",
			input: "create:ok.nohtml->todo,delete:error.nocache=>replace;update:pending->done=>archive",
			expected: []string{
				"EventExpression: {Name:create State::ok Modifier:.nohtml}",
				"Template Target: todo",
				"EventExpression: {Name:delete State::error Modifier:.nocache}",
				"Action Target: replace",
				"EventExpression: {Name:update State::pending Modifier:}",
				"Template Target: done",
				"Action Target: archive",
			},
		},

		// Group 5: Invalid Inputs
		{
			name:      "Invalid Modifier after Template",
			input:     "create:ok->todo.nohtml",
			expectErr: true, // Modifier after template is not allowed
		},
		{
			name:      "Event with Modifier and Invalid State",
			input:     "create:invalid.nohtml",
			expectErr: true, // Invalid state should trigger an error
		},
		{
			name:      "Event with Modifier and Invalid Target",
			input:     "create.nohtml->123",
			expectErr: true, // Invalid template target should trigger an error
		},
		{
			name:      "Event with Modifier and Empty Target",
			input:     "create.nohtml->",
			expectErr: true, // Empty template target should trigger an error
		},
		{
			name:      "Event with Modifier and Multiple Actions",
			input:     "create.nohtml=>replace=>append",
			expectErr: true, // Multiple actions are not allowed
		},
		{
			name:      "Event with Modifier and Invalid Characters in Modifier",
			input:     "create.no_html",
			expectErr: true, // Invalid characters in modifier should trigger an error
		},
		{
			name:      "Event with Modifier and Valid State but Invalid Action",
			input:     "create:ok.nohtml=>123",
			expectErr: true, // Invalid action target should trigger an error
		},
		{
			name:      "Event with Modifier and Multiple States",
			input:     "create:ok:error.nohtml",
			expectErr: true, // Multiple states are not allowed
		},
		{
			name:      "Event with Modifier and Special Characters in Target",
			input:     "create.nohtml->todo@123",
			expectErr: true, // Special characters in target should trigger an error
		},
		{
			name:  "Event with Modifier and Valid State but No Targets",
			input: "create:ok.nohtml",
			expected: []string{
				"EventExpression: {Name:create State::ok Modifier:.nohtml}",
			},
		},
		{
			name:  "Event with Modifier and Valid Action but No Template",
			input: "create.nohtml=>replace",
			expected: []string{
				"EventExpression: {Name:create State: Modifier:.nohtml}",
				"Action Target: replace",
			},
		},
		{
			name:      "Event with Modifier and Empty Input",
			input:     "",
			expectErr: true, // Empty input should trigger an error
		},
		{
			name:      "Event with Modifier and Invalid Characters in Modifier",
			input:     "create.no_html",
			expectErr: true, // Invalid characters in modifier should trigger an error
		},
		{
			name:      "Event with Modifier and Mixed Valid and Invalid Targets",
			input:     "create.nohtml->todo,delete:error=>123",
			expectErr: true, // Invalid action target should trigger an error
		},
		{
			name:  "Event with Modifier and Whitespace Between Tokens",
			input: "create .nohtml -> todo => replace",
			expected: []string{
				"EventExpression: {Name:create State: Modifier:.nohtml}",
				"Template Target: todo",
				"Action Target: replace",
			},
		},
		// Group 6: Whitespace Handling
		{
			name:  "Whitespace Ignored",
			input: "  create: ok  -> todo  , delete: error => replace  ",
			expected: []string{
				"EventExpression: {Name:create State::ok Modifier:}",
				"Template Target: todo",
				"EventExpression: {Name:delete State::error Modifier:}",
				"Action Target: replace",
			},
		},
		{
			name:  "Event with Modifier and Whitespace",
			input: "  create .nohtml  -> todo  ",
			expected: []string{
				"EventExpression: {Name:create State: Modifier:.nohtml}",
				"Template Target: todo",
			},
		},

		// Group 7: Modifiers with Complex Scenarios

		{
			name:      "Event with Modifier and Multiple States",
			input:     "create:ok:error.nohtml",
			expectErr: true, // Multiple states are not allowed
		},
		{
			name:      "Event with Modifier and Special Characters in Target",
			input:     "create.nohtml->todo@123",
			expectErr: true, // Special characters in target should trigger an error
		},
		{
			name:  "Event with Modifier and Valid State but No Targets",
			input: "create:ok.nohtml",
			expected: []string{
				"EventExpression: {Name:create State::ok Modifier:.nohtml}",
			},
		},
		{
			name:  "Event with Modifier and Valid Action but No Template",
			input: "create.nohtml=>replace",
			expected: []string{
				"EventExpression: {Name:create State: Modifier:.nohtml}",
				"Action Target: replace",
			},
		},
		{
			name:      "Event with Modifier and Empty Input",
			input:     "",
			expectErr: true, // Empty input should trigger an error
		},
		{
			name:      "Event with Modifier and Invalid Characters in Modifier",
			input:     "create.no_html",
			expectErr: true, // Invalid characters in modifier should trigger an error
		},
		{
			name:      "Event with Modifier and Mixed Valid and Invalid Targets",
			input:     "create.nohtml->todo,delete:error=>123",
			expectErr: true, // Invalid action target should trigger an error
		},
		{
			name:  "Event with Modifier and Whitespace Between Tokens",
			input: "create .nohtml -> todo => replace",
			expected: []string{
				"EventExpression: {Name:create State: Modifier:.nohtml}",
				"Template Target: todo",
				"Action Target: replace",
			},
		},
		// Group 8: Complex Mixed Inputs
		{
			name:  "Complex Input with Multiple Modifiers and Targets",
			input: "create:ok.nohtml->todo,delete:error.nocache=>replace;update:pending->done=>archive.final",
			expected: []string{
				"EventExpression: {Name:create State::ok Modifier:.nohtml}",
				"Template Target: todo",
				"EventExpression: {Name:delete State::error Modifier:.nocache}",
				"Action Target: replace",
				"EventExpression: {Name:update State::pending Modifier:}",
				"Template Target: done",
				"Action Target: archive.final",
			},
		},
		{
			name:      "Complex Input with Mixed Valid and Invalid Modifiers",
			input:     "create:ok.nohtml->todo,delete:error.nocache=>replace;update:done->archive.no_html",
			expectErr: true, // Invalid modifier should trigger an error
		},
		{
			name:      "Complex Input with Multiple Events and Empty Targets",
			input:     "create:ok.nohtml->,delete:error=>replace",
			expectErr: true, // Empty template target should trigger an error
		},
		{
			name:      "Complex Input with Multiple Events and Invalid Characters",
			input:     "create:ok.nohtml->todo,delete:error=>replace@123",
			expectErr: true, // Invalid characters in action target should trigger an error
		},
		{
			name:  "Complex Input with Multiple Events and Whitespace",
			input: "  create:ok .nohtml -> todo , delete : error => replace  ",
			expected: []string{
				"EventExpression: {Name:create State::ok Modifier:.nohtml}",
				"Template Target: todo",
				"EventExpression: {Name:delete State::error Modifier:}",
				"Action Target: replace",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := parseRenderExpression(parser, tt.input)
			if tt.expectErr {
				if err == nil {
					t.Fatalf("Expected an error but got none")
				}
				return // Test passes if an error is expected and received
			}
			if err != nil {
				t.Fatalf("Failed to parse input: %v", err)
			}

			var output []string
			for _, expr := range parsed.Expressions {
				for _, binding := range expr.Bindings {
					for _, eventExpr := range binding.Eventexpressions {
						output = append(output, fmt.Sprintf(
							"EventExpression: {Name:%s State:%s Modifier:%s}",
							eventExpr.Name,
							eventExpr.State,
							eventExpr.Modifier,
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
