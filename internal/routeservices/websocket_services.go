package routeservices

import (
	"net/http"
	"time"

	"github.com/gorilla/schema"
	"github.com/gorilla/websocket"
	"github.com/livefir/fir/internal/event"
	"github.com/livefir/fir/pubsub"
)

// WebSocketServices defines the interface for WebSocket-related operations
// This interface abstracts WebSocket functionality away from the controller,
// enabling clean separation of concerns and improved testability.
type WebSocketServices interface {
	// WebSocket upgrade configuration
	GetWebSocketUpgrader() *websocket.Upgrader

	// Route and event management
	GetRoutes() map[string]RouteInterface
	GetEventRegistry() event.EventRegistry

	// Session and security
	DecodeSession(sessionID string) (userOrSessionID, routeID string, err error)
	GetCookieName() string

	// Configuration
	GetDropDuplicateInterval() time.Duration
	IsWebSocketDisabled() bool

	// Lifecycle callbacks
	OnSocketConnect(userOrSessionID string) error
	OnSocketDisconnect(userOrSessionID string)
}

// RouteInterface defines the minimal interface needed for route operations in WebSocket context
// This avoids circular imports by not directly referencing the route struct
type RouteInterface interface {
	ID() string
	Options() interface{} // Returns RouteOptions but using interface{} to avoid circular import
	ChannelFunc() func(r *http.Request, viewID string) *string
	PubSub() pubsub.Adapter
	DevelopmentMode() bool
	EventSender() interface{} // Returns chan<- Event but using interface{} to avoid circular import
	Services() *RouteServices
	FormDecoder() *schema.Decoder
	GetRenderer() interface{}       // Returns Renderer but using interface{} to avoid circular import
	GetEventTemplates() interface{} // Returns eventTemplates but using interface{} to avoid circular import
	GetTemplate() interface{}       // Returns *template.Template but using interface{} to avoid circular import
	GetAppName() string             // Returns the app name for template context
	GetCookieName() string          // Returns the session cookie name
	GetSecureCookie() interface{}   // Returns *securecookie.SecureCookie but using interface{} to avoid circular import
}

// MockWebSocketServices provides a test implementation of WebSocketServices
type MockWebSocketServices struct {
	WebSocketUpgrader     *websocket.Upgrader
	Routes                map[string]RouteInterface
	EventRegistry         event.EventRegistry
	CookieName            string
	DropDuplicateInterval time.Duration
	WebSocketDisabled     bool
	OnConnectFunc         func(userOrSessionID string) error
	OnDisconnectFunc      func(userOrSessionID string)
	DecodeSessionFunc     func(sessionID string) (string, string, error)
}

// GetWebSocketUpgrader returns the WebSocket upgrader configuration
func (m *MockWebSocketServices) GetWebSocketUpgrader() *websocket.Upgrader {
	if m.WebSocketUpgrader == nil {
		return &websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}
	}
	return m.WebSocketUpgrader
}

// GetRoutes returns the routes map
func (m *MockWebSocketServices) GetRoutes() map[string]RouteInterface {
	if m.Routes == nil {
		return make(map[string]RouteInterface)
	}
	return m.Routes
}

// GetEventRegistry returns the event registry
func (m *MockWebSocketServices) GetEventRegistry() event.EventRegistry {
	if m.EventRegistry == nil {
		return event.NewEventRegistry()
	}
	return m.EventRegistry
}

// DecodeSession decodes a session ID and returns user/session ID and route ID
func (m *MockWebSocketServices) DecodeSession(sessionID string) (string, string, error) {
	if m.DecodeSessionFunc != nil {
		return m.DecodeSessionFunc(sessionID)
	}
	// Default mock implementation
	return sessionID, "mock-route-id", nil
}

// GetCookieName returns the session cookie name
func (m *MockWebSocketServices) GetCookieName() string {
	if m.CookieName == "" {
		return "_session"
	}
	return m.CookieName
}

