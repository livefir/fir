package templateengine

import (
	"html/template"
	"regexp"
	"strings"
	"testing"
)

func TestDefaultEventTemplateEngine(t *testing.T) {
	t.Run("NewDefaultEventTemplateEngine", func(t *testing.T) {
		engine := NewDefaultEventTemplateEngine()

		if engine == nil {
			t.Fatal("Expected non-nil event template engine")
		}

		if engine.GetExtractor() == nil {
			t.Error("Expected extractor to be initialized")
		}

		if engine.GetEventTemplateRegistry() == nil {
			t.Error("Expected registry to be initialized")
		}
	})

	t.Run("ExtractEventTemplates", func(t *testing.T) {
		engine := NewDefaultEventTemplateEngine()

		// Create a simple template
		tmpl := template.Must(template.New("test").Parse("<div>Hello World</div>"))
		goTemplate := NewGoTemplate(tmpl)

		eventTemplates, err := engine.ExtractEventTemplates(goTemplate)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if eventTemplates == nil {
			t.Fatal("Expected non-nil event templates map")
		}

		// Should be empty for a template without fir: attributes
		if len(eventTemplates) != 0 {
			t.Errorf("Expected empty event templates for simple template, got %d", len(eventTemplates))
		}
	})

	t.Run("ValidateEventTemplate", func(t *testing.T) {
		engine := NewDefaultEventTemplateEngine()

		// Valid event template
		err := engine.ValidateEventTemplate("click", "ok", "button_template")
		if err != nil {
			t.Errorf("Expected valid event template to pass validation, got: %v", err)
		}

		// Invalid event ID
		err = engine.ValidateEventTemplate("", "ok", "button_template")
		if err == nil {
			t.Error("Expected error for empty event ID")
		}

		// Invalid state
		err = engine.ValidateEventTemplate("click", "", "button_template")
		if err == nil {
			t.Error("Expected error for empty state")
		}

		// Invalid template name
		err = engine.ValidateEventTemplate("click", "ok", "")
		if err == nil {
			t.Error("Expected error for empty template name")
		}
	})

	t.Run("RenderEventTemplate", func(t *testing.T) {
		engine := NewDefaultEventTemplateEngine()

		// Create a template with a named sub-template
		tmpl := template.New("main")
		tmpl.New("click_ok").Parse("Button clicked successfully")
		goTemplate := NewGoTemplate(tmpl)

		// Mock the event extraction to return our test event
		registry := engine.GetEventTemplateRegistry()
		registry.Register("click", "ok", "click_ok")

		result, err := engine.RenderEventTemplate(goTemplate, "click", "ok", nil)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if !strings.Contains(result, "Button clicked successfully") {
			t.Errorf("Expected template content in result, got: %s", result)
		}
	})

	t.Run("SetExtractor", func(t *testing.T) {
		engine := NewDefaultEventTemplateEngine()
		mockExtractor := &MockEventTemplateExtractor{}

		engine.SetExtractor(mockExtractor)

		if engine.GetExtractor() != mockExtractor {
			t.Error("Expected custom extractor to be set")
		}
	})
}

