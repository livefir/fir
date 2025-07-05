# Template Engine Package

This package provides a template engine abstraction layer for the Fir framework, decoupling routing logic from template parsing and rendering concerns.

## Overview

The template engine package introduces clean interfaces and configurations that separate template management from routing, making the codebase more maintainable, testable, and extensible.

## Architecture

### Core Interfaces

#### TemplateEngine
The main interface for template operations:
- **Template Loading**: `LoadTemplate()`, `LoadErrorTemplate()`
- **Rendering**: `Render()`, `RenderWithContext()`
- **Event Templates**: `ExtractEventTemplates()`, `RenderEventTemplate()`
- **Caching**: `CacheTemplate()`, `GetCachedTemplate()`, `ClearCache()`

#### Template
Wraps template functionality with a consistent interface:
- **Execution**: `Execute()`, `ExecuteTemplate()`
- **Metadata**: `Name()`, `Templates()`
- **Manipulation**: `Clone()`, `Funcs()`, `Lookup()`

#### TemplateCache
Provides caching functionality for templates:
- **Cache Operations**: `Set()`, `Get()`, `Delete()`, `Clear()`
- **Management**: `Size()`, `Keys()`

### Configuration

#### TemplateConfig
Centralizes all template configuration:
```go
config := DefaultTemplateConfig().
    WithLayout("layout.html").
    WithContent("content.html").
    WithPublicDir("/templates").
    WithDevMode(true)
```

**Configuration Options:**
- **Template Paths**: Layout, content, error templates
- **File System**: Public directory, embedded FS support
- **Function Maps**: Custom template functions
- **Caching**: Enable/disable template caching
- **Development**: Debug mode, file watching

### Event Template Support

#### EventTemplateMap
Maps event IDs to their template states:
```go
type EventTemplateMap map[string]EventTemplateState
type EventTemplateState map[string]struct{}
```

**Example:**
```go
eventMap := EventTemplateMap{
    "create:ok": EventTemplateState{
        "success-template": struct{}{},
        "notification-template": struct{}{},
    },
    "create:error": EventTemplateState{
        "error-template": struct{}{},
    },
}
```

## Usage Examples

### Basic Template Configuration

```go
// Create a basic template configuration
config := DefaultTemplateConfig().
    WithLayout("layouts/main.html").
    WithContent("pages/home.html").
    WithPublicDir("./templates")

// Validate configuration
if err := config.Validate(); err != nil {
    log.Fatal("Invalid template config:", err)
}
```

### Template Engine Interface Implementation

```go
// Example of implementing the TemplateEngine interface
type MyTemplateEngine struct {
    cache TemplateCache
}

func (e *MyTemplateEngine) LoadTemplate(config TemplateConfig) (Template, error) {
    // Implementation for loading templates
    return nil, nil
}

func (e *MyTemplateEngine) Render(template Template, data interface{}, w io.Writer) error {
    // Implementation for rendering templates
    return template.Execute(w, data)
}
```

### Template Context Usage

```go
// Create template context for rendering
ctx := TemplateContext{
    RouteContext: routeCtx,
    Errors: map[string]interface{}{
        "username": "Username is required",
        "email": "Invalid email format",
    },
    FuncMap: template.FuncMap{
        "formatDate": func(t time.Time) string {
            return t.Format("2006-01-02")
        },
    },
    Data: map[string]interface{}{
        "user": currentUser,
        "settings": userSettings,
    },
}
```

## Default Implementations

### GoTemplateEngine
Default implementation using Go's `html/template` package:
- Maintains backward compatibility with existing Fir templates
- Supports all Go template features
- Includes event template extraction and rendering

### InMemoryTemplateCache
Simple in-memory cache implementation:
- Thread-safe operations
- LRU eviction (when size limits are implemented)
- Development-friendly cache clearing

### FileTemplateLoader
File-based template loading:
- Supports multiple file extensions
- Embedded file system support
- Partial template loading

## Migration Strategy

This package is designed to be introduced gradually:

1. **Phase 1** (Current): Interface definitions and configuration structures
2. **Phase 2**: Default implementations with backward compatibility
3. **Phase 3**: Route integration with fallback to current implementation
4. **Phase 4**: Full migration and deprecation of old template handling

## Benefits

### Separation of Concerns
- Template parsing logic is separate from routing logic
- Each component has a single responsibility
- Easier to reason about and maintain

### Testability
- Mock implementations for unit testing
- Template logic can be tested independently
- Error scenarios can be simulated easily

### Extensibility
- New template engines can be added without changing routes
- Custom caching strategies can be implemented
- Template processing can be customized per use case

### Performance
- Template caching reduces parsing overhead
- Concurrent template loading and rendering
- Memory-efficient template management

## Testing

The package includes comprehensive tests:
- **Unit tests** for all interfaces and implementations
- **Mock implementations** for testing template engine consumers
- **Configuration validation** tests
- **Interface compliance** verification

Run tests:
```bash
go test ./internal/templateengine/...
```

## Future Enhancements

- **Template Hot Reloading**: Automatic template reloading in development
- **Template Metrics**: Performance monitoring and statistics
- **Advanced Caching**: LRU, TTL, and memory-aware caching strategies
- **Template Validation**: Compile-time template validation and optimization
- **Multiple Template Engines**: Support for alternative template engines (e.g., Mustache, Handlebars)

## References

- [Go html/template documentation](https://pkg.go.dev/html/template)
- [Fir Framework Architecture](../../ARCHITECTURE.md)
- [Template Engine Decoupling Strategy](../../TEMPLATE_ENGINE_DECOUPLING_STRATEGY.md)
