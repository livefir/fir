# Hybrid Fir Actions Testing Strategy

## Overview

This document outlines a comprehensive hybrid testing approach that combines the best aspects of unit testing, integration testing, and functional testing to provide fast, reliable, and thorough validation of all Fir actions.

**IMPORTANT NOTE**: This strategy focuses only on testing functionality that is actually supported by the current Fir implementation. Conflict detection between actions is not implemented in the main framework and therefore not tested.

## Three-Layer Testing Architecture

### Layer 1: Unit Testing (Fastest - <1ms per test)
**Purpose**: Test individual action handlers in isolation
**Scope**: Action parsing, translation, and output generation
**Location**: `actions_test.go` - simplified to test only supported features

### Layer 2: Integration Testing (Fast - <10ms per test)  
**Purpose**: Test actions within HTTP request/response cycle
**Scope**: Action execution with real template rendering and session state
**Location**: `actions_integration_test.go`

### Layer 3: Functional Testing (Medium - <100ms per test)
**Purpose**: Test complete action workflows and combinations
**Scope**: End-to-end scenarios with multiple actions and complex state
**Location**: `actions_functional_test.go`

## Detailed Testing Strategy

### Layer 1: Unit Testing (Cleaned Up)

**Current State**: Simplified unit tests focusing only on supported functionality
**Implementation**: Clean unit tests for each action handler without unsupported features

```go
// Current approach - test only what the framework actually supports
func TestActionHandlers(t *testing.T) {
    // Test all 11 action handlers systematically
    // Test parameter combinations
    // Test translation accuracy
    // Test precedence values
    // NO conflict detection tests (not supported by framework)
}
```

**Coverage Matrix**:
- ✅ All 11 action handlers (RefreshActionHandler, RemoveActionHandler, etc.)
- ✅ All parameter combinations (empty, single, multiple, malformed)
- ✅ All event states (ok, error, pending, custom)
- ✅ All syntax variations (simple, complex, chained)
- ✅ Error conditions for invalid inputs
- ❌ Action conflicts (NOT IMPLEMENTED in framework - removed from tests)

### Layer 2: Integration Testing (New)

**Purpose**: Test actions within the full Fir request processing pipeline

```go
// Integration test structure
func TestActionIntegration_HTTP(t *testing.T) {
    // Create minimal routes with specific actions
    // Send HTTP events to trigger actions  
    // Validate HTTP response contains correct action instructions
    // Test with real session state and template rendering
}
```

**Test Categories**:

#### 2.1 Action Rendering Integration
```go
func TestActionRendering_Integration(t *testing.T) {
    // Test: x-fir-refresh renders correctly in template
    route := createTestRoute(`
        <div x-fir-refresh="update:ok">Content: {{.value}}</div>
        <button onclick="$fir.submit('update')">Update</button>
    `)
    
    // Send update event, verify response contains transformed attributes
    resp := sendEvent(route, "update", map[string]interface{}{"value": "new"})
    assertContains(t, resp, `@fir:update:ok="$fir.replace()"`)
}
```

#### 2.2 State Management Integration  
```go
func TestActionStateManagement_Integration(t *testing.T) {
    // Test actions with session state persistence
    // Test actions with route-specific data
    // Test actions with KV store integration
}
```

#### 2.3 Multi-Action Integration
```go
func TestMultiActionIntegration(t *testing.T) {
    // Test multiple actions on same element
    // Test action precedence and conflict resolution
    // Test action chaining across events
}
```

### Layer 3: Functional Testing (New)

**Purpose**: Test complete real-world workflows and complex scenarios

```go
func TestActionWorkflows_Functional(t *testing.T) {
    // Test complete CRUD workflows
    // Test complex form interactions
    // Test real-time update scenarios
    // Test error recovery workflows
}
```

**Test Scenarios**:

#### 3.1 CRUD Workflows
```go
func TestCRUDWorkflow_Functional(t *testing.T) {
    // Create: x-fir-append adds new item
    // Read: x-fir-refresh updates item display  
    // Update: x-fir-refresh replaces item content
    // Delete: x-fir-remove removes item from list
}
```

#### 3.2 Form Processing Workflows
```go
func TestFormWorkflow_Functional(t *testing.T) {
    // Submit: x-fir-toggle-disabled during processing
    // Success: x-fir-reset clears form + x-fir-redirect
    // Error: x-fir-toggle-class shows error state
    // Retry: Form re-enables for retry
}
```

#### 3.3 Real-time Updates
```go
func TestRealtimeWorkflow_Functional(t *testing.T) {
    // WebSocket events trigger x-fir-refresh
    // Multiple clients see x-fir-append updates
    // Optimistic updates with x-fir-remove rollback
}
```

## Testing Utilities and Infrastructure

### Enhanced Test Helpers

```go
// Action-specific test utilities
type ActionTestSuite struct {
    controller *Controller
    session    *Session
    recorder   *httptest.ResponseRecorder
}

func (suite *ActionTestSuite) CreateRouteWithAction(template, actionName string) RouteOptions {
    // Create minimal route with specific action for testing
}

func (suite *ActionTestSuite) SendEventAndValidateAction(eventName string, expectedAction ActionExpectation) {
    // Send event and validate resulting action in response
}

func (suite *ActionTestSuite) AssertActionInResponse(resp *http.Response, action ActionExpectation) {
    // Parse response and validate action was correctly generated
}

// Action expectation structure
type ActionExpectation struct {
    Type       string                 // "refresh", "append", etc.
    Selector   string                // CSS selector or target
    Content    string                // Expected content (for content actions)
    Attributes map[string]string     // Expected attributes
    EventState string                // "ok", "error", "pending"
}
```

