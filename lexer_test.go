package fir

import (
	"fmt"
	"testing"

	"github.com/alecthomas/participle/v2"
)

func TestLexer_MultipleCases(t *testing.T) {
	parser, err := getRenderExpressionParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	tests := []struct {
		name      string
		input     string
		expected  []string
		expectErr bool // New field to indicate if an error is expected
	}{
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
		{
			name:  "Event with State and Template",
			input: "create:ok->todo",
			expected: []string{
				"EventExpression: {Name:create State::ok Modifier:}",
				"Template Target: todo",
			},
		},
		{
			name:      "Empty Input",
			input:     "",
			expected:  nil,  // No expected output
			expectErr: true, // Expect an error
		},
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
			name:  "Single Event with State",
			input: "create:ok->todo",
			expected: []string{
				"EventExpression: {Name:create State::ok Modifier:}",
				"Template Target: todo",
			},
		},
		{
			name:  "Multiple Events with States and Actions",
			input: "create:ok,delete:error=>replace;update:pending->todo",
			expected: []string{
				"EventExpression: {Name:create State::ok Modifier:}",
				"EventExpression: {Name:delete State::error Modifier:}",
				"Action Target: replace",
				"EventExpression: {Name:update State::pending Modifier:}",
				"Template Target: todo",
			},
		},
		{
			name:  "No State with Action",
			input: "create=>append",
			expected: []string{
				"EventExpression: {Name:create State: Modifier:}",
				"Action Target: append",
			},
		},
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
			name:  "Valid Input Without Whitespace Between Event and State",
			input: "create:ok->todo",
			expected: []string{
				"EventExpression: {Name:create State::ok Modifier:}",
				"Template Target: todo",
			},
			expectErr: false, // No error expected
		},
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
			name:  "Template Target without Modifier",
			input: "create:ok->todo",
			expected: []string{
				"EventExpression: {Name:create State::ok Modifier:}",
				"Template Target: todo",
			},
		},
		{
			name:      "Invalid Modifier after Template",
			input:     "create:ok->todo.nohtml",
			expectErr: true, // Modifier after template is not allowed
		},
		{
			name:  "Action Target with Modifier",
			input: "create:ok=>fir.replace",
			expected: []string{
				"EventExpression: {Name:create State::ok Modifier:}",
				"Action Target: fir.replace",
			},
		},
		{
			name:  "Multiple Events with Modifiers and States",
			input: "create:ok.nohtml,delete:error.nocache=>replace",
			expected: []string{
				"EventExpression: {Name:create State::ok Modifier:.nohtml}",
				"EventExpression: {Name:delete State::error Modifier:.nocache}",
				"Action Target: replace",
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
		{
			name:  "Event with State, Modifier, and Action Target",
			input: "create:ok.nohtml=>replace",
			expected: []string{
				"EventExpression: {Name:create State::ok Modifier:.nohtml}",
				"Action Target: replace",
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
		{
			name:  "Event with Modifier and Multiple Targets",
			input: "create.nohtml->todo=>replace",
			expected: []string{
				"EventExpression: {Name:create State: Modifier:.nohtml}",
				"Template Target: todo",
				"Action Target: replace",
			},
		},
		{
			name:      "Invalid Modifier after Template Target",
			input:     "create:ok->todo.nohtml",
			expectErr: true, // Modifier after template is not allowed
		},
		{
			name:  "Multiple Events with Mixed Modifiers and Targets",
			input: "create.nohtml,delete:error.nocache->todo=>replace",
			expected: []string{
				"EventExpression: {Name:create State: Modifier:.nohtml}",
				"EventExpression: {Name:delete State::error Modifier:.nocache}",
				"Template Target: todo",
				"Action Target: replace",
			},
		},
		{
			name:  "Event with State, Modifier, and No Targets",
			input: "create:ok.nohtml",
			expected: []string{
				"EventExpression: {Name:create State::ok Modifier:.nohtml}",
			},
		},
		{
			name:  "Event with Modifier and Action Target Only",
			input: "create.nohtml=>replace",
			expected: []string{
				"EventExpression: {Name:create State: Modifier:.nohtml}",
				"Action Target: replace",
			},
		},
		{
			name:  "Complex Input with Multiple Modifiers and Targets",
			input: "create:ok.nohtml->todo,delete:error.nocache=>replace;update:done->archive=>finalize",
			expected: []string{
				"EventExpression: {Name:create State::ok Modifier:.nohtml}",
				"Template Target: todo",
				"EventExpression: {Name:delete State::error Modifier:.nocache}",
				"Action Target: replace",
				"EventExpression: {Name:update State::done Modifier:}",
				"Template Target: archive",
				"Action Target: finalize",
			},
		},
		{
			name:  "Event with Modifier and No State",
			input: "create.nohtml",
			expected: []string{
				"EventExpression: {Name:create State: Modifier:.nohtml}",
			},
		},
		{
			name:  "Event with State, Modifier, and Multiple Targets",
			input: "create:ok.nohtml->todo=>replace",
			expected: []string{
				"EventExpression: {Name:create State::ok Modifier:.nohtml}",
				"Template Target: todo",
				"Action Target: replace",
			},
		},
		{
			name:  "Multiple Events with Modifiers and Mixed Targets",
			input: "create.nohtml->todo,delete:error.nocache=>replace",
			expected: []string{
				"EventExpression: {Name:create State: Modifier:.nohtml}",
				"Template Target: todo",
				"EventExpression: {Name:delete State::error Modifier:.nocache}",
				"Action Target: replace",
			},
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
			name:  "Event with Modifier and State but No Targets",
			input: "create:ok.nohtml",
			expected: []string{
				"EventExpression: {Name:create State::ok Modifier:.nohtml}",
			},
		},
		{
			name:      "Complex Input with Multiple Modifiers and Invalid Target",
			input:     "create:ok.nohtml->todo,delete:error.nocache=>123",
			expectErr: true, // Invalid action target should trigger an error
		},
		{
			name:  "Event with Modifier and Whitespace",
			input: "  create .nohtml  -> todo  ",
			expected: []string{
				"EventExpression: {Name:create State: Modifier:.nohtml}",
				"Template Target: todo",
			},
		},
		{
			name:      "Event with Modifier and Special Characters in Target",
			input:     "create.nohtml->todo@123",
			expectErr: true, // Special characters in target should trigger an error
		},
		{
			name:      "Event with Modifier and Multiple States",
			input:     "create:ok:error.nohtml",
			expectErr: true, // Multiple states are not allowed
		},
		{
			name:      "Event with Modifier and Mixed Valid and Invalid Targets",
			input:     "create.nohtml->todo,delete:error=>123",
			expectErr: true, // Invalid action target should trigger an error
		},
		{
			name:  "Event with Modifier and Valid State but No Action",
			input: "create:ok.nohtml->todo",
			expected: []string{
				"EventExpression: {Name:create State::ok Modifier:.nohtml}",
				"Template Target: todo",
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
			name:      "Event with Modifier and Valid State but Invalid Action",
			input:     "create:ok.nohtml=>123",
			expectErr: true, // Invalid action target should trigger an error
		},
		{
			name:  "Complex Input with Multiple Events and Modifiers",
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

func TestLexer(t *testing.T) {
	parser := participle.MustBuild[Expressions](
		participle.Lexer(lexerRules),
	)

	example := "create:ok,delete:error=>replace;create:ok->todo=>append"
	parsed, err := parser.ParseString("", example)
	if err != nil {
		t.Fatalf("Failed to parse input: %v", err)
	}

	for _, expr := range parsed.Expressions {
		for _, binding := range expr.Bindings {
			for _, eventExpr := range binding.Eventexpressions {
				fmt.Printf("    EventExpression: %+v\n", eventExpr)
			}
			if binding.Target != nil {
				if binding.Target.Template != "" {
					fmt.Printf("    Template Target: %s\n", binding.Target.Template)
				}
				if binding.Target.Action != "" {
					fmt.Printf("    Action Target: %s\n", binding.Target.Action)
				}
			}
		}
	}
}
