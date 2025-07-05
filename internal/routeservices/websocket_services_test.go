package routeservices

import (
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/livefir/fir/internal/event"
	"github.com/livefir/fir/pubsub"
)

// TestWebSocketServices tests the WebSocketServices interface and MockWebSocketServices implementation
func TestWebSocketServices(t *testing.T) {
	t.Run("MockWebSocketServices_DefaultValues", func(t *testing.T) {
		mock := &MockWebSocketServices{}

		// Test default values
		upgrader := mock.GetWebSocketUpgrader()
		if upgrader == nil {
			t.Error("Expected default WebSocket upgrader, got nil")
		}

		routes := mock.GetRoutes()
		if routes == nil {
			t.Error("Expected empty routes map, got nil")
		}

		registry := mock.GetEventRegistry()
		if registry == nil {
			t.Error("Expected default event registry, got nil")
		}

		cookieName := mock.GetCookieName()
		if cookieName != "_session" {
			t.Errorf("Expected default cookie name '_session', got %s", cookieName)
		}

		interval := mock.GetDropDuplicateInterval()
		if interval != 100*time.Millisecond {
			t.Errorf("Expected default interval 100ms, got %v", interval)
		}

		if mock.IsWebSocketDisabled() {
			t.Error("Expected WebSocket to be enabled by default")
		}
	})

	t.Run("MockWebSocketServices_CustomValues", func(t *testing.T) {
		customUpgrader := &websocket.Upgrader{
			ReadBufferSize:  2048,
			WriteBufferSize: 2048,
		}
		customRegistry := event.NewEventRegistry()
		customRoutes := make(map[string]RouteInterface)

		mock := &MockWebSocketServices{
			WebSocketUpgrader:     customUpgrader,
			EventRegistry:         customRegistry,
			Routes:                customRoutes,
			CookieName:            "_custom_session",
			DropDuplicateInterval: 200 * time.Millisecond,
			WebSocketDisabled:     true,
		}

		// Test custom values
		if mock.GetWebSocketUpgrader() != customUpgrader {
			t.Error("Expected custom WebSocket upgrader")
		}

		if mock.GetEventRegistry() != customRegistry {
			t.Error("Expected custom event registry")
		}

		if len(mock.GetRoutes()) != len(customRoutes) {
			t.Error("Expected custom routes map")
		}

		if mock.GetCookieName() != "_custom_session" {
			t.Error("Expected custom cookie name")
		}

		if mock.GetDropDuplicateInterval() != 200*time.Millisecond {
			t.Error("Expected custom interval")
		}

		if !mock.IsWebSocketDisabled() {
			t.Error("Expected WebSocket to be disabled")
		}
	})

	t.Run("MockWebSocketServices_Callbacks", func(t *testing.T) {
		var connectCallCount int
		var disconnectCallCount int
		var lastConnectedUser string
		var lastDisconnectedUser string

		mock := &MockWebSocketServices{
			OnConnectFunc: func(userOrSessionID string) error {
				connectCallCount++
				lastConnectedUser = userOrSessionID
				return nil
			},
			OnDisconnectFunc: func(userOrSessionID string) {
				disconnectCallCount++
				lastDisconnectedUser = userOrSessionID
			},
		}

		// Test connect callback
		err := mock.OnSocketConnect("user123")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if connectCallCount != 1 {
			t.Errorf("Expected connect call count 1, got %d", connectCallCount)
		}
		if lastConnectedUser != "user123" {
			t.Errorf("Expected connected user 'user123', got %s", lastConnectedUser)
		}

		// Test disconnect callback
		mock.OnSocketDisconnect("user456")
		if disconnectCallCount != 1 {
			t.Errorf("Expected disconnect call count 1, got %d", disconnectCallCount)
		}
		if lastDisconnectedUser != "user456" {
			t.Errorf("Expected disconnected user 'user456', got %s", lastDisconnectedUser)
		}
	})

	t.Run("MockWebSocketServices_DecodeSession", func(t *testing.T) {
		mock := &MockWebSocketServices{
			DecodeSessionFunc: func(sessionID string) (string, string, error) {
				return "decoded-" + sessionID, "route-123", nil
			},
		}

		userID, routeID, err := mock.DecodeSession("session456")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if userID != "decoded-session456" {
			t.Errorf("Expected userID 'decoded-session456', got %s", userID)
		}
		if routeID != "route-123" {
			t.Errorf("Expected routeID 'route-123', got %s", routeID)
		}

		// Test default implementation
		defaultMock := &MockWebSocketServices{}
		userID, routeID, err = defaultMock.DecodeSession("test")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if userID != "test" {
			t.Errorf("Expected default userID 'test', got %s", userID)
		}
		if routeID != "mock-route-id" {
			t.Errorf("Expected default routeID 'mock-route-id', got %s", routeID)
		}
	})
}

