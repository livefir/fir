# Fir Framework API Documentation

## Overview

This document provides comprehensive API documentation for the Fir framework's modern handler chain architecture. The framework uses an extensible handler system with service-based dependencies and priority-based routing.

## Core Interfaces

### Handler Interface

The base interface for all request handlers in the chain:

```go
type Handler interface {
    // CanHandle returns true if this handler can process the request
    CanHandle(req *http.Request, route RouteInterface) bool
    
    // Handle processes the request and returns true if handled successfully
    Handle(w http.ResponseWriter, req *http.Request, route RouteInterface) bool
    
    // Priority returns the handler's priority (lower values = higher priority)
    Priority() int
}
```

#### Implementation Guidelines

1. **CanHandle Method**: Should be fast and deterministic
   - Check HTTP method, content-type, headers
   - Validate route capabilities (handlers, templates)
   - Avoid heavy computations or I/O

2. **Handle Method**: Perform actual request processing
   - Use injected services for business logic
   - Return `true` only if request was fully processed
   - Handle errors gracefully with appropriate HTTP status

3. **Priority Method**: Return consistent priority value
   - Lower numbers = higher priority
   - Use standard priorities (see Handler Priority System)

### RouteInterface

Extended interface providing access to route configuration and services:

```go
type RouteInterface interface {
    // Core route information
    ID() string
    Path() string
    Method() string
    
    // Handler configuration
    GetOnEventHandlers() map[string]EventHandler
    GetOnLoadHandler() *OnLoadHandler
    
    // Service access
    GetServices() *routeservices.RouteServices
    
    // Session management
    GetCookieName() string
    GetSecureCookie(req *http.Request, name string) (string, error)
    
    // Template configuration
    GetTemplate() string
    GetTemplateData() any
}
```

## Standard Handlers

### WebSocketHandler

Handles WebSocket upgrade requests.

**Priority**: 5 (Highest)

**Usage**:
```go
handler := &WebSocketHandler{}

// Check if request can be handled
if handler.CanHandle(req, route) {
    success := handler.Handle(w, req, route)
}
```

**Requirements**:
- HTTP method: GET
- Headers: `Connection: Upgrade`, `Upgrade: websocket`
- Route must have WebSocket support enabled

**Dependencies**: Connection management services

### JSONEventHandler

Processes JSON-based event submissions.

**Priority**: 10

**Usage**:
```go
handler := &JSONEventHandler{}

// Example request processing
if handler.CanHandle(req, route) {
    success := handler.Handle(w, req, route)
}
```

**Requirements**:
- HTTP method: POST
- Content-Type: `application/json`
- Route must have OnEvent handlers configured
- Valid JSON payload with event information

**Dependencies**:
- `EventService`: For event processing
- `ResponseBuilder`: For response generation

**Request Format**:
```json
{
    "id": "button-click",
    "event": "click",
    "data": {
        "key": "value"
    }
}
```

### FormHandler

Handles form-encoded POST submissions.

**Priority**: 20

**Usage**:
```go
handler := &FormHandler{}

// Process form submission
if handler.CanHandle(req, route) {
    success := handler.Handle(w, req, route)
}
```

**Requirements**:
- HTTP method: POST
- Content-Type: `application/x-www-form-urlencoded`
- Route must have OnEvent handlers configured
- Valid form data with event fields

**Dependencies**:
- `EventService`: For event processing
- `ResponseBuilder`: For response generation

**Form Format**:
```html
<form method="POST">
    <input name="id" value="form-submit">
    <input name="event" value="submit">
    <input name="data" value="form-data">
    <button type="submit">Submit</button>
</form>
```

### GetHandler

Processes GET requests with onLoad event handlers.

**Priority**: 50 (Lowest)

**Usage**:
```go
handler := &GetHandler{}

// Handle GET request with onLoad
if handler.CanHandle(req, route) {
    success := handler.Handle(w, req, route)
}
```

**Requirements**:
- HTTP method: GET
- Route must have OnLoad handler configured
- Template rendering capability

**Dependencies**:
- `EventService`: For onLoad event processing
- `RenderService`: For DOM rendering
- `TemplateService`: For template processing
- `ResponseBuilder`: For response generation

## Service Interfaces

### EventService

Core service for processing route events:

```go
type EventService interface {
    // ProcessEvent handles route events with context
    ProcessEvent(ctx context.Context, route RouteInterface, event *Event) (*EventResponse, error)
}
```

**Usage Example**:
```go
// Process an event
response, err := eventService.ProcessEvent(ctx, route, event)
if err != nil {
    // Handle error
    return err
}

// Use response data
if response.Redirect != nil {
    // Handle redirect
    http.Redirect(w, req, response.Redirect.URL, response.Redirect.Code)
}
```

### RenderService

Handles DOM rendering and template processing:

```go
type RenderService interface {
    // RenderRoute renders the complete route with template and events
    RenderRoute(ctx context.Context, route RouteInterface, data any) (*RenderResult, error)
}
```

**Usage Example**:
```go
// Render route with data
result, err := renderService.RenderRoute(ctx, route, templateData)
if err != nil {
    // Handle rendering error
    return err
}

// Access rendered content
html := result.HTML
events := result.Events
```

### TemplateService

Template processing and caching:

```go
type TemplateService interface {
    // ExecuteTemplate renders a template with data
    ExecuteTemplate(templateName string, data any) (string, error)
    
    // GetTemplate retrieves a cached template
    GetTemplate(name string) (*template.Template, error)
}
```

### ResponseBuilder

HTTP response construction:

