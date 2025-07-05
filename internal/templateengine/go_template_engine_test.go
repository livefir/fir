package templateengine

import (
	"bytes"
	"html/template"
	"strings"
	"testing"
)

func TestNewGoTemplateEngine(t *testing.T) {
	engine := NewGoTemplateEngine()

	if engine == nil {
		t.Fatal("Expected non-nil engine")
	}

	if engine.cache == nil {
		t.Error("Expected cache to be initialized")
	}

	if !engine.enableCache {
		t.Error("Expected cache to be enabled by default")
	}
}

func TestGoTemplateEngineLoadTemplate(t *testing.T) {
	t.Run("DefaultTemplate", func(t *testing.T) {
		engine := NewGoTemplateEngine()
		config := DefaultTemplateConfig()

		tmpl, err := engine.LoadTemplate(config)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if tmpl == nil {
			t.Fatal("Expected non-nil template")
		}

		// Test rendering
		var buf bytes.Buffer
		err = tmpl.Execute(&buf, nil)
		if err != nil {
			t.Errorf("Unexpected error executing template: %v", err)
		}

		if !strings.Contains(buf.String(), "default page") {
			t.Errorf("Expected default page content, got: %s", buf.String())
		}
	})

	t.Run("InlineContentTemplate", func(t *testing.T) {
		engine := NewGoTemplateEngine()
		config := DefaultTemplateConfig()
		config.ContentTemplate = "<h1>Hello, {{.Name}}!</h1>"

		tmpl, err := engine.LoadTemplate(config)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Test rendering with data
		var buf bytes.Buffer
		data := map[string]interface{}{"Name": "World"}
		err = tmpl.ExecuteTemplate(&buf, "content", data)
		if err != nil {
			t.Errorf("Unexpected error executing template: %v", err)
		}

		if !strings.Contains(buf.String(), "Hello, World!") {
			t.Errorf("Expected rendered content, got: %s", buf.String())
		}
	})
}

func TestGoTemplateEngineRender(t *testing.T) {
	t.Run("BasicRender", func(t *testing.T) {
		engine := NewGoTemplateEngine()
		config := DefaultTemplateConfig()
		config.ContentTemplate = "<h1>Hello, {{.Name}}!</h1>"

		tmpl, err := engine.LoadTemplate(config)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		var buf bytes.Buffer
		data := map[string]interface{}{"Name": "World"}
		err = engine.Render(tmpl, data, &buf)
		if err != nil {
			t.Errorf("Unexpected error rendering: %v", err)
		}
	})

	t.Run("RenderWithContext", func(t *testing.T) {
		engine := NewGoTemplateEngine()
		config := DefaultTemplateConfig()
		config.ContentTemplate = "<h1>{{upper .Name}}</h1>"
		// Set the function map in the config so the template is parsed with it
		config.FuncMap = template.FuncMap{
			"upper": strings.ToUpper,
		}

		tmpl, err := engine.LoadTemplate(config)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Create context with function map
		ctx := TemplateContext{
			FuncMap: template.FuncMap{
				"upper": strings.ToUpper,
			},
		}

		var buf bytes.Buffer
		data := map[string]interface{}{"Name": "world"}
		err = engine.RenderWithContext(tmpl, ctx, data, &buf)
		if err != nil {
			t.Errorf("Unexpected error rendering: %v", err)
		}

		if !strings.Contains(buf.String(), "WORLD") {
			t.Errorf("Expected uppercase content, got: %s", buf.String())
		}
	})

	t.Run("RenderNilTemplate", func(t *testing.T) {
		engine := NewGoTemplateEngine()

		var buf bytes.Buffer
		err := engine.Render(nil, nil, &buf)
		if err != ErrInvalidTemplate {
			t.Errorf("Expected ErrInvalidTemplate, got: %v", err)
		}
	})
}

func TestGoTemplateEngineCache(t *testing.T) {
	t.Run("CacheOperations", func(t *testing.T) {
		engine := NewGoTemplateEngine()

		// Create a mock template
		tmpl := NewGoTemplate(template.Must(template.New("test").Parse("test")))

		// Test cache operations
		engine.CacheTemplate("test-key", tmpl)

		cached, found := engine.GetCachedTemplate("test-key")
		if !found {
			t.Error("Expected template to be found in cache")
		}

		if cached != tmpl {
			t.Error("Expected same template from cache")
		}

		engine.ClearCache()

		_, found = engine.GetCachedTemplate("test-key")
		if found {
			t.Error("Expected template not to be found after cache clear")
		}
	})
}

func TestGoTemplateEngineErrorTemplate(t *testing.T) {
	t.Run("DefaultErrorTemplate", func(t *testing.T) {
		engine := NewGoTemplateEngine()
		config := DefaultTemplateConfig()

		tmpl, err := engine.LoadErrorTemplate(config)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if tmpl == nil {
			t.Fatal("Expected non-nil template")
		}

		// Test rendering
		var buf bytes.Buffer
		err = tmpl.Execute(&buf, nil)
		if err != nil {
			t.Errorf("Unexpected error executing template: %v", err)
		}

		if !strings.Contains(buf.String(), "default page") {
			t.Errorf("Expected default page content, got: %s", buf.String())
		}
	})
}

func TestGoTemplateEngineEventTemplates(t *testing.T) {
	t.Run("ExtractEventTemplates", func(t *testing.T) {
		engine := NewGoTemplateEngine()
		tmpl := NewGoTemplate(template.Must(template.New("test").Parse("test")))

		eventTemplates, err := engine.ExtractEventTemplates(tmpl)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if eventTemplates == nil {
			t.Error("Expected non-nil event templates map")
		}
	})

	t.Run("RenderEventTemplate", func(t *testing.T) {
		engine := NewGoTemplateEngine()
		tmpl := NewGoTemplate(template.Must(template.New("test").Parse("test")))

		result, err := engine.RenderEventTemplate(tmpl, "test-event", "ok", nil)
		if err != ErrEventNotFound {
			t.Errorf("Expected ErrEventNotFound, got: %v", err)
		}

		if result != "" {
			t.Errorf("Expected empty result, got: %s", result)
		}
	})
}
