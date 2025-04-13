package fir

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/v2"
)

func TranslateRenderExpression(parser *participle.Parser[Expressions], input string) (string, error) {
	parsed, err := parseRenderExpression(parser, input)
	if err != nil {
		return "", fmt.Errorf("failed to parse input: %w", err)
	}

	var result []string
	for _, expr := range parsed.Expressions {
		for _, binding := range expr.Bindings {
			for _, eventExpr := range binding.Eventexpressions {
				// Construct the base event expression
				translation := fmt.Sprintf("@fir:%s%s", eventExpr.Name, eventExpr.State)
				if binding.Target != nil {
					if binding.Target.Template != "" {
						translation += fmt.Sprintf("::%s", binding.Target.Template)
					}
					if binding.Target.Action != "" {
						translation += fmt.Sprintf("=\"%s\"", binding.Target.Action)
					}
				}
				result = append(result, translation)
			}
		}
	}

	return strings.Join(result, "\n"), nil
}
