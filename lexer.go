package fir

import (
	"github.com/alecthomas/participle/v2/lexer"
)

// Define the lexer rules
var lexerRules = lexer.MustSimple([]lexer.SimpleRule{
	{Name: "Ident", Pattern: `[a-zA-Z_][a-zA-Z0-9_]*`},   // Matches event names like "create"
	{Name: "State", Pattern: `:(ok|error|pending|done)`}, // Matches states like ":ok" without capturing trailing whitespace
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
	Name  string `parser:"@Ident"`
	State string `parser:"(@State)?"`
}

type Target struct {
	Template string `parser:"( \"->\" @Ident )?"` // Match template target for "->"
	Action   string `parser:"( \"=>\" @Ident )?"` // Match action target for "=>"
}
