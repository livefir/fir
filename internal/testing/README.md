# Fir Testing Infrastructure

This package provides comprehensive testing utilities for the Fir framework, enabling fast, maintainable, and example-independent tests.

## Overview

The testing infrastructure consists of three main components:
- **HTTP Helpers** (`http_helpers.go`): HTTP, session, event, WebSocket, performance, and concurrency testing utilities
- **Builders** (`builders.go`): Fluent builders for routes, events, templates, expectations, and test scenarios
- **Mocks** (`mocks.go`): Mock implementations for Redis, WebSocket, HTTP handlers, network stubs, and session stores

## Quick Start

```go
import testhelpers "github.com/livefir/fir/internal/testing"

func TestMyFeature(t *testing.T) {
    // Create a test server
    controller := NewController("test", DevelopmentMode(true))
    server := httptest.NewServer(controller.RouteFunc(myRouteFunc))
    defer server.Close()

    // Create HTTP test helper
    helper := testhelpers.NewHTTPTestHelper(t, server)
    session := helper.GetSession()

    // Send events and validate responses
    resp := session.SendEvent("increment", nil)
    helper.ValidateResponse(resp).StatusCode(200).ContainsHTML("Count: 1")
}
```

## HTTP Helpers

### HTTPTestHelper

The main testing utility for HTTP-based Fir applications.

```go
helper := testhelpers.NewHTTPTestHelper(t, server)
```

#### Session Management

```go
// Get a new session
session := helper.GetSession()

// Get session with specific ID
session := helper.GetSessionWithID("custom-session-id")

// Access session ID
sessionID := session.SessionID()
```

#### Event Testing

```go
// Send regular event
resp := session.SendEvent("event_name", eventData)

// Send form event
resp := helper.SendFormEvent(sessionID, "event_name", map[string]string{
    "field1": "value1",
    "field2": "value2",
})

// Send event with session
resp := helper.SendEventWithSession(sessionID, "event_name", eventData)
```

#### Response Validation

```go
validator := helper.ValidateResponse(resp)

// Chain validations
validator.StatusCode(200).
    ContainsHTML("Expected content").
    NotContainsHTML("Error message").
    HasHeader("Content-Type", "text/html")
```

### WebSocket Testing

```go
wsHelper := helper.NewWebSocketHelper()

// Connect multiple WebSocket clients
connections := wsHelper.ConnectMultiple(3)
defer wsHelper.CloseConnections(connections)

// Send events via WebSocket
wsHelper.SendEvent(conn, "event_name", eventData)
```

### Performance Testing

```go
perfHelper := helper.NewPerformanceHelper()

// Measure event throughput
throughput, duration := perfHelper.MeasureEventThroughput(sessionID, "event_name", 100, false)

// Measure memory usage
baselineAlloc, finalAlloc := perfHelper.MeasureMemoryUsage(func() {
    // Your test operations
})

// Measure response time
responseTime := perfHelper.MeasureResponseTime(sessionID, "event_name", nil)
```

### Concurrency Testing

```go
concurrencyHelper := helper.NewConcurrencyHelper()

// Run concurrent operations
operations := []func(){
    func() { session.SendEvent("op1", nil) },
    func() { session.SendEvent("op2", nil) },
}
concurrencyHelper.RunConcurrentOperations(operations)

// Run concurrent events with multiple sessions
sessionIDs := []string{"session1", "session2", "session3"}
concurrencyHelper.RunConcurrentEvents(sessionIDs, "event_name", eventData)
```

## Builders

### RouteBuilder

Build route templates for testing:

```go
routeBuilder := testhelpers.NewRouteBuilder().
    WithID("test-route").
    WithCounterTemplate(0).
    WithIncrementEvent("increment").
    WithDecrementEvent("decrement")

// Generate route content
content := routeBuilder.Build()
```

### EventBuilder

Build event data structures:

```go
event := testhelpers.NewEventBuilder().
    WithID("user_update").
    WithParam("user_id", 123).
    WithParam("name", "John Doe").
    AsForm().
    WithSession("session-123").
    Build()
```

### TemplateBuilder

Build HTML templates with Fir directives:

```go
template := testhelpers.NewTemplateBuilder().
    WithRefreshDirective("submit", "reset").
    WithDiv("form-container", "").
    WithInput("username", "Enter username").
    WithButton("Submit", "submit").
    WithConditional(".error", `<div class="error">{{.error}}</div>`).
    Build()
```

### ExpectationBuilder

Build test expectations:

```go
expectation := testhelpers.NewExpectationBuilder().
    StatusCode(200).
    ContainsHTML("Success").
    NotContainsHTML("Error").
    HasKV("count", 5).
    Build()
```

