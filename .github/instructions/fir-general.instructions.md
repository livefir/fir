---
applyTo: '**'
---

# Fir Library: General Development Guidelines

This document provides comprehensive guidelines for developing with the Fir toolkit, a Go library for building reactive web interfaces using Go, html/template, and Alpine.js.

## Core Philosophy

Fir is designed for Go developers with moderate HTML/CSS & JS skills who want to progressively build reactive web apps without mastering complex web frameworks. It emphasizes:

- **Server-rendered HTML**: Use Go's standard html/template library
- **Progressive enhancement**: Start with HTML forms, enhance with JavaScript
- **Web standards compliance**: Build on standard HTTP, WebSocket, and Alpine.js
- **Simplicity over complexity**: Avoid unnecessary abstractions

## Architecture Overview

### Dual-Mode Operation
Fir operates in two distinct modes that provide progressive enhancement:

1. **HTTP Mode (Fallback)**: Event submission via Ajax/fetch with DOM updates (no page reloads when using @submit.prevent)
2. **WebSocket Mode (Enhanced)**: Event submission via WebSocket with real-time DOM updates

Both modes use identical server-side code and `x-fir-*` attributes, ensuring consistent behavior regardless of client capabilities.

### Core Components
- **Controller System**: Main entry point that manages routes and provides configuration options
- **Route System**: Defines application logic using functional options (`OnLoad`, `OnEvent`)
- **Route Context**: Provides interface for handlers to interact with framework (`Data()`, `KV()`, `FieldError()`)
- **Template System**: Go html/template with custom processing for `x-fir-*` attributes
- **Actions System**: 11 different action handlers for declarative DOM updates
- **PubSub System**: Real-time communication between server and clients
- **WebSocket System**: Handles WebSocket connections and real-time event processing

### Data Flow Patterns
1. **Initial Load**: Browser Request → Controller → OnLoad → Template rendering → HTML response  
2. **HTTP Mode**: Form Submit → Ajax/fetch → OnEvent → State update → JSON response → DOM updates via Alpine.js
3. **WebSocket Mode**: Alpine.js → WebSocket → OnEvent → Event publishing → DOM actions

## Development Patterns

### Controller Setup
```go
func main() {
    // Development mode with built-in features
    controller := fir.NewController("app", fir.DevelopmentMode(true))
    
    // Production mode with custom options
    controller := fir.NewController("app",
        fir.WithPubsubAdapter(redisPubsubAdapter),
        fir.WithChannelFunc(authChannelFunc),
        fir.WithErrorHandler(customErrorHandler),
    )
    
    http.Handle("/", controller.RouteFunc(MyRoute))
}
```

### Route Structure
```go
func MyRoute() fir.RouteOptions {
    // State variables (persistent across requests)
    var data []Item
    var message string
    
    return fir.RouteOptions{
        fir.ID("my-route"),
        fir.Content("template.html"),
        
        // Initialize data on page load
        fir.OnLoad(func(ctx fir.RouteContext) error {
            return ctx.Data(map[string]interface{}{
                "items": data,
                "message": message,
            })
```

### Template Structure
```html
<!DOCTYPE html>
<html>
<head>
    <!-- Always include Alpine.js and Fir plugin -->
    {{ if fir.Development }}
        <script defer src="/cdn.js"></script>
    {{ else }}
        <script defer src="https://unpkg.com/@livefir/fir@latest/dist/fir.min.js"></script>
    {{ end }}
    <script defer src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"></script>
</head>
<body>
    <!-- Root Alpine.js component -->
    <div x-data>
        <!-- Progressive enhancement: HTTP forms with WebSocket enhancement -->
        <form method="post" @submit.prevent="$fir.submit()">
            <input name="item_name" placeholder="Enter name" />
            <button formaction="/?event=add-item" type="submit">Add</button>
        </form>
        
        <!-- Dynamic content with template blocks -->
        <div x-fir-append:item="add-item">
            {{ range .items }}
                {{ block "item" . }}
                    <div>{{ .Name }}</div>
                {{ end }}
            {{ end }}
        </div>
    </div>
</body>
</html>
```

## Best Practices

### State Management
- **Persistent State**: Use closure variables in route functions for persistent data
- **Initial Data**: Always use `ctx.Data()` in OnLoad, not multiple `ctx.KV()` calls
- **Event Responses**: Use `ctx.KV()` for individual data updates in event handlers
- **State Reset**: Avoid resetting persistent state on every page load