// TestRouteServicesWithWebSocket tests WebSocket integration with RouteServices
func TestRouteServicesWithWebSocket(t *testing.T) {
	t.Run("NewRouteServicesWithWebSocket", func(t *testing.T) {
		eventRegistry := event.NewEventRegistry()
		pubsub := pubsub.NewInmem()
		renderer := &mockRenderer{}
		options := &Options{AppName: "test"}
		wsServices := &MockWebSocketServices{}

		services := NewRouteServicesWithWebSocket(eventRegistry, pubsub, renderer, options, wsServices)

		if services.EventRegistry != eventRegistry {
			t.Error("EventRegistry not set correctly")
		}
		if services.PubSub != pubsub {
			t.Error("PubSub not set correctly")
		}
		if services.Renderer != renderer {
			t.Error("Renderer not set correctly")
		}
		if services.Options != options {
			t.Error("Options not set correctly")
		}
		if services.WebSocketServices != wsServices {
			t.Error("WebSocketServices not set correctly")
		}
	})

	t.Run("SetAndGetWebSocketServices", func(t *testing.T) {
		services := NewRouteServices(
			event.NewEventRegistry(),
			pubsub.NewInmem(),
			&mockRenderer{},
			&Options{AppName: "test"},
		)

		// Initially no WebSocket services
		if services.HasWebSocketServices() {
			t.Error("Expected no WebSocket services initially")
		}
		if services.GetWebSocketServices() != nil {
			t.Error("Expected nil WebSocket services initially")
		}

		// Set WebSocket services
		wsServices := &MockWebSocketServices{}
		services.SetWebSocketServices(wsServices)

		if !services.HasWebSocketServices() {
			t.Error("Expected WebSocket services to be set")
		}
		if services.GetWebSocketServices() != wsServices {
			t.Error("Expected WebSocket services to match what was set")
		}
	})

	t.Run("ValidateWebSocketServices", func(t *testing.T) {
		services := NewRouteServices(
			event.NewEventRegistry(),
			pubsub.NewInmem(),
			&mockRenderer{},
			&Options{AppName: "test"},
		)

		// Test validation without WebSocket services
		err := services.ValidateWebSocketServices()
		if err == nil {
			t.Error("Expected validation error when WebSocket services are nil")
		}

		// Test validation with valid WebSocket services
		wsServices := &MockWebSocketServices{}
		services.SetWebSocketServices(wsServices)

		err = services.ValidateWebSocketServices()
		if err != nil {
			t.Errorf("Expected validation to pass, got error: %v", err)
		}

		// Test validation with nil upgrader - create a mock that returns nil
		wsServicesWithNilUpgrader := &MockWebSocketServices{}
		// Override the GetWebSocketUpgrader method to return nil
		wsServicesWithNilUpgrader.WebSocketUpgrader = nil

		// We need to create a custom mock that actually returns nil for the upgrader
		customMock := &testMockWebSocketServices{
			eventRegistry: event.NewEventRegistry(),
			upgrader:      nil, // This will cause validation to fail
		}
		services.SetWebSocketServices(customMock)

		err = services.ValidateWebSocketServices()
		if err == nil {
			t.Error("Expected validation error when WebSocket upgrader is nil")
		}
	})

	t.Run("CloneWithWebSocketServices", func(t *testing.T) {
		wsServices := &MockWebSocketServices{CookieName: "test-cookie"}
		services := NewRouteServicesWithWebSocket(
			event.NewEventRegistry(),
			pubsub.NewInmem(),
			&mockRenderer{},
			&Options{AppName: "test"},
			wsServices,
		)

		cloned := services.Clone()

		if cloned.WebSocketServices != wsServices {
			t.Error("WebSocket services should be copied in clone")
		}
		if cloned.WebSocketServices.GetCookieName() != "test-cookie" {
			t.Error("Cloned WebSocket services should have same configuration")
		}
	})
}

