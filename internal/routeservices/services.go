package routeservices

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/gorilla/schema"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/websocket"
	"github.com/livefir/fir/internal/event"
	"github.com/livefir/fir/internal/services"
	"github.com/livefir/fir/pubsub"
	"github.com/patrickmn/go-cache"
)

// RouteServices encapsulates all the services and dependencies that routes need
// from the controller. This provides a clean separation of concerns and makes
// routes testable in isolation.
type RouteServices struct {
	// Event management
	EventRegistry event.EventRegistry

	// New event service layer for improved testability and maintainability
	EventService services.EventService

	// Pub/Sub system
	PubSub pubsub.Adapter

	// Rendering - Using interface{} to avoid circular imports
	// This will be cast to the actual Renderer interface when used
	Renderer interface{}

	// Template engine - Using interface{} to avoid circular imports
	// This will be cast to the actual TemplateEngine interface when used
	TemplateEngine interface{}

	// Request routing and parameters
	ChannelFunc    func(r *http.Request, routeID string) *string
	PathParamsFunc func(r *http.Request) map[string]string

	// Configuration and utilities
	Options *Options

	// WebSocket services - provides WebSocket functionality without controller dependency
	WebSocketServices WebSocketServices
}

// Options contains all the configuration options that routes might need
type Options struct {
	// WebSocket configuration
	OnSocketConnect    func(userOrSessionID string) error
	OnSocketDisconnect func(userOrSessionID string)
	WebsocketUpgrader  websocket.Upgrader

	// Template and caching configuration
	DisableTemplateCache bool
	DisableWebsocket     bool
	EnableWatch          bool
	WatchExts            []string
	PublicDir            string
	DevelopmentMode      bool

	// File system functions
	ReadFile  func(string) (string, []byte, error)
	ExistFile func(string) bool

	// Application configuration
	AppName      string
	FormDecoder  *schema.Decoder
	CookieName   string
	SecureCookie *securecookie.SecureCookie
	Cache        *cache.Cache
	FuncMap      template.FuncMap

	// Performance tuning
	DropDuplicateInterval time.Duration

	// Debug and logging
	DebugLog bool
}

// NewRouteServices creates a new RouteServices instance with the provided dependencies
func NewRouteServices(eventRegistry event.EventRegistry, pubsub pubsub.Adapter, renderer interface{}, options *Options) *RouteServices {
	return &RouteServices{
		EventRegistry: eventRegistry,
		PubSub:        pubsub,
		Renderer:      renderer,
		Options:       options,
	}
}

// NewRouteServicesWithTemplateEngine creates a new RouteServices instance with template engine
func NewRouteServicesWithTemplateEngine(eventRegistry event.EventRegistry, pubsub pubsub.Adapter, renderer interface{}, templateEngine interface{}, options *Options) *RouteServices {
	return &RouteServices{
		EventRegistry:  eventRegistry,
		PubSub:         pubsub,
		Renderer:       renderer,
		TemplateEngine: templateEngine,
		Options:        options,
	}
}

// NewRouteServicesWithWebSocket creates a new RouteServices instance including WebSocket services
func NewRouteServicesWithWebSocket(eventRegistry event.EventRegistry, pubsub pubsub.Adapter, renderer interface{}, options *Options, wsServices WebSocketServices) *RouteServices {
	return &RouteServices{
		EventRegistry:     eventRegistry,
		PubSub:            pubsub,
		Renderer:          renderer,
		Options:           options,
		WebSocketServices: wsServices,
	}
}

// NewRouteServicesWithWebSocketAndTemplateEngine creates a new RouteServices instance with WebSocket services and template engine
func NewRouteServicesWithWebSocketAndTemplateEngine(eventRegistry event.EventRegistry, pubsub pubsub.Adapter, renderer interface{}, templateEngine interface{}, options *Options, wsServices WebSocketServices) *RouteServices {
	return &RouteServices{
		EventRegistry:     eventRegistry,
		PubSub:            pubsub,
		Renderer:          renderer,
		TemplateEngine:    templateEngine,
		Options:           options,
		WebSocketServices: wsServices,
	}
}

// SetChannelFunc sets the function used to determine the channel for a route
func (rs *RouteServices) SetChannelFunc(f func(r *http.Request, routeID string) *string) {
	rs.ChannelFunc = f
}

// SetPathParamsFunc sets the function used to extract path parameters from requests
func (rs *RouteServices) SetPathParamsFunc(f func(r *http.Request) map[string]string) {
	rs.PathParamsFunc = f
}

// UpdateOptions allows runtime updates to the options configuration
func (rs *RouteServices) UpdateOptions(options *Options) {
	rs.Options = options
}

// ValidateServices checks that all required services are properly configured
func (rs *RouteServices) ValidateServices() error {
	if rs.EventRegistry == nil {
		return fmt.Errorf("eventRegistry is required but not set")
	}
	if rs.PubSub == nil {
		return fmt.Errorf("pubSub is required but not set")
	}
	if rs.Renderer == nil {
		return fmt.Errorf("renderer is required but not set")
	}
	if rs.Options == nil {
		return fmt.Errorf("options is required but not set")
	}
	// Note: WebSocketServices is optional and may be nil for non-WebSocket routes
	return nil
}

// ValidateWebSocketServices checks that WebSocket services are properly configured
func (rs *RouteServices) ValidateWebSocketServices() error {
	if rs.WebSocketServices == nil {
		return fmt.Errorf("webSocketServices is required but not set")
	}
	if rs.WebSocketServices.GetWebSocketUpgrader() == nil {
		return fmt.Errorf("webSocket upgrader is required but not set")
	}
	if rs.WebSocketServices.GetEventRegistry() == nil {
		return fmt.Errorf("webSocket event registry is required but not set")
	}
	return nil
}

// Clone creates a copy of the RouteServices for independent testing or modification
func (rs *RouteServices) Clone() *RouteServices {
	return &RouteServices{
		EventRegistry:     rs.EventRegistry,
		PubSub:            rs.PubSub,
		Renderer:          rs.Renderer,
		TemplateEngine:    rs.TemplateEngine, // Include template engine in clone
		ChannelFunc:       rs.ChannelFunc,
		PathParamsFunc:    rs.PathParamsFunc,
		Options:           rs.Options,
		WebSocketServices: rs.WebSocketServices,
	}
}

// SetWebSocketServices sets the WebSocket services for the RouteServices
func (rs *RouteServices) SetWebSocketServices(wsServices WebSocketServices) {
	rs.WebSocketServices = wsServices
}

// GetWebSocketServices returns the WebSocket services
func (rs *RouteServices) GetWebSocketServices() WebSocketServices {
	return rs.WebSocketServices
}

// HasWebSocketServices returns true if WebSocket services are configured
func (rs *RouteServices) HasWebSocketServices() bool {
	return rs.WebSocketServices != nil
}
