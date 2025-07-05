package routeservices

import (
	"context"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/schema"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/websocket"
	"github.com/livefir/fir/internal/event"
	"github.com/livefir/fir/internal/eventstate"
	"github.com/livefir/fir/pubsub"
	"github.com/patrickmn/go-cache"
)

// Integration tests for RouteServices with real HTTP scenarios

// fullMockRenderer provides a more complete mock for integration testing
type fullMockRenderer struct {
	renderCount    int
	lastRenderData interface{}
	domEventCount  int
	lastDOMEvent   interface{}
}

func (r *fullMockRenderer) RenderRoute(ctx interface{}, data interface{}, isError bool) {
	r.renderCount++
	r.lastRenderData = data
}

func (r *fullMockRenderer) RenderDOMEvents(ctx interface{}, event interface{}) interface{} {
	r.domEventCount++
	r.lastDOMEvent = event
	return map[string]interface{}{
		"event":     event,
		"processed": true,
	}
}

func createTestRouteServices() (*RouteServices, *fullMockRenderer) {
	eventRegistry := event.NewEventRegistry()
	pubsubAdapter := pubsub.NewInmem()
	renderer := &fullMockRenderer{}

	options := &Options{
		AppName:               "integration-test",
		DisableTemplateCache:  true,
		DisableWebsocket:      false,
		DevelopmentMode:       true,
		DebugLog:              true,
		PublicDir:             ".",
		DropDuplicateInterval: 50 * time.Millisecond,
		WebsocketUpgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		FormDecoder:  schema.NewDecoder(),
		CookieName:   "_test_session_",
		SecureCookie: securecookie.New([]byte("test-hash-key-32-chars-long!!"), []byte("test-block-key-32-chars-long!")),
		Cache:        cache.New(5*time.Minute, 10*time.Minute),
		FuncMap:      make(template.FuncMap),
		ReadFile: func(filename string) (string, []byte, error) {
			return filename, []byte("test content"), nil
		},
		ExistFile: func(filename string) bool {
			return true
		},
	}

	services := NewRouteServices(eventRegistry, pubsubAdapter, renderer, options)

	// Set up channel and path params functions
	services.SetChannelFunc(func(r *http.Request, routeID string) *string {
		channel := "test-channel-" + routeID
		return &channel
	})

	services.SetPathParamsFunc(func(r *http.Request) map[string]string {
		// Simple path parameter extraction for testing
		return map[string]string{
			"path":   r.URL.Path,
			"method": r.Method,
		}
	})

	return services, renderer
}