```go
type ResponseBuilder interface {
    // BuildResponse creates an HTTP response from event result
    BuildResponse(w http.ResponseWriter, req *http.Request, result *EventResponse) error
    
    // BuildErrorResponse creates an error response
    BuildErrorResponse(w http.ResponseWriter, err error, statusCode int) error
}
```

## Configuration

### RouteServices Setup

Complete service configuration for a route:

```go
services := &routeservices.RouteServices{
    EventService:    &eventservice.DefaultEventService{},
    RenderService:   &renderservice.DefaultRenderService{},
    TemplateService: &templateservice.GoTemplateService{},
    ResponseBuilder: &responsebuilder.DefaultResponseBuilder{},
    Options: &routeservices.Options{
        DisableTemplateCache: false,
        DisableWebsocket:     false,
    },
}

route := &Route{
    // ... route configuration
    services: services,
}
```

### Handler Chain Setup

Configuring the handler chain with custom handlers:

```go
// Create handler chain
chain := []Handler{
    &WebSocketHandler{},           // Priority 5
    &JSONEventHandler{},          // Priority 10
    &FormHandler{},               // Priority 20
    &GetHandler{},                // Priority 50
    &CustomHandler{},             // Custom priority
}

// Sort by priority (automatic in framework)
sort.Slice(chain, func(i, j int) bool {
    return chain[i].Priority() < chain[j].Priority()
})
```

## Testing

### Handler Testing

Testing individual handlers:

```go
func TestJSONEventHandler(t *testing.T) {
    handler := &JSONEventHandler{}
    route := createTestRoute()
    req := createJSONRequest()
    
    // Test CanHandle
    if !handler.CanHandle(req, route) {
        t.Error("Handler should be able to handle JSON request")
    }
    
    // Test Handle
    w := httptest.NewRecorder()
    success := handler.Handle(w, req, route)
    
    if !success {
        t.Error("Handler should successfully process request")
    }
    
    // Verify response
    if w.Code != http.StatusOK {
        t.Errorf("Expected status 200, got %d", w.Code)
    }
}
```

### Mock Services

Using mock services for testing:

```go
// Create mock services
services := &routeservices.RouteServices{
    EventService: &MockEventService{
        ProcessEventFunc: func(ctx context.Context, route RouteInterface, event *Event) (*EventResponse, error) {
            return &EventResponse{Success: true}, nil
        },
    },
    RenderService:   &MockRenderService{},
    TemplateService: &MockTemplateService{},
    ResponseBuilder: &MockResponseBuilder{},
}

// Use in tests
route := createTestRouteWithServices(services)
```

## Error Handling

### Handler Error Patterns

Standard error handling in handlers:

```go
func (h *JSONEventHandler) Handle(w http.ResponseWriter, req *http.Request, route RouteInterface) bool {
    // Parse request
    event, err := h.parseJSONEvent(req)
    if err != nil {
        http.Error(w, "Invalid JSON event", http.StatusBadRequest)
        return true // We handled the error
    }
    
    // Process event
    response, err := route.GetServices().EventService.ProcessEvent(req.Context(), route, event)
    if err != nil {
        http.Error(w, "Event processing failed", http.StatusInternalServerError)
        return true // We handled the error
    }
    
    // Build response
    err = route.GetServices().ResponseBuilder.BuildResponse(w, req, response)
    if err != nil {
        http.Error(w, "Response building failed", http.StatusInternalServerError)
        return true // We handled the error
    }
    
    return true // Success
}
```

### Service Error Handling

Error handling in service implementations:

```go
func (s *EventService) ProcessEvent(ctx context.Context, route RouteInterface, event *Event) (*EventResponse, error) {
    // Validate event
    if err := s.validateEvent(event); err != nil {
        return nil, fmt.Errorf("event validation failed: %w", err)
    }
    
    // Find handler
    handler, exists := route.GetOnEventHandlers()[event.ID]
    if !exists {
        return nil, fmt.Errorf("no handler found for event: %s", event.ID)
    }
    
    // Process event
    result, err := handler.Handle(ctx, event)
    if err != nil {
        return nil, fmt.Errorf("event handler failed: %w", err)
    }
    
    return result, nil
}
```

## Best Practices

### Handler Implementation

1. **Keep CanHandle Fast**: Minimize computation in `CanHandle()` methods
2. **Single Responsibility**: Each handler should handle one request type
3. **Error Handling**: Always handle errors gracefully with appropriate HTTP status codes
4. **Service Dependencies**: Use injected services, don't create dependencies in handlers
5. **Logging**: Add appropriate logging for debugging and monitoring

### Service Implementation

1. **Context Usage**: Always respect context cancellation and timeouts
2. **Error Wrapping**: Use `fmt.Errorf` with `%w` to wrap errors for better debugging
3. **Interface Compliance**: Ensure services implement required interfaces completely
4. **Stateless Design**: Services should be stateless when possible
5. **Testing**: Provide mock implementations for testing

### Route Configuration

1. **Service Injection**: Always provide complete service configuration
2. **Handler Registration**: Register all required event handlers
3. **Template Setup**: Ensure templates are available and cached appropriately
4. **Error Boundaries**: Define clear error handling strategies for each route

## Migration Guidelines

### From Legacy to Modern Handlers

1. **Identify Handler Type**: Determine which modern handler replaces legacy code
2. **Extract Services**: Move business logic to appropriate services
3. **Update Tests**: Replace legacy tests with handler-specific tests
4. **Verify Coverage**: Ensure all request types are covered by modern handlers

### Service Migration

1. **Interface First**: Define service interfaces before implementation
2. **Mock Early**: Create mock implementations for testing
3. **Gradual Migration**: Migrate one service at a time
4. **Validation**: Verify service behavior matches legacy implementation

For more detailed examples and migration patterns, see `MIGRATION_GUIDE.md`.
