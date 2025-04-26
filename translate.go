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

// TranslateRenderExpression translates a fir expression string into its canonical form.
// It accepts an optional actions map. If provided, action targets (e.g., "myfunc" in "event=>myfunc")
// found as keys in the map will be replaced by their corresponding map values in the output.
func TranslateRenderExpression(input string, actions ...map[string]string) (string, error) {
	var actionMap map[string]string
	if len(actions) > 0 {
		actionMap = actions[0] // Use the first map if provided
	}

	// lower all keys in the action map for case-insensitive matching
	for key, value := range actionMap {
		lowerKey := strings.ToLower(key)
		actionMap[lowerKey] = value
		if lowerKey != key {
			delete(actionMap, key) // Remove the original key
		}
	}

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

		// Determine the final action string, considering the actions map
		finalActionValue := ""
		applyDefaultAction := false

		if hasExplicitTarget {
			if effectiveAction != "" {
				// Check if the action exists in the provided map
				if actionMap != nil {
					if replacement, ok := actionMap[strings.ToLower(effectiveAction)]; ok {
						finalActionValue = replacement // Use replacement from map
					} else {
						finalActionValue = effectiveAction // Use original action name
					}
				} else {
					finalActionValue = effectiveAction // Use original action name if no map provided
				}
			} else {
				// Target exists, but action is missing -> apply default
				applyDefaultAction = true
			}
		} else {
			// No explicit target (-> or =>) was present in the expression group
			// Apply default action "$fir.replace()"
			applyDefaultAction = true
		}

		// Append template and modifiers
		if effectiveTemplate != "" {
			translation += fmt.Sprintf("::%s", effectiveTemplate)
			// Template present: append modifiers after template
			translation += modifierString
		} else {
			// No template: append modifiers before action (or at the end if no action)
			translation += modifierString
		}

		// Append the action part
		if finalActionValue != "" {
			// Use the determined action value (either original or from map)
			translation += fmt.Sprintf("=\"%s\"", finalActionValue)
		} else if applyDefaultAction {
			// Apply the default action if needed
			translation += `="$fir.replace()"`
		}

		expressionResults = append(expressionResults, translation)
	}

	// Join results from different expressions (separated by ';') with newline
	return strings.Join(expressionResults, "\n"), nil
}

// TranslateEventExpression translates a render expression string focusing only on event expressions
// into canonical @fir event binding attributes, ignoring any template or action targets.
func TranslateEventExpression(input string, actionValue string) (string, error) {
	parser, err := getRenderExpressionParser()
	if err != nil {
		return "", fmt.Errorf("error creating parser: %w", err)
	}

	parsed, err := parseRenderExpression(parser, input)
	if err != nil {
		return "", fmt.Errorf("error parsing render expression: %w", err)
	}

	var results []string

	for _, expr := range parsed.Expressions {
		for _, binding := range expr.Bindings {
			var eventParts []string
			var modifiers []string
			modifierSet := make(map[string]struct{}) // To store unique modifiers

			for _, eventExpr := range binding.Eventexpressions {
				state := eventExpr.State
				if state == "" {
					state = ":ok" // Default state
				}
				eventPart := eventExpr.Name + state
				eventParts = append(eventParts, eventPart)

				// Collect unique modifiers
				if eventExpr.Modifier != "" {
					if _, exists := modifierSet[eventExpr.Modifier]; !exists {
						modifiers = append(modifiers, eventExpr.Modifier)
						modifierSet[eventExpr.Modifier] = struct{}{}
					}
				}
			}

			// Format event part
			eventStr := ""
			if len(eventParts) == 1 {
				eventStr = eventParts[0]
			} else {
				eventStr = "[" + strings.Join(eventParts, ",") + "]"
			}

			// Append modifiers
			modifierStr := strings.Join(modifiers, "")

			// Construct the final attribute string for this binding, ignoring any parsed target
			attribute := fmt.Sprintf("@fir:%s%s=\"%s\"", eventStr, modifierStr, actionValue)
			results = append(results, attribute)
		}
	}

	return strings.Join(results, "\n"), nil
}

// Assume parseRenderExpression, struct definitions, and getRenderExpressionParser are correctly defined.
// func parseRenderExpression(parser *participle.Parser[Expressions], input string) (*Expressions, error) { ... }
// type Expressions struct { Expressions []*Expression `@(@@ (";" @@)*)? ';' ?` } // Example grammar
// type Expression struct { Bindings []*Binding `@@ ("," @@)*` }
// type Binding struct { Eventexpressions []*Eventexpression `@@+` Target *Target `@@?` }
// type Eventexpression struct { Name string `@Ident`; State string `(":" @("ok" | "error" | "pending" | "done"))?`; Modifier string `("." @Ident)?` }
// type Target struct { Template string `("->" @Ident)?`; Action string `("=>" @Ident)?` }
// func getRenderExpressionParser() (*participle.Parser[Expressions], error) { ... }
