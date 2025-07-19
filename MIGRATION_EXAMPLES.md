# Migration Examples

This document provides practical before/after examples and common migration patterns for transitioning from legacy Fir routes to the modern handler chain architecture.

## Before/After Code Examples

### Example 1: Simple GET Route

**Before (Legacy)**:
```go
func createLegacyRoute() *Route {
    route := &Route{
        id:     "user-profile",
        path:   "/users/{id}",
        method: "GET",
        template: "user_profile.html",
        templateData: UserData{},
        // Legacy fields
        onLoadHandler: func(ctx context.Context) (*LoadResult, error) {
            userID := ctx.Value("id").(string)
            user, err := getUserByID(userID)
            if err != nil {
                return nil, err
            }
            return &LoadResult{Data: user}, nil
        },
    }
    return route
}

// Legacy request handling (automatic)
func (r *Route) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    // Legacy handleGetRequest() called automatically
    // Limited extensibility, mixed concerns
}
```

**After (Modern)**:
```go
func createModernRoute() *Route {
    // 1. Define services
    services := &routeservices.RouteServices{
        EventService:    &eventservice.DefaultEventService{},
        RenderService:   &renderservice.DefaultRenderService{},
        TemplateService: &templateservice.GoTemplateService{},
        ResponseBuilder: &responsebuilder.DefaultResponseBuilder{},
    }
    
    // 2. Create route with modern interface
    route := &Route{
        id:       "user-profile",
        path:     "/users/{id}",
        method:   "GET",
        template: "user_profile.html",
        services: services,
        // Modern onLoad handler
        onLoadHandler: &OnLoadHandler{
            Handler: func(ctx context.Context) (*LoadResult, error) {
                userID := ctx.Value("id").(string)
                user, err := getUserByID(userID)
                if err != nil {
                    return nil, err
                }
                return &LoadResult{Data: user}, nil
            },
        },
    }
    
    return route
}

// Modern request handling (automatic)
func (r *Route) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    // Modern handler chain processes request:
    // 1. GetHandler.CanHandle() checks for onLoad
    // 2. GetHandler.Handle() processes with services
    // 3. Clean separation of concerns
}
```

**Key Changes**:
- Service injection for better testability
- Structured onLoad handler with proper interface
- Handler chain automatically selects GetHandler
- Clear service dependencies

### Example 2: JSON Event Handler

**Before (Legacy)**:
```go
func createLegacyAPIRoute() *Route {
    route := &Route{
        id:     "api-endpoint",
        path:   "/api/users",
        method: "POST",
        // Legacy event handlers (map[string]func)
        onEventHandlers: map[string]EventHandler{
            "create-user": func(ctx context.Context, event *Event) (*EventResponse, error) {
                // Parse JSON manually
                var userData UserData
                if err := json.Unmarshal(event.Data, &userData); err != nil {
                    return nil, err
                }
                
                // Business logic mixed with request handling
                user, err := createUser(userData)
                if err != nil {
                    return nil, err
                }
                
                // Manual response building
                return &EventResponse{
                    Success: true,
                    Data:    user,
                }, nil
            },
        },
    }
    return route
}

// Legacy processing (handleJSONEvent called automatically)
```

**After (Modern)**:
```go
func createModernAPIRoute() *Route {
    // 1. Extract business logic to service
    eventService := &eventservice.DefaultEventService{
        UserService: &userservice.DefaultUserService{},
    }
    
    services := &routeservices.RouteServices{
        EventService:    eventService,
        ResponseBuilder: &responsebuilder.JSONResponseBuilder{},
    }
    
    // 2. Clean event handler definition
    route := &Route{
        id:       "api-endpoint",
        path:     "/api/users",
        method:   "POST",
        services: services,
        onEventHandlers: map[string]EventHandler{
            "create-user": &UserCreationHandler{
                userService: eventService.UserService,
            },
        },
    }
    
    return route
}

// Separate handler implementation
type UserCreationHandler struct {
    userService UserService
}

func (h *UserCreationHandler) Handle(ctx context.Context, event *Event) (*EventResponse, error) {
    // Clean separation: only business logic here
    var userData UserData
    if err := event.UnmarshalData(&userData); err != nil {
        return nil, fmt.Errorf("invalid user data: %w", err)
    }
    
    user, err := h.userService.CreateUser(ctx, userData)
    if err != nil {
        return nil, fmt.Errorf("user creation failed: %w", err)
    }
    
    return &EventResponse{
        Success: true,
        Data:    user,
    }, nil
}

// Modern processing: JSONEventHandler automatically selected
```

