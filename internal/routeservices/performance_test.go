package routeservices

import (
	"context"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync"
	"testing"

	"github.com/livefir/fir/internal/event"
	"github.com/livefir/fir/internal/eventstate"
	"github.com/livefir/fir/pubsub"
)

// Performance and edge case validation tests for RouteServices

// BenchmarkRouteServicesCreationConcurrent tests concurrent creation of RouteServices
func BenchmarkRouteServicesCreationConcurrent(b *testing.B) {
	eventRegistry := event.NewEventRegistry()
	pubsubAdapter := pubsub.NewInmem()
	renderer := &mockRenderer{}
	options := &Options{AppName: "benchmark"}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = NewRouteServices(eventRegistry, pubsubAdapter, renderer, options)
		}
	})
}

// BenchmarkRouteServicesValidationConcurrent tests concurrent validation
func BenchmarkRouteServicesValidationConcurrent(b *testing.B) {
	services := NewRouteServices(
		event.NewEventRegistry(),
		pubsub.NewInmem(),
		&mockRenderer{},
		&Options{AppName: "benchmark"},
	)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = services.ValidateServices()
		}
	})
}

// BenchmarkRouteServicesCloningConcurrent tests concurrent cloning
func BenchmarkRouteServicesCloningConcurrent(b *testing.B) {
	services := NewRouteServices(
		event.NewEventRegistry(),
		pubsub.NewInmem(),
		&mockRenderer{},
		&Options{AppName: "benchmark"},
	)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = services.Clone()
		}
	})
}

// BenchmarkEventRegistryOperations tests performance of event operations
func BenchmarkEventRegistryOperations(b *testing.B) {
	services := NewRouteServices(
		event.NewEventRegistry(),
		pubsub.NewInmem(),
		&mockRenderer{},
		&Options{AppName: "benchmark"},
	)

	handler := func(data interface{}) {}
	routeID := "benchmark-route"

	b.Run("Register", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			eventID := "event-" + string(rune(i%1000))
			_ = services.EventRegistry.Register(routeID, eventID, handler)
		}
	})

	// Pre-register some events for Get benchmark
	for i := 0; i < 100; i++ {
		eventID := "get-event-" + string(rune(i))
		_ = services.EventRegistry.Register(routeID, eventID, handler)
	}

	b.Run("Get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			eventID := "get-event-" + string(rune(i%100))
			_, _ = services.EventRegistry.Get(routeID, eventID)
		}
	})
}

// BenchmarkPubSubOperations tests PubSub performance
func BenchmarkPubSubOperations(b *testing.B) {
	services := NewRouteServices(
		event.NewEventRegistry(),
		pubsub.NewInmem(),
		&mockRenderer{},
		&Options{AppName: "benchmark"},
	)

	ctx := context.Background()
	channel := "benchmark-channel"
	event := pubsub.Event{
		ID:    func() *string { s := "bench-id"; return &s }(),
		State: eventstate.Type("bench-state"),
	}

	b.Run("Publish", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = services.PubSub.Publish(ctx, channel, event)
		}
	})

	b.Run("Subscribe", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sub, _ := services.PubSub.Subscribe(ctx, channel)
			if sub != nil {
				sub.Close()
			}
		}
	})
}

// BenchmarkMemoryUsage tests memory usage patterns
func BenchmarkMemoryUsage(b *testing.B) {
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	services := make([]*RouteServices, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		services[i] = NewRouteServices(
			event.NewEventRegistry(),
			pubsub.NewInmem(),
			&mockRenderer{},
			&Options{AppName: "memory-test"},
		)
	}

	runtime.GC()
	runtime.ReadMemStats(&m2)

	b.ReportMetric(float64(m2.Alloc-m1.Alloc)/float64(b.N), "bytes/op")

	// Prevent optimization from removing the services
	_ = services[0].ValidateServices()
}

// Edge case tests

