package routeservices

import (
	"html/template"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/schema"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/websocket"
	"github.com/livefir/fir/internal/event"
	"github.com/livefir/fir/pubsub"
	"github.com/patrickmn/go-cache"
)

// mockRenderer implements the renderer interface for testing
type mockRenderer struct{}

func (m *mockRenderer) RenderRoute(ctx interface{}, data interface{}, isError bool) {
	// Mock implementation
}

func (m *mockRenderer) RenderDOMEvents(ctx interface{}, event interface{}) interface{} {
	// Mock implementation
	return nil
}

// mockTemplateEngine implements the template engine interface for testing
type mockTemplateEngine struct{}

func (m *mockTemplateEngine) LoadTemplate(config interface{}) (interface{}, error) {
	return nil, nil
}

func (m *mockTemplateEngine) Render(template interface{}, data interface{}, w interface{}) error {
	return nil
}

func TestNewRouteServices(t *testing.T) {
	// Setup test dependencies
	eventRegistry := event.NewEventRegistry()
	pubsubAdapter := pubsub.NewInmem()
	renderer := &mockRenderer{}

	options := &Options{
		AppName:               "test-app",
		DisableTemplateCache:  true,
		DisableWebsocket:      false,
		DevelopmentMode:       true,
		DebugLog:              true,
		PublicDir:             ".",
		DropDuplicateInterval: 250 * time.Millisecond,
		WebsocketUpgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		FormDecoder:  schema.NewDecoder(),
		CookieName:   "_test_session_",
		SecureCookie: securecookie.New([]byte("test-hash-key"), []byte("test-block-key")),
		Cache:        cache.New(5*time.Minute, 10*time.Minute),
		FuncMap:      make(template.FuncMap),
	}

	// Create RouteServices
	services := NewRouteServices(eventRegistry, pubsubAdapter, renderer, options)

	// Test that all services are properly set
	if services.EventRegistry == nil {
		t.Error("EventRegistry should not be nil")
	}

	if services.PubSub == nil {
		t.Error("PubSub should not be nil")
	}

	if services.Renderer == nil {
		t.Error("Renderer should not be nil")
	}

	if services.Options == nil {
		t.Error("Options should not be nil")
	}

	// Test that the options are correctly set
	if services.Options.AppName != "test-app" {
		t.Errorf("Expected AppName to be 'test-app', got %s", services.Options.AppName)
	}

	if !services.Options.DevelopmentMode {
		t.Error("Expected DevelopmentMode to be true")
	}
}

func TestSetChannelFunc(t *testing.T) {
	services := &RouteServices{}

	testChannelFunc := func(r *http.Request, routeID string) *string {
		channel := "test-channel"
		return &channel
	}

	services.SetChannelFunc(testChannelFunc)

	if services.ChannelFunc == nil {
		t.Error("ChannelFunc should not be nil after setting")
	}

	// Test the function works
	req, _ := http.NewRequest("GET", "/test", nil)
	result := services.ChannelFunc(req, "test-route")
	if result == nil || *result != "test-channel" {
		t.Error("ChannelFunc should return 'test-channel'")
	}
}

func TestSetPathParamsFunc(t *testing.T) {
	services := &RouteServices{}

	testPathParamsFunc := func(r *http.Request) map[string]string {
		return map[string]string{"id": "123"}
	}

	services.SetPathParamsFunc(testPathParamsFunc)

	if services.PathParamsFunc == nil {
		t.Error("PathParamsFunc should not be nil after setting")
	}

	// Test the function works
	req, _ := http.NewRequest("GET", "/test/123", nil)
	result := services.PathParamsFunc(req)
	if result["id"] != "123" {
		t.Error("PathParamsFunc should return id=123")
	}
}

func TestRouteServicesIntegration(t *testing.T) {
	// Test that RouteServices can be created with all real dependencies
	eventRegistry := event.NewEventRegistry()
	pubsubAdapter := pubsub.NewInmem()
	renderer := &mockRenderer{}

	options := &Options{
		AppName:               "integration-test",
		DisableTemplateCache:  false,
		DisableWebsocket:      false,
		DevelopmentMode:       false,
		DebugLog:              false,
		PublicDir:             "./templates",
		DropDuplicateInterval: 100 * time.Millisecond,
		WebsocketUpgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		FormDecoder:  schema.NewDecoder(),
		CookieName:   "_fir_session_",
		SecureCookie: securecookie.New(securecookie.GenerateRandomKey(64), securecookie.GenerateRandomKey(32)),
		Cache:        cache.New(10*time.Minute, 20*time.Minute),
		FuncMap:      template.FuncMap{"test": func() string { return "test" }},
	}

	services := NewRouteServices(eventRegistry, pubsubAdapter, renderer, options)

	// Set up functions
	services.SetChannelFunc(func(r *http.Request, routeID string) *string {
		channel := "channel-" + routeID
		return &channel
	})

	services.SetPathParamsFunc(func(r *http.Request) map[string]string {
		return map[string]string{"path": r.URL.Path}
	})

	// Test that all components work together
	if services.EventRegistry == nil || services.PubSub == nil || services.Renderer == nil {
		t.Error("Core services should be available")
	}

	if services.ChannelFunc == nil || services.PathParamsFunc == nil {
		t.Error("Function services should be available")
	}

	// Test channel function
	req, _ := http.NewRequest("GET", "/test", nil)
	channel := services.ChannelFunc(req, "test-route")
	if channel == nil || *channel != "channel-test-route" {
		t.Error("Channel function should work correctly")
	}

	// Test path params function
	params := services.PathParamsFunc(req)
	if params["path"] != "/test" {
		t.Error("Path params function should work correctly")
	}
}