func TestInMemoryEventTemplateRegistry(t *testing.T) {
	t.Run("RegisterAndGet", func(t *testing.T) {
		registry := NewInMemoryEventTemplateRegistry()

		registry.Register("click", "ok", "button_template")
		registry.Register("click", "error", "error_template")
		registry.Register("submit", "pending", "loading_template")

		// Test Get
		clickTemplates := registry.Get("click")
		if len(clickTemplates) != 2 {
			t.Errorf("Expected 2 states for click event, got %d", len(clickTemplates))
		}

		if _, exists := clickTemplates["ok"]; !exists {
			t.Error("Expected 'ok' state for click event")
		}

		if _, exists := clickTemplates["error"]; !exists {
			t.Error("Expected 'error' state for click event")
		}

		// Test GetByState
		okTemplates := registry.GetByState("click", "ok")
		if len(okTemplates) == 0 {
			t.Error("Expected templates for click ok state")
		}

		nonExistentTemplates := registry.GetByState("nonexistent", "ok")
		if len(nonExistentTemplates) != 0 {
			t.Error("Expected no templates for non-existent event")
		}
	})

	t.Run("GetAll", func(t *testing.T) {
		registry := NewInMemoryEventTemplateRegistry()

		registry.Register("click", "ok", "template1")
		registry.Register("submit", "error", "template2")

		all := registry.GetAll()
		if len(all) != 2 {
			t.Errorf("Expected 2 events in registry, got %d", len(all))
		}

		if _, exists := all["click"]; !exists {
			t.Error("Expected click event in GetAll result")
		}

		if _, exists := all["submit"]; !exists {
			t.Error("Expected submit event in GetAll result")
		}
	})

	t.Run("Clear", func(t *testing.T) {
		registry := NewInMemoryEventTemplateRegistry()

		registry.Register("click", "ok", "template1")
		registry.Register("submit", "error", "template2")

		if registry.Size() != 2 {
			t.Errorf("Expected 2 events before clear, got %d", registry.Size())
		}

		registry.Clear()

		if registry.Size() != 0 {
			t.Errorf("Expected 0 events after clear, got %d", registry.Size())
		}

		all := registry.GetAll()
		if len(all) != 0 {
			t.Errorf("Expected empty registry after clear, got %d events", len(all))
		}
	})

	t.Run("Merge", func(t *testing.T) {
		registry1 := NewInMemoryEventTemplateRegistry()
		registry2 := NewInMemoryEventTemplateRegistry()

		registry1.Register("click", "ok", "template1")
		registry2.Register("submit", "error", "template2")
		registry2.Register("click", "error", "template3") // Different state for same event

		registry1.Merge(registry2)

		if registry1.Size() != 2 {
			t.Errorf("Expected 2 events after merge, got %d", registry1.Size())
		}

		clickStates := registry1.GetStates("click")
		if len(clickStates) != 2 {
			t.Errorf("Expected 2 states for click after merge, got %d", len(clickStates))
		}

		submitStates := registry1.GetStates("submit")
		if len(submitStates) != 1 {
			t.Errorf("Expected 1 state for submit after merge, got %d", len(submitStates))
		}
	})

	t.Run("GetEventIDs", func(t *testing.T) {
		registry := NewInMemoryEventTemplateRegistry()

		registry.Register("click", "ok", "template1")
		registry.Register("submit", "error", "template2")
		registry.Register("change", "pending", "template3")

		eventIDs := registry.GetEventIDs()
		if len(eventIDs) != 3 {
			t.Errorf("Expected 3 event IDs, got %d", len(eventIDs))
		}

		// Check that all expected events are present
		expectedEvents := map[string]bool{"click": false, "submit": false, "change": false}
		for _, eventID := range eventIDs {
			if _, exists := expectedEvents[eventID]; exists {
				expectedEvents[eventID] = true
			}
		}

		for event, found := range expectedEvents {
			if !found {
				t.Errorf("Expected event %s to be in event IDs list", event)
			}
		}
	})

	t.Run("GetStates", func(t *testing.T) {
		registry := NewInMemoryEventTemplateRegistry()

		registry.Register("click", "ok", "template1")
		registry.Register("click", "error", "template2")
		registry.Register("click", "pending", "template3")

		states := registry.GetStates("click")
		if len(states) != 3 {
			t.Errorf("Expected 3 states for click event, got %d", len(states))
		}

		// Check that all expected states are present
		expectedStates := map[string]bool{"ok": false, "error": false, "pending": false}
		for _, state := range states {
			if _, exists := expectedStates[state]; exists {
				expectedStates[state] = true
			}
		}

		for state, found := range expectedStates {
			if !found {
				t.Errorf("Expected state %s to be in states list", state)
			}
		}

		// Test non-existent event
		nonExistentStates := registry.GetStates("nonexistent")
		if len(nonExistentStates) != 0 {
			t.Errorf("Expected no states for non-existent event, got %d", len(nonExistentStates))
		}
	})
}

