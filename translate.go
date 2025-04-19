package fir

import (
	"fmt"
	"strings"
)

// Assume Expressions, Binding, Eventexpression, Target structs are defined correctly
// matching the grammar, including Target using "=>" for Action.

func TranslateRenderExpression(input string) (string, error) {
	parser, err := getRenderExpressionParser()
	if err != nil {
		return "", fmt.Errorf("failed to create parser: %w", err)
	}
	// parseRenderExpression uses the parser which MUST have the updated grammar
	parsed, err := parseRenderExpression(parser, input)
	if err != nil {
		return "", fmt.Errorf("failed to parse input: %w", err)
	}

	var result []string
	for _, expr := range parsed.Expressions {
		for _, binding := range expr.Bindings {
			var eventStrings []string
			for _, eventExpr := range binding.Eventexpressions {
				eventStrings = append(eventStrings, fmt.Sprintf("%s%s", eventExpr.Name, eventExpr.State))
			}

			var eventsPart string
			if len(eventStrings) > 1 {
				eventsPart = fmt.Sprintf("[%s]", strings.Join(eventStrings, ","))
			} else if len(eventStrings) == 1 {
				eventsPart = eventStrings[0]
			} else {
				continue // Should not happen with valid grammar
			}

			translation := fmt.Sprintf("@fir:%s", eventsPart)

			if binding.Target != nil {
				if binding.Target.Template != "" {
					translation += fmt.Sprintf("::%s", binding.Target.Template)
				}
				// This part correctly uses the Action field, regardless of how it was parsed ("->" or "=>")
				if binding.Target.Action != "" {
					translation += fmt.Sprintf("=\"%s\"", binding.Target.Action)
				}
			}
			result = append(result, translation)
		}
	}

	return strings.Join(result, "\n"), nil
}

// Assume parseRenderExpression exists and uses a parser with the correct grammar
// func parseRenderExpression(parser *participle.Parser[Expressions], input string) (*Expressions, error) { ... }

// Assume struct definitions match the required grammar (Target uses "=>")
// type Expressions struct { ... }
// type Expression struct { ... }
// type Binding struct { ... }
// type Eventexpression struct { ... }
// type Target struct {
// 	Template string `@Ident`
// 	Action   string `[ "=>" @Ident ]` // Grammar needs "=>" here
// }
