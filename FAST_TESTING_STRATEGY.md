# Fir Framework Testing Strategy Plan

## Problem Analysis

### Current Test Categories

Based on my analysis of the codebase, I've identified three main categories of tests:

#### 1. **E2E Tests (Browser-based)** - SLOW
- **Location**: `examples/e2e/*_test.go`
- **Dependencies**: Chrome browser via chromedp
- **What they test**: Full UI interaction, DOM manipulation, JavaScript execution
- **Examples**: 
  - `counter_test.go` - Tests clicking buttons and seeing count updates
  - `chirper_test.go` - Tests form submissions and real-time updates
  - `autocomplete_test.go` - Tests dynamic search functionality
- **Speed**: Very slow (30+ seconds per test due to browser startup)

#### 2. **Integration Tests (Docker-based)** - MEDIUM-SLOW
- **Location**: `pubsub/pubsub_test.go`, `controller_test.go` (Redis tests)
- **Dependencies**: Docker containers (Redis)
- **What they test**: Redis pub/sub functionality, multi-user scenarios
- **Examples**:
  - `TestRedisPublishAndSubscribe` - Tests Redis pub/sub with real Redis
  - `TestControllerWebsocketEnabledRedis` - Tests WebSocket with Redis backend
- **Speed**: Slow due to Docker container startup

#### 3. **Unit/HTTP Tests** - FAST
- **Location**: `internal/services/*_test.go`, `internal/handlers/*_test.go`, etc.
- **Dependencies**: `net/http/httptest` only
- **What they test**: Individual components, HTTP request/response cycles
- **Examples**:
  - `counter/counter_http_test.go` - Tests HTTP responses without browser
  - Service layer tests - Test business logic in isolation
- **Speed**: Fast (milliseconds)

### Core Issue

**The problem**: E2E tests are testing core functionality (event handling, template rendering, state management) but require slow browser automation. This creates a development bottleneck during refactoring.

**The solution**: Create fast HTTP-based tests that validate the same core functionality without needing a browser.

## Proposed Testing Strategy

### Phase 1: Fast Core Functionality Tests

Create `httptest`-based tests that replicate E2E functionality without browsers:

#### A. **Event Processing Tests** (Replace browser clicking)
```go
// Instead of: chromedp.Click(incrementButtonSelector)
// Do: HTTP POST with event data

func TestCounterIncrementEvent_HTTP(t *testing.T) {
    controller := fir.NewController("counter_test")
    server := httptest.NewServer(controller.RouteFunc(counter.Index))
    defer server.Close()
    
    // 1. Get initial page and extract session
    resp := getPage(t, server.URL)
    session := extractSession(resp)
    
    // 2. Send increment event via HTTP
    eventResp := sendEvent(t, server.URL, session, "inc", nil)
    
    // 3. Verify response contains updated count
    assertHTMLContains(t, eventResp, `Count: 1`)
}
```

#### B. **WebSocket Communication Tests** (Replace browser WebSocket)
```go
func TestCounterWebSocket_HTTP(t *testing.T) {
    controller := fir.NewController("counter_ws_test")
    server := httptest.NewServer(controller.RouteFunc(counter.Index))
    defer server.Close()
    
    // Convert to WebSocket URL and connect
    wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
    ws := connectWebSocket(t, wsURL)
    defer ws.Close()
    
    // Send event via WebSocket, verify response
    sendWSEvent(t, ws, "inc", nil)
    verifyWSResponse(t, ws, `Count: 1`)
}
```

#### C. **Template Rendering Tests** (Replace DOM inspection)
```go
func TestCounterTemplateRendering_HTTP(t *testing.T) {
    controller := fir.NewController("counter_template_test")
    server := httptest.NewServer(controller.RouteFunc(counter.Index))
    defer server.Close()
    
    // Test initial render
    resp := getPage(t, server.URL)
    assertHTMLStructure(t, resp, []string{
        `<div.*Count:.*0.*</div>`,
        `<button.*formaction="/?event=inc"`,
        `<button.*formaction="/?event=dec"`,
    })
    
    // Test post-event render
    session := extractSession(resp)
    eventResp := sendEvent(t, server.URL, session, "inc", nil)
    assertHTMLStructure(t, eventResp, []string{
        `<div.*Count:.*1.*</div>`,
    })
}
```

#### D. **Session Management Tests** (Replace browser session handling)
```go
func TestSessionPersistence_HTTP(t *testing.T) {
    controller := fir.NewController("session_test")
    server := httptest.NewServer(controller.RouteFunc(counter.Index))
    defer server.Close()
    
    // Get session from first request
    resp1 := getPage(t, server.URL)
    session := extractSession(resp1)
    
    // Send event with session
    eventResp := sendEvent(t, server.URL, session, "inc", nil)
    
    // Verify session maintains state across requests
    resp2 := getPageWithSession(t, server.URL, session)
    assertHTMLContains(t, resp2, `Count: 1`) // State persisted
}
```

### Phase 2: Mock-based Fast Integration Tests

Replace Docker-dependent tests with mocks:

#### A. **Redis Pub/Sub Mock Tests**
```go
func TestRedisPubSub_Mock(t *testing.T) {
    // Use go-redis/redismock instead of real Redis
    mockRedis := redismock.NewClientMock()
    
    // Set up expectations
    mockRedis.ExpectPublish("channel", "message").SetVal(1)
    mockRedis.ExpectSubscribe("channel")
    
    // Test pub/sub functionality
    pubsub := fir.NewRedisPubSub(mockRedis)
    // ... test logic
    
    // Verify all expectations met
    if err := mockRedis.ExpectationsWereMet(); err != nil {
        t.Errorf("Redis expectations not met: %v", err)
    }
}
```