// TestRouteServicesEdgeCaseNilValues tests handling of nil values
func TestRouteServicesEdgeCaseNilValues(t *testing.T) {
	tests := []struct {
		name        string
		eventReg    event.EventRegistry
		pubsub      pubsub.Adapter
		renderer    interface{}
		options     *Options
		expectPanic bool
		expectError bool
	}{
		{
			name:        "all nil",
			expectError: true,
		},
		{
			name:        "nil event registry",
			pubsub:      pubsub.NewInmem(),
			renderer:    &mockRenderer{},
			options:     &Options{AppName: "test"},
			expectError: true,
		},
		{
			name:        "nil pubsub",
			eventReg:    event.NewEventRegistry(),
			renderer:    &mockRenderer{},
			options:     &Options{AppName: "test"},
			expectError: true,
		},
		{
			name:        "nil renderer",
			eventReg:    event.NewEventRegistry(),
			pubsub:      pubsub.NewInmem(),
			options:     &Options{AppName: "test"},
			expectError: true,
		},
		{
			name:        "nil options",
			eventReg:    event.NewEventRegistry(),
			pubsub:      pubsub.NewInmem(),
			renderer:    &mockRenderer{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.expectPanic {
						t.Errorf("Unexpected panic: %v", r)
					}
				}
			}()

			services := NewRouteServices(tt.eventReg, tt.pubsub, tt.renderer, tt.options)
			err := services.ValidateServices()

			if tt.expectError && err == nil {
				t.Error("Expected validation error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
		})
	}
}

// TestRouteServicesConcurrentAccess tests concurrent access to RouteServices
func TestRouteServicesConcurrentAccess(t *testing.T) {
	services := NewRouteServices(
		event.NewEventRegistry(),
		pubsub.NewInmem(),
		&mockRenderer{},
		&Options{AppName: "concurrent-test"},
	)

	const numGoroutines = 100
	const numOperations = 50

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numOperations)

	// Test concurrent validation
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				if err := services.ValidateServices(); err != nil {
					errors <- err
				}
			}
		}()
	}

	// Test concurrent cloning
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				clone := services.Clone()
				if err := clone.ValidateServices(); err != nil {
					errors <- err
				}
			}
		}()
	}

	wg.Wait()
	close(errors)

	var errorCount int
	for err := range errors {
		t.Errorf("Concurrent operation error: %v", err)
		errorCount++
	}

	if errorCount > 0 {
		t.Errorf("Got %d errors during concurrent operations", errorCount)
	}
}

// TestRouteServicesWebSocketEdgeCases tests WebSocket-related edge cases
func TestRouteServicesWebSocketEdgeCases(t *testing.T) {
	services := NewRouteServices(
		event.NewEventRegistry(),
		pubsub.NewInmem(),
		&mockRenderer{},
		&Options{
			AppName:          "websocket-test",
			DisableWebsocket: false,
		},
	)

	// Test WebSocket services handling
	mockWSServices := &MockWebSocketServices{}
	services.SetWebSocketServices(mockWSServices)

	retrieved := services.GetWebSocketServices()
	if retrieved != mockWSServices {
		t.Error("WebSocket services should be properly stored and retrieved")
	}

	// Test cloning with WebSocket services
	clone := services.Clone()
	cloneWSServices := clone.GetWebSocketServices()
	if cloneWSServices != mockWSServices {
		t.Error("Clone should have same WebSocket services reference")
	}

	// Test independence after cloning
	differentWSServices := &MockWebSocketServices{}
	clone.SetWebSocketServices(differentWSServices)
	if services.GetWebSocketServices() == clone.GetWebSocketServices() {
		t.Error("WebSocket services should be independent after modification")
	}
}

