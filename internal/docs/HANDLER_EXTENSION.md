# Handler Extension Guide

This guide shows how to extend the Fir framework's handler chain with custom request handlers.

## Overview

The handler chain architecture allows you to add custom handlers for specialized request processing. Common use cases include:

- Custom authentication handlers
- API versioning handlers
- Rate limiting handlers
- Custom content type processors
- Protocol adapters (GraphQL, gRPC-Web, etc.)

## Handler Interface

All handlers must implement the `RequestHandler` interface:

```go
type RequestHandler interface {
    Handle(ctx context.Context, req *RequestModel) (*ResponseModel, error)
    SupportsRequest(req *RequestModel) bool
    HandlerName() string
}
```

## Creating a Custom Handler

### Example: API Key Authentication Handler

```go
package handlers

import (
    "context"
    "fmt"
    "strings"
    
    firHttp "github.com/livefir/fir/internal/http"
)

type APIKeyHandler struct {
    validKeys    map[string]bool
    headerName   string
    nextHandler  RequestHandler
}

func NewAPIKeyHandler(validKeys []string, headerName string) *APIKeyHandler {
    keyMap := make(map[string]bool)
    for _, key := range validKeys {
        keyMap[key] = true
    }
    
    return &APIKeyHandler{
        validKeys:  keyMap,
        headerName: headerName,
    }
}

func (h *APIKeyHandler) HandlerName() string {
    return "APIKeyHandler"
}

func (h *APIKeyHandler) SupportsRequest(req *firHttp.RequestModel) bool {
    // Support all requests that require API key validation
    return strings.HasPrefix(req.URL.Path, "/api/")
}

func (h *APIKeyHandler) Handle(ctx context.Context, req *firHttp.RequestModel) (*firHttp.ResponseModel, error) {
    // Extract API key from header
    apiKey := req.Header.Get(h.headerName)
    if apiKey == "" {
        return &firHttp.ResponseModel{
            StatusCode: 401,
            Headers:    map[string]string{"Content-Type": "application/json"},
            Body:       []byte(`{"error": "API key required"}`),
        }, nil
    }
    
    // Validate API key
    if !h.validKeys[apiKey] {
        return &firHttp.ResponseModel{
            StatusCode: 403,
            Headers:    map[string]string{"Content-Type": "application/json"},
            Body:       []byte(`{"error": "Invalid API key"}`),
        }, nil
    }
    
    // Add authenticated user info to request context
    ctx = context.WithValue(ctx, "authenticated", true)
    ctx = context.WithValue(ctx, "api_key", apiKey)
    
    // Delegate to next handler
    if h.nextHandler != nil {
        return h.nextHandler.Handle(ctx, req)
    }
    
    // If no next handler, return success
    return &firHttp.ResponseModel{
        StatusCode: 200,
        Headers:    map[string]string{"Content-Type": "application/json"},
        Body:       []byte(`{"status": "authenticated"}`),
    }, nil
}
```

### Example: Rate Limiting Handler

```go
type RateLimitHandler struct {
    limiters map[string]*rate.Limiter
    mutex    sync.RWMutex
    rate     rate.Limit
    burst    int
}

func NewRateLimitHandler(requestsPerSecond int, burst int) *RateLimitHandler {
    return &RateLimitHandler{
        limiters: make(map[string]*rate.Limiter),
        rate:     rate.Limit(requestsPerSecond),
        burst:    burst,
    }
}

func (h *RateLimitHandler) HandlerName() string {
    return "RateLimitHandler"
}

func (h *RateLimitHandler) SupportsRequest(req *firHttp.RequestModel) bool {
    // Apply rate limiting to all requests
    return true
}

func (h *RateLimitHandler) Handle(ctx context.Context, req *firHttp.RequestModel) (*firHttp.ResponseModel, error) {
    // Get client identifier (IP address)
    clientID := h.getClientID(req)
    
    // Get or create limiter for this client
    limiter := h.getLimiter(clientID)
    
    // Check if request is allowed
    if !limiter.Allow() {
        return &firHttp.ResponseModel{
            StatusCode: 429,
            Headers: map[string]string{
                "Content-Type":   "application/json",
                "Retry-After":    "1",
                "X-RateLimit-Limit": fmt.Sprintf("%d", int(h.rate)),
            },
            Body: []byte(`{"error": "Rate limit exceeded"}`),
        }, nil
    }
    
    // Request allowed, continue processing
    return nil, nil // Return nil to continue to next handler
}

func (h *RateLimitHandler) getLimiter(clientID string) *rate.Limiter {
    h.mutex.RLock()
    limiter, exists := h.limiters[clientID]
    h.mutex.RUnlock()
    
    if !exists {
        h.mutex.Lock()
        limiter = rate.NewLimiter(h.rate, h.burst)
        h.limiters[clientID] = limiter
        h.mutex.Unlock()
    }
    
    return limiter
}

func (h *RateLimitHandler) getClientID(req *firHttp.RequestModel) string {
    // Use X-Forwarded-For if available, otherwise RemoteAddr
    if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
        return strings.Split(xff, ",")[0]
    }
    return req.RemoteAddr
}
```