**Key Changes**:
- Business logic extracted to dedicated service
- Clean handler interface implementation
- Service injection for dependencies
- Automatic JSON parsing and response building
- Better error handling and context propagation

### Example 3: Form Handler Migration

**Before (Legacy)**:
```go
func createLegacyFormRoute() *Route {
    route := &Route{
        id:     "contact-form",
        path:   "/contact",
        method: "POST",
        onEventHandlers: map[string]EventHandler{
            "submit-contact": func(ctx context.Context, event *Event) (*EventResponse, error) {
                // Manual form parsing
                formData := make(map[string]string)
                for key, value := range event.Data {
                    formData[key] = fmt.Sprintf("%v", value)
                }
                
                // Validation mixed with handler
                if formData["email"] == "" {
                    return &EventResponse{
                        Success: false,
                        Error:   "Email required",
                    }, nil
                }
                
                // Direct email service call
                err := sendEmail(formData["email"], formData["message"])
                if err != nil {
                    return &EventResponse{
                        Success: false,
                        Error:   "Email sending failed",
                    }, nil
                }
                
                // Manual redirect
                return &EventResponse{
                    Success:  true,
                    Redirect: &Redirect{URL: "/thank-you", Code: 302},
                }, nil
            },
        },
    }
    return route
}
```

**After (Modern)**:
```go
// 1. Define form data structure
type ContactForm struct {
    Email   string `json:"email" validate:"required,email"`
    Message string `json:"message" validate:"required,min=10"`
}

// 2. Create dedicated handler
type ContactFormHandler struct {
    emailService EmailService
    validator    Validator
}

func (h *ContactFormHandler) Handle(ctx context.Context, event *Event) (*EventResponse, error) {
    // Parse and validate
    var form ContactForm
    if err := event.UnmarshalData(&form); err != nil {
        return nil, fmt.Errorf("invalid form data: %w", err)
    }
    
    if err := h.validator.Validate(form); err != nil {
        return &EventResponse{
            Success: false,
            Error:   err.Error(),
        }, nil
    }
    
    // Business logic
    err := h.emailService.SendContactEmail(ctx, form.Email, form.Message)
    if err != nil {
        return nil, fmt.Errorf("email service failed: %w", err)
    }
    
    return &EventResponse{
        Success:  true,
        Redirect: &Redirect{URL: "/thank-you", Code: 302},
    }, nil
}

// 3. Modern route setup
func createModernFormRoute() *Route {
    emailService := &emailservice.DefaultEmailService{}
    validator := &validator.DefaultValidator{}
    
    services := &routeservices.RouteServices{
        EventService: &eventservice.DefaultEventService{},
        ResponseBuilder: &responsebuilder.DefaultResponseBuilder{},
    }
    
    route := &Route{
        id:       "contact-form",
        path:     "/contact",
        method:   "POST",
        services: services,
        onEventHandlers: map[string]EventHandler{
            "submit-contact": &ContactFormHandler{
                emailService: emailService,
                validator:    validator,
            },
        },
    }
    
    return route
}
```

**Key Changes**:
- Structured form data with validation tags
- Dedicated handler with injected dependencies
- Clean separation of validation, business logic, and response
- Better error handling with proper error wrapping

## Common Migration Patterns

### Pattern 1: Service Extraction

**Problem**: Legacy handlers mix request handling with business logic

**Solution**: Extract business logic to dedicated services

