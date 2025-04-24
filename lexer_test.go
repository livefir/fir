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
			name:  "Single Event with Template Target",
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
			name:  "Event with State: pending", // New test
			input: "create:pending",
			expected: []string{
				"EventExpression: {Name:create State::pending Modifier:}",
			},
		},
		{
			name:  "Event with State: done", // New test
			input: "create:done",
			expected: []string{
				"EventExpression: {Name:create State::done Modifier:}",
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
				"EventExpression: {Name:create State::ok Modifier:.nohtml}",
			},
		},

		// Group 8: Complex Mixed Inputs
		{
			name:  "Complex Input with Multiple Modifiers and Targets",
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
				"EventExpression: {Name:create State::ok Modifier:}",
				"Template Target: todo",
				"EventExpression: {Name:delete State::error Modifier:}",
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
				"EventExpression: {Name:create State: Modifier:.mod}",
				"Action Target: doAction",
			},
		},
		{
			name:  "Event with Modifier and Template Target",
			input: "create.mod->doTemplate",
			expected: []string{
				"EventExpression: {Name:create State: Modifier:.mod}",
				"Template Target: doTemplate",
			},
		},
		{
			name:  "Single Event with Action Target Only", // New test
			input: "create=>doAction",
			expected: []string{
				"EventExpression: {Name:create State: Modifier:}",
				"Action Target: doAction",
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
			name:  "Comma-Separated Bindings without Target", // New test
			input: "event1:ok, event2.mod",
			expected: []string{
				"EventExpression: {Name:event1 State::ok Modifier:}",
				"EventExpression: {Name:event2 State: Modifier:.mod}",
			},
		},
		{
			name:  "Identifiers with Numbers and Underscores", // New test
			input: "event_1->templateA=>action123_B",
			expected: []string{
				"EventExpression: {Name:event_1 State: Modifier:}",
				"Template Target: templateA",
				"Action Target: action123_B",
			},
		},

		// Group 10: Fir Action Rule Tests
		{
			name:  "Single Event with Fir Action",
			input: "create => $fir.X()",
			expected: []string{
				"EventExpression: {Name:create State: Modifier:}",
				"Action Target: $fir.X()",
			},
		},
		{
			name:  "Event with Template and Fir Action",
			input: "update -> myTemplate => $fir.Y()",
			expected: []string{
				"EventExpression: {Name:update State: Modifier:}",
				"Template Target: myTemplate",
				"Action Target: $fir.Y()",
			},
		},
		{
			name:  "Multiple Bindings with Fir Actions",
			input: "load:ok -> data, save => $fir.Z(); submit => $fir.A()",
			expected: []string{
				"EventExpression: {Name:load State::ok Modifier:}",
				"Template Target: data",
				"EventExpression: {Name:save State: Modifier:}",
				"Action Target: $fir.Z()",
				"EventExpression: {Name:submit State: Modifier:}",
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