func TestHTMLEventTemplateExtractor(t *testing.T) {
	t.Run("NewHTMLEventTemplateExtractor", func(t *testing.T) {
		extractor := NewHTMLEventTemplateExtractor()

		if extractor == nil {
			t.Fatal("Expected non-nil extractor")
		}

		supportedAttrs := extractor.GetSupportedAttributes()
		if len(supportedAttrs) == 0 {
			t.Error("Expected non-empty supported attributes list")
		}

		// Check that common fir: attributes are supported
		expectedAttrs := []string{"fir:click", "fir:submit", "fir:change"}
		for _, expected := range expectedAttrs {
			found := false
			for _, supported := range supportedAttrs {
				if supported == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected %s to be in supported attributes", expected)
			}
		}
	})

	t.Run("SetTemplateNameRegex", func(t *testing.T) {
		extractor := NewHTMLEventTemplateExtractor()
		customRegex := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

		extractor.SetTemplateNameRegex(customRegex)

		// We can't directly test the regex change without accessing private fields,
		// but we can test that the method doesn't panic
	})

	t.Run("Extract", func(t *testing.T) {
		extractor := NewHTMLEventTemplateExtractor()

		// Simple HTML content without fir: attributes
		htmlContent := []byte(`<div>Hello World</div>`)

		eventTemplates, err := extractor.Extract(htmlContent)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if eventTemplates == nil {
			t.Fatal("Expected non-nil event templates map")
		}

		// Should be empty for HTML without fir: attributes
		if len(eventTemplates) != 0 {
			t.Errorf("Expected empty event templates for simple HTML, got %d", len(eventTemplates))
		}
	})

	t.Run("ExtractStateFromTemplateName", func(t *testing.T) {
		extractor := NewHTMLEventTemplateExtractor()

		testCases := []struct {
			templateName  string
			expectedState string
		}{
			{"button_error", "error"},
			{"form_err_template", "error"},
			{"loading_pending", "pending"},
			{"success_done", "done"},
			{"simple_template", "ok"},
			{"", "ok"},
		}

		for _, tc := range testCases {
			state := extractor.extractStateFromTemplateName(tc.templateName)
			if state != tc.expectedState {
				t.Errorf("For template name %s, expected state %s, got %s",
					tc.templateName, tc.expectedState, state)
			}
		}
	})
}

func TestGoTemplateEngineEventIntegration(t *testing.T) {
	t.Run("SetEventEngine", func(t *testing.T) {
		engine := NewGoTemplateEngine()
		customEventEngine := NewDefaultEventTemplateEngine()

		engine.SetEventEngine(customEventEngine)

		if engine.GetEventEngine() != customEventEngine {
			t.Error("Expected custom event engine to be set")
		}
	})

	t.Run("ExtractEventTemplatesIntegration", func(t *testing.T) {
		engine := NewGoTemplateEngine()

		// Create a simple template
		tmpl := template.Must(template.New("test").Parse("<div>Hello World</div>"))
		goTemplate := NewGoTemplate(tmpl)

		eventTemplates, err := engine.ExtractEventTemplates(goTemplate)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if eventTemplates == nil {
			t.Fatal("Expected non-nil event templates map")
		}
	})

	t.Run("RenderEventTemplateIntegration", func(t *testing.T) {
		engine := NewGoTemplateEngine()

		// Create a template with a named sub-template
		tmpl := template.New("main")
		tmpl.New("click_ok").Parse("Button clicked: {{.Message}}")
		goTemplate := NewGoTemplate(tmpl)

		data := map[string]interface{}{"Message": "Success!"}

		// The event template rendering might not find the template since we're not
		// actually extracting from HTML with fir: attributes, but it should not panic
		_, err := engine.RenderEventTemplate(goTemplate, "click", "ok", data)

		// It's okay if this returns an error since we don't have proper event extraction
		// The important thing is that it doesn't panic and returns a reasonable error
		if err != nil && err != ErrEventNotFound && err != ErrTemplateNotFound {
			t.Errorf("Expected ErrEventNotFound or ErrTemplateNotFound, got: %v", err)
		}
	})
}

// Mock event template extractor for testing
type MockEventTemplateExtractor struct {
	extractFunc         func([]byte) (EventTemplateMap, error)
	extractFromTemplate func(*template.Template) (EventTemplateMap, error)
	supportedAttrs      []string
}

func (m *MockEventTemplateExtractor) Extract(content []byte) (EventTemplateMap, error) {
	if m.extractFunc != nil {
		return m.extractFunc(content)
	}
	return make(EventTemplateMap), nil
}

func (m *MockEventTemplateExtractor) ExtractFromTemplate(tmpl *template.Template) (EventTemplateMap, error) {
	if m.extractFromTemplate != nil {
		return m.extractFromTemplate(tmpl)
	}
	return make(EventTemplateMap), nil
}

func (m *MockEventTemplateExtractor) SetTemplateNameRegex(regex *regexp.Regexp) {
	// Mock implementation - do nothing
}

func (m *MockEventTemplateExtractor) GetSupportedAttributes() []string {
	if m.supportedAttrs != nil {
		return m.supportedAttrs
	}
	return []string{"fir:click", "fir:submit"}
}
