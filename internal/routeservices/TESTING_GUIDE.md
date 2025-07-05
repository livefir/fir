# Testing Guide for RouteServices Architecture

## Overview

This guide demonstrates how to test routes and controllers using the RouteServices architecture, which provides better separation of concerns and improved testability. This guide is part of the `internal/routeservices` package documentation.

## Testing RouteServices

### Basic Unit Tests

```go
package myapp_test

import (
    "testing"
    "github.com/livefir/fir/internal/routeservices"
    "github.com/livefir/fir/internal/event"
    "github.com/livefir/fir/pubsub"
)

func TestRouteServicesCreation(t *testing.T) {
    // Create mock dependencies
    eventRegistry := event.NewEventRegistry()
    pubsubAdapter := pubsub.NewInmem()
    renderer := &MockRenderer{}
    options := &routeservices.Options{
        AppName: "test-app",
        DevelopmentMode: true,
    }

    // Create RouteServices
    services := routeservices.NewRouteServices(
        eventRegistry,
        pubsubAdapter,
        renderer,
        options,
    )

    // Validate services
    if err := services.ValidateServices(); err != nil {
        t.Fatalf("Service validation failed: %v", err)
    }

    // Test service access
    if services.EventRegistry == nil {
        t.Error("EventRegistry should be available")
    }
}
```

### Mock Implementations

```go
// MockRenderer for testing
type MockRenderer struct {
    RenderCount int
    LastData    interface{}
}

func (m *MockRenderer) RenderRoute(ctx interface{}, data interface{}, isError bool) {
    m.RenderCount++
    m.LastData = data
}

func (m *MockRenderer) RenderDOMEvents(ctx interface{}, event interface{}) interface{} {
    return map[string]interface{}{
        "processed": true,
        "event":     event,
    }
}
```

## Testing Routes with RouteServices

### Independent Route Testing

```go
func TestRouteHandling(t *testing.T) {
    // Create test services
    services := createTestRouteServices()
    
    // Configure services for testing
    services.SetChannelFunc(func(r *http.Request, routeID string) *string {
        channel := "test-channel"
        return &channel
    })
    
    services.SetPathParamsFunc(func(r *http.Request) map[string]string {
        return map[string]string{"id": "123"}
    })

    // Create route with services (using internal API for testing)
    route := newRoute("/test/{id}", testHandler, services)

    // Test route functionality
    req := httptest.NewRequest("GET", "/test/123", nil)
    // ... perform route testing
}

func testHandler(ctx RouteContext) {
    // Test handler implementation
}
```

### Integration Testing with Controller

```go
func TestControllerRouteIntegration(t *testing.T) {
    // Create controller
    controller := NewController()
    
    // Create route through controller (uses RouteServices internally)
    route := controller.Route("/api/users", userHandler)
    
    // Test that route was created with proper services
    // Access through controller's RouteServices if needed
    services := controller.GetRouteServices()
    
    if err := services.ValidateServices(); err != nil {
        t.Errorf("Controller services should be valid: %v", err)
    }
}
```

## Testing Event Handling

### Event Registry Testing

```go
func TestEventRegistration(t *testing.T) {
    services := createTestRouteServices()
    
    routeID := "test-route"
    eventID := "click"
    handler := func(data interface{}) {
        // Event handler logic
    }

    // Register event
    err := services.EventRegistry.Register(routeID, eventID, handler)
    if err != nil {
        t.Fatalf("Failed to register event: %v", err)
    }

    // Verify registration
    retrievedHandler, found := services.EventRegistry.Get(routeID, eventID)
    if !found {
        t.Error("Event handler should be registered")
    }
    if retrievedHandler == nil {
        t.Error("Retrieved handler should not be nil")
    }
}
```

### PubSub Testing