### Example: GraphQL Handler

```go
type GraphQLHandler struct {
    schema     graphql.Schema
    rootValue  interface{}
}

func NewGraphQLHandler(schema graphql.Schema) *GraphQLHandler {
    return &GraphQLHandler{
        schema: schema,
    }
}

func (h *GraphQLHandler) HandlerName() string {
    return "GraphQLHandler"
}

func (h *GraphQLHandler) SupportsRequest(req *firHttp.RequestModel) bool {
    return req.Method == "POST" && 
           req.URL.Path == "/graphql" &&
           strings.Contains(req.Header.Get("Content-Type"), "application/json")
}

func (h *GraphQLHandler) Handle(ctx context.Context, req *firHttp.RequestModel) (*firHttp.ResponseModel, error) {
    // Parse GraphQL query from request body
    body, err := io.ReadAll(req.Body)
    if err != nil {
        return h.errorResponse("Failed to read request body"), nil
    }
    
    var query struct {
        Query     string                 `json:"query"`
        Variables map[string]interface{} `json:"variables"`
    }
    
    if err := json.Unmarshal(body, &query); err != nil {
        return h.errorResponse("Invalid JSON"), nil
    }
    
    // Execute GraphQL query
    params := graphql.Params{
        Schema:        h.schema,
        RequestString: query.Query,
        VariableValues: query.Variables,
        Context:       ctx,
        RootObject:    h.rootValue,
    }
    
    result := graphql.Do(params)
    
    // Build response
    responseBody, err := json.Marshal(result)
    if err != nil {
        return h.errorResponse("Failed to marshal response"), nil
    }
    
    return &firHttp.ResponseModel{
        StatusCode: 200,
        Headers:    map[string]string{"Content-Type": "application/json"},
        Body:       responseBody,
    }, nil
}

func (h *GraphQLHandler) errorResponse(message string) *firHttp.ResponseModel {
    errorBody, _ := json.Marshal(map[string]string{"error": message})
    return &firHttp.ResponseModel{
        StatusCode: 400,
        Headers:    map[string]string{"Content-Type": "application/json"},
        Body:       errorBody,
    }
}
```

## Adding Handlers to the Chain

### Option 1: Modify Default Setup

Modify the `SetupDefaultHandlerChain` function to include your custom handlers:

```go
func SetupDefaultHandlerChain(services *routeservices.RouteServices) HandlerChain {
    chain := NewPriorityHandlerChain(&defaultHandlerLogger{}, &defaultHandlerMetrics{})
    
    // Add custom authentication handler (highest priority)
    apiKeyHandler := NewAPIKeyHandler([]string{"secret-key-1", "secret-key-2"}, "X-API-Key")
    chain.AddHandlerWithConfig(apiKeyHandler, HandlerConfig{
        Name:     "APIKeyHandler",
        Priority: 1, // Highest priority
        Enabled:  true,
    })
    
    // Add rate limiting (second priority)
    rateLimitHandler := NewRateLimitHandler(100, 10) // 100 req/sec, burst 10
    chain.AddHandlerWithConfig(rateLimitHandler, HandlerConfig{
        Name:     "RateLimitHandler", 
        Priority: 2,
        Enabled:  true,
    })
    
    // Add existing handlers with lower priorities
    if services.EventService != nil {
        wsHandler := NewWebSocketHandler(services.EventService, services.ResponseBuilder)
        chain.AddHandlerWithConfig(wsHandler, HandlerConfig{
            Name:     "WebSocketHandler",
            Priority: 10,
            Enabled:  true,
        })
    }
    
    // ... add other default handlers
    
    return chain
}
```

### Option 2: Custom Factory

Create a custom factory that builds chains with your handlers:

