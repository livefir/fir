package templateengine

import (
	"bytes"
	"html/template"
	"strings"
	"testing"
)

func TestRouteFuncMapBuilder(t *testing.T) {
	t.Run("BuildFuncMap", func(t *testing.T) {
		builder := NewRouteFuncMapBuilder()

		ctx := FuncMapContext{
			URLPath:         "/test/path",
			AppName:         "testapp",
			DevelopmentMode: true,
			Errors: map[string]interface{}{
				"field1": "error message",
			},
		}

		funcMap := builder.BuildFuncMap(ctx)

		// Check that base 'fir' function exists
		if _, exists := funcMap["fir"]; !exists {
			t.Error("Expected 'fir' function to exist from base provider")
		}

		// Check that route-specific function exists
		routeInfoFunc, exists := funcMap["routeInfo"]
		if !exists {
			t.Error("Expected 'routeInfo' function to exist")
		}

		// Test the routeInfo function
		routeInfo := routeInfoFunc.(func() map[string]interface{})()
		if routeInfo["path"] != "/test/path" {
			t.Errorf("Expected path '/test/path', got '%v'", routeInfo["path"])
		}

		if routeInfo["app"] != "testapp" {
			t.Errorf("Expected app 'testapp', got '%v'", routeInfo["app"])
		}

		if routeInfo["development"] != true {
			t.Errorf("Expected development true, got '%v'", routeInfo["development"])
		}
	})

	t.Run("SetBaseProvider", func(t *testing.T) {
		builder := NewRouteFuncMapBuilder()

		customProvider := &mockFuncMapProvider{
			funcs: template.FuncMap{
				"custom": func() string { return "custom function" },
			},
			name: "CustomProvider",
		}

		builder.SetBaseProvider(customProvider)

		ctx := FuncMapContext{}
		funcMap := builder.BuildFuncMap(ctx)

		// Should have custom function from base provider
		if _, exists := funcMap["custom"]; !exists {
			t.Error("Expected custom function from base provider")
		}

		// Should still have route-specific function
		if _, exists := funcMap["routeInfo"]; !exists {
			t.Error("Expected routeInfo function to still exist")
		}
	})
}

