package templateengine

import (
	"bytes"
	"html/template"
	"io"
	"strings"
	"testing"
)

// Test for TemplateConfig default values and validation
func TestTemplateConfig(t *testing.T) {
	t.Run("DefaultTemplateConfig", func(t *testing.T) {
		config := DefaultTemplateConfig()

		// Verify default values
		if config.LayoutContentName != "content" {
			t.Errorf("Expected LayoutContentName 'content', got '%s'", config.LayoutContentName)
		}

		if config.ErrorLayoutContentName != "content" {
			t.Errorf("Expected ErrorLayoutContentName 'content', got '%s'", config.ErrorLayoutContentName)
		}

		if config.MissingKeyBehavior != "zero" {
			t.Errorf("Expected MissingKeyBehavior 'zero', got '%s'", config.MissingKeyBehavior)
		}

		if len(config.Extensions) == 0 {
			t.Error("Expected default extensions to be set")
		}

		expectedExtensions := []string{".html", ".tmpl", ".gohtml"}
		if len(config.Extensions) != len(expectedExtensions) {
			t.Errorf("Expected %d extensions, got %d", len(expectedExtensions), len(config.Extensions))
		}
	})

	t.Run("ConfigBuilder", func(t *testing.T) {
		config := DefaultTemplateConfig().
			WithLayout("layout.html").
			WithContent("content.html").
			WithPublicDir("/public").
			WithDevMode(true)

		if config.LayoutPath != "layout.html" {
			t.Errorf("Expected LayoutPath 'layout.html', got '%s'", config.LayoutPath)
		}

		if config.ContentPath != "content.html" {
			t.Errorf("Expected ContentPath 'content.html', got '%s'", config.ContentPath)
		}

		if config.PublicDir != "/public" {
			t.Errorf("Expected PublicDir '/public', got '%s'", config.PublicDir)
		}

		if !config.DevMode {
			t.Error("Expected DevMode to be true")
		}
	})

	t.Run("ConfigValidation", func(t *testing.T) {
		config := DefaultTemplateConfig()
		err := config.Validate()
		if err != nil {
			t.Errorf("Expected valid config, got error: %v", err)
		}
	})

	t.Run("ConfigCheckers", func(t *testing.T) {
		emptyConfig := TemplateConfig{}
		if !emptyConfig.IsEmpty() {
			t.Error("Expected empty config to be detected as empty")
		}

		layoutConfig := DefaultTemplateConfig().WithLayout("layout.html")
		if !layoutConfig.HasLayout() {
			t.Error("Expected config with layout to be detected")
		}

		contentConfig := DefaultTemplateConfig().WithContent("content.html")
		if !contentConfig.HasContent() {
			t.Error("Expected config with content to be detected")
		}
	})

	t.Run("ConfigClone", func(t *testing.T) {
		original := DefaultTemplateConfig().
			WithPartials("partial1.html", "partial2.html").
			WithFuncMap(template.FuncMap{"test": func() string { return "test" }})

		clone := original.Clone()

		// Modify original
		original.Partials[0] = "modified.html"
		original.FuncMap["new"] = func() string { return "new" }

		// Verify clone is unaffected
		if clone.Partials[0] == "modified.html" {
			t.Error("Clone was affected by modification to original")
		}

		if _, exists := clone.FuncMap["new"]; exists {
			t.Error("Clone function map was affected by modification to original")
		}
	})
}

// Test for EventTemplateMap functionality
func TestEventTemplateMap(t *testing.T) {
	t.Run("EventTemplateMapCreation", func(t *testing.T) {
		eventMap := make(EventTemplateMap)
		eventMap["create:ok"] = EventTemplateState{
			"template1": struct{}{},
			"template2": struct{}{},
		}

		if len(eventMap) != 1 {
			t.Errorf("Expected 1 event, got %d", len(eventMap))
		}

		state, exists := eventMap["create:ok"]
		if !exists {
			t.Error("Expected event 'create:ok' to exist")
		}

		if len(state) != 2 {
			t.Errorf("Expected 2 templates for event state, got %d", len(state))
		}

		if _, exists := state["template1"]; !exists {
			t.Error("Expected template1 to exist in event state")
		}
	})
}