```go
type CustomHandlerChainFactory struct {
    services *routeservices.RouteServices
    config   CustomHandlerConfig
}

type CustomHandlerConfig struct {
    APIKeysEnabled    bool
    APIKeys          []string
    RateLimitEnabled bool
    RateLimit        int
    GraphQLEnabled   bool
    GraphQLSchema    graphql.Schema
}

func (f *CustomHandlerChainFactory) CreateHandlerChain() handlers.HandlerChain {
    chain := handlers.NewPriorityHandlerChain(&defaultHandlerLogger{}, &defaultHandlerMetrics{})
    
    priority := 1
    
    // Add optional handlers based on configuration
    if f.config.APIKeysEnabled {
        apiHandler := NewAPIKeyHandler(f.config.APIKeys, "X-API-Key")
        chain.AddHandlerWithConfig(apiHandler, handlers.HandlerConfig{
            Name:     "APIKeyHandler",
            Priority: priority,
            Enabled:  true,
        })
        priority++
    }
    
    if f.config.RateLimitEnabled {
        rateHandler := NewRateLimitHandler(f.config.RateLimit, 10)
        chain.AddHandlerWithConfig(rateHandler, handlers.HandlerConfig{
            Name:     "RateLimitHandler",
            Priority: priority,
            Enabled:  true,
        })
        priority++
    }
    
    if f.config.GraphQLEnabled {
        graphqlHandler := NewGraphQLHandler(f.config.GraphQLSchema)
        chain.AddHandlerWithConfig(graphqlHandler, handlers.HandlerConfig{
            Name:     "GraphQLHandler",
            Priority: priority,
            Enabled:  true,
        })
        priority++
    }
    
    // Add default Fir handlers
    f.addDefaultHandlers(chain, priority)
    
    return chain
}
```

### Option 3: Runtime Handler Addition

Add handlers to an existing chain at runtime:

```go
func ConfigureCustomHandlers(chain handlers.HandlerChain, config Config) {
    if config.EnableAuth {
        authHandler := NewAPIKeyHandler(config.APIKeys, "X-API-Key")
        chain.AddHandler(authHandler)
    }
    
    if config.EnableRateLimit {
        rateLimitHandler := NewRateLimitHandler(config.RateLimit, config.RateBurst)
        chain.AddHandler(rateLimitHandler)
    }
}

// Usage
chain := handlers.SetupDefaultHandlerChain(services)
ConfigureCustomHandlers(chain, myConfig)
```

## Handler Chain Composition

### Middleware Pattern

Handlers can implement middleware patterns by wrapping other handlers:

```go
type LoggingHandler struct {
    next   RequestHandler
    logger Logger
}

func NewLoggingHandler(next RequestHandler, logger Logger) *LoggingHandler {
    return &LoggingHandler{
        next:   next,
        logger: logger,
    }
}

func (h *LoggingHandler) HandlerName() string {
    return "LoggingHandler"
}

func (h *LoggingHandler) SupportsRequest(req *firHttp.RequestModel) bool {
    return h.next.SupportsRequest(req)
}

func (h *LoggingHandler) Handle(ctx context.Context, req *firHttp.RequestModel) (*firHttp.ResponseModel, error) {
    start := time.Now()
    
    h.logger.Info("Request started", "method", req.Method, "path", req.URL.Path)
    
    response, err := h.next.Handle(ctx, req)
    
    duration := time.Since(start)
    h.logger.Info("Request completed", 
        "method", req.Method, 
        "path", req.URL.Path,
        "status", response.StatusCode,
        "duration", duration,
    )
    
    return response, err
}
```

### Conditional Handlers

Handlers can implement complex conditional logic:

```go
type ConditionalHandler struct {
    condition func(*firHttp.RequestModel) bool
    handler   RequestHandler
}

func NewConditionalHandler(condition func(*firHttp.RequestModel) bool, handler RequestHandler) *ConditionalHandler {
    return &ConditionalHandler{
        condition: condition,
        handler:   handler,
    }
}

func (h *ConditionalHandler) SupportsRequest(req *firHttp.RequestModel) bool {
    return h.condition(req) && h.handler.SupportsRequest(req)
}

func (h *ConditionalHandler) Handle(ctx context.Context, req *firHttp.RequestModel) (*firHttp.ResponseModel, error) {
    if h.condition(req) {
        return h.handler.Handle(ctx, req)
    }
    return nil, nil // Continue to next handler
}

// Usage
adminOnlyHandler := NewConditionalHandler(
    func(req *firHttp.RequestModel) bool {
        return strings.HasPrefix(req.URL.Path, "/admin/")
    },
    NewAPIKeyHandler(adminAPIKeys, "X-Admin-Key"),
)
```

## Testing Custom Handlers

