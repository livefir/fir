package fir

import (
	"fmt"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// Define the lexer rules
var lexerRules = lexer.MustSimple([]lexer.SimpleRule{
	{Name: "Ident", Pattern: `[a-zA-Z_][a-zA-Z0-9_]*`},   // Matches event names like "create"
	{Name: "State", Pattern: `:(ok|error|pending|done)`}, // Matches states like ":ok" without capturing trailing whitespace
	{Name: "Modifier", Pattern: `\.[a-zA-Z]+`},           // Matches modifiers like ".nohtml" (alphabetic only)
	{Name: "Arrow", Pattern: `->`},                       // Matches "->" without trailing whitespace
	{Name: "DoubleArrow", Pattern: `=>`},                 // Matches "=>" without trailing whitespace
	{Name: "Comma", Pattern: `,`},                        // Matches ","
	{Name: "Semicolon", Pattern: `;`},                    // Matches ";"
	{Name: "Whitespace", Pattern: `\s+`},                 // Ignore standalone whitespace
})

// Define the grammar structure
type Expressions struct {
	Expressions []*Expression `parser:"@@ ( ';' @@ )*"`
}

type Expression struct {
	Bindings []*Binding `parser:"@@ ( ',' @@ )*"`
}

type Binding struct {
	Eventexpressions []*EventExpression `parser:"@@ ( ',' @@ )*"`
	Target           *Target            `parser:"@@"` // Parse both template and action targets
}

type EventExpression struct {
	Name     string `parser:"@Ident"`
	State    string `parser:"(@State)?"`
	Modifier string `parser:"(@Modifier)?"` // Optional modifier after event or state
}

type Target struct {
	Template string `parser:"( \"->\" @Ident )?"`             // Match template target for "->"
	Action   string `parser:"( \"=>\" @Ident (@Modifier)?)?"` // Match action target with optional modifier
}

// removeAllWhitespace is helper function to remove all whitespace from a string
func removeAllWhitespace(input string) string {
	result := ""
	for _, r := range input {
		if r != ' ' && r != '\t' && r != '\n' && r != '\r' {
			result += string(r)
		}
	}
	return result
}

// getRenderExpressionParser parser function to parse the input string
func getRenderExpressionParser() (*participle.Parser[Expressions], error) {
	parser, err := participle.Build[Expressions](
		participle.Lexer(lexerRules),
		participle.Elide("Whitespace"), // Ignore standalone whitespace globally
	)
	return parser, err
}

func parseRenderExpression(parser *participle.Parser[Expressions], input string) (*Expressions, error) {
	if input == "" {
		return nil, fmt.Errorf("render expression cannot be empty")
	}
	input = removeAllWhitespace(input)
	parsed, err := parser.ParseString("", input)
	if err != nil {

		return nil, err
	}

	return parsed, nil
}
