package fir

import (
	"fmt"
	"regexp" // Import regexp
	"strings"
	"sync" // Import sync for caching

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// Define the lexer rules for parsing the render expression
var renderExpressionlexerRules = lexer.MustSimple([]lexer.SimpleRule{
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
		participle.Lexer(renderExpressionlexerRules),
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
