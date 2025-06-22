package firattr

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/valyala/bytebufferpool"
	"golang.org/x/net/html"
)

// EventFilter represents a parsed event filter like [event1:ok,event2:ok]
type EventFilter struct {
	BeforeBracket string
	Values        []string
	AfterBracket  string
}

var ErrorEventFilterFormat = fmt.Errorf("error parsing event filter. must match ^[a-zA-Z0-9-]+:(ok|pending|error|done)$")

// GetEventNsList checks if the event string is of the format [event1:ok,event2:ok]:tmpl1
// and returns the unbundled list of event strings: event1:ok:tmpl1,event2:ok:tmpl1.
// If not, returns original event string.
func GetEventNsList(input string) ([]string, bool) {
	ef, err := GetEventFilter(input)
	if err != nil {
		return []string{input}, false
	}
	if ef == nil {
		return []string{input}, false
	}
	if len(ef.Values) == 0 {
		return []string{input}, false
	}
	var eventnsList []string
	for _, v := range ef.Values {
		eventnsList = append(eventnsList, ef.BeforeBracket+v+ef.AfterBracket)
	}
	return eventnsList, true
}

// GetEventFilter parses event filter syntax like [event1:ok,event2:ok]
func GetEventFilter(input string) (*EventFilter, error) {
	// Extract the part of the string before the open square bracket
	beforeRe := regexp.MustCompile(`^(.*?)\[`)
	beforeMatch := beforeRe.FindStringSubmatch(input)

	var beforeBracket string
	if len(beforeMatch) == 2 {
		beforeBracket = beforeMatch[1]
	}

	// Extract the part of the string after the closed square bracket
	afterRe := regexp.MustCompile(`\](.*)$`)
	afterMatch := afterRe.FindStringSubmatch(input)
	var afterBracket string
	if len(afterMatch) == 2 {
		afterBracket = afterMatch[1]
	}

	// Extract the contents of a closed square bracket
	re := regexp.MustCompile(`\[(.*?)\]`)
	matches := re.FindStringSubmatch(input)
	if len(matches) < 2 {
		return nil, nil
	}

	// Remove whitespace and split the contents by comma
	contents := strings.ReplaceAll(matches[1], " ", "")
	values := strings.Split(contents, ",")

	// Validate and format each value
	validValues := make([]string, 0)
	for _, value := range values {
		if !IsValidEventFilterFormat(value) {
			return nil, ErrorEventFilterFormat
		}
		validValues = append(validValues, FormatValue(value))
	}

	extractedValues := &EventFilter{
		BeforeBracket: beforeBracket,
		Values:        validValues,
		AfterBracket:  afterBracket,
	}

	return extractedValues, nil
}