```go
// Before: Mixed concerns
func legacyHandler(ctx context.Context, event *Event) (*EventResponse, error) {
    // Request parsing
    data := parseEventData(event)
    
    // Validation
    if !isValid(data) {
        return errorResponse("Invalid data")
    }
    
    // Database operations
    result, err := db.Query("SELECT ...", data.ID)
    if err != nil {
        return errorResponse("Database error")
    }
    
    // Business logic
    processedResult := processData(result)
    
    // Response building
    return &EventResponse{Data: processedResult}, nil
}

// After: Clean separation
type BusinessService struct {
    db       Database
    validator Validator
}

func (s *BusinessService) ProcessRequest(ctx context.Context, data RequestData) (*ProcessResult, error) {
    // Pure business logic
    if err := s.validator.Validate(data); err != nil {
        return nil, err
    }
    
    result, err := s.db.GetData(ctx, data.ID)
    if err != nil {
        return nil, err
    }
    
    return s.processData(result), nil
}

func modernHandler(ctx context.Context, event *Event) (*EventResponse, error) {
    // Only request/response handling
    var data RequestData
    if err := event.UnmarshalData(&data); err != nil {
        return nil, err
    }
    
    result, err := businessService.ProcessRequest(ctx, data)
    if err != nil {
        return nil, err
    }
    
    return &EventResponse{Data: result}, nil
}
```

### Pattern 2: Handler Interface Implementation

**Problem**: Function-based handlers are hard to test and extend

**Solution**: Implement proper handler interfaces

```go
// Before: Function-based
onEventHandlers: map[string]EventHandler{
    "action": func(ctx context.Context, event *Event) (*EventResponse, error) {
        // Handler logic
    },
}

// After: Interface-based
type ActionHandler struct {
    service ActionService
    logger  Logger
}

func (h *ActionHandler) Handle(ctx context.Context, event *Event) (*EventResponse, error) {
    h.logger.Info("Processing action", "event", event.ID)
    
    result, err := h.service.ProcessAction(ctx, event)
    if err != nil {
        h.logger.Error("Action failed", "error", err)
        return nil, err
    }
    
    return result, nil
}

// Register handler
onEventHandlers: map[string]EventHandler{
    "action": &ActionHandler{
        service: actionService,
        logger:  logger,
    },
}
```

### Pattern 3: Configuration Migration

**Problem**: Route configuration scattered across initialization

**Solution**: Centralized service configuration

```go
// Before: Scattered configuration
func createRoute() *Route {
    route := &Route{
        id: "example",
        // Various fields set individually
    }
    
    // Separate service setup
    route.setupEventHandlers()
    route.configureTemplates()
    route.setupWebSocket()
    
    return route
}

// After: Centralized configuration
func createRoute() *Route {
    // 1. Service configuration
    services := &routeservices.RouteServices{
        EventService:    createEventService(),
        RenderService:   createRenderService(),
        TemplateService: createTemplateService(),
        ResponseBuilder: createResponseBuilder(),
        Options: &routeservices.Options{
            DisableTemplateCache: false,
            DisableWebsocket:     false,
        },
    }
    
    // 2. Handler configuration
    handlers := map[string]EventHandler{
        "action1": &Action1Handler{},
        "action2": &Action2Handler{},
    }
    
    // 3. Route creation
    return &Route{
        id:              "example",
        services:        services,
        onEventHandlers: handlers,
        onLoadHandler:   &OnLoadHandler{},
    }
}

func createEventService() EventService {
    return &eventservice.DefaultEventService{
        Validator: &validator.DefaultValidator{},
        Logger:    &logger.DefaultLogger{},
    }
}
```

## Testing Migration

### Before: Legacy Testing

```go
func TestLegacyRoute(t *testing.T) {
    route := createLegacyRoute()
    
    // Test through internal methods (brittle)
    result, err := route.handleJSONEvent(context.Background(), event)
    if err != nil {
        t.Fatal(err)
    }
    
    // Manual assertions
    if result.Success != true {
        t.Error("Expected success")
    }
}
```

### After: Modern Testing