### ScenarioBuilder

Build complex test scenarios:

```go
scenario := testhelpers.NewScenarioBuilder().
    WithName("User Registration Flow").
    WithStep("Load form", "load", nil).
    WithStep("Submit form", "submit", formData).
    WithStep("Verify success", "verify", nil).
    WithExpectation(expectation).
    Build()
```

## Mocks

### MockRedisClient

In-memory Redis client for testing:

```go
mockRedis := testhelpers.NewMockRedisClient()

// Basic operations
mockRedis.Set("key", "value")
value, err := mockRedis.Get("key")

// Pub/Sub
subscriber := mockRedis.Subscribe("channel")
mockRedis.Publish("channel", "message")

// Clear all data
mockRedis.Clear()
```

### MockWebSocketConnection

Mock WebSocket for testing:

```go
mockWS := testhelpers.NewMockWebSocketConnection()

// Send and receive messages
mockWS.WriteMessage(websocket.TextMessage, []byte("test"))
messageType, data, err := mockWS.ReadMessage()

// Check connection state
if mockWS.IsClosed() {
    // Handle closed connection
}
```

### MockHTTPHandler

Mock HTTP handler for external services:

```go
mockHandler := testhelpers.NewMockHTTPHandler()

// Set responses
mockHandler.SetResponse("/api/users", `{"users": []}`)
mockHandler.SetStatusCode("/api/error", 500)

// Get request history
requests := mockHandler.GetRequests()
```

### NetworkStub

Stub external network calls:

```go
networkStub := testhelpers.NewNetworkStub()

// Set responses
networkStub.SetResponse("/api/external", map[string]string{"status": "ok"})

// Simulate calls
response, err := networkStub.SimulateCall("/api/external")
```

### MockSessionStore

Mock session store for testing:

```go
sessionStore := testhelpers.NewMockSessionStore()

// Store and retrieve session data
sessionStore.Set("session-id", "user_id", 123)
value, exists := sessionStore.Get("session-id", "user_id")
```

### TestFixture

Integrated testing fixture with all mocks:

```go
fixture := testhelpers.NewTestFixture()
defer fixture.Cleanup()

// Setup scenarios
fixture.SetupRedisScenario()
wsConnections := fixture.SetupWebSocketScenario(3)

// Access mocks
fixture.RedisClient.Set("key", "value")
fixture.SessionStore.Set("session", "key", "value")
fixture.NetworkStub.SetResponse("/api", response)
```

## Best Practices

### 1. Use Session Managers

Always use the session manager for consistent session handling:

```go
session := helper.GetSession()
resp := session.SendEvent("event", data)
```

### 2. Chain Validations

Chain response validations for comprehensive testing:

```go
helper.ValidateResponse(resp).
    StatusCode(200).
    ContainsHTML("Expected").
    NotContainsHTML("Error").
    HasHeader("Content-Type", "text/html")
```

### 3. Use Test Fixtures for Integration Tests

For complex tests involving multiple components:

```go
fixture := testhelpers.NewTestFixture()
defer fixture.Cleanup()

fixture.SetupRedisScenario()
fixture.SetupWebSocketScenario(3)
```

### 4. Measure Performance

Include performance measurements in your tests:

```go
perfHelper := helper.NewPerformanceHelper()
throughput, duration := perfHelper.MeasureEventThroughput(sessionID, "event", 100, false)
t.Logf("Throughput: %.2f events/second", throughput)
```

### 5. Test Concurrency

Test concurrent scenarios to catch race conditions:

```go
concurrencyHelper := helper.NewConcurrencyHelper()
operations := make([]func(), 10)
for i := 0; i < 10; i++ {
    operations[i] = func() { /* concurrent operation */ }
}
concurrencyHelper.RunConcurrentOperations(operations)
```

## Examples

See `core_testing_infrastructure_demo_test.go` for comprehensive examples of all testing utilities in action.

## Architecture

The testing infrastructure is designed for:
- **Speed**: In-memory mocks and minimal setup overhead
- **Maintainability**: Fluent APIs and reusable components
- **Independence**: No dependencies on external examples or services
- **Comprehensive Coverage**: HTTP, WebSocket, performance, and concurrency testing

## Migration Guide

To migrate existing tests to use this infrastructure:

1. Replace manual HTTP clients with `HTTPTestHelper`
2. Use `SessionManager` for session handling
3. Replace custom event sending with helper methods
4. Use `ValidateResponse` for response assertions
5. Replace external dependencies with mocks from `TestFixture`

This infrastructure significantly reduces test complexity while providing comprehensive testing capabilities for Fir applications.