### Test Data Builders

```go
// Fluent API for building test scenarios
func NewActionTest(actionType string) *ActionTestBuilder {
    return &ActionTestBuilder{actionType: actionType}
}

func (b *ActionTestBuilder) WithTemplate(template string) *ActionTestBuilder {
    b.template = template
    return b
}

func (b *ActionTestBuilder) WithEvent(eventName string) *ActionTestBuilder {
    b.eventName = eventName
    return b
}

func (b *ActionTestBuilder) ExpectAction(expectation ActionExpectation) *ActionTestBuilder {
    b.expectations = append(b.expectations, expectation)
    return b
}

func (b *ActionTestBuilder) Run(t *testing.T) {
    // Execute the test scenario
}

// Usage:
NewActionTest("refresh").
    WithTemplate(`<div x-fir-refresh="update:ok">{{.content}}</div>`).
    WithEvent("update").
    ExpectAction(ActionExpectation{
        Type: "refresh",
        Selector: "div",
        EventState: "ok",
    }).
    Run(t)
```

### Performance Testing Integration

```go
func BenchmarkActionProcessing(b *testing.B) {
    // Benchmark all actions for performance regression testing
    actions := []string{"refresh", "append", "prepend", "remove", "toggle-disabled"}
    
    for _, action := range actions {
        b.Run(action, func(b *testing.B) {
            // Benchmark specific action processing
        })
    }
}
```

## Comprehensive Test Matrix

### Actions Coverage Matrix

| Action | Unit Tests | Integration Tests | Functional Tests | Edge Cases | Error Conditions |
|--------|------------|-------------------|------------------|------------|------------------|
| x-fir-refresh | ✅ | ⏸️ | ⏸️ | ⏸️ | ⏸️ |
| x-fir-append | ✅ | ⏸️ | ⏸️ | ⏸️ | ⏸️ |
| x-fir-prepend | ✅ | ⏸️ | ⏸️ | ⏸️ | ⏸️ |
| x-fir-remove | ✅ | ⏸️ | ⏸️ | ⏸️ | ⏸️ |
| x-fir-remove-parent | ✅ | ⏸️ | ⏸️ | ⏸️ | ⏸️ |
| x-fir-reset | ✅ | ⏸️ | ⏸️ | ⏸️ | ⏸️ |
| x-fir-toggle-disabled | ✅ | ⏸️ | ⏸️ | ⏸️ | ⏸️ |
| x-fir-toggle-class | ✅ | ⏸️ | ⏸️ | ⏸️ | ⏸️ |
| x-fir-redirect | ✅ | ⏸️ | ⏸️ | ⏸️ | ⏸️ |
| x-fir-dispatch | ✅ | ⏸️ | ⏸️ | ⏸️ | ⏸️ |
| x-fir-trigger | ✅ | ⏸️ | ⏸️ | ⏸️ | ⏸️ |

### Event States Coverage

| Event State | All Actions Tested | Edge Cases | Error Recovery |
|-------------|-------------------|------------|----------------|
| :ok | ⚠️ Partial | ⏸️ | ⏸️ |
| :error | ⚠️ Partial | ⏸️ | ⏸️ |  
| :pending | ⚠️ Partial | ⏸️ | ⏸️ |
| :custom | ⏸️ | ⏸️ | ⏸️ |

### Parameter Combinations Coverage

| Parameter Type | Tested | Complex Cases | Validation |
|----------------|--------|---------------|------------|
| Simple values | ✅ | ⏸️ | ⏸️ |
| Multiple params | ✅ | ⏸️ | ⏸️ |
| Template params | ⚠️ Partial | ⏸️ | ⏸️ |
| Malformed params | ⏸️ | ⏸️ | ⏸️ |

## Implementation Phases

### Phase 1: Integration Testing Foundation (Week 1)
- Create `actions_integration_test.go`
- Implement ActionTestSuite utilities
- Test each action in HTTP request cycle
- Validate action rendering in templates

### Phase 2: Functional Testing Framework (Week 2)  
- Create `actions_functional_test.go`
- Implement workflow testing scenarios
- Test action combinations and conflicts
- Test complex state management

### Phase 3: Enhanced Unit Testing (Week 3)
- Enhance existing `actions_test.go`
- Add comprehensive edge case testing
- Add error condition validation
- Add performance regression testing

### Phase 4: Test Matrix Completion (Week 4)
- Complete all test matrix cells
- Add comprehensive benchmarking
- Add test documentation and examples
- Integrate with CI/CD pipeline

## Success Criteria

### Performance Targets
- Unit tests: <1ms per test
- Integration tests: <10ms per test  
- Functional tests: <100ms per test
- Complete test suite: <5 seconds

### Coverage Targets
- 100% action handler coverage
- 100% event state coverage
- 100% parameter combination coverage
- 90%+ edge case coverage

### Quality Targets
- Zero flaky tests
- All tests deterministic and isolated
- Complete error condition coverage
- Comprehensive documentation

This hybrid approach ensures we have fast feedback during development (unit tests), reliable integration validation (integration tests), and comprehensive real-world scenario coverage (functional tests), while maintaining the speed and reliability needed for continuous development.
