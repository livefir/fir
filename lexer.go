package fir

import (
	"fmt"
	"regexp" // Import regexp

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// Define the lexer rules
var lexerRules = lexer.MustSimple([]lexer.SimpleRule{
	// Updated pattern to allow one or more letters [a-zA-Z]+
	{Name: "FirAction", Pattern: `\$fir\.[a-zA-Z]+\(\)`}, // Matches $fir.Save(), $fir.Load(), etc.
	{Name: "Ident", Pattern: `[a-zA-Z_][a-zA-Z0-9_]*`},   // Matches event names like "create"
	{Name: "State", Pattern: `:(ok|error|pending|done)`}, // Matches states like ":ok"
	{Name: "Modifier", Pattern: `\.[a-zA-Z]+`},           // Matches modifiers like ".nohtml"
	{Name: "Arrow", Pattern: `->`},                       // Matches "->"
	{Name: "DoubleArrow", Pattern: `=>`},                 // Matches "=>"
	{Name: "Comma", Pattern: `,`},                        // Matches ","
	{Name: "Semicolon", Pattern: `;`},                    // Matches ";"
	{Name: "Whitespace", Pattern: `\s+`},                 // Ignore standalone whitespace
})

// Define the grammar structure
type Expressions struct {
	// Allow optional trailing semicolon by adding ';'? at the end
	Expressions []*Expression `parser:"@@ ( ';' @@ )* ';'? "`
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
	Template string `parser:"( \"->\" @Ident )?"` // Match template target for "->"
	// Action accepts Ident or FirAction. Modifier is removed. The entire Action part ("=> ...") is optional.
	Action string `parser:"( \"=>\" ( @Ident | @FirAction ) )?"`
}

// removeAllWhitespace is helper function to remove all whitespace from a string
// Using regexp is more concise
var whitespaceRegex = regexp.MustCompile(`\s+`)

func removeAllWhitespace(input string) string {
	return whitespaceRegex.ReplaceAllString(input, "")
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
