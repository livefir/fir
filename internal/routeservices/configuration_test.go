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

// Configuration validation tests for different RouteServices options

// TestRouteServicesConfigurationOptions tests various configuration combinations
func TestRouteServicesConfigurationOptions(t *testing.T) {
	baseOptions := &Options{
		AppName:         "config-test",
		DevelopmentMode: true,
	}

	tests := []struct {
		name    string
		options *Options
		valid   bool
	}{
		{
			name:    "minimal configuration",
			options: baseOptions,
			valid:   true,
		},
		{
			name: "full production configuration",
			options: &Options{
				AppName:               "prod-app",
				DisableTemplateCache:  false,
				DisableWebsocket:      false,
				EnableWatch:           false,
				WatchExts:             []string{".html", ".css", ".js"},
				PublicDir:             "/static",
				DevelopmentMode:       false,
				DropDuplicateInterval: 100 * time.Millisecond,
				DebugLog:              false,
				FormDecoder:           schema.NewDecoder(),
				CookieName:            "_app_session_",
				SecureCookie:          securecookie.New([]byte("hash-key"), []byte("block-key")),
				Cache:                 cache.New(30*time.Minute, 1*time.Hour),
				FuncMap:               template.FuncMap{"upper": func(s string) string { return s }},
				WebsocketUpgrader: websocket.Upgrader{
					ReadBufferSize:  1024,
					WriteBufferSize: 1024,
					CheckOrigin:     func(r *http.Request) bool { return true },
				},
				OnSocketConnect: func(userOrSessionID string) error {
					return nil
				},
				OnSocketDisconnect: func(userOrSessionID string) {
					// Cleanup logic
				},
				ReadFile: func(filename string) (string, []byte, error) {
					return filename, []byte("file content"), nil
				},
				ExistFile: func(filename string) bool {
					return true
				},
			},
			valid: true,
		},
		{
			name: "development configuration",
			options: &Options{
				AppName:               "dev-app",
				DisableTemplateCache:  true,
				DisableWebsocket:      false,
				EnableWatch:           true,
				WatchExts:             []string{".go", ".html", ".css", ".js"},
				PublicDir:             "./public",
				DevelopmentMode:       true,
				DropDuplicateInterval: 50 * time.Millisecond,
				DebugLog:              true,
				FormDecoder:           schema.NewDecoder(),
				CookieName:            "_dev_session_",
				Cache:                 cache.New(5*time.Minute, 10*time.Minute),
				FuncMap:               make(template.FuncMap),
				WebsocketUpgrader: websocket.Upgrader{
					CheckOrigin: func(r *http.Request) bool { return true },
				},
				ReadFile: func(filename string) (string, []byte, error) {
					return filename, []byte("dev content"), nil
				},
				ExistFile: func(filename string) bool {
					return filename != "nonexistent.file"
				},
			},
			valid: true,
		},
		{
			name: "websocket disabled configuration",
			options: &Options{
				AppName:              "no-ws-app",
				DisableWebsocket:     true,
				DisableTemplateCache: false,
				DevelopmentMode:      false,
				DebugLog:             false,
				FormDecoder:          schema.NewDecoder(),
				CookieName:           "_noweb_session_",
				Cache:                cache.New(1*time.Hour, 2*time.Hour),
				FuncMap:              template.FuncMap{},
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			services := NewRouteServices(
				event.NewEventRegistry(),
				pubsub.NewInmem(),
				&mockRenderer{},
				tt.options,
			)

			err := services.ValidateServices()
			if tt.valid && err != nil {
				t.Errorf("Expected valid configuration but got error: %v", err)
			} else if !tt.valid && err == nil {
				t.Error("Expected invalid configuration but validation passed")
			}

			// Test configuration values
			if services.Options.AppName != tt.options.AppName {
				t.Errorf("Expected AppName %s, got %s", tt.options.AppName, services.Options.AppName)
			}

			if services.Options.DevelopmentMode != tt.options.DevelopmentMode {
				t.Errorf("Expected DevelopmentMode %v, got %v", tt.options.DevelopmentMode, services.Options.DevelopmentMode)
			}
		})
	}
}

// TestRouteServicesConfigurationUpdates tests runtime configuration updates
func TestRouteServicesConfigurationUpdates(t *testing.T) {
	initialOptions := &Options{
		AppName:         "initial-app",
		DevelopmentMode: false,
		DebugLog:        false,
	}

	services := NewRouteServices(
		event.NewEventRegistry(),
		pubsub.NewInmem(),
		&mockRenderer{},
		initialOptions,
	)

	// Verify initial configuration
	if services.Options.AppName != "initial-app" {
		t.Error("Initial app name should be set correctly")
	}

	if services.Options.DevelopmentMode {
		t.Error("Initial development mode should be false")
	}

	// Update configuration
	updatedOptions := &Options{
		AppName:              "updated-app",
		DevelopmentMode:      true,
		DebugLog:             true,
		DisableTemplateCache: true,
		Cache:                cache.New(1*time.Minute, 2*time.Minute),
	}

	services.UpdateOptions(updatedOptions)

	// Verify updated configuration
	if services.Options.AppName != "updated-app" {
		t.Error("App name should be updated")
	}

	if !services.Options.DevelopmentMode {
		t.Error("Development mode should be enabled")
	}

	if !services.Options.DebugLog {
		t.Error("Debug log should be enabled")
	}

	if !services.Options.DisableTemplateCache {
		t.Error("Template cache should be disabled")
	}

	// Verify services still validate after update
	if err := services.ValidateServices(); err != nil {
		t.Errorf("Services should still validate after configuration update: %v", err)
	}
}