```go
func TestPubSubIntegration(t *testing.T) {
    services := createTestRouteServices()
    
    ctx := context.Background()
    channel := "test-channel"
    
    // Subscribe to events
    subscription, err := services.PubSub.Subscribe(ctx, channel)
    if err != nil {
        t.Fatalf("Failed to subscribe: %v", err)
    }
    defer subscription.Close()

    // Publish event
    event := pubsub.Event{
        ID: func() *string { s := "test-id"; return &s }(),
        State: eventstate.Type("test-state"),
    }
    
    err = services.PubSub.Publish(ctx, channel, event)
    if err != nil {
        t.Fatalf("Failed to publish: %v", err)
    }

    // Verify event received
    select {
    case receivedEvent := <-subscription.C():
        if *receivedEvent.ID != *event.ID {
            t.Error("Event ID should match")
        }
    case <-time.After(time.Second):
        t.Error("Event not received")
    }
}
```

## Testing Utilities

### Helper Functions

```go
func createTestRouteServices() *routeservices.RouteServices {
    eventRegistry := event.NewEventRegistry()
    pubsubAdapter := pubsub.NewInmem()
    renderer := &MockRenderer{}
    
    options := &routeservices.Options{
        AppName:         "test-app",
        DevelopmentMode: true,
        DebugLog:        true,
    }

    services := routeservices.NewRouteServices(
        eventRegistry,
        pubsubAdapter,
        renderer,
        options,
    )

    return services
}

func createTestController() *Controller {
    controller := NewController()
    // Configure controller for testing
    return controller
}
```

### Test Configuration

```go
func createTestOptions() *routeservices.Options {
    return &routeservices.Options{
        AppName:               "test-app",
        DisableTemplateCache:  true,
        DisableWebsocket:      true, // Disable for simpler testing
        DevelopmentMode:       true,
        DebugLog:              true,
        ReadFile: func(filename string) (string, []byte, error) {
            return filename, []byte("test content"), nil
        },
        ExistFile: func(filename string) bool {
            return true
        },
    }
}
```

## Best Practices

### Test Organization

1. **Unit Tests**: Test RouteServices in isolation
2. **Integration Tests**: Test routes with services
3. **End-to-End Tests**: Test full request/response cycle

### Mocking Strategy

1. **Mock at Service Level**: Create mock implementations of services
2. **Use Real Services**: When testing service interactions
3. **Isolate Dependencies**: Mock external dependencies only

### Testing Patterns

```go
// Pattern: Test service validation
func TestServiceValidation(t *testing.T) {
    tests := []struct {
        name        string
        setupFunc   func() *routeservices.RouteServices
        expectError bool
    }{
        // Test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            services := tt.setupFunc()
            err := services.ValidateServices()
            // Validate results
        })
    }
}
```

### Performance Testing

```go
func BenchmarkRouteServiceCreation(b *testing.B) {
    eventRegistry := event.NewEventRegistry()
    pubsubAdapter := pubsub.NewInmem()
    renderer := &MockRenderer{}
    options := createTestOptions()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = routeservices.NewRouteServices(
            eventRegistry,
            pubsubAdapter,
            renderer,
            options,
        )
    }
}
```

## Troubleshooting Tests

### Common Issues

1. **Service Validation Failures**: Ensure all required services are provided
2. **Mock Setup**: Verify mock implementations match expected interfaces
3. **Context Handling**: Use appropriate context for PubSub operations
4. **Event Registration**: Use correct routeID and eventID parameters

### Debug Techniques

```go
// Enable debug logging in tests
options := &routeservices.Options{
    DebugLog: true,
    // other options
}

// Validate services step by step
if services.EventRegistry == nil {
    t.Error("EventRegistry not set")
}
if services.PubSub == nil {
    t.Error("PubSub not set")
}
// etc.
```

This testing approach provides confidence in the route/controller decoupling while maintaining the ability to test components in isolation and integration.

## Example Code

For complete working examples of testing RouteServices, see the test files in this package:

- `services_test.go` - Comprehensive unit tests for RouteServices
- `integration_test.go` - Integration tests with HTTP, PubSub, and EventRegistry
- `examples_test.go` - Additional examples demonstrating testing patterns

These files demonstrate:

- Creating and validating RouteServices
- Testing with different configurations
- Event registration and management patterns
- Integration testing with real HTTP scenarios
- Performance benchmarking
- Service cloning and independence verification

Run the tests with:

```bash
go test ./internal/routeservices/... -v
```
