// Package firattr provides self-contained parsing and analysis of fir: attributes.
// This package consolidates all fir: attribute parsing logic that was previously
// scattered across lexer.go, translate.go, actions.go, parse.go, readattr.go, and writeattr.go.
package firattr

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// ParsedAttribute represents a parsed fir: attribute with its components
type ParsedAttribute struct {
	Events    []EventInfo `json:"events"`    // List of events this attribute listens for
	Template  string      `json:"template"`  // Optional template target
	Action    string      `json:"action"`    // Action to execute
	Modifiers []string    `json:"modifiers"` // Event modifiers (e.g., debounce, throttle)
}

// EventInfo represents information about a single event
type EventInfo struct {
	Name  string `json:"name"`  // Event name (e.g., "click", "submit")
	State string `json:"state"` // Event state (e.g., "ok", "error", "pending", "done")
}

// AttributeExtractor can extract fir: attributes from template content
type AttributeExtractor struct {
	parser *participle.Parser[Expressions]
}

// NewAttributeExtractor creates a new attribute extractor
func NewAttributeExtractor() (*AttributeExtractor, error) {
	parser, err := getRenderExpressionParser()
	if err != nil {
		return nil, fmt.Errorf("failed to create parser: %w", err)
	}
	return &AttributeExtractor{parser: parser}, nil
}

// ParseExpression parses a fir: expression string into structured information
func (e *AttributeExtractor) ParseExpression(input string) (*ParsedAttribute, error) {
	if input == "" {
		return nil, fmt.Errorf("expression cannot be empty")
	}

	parsed, err := parseRenderExpression(e.parser, input)
	if err != nil {
		return nil, fmt.Errorf("failed to parse expression: %w", err)
	}

	if len(parsed.Expressions) == 0 {
		return nil, fmt.Errorf("no expressions found")
	}

	// For simplicity, we'll process the first expression
	// In practice, you might want to handle multiple expressions differently
	expr := parsed.Expressions[0]

	var events []EventInfo
	var template string
	var action string
	modifierSet := make(map[string]struct{})

	// Collect events and determine targets
	for _, binding := range expr.Bindings {
		for _, eventExpr := range binding.Eventexpressions {
			state := eventExpr.State
			if state == "" {
				state = ":ok" // Default state
			}
			events = append(events, EventInfo{
				Name:  eventExpr.Name,
				State: strings.TrimPrefix(state, ":"),
			})

			// Collect modifiers
			for _, mod := range eventExpr.Modifiers {
				modifierSet[strings.TrimPrefix(mod, ".")] = struct{}{}
			}
		}

		// Process target information
		if binding.Target != nil {
			if binding.Target.Template != "" {
				template = binding.Target.Template
			}
			if binding.Target.Action != "" {
				action = binding.Target.Action
			}
		}
	}

	// Convert modifier set to sorted slice
	modifiers := make([]string, 0)
	for mod := range modifierSet {
		modifiers = append(modifiers, mod)
	}
	sort.Strings(modifiers)

	// Default action if none specified
	if action == "" {
		action = "$fir.replace()"
	}

	return &ParsedAttribute{
		Events:    events,
		Template:  template,
		Action:    action,
		Modifiers: modifiers,
	}, nil
}

// ExtractFromTemplate extracts all fir: attributes from template content
// This is a simplified version - in practice, you'd want to use a proper HTML parser
func (e *AttributeExtractor) ExtractFromTemplate(templateContent string) ([]ParsedAttribute, error) {
	var attributes []ParsedAttribute

	// Simple regex-based extraction for demonstration
	// In practice, you'd want to use html/template parsing or HTML parsing
	lines := strings.Split(templateContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for fir: attributes (this is a simplified approach)
		if strings.Contains(line, "fir:") {
			// Extract the fir: expression (this is very basic)
			// In a real implementation, you'd parse the HTML properly
			start := strings.Index(line, "fir:")
			if start != -1 {
				// Find the end of the attribute
				remaining := line[start:]
				if idx := strings.Index(remaining, "\""); idx != -1 {
					remaining = remaining[idx+1:]
					if endIdx := strings.Index(remaining, "\""); endIdx != -1 {
						expression := remaining[:endIdx]
						if parsed, err := e.ParseExpression(expression); err == nil {
							attributes = append(attributes, *parsed)
						}
					}
				}
			}
		}
	}

	return attributes, nil
}

// GetEventIDs returns a list of all event IDs that this attribute listens for
func (attr *ParsedAttribute) GetEventIDs() []string {
	var eventIDs []string
	for _, event := range attr.Events {
		eventIDs = append(eventIDs, event.Name)
	}
	return eventIDs
}