#### B. **Multi-User WebSocket Tests**
```go
func TestMultiUserWebSocket_InMemory(t *testing.T) {
    // Use in-memory pub/sub instead of Redis
    controller := fir.NewController("multi_user_test", 
        fir.WithInMemoryPubSub()) // Force in-memory for testing
    
    server := httptest.NewServer(controller.RouteFunc(counter.Index))
    defer server.Close()
    
    // Connect multiple WebSocket clients
    ws1 := connectWebSocket(t, wsURL)
    ws2 := connectWebSocket(t, wsURL)
    
    // Send event from client 1
    sendWSEvent(t, ws1, "inc", nil)
    
    // Verify both clients receive update
    verifyWSResponse(t, ws1, `Count: 1`)
    verifyWSResponse(t, ws2, `Count: 1`) // Broadcast received
}
```

### Phase 3: Test Utilities and Helpers

Create reusable testing utilities:

#### A. **HTTP Test Helpers**
```go
// internal/testing/http_helpers.go
package testing

func GetPage(t *testing.T, url string) *http.Response { ... }
func GetPageWithSession(t *testing.T, url, session string) *http.Response { ... }
func SendEvent(t *testing.T, url, session, eventName string, params map[string]interface{}) *http.Response { ... }
func ExtractSession(resp *http.Response) string { ... }
func AssertHTMLContains(t *testing.T, resp *http.Response, expected string) { ... }
func AssertHTMLStructure(t *testing.T, resp *http.Response, patterns []string) { ... }
```

#### B. **WebSocket Test Helpers**
```go
// internal/testing/websocket_helpers.go
func ConnectWebSocket(t *testing.T, url string) *websocket.Conn { ... }
func ConnectWebSocketWithSession(t *testing.T, url, session string) *websocket.Conn { ... }
func SendWSEvent(t *testing.T, ws *websocket.Conn, eventName string, params map[string]interface{}) { ... }
func VerifyWSResponse(t *testing.T, ws *websocket.Conn, expected string) { ... }
```

#### C. **Mock Factories**
```go
// internal/testing/mocks.go
func NewMockRedisClient() *redismock.ClientMock { ... }
func NewTestController(name string, opts ...fir.ControllerOption) *fir.Controller { ... }
func NewTestServer(routeFunc fir.RouteFunc) *httptest.Server { ... }
```

## Implementation Plan

### Priority 1: Core HTTP Tests (Week 1)
1. **Counter Example HTTP Tests**
   - `examples/counter/counter_core_test.go`
   - Test all counter functionality without browser
   - Validate event processing, state management, template rendering

2. **WebSocket Core Tests**
   - `websocket_core_test.go`
   - Test WebSocket event handling without browser
   - Use `net/http/httptest` + `gorilla/websocket`

### Priority 2: Complex Example Tests (Week 2)
1. **Chirper HTTP Tests**
   - `examples/chirper/chirper_core_test.go`
   - Test form submissions, user interactions via HTTP
   
2. **Autocomplete HTTP Tests**
   - `examples/autocomplete/autocomplete_core_test.go`
   - Test dynamic search via HTTP API calls

### Priority 3: Mock Integration Tests (Week 3)
1. **Redis Mock Tests**
   - Replace `pubsub/pubsub_test.go` Docker tests with mocks
   - Faster CI/CD pipeline
   
2. **Multi-User Scenarios**
   - Test multi-user WebSocket scenarios with in-memory pub/sub
   - Validate broadcasting without Redis dependency

### Priority 4: Test Infrastructure (Week 4)
1. **Test Utilities Package**
   - Create `internal/testing/` package with helpers
   - Standardize test patterns across codebase
   
2. **Test Documentation**
   - Update testing guide with new patterns
   - Examples of fast vs slow test approaches

## Expected Benefits

### 🚀 **Speed Improvements**
- **E2E Tests**: 30+ seconds → 100-500ms per test
- **Integration Tests**: 5-10 seconds → 10-50ms per test
- **Total CI Time**: 5+ minutes → 30 seconds

### 🔧 **Development Experience**
- **Refactoring**: Fast feedback during architectural changes
- **Debugging**: Easier to debug HTTP tests vs browser tests
- **Parallelization**: HTTP tests can run in parallel easily

### 🏗️ **Architecture Validation**
- **Core Logic**: Validate business logic without UI dependencies
- **API Contracts**: Test HTTP/WebSocket APIs independently
- **Session Handling**: Verify session management across requests

### 📊 **Test Coverage**
- **More Comprehensive**: Easier to test edge cases
- **Faster Iteration**: Write more tests due to fast execution
- **Better CI/CD**: Reliable tests that don't depend on browser/Docker

## Migration Strategy

### Phase 1: Parallel Implementation
- Keep existing E2E/Docker tests
- Add new HTTP-based tests alongside
- Validate both approaches produce same results

### Phase 2: Selective Replacement
- Replace slow tests with fast equivalents for CI
- Keep some E2E tests for final validation
- Use fast tests for development, slow tests for releases

### Phase 3: Optimization
- Fine-tune test suite for optimal balance
- Keep critical E2E tests, remove redundant ones
- Optimize CI pipeline with test categorization

This strategy maintains test quality while dramatically improving development velocity during major refactoring efforts like Milestones 5 and 6.