func TestRouteServicesHTTPIntegration(t *testing.T) {
	services, renderer := createTestRouteServices()

	// Validate services are properly configured
	if err := services.ValidateServices(); err != nil {
		t.Fatalf("Services validation failed: %v", err)
	}

	// Test channel function with HTTP request
	req := httptest.NewRequest("GET", "/test/123", nil)
	channel := services.ChannelFunc(req, "test-route")

	if channel == nil {
		t.Fatal("Channel function should return a channel")
	}

	if *channel != "test-channel-test-route" {
		t.Errorf("Expected channel 'test-channel-test-route', got '%s'", *channel)
	}

	// Test path params function
	params := services.PathParamsFunc(req)
	if params["path"] != "/test/123" {
		t.Errorf("Expected path '/test/123', got '%s'", params["path"])
	}

	if params["method"] != "GET" {
		t.Errorf("Expected method 'GET', got '%s'", params["method"])
	}

	// Test renderer integration
	renderer.RenderRoute(nil, map[string]string{"test": "data"}, false)
	if renderer.renderCount != 1 {
		t.Errorf("Expected render count 1, got %d", renderer.renderCount)
	}

	// Test DOM events
	eventData := map[string]interface{}{"type": "click", "target": "button"}
	result := renderer.RenderDOMEvents(nil, eventData)
	if renderer.domEventCount != 1 {
		t.Errorf("Expected DOM event count 1, got %d", renderer.domEventCount)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok || !resultMap["processed"].(bool) {
		t.Error("DOM event should be processed correctly")
	}
}

func TestRouteServicesWebSocketIntegration(t *testing.T) {
	services, _ := createTestRouteServices()

	// Create mock WebSocket services
	wsServices := &MockWebSocketServices{}
	services.SetWebSocketServices(wsServices)

	// Test WebSocket integration
	retrievedWSServices := services.GetWebSocketServices()
	if retrievedWSServices != wsServices {
		t.Error("WebSocket services should be properly stored and retrieved")
	}

	// Test WebSocket validation
	if err := services.ValidateWebSocketServices(); err != nil {
		t.Errorf("WebSocket services should validate: %v", err)
	}
}

func TestRouteServicesPubSubIntegration(t *testing.T) {
	services, _ := createTestRouteServices()

	// Test PubSub integration
	testChannel := "test-integration-channel"
	testEvent := pubsub.Event{
		ID:    func() *string { s := "test-event-id"; return &s }(),
		State: func() eventstate.Type { return eventstate.Type("test-state") }(),
	}

	ctx := context.Background()

	// Subscribe to a channel
	subscription, err := services.PubSub.Subscribe(ctx, testChannel)
	if err != nil {
		t.Fatalf("Failed to subscribe to channel: %v", err)
	}
	defer subscription.Close()

	// Publish a message
	err = services.PubSub.Publish(ctx, testChannel, testEvent)
	if err != nil {
		t.Fatalf("Failed to publish message: %v", err)
	}

	// Receive the message
	select {
	case msg := <-subscription.C():
		if msg.ID == nil || *msg.ID != *testEvent.ID {
			t.Errorf("Expected event ID '%s', got '%v'", *testEvent.ID, msg.ID)
		}
	case <-time.After(1 * time.Second):
		t.Error("Message not received within timeout")
	}
}

func TestRouteServicesEventRegistryIntegration(t *testing.T) {
	services, _ := createTestRouteServices()

	// Test EventRegistry integration
	testRouteID := "test-route"
	testEventID := "test-event"
	testHandler := func(data interface{}) {
		// Test handler function
	}

	// Register an event handler
	err := services.EventRegistry.Register(testRouteID, testEventID, testHandler)
	if err != nil {
		t.Fatalf("Failed to register event handler: %v", err)
	}

	// Get the registered handler
	handler, found := services.EventRegistry.Get(testRouteID, testEventID)
	if !found {
		t.Error("Event handler should be found")
	}
	if handler == nil {
		t.Error("Event handler should not be nil")
	}

	// Test getting route events
	routeEvents := services.EventRegistry.GetRouteEvents(testRouteID)
	if len(routeEvents) != 1 {
		t.Errorf("Expected 1 event for route, got %d", len(routeEvents))
	}

	// Test getting all events
	allEvents := services.EventRegistry.GetAllEvents()
	if len(allEvents) == 0 {
		t.Error("Should have registered events")
	}

	if routeEvents, exists := allEvents[testRouteID]; !exists || len(routeEvents) != 1 {
		t.Error("Route should have exactly one registered event")
	}
}

func TestRouteServicesConfigurationIntegration(t *testing.T) {
	services, _ := createTestRouteServices()

	// Test various configuration options
	options := services.Options

	// Test file system integration
	filename, content, err := options.ReadFile("test.txt")
	if err != nil {
		t.Errorf("ReadFile should not error: %v", err)
	}
	if filename != "test.txt" {
		t.Errorf("Expected filename 'test.txt', got '%s'", filename)
	}
	if string(content) != "test content" {
		t.Errorf("Expected content 'test content', got '%s'", string(content))
	}

	// Test file existence check
	if !options.ExistFile("any-file") {
		t.Error("ExistFile should return true for any file in test")
	}

	// Test cache integration
	cacheKey := "test-key"
	cacheValue := "test-value"
	options.Cache.Set(cacheKey, cacheValue, time.Minute)

	retrievedValue, found := options.Cache.Get(cacheKey)
	if !found {
		t.Error("Cache value should be found")
	}
	if retrievedValue != cacheValue {
		t.Errorf("Expected cache value '%s', got '%v'", cacheValue, retrievedValue)
	}

	// Test form decoder
	if options.FormDecoder == nil {
		t.Error("FormDecoder should be available")
	}

	// Test secure cookie
	if options.SecureCookie == nil {
		t.Error("SecureCookie should be available")
	}
}

func TestRouteServicesCloneIndependence(t *testing.T) {
	original, _ := createTestRouteServices()

	// Modify original
	original.Options.AppName = "original-app"
	originalWSServices := &MockWebSocketServices{}
	original.SetWebSocketServices(originalWSServices)

	// Clone and modify
	clone := original.Clone()
	clone.UpdateOptions(&Options{
		AppName:         "cloned-app",
		DevelopmentMode: false,
	})
	cloneWSServices := &MockWebSocketServices{}
	clone.SetWebSocketServices(cloneWSServices)

	// Test independence
	if original.Options.AppName == clone.Options.AppName {
		t.Error("Clone should have independent options")
	}

	if original.GetWebSocketServices() == clone.GetWebSocketServices() {
		t.Error("Clone should have independent WebSocket services reference")
	}

	// Test that both still validate and work
	if err := original.ValidateServices(); err != nil {
		t.Errorf("Original should still validate: %v", err)
	}

	if err := clone.ValidateServices(); err != nil {
		t.Errorf("Clone should validate: %v", err)
	}

	// Test that core services are still shared references (efficient)
	if original.EventRegistry != clone.EventRegistry {
		t.Error("EventRegistry should be shared reference")
	}

	if original.PubSub != clone.PubSub {
		t.Error("PubSub should be shared reference")
	}
}

// Benchmark tests for performance validation
func BenchmarkRouteServicesCreation(b *testing.B) {
	eventRegistry := event.NewEventRegistry()
	pubsubAdapter := pubsub.NewInmem()
	renderer := &mockRenderer{}
	options := &Options{AppName: "benchmark"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewRouteServices(eventRegistry, pubsubAdapter, renderer, options)
	}
}

func BenchmarkRouteServicesValidation(b *testing.B) {
	services, _ := createTestRouteServices()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = services.ValidateServices()
	}
}

func BenchmarkRouteServicesClone(b *testing.B) {
	services, _ := createTestRouteServices()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = services.Clone()
	}
}