// ToCanonicalForm converts the parsed attribute back to canonical @fir: format
func (attr *ParsedAttribute) ToCanonicalForm() string {
	if len(attr.Events) == 0 {
		return ""
	}

	// Format events
	var eventStrings []string
	for _, event := range attr.Events {
		eventStrings = append(eventStrings, fmt.Sprintf("%s:%s", event.Name, event.State))
	}

	var eventsPart string
	if len(eventStrings) == 1 {
		eventsPart = eventStrings[0]
	} else {
		eventsPart = fmt.Sprintf("[%s]", strings.Join(eventStrings, ","))
	}

	// Build the canonical form
	result := fmt.Sprintf("@fir:%s", eventsPart)

	// Add template if present
	if attr.Template != "" {
		result += fmt.Sprintf("::%s", attr.Template)
	}

	// Add modifiers
	if len(attr.Modifiers) > 0 {
		result += "." + strings.Join(attr.Modifiers, ".")
	}

	// Add action
	result += fmt.Sprintf("=\"%s\"", attr.Action)

	return result
}

// Example usage for static analysis tools
func AnalyzeTemplate(templateContent string) (map[string][]string, error) {
	extractor, err := NewAttributeExtractor()
	if err != nil {
		return nil, fmt.Errorf("failed to create extractor: %w", err)
	}

	attributes, err := extractor.ExtractFromTemplate(templateContent)
	if err != nil {
		return nil, fmt.Errorf("failed to extract attributes: %w", err)
	}

	// Build a map of event names to their states for analysis
	eventMap := make(map[string][]string)
	for _, attr := range attributes {
		for _, event := range attr.Events {
			eventMap[event.Name] = append(eventMap[event.Name], event.State)
		}
	}

	return eventMap, nil
}

// --- Action Expression Parsing (from lexer.go) ---

// ActionExpression represents a parsed x-fir-* action attribute
type ActionExpression struct {
	Prefix     string      `parser:"@Prefix"` // Expect "x-fir-"
	ActionName string      `parser:"@Ident"`
	Params     *Parameters `parser:"(':' @@)?"` // Optional parameters part starting with ':'
}

// Parameters represents the parameters part of an action expression
type Parameters struct {
	// Use pointers to distinguish between matched cases
	// Handles cases like [param1, param2,...] or []
	BracketedParams *[]string `parser:"'[' (@Ident (',' @Ident)*)? ']' |"`
	SingleParam     *string   `parser:"@Ident"`
}

// Define the lexer rules for parsing action expressions
var actionExpressionLexerRules = lexer.MustSimple([]lexer.SimpleRule{
	{Name: "Prefix", Pattern: `x-fir-`},
	{Name: "Ident", Pattern: `[a-zA-Z0-9_-]+`}, // Action names and parameters
	{Name: "LBracket", Pattern: `\[`},
	{Name: "RBracket", Pattern: `\]`},
	{Name: "Comma", Pattern: `,`},
	{Name: "Colon", Pattern: `:`},
	{Name: "Whitespace", Pattern: `\s+`},
})

// Cache for the action parser
var (
	actionParser     *participle.Parser[ActionExpression]
	actionParserErr  error
	actionParserOnce sync.Once
)

// getActionExpressionParser builds and returns the parser for action expressions
func getActionExpressionParser() (*participle.Parser[ActionExpression], error) {
	actionParserOnce.Do(func() {
		actionParser, actionParserErr = participle.Build[ActionExpression](
			participle.Lexer(actionExpressionLexerRules),
			participle.Elide("Whitespace"),
		)
	})
	return actionParser, actionParserErr
}

// ParseActionExpression parses an attribute key string like "x-fir-actionName:[param1,param2]" or "x-fir-actionName:param1".
// It returns the parsed action name and parameters.
func ParseActionExpression(key string) (actionName string, params []string, err error) {
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

// --- Render Expression Parsing (consolidating from lexer.go and translate.go) ---

// Define the lexer rules for parsing the render expression
var renderExpressionLexerRules = lexer.MustSimple([]lexer.SimpleRule{
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
	Expressions []*Expression `parser:"@@ ( ';' @@ )* ';'?"`
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

// Cache for the parser
var (
	renderParser     *participle.Parser[Expressions]
	renderParserErr  error
	renderParserOnce sync.Once
)

// getRenderExpressionParser builds and returns the parser for render expressions,
// caching the result for efficiency.
func getRenderExpressionParser() (*participle.Parser[Expressions], error) {
	renderParserOnce.Do(func() {
		renderParser, renderParserErr = participle.Build[Expressions](
			participle.Lexer(renderExpressionLexerRules),
			participle.Elide("Whitespace"), // Ignore standalone whitespace globally
		)
	})
	return renderParser, renderParserErr
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

// parseRenderExpression parses the input using the cached parser
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

// --- Public API for external packages ---

// Parser is a type alias for participle parser
type Parser[T any] = participle.Parser[T]

// GetRenderExpressionParser returns a parser for render expressions.
// This is used by translate.go, actions.go, and other packages.
func GetRenderExpressionParser() (*Parser[Expressions], error) {
	return getRenderExpressionParser()
}

// ParseRenderExpression parses a render expression string.
// This consolidates the parseRenderExpression function from lexer.go and other files.
func ParseRenderExpression(input string) (*Expressions, error) {
	parser, err := getRenderExpressionParser()
	if err != nil {
		return nil, fmt.Errorf("failed to get parser: %w", err)
	}
	return parseRenderExpression(parser, input)
}

// --- Exported types for external use ---

// These types are exported so other packages can use them directly
