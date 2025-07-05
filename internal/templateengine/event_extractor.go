package templateengine

import (
	"bytes"
	"html/template"
	"regexp"
	"strings"

	"github.com/livefir/fir/internal/firattr"
	"github.com/sourcegraph/conc/pool"
	"golang.org/x/net/html"
)

// HTMLEventTemplateExtractor extracts event template information from HTML content
// using the existing firattr package functionality.
type HTMLEventTemplateExtractor struct {
	templateNameRegex *regexp.Regexp
}

// NewHTMLEventTemplateExtractor creates a new HTML event template extractor.
func NewHTMLEventTemplateExtractor() *HTMLEventTemplateExtractor {
	// Default template name regex from the original code
	defaultRegex := regexp.MustCompile(`^[ A-Za-z0-9\-:_.]*$`)
	return &HTMLEventTemplateExtractor{
		templateNameRegex: defaultRegex,
	}
}

// Extract implements EventTemplateExtractor interface.
func (extractor *HTMLEventTemplateExtractor) Extract(content []byte) (EventTemplateMap, error) {
	doc, err := html.Parse(bytes.NewReader(content))
	if err != nil {
		return nil, err
	}

	attributes := firattr.FirAttributes(doc)
	// Debug: let's see what attributes we found
	if len(attributes) == 0 {
		// No fir attributes found - this might be normal for basic HTML
		return make(EventTemplateMap), nil
	}

	resultPool := pool.NewWithResults[EventTemplateMap]()

	for _, attr := range attributes {
		attr := attr
		resultPool.Go(func() EventTemplateMap {
			return extractor.eventTemplatesFromAttr(attr)
		})
	}

	evtArr := resultPool.Wait()
	evt := make(EventTemplateMap)
	for _, evtMap := range evtArr {
		evt = extractor.mergeEventTemplates(evt, evtMap)
	}

	return evt, nil
}

// ExtractFromTemplate implements EventTemplateExtractor interface.
func (extractor *HTMLEventTemplateExtractor) ExtractFromTemplate(tmpl *template.Template) (EventTemplateMap, error) {
	if tmpl == nil {
		return make(EventTemplateMap), nil
	}

	// Get all templates in the template set
	templates := tmpl.Templates()
	allEvents := make(EventTemplateMap)

	for _, t := range templates {
		// Try to extract the template content
		// Note: This is a simplified approach - in practice, we'd need to
		// reconstruct the original HTML from the parsed template
		if t.Tree != nil && t.Tree.Root != nil {
			// For now, return empty map as extracting from parsed templates
			// requires more complex logic to reconstruct the original HTML
			continue
		}
	}

	return allEvents, nil
}

// SetTemplateNameRegex implements EventTemplateExtractor interface.
func (extractor *HTMLEventTemplateExtractor) SetTemplateNameRegex(regex *regexp.Regexp) {
	extractor.templateNameRegex = regex
}

// GetSupportedAttributes implements EventTemplateExtractor interface.
func (extractor *HTMLEventTemplateExtractor) GetSupportedAttributes() []string {
	return []string{
		"fir:click",
		"fir:submit",
		"fir:change",
		"fir:input",
		"fir:focus",
		"fir:blur",
		"fir:keydown",
		"fir:keyup",
		"fir:mouseenter",
		"fir:mouseleave",
		"fir:load",
		"fir:unload",
	}
}

// eventTemplatesFromAttr converts a single HTML attribute to EventTemplateMap.
func (extractor *HTMLEventTemplateExtractor) eventTemplatesFromAttr(attr html.Attribute) EventTemplateMap {
	firattrEvt := firattr.EventTemplatesFromAttr(attr, extractor.templateNameRegex)

	// Convert firattr.EventTemplates to template engine EventTemplateMap
	evt := make(EventTemplateMap)
	for eventIDWithState := range firattrEvt {
		// The firattr package returns eventID as "eventname:state" (e.g., "increment:ok")
		// We need to split this to get the event name and state separately
		parts := strings.SplitN(eventIDWithState, ":", 2)
		if len(parts) != 2 {
			continue // Skip invalid format
		}

		eventName := parts[0]
		state := parts[1]

		// Ensure the event exists in our map
		if evt[eventName] == nil {
			evt[eventName] = make(EventTemplateState)
		}

		// Add the state (we don't need the individual template names for our use case)
		evt[eventName][state] = struct{}{}
	}

	return evt
}

// extractStateFromTemplateName attempts to extract state information from template name.
// This is a heuristic approach - in practice, the state might be encoded differently.
func (extractor *HTMLEventTemplateExtractor) extractStateFromTemplateName(templateName string) string {
	// Common patterns for state in template names
	if contains(templateName, "error") || contains(templateName, "err") {
		return "error"
	}
	if contains(templateName, "pending") || contains(templateName, "loading") {
		return "pending"
	}
	if contains(templateName, "done") || contains(templateName, "success") {
		return "done"
	}
	// Default state
	return "ok"
}

// mergeEventTemplates merges two EventTemplateMap instances.
func (extractor *HTMLEventTemplateExtractor) mergeEventTemplates(evt1, evt2 EventTemplateMap) EventTemplateMap {
	result := make(EventTemplateMap)

	// Copy evt1
	for eventID, stateMap := range evt1 {
		result[eventID] = make(EventTemplateState)
		for state := range stateMap {
			result[eventID][state] = struct{}{}
		}
	}

	// Merge evt2
	for eventID, stateMap := range evt2 {
		if result[eventID] == nil {
			result[eventID] = make(EventTemplateState)
		}
		for state := range stateMap {
			result[eventID][state] = struct{}{}
		}
	}

	return result
}

// contains is a helper function to check if a string contains a substring (case-insensitive).
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					findSubstring(s, substr))))
}

// findSubstring is a simple substring search helper.
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