func TestTemplateEngineAdapter(t *testing.T) {
	t.Run("ConvertRouteContextToTemplateContext", func(t *testing.T) {
		engine := NewGoTemplateEngine()
		adapter := NewTemplateEngineAdapter(engine)

		routeContext := map[string]interface{}{"test": "data"}
		errors := map[string]interface{}{"field": "error"}

		templateCtx := adapter.ConvertRouteContextToTemplateContext(
			routeContext, errors, "/test", "myapp", true,
		)

		if templateCtx.RouteContext == nil {
			t.Error("Expected route context to be preserved")
		}

		if templateCtx.Errors == nil {
			t.Error("Expected errors to be preserved")
		}

		if templateCtx.Data["URLPath"] != "/test" {
			t.Errorf("Expected URLPath '/test', got '%v'", templateCtx.Data["URLPath"])
		}

		if templateCtx.Data["AppName"] != "myapp" {
			t.Errorf("Expected AppName 'myapp', got '%v'", templateCtx.Data["AppName"])
		}

		if templateCtx.Data["DevelopmentMode"] != true {
			t.Errorf("Expected DevelopmentMode true, got '%v'", templateCtx.Data["DevelopmentMode"])
		}
	})

	t.Run("LoadTemplateForRoute", func(t *testing.T) {
		engine := NewGoTemplateEngine()
		adapter := NewTemplateEngineAdapter(engine)

		config := DefaultTemplateConfig()
		config.ContentTemplate = "Hello {{.Name}}, app: {{fir.Name}}"

		tmpl, err := adapter.LoadTemplateForRoute(
			config,
			nil, // routeContext
			nil, // errors
			"/test",
			"myapp",
			true, // developmentMode
		)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if tmpl == nil {
			t.Fatal("Expected non-nil template")
		}

		// Test that the template can be executed with the fir function
		var buf bytes.Buffer
		data := map[string]interface{}{"Name": "world"}
		err = tmpl.Execute(&buf, data)
		if err != nil {
			t.Errorf("Unexpected error executing template: %v", err)
		}

		result := buf.String()
		if !strings.Contains(result, "Hello world") {
			t.Errorf("Expected 'Hello world' in result, got: %s", result)
		}

		if !strings.Contains(result, "myapp") {
			t.Errorf("Expected 'myapp' in result, got: %s", result)
		}
	})

	t.Run("LoadErrorTemplateForRoute", func(t *testing.T) {
		engine := NewGoTemplateEngine()
		adapter := NewTemplateEngineAdapter(engine)

		config := DefaultTemplateConfig()
		config.ErrorContentTemplate = "Error: {{fir.Name}}"

		tmpl, err := adapter.LoadErrorTemplateForRoute(
			config,
			nil, // routeContext
			nil, // errors
			"/error",
			"errorapp",
			false, // developmentMode
		)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if tmpl == nil {
			t.Fatal("Expected non-nil template")
		}

		// Test that the error template can be executed
		var buf bytes.Buffer
		err = tmpl.Execute(&buf, nil)
		if err != nil {
			t.Errorf("Unexpected error executing template: %v", err)
		}

		result := buf.String()
		if !strings.Contains(result, "errorapp") {
			t.Errorf("Expected 'errorapp' in result, got: %s", result)
		}
	})

	t.Run("GetEngine", func(t *testing.T) {
		engine := NewGoTemplateEngine()
		adapter := NewTemplateEngineAdapter(engine)

		retrieved := adapter.GetEngine()
		if retrieved != engine {
			t.Error("Expected to retrieve the same engine instance")
		}
	})

	t.Run("SetFuncMapProvider", func(t *testing.T) {
		engine := NewGoTemplateEngine()
		adapter := NewTemplateEngineAdapter(engine)

		customProvider := &mockFuncMapProvider{
			funcs: template.FuncMap{
				"custom": func() string { return "custom" },
			},
			name: "CustomProvider",
		}

		adapter.SetFuncMapProvider(customProvider)

		retrieved := adapter.GetFuncMapProvider()
		if retrieved != customProvider {
			t.Error("Expected custom provider to be set")
		}
	})

	t.Run("NewTemplateEngineAdapterWithProvider", func(t *testing.T) {
		engine := NewGoTemplateEngine()
		customProvider := &mockFuncMapProvider{
			funcs: template.FuncMap{
				"custom": func() string { return "custom" },
			},
			name: "CustomProvider",
		}

		adapter := NewTemplateEngineAdapterWithProvider(engine, customProvider)

		if adapter.GetEngine() != engine {
			t.Error("Expected engine to be set correctly")
		}

		if adapter.GetFuncMapProvider() != customProvider {
			t.Error("Expected custom provider to be set")
		}
	})

	t.Run("ContextAwareEngineIntegration", func(t *testing.T) {
		// Create an engine that supports context-aware loading
		engine := NewGoTemplateEngine()
		adapter := NewTemplateEngineAdapter(engine)

		config := DefaultTemplateConfig()
		config.ContentTemplate = "Route: {{routeInfo.path}}, App: {{routeInfo.app}}"

		// Set a custom provider that includes route-specific functions
		adapter.SetFuncMapProvider(NewRouteFuncMapBuilder())

		// Load template through the adapter
		tmpl, err := adapter.LoadTemplateForRoute(
			config,
			nil, // routeContext
			nil, // errors
			"/test/route",
			"testapp",
			true, // developmentMode
		)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Test template execution
		var buf bytes.Buffer
		err = tmpl.Execute(&buf, nil)
		if err != nil {
			t.Errorf("Unexpected error executing template: %v", err)
		}

		result := buf.String()
		if !strings.Contains(result, "/test/route") {
			t.Errorf("Expected route path in result, got: %s", result)
		}

		if !strings.Contains(result, "testapp") {
			t.Errorf("Expected app name in result, got: %s", result)
		}
	})
}
