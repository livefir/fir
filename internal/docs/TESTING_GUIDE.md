# Fir Framework Testing Guide

This guide demonstrates how to write unit tests for the new service layer architecture in Fir framework.

## Overview

The Fir framework now uses a service layer architecture that makes components highly testable:

- **Service Layer**: Event processing, rendering, and template services can be unit tested with mocks
- **Handler Layer**: Individual request handlers can be tested independently  
- **HTTP Abstraction**: Request/response models enable testing without HTTP infrastructure
- **Dependency Injection**: All dependencies are injected via interfaces for easy mocking

## Testing Strategy

### 1. Unit Testing Services

Services can be tested independently by mocking their dependencies.

#### Testing Event Service

```go
func TestEventService_ProcessEvent(t *testing.T) {
    // Create mocks
    mockRegistry := &MockEventRegistry{}
    mockValidator := &MockEventValidator{}
    mockPublisher := &MockEventPublisher{}
    
    // Setup expectations
    mockRegistry.EXPECT().Get("route1", "click").Return(mockHandler, true)
    mockValidator.EXPECT().ValidateEvent(gomock.Any()).Return(nil)
    
    // Create service with mocks
    service := services.NewDefaultEventService(
        mockRegistry,
        mockValidator, 
        mockPublisher,
        &MockEventLogger{},
        &MockEventMetrics{},
    )
    
    // Create test request
    req := services.EventRequest{
        ID:        "click",
        SessionID: "session1",
        Context:   context.Background(),
        Params:    map[string]interface{}{"value": "test"},
    }
    
    // Execute and verify
    response, err := service.ProcessEvent(context.Background(), req)
    assert.NoError(t, err)
    assert.Equal(t, 200, response.StatusCode)
}
```

#### Testing Render Service

```go
func TestRenderService_RenderTemplate(t *testing.T) {
    // Create mocks
    mockTemplateService := &MockTemplateService{}
    mockRenderer := &MockRenderer{}
    
    // Setup expectations
    mockTemplateService.EXPECT().
        LoadTemplate(gomock.Any()).
        Return(mockTemplate, nil)
    mockRenderer.EXPECT().
        RenderTemplate(mockTemplate, gomock.Any()).
        Return([]byte("<div>Test</div>"), nil)
    
    // Create service
    service := services.NewDefaultRenderService(
        mockTemplateService,
        mockRenderer,
        &MockRenderLogger{},
    )
    
    // Create test context
    ctx := services.RenderContext{
        RouteID: "test-route",
        Data:    map[string]interface{}{"name": "test"},
    }
    
    // Execute and verify
    result, err := service.RenderTemplate(ctx)
    assert.NoError(t, err)
    assert.Contains(t, string(result.HTML), "Test")
}
```

### 2. Testing Request Handlers

Handlers can be tested by providing mock services and request models.

#### Testing JSON Event Handler

```go
func TestJSONEventHandler_Handle(t *testing.T) {
    // Create mocks
    mockEventService := &MockEventService{}
    mockRenderService := &MockRenderService{}
    mockResponseBuilder := &MockResponseBuilder{}
    
    // Setup handler
    handler := handlers.NewJSONEventHandler(
        mockEventService,
        mockRenderService,
        mockResponseBuilder,
        nil, // validator optional
    )
    
    // Create test request
    req := &firHttp.RequestModel{
        Method: "POST",
        Header: http.Header{
            "Content-Type": []string{"application/json"},
        },
        Body: io.NopCloser(strings.NewReader(`{"event":"click","target":"#btn"}`)),
    }
    
    // Setup expectations
    mockEventService.EXPECT().
        ProcessEvent(gomock.Any(), gomock.Any()).
        Return(&services.EventResponse{StatusCode: 200}, nil)
    mockResponseBuilder.EXPECT().
        BuildEventResponse(gomock.Any(), gomock.Any()).
        Return(&firHttp.ResponseModel{StatusCode: 200}, nil)
    
    // Execute
    response, err := handler.Handle(context.Background(), req)
    
    // Verify
    assert.NoError(t, err)
    assert.Equal(t, 200, response.StatusCode)
}
```

#### Testing Form Handler

```go
func TestFormHandler_Handle(t *testing.T) {
    // Create mocks
    mockEventService := &MockEventService{}
    mockRenderService := &MockRenderService{}
    mockResponseBuilder := &MockResponseBuilder{}
    
    handler := handlers.NewFormHandler(
        mockEventService,
        mockRenderService,
        mockResponseBuilder,
        nil,
    )
    
    // Create form data
    formData := url.Values{}
    formData.Set("event", "submit")
    formData.Set("name", "test")
    
    req := &firHttp.RequestModel{
        Method:   "POST",
        PostForm: formData,
        Header: http.Header{
            "Content-Type": []string{"application/x-www-form-urlencoded"},
        },
    }
    
    // Test execution
    response, err := handler.Handle(context.Background(), req)
    assert.NoError(t, err)
}
```

### 3. Testing Handler Chain

