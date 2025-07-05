package templateengine

import (
	"bytes"
	"html/template"
	"strings"
	"testing"
)

func TestDefaultFuncMapProvider(t *testing.T) {
	t.Run("BuildFuncMap", func(t *testing.T) {
		provider := NewDefaultFuncMapProvider()

		ctx := FuncMapContext{
			URLPath:         "/test",
			AppName:         "testapp",
			DevelopmentMode: true,
			Errors: map[string]interface{}{
				"field1": "error message",
			},
		}

		funcMap := provider.BuildFuncMap(ctx)

		// Check that 'fir' function exists
		firFunc, exists := funcMap["fir"]
		if !exists {
			t.Fatal("Expected 'fir' function to exist in function map")
		}

		// Test the fir function
		firContext := firFunc.(func() *RouteDOMContext)()
		if firContext == nil {
			t.Fatal("Expected non-nil RouteDOMContext")
		}

		if firContext.URLPath != "/test" {
			t.Errorf("Expected URLPath '/test', got '%s'", firContext.URLPath)
		}

		if firContext.Name != "testapp" {
			t.Errorf("Expected Name 'testapp', got '%s'", firContext.Name)
		}

		if !firContext.Development {
			t.Error("Expected Development mode to be true")
		}
	})

	t.Run("GetName", func(t *testing.T) {
		provider := NewDefaultFuncMapProvider()
		name := provider.GetName()

		if name == "" {
			t.Error("Expected non-empty provider name")
		}
	})
}

func TestCompositeFuncMapProvider(t *testing.T) {
	t.Run("CombineProviders", func(t *testing.T) {
		// Create two providers with different functions
		provider1 := &mockFuncMapProvider{
			funcs: template.FuncMap{
				"func1":  func() string { return "from provider 1" },
				"shared": func() string { return "from provider 1" },
			},
			name: "Provider1",
		}

		provider2 := &mockFuncMapProvider{
			funcs: template.FuncMap{
				"func2":  func() string { return "from provider 2" },
				"shared": func() string { return "from provider 2" }, // Should override provider1
			},
			name: "Provider2",
		}

		composite := NewCompositeFuncMapProvider("Composite", provider1, provider2)

		ctx := FuncMapContext{}
		funcMap := composite.BuildFuncMap(ctx)

		// Check that both functions exist
		if _, exists := funcMap["func1"]; !exists {
			t.Error("Expected func1 to exist")
		}

		if _, exists := funcMap["func2"]; !exists {
			t.Error("Expected func2 to exist")
		}

		// Check that provider2 overrode the shared function
		sharedFunc := funcMap["shared"].(func() string)
		result := sharedFunc()
		if result != "from provider 2" {
			t.Errorf("Expected shared function from provider 2, got '%s'", result)
		}
	})

	t.Run("AddProvider", func(t *testing.T) {
		composite := NewCompositeFuncMapProvider("Test")

		provider := &mockFuncMapProvider{
			funcs: template.FuncMap{
				"testfunc": func() string { return "test" },
			},
			name: "TestProvider",
		}

		composite.AddProvider(provider)

		ctx := FuncMapContext{}
		funcMap := composite.BuildFuncMap(ctx)

		if _, exists := funcMap["testfunc"]; !exists {
			t.Error("Expected testfunc to exist after adding provider")
		}
	})
}

func TestFuncMapRegistry(t *testing.T) {
	t.Run("RegisterAndGet", func(t *testing.T) {
		registry := NewFuncMapRegistry()

		provider := &mockFuncMapProvider{
			funcs: template.FuncMap{
				"custom": func() string { return "custom function" },
			},
			name: "CustomProvider",
		}

		registry.Register("custom", provider)

		retrieved := registry.Get("custom")
		if retrieved != provider {
			t.Error("Expected to retrieve the same provider instance")
		}

		// Test non-existent provider returns default
		defaultProvider := registry.Get("nonexistent")
		if defaultProvider == nil {
			t.Error("Expected default provider for non-existent name")
		}
	})

	t.Run("SetDefault", func(t *testing.T) {
		registry := NewFuncMapRegistry()

		customDefault := &mockFuncMapProvider{
			funcs: template.FuncMap{
				"default": func() string { return "custom default" },
			},
			name: "CustomDefault",
		}

		registry.SetDefault(customDefault)

		retrieved := registry.GetDefault()
		if retrieved != customDefault {
			t.Error("Expected custom default provider")
		}
	})

	t.Run("List", func(t *testing.T) {
		registry := NewFuncMapRegistry()

		registry.Register("provider1", &mockFuncMapProvider{name: "Provider1"})
		registry.Register("provider2", &mockFuncMapProvider{name: "Provider2"})

		names := registry.List()
		if len(names) != 2 {
			t.Errorf("Expected 2 provider names, got %d", len(names))
		}

		// Check that both names are present
		found1, found2 := false, false
		for _, name := range names {
			if name == "provider1" {
				found1 = true
			}
			if name == "provider2" {
				found2 = true
			}
		}

		if !found1 || !found2 {
			t.Error("Expected both provider names to be in the list")
		}
	})
}

