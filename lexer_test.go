package fir

import (
	"fmt"
	"testing"

	"github.com/alecthomas/participle/v2"
)

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

func TestLexer_MultipleCases(t *testing.T) {
	parser := participle.MustBuild[Expressions](
		participle.Lexer(lexerRules),
		participle.Elide("Whitespace"), // Globally ignore whitespace
	)

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:  "Single Event with State",
			input: "create:ok->todo",
			expected: []string{
				"EventExpression: {Name:create State::ok}",
				"Template Target: todo",
			},
		},
		{
			name:  "Multiple Events with States and Actions",
			input: "create:ok,delete:error=>replace;update:pending->todo",
			expected: []string{
				"EventExpression: {Name:create State::ok}",
				"EventExpression: {Name:delete State::error}",
				"Action Target: replace",
				"EventExpression: {Name:update State::pending}",
				"Template Target: todo",
			},
		},
		{
			name:  "No State with Action",
			input: "create=>append",
			expected: []string{
				"EventExpression: {Name:create State:}",
				"Action Target: append",
			},
		},
		{
			name:  "Whitespace Ignored",
			input: "  create:ok  -> todo  , delete:error => replace  ",
			expected: []string{
				"EventExpression: {Name:create State::ok}",
				"Template Target: todo",
				"EventExpression: {Name:delete State::error}",
				"Action Target: replace",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := parser.ParseString("", tt.input)
			if err != nil {
				t.Fatalf("Failed to parse input: %v", err)
			}

			var output []string
			for _, expr := range parsed.Expressions {
				for _, binding := range expr.Bindings {
					for _, eventExpr := range binding.Eventexpressions {
						output = append(output, fmt.Sprintf("EventExpression: {Name:%s State:%s}", eventExpr.Name, eventExpr.State))
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
