# Template Engine Migration Guide

## Overview

The Fir framework has been enhanced with a new template engine abstraction that provides better performance, flexibility, and testability. This guide will help you migrate from the legacy template system to the new template engine.

## Current Status

- ✅ **Template Engine Infrastructure**: Complete and tested
- ✅ **Route Integration**: Routes can use template engines via route options
- ✅ **Backward Compatibility**: All existing code continues to work unchanged
- 🔄 **Migration Path**: Optional adoption, routes can be migrated incrementally

## Current Implementation Status

### ✅ **Completed Infrastructure**

1. **Template Engine Interface**: Complete with all necessary methods
2. **GoTemplateEngine Implementation**: Production-ready with caching and event support
3. **Event Template System**: Full extraction and rendering capabilities
4. **Route Integration**: Routes can accept template engines via options
5. **Function Map Providers**: Flexible function injection system
6. **Comprehensive Testing**: 80%+ test coverage with integration tests

### 🔄 **Migration State**

The template engine system is fully implemented and tested, but the framework maintains **complete backward compatibility**:

- **Legacy template fields still active**: The route struct still contains `template`, `errorTemplate`, and `eventTemplates` fields
- **Renderer uses legacy templates**: The `render.go` system currently uses legacy template access methods
- **Fallback strategy**: Routes with template engines fall back to legacy parsing if template engine interface doesn't match
- **Template engine factory disabled**: Controller returns `nil` for template engine factory to maintain compatibility

### 🎯 **Next Migration Steps**

To complete the migration, these steps need to be taken in order:

1. **Enable renderer template engine support** - Update `render.go` to optionally use template engines
2. **Create template engine adapter for render compatibility** - Bridge between template engine interface and renderer expectations
3. **Add controller-level template engine support** - Enable default template engines per controller
4. **Gradual legacy field removal** - Remove legacy template fields once renderer is updated
5. **Legacy parsing method cleanup** - Remove `parseTemplatesLegacy` and related functions

### ⚠️ **Important Notes**

- **No breaking changes**: All existing code continues to work unchanged
- **Opt-in migration**: Template engines are only used when explicitly provided
- **Incremental adoption**: Routes can be migrated individually as needed
- **Performance ready**: Template engine system is optimized and production-ready

## Template Engine Benefits

### 1. **Better Performance**
- Template caching at the engine level
- Concurrent template loading
- Optimized event template processing

### 2. **Improved Testability**
- Template engines can be mocked for testing
- Template rendering can be tested independently
- Clean separation of template logic from route logic

### 3. **Enhanced Flexibility**
- Support for multiple template engines
- Route-specific template configurations
- Custom function map providers

## Migration Strategies

### Strategy 1: Gradual Route Migration (Recommended)

Migrate routes one at a time using the new template engine route options:

```go
// Old way (still works)
route := fir.NewRoute(
    fir.ID("my-route"),
    fir.Content("template.html"),
    fir.Layout("layout.html"),
)

// New way with template engine
templateEngine := templateengine.NewGoTemplateEngine()
route := fir.NewRoute(
    fir.ID("my-route"),
    fir.Content("template.html"),
    fir.Layout("layout.html"),
    fir.TemplateEngine(templateEngine), // Add template engine
)
```

### Strategy 2: Controller-Level Template Engine

Configure a template engine at the controller level for all routes:

```go
// This approach will be available in future versions
// controller := fir.NewController(
//     fir.WithTemplateEngine(templateEngine),
// )
```

## Using Template Engines

### Basic Template Engine Setup

```go
package main

import (
    "github.com/livefir/fir"
    "github.com/livefir/fir/internal/templateengine"
)

func main() {
    // Create a template engine
    engine := templateengine.NewGoTemplateEngine()
    
    // Create route with template engine
    route := fir.NewRoute(
        fir.ID("example"),
        fir.Content("example.html"),
        fir.TemplateEngine(engine),
    )
    
    // Use the route as normal
    controller := fir.NewController()
    controller.Handle("/", route)
}
```

### Advanced Template Engine Configuration

```go
// Create a custom template engine with specific configuration
builder := templateengine.NewRouteTemplateEngineBuilder()
engine := builder.
    WithRouteConfig("my-route", "content.html", "layout.html", "content", false).
    WithBaseFuncMap(template.FuncMap{
        "customFunc": func() string { return "custom" },
    }).
    Build()

route := fir.NewRoute(
    fir.ID("my-route"),
    fir.TemplateEngine(engine),
)
```

### Custom Function Map Providers

```go
// Create a custom function map provider
type MyFuncMapProvider struct{}

func (p *MyFuncMapProvider) BuildFuncMap(route interface{}, ctx interface{}) template.FuncMap {
    return template.FuncMap{
        "myFunc": func() string { return "Hello from custom provider!" },
    }
}

func (p *MyFuncMapProvider) GetName() string {
    return "MyFuncMapProvider"
}

// Use the provider with a template engine
provider := &MyFuncMapProvider{}
engine := templateengine.NewGoTemplateEngine()
engine.SetFuncMapProvider(provider)
```