// TestRouteServicesConfigurationCloning tests configuration behavior during cloning
func TestRouteServicesConfigurationCloning(t *testing.T) {
	originalOptions := &Options{
		AppName:         "original-app",
		DevelopmentMode: true,
		DebugLog:        true,
		Cache:           cache.New(10*time.Minute, 20*time.Minute),
	}

	original := NewRouteServices(
		event.NewEventRegistry(),
		pubsub.NewInmem(),
		&mockRenderer{},
		originalOptions,
	)

	// Clone the services
	clone := original.Clone()

	// Verify clone has same configuration reference (shallow copy)
	if clone.Options != original.Options {
		t.Error("Clone should share the same Options reference")
	}

	// Update clone's options
	cloneOptions := &Options{
		AppName:         "cloned-app",
		DevelopmentMode: false,
		DebugLog:        false,
	}

	clone.UpdateOptions(cloneOptions)

	// Verify independence
	if original.Options.AppName == clone.Options.AppName {
		t.Error("Original and clone should have different options after update")
	}

	if original.Options.AppName != "original-app" {
		t.Error("Original options should be unchanged")
	}

	if clone.Options.AppName != "cloned-app" {
		t.Error("Clone options should be updated")
	}
}

// TestRouteServicesConfigurationValidation tests configuration validation edge cases
func TestRouteServicesConfigurationValidation(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func() *Options
		expectPass bool
	}{
		{
			name: "empty app name",
			setupFunc: func() *Options {
				return &Options{
					AppName: "",
				}
			},
			expectPass: true, // Empty app name is allowed
		},
		{
			name: "very long app name",
			setupFunc: func() *Options {
				return &Options{
					AppName: string(make([]byte, 10000)),
				}
			},
			expectPass: true,
		},
		{
			name: "zero drop duplicate interval",
			setupFunc: func() *Options {
				return &Options{
					AppName:               "test",
					DropDuplicateInterval: 0,
				}
			},
			expectPass: true,
		},
		{
			name: "negative drop duplicate interval",
			setupFunc: func() *Options {
				return &Options{
					AppName:               "test",
					DropDuplicateInterval: -1 * time.Millisecond,
				}
			},
			expectPass: true, // Negative values are allowed (might disable feature)
		},
		{
			name: "nil function map",
			setupFunc: func() *Options {
				return &Options{
					AppName: "test",
					FuncMap: nil,
				}
			},
			expectPass: true,
		},
		{
			name: "nil cache",
			setupFunc: func() *Options {
				return &Options{
					AppName: "test",
					Cache:   nil,
				}
			},
			expectPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := tt.setupFunc()
			services := NewRouteServices(
				event.NewEventRegistry(),
				pubsub.NewInmem(),
				&mockRenderer{},
				options,
			)

			err := services.ValidateServices()
			if tt.expectPass && err != nil {
				t.Errorf("Expected validation to pass but got error: %v", err)
			} else if !tt.expectPass && err == nil {
				t.Error("Expected validation to fail but it passed")
			}
		})
	}
}

// TestRouteServicesConfigurationPerformance tests performance with different configurations
func TestRouteServicesConfigurationPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping configuration performance test in short mode")
	}

	configurations := []*Options{
		// Minimal configuration
		{AppName: "minimal"},

		// Full configuration
		{
			AppName:               "full",
			DisableTemplateCache:  false,
			DisableWebsocket:      false,
			EnableWatch:           true,
			WatchExts:             []string{".go", ".html", ".css", ".js"},
			PublicDir:             "/static",
			DevelopmentMode:       true,
			DropDuplicateInterval: 100 * time.Millisecond,
			DebugLog:              true,
			FormDecoder:           schema.NewDecoder(),
			CookieName:            "_session_",
			SecureCookie:          securecookie.New([]byte("key"), []byte("block")),
			Cache:                 cache.New(30*time.Minute, 1*time.Hour),
			FuncMap:               template.FuncMap{"test": func() string { return "test" }},
			ReadFile: func(filename string) (string, []byte, error) {
				return filename, []byte("content"), nil
			},
			ExistFile: func(filename string) bool {
				return true
			},
		},
	}

	for i, config := range configurations {
		t.Run("config_"+string(rune(i)), func(t *testing.T) {
			const numIterations = 1000

			start := time.Now()
			for j := 0; j < numIterations; j++ {
				services := NewRouteServices(
					event.NewEventRegistry(),
					pubsub.NewInmem(),
					&mockRenderer{},
					config,
				)

				if err := services.ValidateServices(); err != nil {
					t.Fatalf("Validation failed: %v", err)
				}
			}
			duration := time.Since(start)

			t.Logf("Configuration %d: %d iterations in %v (%.2f μs/op)",
				i, numIterations, duration, float64(duration.Nanoseconds())/float64(numIterations)/1000)
		})
	}
}