func TestValidateServices(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func() *RouteServices
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid services",
			setupFunc: func() *RouteServices {
				return &RouteServices{
					EventRegistry: event.NewEventRegistry(),
					PubSub:        pubsub.NewInmem(),
					Renderer:      &mockRenderer{},
					Options:       &Options{AppName: "test"},
				}
			},
			expectError: false,
		},
		{
			name: "missing EventRegistry",
			setupFunc: func() *RouteServices {
				return &RouteServices{
					PubSub:   pubsub.NewInmem(),
					Renderer: &mockRenderer{},
					Options:  &Options{AppName: "test"},
				}
			},
			expectError: true,
			errorMsg:    "eventRegistry is required but not set",
		},
		{
			name: "missing PubSub",
			setupFunc: func() *RouteServices {
				return &RouteServices{
					EventRegistry: event.NewEventRegistry(),
					Renderer:      &mockRenderer{},
					Options:       &Options{AppName: "test"},
				}
			},
			expectError: true,
			errorMsg:    "pubSub is required but not set",
		},
		{
			name: "missing Renderer",
			setupFunc: func() *RouteServices {
				return &RouteServices{
					EventRegistry: event.NewEventRegistry(),
					PubSub:        pubsub.NewInmem(),
					Options:       &Options{AppName: "test"},
				}
			},
			expectError: true,
			errorMsg:    "renderer is required but not set",
		},
		{
			name: "missing Options",
			setupFunc: func() *RouteServices {
				return &RouteServices{
					EventRegistry: event.NewEventRegistry(),
					PubSub:        pubsub.NewInmem(),
					Renderer:      &mockRenderer{},
				}
			},
			expectError: true,
			errorMsg:    "options is required but not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			services := tt.setupFunc()
			err := services.ValidateServices()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestUpdateOptions(t *testing.T) {
	services := &RouteServices{
		Options: &Options{AppName: "old-name"},
	}

	newOptions := &Options{
		AppName:         "new-name",
		DevelopmentMode: true,
		DebugLog:        true,
	}

	services.UpdateOptions(newOptions)

	if services.Options.AppName != "new-name" {
		t.Errorf("Expected AppName to be 'new-name', got %s", services.Options.AppName)
	}

	if !services.Options.DevelopmentMode {
		t.Error("Expected DevelopmentMode to be true")
	}

	if !services.Options.DebugLog {
		t.Error("Expected DebugLog to be true")
	}
}

func TestClone(t *testing.T) {
	mockTemplateEngine := &mockTemplateEngine{}

	original := &RouteServices{
		EventRegistry:  event.NewEventRegistry(),
		PubSub:         pubsub.NewInmem(),
		Renderer:       &mockRenderer{},
		TemplateEngine: mockTemplateEngine,
		Options:        &Options{AppName: "original"},
	}

	original.SetChannelFunc(func(r *http.Request, routeID string) *string {
		channel := "test"
		return &channel
	})

	original.SetPathParamsFunc(func(r *http.Request) map[string]string {
		return map[string]string{"test": "value"}
	})

	clone := original.Clone()

	// Test that all fields are copied
	if clone.EventRegistry != original.EventRegistry {
		t.Error("EventRegistry should be the same reference")
	}

	if clone.PubSub != original.PubSub {
		t.Error("PubSub should be the same reference")
	}

	if clone.Renderer != original.Renderer {
		t.Error("Renderer should be the same reference")
	}

	if clone.TemplateEngine != original.TemplateEngine {
		t.Error("TemplateEngine should be the same reference")
	}

	if clone.Options != original.Options {
		t.Error("Options should be the same reference")
	}

	// Test that functions work
	if clone.ChannelFunc == nil {
		t.Error("ChannelFunc should be copied")
	}

	if clone.PathParamsFunc == nil {
		t.Error("PathParamsFunc should be copied")
	}

	// Test that clone is independent (modify clone options)
	clone.Options = &Options{AppName: "cloned"}
	if original.Options.AppName == "cloned" {
		t.Error("Original options should not be affected by clone modification")
	}
}

// Integration test for RouteServices lifecycle
func TestRouteServicesLifecycle(t *testing.T) {
	// Create services
	eventRegistry := event.NewEventRegistry()
	pubsubAdapter := pubsub.NewInmem()
	renderer := &mockRenderer{}
	options := &Options{
		AppName:         "lifecycle-test",
		DevelopmentMode: true,
	}

	services := NewRouteServices(eventRegistry, pubsubAdapter, renderer, options)

	// Validate initial state
	if err := services.ValidateServices(); err != nil {
		t.Fatalf("Initial validation failed: %v", err)
	}

	// Set functions
	services.SetChannelFunc(func(r *http.Request, routeID string) *string {
		channel := "channel-" + routeID
		return &channel
	})

	services.SetPathParamsFunc(func(r *http.Request) map[string]string {
		return map[string]string{"id": "123"}
	})

	// Clone and test independence
	clone := services.Clone()
	clone.UpdateOptions(&Options{AppName: "cloned-app"})

	// Original should be unchanged
	if services.Options.AppName != "lifecycle-test" {
		t.Error("Original services should not be affected by clone updates")
	}

	// Clone should be updated
	if clone.Options.AppName != "cloned-app" {
		t.Error("Clone should have updated options")
	}

	// Both should still validate
	if err := services.ValidateServices(); err != nil {
		t.Errorf("Original services validation failed: %v", err)
	}

	if err := clone.ValidateServices(); err != nil {
		t.Errorf("Clone services validation failed: %v", err)
	}
}
