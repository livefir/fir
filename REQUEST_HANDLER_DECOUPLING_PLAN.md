# Request Handler Decoupling Implementation Plan

## Overview

This plan implements a systematic decoupling of request handling from route implementation to improve testability, maintainability, and architectural clarity. The strategy follows the Single Responsibility Principle and Dependency Inversion Principle while maintaining full backward compatibility.

## Current Architecture Issues

- **Monolithic `ServeHTTP`**: Handles routing, parsing, business logic, and response generation
- **Direct HTTP Dependencies**: Business logic tightly coupled to `http.Request`/`http.ResponseWriter`
- **Embedded State**: Route struct mixes template management with request handling
- **Complex Context Creation**: Inconsistent `RouteContext` creation across handlers
- **No Unit Tests**: Only integration tests due to tight coupling
- **Mixed Concerns**: Transport, business logic, and presentation layers intertwined

## Target Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   HTTP Layer    │───▶│  Service Layer   │───▶│  Domain Layer   │
│  (Transport)    │    │ (Business Logic) │    │   (Pure Logic)  │
├─────────────────┤    ├──────────────────┤    ├─────────────────┤
│ • Request       │    │ • EventService   │    │ • Event Models  │
│   Handlers      │    │ • RenderService  │    │ • Error Types   │
│ • Response      │    │ • TemplateService│    │ • Domain Rules  │
│   Writers       │    │ • ValidationSvc  │    │ • Pure Functions│
│ • HTTP Adapters │    │ • Interfaces     │    │ • Value Objects │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

---

## 📋 MILESTONE 1: Request/Response Abstractions

**Goal**: Create transport-agnostic request/response models and interfaces

**Duration**: 1-2 days  
**Risk**: Low - Pure additive changes  

### Tasks

#### 1.1 Create Request Abstractions
- [ ] Create `internal/http/request.go` with request models:
  ```go
  type RequestType int
  const (
      RequestTypeJSON RequestType = iota
      RequestTypeForm
      RequestTypeGet
      RequestTypeWebSocket
  )
  
  type RequestModel struct {
      Type      RequestType
      Method    string
      Headers   map[string]string
      Body      []byte
      Query     map[string][]string
      Form      map[string][]string
      PathParams map[string]string
  }
  ```

#### 1.2 Create Response Abstractions  
- [ ] Create `internal/http/response.go` with response models:
  ```go
  type ResponseModel struct {
      StatusCode int
      Headers    map[string]string
      Body       []byte
      Events     []DOMEvent
  }
  
  type ResponseWriter interface {
      WriteResponse(ResponseModel) error
      WriteError(int, string) error
      Redirect(string, int) error
  }
  ```

#### 1.3 Add HTTP Adapters
- [ ] Create `internal/http/adapters.go`:
  ```go
  func FromHTTPRequest(r *http.Request) (*RequestModel, error)
  func ToHTTPResponse(w http.ResponseWriter, resp ResponseModel) error
  ```

#### 1.4 Unit Tests
- [ ] Test request model creation from various HTTP requests
- [ ] Test response writing with different status codes and content
- [ ] Test adapter error handling and edge cases

### Acceptance Criteria
- [ ] All adapters have 100% test coverage
- [ ] Request models support all current request types
- [ ] Response writing handles all current response scenarios
- [ ] No changes to existing route behavior
- [ ] `./scripts/pre-commit-check.sh --fast` passes (quick validation)
- [ ] Ready for commit via `./scripts/commit.sh` (full validation)

---

## 📋 MILESTONE 2: Event Processing Service Layer

**Goal**: Extract event processing logic into testable service layer

**Duration**: 2-3 days  
**Risk**: Medium - Touches core event handling  

### Tasks

#### 2.1 Create Event Processing Interfaces
- [ ] Create `internal/services/interfaces.go`:
  ```go
  type EventProcessor interface {
      ProcessEvent(ctx context.Context, req EventRequest) (*EventResponse, error)
  }
  
  type EventRequest struct {
      EventID    string
      RouteID    string
      Params     json.RawMessage
      IsForm     bool
      SessionID  *string
      Target     *string
  }
  
  type EventResponse struct {
      Data      interface{}
      State     interface{}
      Events    []pubsub.Event
      Errors    map[string]interface{}
      Redirect  *string
  }
  ```

#### 2.2 Implement Event Service
- [ ] Create `internal/services/event_service.go`:
  ```go
  type EventService struct {
      registry EventRegistry
      logger   Logger
  }
  
  func (s *EventService) ProcessEvent(ctx context.Context, req EventRequest) (*EventResponse, error)
  ```

#### 2.3 Extract Event Handler Logic
- [ ] Move event registry lookup logic to service
- [ ] Move error handling logic to service  
- [ ] Move result processing logic to service
- [ ] Maintain exact same behavior as current handlers