// IsValidEventFilterFormat validates event filter value format
func IsValidEventFilterFormat(value string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9-]+:(ok|pending|error|done)$`)
	return re.MatchString(value)
}

// FormatValue formats an event filter value
func FormatValue(value string) string {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return value
	}
	return fmt.Sprintf("%s:%s", parts[0], parts[1])
}

// EventTemplate represents a template mapping for an event
type EventTemplate map[string]struct{}

// EventTemplates represents a mapping of event IDs to templates
type EventTemplates map[string]EventTemplate

// FirAttributes extracts all fir: and x-on:fir: attributes from an HTML node tree
func FirAttributes(n *html.Node) []html.Attribute {
	var attributes []html.Attribute
	if n.Type == html.ElementNode {
		for _, attr := range n.Attr {
			if !strings.HasPrefix(attr.Key, "@fir:") && !strings.HasPrefix(attr.Key, "x-on:fir:") {
				continue
			}
			attributes = append(attributes, attr)
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		attributes = append(FirAttributes(c), attributes...)
	}
	return attributes
}

// EventTemplatesFromAttr converts an HTML attribute to event templates
func EventTemplatesFromAttr(attr html.Attribute, templateNameRegex *regexp.Regexp) EventTemplates {
	evt := make(EventTemplates)
	eventns := strings.TrimPrefix(attr.Key, "@fir:")
	eventns = strings.TrimPrefix(eventns, "x-on:fir:")

	eventns = RemoveModifiers(eventns)
	// eventns might have a filter:[e1:ok,e2:ok] containing multiple event:state separated by comma
	eventnsList, _ := GetEventNsList(eventns)

	for _, eventns := range eventnsList {
		eventns = strings.TrimSpace(eventns)
		// set @fir|x-on:fir:eventns attribute to the node

		// myevent:ok::myblock
		eventnsParts := strings.SplitN(eventns, "::", -1)
		if len(eventnsParts) == 0 {
			continue
		}

		// [myevent:ok, myblock]
		if len(eventnsParts) > 2 {
			continue // Skip invalid format
		}

		// myevent:ok
		eventID := eventnsParts[0]
		// [myevent, ok]
		eventIDParts := strings.SplitN(eventID, ":", -1)
		if len(eventIDParts) != 2 {
			continue // Skip invalid format
		}
		// event name can only be followed by ok, error, pending, done
		if !slices.Contains([]string{"ok", "error", "pending", "done"}, eventIDParts[1]) {
			continue // Skip invalid state
		}
		// assert myevent:ok::myblock or myevent:error::myblock and skip if not
		if len(eventnsParts) == 2 && !slices.Contains([]string{"ok", "error"}, eventIDParts[1]) {
			continue
		}
		// template name is declared for event state i.e. myevent:ok::myblock
		templateName := "-"
		if len(eventnsParts) == 2 {
			templateName = eventnsParts[1]
		}

		templates, ok := evt[eventID]
		if !ok {
			templates = make(EventTemplate)
		}

		if templateNameRegex != nil && !templateNameRegex.MatchString(templateName) {
			continue // Skip invalid template name
		}

		templates[templateName] = struct{}{}
		evt[eventID] = templates
	}

	return evt
}

// DeepMergeEventTemplates merges two EventTemplates maps
func DeepMergeEventTemplates(evt1, evt2 EventTemplates) EventTemplates {
	merged := make(EventTemplates)
	for eventID, templatesMap := range evt1 {
		merged[eventID] = templatesMap
	}
	for eventID, templatesMap := range evt2 {
		templatesMap1, ok := merged[eventID]
		if !ok {
			merged[eventID] = templatesMap
			continue
		}
		for templateName := range templatesMap {
			templatesMap1[templateName] = struct{}{}
		}
	}
	return merged
}

// GetClassName generates a CSS class name from an event namespace
// Converts eventns like "myevent:ok" to "myevent-ok"
func GetClassName(eventns string) string {
	return strings.ReplaceAll(eventns, ":", "-")
}

// GetClassNameWithKey generates a CSS class name with an optional key suffix
// If key is provided and not empty, appends "--{key}" to the class name
func GetClassNameWithKey(eventns string, key *string) string {
	className := GetClassName(eventns)
	if key != nil && *key != "" {
		className = fmt.Sprintf("%s--%s", className, *key)
	}
	return className
}

// HTML Node Utility Functions

// RemoveAttr removes an attribute from an HTML node
func RemoveAttr(n *html.Node, attr string) {
	for i, a := range n.Attr {
		if a.Key == attr {
			n.Attr = append(n.Attr[:i], n.Attr[i+1:]...)
			break
		}
	}
}

// HasAttr checks if an HTML node has a specific attribute
func HasAttr(n *html.Node, attr string) bool {
	for _, a := range n.Attr {
		if a.Key == attr {
			return true
		}
	}
	return false
}

// SetAttr sets an attribute on an HTML node
func SetAttr(n *html.Node, key, val string) {
	n.Attr = append(n.Attr, html.Attribute{Key: key, Val: val})
}

// GetAttr gets the value of an attribute from an HTML node
func GetAttr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

// RemoveModifiers removes modifiers from an event namespace string and returns the base event
func RemoveModifiers(eventns string) string {
	baseEvent, _ := ProcessEventNamespace(eventns)
	return baseEvent
}

// ProcessEventNamespace processes an event namespace string by removing modifiers
// and extracting the base event name and modifiers separately
func ProcessEventNamespace(eventns string) (baseEvent string, modifiers string) {
	eventnsParts := strings.SplitN(eventns, ".", -1)
	if len(eventnsParts) > 0 {
		baseEvent = eventnsParts[0]
	}
	if len(eventnsParts) > 1 {
		modifierParts := eventnsParts[1:]
		modifiers = strings.Join(modifierParts, ".")
	}
	return baseEvent, modifiers
}

// IsFirEvent checks if an attribute key is a fir event attribute
func IsFirEvent(key string) bool {
	return strings.HasPrefix(key, "@fir") || strings.HasPrefix(key, "x-on:fir")
}

// ExtractTemplateName extracts the template name from an event string containing "::"
func ExtractTemplateName(key string) string {
	parts := strings.Split(key, "::")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

// IsHTMLTemplate checks if a block contains HTML template syntax
func IsHTMLTemplate(block string) bool {
	return strings.Contains(block, "{{") && strings.Contains(block, "}}")
}

// SetKeyToChildren recursively sets the "fir-key" attribute to all nested children
func SetKeyToChildren(node *html.Node, key string) {
	if node == nil || node.Type != html.ElementNode {
		return
	}

	if key == "" {
		for _, attr := range node.Attr {
			if attr.Key == "fir-key" {
				key = attr.Val
				break
			}
		}
	} else {
		for _, attr := range node.Attr {
			if attr.Key == "fir-key" {
				if key != attr.Val {
					SetKeyToChildren(node, attr.Val)
				}
				break
			}
		}
	}

	if key == "" {
		return
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		SetKeyToChildren(child, key)

		if child.Type == html.ElementNode {
			hasPrefixAttribute := false
			for _, attr := range child.Attr {
				if strings.HasPrefix(attr.Key, "@") || strings.HasPrefix(attr.Key, "x-on") {
					hasPrefixAttribute = true
					break
				}
			}

			if !hasPrefixAttribute {
				continue
			}

			hasKeyAttribute := false
			for _, attr := range child.Attr {
				if attr.Key == "fir-key" {
					hasKeyAttribute = true
					break
				}
			}

			if !hasKeyAttribute {
				child.Attr = append(child.Attr, html.Attribute{Key: "fir-key", Val: key})
			}
		}
	}
}

// HTMLNodeToBytes converts an HTML node to bytes
func HTMLNodeToBytes(n *html.Node) []byte {
	return []byte(HTMLNodeToString(n))
}

// HTMLNodeToString converts an HTML node to string
func HTMLNodeToString(n *html.Node) string {
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)
	err := html.Render(buf, n)
	if err != nil {
		panic(fmt.Sprintf("failed to render HTML: %v", err))
	}
	return html.UnescapeString(buf.String())
}

// EventFormatError returns a formatted error message for invalid event namespace
func EventFormatError(eventns string) string {
	return fmt.Sprintf(`
	error: invalid event namespace: %s. must be of either of the three formats =>
	1. @fir:<event>:<state:ok|error|pending|done>::<block-name(optional)>
	2. @fir:[event1:state,event2:state]::<block-name(optional)>
	`, eventns)
}