The handler chain can be tested to ensure proper request routing.

```go
func TestHandlerChain_Handle(t *testing.T) {
    // Create chain
    chain := handlers.NewDefaultHandlerChain(&MockLogger{}, &MockMetrics{})
    
    // Add test handlers
    jsonHandler := &MockJSONEventHandler{}
    formHandler := &MockFormHandler{}
    
    chain.AddHandler(jsonHandler)
    chain.AddHandler(formHandler)
    
    // Create JSON request
    jsonReq := &firHttp.RequestModel{
        Method: "POST",
        Header: http.Header{"Content-Type": []string{"application/json"}},
    }
    
    // Setup expectations
    jsonHandler.EXPECT().SupportsRequest(jsonReq).Return(true)
    jsonHandler.EXPECT().Handle(gomock.Any(), jsonReq).Return(&firHttp.ResponseModel{}, nil)
    
    // Execute
    response, err := chain.Handle(context.Background(), jsonReq)
    assert.NoError(t, err)
}
```

### 4. Integration Testing with HTTP Adapter

Test the full request flow including HTTP conversion.

```go
func TestRouteIntegration_FullFlow(t *testing.T) {
    // Create HTTP request
    httpReq := httptest.NewRequest("POST", "/test", strings.NewReader(`{"event":"click"}`))
    httpReq.Header.Set("Content-Type", "application/json")
    
    // Create response recorder
    w := httptest.NewRecorder()
    
    // Create HTTP adapter
    adapter := firHttp.NewStandardHTTPAdapter(w, httpReq, nil)
    
    // Parse request
    requestModel, err := adapter.ParseRequest(httpReq)
    assert.NoError(t, err)
    
    // Create mock handler chain
    mockChain := &MockHandlerChain{}
    mockChain.EXPECT().
        Handle(gomock.Any(), requestModel).
        Return(&firHttp.ResponseModel{
            StatusCode: 200,
            Body:       []byte("OK"),
        }, nil)
    
    // Execute through chain
    response, err := mockChain.Handle(context.Background(), requestModel)
    assert.NoError(t, err)
    assert.Equal(t, 200, response.StatusCode)
}
```

## Mock Generation

Use mockgen to generate mocks for interfaces:

```bash
# Generate mocks for services
go generate ./internal/services/...

# Generate mocks for handlers  
go generate ./internal/handlers/...

# Generate mocks for HTTP layer
go generate ./internal/http/...
```

Add `//go:generate` comments to interface files:

```go
//go:generate mockgen -source=interfaces.go -destination=mocks/mock_interfaces.go
type EventService interface {
    ProcessEvent(ctx context.Context, req EventRequest) (*EventResponse, error)
}
```

## Best Practices

### 1. Test Structure

```go
func TestServiceMethod_Scenario(t *testing.T) {
    // Arrange - setup mocks and test data
    
    // Act - execute the method under test
    
    // Assert - verify results and mock expectations
}
```

### 2. Use Table Tests for Multiple Scenarios

```go
func TestEventService_ProcessEvent(t *testing.T) {
    tests := []struct {
        name        string
        request     services.EventRequest
        setupMocks  func(*MockEventRegistry, *MockEventValidator)
        expectError bool
        expectCode  int
    }{
        {
            name: "successful event processing",
            request: services.EventRequest{ID: "click"},
            setupMocks: func(registry *MockEventRegistry, validator *MockEventValidator) {
                registry.EXPECT().Get(gomock.Any(), "click").Return(mockHandler, true)
                validator.EXPECT().ValidateEvent(gomock.Any()).Return(nil)
            },
            expectError: false,
            expectCode:  200,
        },
        // More test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### 3. Test Error Paths

```go
func TestEventService_ProcessEvent_ValidationError(t *testing.T) {
    mockValidator := &MockEventValidator{}
    mockValidator.EXPECT().
        ValidateEvent(gomock.Any()).
        Return(errors.New("validation failed"))
    
    service := services.NewDefaultEventService(
        &MockEventRegistry{}, 
        mockValidator,
        // ... other mocks
    )
    
    _, err := service.ProcessEvent(context.Background(), services.EventRequest{})
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "validation failed")
}
```

### 4. Performance Testing

```go
func BenchmarkEventService_ProcessEvent(b *testing.B) {
    service := setupEventService() // Setup with real or optimized mocks
    req := services.EventRequest{ID: "test"}
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := service.ProcessEvent(context.Background(), req)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## Running Tests

```bash
# Run all unit tests
go test ./internal/services/... ./internal/handlers/... ./internal/http/...

# Run with coverage
go test -cover ./...

# Run specific test
go test -run TestEventService_ProcessEvent ./internal/services

# Run benchmarks
go test -bench=. ./internal/services
```

## Test Coverage Goals

- **Services**: 90%+ coverage
- **Handlers**: 85%+ coverage  
- **HTTP Adapters**: 95%+ coverage
- **Integration**: 70%+ coverage

The new architecture enables comprehensive unit testing that was not possible with the previous monolithic route design.