### Unit Testing

```go
func TestAPIKeyHandler_ValidKey(t *testing.T) {
    handler := NewAPIKeyHandler([]string{"valid-key"}, "X-API-Key")
    
    req := &firHttp.RequestModel{
        Method: "GET",
        URL:    &url.URL{Path: "/api/test"},
        Header: http.Header{
            "X-API-Key": []string{"valid-key"},
        },
    }
    
    response, err := handler.Handle(context.Background(), req)
    
    assert.NoError(t, err)
    assert.Equal(t, 200, response.StatusCode)
}

func TestAPIKeyHandler_InvalidKey(t *testing.T) {
    handler := NewAPIKeyHandler([]string{"valid-key"}, "X-API-Key")
    
    req := &firHttp.RequestModel{
        Method: "GET", 
        URL:    &url.URL{Path: "/api/test"},
        Header: http.Header{
            "X-API-Key": []string{"invalid-key"},
        },
    }
    
    response, err := handler.Handle(context.Background(), req)
    
    assert.NoError(t, err)
    assert.Equal(t, 403, response.StatusCode)
}
```

### Integration Testing

```go
func TestHandlerChain_WithCustomHandlers(t *testing.T) {
    // Create services
    services := createTestServices()
    
    // Create custom chain
    chain := handlers.NewPriorityHandlerChain(&MockLogger{}, &MockMetrics{})
    
    // Add custom handlers
    apiHandler := NewAPIKeyHandler([]string{"test-key"}, "X-API-Key")
    chain.AddHandler(apiHandler)
    
    // Add default handlers
    jsonHandler := handlers.NewJSONEventHandler(services.EventService, services.RenderService, services.ResponseBuilder, nil)
    chain.AddHandler(jsonHandler)
    
    // Test API endpoint with valid key
    req := &firHttp.RequestModel{
        Method: "POST",
        URL:    &url.URL{Path: "/api/event"},
        Header: http.Header{
            "X-API-Key":    []string{"test-key"},
            "Content-Type": []string{"application/json"},
        },
        Body: io.NopCloser(strings.NewReader(`{"event":"click"}`)),
    }
    
    response, err := chain.Handle(context.Background(), req)
    
    assert.NoError(t, err)
    assert.Equal(t, 200, response.StatusCode)
}
```

## Best Practices

### 1. Handler Ordering

Order handlers by priority:
1. **Authentication/Authorization** (Priority 1-5)
2. **Rate Limiting** (Priority 6-10)
3. **Content Processing** (Priority 11-20)
4. **Business Logic** (Priority 21+)

### 2. Error Handling

Always return proper HTTP responses for errors:

```go
func (h *MyHandler) Handle(ctx context.Context, req *firHttp.RequestModel) (*firHttp.ResponseModel, error) {
    if err := h.validate(req); err != nil {
        return &firHttp.ResponseModel{
            StatusCode: 400,
            Headers:    map[string]string{"Content-Type": "application/json"},
            Body:       []byte(fmt.Sprintf(`{"error": "%s"}`, err.Error())),
        }, nil // Return nil error to prevent handler chain from stopping
    }
    
    // Continue processing...
}
```

### 3. Resource Management

Clean up resources properly:

```go
func (h *DatabaseHandler) Handle(ctx context.Context, req *firHttp.RequestModel) (*firHttp.ResponseModel, error) {
    conn, err := h.pool.Get()
    if err != nil {
        return h.errorResponse("Database unavailable"), nil
    }
    defer h.pool.Put(conn) // Always return connection to pool
    
    // Use connection...
}
```

### 4. Context Usage

Use context for request-scoped data:

```go
func (h *AuthHandler) Handle(ctx context.Context, req *firHttp.RequestModel) (*firHttp.ResponseModel, error) {
    user, err := h.authenticate(req)
    if err != nil {
        return h.unauthorizedResponse(), nil
    }
    
    // Add user to context for downstream handlers
    ctx = context.WithValue(ctx, "user", user)
    
    return h.next.Handle(ctx, req)
}
```

### 5. Configuration

Make handlers configurable:

```go
type HandlerConfig struct {
    Enabled       bool
    Priority      int
    CustomOptions map[string]interface{}
}

func NewConfigurableHandler(config HandlerConfig) RequestHandler {
    if !config.Enabled {
        return &NoOpHandler{}
    }
    
    return &MyHandler{
        options: config.CustomOptions,
    }
}
```

This extension system provides powerful capabilities for customizing request processing while maintaining the framework's performance and reliability characteristics.