#### 2.4 Add Validation Service
- [ ] Create `internal/services/validation_service.go`:
  ```go
  type ValidationService interface {
      ValidateEvent(req EventRequest) error
      ValidateRouteAccess(routeID string) error
  }
  ```

#### 2.5 Unit Tests
- [ ] Test event processing with various event types
- [ ] Test error scenarios and error transformation
- [ ] Test validation rules and edge cases
- [ ] Mock all dependencies for isolated testing

### Acceptance Criteria
- [ ] Event processing fully extracted from HTTP handlers
- [ ] All current event handling behavior preserved
- [ ] Service layer has 90%+ test coverage
- [ ] Services can be tested without HTTP infrastructure
- [ ] `./scripts/pre-commit-check.sh --fast` passes (quick validation)
- [ ] Ready for commit via `./scripts/commit.sh` (full validation)

---

## 📋 MILESTONE 3: Template and Rendering Service Layer

**Goal**: Decouple template processing and response rendering from routes

**Duration**: 2-3 days  
**Risk**: Medium - Affects template engine integration  

### Tasks

#### 3.1 Create Rendering Interfaces
- [ ] Create `internal/services/render_interfaces.go`:
  ```go
  type RenderService interface {
      RenderTemplate(ctx RenderContext) (*RenderResult, error)
      RenderError(ctx ErrorContext) (*RenderResult, error)
      RenderEvents(events []pubsub.Event) ([]DOMEvent, error)
  }
  
  type RenderContext struct {
      RouteID      string
      Data         interface{}
      IsError      bool
      TemplateType TemplateType
  }
  
  type RenderResult struct {
      HTML   []byte
      Events []DOMEvent
  }
  ```

#### 3.2 Implement Template Service
- [ ] Create `internal/services/template_service.go`:
  ```go
  type TemplateService struct {
      engine    TemplateEngine
      cache     TemplateCache
      funcMap   template.FuncMap
  }
  
  func (s *TemplateService) LoadTemplate(config TemplateConfig) (*template.Template, error)
  func (s *TemplateService) RenderTemplate(tmpl *template.Template, data interface{}) ([]byte, error)
  ```

#### 3.3 Extract Rendering Logic
- [ ] Move template parsing from route to service
- [ ] Move template rendering from route to service
- [ ] Move event template extraction to service
- [ ] Maintain backward compatibility with current renderer interface

#### 3.4 Create Response Building Service
- [ ] Create `internal/services/response_builder.go`:
  ```go
  type ResponseBuilder interface {
      BuildEventResponse(result *EventResponse, request *RequestModel) (*ResponseModel, error)
      BuildTemplateResponse(render *RenderResult) (*ResponseModel, error)
      BuildErrorResponse(err error, code int) (*ResponseModel, error)
  }
  ```

#### 3.5 Unit Tests
- [ ] Test template loading with various configurations
- [ ] Test rendering with different data types and templates
- [ ] Test error template rendering
- [ ] Test response building for all response types
- [ ] Mock template engines for testing

### Acceptance Criteria
- [ ] Template processing extracted from route struct
- [ ] All rendering logic moved to services
- [ ] Template caching behavior preserved
- [ ] Custom template engines still supported
- [ ] Service layer has 90%+ test coverage
- [ ] `pre-commit-check.sh` passes

---

## 📋 MILESTONE 4: Request Handler Interfaces

**Goal**: Create pluggable request handlers with clear interfaces

**Duration**: 2-3 days  
**Risk**: Medium - Changes core request handling flow  

### Tasks

#### 4.1 Define Handler Interfaces
- [ ] Create `internal/handlers/interfaces.go`:
  ```go
  type RequestHandler interface {
      Handle(ctx context.Context, req *RequestModel) (*ResponseModel, error)
      SupportsRequest(req *RequestModel) bool
  }
  
  type HandlerChain interface {
      Handle(ctx context.Context, req *RequestModel) (*ResponseModel, error)
      AddHandler(handler RequestHandler)
  }
  ```

#### 4.2 Implement Specific Handlers
- [ ] Create `internal/handlers/json_event_handler.go`:
  ```go
  type JSONEventHandler struct {
      eventService EventService
      renderService RenderService
      responseBuilder ResponseBuilder
  }
  ```

- [ ] Create `internal/handlers/form_handler.go`:
  ```go
  type FormHandler struct {
      eventService EventService
      renderService RenderService
      responseBuilder ResponseBuilder
  }
  ```

- [ ] Create `internal/handlers/get_handler.go`:
  ```go
  type GetHandler struct {
      eventService EventService
      renderService RenderService
      responseBuilder ResponseBuilder
  }
  ```

- [ ] Create `internal/handlers/websocket_handler.go`:
  ```go
  type WebSocketHandler struct {
      wsServices WebSocketServices
  }
  ```