func TestGoTemplateEngineWithFuncMapProvider(t *testing.T) {
	t.Run("LoadTemplateWithContext", func(t *testing.T) {
		// Create a custom function map provider
		customProvider := &mockFuncMapProvider{
			funcs: template.FuncMap{
				"upper": strings.ToUpper,
				"fir": func() *RouteDOMContext {
					return &RouteDOMContext{
						Name:        "testapp",
						URLPath:     "/custom",
						Development: true,
					}
				},
			},
			name: "CustomProvider",
		}

		engine := NewGoTemplateEngineWithProvider(customProvider)

		config := DefaultTemplateConfig()
		config.ContentTemplate = "Hello {{upper .Name}}, app: {{fir.Name}}"

		ctx := TemplateContext{
			Data: map[string]interface{}{
				"URLPath":         "/custom",
				"AppName":         "testapp",
				"DevelopmentMode": true,
			},
		}

		tmpl, err := engine.LoadTemplateWithContext(config, ctx)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		var buf bytes.Buffer
		data := map[string]interface{}{"Name": "world"}
		err = tmpl.Execute(&buf, data)
		if err != nil {
			t.Errorf("Unexpected error executing template: %v", err)
		}

		result := buf.String()
		if !strings.Contains(result, "WORLD") {
			t.Errorf("Expected uppercase 'WORLD' in result, got: %s", result)
		}

		if !strings.Contains(result, "testapp") {
			t.Errorf("Expected 'testapp' in result, got: %s", result)
		}
	})

	t.Run("SetFuncMapProvider", func(t *testing.T) {
		engine := NewGoTemplateEngine()

		customProvider := &mockFuncMapProvider{
			funcs: template.FuncMap{
				"custom": func() string { return "custom function" },
			},
			name: "CustomProvider",
		}

		engine.SetFuncMapProvider(customProvider)

		retrieved := engine.GetFuncMapProvider()
		if retrieved != customProvider {
			t.Error("Expected custom provider to be set")
		}
	})

	t.Run("FuncMapMerging", func(t *testing.T) {
		// Test that config function maps override provider function maps
		provider := &mockFuncMapProvider{
			funcs: template.FuncMap{
				"shared":        func() string { return "from provider" },
				"provider_only": func() string { return "provider function" },
			},
			name: "TestProvider",
		}

		engine := NewGoTemplateEngineWithProvider(provider)

		config := DefaultTemplateConfig()
		config.ContentTemplate = "{{shared}} {{provider_only}}"
		config.FuncMap = template.FuncMap{
			"shared": func() string { return "from config" },
		}

		ctx := TemplateContext{}

		tmpl, err := engine.LoadTemplateWithContext(config, ctx)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		var buf bytes.Buffer
		err = tmpl.Execute(&buf, nil)
		if err != nil {
			t.Errorf("Unexpected error executing template: %v", err)
		}

		result := buf.String()
		if !strings.Contains(result, "from config") {
			t.Errorf("Expected config function to override provider, got: %s", result)
		}

		if !strings.Contains(result, "provider function") {
			t.Errorf("Expected provider function to be available, got: %s", result)
		}
	})
}

func TestRouteDOMContext(t *testing.T) {
	t.Run("NewRouteDOMContext", func(t *testing.T) {
		ctx := FuncMapContext{
			URLPath:         "/test/path",
			AppName:         "myapp",
			DevelopmentMode: false,
			Errors: map[string]interface{}{
				"field1": "error message",
			},
		}

		domCtx := NewRouteDOMContext(ctx)

		if domCtx.URLPath != "/test/path" {
			t.Errorf("Expected URLPath '/test/path', got '%s'", domCtx.URLPath)
		}

		if domCtx.Name != "myapp" {
			t.Errorf("Expected Name 'myapp', got '%s'", domCtx.Name)
		}

		if domCtx.Development {
			t.Error("Expected Development to be false")
		}

		if domCtx.errors == nil {
			t.Error("Expected errors to be initialized")
		}
	})

	t.Run("ActiveRoute", func(t *testing.T) {
		domCtx := &RouteDOMContext{URLPath: "/current"}

		// Test active route
		result := domCtx.ActiveRoute("/current", "active")
		if result != "active" {
			t.Errorf("Expected 'active', got '%s'", result)
		}

		// Test inactive route
		result = domCtx.ActiveRoute("/other", "active")
		if result != "" {
			t.Errorf("Expected empty string, got '%s'", result)
		}
	})

	t.Run("NotActiveRoute", func(t *testing.T) {
		domCtx := &RouteDOMContext{URLPath: "/current"}

		// Test non-active route
		result := domCtx.NotActiveRoute("/other", "inactive")
		if result != "inactive" {
			t.Errorf("Expected 'inactive', got '%s'", result)
		}

		// Test active route
		result = domCtx.NotActiveRoute("/current", "inactive")
		if result != "" {
			t.Errorf("Expected empty string, got '%s'", result)
		}
	})

	t.Run("Error", func(t *testing.T) {
		domCtx := &RouteDOMContext{
			errors: map[string]interface{}{
				"field1": "simple error",
				"nested": map[string]interface{}{
					"field2": "nested error",
				},
			},
		}

		// Test simple error lookup
		result := domCtx.Error("field1")
		if result != "simple error" {
			t.Errorf("Expected 'simple error', got '%v'", result)
		}

		// Test nested error lookup
		result = domCtx.Error("nested", "field2")
		if result != "nested error" {
			t.Errorf("Expected 'nested error', got '%v'", result)
		}

		// Test non-existent error
		result = domCtx.Error("nonexistent")
		if result != nil {
			t.Errorf("Expected nil for non-existent error, got '%v'", result)
		}
	})
}

// Mock function map provider for testing
type mockFuncMapProvider struct {
	funcs template.FuncMap
	name  string
}

func (m *mockFuncMapProvider) BuildFuncMap(ctx FuncMapContext) template.FuncMap {
	return m.funcs
}

func (m *mockFuncMapProvider) GetName() string {
	return m.name
}
