package fir

import (
	"fmt"
	"sort"
	"strings"
)

// Assume Expressions, Binding, Eventexpression, Target structs are defined correctly
// matching the grammar, including Target using "=>" for Action.
// Assume Eventexpression has a Modifier field, e.g., Modifier string `@("." @Ident)?`

// Helper function to get state or default to ":ok"
func getStateOrDefault(state string) string {
	if state == "" {
		return ":ok" // Default state
	}
	return state
}

func TranslateRenderExpression(input string) (string, error) {
	parser, err := getRenderExpressionParser()
	if err != nil {
		return "", fmt.Errorf("failed to create parser: %w", err)
	}
	parsed, err := parseRenderExpression(parser, input)
	if err != nil {
		return "", fmt.Errorf("failed to parse input: %w", err)
	}

	var expressionResults []string
	// Process each Expression (group separated by ';')
	for _, expr := range parsed.Expressions {
		var allEventStrings []string
		allModifierSet := make(map[string]struct{})
		var effectiveTemplate string
		var effectiveAction string
		var hasExplicitTarget bool // Track if any binding in the group had a target

		// Collect events, modifiers, and determine target from ALL bindings within the expression (group separated by ',')
		for _, binding := range expr.Bindings {
			for _, eventExpr := range binding.Eventexpressions {
				// Use helper to apply default state ":ok" if missing
				eventState := getStateOrDefault(eventExpr.State)
				allEventStrings = append(allEventStrings, fmt.Sprintf("%s%s", eventExpr.Name, eventState))
				if eventExpr.Modifier != "" {
					allModifierSet[strings.TrimPrefix(eventExpr.Modifier, ".")] = struct{}{}
				}
			}
			// "Last target wins" rule for comma-separated bindings within an expression
			if binding.Target != nil {
				hasExplicitTarget = true // Mark that a target was found in this group
				if binding.Target.Template != "" {
					effectiveTemplate = binding.Target.Template
				}
				if binding.Target.Action != "" {
					effectiveAction = binding.Target.Action
				}
			}
		}

		if len(allEventStrings) == 0 {
			continue // Skip if an expression somehow has no events
		}

		// Sort modifiers
		var allModifiers []string
		for mod := range allModifierSet {
			allModifiers = append(allModifiers, mod)
		}
		sort.Strings(allModifiers)
		modifierString := ""
		if len(allModifiers) > 0 {
			modifierString = "." + strings.Join(allModifiers, ".")
		}

		// Format events part (use brackets if multiple events collected OR if multiple bindings were grouped by comma)
		var eventsPart string
		if len(allEventStrings) > 1 || len(expr.Bindings) > 1 {
			eventsPart = fmt.Sprintf("[%s]", strings.Join(allEventStrings, ","))
		} else {
			eventsPart = allEventStrings[0] // Single event overall in this expression from a single binding
		}

		// Construct the translation string for the expression
		translation := fmt.Sprintf("@fir:%s", eventsPart)

		if !hasExplicitTarget { // Check if any target was specified in the group
			// No target: append modifiers directly to events part
			translation += modifierString
		} else {
			// Target was specified (-> or =>)
			if effectiveTemplate != "" {
				translation += fmt.Sprintf("::%s", effectiveTemplate)
				// Template present: append modifiers after template
				translation += modifierString
			} else {
				// No template, but action might be present or defaulted: append modifiers before action
				translation += modifierString
			}

			// Append action: either the specified one or the default if target existed but action was missing
			if effectiveAction != "" {
				translation += fmt.Sprintf("=\"%s\"", effectiveAction)
			} else {
				// Apply default action only if a target arrow (-> or =>) was present
				translation += `="$fir.replace()"` // Default action
			}
		}
		expressionResults = append(expressionResults, translation)
	}

	// Join results from different expressions (separated by ';') with newline
	return strings.Join(expressionResults, "\n"), nil
}

// Assume parseRenderExpression, struct definitions, and getRenderExpressionParser are correctly defined.
// func parseRenderExpression(parser *participle.Parser[Expressions], input string) (*Expressions, error) { ... }
// type Expressions struct { Expressions []*Expression `@(@@ (";" @@)*)?` }
// type Expression struct { Bindings []*Binding `@@ ("," @@)*` }
// type Binding struct { Eventexpressions []*Eventexpression `@@ ("," @@)*` Target *Target `@@?` }
// type Eventexpression struct { Name string `@Ident`; State string `(":" @("ok" | "error" | "pending" | "done" | "cancel"))?`; Modifier string `("." @Ident)?` }
// type Target struct { Template string `("->" @Ident)?`; Action string `("=>" @Ident)?` }
// func getRenderExpressionParser() (*participle.Parser[Expressions], error) { ... }