#### 4.3 Implement Handler Chain
- [ ] Create `internal/handlers/chain.go`:
  ```go
  type Chain struct {
      handlers []RequestHandler
      logger   Logger
  }
  
  func (c *Chain) Handle(ctx context.Context, req *RequestModel) (*ResponseModel, error)
  ```

#### 4.4 Extract Handler Logic
- [ ] Move JSON event handling logic to JSONEventHandler
- [ ] Move form handling logic to FormHandler
- [ ] Move GET request logic to GetHandler
- [ ] Move WebSocket logic to WebSocketHandler
- [ ] Preserve all current behavior and error handling

#### 4.5 Unit Tests
- [ ] Test each handler individually with mocked services
- [ ] Test handler chain routing and execution
- [ ] Test error propagation through handler chain
- [ ] Test that handlers correctly identify supported requests
- [ ] Test WebSocket handler integration

### Acceptance Criteria
- [ ] All request handling extracted to dedicated handlers
- [ ] Handler chain correctly routes requests to appropriate handlers
- [ ] Each handler can be unit tested independently
- [ ] All current request handling behavior preserved
- [ ] Handler interfaces support extensibility
- [ ] `pre-commit-check.sh` passes

---

## 📋 MILESTONE 5: Route Refactoring and Integration

**Goal**: Refactor route to use new handler chain while maintaining compatibility

**Duration**: 2-3 days  
**Risk**: High - Changes core route implementation  

### Tasks

#### 5.1 Create Route Service Factory
- [ ] Create `internal/route/factory.go`:
  ```go
  type RouteServiceFactory struct {
      services *routeservices.RouteServices
  }
  
  func (f *RouteServiceFactory) CreateEventService() EventService
  func (f *RouteServiceFactory) CreateRenderService() RenderService
  func (f *RouteServiceFactory) CreateHandlerChain() HandlerChain
  ```

#### 5.2 Refactor Route Constructor
- [ ] Update `newRoute()` to create service dependencies
- [ ] Initialize handler chain in route constructor
- [ ] Preserve all current route options and behavior

#### 5.3 Simplify ServeHTTP Method
- [ ] Refactor `ServeHTTP` to use handler chain:
  ```go
  func (rt *route) ServeHTTP(w http.ResponseWriter, r *http.Request) {
      timing := servertiming.FromContext(r.Context())
      defer timing.NewMetric("route").Start().Stop()
      
      req, err := FromHTTPRequest(r)
      if err != nil {
          http.Error(w, err.Error(), http.StatusBadRequest)
          return
      }
      
      resp, err := rt.handlerChain.Handle(r.Context(), req)
      if err != nil {
          http.Error(w, err.Error(), http.StatusInternalServerError)
          return
      }
      
      ToHTTPResponse(w, *resp)
  }
  ```

#### 5.4 Remove Old Handler Methods
- [ ] Remove `handleJSONEvent()` method
- [ ] Remove `handleFormPost()` method  
- [ ] Remove `handleGetRequest()` method
- [ ] Remove `handleWebSocketUpgrade()` method
- [ ] Remove helper methods (`parseFormEvent`, `determineFormAction`, etc.)

#### 5.5 Update RouteContext Creation
- [ ] Move RouteContext creation to handlers where needed
- [ ] Standardize context creation across all handlers
- [ ] Maintain compatibility with existing OnEventFunc signatures

#### 5.6 Integration Tests
- [ ] Test all existing examples still work
- [ ] Test WebSocket functionality preserved
- [ ] Test error handling scenarios
- [ ] Test template rendering with custom engines
- [ ] Run full e2e test suite

### Acceptance Criteria
- [ ] Route struct simplified to core responsibilities
- [ ] All request handling delegated to handler chain
- [ ] RouteContext creation consistent across handlers
- [ ] No breaking changes to public API
- [ ] All existing functionality preserved
- [ ] All examples and e2e tests pass
- [ ] `pre-commit-check.sh` passes

---

## 📋 MILESTONE 6: Legacy Cleanup and Documentation

**Goal**: Remove unused code and document new architecture

**Duration**: 1-2 days  
**Risk**: Low - Cleanup and documentation  

### Tasks

#### 6.1 Remove Unused Code
- [ ] Remove unused helper functions from route.go
- [ ] Remove unused result handling functions
- [ ] Clean up imports and dependencies
- [ ] Remove dead code identified by static analysis

#### 6.2 Update Architecture Documentation
- [ ] Update `ARCHITECTURE.md` with new request handling flow
- [ ] Document service layer interfaces and responsibilities
- [ ] Add examples of testing individual components
- [ ] Document migration guide for custom implementations

