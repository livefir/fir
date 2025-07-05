# RouteServices Architecture Documentation

## Overview

The `RouteServices` package provides a clean abstraction layer that encapsulates all the services and dependencies that routes need from the controller. This design enables better separation of concerns, improved testability, and more maintainable code.

## Architecture

### Core Components

#### RouteServices Struct
```go
type RouteServices struct {
    EventRegistry   event.EventRegistry
    PubSub         pubsub.Adapter
    Renderer       interface{}
    ChannelFunc    func(*http.Request, string) *string
    PathParamsFunc func(*http.Request) map[string]string
    Options        *Options
    controller     interface{} // Temporary for WebSocket handling
}
```

### Key Features

1. **Service Encapsulation**: All route dependencies are centralized in one struct
2. **Clean Separation**: Routes no longer directly reference the controller
3. **Testability**: Services can be mocked and tested independently
4. **Configuration Management**: Centralized options and settings
5. **Validation**: Built-in service validation for error detection

## Usage Patterns

### Creating RouteServices

```go
// Create RouteServices with all dependencies
services := routeservices.NewRouteServices(
    eventRegistry,
    pubsubAdapter,
    renderer,
    options,
)

// Configure additional functions
services.SetChannelFunc(channelFunc)
services.SetPathParamsFunc(pathParamsFunc)

// Validate all services are properly configured
if err := services.ValidateServices(); err != nil {
    log.Fatalf("Service validation failed: %v", err)
}
```

### Using with Routes

Routes access services through the `RouteServices` instead of directly referencing the controller:

```go
// Old pattern (deprecated)
route.cntrl.EventRegistry.Register(...)

// New pattern
route.services.EventRegistry.Register(...)
```

### Factory Pattern Integration

The controller uses a factory pattern for route creation:

```go
// Controller creates routes using RouteServices
route := c.createRouteHandler(pattern, handler, options)
```

## Testing

### Unit Testing RouteServices

```go
func TestRouteServices(t *testing.T) {
    services := routeservices.NewRouteServices(
        mockEventRegistry,
        mockPubSub,
        mockRenderer,
        testOptions,
    )
    
    if err := services.ValidateServices(); err != nil {
        t.Errorf("Validation failed: %v", err)
    }
}
```

### Integration Testing

```go
func TestRouteWithServices(t *testing.T) {
    services := createTestRouteServices()
    
    // Test route functionality using services
    route := newRoute(pattern, handler, services)
    
    // Validate route behavior
    // ...
}
```

### Mocking for Tests

```go
type mockRenderer struct{}
func (m *mockRenderer) RenderRoute(ctx, data interface{}, isError bool) {}
func (m *mockRenderer) RenderDOMEvents(ctx, event interface{}) interface{} {
    return nil
}

// Use in tests
services := routeservices.NewRouteServices(
    event.NewEventRegistry(),
    pubsub.NewInmem(),
    &mockRenderer{},
    testOptions,
)
```

## Configuration Options

### Core Options
- `AppName`: Application identifier
- `DevelopmentMode`: Enable development features
- `DebugLog`: Enable debug logging
- `PublicDir`: Static file directory

### WebSocket Options
- `DisableWebsocket`: Disable WebSocket functionality
- `WebsocketUpgrader`: WebSocket connection upgrader
- `OnSocketConnect/Disconnect`: Connection lifecycle callbacks

### Performance Options
- `DisableTemplateCache`: Disable template caching
- `DropDuplicateInterval`: Event deduplication timing
- `Cache`: In-memory cache instance

### File System Options
- `ReadFile`: File reading function
- `ExistFile`: File existence checking function
- `WatchExts`: File extensions to watch for changes

## Benefits

### Improved Testability
- Services can be mocked independently
- Routes can be tested without full controller setup
- Clear dependency injection patterns

### Better Separation of Concerns
- Routes focus on request handling logic
- Controller focuses on application orchestration
- Services encapsulate shared functionality

### Enhanced Maintainability
- Changes to service interfaces are centralized
- Dependencies are explicit and documented
- Easier to reason about code relationships

### Performance Benefits
- Service instances are reused across routes
- No unnecessary controller references
- Optimized for memory usage

## Migration Guide

### From Direct Controller References

**Before:**
```go
route.cntrl.EventRegistry.Register(routeID, eventID, handler)
channel := route.cntrl.channelFunc(req, routeID)
```

**After:**
```go
route.services.EventRegistry.Register(routeID, eventID, handler)
channel := route.services.ChannelFunc(req, routeID)
```

### Testing Migration

**Before:**
```go
// Had to create full controller for testing
controller := NewController(...)
route := controller.Route(pattern, handler)
```

**After:**
```go
// Can test with just the needed services
services := routeservices.NewRouteServices(mockRegistry, mockPubSub, mockRenderer, options)
route := newRoute(pattern, handler, services)
```

## Future Enhancements

### Planned Improvements
1. Remove temporary controller reference when WebSocket handling is refactored
2. Add middleware support through RouteServices
3. Enhance validation with custom validation rules
4. Add service discovery and dependency injection features

### Extension Points
- Custom service providers
- Pluggable middleware system
- Service lifecycle management
- Advanced configuration validation

## Best Practices

### Service Creation
- Always validate services after creation
- Use dependency injection for service creation
- Keep service configuration immutable when possible

### Testing
- Mock services at the appropriate level of abstraction
- Use integration tests for end-to-end validation
- Test service validation logic thoroughly

### Performance
- Reuse RouteServices instances across routes
- Clone services only when necessary for isolation
- Monitor service memory usage in production

## Troubleshooting

### Common Issues

#### Service Validation Failures
```go
if err := services.ValidateServices(); err != nil {
    // Check that all required services are set
    // Ensure EventRegistry, PubSub, Renderer, and Options are non-nil
}
```

#### Function Configuration
```go
// Ensure channel and path param functions are set
services.SetChannelFunc(channelFunc)
services.SetPathParamsFunc(pathParamsFunc)
```

#### Testing Issues
```go
// Use appropriate mocks for each service type
mockRenderer := &MockRenderer{}
services := routeservices.NewRouteServices(eventRegistry, pubsub, mockRenderer, options)
```

## Related Documentation

- [Route Factory Pattern](./ROUTE_FACTORY.md)
- [Controller Architecture](./CONTROLLER_ARCHITECTURE.md)
- [Testing Guide](./TESTING_GUIDE.md)
- [Migration Guide](./MIGRATION_GUIDE.md)