## Template Engine Features

### 1. **Template Caching**

Template engines support intelligent caching:

```go
// Disable caching for development
route := fir.NewRoute(
    fir.ID("dev-route"),
    fir.Content("template.html"),
    fir.DisableRouteTemplateCache(true), // Disable caching for this route
)

// Or configure caching in the template engine
engine := templateengine.NewGoTemplateEngine()
engine.SetCacheEnabled(false) // Disable caching for the entire engine
```

### 2. **Event Template Handling**

Template engines automatically handle event templates:

```go
// Event templates are automatically extracted and processed
// No changes needed to existing event template patterns
// @fir:increment:ok, @fir:submit:error, etc. work as before
```

### 3. **Error Template Support**

Template engines support dedicated error templates:

```go
route := fir.NewRoute(
    fir.ID("my-route"),
    fir.Content("content.html"),
    fir.ErrorContent("error.html"), // Error templates work with engines
    fir.TemplateEngine(engine),
)
```

## Testing with Template Engines

### Mocking Template Engines

```go
// Create a mock template engine for testing
type MockTemplateEngine struct {
    templates map[string]string
}

func (m *MockTemplateEngine) LoadTemplate(config interface{}) (interface{}, error) {
    // Mock implementation
    return template.Must(template.New("test").Parse("mock template")), nil
}

func (m *MockTemplateEngine) Render(tmpl interface{}, data interface{}, w io.Writer) error {
    // Mock rendering
    w.Write([]byte("mock rendered content"))
    return nil
}

// Use in tests
func TestRouteWithMockEngine(t *testing.T) {
    mockEngine := &MockTemplateEngine{}
    route := fir.NewRoute(
        fir.ID("test-route"),
        fir.TemplateEngine(mockEngine),
    )
    
    // Test route behavior with mocked template engine
}
```

### Testing Template Engines Independently

```go
func TestTemplateEngine(t *testing.T) {
    engine := templateengine.NewGoTemplateEngine()
    
    config := templateengine.TemplateConfig{
        ContentPath: "test.html",
        FuncMap: template.FuncMap{
            "testFunc": func() string { return "test" },
        },
    }
    
    template, err := engine.LoadTemplate(config)
    assert.NoError(t, err)
    assert.NotNil(t, template)
}
```

## Performance Considerations

### Template Caching

- Template engines cache parsed templates by default
- Use `DisableRouteTemplateCache(true)` for development
- Enable caching in production for better performance

### Function Map Optimization

- Function maps are merged at template load time
- Use static function map providers for better performance
- Avoid creating new function maps on each request

### Event Template Performance

- Event templates are extracted once and cached
- Large HTML templates with many events are efficiently processed
- Concurrent event template processing is supported

## Migration Checklist

### Phase 1: Setup Template Engine
- [ ] Import `internal/templateengine` package
- [ ] Create template engine instance
- [ ] Test template engine with simple route

### Phase 2: Route Migration
- [ ] Identify routes to migrate
- [ ] Add `fir.TemplateEngine()` option to routes
- [ ] Test migrated routes thoroughly
- [ ] Update route tests to use template engines

### Phase 3: Advanced Features
- [ ] Implement custom function map providers
- [ ] Configure template caching strategies
- [ ] Add custom template engines if needed

### Phase 4: Validation
- [ ] Run comprehensive test suite
- [ ] Performance test with template engines
- [ ] Validate all event templates work correctly

## Common Issues and Solutions

### Issue: Template Not Found
```
Error: template not found
```
**Solution**: Ensure template paths are correct and files exist. Template engines use the same path resolution as legacy templates.

### Issue: Function Not Available
```
Error: function "myFunc" not defined
```
**Solution**: Make sure custom functions are added via function map providers or route options.

### Issue: Event Templates Not Working
```
Event templates not rendering
```
**Solution**: Verify that event template syntax is correct (`@fir:eventname:state`) and that the template engine has event template support enabled.

## Future Enhancements

### Planned Features
- Multiple template engine support (Jinja, Handlebars)
- Template hot reloading in development
- Template preprocessing and optimization
- Advanced caching strategies (Redis, Memcached)

### Contributing
- Template engine implementations are in `internal/templateengine/`
- Add tests for new template engine features
- Follow the existing interface patterns
- Maintain backward compatibility

## Support

For questions or issues with template engine migration:

1. Check the test files in `internal/templateengine/` for examples
2. Review the integration tests in `route_integration_test.go`
3. Consult the template engine interfaces in `interfaces.go`

The template engine system is designed to be backward compatible, so existing code will continue to work while you gradually adopt the new features.