// Test for TemplateContext functionality
func TestTemplateContext(t *testing.T) {
	t.Run("TemplateContextCreation", func(t *testing.T) {
		ctx := TemplateContext{
			Errors: map[string]interface{}{
				"field1": "error message 1",
				"field2": "error message 2",
			},
			FuncMap: template.FuncMap{
				"test": func() string { return "test" },
			},
			Data: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		}

		if len(ctx.Errors) != 2 {
			t.Errorf("Expected 2 errors, got %d", len(ctx.Errors))
		}

		if len(ctx.FuncMap) != 1 {
			t.Errorf("Expected 1 function, got %d", len(ctx.FuncMap))
		}

		if len(ctx.Data) != 2 {
			t.Errorf("Expected 2 data items, got %d", len(ctx.Data))
		}
	})
}

// Mock implementations for testing interface compliance

// MockTemplate implements the Template interface for testing
type MockTemplate struct {
	name      string
	templates []Template
	funcMap   template.FuncMap
}

func NewMockTemplate(name string) *MockTemplate {
	return &MockTemplate{
		name:      name,
		templates: []Template{},
		funcMap:   make(template.FuncMap),
	}
}

func (m *MockTemplate) Execute(wr io.Writer, data interface{}) error {
	// Simple mock implementation
	_, err := wr.Write([]byte("mock template output"))
	return err
}

func (m *MockTemplate) ExecuteTemplate(wr io.Writer, name string, data interface{}) error {
	// Simple mock implementation
	_, err := wr.Write([]byte("mock template: " + name))
	return err
}

func (m *MockTemplate) Name() string {
	return m.name
}

func (m *MockTemplate) Templates() []Template {
	return m.templates
}

func (m *MockTemplate) Clone() (Template, error) {
	clone := &MockTemplate{
		name:      m.name,
		templates: make([]Template, len(m.templates)),
		funcMap:   make(template.FuncMap),
	}

	copy(clone.templates, m.templates)
	for k, v := range m.funcMap {
		clone.funcMap[k] = v
	}

	return clone, nil
}

func (m *MockTemplate) Funcs(funcMap template.FuncMap) Template {
	for k, v := range funcMap {
		m.funcMap[k] = v
	}
	return m
}

func (m *MockTemplate) Lookup(name string) Template {
	for _, tmpl := range m.templates {
		if tmpl.Name() == name {
			return tmpl
		}
	}
	return nil
}

// MockTemplateCache implements the TemplateCache interface for testing
type MockTemplateCache struct {
	cache map[string]Template
}

func NewMockTemplateCache() *MockTemplateCache {
	return &MockTemplateCache{
		cache: make(map[string]Template),
	}
}

func (m *MockTemplateCache) Set(key string, template Template) error {
	m.cache[key] = template
	return nil
}

func (m *MockTemplateCache) Get(key string) (Template, bool) {
	template, exists := m.cache[key]
	return template, exists
}

func (m *MockTemplateCache) Delete(key string) bool {
	if _, exists := m.cache[key]; exists {
		delete(m.cache, key)
		return true
	}
	return false
}

func (m *MockTemplateCache) Clear() {
	m.cache = make(map[string]Template)
}

func (m *MockTemplateCache) Size() int {
	return len(m.cache)
}

func (m *MockTemplateCache) Keys() []string {
	keys := make([]string, 0, len(m.cache))
	for k := range m.cache {
		keys = append(keys, k)
	}
	return keys
}

// Test template cache functionality
func TestTemplateCache(t *testing.T) {
	t.Run("MockTemplateCacheOperations", func(t *testing.T) {
		cache := NewMockTemplateCache()
		template := NewMockTemplate("test")

		// Test Set and Get
		err := cache.Set("key1", template)
		if err != nil {
			t.Errorf("Unexpected error setting template: %v", err)
		}

		retrieved, exists := cache.Get("key1")
		if !exists {
			t.Error("Expected template to exist in cache")
		}

		if retrieved.Name() != "test" {
			t.Errorf("Expected template name 'test', got '%s'", retrieved.Name())
		}

		// Test Size
		if cache.Size() != 1 {
			t.Errorf("Expected cache size 1, got %d", cache.Size())
		}

		// Test Keys
		keys := cache.Keys()
		if len(keys) != 1 || keys[0] != "key1" {
			t.Errorf("Expected keys ['key1'], got %v", keys)
		}

		// Test Delete
		deleted := cache.Delete("key1")
		if !deleted {
			t.Error("Expected template to be deleted")
		}

		if cache.Size() != 0 {
			t.Errorf("Expected cache size 0 after deletion, got %d", cache.Size())
		}

		// Test Clear
		cache.Set("key1", template)
		cache.Set("key2", template)
		cache.Clear()

		if cache.Size() != 0 {
			t.Errorf("Expected cache size 0 after clear, got %d", cache.Size())
		}
	})
}