### Error Handling
```go
fir.OnEvent("create-item", func(ctx fir.RouteContext) error {
    name := ctx.Request().FormValue("name")
    if name == "" {
        return ctx.FieldError("name", errors.New("name is required"))
    }
    // Process successfully...
    return ctx.KV("items", items)
})
```

### Route Context Methods
- **`ctx.Data(map[string]interface{})`**: Used in OnLoad for initial page data
- **`ctx.KV(key, value)`**: Used in OnEvent for incremental updates
- **`ctx.FieldError(field, error)`**: Used for validation errors
- **`ctx.Bind(dst)`**: Used for form data binding
- **`ctx.Redirect(url)`**: Used for navigation
- **`ctx.Request()`**: Access to HTTP request
- **`ctx.Response()`**: Access to HTTP response writer

### Template Patterns
- **Use template blocks** for dynamic content that will be updated
- **Include error displays** with appropriate x-fir-refresh attributes
- **Provide fallback forms** for HTTP mode operation
- **Use semantic HTML** with proper form elements and actions

### Progressive Enhancement
1. **Start with HTTP forms**: Ensure functionality works without JavaScript
2. **Add Alpine.js handlers**: Enhance with `@submit.prevent="$fir.submit()"`
3. **Include x-fir-* attributes**: Define how DOM should update
4. **Test both modes**: Verify HTTP fallback and WebSocket enhancement

## Action System

### 11 Available Actions
```go
var registry = map[string]ActionHandler{
    "refresh":         &RefreshActionHandler{},
    "reset":           &ResetActionHandler{},
    "remove":          &RemoveActionHandler{},
    "remove-parent":   &RemoveParentActionHandler{},
    "append":          &AppendActionHandler{},
    "prepend":         &PrependActionHandler{},
    "toggle-disabled": &ToggleDisabledActionHandler{},
    "toggleClass":     &ToggleClassActionHandler{},
    "dispatch":        &DispatchActionHandler{},
    "runjs":           &TriggerActionHandler{},
    "js":              &ActionPrefixHandler{},
    "redirect":        &RedirectActionHandler{},
}
```

### Action Usage Examples
```html
<!-- Refresh content -->
<div x-fir-refresh="update:ok">Content to refresh</div>

<!-- Append new items -->
<ul x-fir-append:item="add-item">
    {{ range .items }}
        {{ block "item" . }}<li>{{ . }}</li>{{ end }}
    {{ end }}
</ul>

<!-- Remove elements -->
<button x-fir-remove="delete-item">Delete</button>

<!-- Toggle classes -->
<div x-fir-toggleClass:[highlight]="toggle-highlight">Toggle me</div>

<!-- Dispatch events -->
<button x-fir-dispatch:[custom-event]="dispatch-event">Dispatch</button>

<!-- Execute JavaScript -->
<button x-fir-runjs:myAction="execute-js">Run JS</button>
```

## Common Patterns

### CRUD Operations
```go
// Create
fir.OnEvent("create", func(ctx fir.RouteContext) error {
    item := extractFromForm(ctx.Request())
    if err := validate(item); err != nil {
        return ctx.FieldError("create.field", err)
    }
    items = append(items, item)
    return ctx.KV("items", items)
}),

// Update
fir.OnEvent("update", func(ctx fir.RouteContext) error {
    id := ctx.Request().FormValue("id")
    // Update logic...
    return ctx.KV("items", items)
}),

// Delete
fir.OnEvent("delete", func(ctx fir.RouteContext) error {
    id := ctx.Request().FormValue("id")
    // Delete logic...
    return ctx.KV("items", items)
}),
```

### Form Validation
```html
<form method="post" @submit.prevent="$fir.submit()">
    <input name="email" type="email" />
    <p x-fir-refresh="validate:error">{{ fir.Error "validate.email" }}</p>
    
    <button formaction="/?event=validate" type="submit">Submit</button>
</form>
```

### Real-time Updates
All events are automatically broadcast to connected WebSocket clients:

```go
fir.OnEvent("broadcast-update", func(ctx fir.RouteContext) error {
    // Update local state
    updateData()
    
    // Automatically broadcasts to other clients via PubSub
    return ctx.KV("items", items)
})
```
    return ctx.KV("items", items)
})
```

### Template Patterns
- **Use template blocks** for dynamic content that will be updated
- **Include error displays** with appropriate x-fir-refresh attributes
- **Provide fallback forms** for HTTP mode operation
- **Use semantic HTML** with proper form elements and actions

### Progressive Enhancement
1. **Start with HTTP forms**: Ensure functionality works without JavaScript
2. **Add Alpine.js handlers**: Enhance with `@submit.prevent="$fir.submit()"`
3. **Include x-fir-* attributes**: Define how DOM should update
4. **Test both modes**: Verify HTTP fallback and WebSocket enhancement

## Common Patterns

### CRUD Operations
```go
// Create
fir.OnEvent("create", func(ctx fir.RouteContext) error {
    item := extractFromForm(ctx.Request())
    if err := validate(item); err != nil {
        return ctx.FieldError("create.field", err)
    }
    items = append(items, item)
    return ctx.KV("items", items)
}),

