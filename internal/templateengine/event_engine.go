package templateengine

import (
	"html/template"
	"regexp"
	"strings"
)

// EventTemplateEngine defines the interface for extracting and managing event templates.
// Event templates are HTML templates that are associated with specific events and states.
type EventTemplateEngine interface {
	// ExtractEventTemplates extracts event template mappings from an HTML template.
	// Returns a map of event IDs to their associated templates and states.
	ExtractEventTemplates(template Template) (EventTemplateMap, error)

	// RenderEventTemplate renders a specific event template with the given state and data.
	RenderEventTemplate(template Template, eventID string, state string, data interface{}) (string, error)

	// GetEventTemplateRegistry returns the registry for managing event templates.
	GetEventTemplateRegistry() EventTemplateRegistry

	// ValidateEventTemplate validates that an event template is properly formed.
	ValidateEventTemplate(eventID string, state string, templateName string) error
}

// EventTemplateRegistry manages the storage and retrieval of event templates.
// This allows for caching and efficient lookup of event template associations.
type EventTemplateRegistry interface {
	// Register associates an event template with an event ID and state.
	Register(eventID string, state string, templateName string)

	// Get retrieves all templates associated with an event ID.
	Get(eventID string) map[string][]string

	// GetByState retrieves templates for a specific event ID and state.
	GetByState(eventID string, state string) []string

	// Clear removes all event template associations.
	Clear()

	// GetAll returns all event template associations.
	GetAll() EventTemplateMap

	// Merge combines this registry with another registry.
	Merge(other EventTemplateRegistry)
}

// EventTemplateExtractor extracts event template information from HTML content.
// This interface abstracts the HTML parsing and attribute extraction logic.
type EventTemplateExtractor interface {
	// Extract parses HTML content and extracts event template associations.
	Extract(content []byte) (EventTemplateMap, error)

	// ExtractFromTemplate extracts event templates from a parsed template.
	ExtractFromTemplate(tmpl *template.Template) (EventTemplateMap, error)

	// SetTemplateNameRegex sets the regex pattern for validating template names.
	SetTemplateNameRegex(regex *regexp.Regexp)

	// GetSupportedAttributes returns the list of supported fir: attributes.
	GetSupportedAttributes() []string
}

// DefaultEventTemplateEngine implements EventTemplateEngine using the existing
// Fir framework's event template extraction logic.
type DefaultEventTemplateEngine struct {
	extractor EventTemplateExtractor
	registry  EventTemplateRegistry
}

// NewDefaultEventTemplateEngine creates a new default event template engine.
func NewDefaultEventTemplateEngine() *DefaultEventTemplateEngine {
	return &DefaultEventTemplateEngine{
		extractor: NewHTMLEventTemplateExtractor(),
		registry:  NewInMemoryEventTemplateRegistry(),
	}
}

// NewDefaultEventTemplateEngineWithExtractor creates a new event template engine
// with a custom extractor.
func NewDefaultEventTemplateEngineWithExtractor(extractor EventTemplateExtractor) *DefaultEventTemplateEngine {
	return &DefaultEventTemplateEngine{
		extractor: extractor,
		registry:  NewInMemoryEventTemplateRegistry(),
	}
}

// ExtractEventTemplates implements EventTemplateEngine interface.
func (dete *DefaultEventTemplateEngine) ExtractEventTemplates(template Template) (EventTemplateMap, error) {
	if template == nil {
		return make(EventTemplateMap), nil
	}

	// Try to get the underlying template if it's a GoTemplate
	if goTemplate, ok := template.(*GoTemplate); ok {
		underlyingTemplate := goTemplate.GetUnderlyingTemplate()
		if underlyingTemplate != nil {
			return dete.extractor.ExtractFromTemplate(underlyingTemplate)
		}
	}

	// Fall back to extracting from template content if possible
	// This would require the template to have a way to get its source content
	// For now, return empty map as this is a placeholder implementation
	return make(EventTemplateMap), nil
}

// RenderEventTemplate implements EventTemplateEngine interface.
func (dete *DefaultEventTemplateEngine) RenderEventTemplate(template Template, eventID string, state string, data interface{}) (string, error) {
	if template == nil {
		return "", ErrInvalidTemplate
	}

	// First check if the template is registered in the registry
	registeredTemplates := dete.registry.GetByState(eventID, state)
	var templateName string

	if len(registeredTemplates) > 0 {
		// Use the first registered template name
		templateName = registeredTemplates[0]
	} else {
		// Fall back to extracting event templates from the current template
		eventTemplates, err := dete.ExtractEventTemplates(template)
		if err != nil {
			return "", err
		}

		// Check if the event and state exist
		stateMap, exists := eventTemplates[eventID]
		if !exists {
			return "", ErrEventNotFound
		}

		_, exists = stateMap[state]
		if !exists {
			return "", ErrEventStateNotFound
		}

		// Use a simple template name construction
		templateName = eventID + "_" + state
	}

	// Look up the template by name
	namedTemplate := template.Lookup(templateName)
	if namedTemplate == nil {
		return "", ErrTemplateNotFound
	}

	// Render the template
	var buf strings.Builder
	err := namedTemplate.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// GetEventTemplateRegistry implements EventTemplateEngine interface.
func (dete *DefaultEventTemplateEngine) GetEventTemplateRegistry() EventTemplateRegistry {
	return dete.registry
}

// ValidateEventTemplate implements EventTemplateEngine interface.
func (dete *DefaultEventTemplateEngine) ValidateEventTemplate(eventID string, state string, templateName string) error {
	if eventID == "" {
		return ErrInvalidEventID
	}
	if state == "" {
		return ErrInvalidEventState
	}
	if templateName == "" {
		return ErrInvalidTemplateName
	}
	return nil
}

// SetExtractor sets the event template extractor.
func (dete *DefaultEventTemplateEngine) SetExtractor(extractor EventTemplateExtractor) {
	dete.extractor = extractor
}

// GetExtractor returns the current event template extractor.
func (dete *DefaultEventTemplateEngine) GetExtractor() EventTemplateExtractor {
	return dete.extractor
}