// TestRouteServicesChannelFunctionEdgeCases tests channel function edge cases
func TestRouteServicesChannelFunctionEdgeCases(t *testing.T) {
	services := NewRouteServices(
		event.NewEventRegistry(),
		pubsub.NewInmem(),
		&mockRenderer{},
		&Options{AppName: "channel-test"},
	)

	// Test with nil channel function
	req := httptest.NewRequest("GET", "/test", nil)
	if services.ChannelFunc != nil {
		t.Error("ChannelFunc should be nil initially")
	}

	// Test setting channel function that returns nil
	services.SetChannelFunc(func(r *http.Request, routeID string) *string {
		return nil
	})

	result := services.ChannelFunc(req, "test-route")
	if result != nil {
		t.Error("Expected nil channel result")
	}

	// Test channel function with empty string
	services.SetChannelFunc(func(r *http.Request, routeID string) *string {
		empty := ""
		return &empty
	})

	result = services.ChannelFunc(req, "test-route")
	if result == nil || *result != "" {
		t.Error("Expected empty string channel result")
	}

	// Test channel function with long string
	longString := string(make([]byte, 10000))
	services.SetChannelFunc(func(r *http.Request, routeID string) *string {
		return &longString
	})

	result = services.ChannelFunc(req, "test-route")
	if result == nil || len(*result) != 10000 {
		t.Error("Expected long string channel result")
	}
}

// TestRouteServicesPathParamsEdgeCases tests path params function edge cases
func TestRouteServicesPathParamsEdgeCases(t *testing.T) {
	services := NewRouteServices(
		event.NewEventRegistry(),
		pubsub.NewInmem(),
		&mockRenderer{},
		&Options{AppName: "pathparams-test"},
	)

	req := httptest.NewRequest("GET", "/test/path", nil)

	// Test with nil path params function
	if services.PathParamsFunc != nil {
		t.Error("PathParamsFunc should be nil initially")
	}

	// Test path params function returning nil
	services.SetPathParamsFunc(func(r *http.Request) map[string]string {
		return nil
	})

	result := services.PathParamsFunc(req)
	if result != nil {
		t.Error("Expected nil path params result")
	}

	// Test path params function returning empty map
	services.SetPathParamsFunc(func(r *http.Request) map[string]string {
		return make(map[string]string)
	})

	result = services.PathParamsFunc(req)
	if result == nil || len(result) != 0 {
		t.Error("Expected empty map path params result")
	}

	// Test path params function with large map
	services.SetPathParamsFunc(func(r *http.Request) map[string]string {
		large := make(map[string]string)
		for i := 0; i < 1000; i++ {
			large["key"+string(rune(i))] = "value" + string(rune(i))
		}
		return large
	})

	result = services.PathParamsFunc(req)
	if result == nil || len(result) != 1000 {
		t.Error("Expected large map path params result")
	}
}

// TestRouteServicesStressTest performs stress testing
func TestRouteServicesStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	const (
		numServices = 1000
		numEvents   = 100
		numClones   = 10
	)

	services := make([]*RouteServices, numServices)

	// Create many RouteServices instances
	for i := 0; i < numServices; i++ {
		services[i] = NewRouteServices(
			event.NewEventRegistry(),
			pubsub.NewInmem(),
			&mockRenderer{},
			&Options{AppName: "stress-test"},
		)

		// Register events
		for j := 0; j < numEvents; j++ {
			routeID := "route-" + string(rune(i))
			eventID := "event-" + string(rune(j))
			handler := func(data interface{}) {}

			err := services[i].EventRegistry.Register(routeID, eventID, handler)
			if err != nil {
				t.Fatalf("Failed to register event %s:%s - %v", routeID, eventID, err)
			}
		}

		// Create clones
		for k := 0; k < numClones; k++ {
			clone := services[i].Clone()
			if err := clone.ValidateServices(); err != nil {
				t.Fatalf("Clone validation failed: %v", err)
			}
		}
	}

	// Validate all services
	for i, service := range services {
		if err := service.ValidateServices(); err != nil {
			t.Errorf("Service %d validation failed: %v", i, err)
		}
	}

	t.Logf("Successfully created and validated %d RouteServices with %d events each", numServices, numEvents)
}
