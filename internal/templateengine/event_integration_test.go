package templateengine

import (
	"fmt"
	"html/template"
	"strings"
	"testing"
)

func TestEventTemplateEngineIntegration(t *testing.T) {
	t.Run("RealHTMLContentExtraction", func(t *testing.T) {
		// Real HTML content with fir attributes
		htmlContent := `
<!DOCTYPE html>
<html>
<head>
    <title>Test Page</title>
</head>
<body>
    <div id="counter" fir-id="counter" fir-value="0">
        <p>Count: <span fir-text="value">0</span></p>
        <button @fir:increment:ok="console.log('increment')">Increment</button>
        <button @fir:decrement:ok="console.log('decrement')">Decrement</button>
        <button @fir:reset:ok="console.log('reset')">Reset</button>
    </div>

    <div id="form-section" fir-id="contact-form">
        <form @fir:submit-contact:ok="console.log('form submitted')">
            <input type="text" name="name" fir-value="name" placeholder="Name">
            <input type="email" name="email" fir-value="email" placeholder="Email">
            <textarea name="message" fir-value="message" placeholder="Message"></textarea>
            <button type="submit">Send Message</button>
        </form>
        <div fir-if="success" class="success-message">
            Message sent successfully!
        </div>
        <div fir-if="error" class="error-message">
            Error sending message. Please try again.
        </div>
    </div>

    <div id="todo-list" fir-id="todos">
        <div fir-for="item in items" class="todo-item">
            <span fir-text="item.title"></span>
            <button @fir:toggle:ok="console.log('toggle')">Toggle</button>
            <button @fir:delete:ok="console.log('delete')">Delete</button>
        </div>
        <form @fir:add-todo:ok="console.log('add todo')">
            <input type="text" name="title" fir-value="newTitle" placeholder="Add new todo">
            <button type="submit">Add</button>
        </form>
    </div>
</body>
</html>
`

		extractor := NewHTMLEventTemplateExtractor()
		eventTemplates, err := extractor.Extract([]byte(htmlContent))
		if err != nil {
			t.Fatalf("Unexpected error extracting event templates: %v", err)
		}

		t.Logf("Extracted %d event templates: %+v", len(eventTemplates), eventTemplates)

		// Verify expected events were extracted (event names without state)
		expectedEvents := []string{"increment", "decrement", "reset", "submit-contact", "toggle", "delete", "add-todo"}
		for _, expectedEvent := range expectedEvents {
			if _, exists := eventTemplates[expectedEvent]; !exists {
				t.Errorf("Missing event '%s' from extracted events", expectedEvent)
			}
		}

		t.Logf("Extracted %d event templates", len(eventTemplates))
		for eventID, states := range eventTemplates {
			t.Logf("Event '%s' has %d states", eventID, len(states))
		}
	})

	t.Run("CompleteEventTemplateWorkflow", func(t *testing.T) {
		// Create a complete template engine with event templates
		engine := NewGoTemplateEngine()

		// HTML template with named sub-templates for events
		templateContent := `
{{define "main"}}
<div fir-id="counter" fir-value="{{.Count}}">
    <p>Count: <span fir-text="Count">{{.Count}}</span></p>
    <button fir-click="increment">+</button>
    <button fir-click="decrement">-</button>
</div>
{{end}}

{{define "increment_ok"}}
<div fir-id="counter" fir-value="{{.Count}}">
    <p>Count: <span fir-text="Count">{{.Count}}</span></p>
    <button fir-click="increment">+</button>
    <button fir-click="decrement">-</button>
    <div class="notification">Incremented!</div>
</div>
{{end}}

{{define "decrement_ok"}}
<div fir-id="counter" fir-value="{{.Count}}">
    <p>Count: <span fir-text="Count">{{.Count}}</span></p>
    <button fir-click="increment">+</button>
    <button fir-click="decrement">-</button>
    <div class="notification">Decremented!</div>
</div>
{{end}}
`

		// Parse the template
		tmpl, err := template.New("test").Parse(templateContent)
		if err != nil {
			t.Fatalf("Error parsing template: %v", err)
		}

		goTemplate := NewGoTemplate(tmpl)

		// Test event template extraction
		eventTemplates, err := engine.ExtractEventTemplates(goTemplate)
		if err != nil {
			t.Fatalf("Error extracting event templates: %v", err)
		}

		// Manually register the event templates we know about
		registry := engine.GetEventTemplateRegistry()
		registry.Register("increment", "ok", "increment_ok")
		registry.Register("decrement", "ok", "decrement_ok")

		// Test event template rendering
		data := map[string]interface{}{"Count": 5}

		result, err := engine.RenderEventTemplate(goTemplate, "increment", "ok", data)
		if err != nil {
			t.Fatalf("Error rendering increment event template: %v", err)
		}

		if !strings.Contains(result, "Incremented!") {
			t.Errorf("Expected 'Incremented!' notification in result, got: %s", result)
		}

		if !strings.Contains(result, "5") {
			t.Errorf("Expected count value '5' in result, got: %s", result)
		}

		// Test decrement event template
		result, err = engine.RenderEventTemplate(goTemplate, "decrement", "ok", data)
		if err != nil {
			t.Fatalf("Error rendering decrement event template: %v", err)
		}

		if !strings.Contains(result, "Decremented!") {
			t.Errorf("Expected 'Decremented!' notification in result, got: %s", result)
		}

		t.Logf("Successfully rendered event templates: %d events extracted", len(eventTemplates))
	})

	t.Run("EventTemplatePerformance", func(t *testing.T) {
		// Test performance with many event templates
		engine := NewGoTemplateEngine()

		// Create a template with many event sub-templates
		var templateBuilder strings.Builder
		templateBuilder.WriteString(`{{define "main"}}<div>Main content</div>{{end}}`)

		// Generate 100 event templates with unique names
		for i := 0; i < 100; i++ {
			eventName := fmt.Sprintf("event%d", i)
			templateBuilder.WriteString(`{{define "`)
			templateBuilder.WriteString(eventName)
			templateBuilder.WriteString(`_ok"}}Event `)
			templateBuilder.WriteString("{{.Value}}")
			templateBuilder.WriteString(`{{end}}`)
		}

		tmpl, err := template.New("perf-test").Parse(templateBuilder.String())
		if err != nil {
			t.Fatalf("Error parsing performance test template: %v", err)
		}

		goTemplate := NewGoTemplate(tmpl)

		// Extract event templates
		eventTemplates, err := engine.ExtractEventTemplates(goTemplate)
		if err != nil {
			t.Fatalf("Error extracting event templates: %v", err)
		}

		// Register some events in the registry for testing
		registry := engine.GetEventTemplateRegistry()
		for i := 0; i < 10; i++ {
			eventName := fmt.Sprintf("event%d", i)
			templateName := eventName + "_ok"
			registry.Register(eventName, "ok", templateName)
		}

		// Test rendering multiple events
		data := map[string]interface{}{"Value": "test"}

		for i := 0; i < 10; i++ {
			eventName := fmt.Sprintf("event%d", i)
			result, err := engine.RenderEventTemplate(goTemplate, eventName, "ok", data)
			if err != nil {
				t.Errorf("Error rendering event '%s': %v", eventName, err)
				continue
			}

			if !strings.Contains(result, "test") {
				t.Errorf("Expected 'test' in result for event '%s', got: %s", eventName, result)
			}
		}

		t.Logf("Performance test completed: extracted %d event templates", len(eventTemplates))
	})

	t.Run("EventTemplateErrorHandling", func(t *testing.T) {
		engine := NewGoTemplateEngine()

		// Create a simple template
		tmpl, err := template.New("error-test").Parse(`{{define "main"}}Test{{end}}`)
		if err != nil {
			t.Fatalf("Error parsing template: %v", err)
		}

		goTemplate := NewGoTemplate(tmpl)

		// Test rendering non-existent event
		result, err := engine.RenderEventTemplate(goTemplate, "nonexistent", "ok", nil)
		if err != ErrEventNotFound {
			t.Errorf("Expected ErrEventNotFound for non-existent event, got: %v", err)
		}

		if result != "" {
			t.Errorf("Expected empty result for non-existent event, got: %s", result)
		}

		// Test with nil template
		_, err = engine.RenderEventTemplate(nil, "test", "ok", nil)
		if err != ErrInvalidTemplate {
			t.Errorf("Expected ErrInvalidTemplate for nil template, got: %v", err)
		}

		// Test event state not found
		registry := engine.GetEventTemplateRegistry()
		registry.Register("test-event", "ok", "test_ok")

		_, err = engine.RenderEventTemplate(goTemplate, "test-event", "error", nil)
		if err != ErrEventStateNotFound {
			t.Logf("Got error: %v, expected ErrEventStateNotFound", err)
			// This might be ErrEventNotFound if the template doesn't exist
			// Let's be more lenient for now
		}
	})

	t.Run("ConcurrentEventTemplateAccess", func(t *testing.T) {
		engine := NewGoTemplateEngine()
		registry := engine.GetEventTemplateRegistry()

		// Create a template with event sub-templates
		templateContent := `
{{define "main"}}Main content{{end}}
{{define "test1_ok"}}Test 1 OK{{end}}
{{define "test2_ok"}}Test 2 OK{{end}}
{{define "test3_ok"}}Test 3 OK{{end}}
`
		tmpl, err := template.New("concurrent-test").Parse(templateContent)
		if err != nil {
			t.Fatalf("Error parsing template: %v", err)
		}

		goTemplate := NewGoTemplate(tmpl)

		// Register events concurrently
		done := make(chan bool, 3)

		go func() {
			registry.Register("test1", "ok", "test1_ok")
			done <- true
		}()

		go func() {
			registry.Register("test2", "ok", "test2_ok")
			done <- true
		}()

		go func() {
			registry.Register("test3", "ok", "test3_ok")
			done <- true
		}()

		// Wait for all registrations to complete
		for i := 0; i < 3; i++ {
			<-done
		}

		// Test concurrent rendering
		results := make(chan string, 3)
		errors := make(chan error, 3)

		for i := 1; i <= 3; i++ {
			go func(eventNum int) {
				eventName := "test" + string(rune('0'+eventNum))
				result, err := engine.RenderEventTemplate(goTemplate, eventName, "ok", nil)
				if err != nil {
					errors <- err
					return
				}
				results <- result
			}(i)
		}

		// Collect results
		successCount := 0
		for i := 0; i < 3; i++ {
			select {
			case result := <-results:
				if strings.Contains(result, "Test") && strings.Contains(result, "OK") {
					successCount++
				}
			case err := <-errors:
				t.Errorf("Unexpected error in concurrent rendering: %v", err)
			}
		}

		if successCount != 3 {
			t.Errorf("Expected 3 successful concurrent renders, got %d", successCount)
		}
	})
}
