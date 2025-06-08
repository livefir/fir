package fir

import (
	"fmt" // Import regexp
	"strings"
	"sync" // Import sync for caching

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// Define the lexer rules for parsing the render expression
var renderExpressionlexerRules = lexer.MustSimple([]lexer.SimpleRule{
	// Whitespace - handled by participle.Elide("Whitespace") in the parser
	{Name: "Whitespace", Pattern: `\s+`},

	// Arrow operators
	{Name: "Arrow", Pattern: `\->`},
	{Name: "DoubleArrow", Pattern: `=>`},

	// Other tokens
	{Name: "FirAction", Pattern: `\$fir\.[a-zA-Z]+\(\)`},
	{Name: "Comma", Pattern: `,`},
	{Name: "Semicolon", Pattern: `;`},
	{Name: "State", Pattern: `:(ok|error|pending|done)`},
	{Name: "Modifier", Pattern: `\.[a-zA-Z]+`},

	// Identifiers - covers both regular and hyphenated identifiers
	{Name: "Ident", Pattern: `[a-zA-Z_][a-zA-Z0-9_\-]*`},
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
	Name      string   `parser:"@Ident"`
	State     string   `parser:"(@State)?"`
	Modifiers []string `parser:"(@Modifier)*"` // Changed to slice and '*' for zero or more
}

type Target struct {
	Template string `parser:"( \"->\" @Ident )?"`
	Action   string `parser:"( \"=>\" ( @Ident | @FirAction ) )?"`
}

// getRenderExpressionParser parser function to parse the input string
func getRenderExpressionParser() (*participle.Parser[Expressions], error) {
	parser, err := participle.Build[Expressions](
		participle.Lexer(renderExpressionlexerRules),
		participle.Elide("Whitespace"), // Ignore standalone whitespace globally
	)
	return parser, err
}

// Add a preprocessing step before parsing
func preProcessExpression(input string) string {
	// Add spaces around the operators to ensure they're recognized as separate tokens
	input = strings.ReplaceAll(input, "->", " -> ")
	input = strings.ReplaceAll(input, "=>", " => ")

	// Normalize multiple spaces to single space
	for strings.Contains(input, "  ") {
		input = strings.ReplaceAll(input, "  ", " ")
	}

	return strings.TrimSpace(input)
}

// Call this function before passing to the parser
func parseRenderExpression(parser *participle.Parser[Expressions], input string) (*Expressions, error) {
	if input == "" {
		return nil, fmt.Errorf("render expression cannot be empty")
	}

	// Preprocess the input to add spaces around arrows
	input = preProcessExpression(input)

	parsed, err := parser.ParseString("", input)
	if err != nil {
		return nil, err
	}

	return parsed, nil
}

// --- Action Expression Lexer and Parser ---

// Define the lexer rules for parsing the action expression
var actionExpressionLexerRules = lexer.MustSimple([]lexer.SimpleRule{
	{Name: "Prefix", Pattern: `x-fir-`},
	{Name: "Ident", Pattern: `[a-zA-Z0-9_-]+`}, // Action names and parameters
	{Name: "LBracket", Pattern: `\[`},
	{Name: "RBracket", Pattern: `\]`},
	{Name: "Comma", Pattern: `,`},
	{Name: "Colon", Pattern: `:`},
	{Name: "Whitespace", Pattern: `\s+`},
})

// Define the grammar structure for action expressions
type ActionExpression struct {
	Prefix     string      `parser:"@Prefix"` // Expect "x-fir-"
	ActionName string      `parser:"@Ident"`
	Params     *Parameters `parser:"(':' @@)?"` // Optional parameters part starting with ':'
}

type Parameters struct {
	// Use pointers to distinguish between matched cases
	// Handles cases like [param1, param2,...] or []
	BracketedParams *[]string `parser:"( '[' (@Ident (',' @Ident)*)? ']' "`
	// Handles the case like :param1
	SingleParam *string `parser:"| @Ident )"`
}

// --- End Action Expression ---

// Cache for the action expression parser
var (
	actionParser     *participle.Parser[ActionExpression]
	actionParserErr  error
	actionParserOnce sync.Once
)

// getActionExpressionParser builds and returns the parser for action expressions,
// caching the result for efficiency.
func getActionExpressionParser() (*participle.Parser[ActionExpression], error) {
	actionParserOnce.Do(func() {
		actionParser, actionParserErr = participle.Build[ActionExpression](
			participle.Lexer(actionExpressionLexerRules),
			participle.Elide("Whitespace"),
		)
	})
	return actionParser, actionParserErr
}

// parseActionExpression parses an attribute key string like "x-fir-actionName:[param1,param2]" or "x-fir-actionName:param1".
// It returns the parsed action name and parameters.
func parseActionExpression(key string) (actionName string, params []string, err error) {
	parser, buildErr := getActionExpressionParser()
	if buildErr != nil {
		return "", nil, fmt.Errorf("error building action key parser: %w", buildErr)
	}

	// Trim whitespace just in case (although Elide should handle most)
	key = strings.TrimSpace(key)
	if key == "" {
		return "", nil, fmt.Errorf("action key cannot be empty")
	}
	// Basic prefix check before parsing
	if !strings.HasPrefix(key, "x-fir-") {
		// This check might be redundant if the lexer is strict, but good for early exit.
		return "", nil, fmt.Errorf("invalid prefix for action key: expected 'x-fir-'")
	}

	parsed, err := parser.ParseString("", key) // Parse ONLY the key
	if err != nil {
		// Wrap the participle error for more context
		return "", nil, fmt.Errorf("error parsing action key '%s': %w", key, err)
	}

	actionName = parsed.ActionName
	params = []string{} // Default to empty slice

	if parsed.Params != nil { // Check if the optional ':' and parameters were present
		if parsed.Params.BracketedParams != nil {
			// Use the pointer dereference safely, ensuring it's not nil first
			// The parser rule ('[' (@Ident (',' @Ident)*)? ']') ensures BracketedParams is non-nil if matched,
			// even for `[]`, where it will point to an empty slice.
			params = *parsed.Params.BracketedParams
		} else if parsed.Params.SingleParam != nil {
			// Create a slice with the single parameter
			params = []string{*parsed.Params.SingleParam}
		}
		// If parsed.Params is not nil, but both BracketedParams and SingleParam are nil,
		// it implies an issue with the grammar or an unexpected parse state.
		// The current grammar ('[' ... ']' | @Ident) should ensure one of them is non-nil if Params itself is non-nil.
	}

	return actionName, params, nil
}

// ... rest of parse.go ...
