# Fir Framework Architecture Documentation

*Following the arc42 template*

## Table of Contents

1. [Introduction and Goals](#1-introduction-and-goals)
2. [Constraints](#2-constraints)
3. [Context and Scope](#3-context-and-scope)
4. [Solution Strategy](#4-solution-strategy)
5. [Building Block View](#5-building-block-view)
6. [Runtime View](#6-runtime-view)
7. [Deployment View](#7-deployment-view)
8. [Crosscutting Concepts](#8-crosscutting-concepts)
9. [Architecture Decisions](#9-architecture-decisions)
10. [Quality](#10-quality)
11. [Risks and Technical Debt](#11-risks-and-technical-debt)
12. [Glossary](#12-glossary)

---

## 1. Introduction and Goals

### 1.1 Requirements Overview

**What is Fir?**

Fir is a Go toolkit for building reactive web interfaces that combines Go's `html/template` with Alpine.js for DOM manipulation. It enables Go developers to create interactive web applications without mastering complex JavaScript frameworks.

**Key Requirements:**

- Server-side HTML rendering using Go templates
- Real-time DOM updates without full page reloads
- WebSocket-based communication with HTTP fallback
- Progressive enhancement approach
- Event-driven architecture
- Pub/sub system for broadcasting updates

### 1.2 Quality Goals

| Priority | Quality Goal | Scenario |
|----------|-------------|----------|
| 1 | **Developer Experience** | Go developers with basic HTML/CSS/JS skills can build reactive UIs efficiently |
| 2 | **Performance** | Real-time updates with minimal latency, efficient DOM patching |
| 3 | **Reliability** | Graceful fallback from WebSocket to HTTP, connection resilience |
| 4 | **Maintainability** | Clean separation between server logic and client presentation |
| 5 | **Scalability** | Support for multiple concurrent connections and pub/sub broadcasting |

### 1.3 Stakeholders

| Role | Contact | Expectations |
|------|---------|-------------|
| **Go Developers** | Primary users | Simple API, familiar patterns, minimal JavaScript required |
| **Frontend Developers** | Secondary users | Standard HTML/CSS/JS integration, Alpine.js compatibility |
| **DevOps Engineers** | Deployment | Easy deployment, monitoring capabilities, performance metrics |
| **End Users** | Web browsers | Fast, responsive interfaces with real-time updates |

---

## 2. Constraints

### 2.1 Technical Constraints

| Constraint | Description | Impact |
|------------|-------------|---------|
| **Go Language** | Framework implemented in Go 1.24+ | Leverages Go's template system and concurrency |
| **Alpine.js Dependency** | Client-side requires Alpine.js framework | Provides reactive DOM manipulation |
| **WebSocket Protocol** | Real-time communication via WebSockets | Enables bidirectional communication |
| **HTML Template System** | Uses Go's `html/template` package | Server-side rendering with type safety |

### 2.2 Organizational Constraints

| Constraint | Description |
|------------|-------------|
| **Open Source** | MIT licensed, community-driven development |
| **Backward Compatibility** | Must maintain API stability for existing applications |
| **Documentation** | Comprehensive docs required for adoption |

### 2.3 Conventions

- RESTful HTTP patterns for fallback communication
- Standard WebSocket message formats
- Alpine.js directive naming conventions (`x-fir-*`)
- Go package structure following standard conventions

---

## 3. Context and Scope

### 3.1 Business Context

```mermaid
graph TB
    Dev[Go Developer] --> Fir[Fir Framework]
    Fir --> App[Web Application]
    App --> Browser[User Browser]
    Browser --> Alpine[Alpine.js Plugin]
    Alpine --> Fir
    
    DB[(Database)] --> App
    Cache[(Cache)] --> App
```

**External Entities:**
- **Go Developers**: Create web applications using Fir
- **Web Browsers**: Execute client-side Alpine.js plugin
- **Databases**: Persist application data
- **Cache Systems**: Store session and template data

### 3.2 Technical Context

```mermaid
graph LR
    Browser[Web Browser] <--> |WebSocket/HTTP| Controller[Fir Controller]
    Controller <--> |Templates| Templates[HTML Templates]
    Controller <--> |Events| EventRegistry[Event Registry]
    Controller <--> |Pub/Sub| PubSub[Pub/Sub System]
    Controller <--> |Session| Session[Session Management]
    
    Alpine[Alpine.js Plugin] --> Browser
    Templates --> Renderer[Template Renderer]
```

**Technical Interfaces:**
- **HTTP/WebSocket**: Client-server communication
- **Template System**: Go html/template integration
- **Pub/Sub**: Redis or in-memory event broadcasting
- **Session Management**: Secure cookie-based sessions

---

## 4. Solution Strategy

### 4.1 Technology Decisions

| Area | Decision | Rationale |
|------|----------|-----------|
| **Server Language** | Go | Type safety, concurrency, template system |
| **Client Framework** | Alpine.js | Lightweight, declarative, easy integration |
| **Communication** | WebSocket + HTTP fallback | Real-time with reliability |
| **Templating** | Go html/template | Server-side rendering, XSS protection |
| **State Management** | Server-side with event sourcing | Simplified client, consistent state |

### 4.2 Architectural Patterns

- **Event-Driven Architecture**: User and server events drive application flow
- **Pub/Sub Pattern**: Broadcast updates to multiple clients
- **Progressive Enhancement**: Start with HTML, enhance with JavaScript
- **Template-Based Rendering**: Server-side HTML generation
- **Connection Pooling**: Efficient WebSocket connection management

### 4.3 Core Design Principles

1. **Simplicity**: Minimal learning curve for Go developers
2. **Standards Compliance**: Use web standards where possible
3. **Progressive Enhancement**: Works without JavaScript, better with it
4. **Type Safety**: Leverage Go's type system for templates and events
5. **Performance**: Efficient rendering and minimal client-side code

---

## 5. Building Block View

### 5.1 Whitebox Overall System

```mermaid
graph TB
    subgraph "Fir Framework"
        Controller[Controller]
        Route[Route]
        Connection[Connection]
        EventRegistry[Event Registry]
        Renderer[Renderer]
        PubSub[Pub/Sub Adapter]
    end
    
    subgraph "Client"
        Alpine[Alpine.js Plugin]
        WebSocket[WebSocket Client]
        EventDispatcher[Event Dispatcher]
    end
    
    Controller --> Route
    Controller --> EventRegistry
    Route --> Renderer
    Route --> Connection
    Connection --> PubSub
    
    Alpine --> WebSocket
    Alpine --> EventDispatcher
    WebSocket <--> Connection
```

**Component Responsibilities:**

| Component | Responsibility |
|-----------|---------------|
| **Controller** | Main framework entry point, route management, WebSocket handling |
| **Route** | HTTP request handling, template rendering, event processing |
| **Connection** | WebSocket connection management, message handling |
| **Event Registry** | Event handler registration and lookup |
| **Renderer** | Template execution and DOM event generation |
| **Pub/Sub Adapter** | Message broadcasting (Redis/in-memory) |
| **Alpine.js Plugin** | Client-side DOM manipulation and event handling |

### 5.2 Level 2 - Controller Subsystem

```mermaid
graph TB
    subgraph "Controller"
        RouteFunc[RouteFunc]
        WebSocketHandler[WebSocket Handler]
        SessionManager[Session Manager]
        EventRegistry[Event Registry]
        Config[Configuration Options]
    end
    
    HTTP[HTTP Request] --> RouteFunc
    WebSocket[WebSocket Upgrade] --> WebSocketHandler
    RouteFunc --> SessionManager
    WebSocketHandler --> SessionManager
    RouteFunc --> EventRegistry
```

### 5.3 Level 2 - Route Subsystem

```mermaid
graph TB
    subgraph "Route"
        TemplateParser[Template Parser]
        EventHandler[Event Handler]
        Renderer[Renderer]
        OnLoad[OnLoad Handler]
        OnEvent[OnEvent Handlers]
    end
    
    Request[HTTP Request] --> EventHandler
    EventHandler --> OnLoad
    EventHandler --> OnEvent
    OnEvent --> Renderer
    TemplateParser --> Renderer
```

### 5.4 Level 2 - Client Plugin

```mermaid
graph TB
    subgraph "Alpine.js Plugin"
        WebSocketClient[WebSocket Client]
        EventDispatcher[Event Dispatcher]
        DOMUtils[DOM Utilities]
        MagicFunctions[Magic Functions]
        Directives[Fir Directives]
    end
    
    ServerEvents[Server Events] --> EventDispatcher
    EventDispatcher --> DOMUtils
    EventDispatcher --> MagicFunctions
    UserEvents[User Events] --> WebSocketClient
    DOMUtils --> Directives
```

---

## 5.5. Low-Level Design

### 5.5.1 Controller Component

**Class Structure:**
```go
type controller struct {
    routes         map[string]*route
    eventRegistry  event.EventRegistry
    opt           *opt
    secureCookie  *securecookie.SecureCookie
    pubsub        pubsub.Adapter
    cache         *cache.Cache
    websocketUpgrader websocket.Upgrader
}

type opt struct {
    onSocketConnect    func(userOrSessionID string) error
    onSocketDisconnect func(userOrSessionID string)
    channelFunc        func(r *http.Request, viewID string) *string
    pathParamsFunc     func(r *http.Request) PathParams
    disableWebsocket   bool
    publicDir          string
    formDecoder        *schema.Decoder
    funcMap            template.FuncMap
    renderer           Renderer
}
```

**Key Methods:**
- `NewController(appName string, options ...ControllerOption) Controller`
- `RouteFunc(opts RouteFunc) http.HandlerFunc`
- `GetEventRegistry() event.EventRegistry`

**Responsibilities:**
- Route registration and management
- WebSocket connection handling
- Event registry management
- Session management
- Configuration management

### 5.5.2 Route Component

**Class Structure:**
```go
type route struct {
    template       *template.Template
    errorTemplate  *template.Template
    eventTemplates eventTemplates
    renderer       Renderer
    cntrl          *controller
    routeOpt
    sync.RWMutex
}

type routeOpt struct {
    id            string
    layoutPath    string
    templatePath  string
    onLoad        OnLoadFunc
    onEvents      map[string]OnEventFunc
    onForm        OnFormFunc
    formStruct    any
    pathParams    PathParams
    beforeRender  BeforeRenderFunc
}
```

**Key Methods:**
- `ServeHTTP(w http.ResponseWriter, r *http.Request)`
- `parseTemplates() error`
- `handleJSONEvent(w http.ResponseWriter, r *http.Request)`
- `handleWebSocketUpgrade(w http.ResponseWriter, r *http.Request)`
- `setTemplate(t *template.Template)`
- `getTemplate() *template.Template`

**Template Management:**
- Thread-safe template storage with RWMutex
- Event-specific template extraction and caching
- Dynamic template parsing and compilation
- Template function map injection

### 5.5.3 Connection Component

**Class Structure:**
```go
type Connection struct {
    conn          *websocket.Conn
    controller    *controller
    request       *http.Request
    response      http.ResponseWriter
    sessionID     string
    routeID       string
    user          string
    send          chan []byte
    writePumpDone chan struct{}
    lastEvent     Event
    ctx           context.Context
    cancel        context.CancelFunc
    mu            sync.Mutex
    closed        bool
}
```

**Key Methods:**
- `NewConnection(w http.ResponseWriter, r *http.Request, cntrl *controller) (*Connection, error)`
- `Upgrade() error`
- `StartPubSubListeners() error`
- `ReadLoop()`
- `StartWritePump()`
- `handleMessage(message []byte) error`
- `processEvent(event Event, eventRouteID string)`
- `Close()`

**Connection Lifecycle:**
1. **Creation**: Validate session, extract user info
2. **Upgrade**: HTTP to WebSocket protocol upgrade
3. **Authentication**: Verify session and route authorization
4. **Event Loop**: Handle incoming messages and pub/sub events
5. **Cleanup**: Graceful connection termination

**Message Processing:**
- JSON message parsing and validation
- Duplicate event detection and filtering
- Session authorization verification
- Event routing to appropriate handlers

### 5.5.4 Event Registry Component

**Interface:**
```go
type EventRegistry interface {
    Register(routeID, eventID string, handler interface{}) error
    Get(routeID, eventID string) (interface{}, bool)
    GetRouteEvents(routeID string) map[string]interface{}
    GetAllEvents() map[string]map[string]interface{}
}
```

**Implementation:**
```go
type eventRegistry struct {
    events map[string]map[string]interface{}
    mu     sync.RWMutex
}
```

**Key Features:**
- Thread-safe event handler storage
- Route-scoped event organization
- Handler type validation
- Debug introspection support

### 5.5.5 Renderer Component

**Interface:**
```go
type Renderer interface {
    RenderRoute(ctx RouteContext, data routeData, useErrorTemplate bool) error
    RenderDOMEvents(ctx RouteContext, pubsubEvent pubsub.Event) []dom.Event
}
```

**Template Renderer Implementation:**
```go
type TemplateRenderer struct{}

func (tr *TemplateRenderer) RenderRoute(ctx RouteContext, data routeData, useErrorTemplate bool) error
func (tr *TemplateRenderer) RenderDOMEvents(ctx RouteContext, pubsubEvent pubsub.Event) []dom.Event
```

**Rendering Pipeline:**
1. **Template Selection**: Choose main or error template
2. **Function Map Injection**: Add Fir-specific template functions
3. **Template Execution**: Render with provided data
4. **DOM Event Generation**: Extract and process event templates
5. **Response Writing**: Send HTML or JSON response

### 5.5.6 Pub/Sub Adapter

**Interface:**
```go
type Adapter interface {
    Publish(channel string, event Event) error
    Subscribe(channel string, handler func(Event)) error
    Unsubscribe(channel string) error
    Close() error
}
```

**Redis Implementation:**
```go
type redisAdapter struct {
    client *redis.Client
    pubsub *redis.PubSub
    subscriptions map[string]func(Event)
}
```

**In-Memory Implementation:**
```go
type inMemoryAdapter struct {
    channels map[string][]func(Event)
    mu       sync.RWMutex
}
```

**Features:**
- Pluggable adapter architecture
- Channel-based message routing
- Automatic subscription management
- Graceful error handling

### 5.5.7 Alpine.js Plugin Architecture

**Core Structure:**
```javascript
// Main plugin entry point
export default function firPlugin(Alpine) {
    // WebSocket management
    let ws = null;
    
    // Event dispatcher
    const eventDispatcher = new EventDispatcher();
    
    // Magic functions
    Alpine.magic('fir', () => ({
        submit: (form) => submitForm(form),
        replace: () => domAction('replace'),
        append: () => domAction('append'),
        prepend: () => domAction('prepend')
    }));
    
    // Directives
    Alpine.directive('fir-replace', firReplaceDirective);
    Alpine.directive('fir-append', firAppendDirective);
}
```

**WebSocket Client:**
```javascript
class WebSocketClient {
    constructor(url, options, eventCallback) {
        this.url = url;
        this.options = options;
        this.eventCallback = eventCallback;
        this.reconnectAttempts = 0;
        this.socket = null;
    }
    
    connect() { /* Connection logic */ }
    send(data) { /* Send message */ }
    close() { /* Cleanup */ }
    handleReconnect() { /* Reconnection logic */ }
}
```

**Event Dispatcher:**
```javascript
class EventDispatcher {
    processServerEvents(events) {
        events.forEach(event => {
            this.dispatchEvent(event);
        });
    }
    
    dispatchEvent(serverEvent) {
        const customEvent = this.createCustomEvent(serverEvent);
        this.targetElement(serverEvent.target, customEvent);
    }
    
    createCustomEvent(serverEvent) { /* Event creation */ }
    targetElement(selector, event) { /* Element targeting */ }
}
```

### 5.5.8 Session Management

**Session Structure:**
```go
type Session struct {
    ID      string
    RouteID string
    UserID  string
    Data    map[string]interface{}
    Expires time.Time
}
```

**Cookie Encoding:**
```go
func encodeSession(opt routeOpt, w http.ResponseWriter, r *http.Request) error {
    sessionData := fmt.Sprintf("%s:%s", sessionID, routeID)
    encoded, err := opt.secureCookie.Encode(opt.cookieName, sessionData)
    if err != nil {
        return err
    }
    
    cookie := &http.Cookie{
        Name:     opt.cookieName,
        Value:    encoded,
        Path:     "/",
        HttpOnly: true,
        Secure:   isHTTPS(r),
        SameSite: http.SameSiteStrictMode,
    }
    
    http.SetCookie(w, cookie)
    return nil
}
```

**Security Features:**
- Cryptographically signed cookies
- Session expiration handling
- CSRF protection
- Secure flag for HTTPS
- HttpOnly to prevent XSS

### 5.5.9 Template System Integration

**Template Function Map:**
```go
func newFirFuncMap(ctx RouteContext, errs map[string]any) template.FuncMap {
    return template.FuncMap{
        "fir": func(parts ...string) *string {
            return firattr.GetClassName(strings.Join(parts, ":"))
        },
        "Error": func(key string) string {
            if errs == nil {
                return ""
            }
            if err, ok := errs[key]; ok {
                return fmt.Sprintf("%v", err)
            }
            return ""
        },
        "HasError": func(key string) bool {
            if errs == nil {
                return false
            }
            _, ok := errs[key]
            return ok
        },
    }
}
```

**Event Template Extraction:**
```go
type eventTemplates map[string]map[string]*template.Template

func extractEventTemplates(tmpl *template.Template) eventTemplates {
    templates := make(eventTemplates)
    
    for _, t := range tmpl.Templates() {
        if matches := eventTemplateRegex.FindAllString(t.Name(), -1); matches != nil {
            eventID := extractEventID(t.Name())
            state := extractState(t.Name())
            
            if templates[eventID] == nil {
                templates[eventID] = make(map[string]*template.Template)
            }
            templates[eventID][state] = t
        }
    }
    
    return templates
}
```

### 5.5.10 Error Handling Strategy

**Error Types:**
```go
type FirError struct {
    Type    ErrorType
    Message string
    Cause   error
    Context map[string]interface{}
}

type ErrorType int

const (
    ErrorTypeTemplate ErrorType = iota
    ErrorTypeEvent
    ErrorTypeConnection
    ErrorTypeValidation
    ErrorTypeAuth
)
```

**Error Recovery:**
```go
func handleEventError(err error, ctx RouteContext) *pubsub.Event {
    if err == nil {
        return nil
    }
    
    // Convert error to appropriate event state
    errorEvent := &pubsub.Event{
        ID:    &ctx.event.ID,
        State: eventstate.Error,
        Detail: &pubsub.Detail{
            Data: map[string]any{
                "error": err.Error(),
            },
        },
    }
    
    return errorEvent
}
```

---

## 6. Runtime View

### 6.1 Page Load Scenario

```mermaid
sequenceDiagram
    participant Browser
    participant Route
    participant Controller
    participant Templates
    participant Alpine
    
    Browser->>Route: GET /page
    Route->>Controller: Handle request
    Controller->>Route: Create route context
    Route->>Templates: Parse templates
    Route->>Route: Execute OnLoad handler
    Route->>Templates: Render HTML
    Templates-->>Browser: HTML response + session cookie
    Browser->>Alpine: Initialize Alpine.js
    Alpine->>Browser: Set up WebSocket connection
```

### 6.2 WebSocket Event Processing

```mermaid
sequenceDiagram
    participant Alpine
    participant Connection
    participant EventRegistry
    participant Route
    participant PubSub
    participant OtherClients
    
    Alpine->>Connection: Send user event (JSON)
    Connection->>Connection: Validate session
    Connection->>EventRegistry: Lookup event handler
    EventRegistry->>Route: Execute OnEvent handler
    Route->>Route: Process business logic
    Route->>PubSub: Publish result event
    PubSub->>Connection: Broadcast to subscribers
    Connection->>Alpine: Send DOM events
    PubSub->>OtherClients: Broadcast to other sessions
```

### 6.3 HTTP Fallback Scenario

```mermaid
sequenceDiagram
    participant Alpine
    participant Route
    participant EventHandler
    participant Templates
    
    Alpine->>Route: POST /page (event JSON)
    Route->>EventHandler: Process event
    EventHandler->>EventHandler: Execute business logic
    EventHandler->>Templates: Render updated templates
    Templates-->>Alpine: DOM events (JSON)
    Alpine->>Alpine: Apply DOM updates
```

### 6.4 Template Rendering Flow

```mermaid
sequenceDiagram
    participant EventHandler
    participant Renderer
    participant Templates
    participant FuncMap
    participant Connection
    
    EventHandler->>Renderer: Render DOM events
    Renderer->>Templates: Get event templates
    Renderer->>FuncMap: Create context functions
    Renderer->>Templates: Execute template with data
    Templates-->>Renderer: Generated HTML
    Renderer->>Renderer: Create DOM events
    Renderer-->>Connection: DOM events array
```

---

## 7. Deployment View

### 7.1 Development Environment

```mermaid
graph TB
    subgraph "Developer Machine"
        IDE[Go IDE/Editor]
        App[Fir Application]
        Browser[Web Browser]
        FileWatcher[File Watcher]
    end
    
    IDE --> App
    App --> Browser
    FileWatcher --> App
    App -.->|Live Reload| Browser
```

### 7.2 Production Environment

```mermaid
graph TB
    subgraph "Load Balancer"
        LB[Load Balancer]
    end
    
    subgraph "Application Servers"
        App1[Fir App Instance 1]
        App2[Fir App Instance 2]
        App3[Fir App Instance N]
    end
    
    subgraph "Shared Services"
        Redis[(Redis Pub/Sub)]
        DB[(Database)]
        Cache[(Cache)]
    end
    
    LB --> App1
    LB --> App2
    LB --> App3
    
    App1 --> Redis
    App2 --> Redis
    App3 --> Redis
    
    App1 --> DB
    App2 --> DB
    App3 --> DB
    
    App1 --> Cache
    App2 --> Cache
    App3 --> Cache
```

### 7.3 Container Deployment

```mermaid
graph TB
    subgraph "Container Runtime"
        subgraph "App Container"
            FirApp[Fir Application]
            AlpinePlugin[Alpine.js Plugin Assets]
        end
        
        subgraph "Redis Container"
            RedisServer[Redis Server]
        end
        
        subgraph "Database Container"
            PostgreSQL[PostgreSQL]
        end
    end
    
    FirApp --> RedisServer
    FirApp --> PostgreSQL
```

**Deployment Artifacts:**
- **Fir Application Binary**: Compiled Go executable
- **Alpine.js Plugin**: JavaScript bundle (CDN or local)
- **HTML Templates**: Template files (embedded or external)
- **Static Assets**: CSS, images, fonts
- **Configuration**: Environment-specific settings

---

## 8. Crosscutting Concepts

### 8.1 Security

**Session Management:**
- Secure cookie-based sessions with CSRF protection
- Session validation for WebSocket connections
- Automatic session renewal

**Input Validation:**
- Form data validation using Go struct tags
- JSON schema validation for events
- XSS prevention through template escaping

**WebSocket Security:**
- Origin validation for WebSocket upgrades
- Session-based authorization
- Rate limiting for event processing

### 8.2 Error Handling

**Error Types:**
- **Template Errors**: Compilation and rendering failures
- **Event Errors**: Handler execution failures
- **Connection Errors**: WebSocket disconnections and timeouts
- **Validation Errors**: User input validation failures

**Error Recovery:**
- Graceful degradation from WebSocket to HTTP
- Error state rendering in templates
- Client-side error boundaries

### 8.3 Logging and Monitoring

**Logging Levels:**
- Debug: Detailed execution flow
- Info: Request/response lifecycle
- Warn: Recoverable errors
- Error: System failures

**Metrics:**
- WebSocket connection count
- Event processing latency
- Template rendering performance
- Memory and CPU usage

### 8.4 Configuration Management

**Configuration Sources:**
- Environment variables
- Configuration files (JSON/YAML)
- Command-line flags
- Programmatic options

**Configurable Aspects:**
- WebSocket settings (timeouts, buffer sizes)
- Template caching behavior
- Pub/sub adapter selection
- Development vs production modes

### 8.5 Testing Strategy

**Testing Levels:**
- **Unit Tests**: Individual component testing
- **Integration Tests**: Component interaction testing
- **E2E Tests**: Full user workflow testing
- **Performance Tests**: Load and stress testing

**Test Tools:**
- Go standard testing framework
- Testify for assertions
- ChromeDP for browser automation
- WebSocket testing utilities

---

## 9. Architecture Decisions

### ADR-001: Event-Driven Architecture
**Status:** Accepted  
**Context:** Need reactive UI updates without full page reloads  
**Decision:** Use event-driven architecture with pub/sub pattern  
**Consequences:** Enables real-time updates, requires careful event design

### ADR-002: WebSocket with HTTP Fallback
**Status:** Accepted  
**Context:** Need real-time communication with reliability  
**Decision:** Primary WebSocket transport with HTTP POST fallback  
**Consequences:** Real-time capabilities with universal compatibility

### ADR-003: Server-Side Template Rendering
**Status:** Accepted  
**Context:** Balance between server control and client performance  
**Decision:** Render templates on server, send HTML fragments to client  
**Consequences:** Consistent state management, reduced client complexity

### ADR-004: Alpine.js for Client-Side Enhancement
**Status:** Accepted  
**Context:** Need lightweight client-side framework integration  
**Decision:** Use Alpine.js for DOM manipulation and event handling  
**Consequences:** Minimal JavaScript footprint, familiar syntax

### ADR-005: Go html/template Integration

**Status:** Accepted  
**Context:** Need type-safe templating with Go integration  
**Decision:** Use Go's built-in html/template package  
**Consequences:** Type safety, XSS protection, familiar to Go developers

### ADR-006: Pluggable Pub/Sub Adapters

**Status:** Accepted  
**Context:** Different deployment scenarios need different messaging  
**Decision:** Abstract pub/sub interface with Redis and in-memory implementations  
**Consequences:** Flexible deployment options, scalable broadcasting

---

## 10. Quality

### 10.1 Quality Tree

```
Quality
├── Performance
│   ├── Response Time (< 100ms for DOM updates)
│   ├── Throughput (1000+ concurrent connections)
│   └── Resource Usage (< 50MB RAM per 1000 connections)
├── Reliability
│   ├── Availability (99.9% uptime)
│   ├── Connection Recovery (automatic reconnection)
│   └── Graceful Degradation (HTTP fallback)
├── Usability
│   ├── Developer Experience (< 1 hour to first app)
│   ├── Learning Curve (familiar Go patterns)
│   └── Documentation (comprehensive examples)
├── Maintainability
│   ├── Code Quality (90%+ test coverage)
│   ├── Modularity (clean interfaces)
│   └── Debugging (comprehensive logging)
└── Security
    ├── Authentication (session-based)
    ├── Authorization (event-level security)
    └── Data Protection (XSS prevention)
```

### 10.2 Quality Scenarios

**Performance Scenario P1:**

- **Source:** User clicks button
- **Stimulus:** DOM update event
- **Environment:** Production with 500 concurrent users
- **Response:** Updated DOM content displayed
- **Measure:** < 100ms from click to visual update

**Reliability Scenario R1:**

- **Source:** Network interruption
- **Stimulus:** WebSocket connection lost
- **Environment:** Production system
- **Response:** Automatic fallback to HTTP POST
- **Measure:** < 5 seconds to resume functionality

**Usability Scenario U1:**

- **Source:** New Go developer
- **Stimulus:** Wants to create reactive form
- **Environment:** Documentation and examples
- **Response:** Working form with real-time validation
- **Measure:** < 30 minutes from start to working code

---

## 11. Risks and Technical Debt

### 11.1 Identified Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|---------|------------|
| **WebSocket Scaling Limits** | Medium | High | Implement connection pooling and load balancing |
| **Template Performance** | Low | Medium | Template caching and compilation optimization |
| **Alpine.js Version Compatibility** | Medium | Medium | Pin versions, maintain compatibility matrix |
| **Memory Leaks in Long Connections** | Low | High | Connection cleanup, memory monitoring |
| **Event Ordering Issues** | Medium | Medium | Event sequence numbers, ordered processing |
| **Goroutine Leak in Connection Management** | Medium | High | Proper context cancellation, connection timeouts |
| **Template Compilation Race Conditions** | Low | High | Enhance template parsing synchronization |
| **Session Cookie Security** | Medium | High | Regular security audits, secure defaults |
| **Event Registry Memory Growth** | Medium | Medium | Implement event handler cleanup, memory limits |
| **Pub/Sub Message Overflow** | Medium | Medium | Message queuing limits, backpressure handling |

### 11.2 Technical Debt - Low-Level Components

#### 11.2.1 Controller Component Debt
**Current Issues:**
- **Route Map Concurrency**: Routes map lacks proper read/write synchronization
- **Configuration Validation**: Limited validation of controller options at startup
- **Event Registry Integration**: Tight coupling between controller and event registry

**Impact:** Potential race conditions, runtime errors with invalid configurations

**Remediation Plan:**
```go
// Add proper synchronization
type controller struct {
    routes    map[string]*route
    routesMu  sync.RWMutex
    // ... other fields
}

// Add configuration validation
func (c *controller) validateOptions() error {
    if c.opt.publicDir != "" && !fileExists(c.opt.publicDir) {
        return fmt.Errorf("public directory does not exist: %s", c.opt.publicDir)
    }
    // Additional validations...
}
```

#### 11.2.2 Route Component Debt
**Current Issues:**
- **Template Parsing Performance**: Templates parsed on every request in development mode
- **Error Template Fallback**: Inconsistent error template resolution
- **Event Template Caching**: Event templates not cached efficiently

**Impact:** Performance degradation, inconsistent error handling

**Remediation Plan:**
```go
// Implement template caching
type templateCache struct {
    templates map[string]*template.Template
    mu        sync.RWMutex
    ttl       time.Duration
}

// Add error template hierarchy
func (rt *route) resolveErrorTemplate() *template.Template {
    if rt.errorTemplate != nil {
        return rt.errorTemplate
    }
    // Fall back to default error template
    return rt.cntrl.defaultErrorTemplate
}
```

#### 11.2.3 Connection Component Debt
**Current Issues:**
- **Message Buffer Overflow**: Fixed-size send channel can block
- **Connection State Management**: No explicit connection state machine
- **Heartbeat Implementation**: Simple heartbeat without adaptive intervals

**Impact:** Connection blocking, unclear connection states, network inefficiency

**Remediation Plan:**
```go
// Implement connection state machine
type ConnectionState int

const (
    StateConnecting ConnectionState = iota
    StateConnected
    StateDisconnecting
    StateDisconnected
)

// Add adaptive heartbeat
type adaptiveHeartbeat struct {
    interval    time.Duration
    maxInterval time.Duration
    missCount   int
}
```

#### 11.2.4 Event Registry Debt
**Current Issues:**
- **Memory Overhead**: Event handlers stored with full reflection information
- **No Handler Lifecycle**: No mechanism to unregister event handlers
- **Type Safety**: Runtime type assertions for event handlers

**Impact:** Memory growth, potential memory leaks, runtime panics

**Remediation Plan:**
```go
// Add handler lifecycle management
type HandlerRegistration struct {
    ID        string
    RouteID   string
    EventID   string
    Handler   interface{}
    CreatedAt time.Time
}

func (er *eventRegistry) Unregister(registrationID string) error {
    // Remove handler and clean up resources
}

// Add compile-time type safety
type TypedHandler[T any] interface {
    Handle(ctx RouteContext, event T) error
}
```

#### 11.2.5 Alpine.js Plugin Debt
**Current Issues:**
- **Error Handling Inconsistency**: Different error handling patterns across modules
- **WebSocket Reconnection Logic**: Complex reconnection state management
- **DOM Manipulation Safety**: Limited validation of DOM operations

**Impact:** Unpredictable error behavior, connection instability, potential XSS vulnerabilities

**Remediation Plan:**
```javascript
// Standardize error handling
class FirError extends Error {
    constructor(type, message, context = {}) {
        super(message);
        this.type = type;
        this.context = context;
        this.timestamp = Date.now();
    }
}

// Simplify reconnection state
class ConnectionStateMachine {
    constructor() {
        this.state = 'disconnected';
        this.transitions = new Map([
            ['disconnected', ['connecting']],
            ['connecting', ['connected', 'disconnected']],
            ['connected', ['disconnecting', 'disconnected']],
            ['disconnecting', ['disconnected']]
        ]);
    }
}
```

#### 11.2.6 Session Management Debt
**Current Issues:**
- **Cookie Security Headers**: Inconsistent security header application
- **Session Cleanup**: No automatic cleanup of expired sessions
- **Session Storage**: In-memory storage doesn't scale across instances

**Impact:** Security vulnerabilities, memory leaks, scalability issues

**Remediation Plan:**
```go
// Add session store interface
type SessionStore interface {
    Get(sessionID string) (*Session, error)
    Set(session *Session) error
    Delete(sessionID string) error
    Cleanup() error // Remove expired sessions
}

// Implement distributed session store
type RedisSessionStore struct {
    client *redis.Client
    ttl    time.Duration
}
```

### 11.3 Performance Technical Debt

#### 11.3.1 Template Rendering
**Issues:**
- Template compilation happens synchronously on request path
- No template precompilation for production deployments
- Template function map recreation on every render

**Metrics Impact:**
- Average response time: +50ms per template compilation
- Memory allocation: +2MB per template parse
- CPU usage: +15% during template operations

#### 11.3.2 WebSocket Message Processing
**Issues:**
- JSON marshaling/unmarshaling on hot path
- No message pooling or reuse
- Synchronous event processing blocks connection

**Metrics Impact:**
- Message latency: +10ms per JSON operation
- GC pressure: +30% due to JSON allocations
- Connection throughput: Limited to ~1000 msg/sec per connection

#### 11.3.3 Memory Management
**Issues:**
- No object pooling for frequent allocations
- Template AST not shared between instances
- Event data copied instead of referenced

**Metrics Impact:**
- Memory usage: +40% overhead from allocations
- GC frequency: +50% more collections
- Response time variance: +25% due to GC pauses

### 11.4 Security Technical Debt

#### 11.4.1 Input Validation
**Current Gaps:**
- WebSocket message size limits not enforced
- Event parameter validation inconsistent
- Template input sanitization incomplete

**Security Impact:**
- DoS via large message attacks
- Injection attacks through event parameters
- XSS through template variables

#### 11.4.2 Session Security
**Current Gaps:**
- No session fixation protection
- Cookie flags not consistently applied
- Session invalidation on security events missing

**Security Impact:**
- Session hijacking vulnerabilities
- Cross-site request forgery
- Privilege escalation attacks

### 11.5 Monitoring and Alerting Enhancements

**Enhanced Metrics to Monitor:**
- **Component-Specific Metrics:**
  - Template compilation time and cache hit ratio
  - Event registry lookup performance
  - Connection pool utilization
  - Message queue depths

- **Low-Level Performance Metrics:**
  - Goroutine count per component
  - Memory allocation rates
  - GC pause times by operation type
  - WebSocket frame processing latency

**Alert Thresholds:**
- Template compilation time > 100ms (Warning)
- Event registry lookup > 1ms (Critical)
- Connection pool exhaustion > 90% (Warning)
- Memory allocation rate > 10MB/s (Warning)
- Goroutine leak detection > 1000 active (Critical)

### 11.7 Architectural Coupling Technical Debt

#### 11.7.1 Controller-Route Tight Coupling
**Current Coupling Issues:**
- Route directly accesses controller's internal fields (`cntrl *controller`)
- Route initialization requires full controller instance
- Event registry tightly bound to controller lifecycle

**Impact on Maintainability:**
```go
// Current tightly coupled design
type route struct {
    cntrl *controller  // Direct dependency on entire controller
    // ... other fields
}

// Route must know about controller internals
func (rt *route) handleEvent() {
    handler := rt.cntrl.eventRegistry.Get(rt.id, eventID)  // Knows internal structure
    rt.cntrl.pubsub.Publish(channel, event)               // Direct access to pubsub
}
```

**Impact Analysis:**
- **Maintainability**: Changes to controller internals break route implementations
- **Extensibility**: Cannot easily swap controller implementations
- **Testability**: Routes cannot be unit tested without full controller setup

**Remediation Strategy:**
```go
// Proposed decoupled design
type RouteServices interface {
    EventRegistry() EventRegistry
    PubSub() pubsub.Adapter
    Renderer() Renderer
    SessionManager() SessionManager
}

type route struct {
    services RouteServices  // Dependency injection through interface
    config   RouteConfig   // Configuration separate from services
}

// Clean separation of concerns
func (rt *route) handleEvent() {
    handler := rt.services.EventRegistry().Get(rt.id, eventID)
    rt.services.PubSub().Publish(channel, event)
}
```

#### 11.7.2 Connection-Controller Circular Dependency
**Current Coupling Issues:**
- Connection holds direct reference to controller
- Controller manages connections through WebSocket handler
- Circular dependency prevents independent testing

**Impact on Architecture:**
```go
// Current circular dependency
type Connection struct {
    controller *controller  // Knows about controller
}

type controller struct {
    // Implicitly manages connections through websocket handler
}

// Connection directly accesses controller internals
func (c *Connection) processEvent() {
    routes := c.controller.routes  // Direct access
    registry := c.controller.eventRegistry  // Tight coupling
}
```

**Impact Analysis:**
- **Maintainability**: Changes ripple between components unpredictably
- **Extensibility**: Cannot replace connection management without controller changes
- **Testability**: Circular dependencies make isolated testing impossible

**Remediation Strategy:**
```go
// Proposed architecture with dependency inversion
type ConnectionManager interface {
    HandleConnection(conn *websocket.Conn, services ConnectionServices) error
    CloseConnection(sessionID string) error
}

type ConnectionServices interface {
    RouteResolver() RouteResolver
    EventProcessor() EventProcessor
    SessionValidator() SessionValidator
    MessageBroadcaster() MessageBroadcaster
}

// Clean separation with interface-based communication
type connection struct {
    services ConnectionServices
    state    ConnectionState
}

func (c *connection) processEvent(event Event) {
    processor := c.services.EventProcessor()
    result := processor.Process(event)
    c.services.MessageBroadcaster().Broadcast(result)
}
```

#### 11.7.3 Renderer-Template Engine Coupling
**Current Coupling Issues:**
- Renderer directly depends on Go's `html/template` package
- Template functions hardcoded in renderer
- No abstraction for different rendering engines

**Impact on Flexibility:**
```go
// Current tightly coupled renderer
type TemplateRenderer struct {
    // Hardcoded dependency on html/template
}

func (tr *TemplateRenderer) RenderRoute(ctx RouteContext, data routeData) error {
    tmpl := ctx.route.getTemplate()  // Direct template access
    tmpl = tmpl.Funcs(newFirFuncMap(ctx, errs))  // Hardcoded function injection
    return tmpl.Execute(ctx.response, data)  // Direct execution
}
```

**Impact Analysis:**
- **Maintainability**: Cannot change template engines without renderer rewrite
- **Extensibility**: No support for alternative rendering strategies
- **Testability**: Difficult to mock template rendering behavior

**Remediation Strategy:**
```go
// Proposed pluggable rendering architecture
type TemplateEngine interface {
    Parse(templatePath string) (Template, error)
    ParseFromString(content string) (Template, error)
}

type Template interface {
    Execute(wr io.Writer, data interface{}) error
    AddFunctions(funcMap FuncMap) Template
}

type FuncMapProvider interface {
    GetFunctions(ctx RenderContext) FuncMap
}

// Decoupled renderer with dependency injection
type renderer struct {
    engine      TemplateEngine
    funcProvider FuncMapProvider
}

func (r *renderer) Render(ctx RenderContext, data interface{}) error {
    template := ctx.Template()
    template = template.AddFunctions(r.funcProvider.GetFunctions(ctx))
    return template.Execute(ctx.Writer(), data)
}
```

#### 11.7.4 Event Registry-Route Coupling
**Current Coupling Issues:**
- Event registry knows about route-specific details
- Routes directly register with global registry
- No abstraction for event routing strategies

**Impact on Scalability:**
```go
// Current tightly coupled event registration
func (c *controller) RouteFunc(opts RouteFunc) http.HandlerFunc {
    // Route registration tightly coupled to controller
    for eventID, handler := range routeOpt.onEvents {
        err := c.eventRegistry.Register(routeOpt.id, eventID, handler)
    }
}

// Event registry knows about route internals
type eventRegistry struct {
    events map[string]map[string]interface{}  // Route-specific organization
}
```

**Impact Analysis:**
- **Maintainability**: Event routing logic scattered across components
- **Extensibility**: Cannot implement alternative event routing strategies
- **Testability**: Event registration requires full route setup

**Remediation Strategy:**
```go
// Proposed event-driven architecture with loose coupling
type EventBus interface {
    Subscribe(topic string, handler EventHandler) Subscription
    Publish(topic string, event Event) error
    Unsubscribe(subscription Subscription) error
}

type EventHandler interface {
    Handle(ctx EventContext, event Event) error
    CanHandle(event Event) bool
}

type RouteEventHandler struct {
    routeID string
    handler OnEventFunc
}

// Decoupled event routing
func (r *RouteEventHandler) Handle(ctx EventContext, event Event) error {
    if ctx.RouteID() != r.routeID {
        return ErrEventNotHandled
    }
    return r.handler(ctx.RouteContext())
}

// Route registration becomes simple subscription
func (rt *route) registerEvents(bus EventBus) error {
    for eventID, handler := range rt.onEvents {
        topic := fmt.Sprintf("route.%s.event.%s", rt.id, eventID)
        subscription := bus.Subscribe(topic, &RouteEventHandler{
            routeID: rt.id,
            handler: handler,
        })
        rt.subscriptions = append(rt.subscriptions, subscription)
    }
    return nil
}
```

#### 11.7.5 Alpine.js Plugin Monolithic Structure
**Current Coupling Issues:**
- Single plugin file with multiple responsibilities
- WebSocket client tightly coupled to event dispatcher
- DOM utilities mixed with business logic

**Impact on Client Architecture:**
```javascript
// Current monolithic plugin structure
export default function firPlugin(Alpine) {
    // WebSocket logic mixed with Alpine.js integration
    let ws = null;
    
    // Event handling tightly coupled to DOM manipulation
    const processEvents = (events) => {
        events.forEach(event => {
            // Direct DOM manipulation mixed with event processing
            const element = document.querySelector(event.target);
            element.innerHTML = event.detail.html;
        });
    };
    
    // Magic functions coupled to internal state
    Alpine.magic('fir', () => ({
        submit: (form) => {
            // Directly accesses global WebSocket instance
            ws.send(JSON.stringify(formData));
        }
    }));
}
```

**Impact Analysis:**
- **Maintainability**: Changes to one feature affect unrelated functionality
- **Extensibility**: Cannot easily add new client-side behaviors
- **Testability**: Monolithic structure makes unit testing difficult

**Remediation Strategy:**
```javascript
// Proposed modular client architecture
class FirCore {
    constructor(config) {
        this.eventBus = new EventBus();
        this.connectionManager = new ConnectionManager(config.websocket);
        this.domManager = new DOMManager();
        this.formManager = new FormManager();
        
        this.setupEventHandlers();
    }
    
    setupEventHandlers() {
        this.eventBus.on('server:event', (event) => {
            this.domManager.processEvent(event);
        });
        
        this.eventBus.on('form:submit', (formData) => {
            this.connectionManager.send(formData);
        });
    }
}

// Separate, testable modules
class ConnectionManager {
    constructor(config) {
        this.config = config;
        this.eventBus = null;
    }
    
    setEventBus(eventBus) {
        this.eventBus = eventBus;
    }
    
    send(data) {
        // Clean separation of concerns
        if (this.isConnected()) {
            this.websocket.send(JSON.stringify(data));
        } else {
            this.fallbackToHTTP(data);
        }
    }
}

// Alpine.js plugin becomes thin integration layer
export default function firPlugin(Alpine) {
    const fir = new FirCore({
        websocket: { /* config */ }
    });
    
    // Clean integration without tight coupling
    Alpine.magic('fir', () => fir.getPublicAPI());
    Alpine.directive('fir-submit', (el, { expression }) => {
        fir.formManager.bindForm(el, expression);
    });
}
```

#### 11.7.6 Session Management Coupling
**Current Coupling Issues:**
- Session encoding/decoding spread across multiple components
- Cookie management tightly coupled to HTTP response handling
- Session validation mixed with business logic

**Impact on Security and Maintainability:**
```go
// Current scattered session management
func (rt *route) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Session management mixed with route logic
    sessionID := generateSessionID()
    encoded, _ := rt.cntrl.secureCookie.Encode(rt.cookieName, sessionData)
    cookie := &http.Cookie{/* cookie setup */}
    http.SetCookie(w, cookie)
}

func (c *Connection) handleMessage(message []byte) error {
    // Session validation mixed with message processing
    eventSessionID, eventRouteID, err := decodeSession(
        *c.controller.secureCookie, 
        c.controller.cookieName, 
        *event.SessionID
    )
}
```

**Impact Analysis:**
- **Maintainability**: Session logic scattered, hard to audit security
- **Extensibility**: Cannot easily change session storage or validation
- **Testability**: Session testing requires full component setup

**Remediation Strategy:**
```go
// Proposed centralized session management
type SessionManager interface {
    CreateSession(userID string, routeID string) (*Session, error)
    ValidateSession(sessionToken string) (*Session, error)
    RefreshSession(sessionID string) (*Session, error)
    InvalidateSession(sessionID string) error
}

type CookieManager interface {
    SetSessionCookie(w http.ResponseWriter, session *Session) error
    GetSessionToken(r *http.Request) (string, error)
    ClearSessionCookie(w http.ResponseWriter) error
}

// Clean separation of concerns
type sessionManager struct {
    store      SessionStore
    cookieMgr  CookieManager
    config     SessionConfig
}

func (sm *sessionManager) AuthenticateRequest(r *http.Request) (*Session, error) {
    token, err := sm.cookieMgr.GetSessionToken(r)
    if err != nil {
        return nil, err
    }
    return sm.ValidateSession(token)
}

// Routes use session manager through dependency injection
type route struct {
    sessionMgr SessionManager
    // ... other dependencies
}

func (rt *route) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    session, err := rt.sessionMgr.AuthenticateRequest(r)
    if err != nil {
        // Handle authentication error
        return
    }
    // Continue with authenticated request
}
```

### 11.8 Coupling Impact Assessment Matrix

| Component Pair | Coupling Level | Maintainability Impact | Extensibility Impact | Testability Impact | Priority |
|---------------|----------------|------------------------|---------------------|-------------------|----------|
| Controller-Route | High | High | High | High | P0 |
| Connection-Controller | High | High | Medium | High | P0 |
| Renderer-Template | Medium | Medium | High | Medium | P1 |
| EventRegistry-Route | High | Medium | High | High | P1 |
| Alpine.js Modules | High | High | Medium | High | P1 |
| Session-Components | Medium | High | Medium | Medium | P2 |

### 11.9 Decoupling Roadmap

**Phase 1 (Q1 2025): Critical Coupling Reduction**
- Implement Controller-Route interface abstraction
- Break Connection-Controller circular dependency
- Create event bus architecture for loose coupling

**Phase 2 (Q2 2025): Service Layer Introduction**
- Implement session management service
- Create pluggable renderer architecture
- Modularize Alpine.js plugin structure

**Phase 3 (Q3 2025): Interface Standardization**
- Define standard service interfaces
- Implement dependency injection container
- Create configuration-driven component assembly

**Phase 4 (Q4 2025): Testing and Validation**
- Comprehensive unit test coverage with mocked dependencies
- Integration test suite with component isolation
- Performance validation of decoupled architecture

**Success Metrics:**
- Reduce coupling metrics by 60% (measured by code analysis tools)
- Achieve 90%+ unit test coverage with isolated component testing
- Enable component replacement without breaking changes
- Reduce time-to-implement new features by 40%

---

## 11.8. Fir Actions System: HTML Attributes and Framework Comparison

### 11.8.1. Overview of x-fir-* Actions

Fir introduces a unique approach to DOM manipulation through declarative HTML attributes called "x-fir-* actions." These attributes are processed server-side during template compilation and converted to @fir: Alpine.js event listeners, providing a clean, HTML-first approach to reactive user interfaces.

### 11.8.2. Core x-fir-* Action Syntax and Behavior

#### Basic Syntax Pattern
```html
<!-- Standard pattern -->
<element x-fir-action="event:state" />

<!-- With template targeting -->
<element x-fir-action:template="event:state" />

<!-- With parameters -->
<element x-fir-action-param="event:state" />
```

#### Compilation Process
x-fir-* attributes are processed server-side during template compilation and converted to @fir: Alpine.js event listeners:

```html
<!-- Source -->
<div x-fir-refresh="update:ok">Content</div>

<!-- Compiled -->
<div @fir:update:ok="$fir.replace()">Content</div>
```

### 11.8.3. Available Action Types

| Action | Purpose | JavaScript Function | Example |
| --- | --- | --- | --- |
| `x-fir-refresh` | Replace element content | `$fir.replace()` | `<div x-fir-refresh="save:ok">` |
| `x-fir-append` | Add content to end | `$fir.appendEl()` | `<div x-fir-append:item="create:ok">` |
| `x-fir-prepend` | Add content to beginning | `$fir.prependEl()` | `<div x-fir-prepend:item="create:ok">` |
| `x-fir-remove` | Remove element | `$fir.removeEl()` | `<div x-fir-remove="delete:ok">` |
| `x-fir-remove-parent` | Remove parent element | `$fir.removeParentEl()` | `<button x-fir-remove-parent="delete:ok">` |
| `x-fir-reset` | Reset form element | `$el.reset()` | `<form x-fir-reset="submit:ok">` |
| `x-fir-toggle-disabled` | Toggle disabled state | `$fir.toggleDisabled()` | `<button x-fir-toggle-disabled="submit:pending">` |
| `x-fir-toggle-class` | Toggle CSS classes | `$fir.toggleClass()` | `<div x-fir-toggle-class:loading="submit:pending">` |
| `x-fir-redirect` | Navigate to URL | `$fir.redirect()` | `<form x-fir-redirect:home="success:ok">` |

### 11.8.4. Event State System

Fir uses a sophisticated event state system with four lifecycle states:

- **ok**: Successful completion
- **error**: Error occurred
- **pending**: Operation in progress
- **done**: Operation completed (regardless of success/failure)

```html
<!-- Respond to different states -->
<div x-fir-refresh="save:ok">Success content</div>
<div x-fir-refresh="save:error">Error content</div>
<button x-fir-toggle-disabled="save:pending">Submit</button>
```

### 11.8.5. Template Targeting and Extraction

Fir automatically extracts templates from elements containing Go template syntax:

```html
<!-- Auto-extracted template -->
<p x-fir-refresh="create:error">
    {{ fir.Error "create.text" }}
</p>

<!-- Compiles to -->
<p x-fir-refresh="create:error" @fir:create:error::fir-auto-123="$fir.replace()">
    {{ fir.Error "create.text" }}
</p>
```

Explicit template targeting:
```html
<div x-fir-append:todo="create:ok">
    <!-- Targets named template "todo" -->
</div>
```

### 11.8.6. Framework Comparison: Developer Experience Analysis

#### 11.8.6.1. HTMX Comparison

**HTMX Approach:**
```html
<!-- HTMX: HTTP-centric, attribute-based -->
<button hx-post="/api/todos" 
        hx-target="#todo-list" 
        hx-swap="afterbegin"
        hx-indicator="#spinner">
    Add Todo
</button>

<div id="todo-list"></div>
<div id="spinner" class="htmx-indicator">Loading...</div>
```

**Fir Equivalent:**
```html
<!-- Fir: Event-centric, server-rendered -->
<form method="post" action="/?event=create-todo" @submit.prevent="$fir.submit()">
    <button x-fir-toggle-disabled="create-todo:pending" type="submit">
        Add Todo
    </button>
</form>

<div x-fir-prepend:todo="create-todo:ok">
    {{ range .todos }}
        {{ template "todo" . }}
    {{ end }}
</div>
```

**Key Differences:**
- **HTMX**: Direct HTTP requests, CSS selectors for targeting
- **Fir**: Event-driven, template-based rendering, WebSocket communication
- **State Management**: HTMX uses indicators; Fir uses event states
- **Targeting**: HTMX uses CSS selectors; Fir uses template names

#### 11.8.6.2. Phoenix LiveView Comparison

**Phoenix LiveView Approach:**
```elixir
# LiveView: Component-based, declarative assigns
def handle_event("create_todo", %{"text" => text}, socket) do
  case create_todo(text) do
    {:ok, todo} ->
      {:noreply, assign(socket, :todos, [todo | socket.assigns.todos])}
    {:error, changeset} ->
      {:noreply, assign(socket, :changeset, changeset)}
  end
end
```

```heex
<!-- LiveView template -->
<form phx-submit="create_todo">
  <input name="text" />
  <button phx-disable-with="Creating...">Add</button>
</form>

<div id="todos" phx-update="prepend">
  <%= for todo <- @todos do %>
    <div id={"todo-#{todo.id}"}>
      <%= todo.text %>
      <button phx-click="delete_todo" phx-value-id={todo.id}>Delete</button>
    </div>
  <% end %>
</div>
```

**Fir Equivalent:**
```go
// Fir: Route-based handlers with OnEvent functions
func createTodo(db *Database) fir.OnEventFunc {
    return func(ctx fir.RouteContext) error {
        todo := new(Todo)
        if err := ctx.Bind(todo); err != nil {
            return err
        }
        if len(todo.Text) < 3 {
            return ctx.FieldError("text", errors.New("todo is too short"))
        }
        
        // Save to database
        if err := db.Insert(todo); err != nil {
            return err
        }
        
        return ctx.Data(todo)
    }
}

// Route configuration
func TodoRoute(db *Database) fir.RouteFunc {
    return func() fir.RouteOptions {
        return fir.RouteOptions{
            fir.ID("todos"),
            fir.Content("todos.html"),
            fir.OnEvent("create-todo", createTodo(db)),
        }
    }
}
```

```html
<!-- Fir template -->
<form method="post" action="/?event=create-todo" @submit.prevent="$fir.submit()">
    <input name="text" />
    <button x-fir-toggle-disabled="create-todo:pending">Add</button>
</form>

<div x-fir-prepend:todo="create-todo:ok">
    {{ range .todos }}
        {{ block "todo" . }}
            <div fir-key="{{ .ID }}">
                {{ .Text }}
                <button x-fir-remove="delete-todo:ok">Delete</button>
            </div>
        {{ end }}
    {{ end }}
</div>
```

**Key Differences:**
- **LiveView**: Stateful processes, automatic state synchronization
- **Fir**: Stateless routes, explicit event handling
- **Updates**: LiveView uses `phx-update` strategies; Fir uses action attributes
- **Targeting**: LiveView uses DOM IDs; Fir uses template blocks

#### 11.8.6.3. Laravel Livewire Comparison

**Laravel Livewire Approach:**
```php
// Livewire: Component classes with reactive properties
class TodoComponent extends Component
{
    public $todos = [];
    public $text = '';

    public function addTodo()
    {
        $this->todos[] = ['text' => $this->text, 'id' => uniqid()];
        $this->text = '';
    }

    public function deleteTodo($id)
    {
        $this->todos = array_filter($this->todos, fn($todo) => $todo['id'] !== $id);
    }
}
```

```html
<!-- Livewire template -->
<form wire:submit.prevent="addTodo">
    <input wire:model="text" />
    <button type="submit" wire:loading.attr="disabled">Add</button>
</form>

@foreach($todos as $todo)
    <div wire:key="todo-{{ $todo['id'] }}">
        {{ $todo['text'] }}
        <button wire:click="deleteTodo('{{ $todo['id'] }}')">Delete</button>
    </div>
@endforeach
```

**Fir Equivalent:**
```go
// Fir: OnEvent functions with explicit data handling
func addTodo(todos *[]Todo) fir.OnEventFunc {
    return func(ctx fir.RouteContext) error {
        todo := new(Todo)
        if err := ctx.Bind(todo); err != nil {
            return err
        }
        
        todo.ID = generateID()
        *todos = append(*todos, *todo)
        
        return ctx.Data(todo)
    }
}

func deleteTodo(todos *[]Todo) fir.OnEventFunc {
    type deleteReq struct {
        ID string `json:"id"`
    }
    return func(ctx fir.RouteContext) error {
        req := new(deleteReq)
        if err := ctx.Bind(req); err != nil {
            return err
        }
        
        for i, todo := range *todos {
            if todo.ID == req.ID {
                *todos = append((*todos)[:i], (*todos)[i+1:]...)
                break
            }
        }
        return nil
    }
}

// Route setup
func TodoRoute() fir.RouteFunc {
    todos := []Todo{}
    return func() fir.RouteOptions {
        return fir.RouteOptions{
            fir.ID("todos"),
            fir.Content("todos.html"),
            fir.OnEvent("add-todo", addTodo(&todos)),
            fir.OnEvent("delete-todo", deleteTodo(&todos)),
        }
    }
}
```

```html
<!-- Fir template -->
<form method="post" action="/?event=add-todo" @submit.prevent="$fir.submit()" x-fir-reset="add-todo:ok">
    <input name="text" />
    <button x-fir-toggle-disabled="add-todo:pending">Add</button>
</form>

<div x-fir-append:todo="add-todo:ok">
    {{ range .todos }}
        {{ block "todo" . }}
            <div fir-key="{{ .ID }}">
                {{ .Text }}
                <button x-fir-remove="delete-todo:ok">Delete</button>
            </div>
        {{ end }}
    {{ end }}
</div>
```

**Key Differences:**
- **Livewire**: Component state management, automatic property binding
- **Fir**: Manual state management, explicit event emission
- **Updates**: Livewire uses `wire:key` for tracking; Fir uses `fir-key`
- **Loading States**: Livewire uses `wire:loading`; Fir uses event states

#### 11.8.6.4. Hotwire Turbo Comparison

**Hotwire Turbo Approach:**
```html
<!-- Turbo: Frame-based updates -->
<turbo-frame id="todo-form">
    <form action="/todos" method="post">
        <input name="text" />
        <button>Add Todo</button>
    </form>
</turbo-frame>

<turbo-frame id="todo-list">
    <div id="todos">
        <!-- Server renders this section -->
    </div>
</turbo-frame>
```

**Fir Equivalent:**
```html
<!-- Fir: Attribute-based targeting -->
<form method="post" action="/?event=create-todo" @submit.prevent="$fir.submit()">
    <input name="text" />
    <button x-fir-toggle-disabled="create-todo:pending">Add Todo</button>
</form>

<div x-fir-prepend:todo="create-todo:ok">
    {{ range .todos }}
        {{ template "todo" . }}
    {{ end }}
</div>
```

**Key Differences:**
- **Turbo**: Frame-based architecture, HTTP-centric
- **Fir**: Attribute-based targeting, WebSocket-centric
- **Granularity**: Turbo updates entire frames; Fir updates specific elements
- **Real-time**: Turbo requires Streams for real-time; Fir has built-in WebSocket support

### 11.8.7. Developer Experience Analysis

#### 11.8.7.1. Benefits of Fir's Approach

**Strengths:**
1. **HTML-First Design**: Declarative attributes keep behavior close to markup
2. **Event State System**: Built-in loading, error, and success states
3. **Template Integration**: Seamless Go template extraction and targeting
4. **WebSocket Native**: Real-time updates without additional setup
5. **Compile-time Processing**: x-fir-* attributes are validated and optimized at build time
6. **Alpine.js Integration**: Leverages existing Alpine.js ecosystem and knowledge
7. **Multi-Target Updates**: Single user event can update multiple disparate parts of the application
8. **Real-time Broadcasting**: Events are automatically sent via pub/sub to all connected clients (tabs, mobile, web browsers)

**Developer Productivity:**
```html
<!-- Complex interaction with minimal code -->
<form method="post" action="/?event=save" @submit.prevent="$fir.submit()" x-fir-reset="save:ok">
    <input name="title" x-fir-refresh="save:error" />
    <p x-fir-refresh="save:error">{{ fir.Error "save.title" }}</p>
    
    <button x-fir-toggle-disabled="save:pending" 
            x-fir-toggle-class:loading="save:pending">
        <span x-fir-refresh="save:pending">Saving...</span>
        <span x-fir-refresh="save:ok,save:error">Save</span>
    </button>
</form>
```

**Multi-Target Updates:**

One of Fir's most powerful features is the ability to update multiple disparate parts of the application with a single user event. Unlike other frameworks that typically update one component or section, Fir can coordinate complex UI updates across the entire application:

```html
<!-- Single "create-post" event updates multiple sections -->
<nav>
    <!-- Update notification count -->
    <span x-fir-refresh="create-post:ok">{{ .notificationCount }}</span>
</nav>

<aside>
    <!-- Update sidebar stats -->
    <div x-fir-refresh="create-post:ok">
        Posts: {{ .totalPosts }}
    </div>
</aside>

<main>
    <!-- Add new post to main feed -->
    <div x-fir-prepend:post="create-post:ok">
        {{ range .posts }}
            {{ template "post" . }}
        {{ end }}
    </div>
</main>

<footer>
    <!-- Update footer analytics -->
    <span x-fir-refresh="create-post:ok">{{ .dailyPostCount }} posts today</span>
</footer>
```

This single `create-post` event can simultaneously:
- Add the new post to the main feed
- Update the notification count in the navigation
- Refresh sidebar statistics
- Update footer analytics
- All while showing appropriate loading states

**Real-time Broadcasting:**

Another unique advantage is automatic real-time synchronization across all connected clients. When one user performs an action, all other connected clients (different browser tabs, mobile apps, other users) automatically receive the same event updates:

```go
// Server-side: One event handler updates all clients
func createPost(db *Database) fir.OnEventFunc {
    return func(ctx fir.RouteContext) error {
        post := new(Post)
        if err := ctx.Bind(post); err != nil {
            return err
        }
        
        if err := db.Save(post); err != nil {
            return err
        }
        
        // This data is automatically broadcast to ALL connected clients
        return ctx.Data(map[string]any{
            "post": post,
            "notificationCount": getNotificationCount(ctx.UserID()),
            "totalPosts": getTotalPosts(),
            "dailyPostCount": getDailyPostCount(),
        })
    }
}
```

When User A creates a post, Users B and C in different browser tabs/devices automatically see:
- The new post appear in their feeds
- Updated notification counts
- Refreshed statistics
- All without any additional code or WebSocket management

**Comparison with Other Frameworks:**

- **HTMX**: Limited to single HTTP requests, no built-in real-time capabilities
- **LiveView**: Updates single LiveView process, but has built-in Phoenix PubSub for broadcasting to other processes
- **Livewire**: Updates single component, requires manual event broadcasting for real-time features
- **Turbo**: Updates single frame, requires Turbo Streams with additional setup for real-time

#### 11.8.7.2. Limitations and Trade-offs

**Drawbacks:**
1. **Learning Curve**: Unique syntax requires Fir-specific knowledge
2. **Go Template Dependency**: Tied to Go's template system
3. **Compilation Step**: Requires build-time processing of templates
4. **Limited Ecosystem**: Smaller community compared to established frameworks
5. **WebSocket Requirement**: Always-on connection may not suit all use cases

**Complexity Comparison:**
- **HTMX**: Simpler for basic HTTP interactions, limited real-time capabilities
- **LiveView**: More complex setup, but powerful state management
- **Livewire**: Similar complexity, but PHP-centric
- **Turbo**: Simpler conceptually, but limited granularity

### 11.8.8. Performance and Architecture Implications

#### 11.8.8.1. Runtime Performance
- **Compilation**: x-fir-* attributes processed at build time, not runtime
- **Bundle Size**: Minimal JavaScript footprint (piggybacks on Alpine.js)
- **Network**: WebSocket reduces HTTP overhead for dynamic updates
- **DOM Updates**: Precise targeting reduces unnecessary re-renders

#### 11.8.8.2. Development Experience
- **Debugging**: Event states provide clear visibility into operation lifecycle
- **Testing**: Declarative attributes are easier to test than imperative JavaScript
- **Maintenance**: Co-location of behavior and markup improves maintainability

### 11.8.9. Recommendations for Adoption

**Choose Fir When:**
- Building Go web applications with real-time requirements
- Team prefers HTML-first, declarative approaches
- Need fine-grained control over DOM updates
- WebSocket communication fits the application architecture

**Consider Alternatives When:**
- Working in non-Go environments
- Simple HTTP request/response patterns are sufficient
- Team lacks Alpine.js/Go template experience
- Offline-first functionality is required

**Migration Strategy:**
1. Start with simple x-fir-refresh patterns
2. Gradually adopt event states for better UX
3. Introduce real-time features using WebSocket events
4. Optimize with template targeting and custom actions

---

## 12. Glossary

| Term | Definition |
|------|------------|
| **Alpine.js** | Lightweight JavaScript framework for adding reactivity to HTML |
| **Connection** | WebSocket connection instance managing client-server communication |
| **Controller** | Main Fir component coordinating routes, events, and connections |
| **DOM Event** | Client-side event representing changes to be applied to the DOM |
| **Event Handler** | Server-side function processing user or system events |
| **Event Registry** | Component managing registration and lookup of event handlers |
| **Event State** | Lifecycle state of an event (ok, error, pending, done) |
| **Fir Actions** | HTML attributes (x-fir-*) that are processed server-side and converted to Alpine.js event listeners |
| **Fir Event** | Custom event format used for client-server communication |
| **HTMX** | Hypermedia framework for building dynamic web applications |
| **LiveView** | Phoenix framework component for real-time server-rendered UIs |
| **OnEvent** | User-defined function handling specific application events |
| **OnLoad** | Function executed when a route is initially loaded |
| **Partial Rendering** | Rendering only specific parts of a template rather than the entire page |
| **Pub/Sub** | Publish-subscribe pattern for broadcasting events |
| **Route** | Fir component handling HTTP requests and WebSocket events for a specific path |
| **Route Context** | Request-scoped data structure containing event, request, response, and route |
| **Session ID** | Unique identifier for user sessions, used for WebSocket authorization |
| **Template** | Go html/template used for server-side HTML generation |
| **Template Block** | Named template section that can be rendered independently |
| **Template Targeting** | Mechanism for directing server-rendered content to specific template blocks |
| **WebSocket** | Full-duplex communication protocol for real-time updates |
| **WebSocket Upgrade** | HTTP protocol upgrade to establish WebSocket connection |
| **x-fir-* Actions** | Declarative HTML attributes that are processed server-side and converted to Alpine.js event handlers |

---

*This architecture documentation follows the arc42 template and provides a comprehensive overview of the Fir framework's architecture, design decisions, and operational considerations.*