// GetDropDuplicateInterval returns the event deduplication interval
func (m *MockWebSocketServices) GetDropDuplicateInterval() time.Duration {
	if m.DropDuplicateInterval == 0 {
		return 100 * time.Millisecond
	}
	return m.DropDuplicateInterval
}

// IsWebSocketDisabled returns whether WebSocket is disabled
func (m *MockWebSocketServices) IsWebSocketDisabled() bool {
	return m.WebSocketDisabled
}

// OnSocketConnect handles socket connection events
func (m *MockWebSocketServices) OnSocketConnect(userOrSessionID string) error {
	if m.OnConnectFunc != nil {
		return m.OnConnectFunc(userOrSessionID)
	}
	return nil
}

// OnSocketDisconnect handles socket disconnection events
func (m *MockWebSocketServices) OnSocketDisconnect(userOrSessionID string) {
	if m.OnDisconnectFunc != nil {
		m.OnDisconnectFunc(userOrSessionID)
	}
}

// MockRoute provides a test implementation of RouteInterface
type MockRoute struct {
	RouteID            string
	RouteOptions       interface{}
	ChannelFunction    func(r *http.Request, viewID string) *string
	PubSubAdapter      pubsub.Adapter
	DevMode            bool
	EventSenderChannel interface{} // chan<- Event but using interface{} to avoid circular import
	RouteServices      *RouteServices
	Decoder            *schema.Decoder
	Renderer           interface{} // Renderer but using interface{} to avoid circular import
	EventTemplates     interface{} // eventTemplates but using interface{} to avoid circular import
	Template           interface{} // *template.Template but using interface{} to avoid circular import
	AppName            string      // App name for template context
	CookieName         string      // Session cookie name
	SecureCookie       interface{} // *securecookie.SecureCookie but using interface{} to avoid circular import
}

// ID returns the route ID
func (m *MockRoute) ID() string {
	return m.RouteID
}

// Options returns the route options
func (m *MockRoute) Options() interface{} {
	return m.RouteOptions
}

// ChannelFunc returns the channel function
func (m *MockRoute) ChannelFunc() func(r *http.Request, viewID string) *string {
	if m.ChannelFunction == nil {
		return func(r *http.Request, viewID string) *string {
			channel := "default-channel"
			return &channel
		}
	}
	return m.ChannelFunction
}

// PubSub returns the pubsub adapter
func (m *MockRoute) PubSub() pubsub.Adapter {
	return m.PubSubAdapter
}

// DevelopmentMode returns whether development mode is enabled
func (m *MockRoute) DevelopmentMode() bool {
	return m.DevMode
}

// EventSender returns the event sender channel
func (m *MockRoute) EventSender() interface{} {
	return m.EventSenderChannel
}

// Services returns the route services
func (m *MockRoute) Services() *RouteServices {
	return m.RouteServices
}

// FormDecoder returns the form decoder
func (m *MockRoute) FormDecoder() *schema.Decoder {
	if m.Decoder == nil {
		decoder := schema.NewDecoder()
		decoder.IgnoreUnknownKeys(true)
		return decoder
	}
	return m.Decoder
}

// GetRenderer returns the renderer
func (m *MockRoute) GetRenderer() interface{} {
	return m.Renderer
}

// GetEventTemplates returns the event templates
func (m *MockRoute) GetEventTemplates() interface{} {
	return m.EventTemplates
}

// GetTemplate returns the route template
func (m *MockRoute) GetTemplate() interface{} {
	return m.Template
}

// GetAppName returns the app name
func (m *MockRoute) GetAppName() string {
	return m.AppName
}

// GetCookieName returns the session cookie name
func (m *MockRoute) GetCookieName() string {
	if m.CookieName == "" {
		return "_session"
	}
	return m.CookieName
}

// GetSecureCookie returns the secure cookie instance
func (m *MockRoute) GetSecureCookie() interface{} {
	return m.SecureCookie
}