#### 6.3 Add Developer Documentation
- [ ] Create `internal/docs/TESTING_GUIDE.md` with unit testing examples
- [ ] Create `internal/docs/REQUEST_HANDLING.md` explaining new flow
- [ ] Create `internal/docs/SERVICE_LAYER.md` documenting service interfaces
- [ ] Create `internal/docs/HANDLER_EXTENSION.md` with examples for extending handlers
- [ ] Add code examples and best practices for each component

#### 6.4 Performance Optimization
- [ ] Profile new request handling flow
- [ ] Optimize handler chain routing if needed
- [ ] Ensure no performance regression from abstraction layers
- [ ] Add benchmarks for critical paths

#### 6.5 Final Validation
- [ ] Run complete test suite including e2e tests
- [ ] Validate all examples work correctly
- [ ] Check for any API compatibility issues
- [ ] Validate custom template engine examples

### Acceptance Criteria
- [ ] No unused code remains in codebase
- [ ] Architecture documentation fully updated
- [ ] Developer documentation complete with examples
- [ ] No performance regression from changes
- [ ] All tests pass including e2e test suite
- [ ] `pre-commit-check.sh` passes

---

## 🎯 Success Metrics

### Code Quality Metrics
- [ ] **Test Coverage**: Increase from 33% to 70%+
- [ ] **Unit Test Count**: Add 150+ new unit tests
- [ ] **Cyclomatic Complexity**: Reduce `ServeHTTP` complexity from 15+ to <5
- [ ] **Lines of Code**: Route struct methods reduced by 60%

### Testability Improvements
- [ ] **Service Layer**: 100% unit testable without HTTP
- [ ] **Handler Layer**: Individual handlers fully unit testable
- [ ] **Test Speed**: Unit tests run in <1s (vs current 30s+ integration tests)
- [ ] **Mock Support**: All dependencies mockable for isolated testing

### Architecture Quality
- [ ] **Separation of Concerns**: Clear boundaries between transport/business/domain
- [ ] **Dependency Inversion**: Services depend on interfaces, not implementations
- [ ] **Single Responsibility**: Each class/function has one clear purpose
- [ ] **Open/Closed**: Easy to extend handlers without modifying existing code

---

## 🚨 Risk Mitigation

### High-Risk Areas
1. **Route ServeHTTP Changes** (Milestone 5)
   - **Mitigation**: Extensive integration testing at each step
   - **Rollback**: Keep old methods as private fallbacks initially
   
2. **Template Engine Integration** (Milestone 3)
   - **Mitigation**: Test custom template engine examples thoroughly
   - **Rollback**: Maintain current template parsing as fallback

3. **RouteContext Changes** (Milestone 5)
   - **Mitigation**: Maintain exact same context fields and behavior
   - **Rollback**: Keep old context creation methods as utilities

### Testing Strategy

- Run `./scripts/pre-commit-check.sh --fast` for quick milestone validation (10-11s)
- Commit after each successful milestone using `./scripts/commit.sh` (full validation ~58s)
- Run full e2e test suite before finalizing each milestone
- Test examples manually after major changes

### Development Workflow

1. **Development iterations**: Use `./scripts/pre-commit-check.sh --fast` for rapid feedback
   - Excludes slow e2e tests for 75.8% speedup (11s vs 48s)
   - Includes core tests, static analysis, and build validation
   - Perfect for frequent testing during implementation

2. **Milestone validation**: Use `./scripts/commit.sh` for comprehensive validation
   - Runs full test suite including e2e tests
   - Includes coverage analysis and example compilation
   - Creates validated commits only after all quality gates pass

3. **Quality gates**: Both workflows include:
   - Build compilation (`go build ./...`)
   - Static analysis (`go vet`, `staticcheck`)
   - Go modules validation
   - Alpine.js plugin testing (if changes detected)

### Backward Compatibility

- All public APIs remain unchanged
- Existing route options continue to work
- Custom template engines continue to work
- OnEventFunc signatures unchanged
- RouteContext interface preserved

---

## 📅 Implementation Timeline

**Total Duration**: 12-16 days

| Milestone | Duration | Dependencies | Risk Level |
|-----------|----------|--------------|------------|
| 1. Request/Response Abstractions | 1-2 days | None | Low |
| 2. Event Processing Service | 2-3 days | Milestone 1 | Medium |
| 3. Template/Rendering Service | 2-3 days | Milestone 1 | Medium |
| 4. Request Handler Interfaces | 2-3 days | Milestones 2,3 | Medium |
| 5. Route Refactoring | 2-3 days | Milestone 4 | High |
| 6. Cleanup & Documentation | 1-2 days | Milestone 5 | Low |

**Checkpoints**:

- Commit after each milestone passes `pre-commit-check.sh`
- Full e2e test validation before Milestone 5
- Architecture review before Milestone 6

This plan ensures systematic decoupling while maintaining stability and backward compatibility throughout the process.