// TestWebSocketServicesIntegration tests integration scenarios
func TestWebSocketServicesIntegration(t *testing.T) {
	t.Run("WebSocketUpgraderConfiguration", func(t *testing.T) {
		upgrader := &websocket.Upgrader{
			ReadBufferSize:  4096,
			WriteBufferSize: 4096,
			CheckOrigin: func(r *http.Request) bool {
				return r.Header.Get("Origin") == "https://example.com"
			},
		}

		wsServices := &MockWebSocketServices{
			WebSocketUpgrader: upgrader,
		}

		services := NewRouteServicesWithWebSocket(
			event.NewEventRegistry(),
			pubsub.NewInmem(),
			&mockRenderer{},
			&Options{AppName: "test"},
			wsServices,
		)

		retrievedUpgrader := services.WebSocketServices.GetWebSocketUpgrader()
		if retrievedUpgrader.ReadBufferSize != 4096 {
			t.Error("WebSocket upgrader configuration not preserved")
		}
		if retrievedUpgrader.WriteBufferSize != 4096 {
			t.Error("WebSocket upgrader configuration not preserved")
		}

		// Test CheckOrigin function
		req := &http.Request{
			Header: make(http.Header),
		}
		req.Header.Set("Origin", "https://example.com")
		if !retrievedUpgrader.CheckOrigin(req) {
			t.Error("CheckOrigin should allow https://example.com")
		}

		req.Header.Set("Origin", "https://malicious.com")
		if retrievedUpgrader.CheckOrigin(req) {
			t.Error("CheckOrigin should reject https://malicious.com")
		}
	})

	t.Run("EventRegistryIntegration", func(t *testing.T) {
		eventRegistry := event.NewEventRegistry()

		// Register some test events
		eventRegistry.Register("test-route", "socket-connect", func() {})
		eventRegistry.Register("test-route", "socket-disconnect", func() {})

		wsServices := &MockWebSocketServices{
			EventRegistry: eventRegistry,
		}

		services := NewRouteServicesWithWebSocket(
			eventRegistry,
			pubsub.NewInmem(),
			&mockRenderer{},
			&Options{AppName: "test"},
			wsServices,
		)

		// Verify event registry is accessible through WebSocket services
		wsEventRegistry := services.WebSocketServices.GetEventRegistry()
		if wsEventRegistry != eventRegistry {
			t.Error("Event registry should be the same instance")
		}

		// Verify events are accessible
		_, exists := wsEventRegistry.Get("test-route", "socket-connect")
		if !exists {
			t.Error("Registered event should be accessible through WebSocket services")
		}
	})
}

// testMockWebSocketServices is a custom mock for testing edge cases
type testMockWebSocketServices struct {
	upgrader      *websocket.Upgrader
	eventRegistry event.EventRegistry
	routes        map[string]RouteInterface
}

func (t *testMockWebSocketServices) GetWebSocketUpgrader() *websocket.Upgrader {
	return t.upgrader
}

func (t *testMockWebSocketServices) GetRoutes() map[string]RouteInterface {
	if t.routes == nil {
		return make(map[string]RouteInterface)
	}
	return t.routes
}

func (t *testMockWebSocketServices) GetEventRegistry() event.EventRegistry {
	return t.eventRegistry
}

func (t *testMockWebSocketServices) DecodeSession(sessionID string) (string, string, error) {
	return sessionID, "test-route", nil
}

func (t *testMockWebSocketServices) GetCookieName() string {
	return "_test_session"
}

func (t *testMockWebSocketServices) GetDropDuplicateInterval() time.Duration {
	return 100 * time.Millisecond
}

func (t *testMockWebSocketServices) IsWebSocketDisabled() bool {
	return false
}

func (t *testMockWebSocketServices) OnSocketConnect(userOrSessionID string) error {
	return nil
}

func (t *testMockWebSocketServices) OnSocketDisconnect(userOrSessionID string) {
	// No-op
}
