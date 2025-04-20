package fir

import (
	"fmt"
	"sort"
	"strings"
)

// Assume Expressions, Binding, Eventexpression, Target structs are defined correctly
// matching the grammar, including Target using "=>" for Action.
// Assume Eventexpression has a Modifier field, e.g., Modifier string `@("." @Ident)?`

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
			modifierSet := make(map[string]struct{}) // Use a map to collect unique modifiers
			for _, eventExpr := range binding.Eventexpressions {
				eventStrings = append(eventStrings, fmt.Sprintf("%s%s", eventExpr.Name, eventExpr.State))
				// Assume Eventexpression struct has a Modifier field
				if eventExpr.Modifier != "" {
					// Remove leading dot if present from parsing, add to set
					modifierSet[strings.TrimPrefix(eventExpr.Modifier, ".")] = struct{}{}
				}
			}

			// Sort modifiers alphabetically for consistent output
			var modifiers []string
			for mod := range modifierSet {
				modifiers = append(modifiers, mod)
			}
			sort.Strings(modifiers)
			modifierString := ""
			if len(modifiers) > 0 {
				modifierString = "." + strings.Join(modifiers, ".")
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

			// Apply modifiers based on target presence
			if binding.Target == nil {
				// No target: append modifiers directly to events part
				translation += modifierString
			} else {
				if binding.Target.Template != "" {
					translation += fmt.Sprintf("::%s", binding.Target.Template)
					// Template present: append modifiers after template
					translation += modifierString
				} else {
					// No template, but action might be present: append modifiers before action
					translation += modifierString
				}

				if binding.Target.Action != "" {
					// Action present: append action assignment
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
// type Eventexpression struct {
// 	Name     string `@Ident`
// 	State    string `[@(":" @("ok" | "error" | "pending" | "done" | "cancel"))]`
// 	Modifier string `@("." @Ident)?` // Assumed field for modifier
// }
// type Target struct {
// 	Template string `@("->" @Ident)?` // Adjusted template parsing
// 	Action   string `[ "=>" @Ident ]` // Action parsing
// }

// Assume getRenderExpressionParser is defined and returns a parser configured
// with the correct grammar including the Modifier field in Eventexpression.
// func getRenderExpressionParser() (*participle.Parser[Expressions], error) { ... }