// Test template functionality
func TestTemplate(t *testing.T) {
	t.Run("MockTemplateOperations", func(t *testing.T) {
		tmpl := NewMockTemplate("test-template")

		// Test Name
		if tmpl.Name() != "test-template" {
			t.Errorf("Expected name 'test-template', got '%s'", tmpl.Name())
		}

		// Test Execute
		var buf bytes.Buffer
		err := tmpl.Execute(&buf, nil)
		if err != nil {
			t.Errorf("Unexpected error executing template: %v", err)
		}

		if !strings.Contains(buf.String(), "mock template output") {
			t.Errorf("Expected template output to contain 'mock template output', got '%s'", buf.String())
		}

		// Test ExecuteTemplate
		buf.Reset()
		err = tmpl.ExecuteTemplate(&buf, "subtemplate", nil)
		if err != nil {
			t.Errorf("Unexpected error executing subtemplate: %v", err)
		}

		if !strings.Contains(buf.String(), "subtemplate") {
			t.Errorf("Expected template output to contain 'subtemplate', got '%s'", buf.String())
		}

		// Test Funcs
		funcMap := template.FuncMap{
			"testFunc": func() string { return "test" },
		}
		tmpl.Funcs(funcMap)

		if tmpl.funcMap["testFunc"] == nil {
			t.Error("Expected function to be added to template")
		}

		// Test Clone
		clone, err := tmpl.Clone()
		if err != nil {
			t.Errorf("Unexpected error cloning template: %v", err)
		}

		if clone.Name() != tmpl.Name() {
			t.Errorf("Expected clone name to match original, got '%s'", clone.Name())
		}
	})
}

// Test for default implementations

// Test DefaultFuncMapBuilder
func TestDefaultFuncMapBuilder(t *testing.T) {
	t.Run("DefaultFuncMapBuilder", func(t *testing.T) {
		builder := &DefaultFuncMapBuilder{}

		// Test with empty context
		ctx := TemplateContext{}
		funcMap := builder.BuildFuncMap(ctx)
		if funcMap == nil {
			t.Error("Expected non-nil function map")
		}
		if len(funcMap) != 0 {
			t.Errorf("Expected empty function map, got %d functions", len(funcMap))
		}

		// Test with existing function map
		testFuncMap := template.FuncMap{
			"test": func() string { return "test" },
		}
		ctx.FuncMap = testFuncMap
		funcMap = builder.BuildFuncMap(ctx)
		if len(funcMap) != 1 {
			t.Errorf("Expected 1 function, got %d", len(funcMap))
		}
		if funcMap["test"] == nil {
			t.Error("Expected test function to exist")
		}
	})
}

// Test FileTemplateLoader
func TestFileTemplateLoader(t *testing.T) {
	t.Run("FileTemplateLoader", func(t *testing.T) {
		loader := &FileTemplateLoader{}
		config := DefaultTemplateConfig() // Test LoadFromFile - should work with proper config
		_, err := loader.LoadFromFile("test.html", config)
		if err == nil {
			t.Error("Expected error for missing file")
		}

		// Test LoadFromString - should work
		tmpl, err := loader.LoadFromString("<h1>Test</h1>", config)
		if err != nil {
			t.Errorf("Expected no error for string template, got %v", err)
		}
		if tmpl == nil {
			t.Error("Expected non-nil template")
		}

		// Test LoadFromBytes - should work
		tmpl, err = loader.LoadFromBytes([]byte("<h1>Test</h1>"), config)
		if err != nil {
			t.Errorf("Expected no error for bytes template, got %v", err)
		}
		if tmpl == nil {
			t.Error("Expected non-nil template")
		}

		// Test LoadPartials - should return empty list for empty input
		templates, err := loader.LoadPartials([]string{}, config)
		if err != nil {
			t.Errorf("Expected no error for empty partials, got %v", err)
		}
		if len(templates) != 0 {
			t.Errorf("Expected empty templates slice, got %d templates", len(templates))
		}
	})
}