// Update
fir.OnEvent("update", func(ctx fir.RouteContext) error {
    id := ctx.Request().FormValue("id")
    // Update logic...
    return ctx.KV("items", items)
}),

// Delete
fir.OnEvent("delete", func(ctx fir.RouteContext) error {
    id := ctx.Request().FormValue("id")
    // Delete logic...
    return ctx.KV("items", items)
}),
```

## Development Workflow

### Project Setup
1. **Initialize Go module**: `go mod init myproject`
2. **Add Fir dependency**: `go get github.com/livefir/fir`
3. **Create controller**: `controller := fir.NewController("app", fir.DevelopmentMode(true))`
4. **Setup routes**: `http.Handle("/", controller.RouteFunc(MyRoute))`
5. **Enable dev server**: Use `fir/internal/dev.SetupAlpinePluginServer()` for development

### Testing Strategy
- **Unit Tests**: Test individual event handlers and state management
- **Integration Tests**: Test complete request/response cycles
- **E2E Tests**: Test both HTTP and WebSocket modes using ChromeDP
- **Manual Testing**: Test with and without JavaScript enabled

### Debugging Tips
- **Enable Development Mode**: `fir.DevelopmentMode(true)` for detailed logging
- **Check WebSocket Status**: Use browser dev tools to monitor WebSocket connections
- **Monitor Console**: Watch for Fir events and Alpine.js errors
- **Inspect DOM**: Verify x-fir-* attributes are processed correctly
- **Test Fallbacks**: Disable JavaScript to test HTTP mode functionality

## Performance Considerations

### Server-Side
- **Minimize State Size**: Keep route state lightweight
- **Efficient Templates**: Use template blocks for partial updates
- **Connection Management**: Monitor WebSocket connection counts
- **Memory Usage**: Be aware of persistent state in route closures

### Client-Side
- **Bundle Size**: Fir plugin is lightweight, but monitor total JS payload
- **DOM Updates**: Actions are optimized for minimal DOM manipulation
- **WebSocket Efficiency**: Events are JSON-serialized for transmission
- **Progressive Loading**: Consider lazy-loading for complex interfaces

## Security Considerations

### Input Validation
- **Always validate** user input in event handlers
- **Use Go's html/template** for automatic XSS protection
- **Sanitize data** before storing or broadcasting
- **Implement CSRF protection** for sensitive operations

### Authentication & Authorization
```go
fir.OnEvent("protected-action", func(ctx fir.RouteContext) error {
    user := getUserFromContext(ctx.Request())
    if user == nil {
        return ctx.FieldError("auth", errors.New("authentication required"))
    }
    if !user.CanPerform("action") {
        return ctx.FieldError("auth", errors.New("insufficient permissions"))
    }
    // Proceed with action...
})
```

## Common Pitfalls

1. **Resetting State**: Don't reset persistent variables on every OnLoad
2. **Missing Fallbacks**: Always provide HTTP mode functionality
3. **Template Blocks**: Remember to define blocks for dynamic content
4. **Event Names**: Keep event names consistent between HTML and Go handlers
5. **Error Handling**: Always handle and display errors appropriately
6. **WebSocket Dependencies**: Don't assume WebSocket is always available
7. **State Races**: Be careful with concurrent access to shared state
8. **Memory Leaks**: Monitor persistent state for memory growth

## Migration and Upgrades

- **Backward Compatibility**: Follow semantic versioning for breaking changes
- **Incremental Adoption**: Fir can be gradually introduced to existing projects
- **Template Migration**: Existing html/template code works with minimal changes
- **Alpine.js Integration**: Existing Alpine.js code is compatible

## Resources

- **Documentation**: `/Users/adnaan/code/livefir/fir/README.md`
- **Examples**: `/Users/adnaan/code/livefir/fir/examples/`
- **Alpine.js Plugin API**: `/Users/adnaan/code/livefir/fir/alpinejs-plugin/api.md`
- **Action Behavior Guide**: `.github/instructions/fir-actions-behavior.instructions.md`