```go
func TestModernRoute(t *testing.T) {
    // 1. Create mock services
    services := &routeservices.RouteServices{
        EventService: &MockEventService{
            ProcessEventFunc: func(ctx context.Context, route RouteInterface, event *Event) (*EventResponse, error) {
                return &EventResponse{Success: true}, nil
            },
        },
        ResponseBuilder: &MockResponseBuilder{},
    }
    
    // 2. Create route with mocks
    route := createRouteWithServices(services)
    
    // 3. Test through public API
    req := createJSONEventRequest()
    w := httptest.NewRecorder()
    
    route.ServeHTTP(w, req)
    
    // 4. Assert response
    assert.Equal(t, http.StatusOK, w.Code)
    assert.Contains(t, w.Body.String(), "success")
}

func TestHandlerDirectly(t *testing.T) {
    // Test handler in isolation
    handler := &ActionHandler{
        service: &MockActionService{},
    }
    
    result, err := handler.Handle(context.Background(), &Event{
        ID: "test-action",
        Data: map[string]interface{}{"key": "value"},
    })
    
    assert.NoError(t, err)
    assert.True(t, result.Success)
}
```

## Troubleshooting Migration

### Issue 1: Handler Not Found

**Symptom**: Requests fall back to legacy system unexpectedly

**Diagnosis**:
```go
// Check handler chain coverage
req := createTestRequest()
route := getRoute()

coverage := route.canHandlerChainHandle(req)
if !coverage.CanHandle {
    // Check specific reasons
    for _, reason := range coverage.Reasons {
        log.Printf("Coverage issue: %s", reason)
    }
}
```

**Solution**: Verify handler registration and requirements

### Issue 2: Service Dependencies

**Symptom**: Nil pointer errors when calling services

**Diagnosis**:
```go
services := route.GetServices()
if services.EventService == nil {
    log.Error("EventService not configured")
}
```

**Solution**: Ensure complete service configuration

### Issue 3: Handler Priority Conflicts

**Symptom**: Wrong handler processes requests

**Diagnosis**:
```go
// Check handler priorities
for _, handler := range handlerChain {
    log.Printf("Handler: %T, Priority: %d", handler, handler.Priority())
}
```

**Solution**: Adjust handler priorities or CanHandle logic

## Performance Considerations

### Handler Chain Overhead

The modern handler chain adds minimal overhead:

```go
// Benchmark comparison
func BenchmarkLegacyRequest(b *testing.B) {
    route := createLegacyRoute()
    req := createTestRequest()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        w := httptest.NewRecorder()
        route.ServeHTTP(w, req)
    }
}

func BenchmarkModernRequest(b *testing.B) {
    route := createModernRoute()
    req := createTestRequest()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        w := httptest.NewRecorder()
        route.ServeHTTP(w, req)
    }
}

// Typical results:
// BenchmarkLegacyRequest-8    10000    150000 ns/op
// BenchmarkModernRequest-8    10000    152000 ns/op
// Overhead: ~1.3% (2μs per request)
```

### Memory Usage

```go
// Memory profile comparison
func TestMemoryUsage(t *testing.T) {
    var m1, m2 runtime.MemStats
    
    // Legacy route memory
    runtime.GC()
    runtime.ReadMemStats(&m1)
    legacyRoute := createLegacyRoute()
    runtime.ReadMemStats(&m2)
    legacyMem := m2.TotalAlloc - m1.TotalAlloc
    
    // Modern route memory
    runtime.GC()
    runtime.ReadMemStats(&m1)
    modernRoute := createModernRoute()
    runtime.ReadMemStats(&m2)
    modernMem := m2.TotalAlloc - m1.TotalAlloc
    
    t.Logf("Legacy route memory: %d bytes", legacyMem)
    t.Logf("Modern route memory: %d bytes", modernMem)
    // Typical: Modern uses ~500 bytes more per route
}
```

## Next Steps

After completing migration:

1. **Run Tests**: Ensure all functionality preserved
   ```bash
   go test ./... -v
   ```

2. **Performance Testing**: Validate performance impact
   ```bash
   go test -bench=. -benchmem
   ```

3. **Coverage Analysis**: Verify handler chain coverage
   ```bash
   go test -run TestHandlerChainCoverage -v
   ```

4. **Pre-commit Validation**: Run quality checks
   ```bash
   ./scripts/pre-commit-check.sh
   ```

5. **Documentation**: Update route-specific documentation

For more migration guidance, see `MIGRATION_GUIDE.md` and `API_DOCS.md`.