// Test DefaultTemplateValidator
func TestDefaultTemplateValidator(t *testing.T) {
	t.Run("DefaultTemplateValidator", func(t *testing.T) {
		validator := &DefaultTemplateValidator{}

		// Test ValidateConfig with valid config
		config := DefaultTemplateConfig()
		err := validator.ValidateConfig(config)
		if err != nil {
			t.Errorf("Expected no error for valid config, got %v", err)
		}

		// Test ValidateTemplate with valid template
		template := NewMockTemplate("test")
		err = validator.ValidateTemplate(template)
		if err != nil {
			t.Errorf("Expected no error for valid template, got %v", err)
		}

		// Test ValidateTemplate with nil template
		err = validator.ValidateTemplate(nil)
		if err != ErrInvalidTemplate {
			t.Errorf("Expected ErrInvalidTemplate for nil template, got %v", err)
		}

		// Test ValidateEventTemplates with valid event templates
		eventTemplates := make(EventTemplateMap)
		eventTemplates["test:ok"] = EventTemplateState{"template1": struct{}{}}
		err = validator.ValidateEventTemplates(eventTemplates)
		if err != nil {
			t.Errorf("Expected no error for valid event templates, got %v", err)
		}

		// Test ValidateEventTemplates with nil event templates
		err = validator.ValidateEventTemplates(nil)
		if err != ErrInvalidEventTemplate {
			t.Errorf("Expected ErrInvalidEventTemplate for nil event templates, got %v", err)
		}
	})
}

// Interface compliance tests for default implementations
func TestDefaultImplementationsCompliance(t *testing.T) {
	t.Run("DefaultFuncMapBuilderImplementsInterface", func(t *testing.T) {
		var _ FuncMapBuilder = &DefaultFuncMapBuilder{}
	})

	t.Run("FileTemplateLoaderImplementsInterface", func(t *testing.T) {
		var _ TemplateLoader = &FileTemplateLoader{}
	})

	t.Run("DefaultTemplateValidatorImplementsInterface", func(t *testing.T) {
		var _ TemplateValidator = &DefaultTemplateValidator{}
	})
}

// Test error constants
func TestErrorConstants(t *testing.T) {
	t.Run("ErrorConstants", func(t *testing.T) {
		if ErrNotImplemented.Error() != "not implemented" {
			t.Errorf("Expected 'not implemented', got '%s'", ErrNotImplemented.Error())
		}

		if ErrInvalidTemplate.Error() != "invalid template" {
			t.Errorf("Expected 'invalid template', got '%s'", ErrInvalidTemplate.Error())
		}

		if ErrInvalidEventTemplate.Error() != "invalid event template" {
			t.Errorf("Expected 'invalid event template', got '%s'", ErrInvalidEventTemplate.Error())
		}

		if ErrTemplateNotFound.Error() != "template not found" {
			t.Errorf("Expected 'template not found', got '%s'", ErrTemplateNotFound.Error())
		}

		if ErrTemplateParseFailed.Error() != "template parse failed" {
			t.Errorf("Expected 'template parse failed', got '%s'", ErrTemplateParseFailed.Error())
		}
	})
}

// Interface compliance tests
func TestInterfaceCompliance(t *testing.T) {
	t.Run("MockTemplateImplementsInterface", func(t *testing.T) {
		var _ Template = &MockTemplate{}
	})

	t.Run("MockTemplateCacheImplementsInterface", func(t *testing.T) {
		var _ TemplateCache = &MockTemplateCache{}
	})
}

// Test GoTemplateEngine basic functionality
func TestGoTemplateEngineBasic(t *testing.T) {
	t.Run("NewGoTemplateEngine", func(t *testing.T) {
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
	})

	t.Run("LoadDefaultTemplate", func(t *testing.T) {
		engine := NewGoTemplateEngine()
		config := DefaultTemplateConfig()

		tmpl, err := engine.LoadTemplate(config)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if tmpl == nil {
			t.Fatal("Expected non-nil template")
		}
	})

	t.Run("TemplateEngineInterfaceCompliance", func(t *testing.T) {
		var _ TemplateEngine = &GoTemplateEngine{}
	})
}
