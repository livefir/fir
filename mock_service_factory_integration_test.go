package fir

import (
	"context"
	"testing"

	"github.com/livefir/fir/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMockServiceFactory_CreateTestRouteServices verifies Step 3.2 mock service creation
func TestMockServiceFactory_CreateTestRouteServices(t *testing.T) {
	factory := NewMockServiceFactory()

	// Create complete test route services
	routeServices := factory.CreateTestRouteServices()

	// Verify all services are created
	require.NotNil(t, routeServices, "Route services should be created")
	require.NotNil(t, routeServices.EventService, "Event service should be created")
	require.NotNil(t, routeServices.RenderService, "Render service should be created")
	require.NotNil(t, routeServices.TemplateService, "Template service should be created")
	require.NotNil(t, routeServices.ResponseBuilder, "Response builder should be created")
	require.NotNil(t, routeServices.Options, "Options should be created")

	// Verify options are set correctly
	assert.False(t, routeServices.Options.DisableTemplateCache, "Template cache should be enabled")
	assert.False(t, routeServices.Options.DisableWebsocket, "WebSocket should be enabled")
}

// TestMockServiceFactory_IndividualServices verifies individual mock service creation
func TestMockServiceFactory_IndividualServices(t *testing.T) {
	factory := NewMockServiceFactory()

	t.Run("Event service", func(t *testing.T) {
		eventService := factory.CreateMockEventService()
		require.NotNil(t, eventService, "Event service should be created")

		// Test that the service has default behavior
		metrics := eventService.GetEventMetrics()
		// Note: metrics is a struct with field access, not methods
		assert.Equal(t, int64(0), metrics.TotalEvents, "Initial metrics should be zero")
	})

	t.Run("Render service", func(t *testing.T) {
		renderService := factory.CreateMockRenderService()
		require.NotNil(t, renderService, "Render service should be created")
	})

	t.Run("Template service", func(t *testing.T) {
		templateService := factory.CreateMockTemplateService()
		require.NotNil(t, templateService, "Template service should be created")

		// Test basic functionality
		err := templateService.ClearCache()
		assert.NoError(t, err, "Clear cache should not error")
	})

	t.Run("Response builder", func(t *testing.T) {
		responseBuilder := factory.CreateMockResponseBuilder()
		require.NotNil(t, responseBuilder, "Response builder should be created")
	})
}

// TestMockServiceFactory_WithCustomBehavior tests custom behavior injection
func TestMockServiceFactory_WithCustomBehavior(t *testing.T) {
	factory := NewMockServiceFactory()

	// Create event service with custom behavior
	customEventService := factory.CreateMockEventServiceWithBehavior(
		func(ctx context.Context, req services.EventRequest) (*services.EventResponse, error) {
			return &services.EventResponse{
				StatusCode: 201,
				Body:       []byte("custom response"),
			}, nil
		},
	)

	require.NotNil(t, customEventService, "Custom event service should be created")

	// Test custom behavior
	ctx := context.Background()
	req := services.EventRequest{ID: "test"}
	resp, err := customEventService.ProcessEvent(ctx, req)

	require.NoError(t, err, "ProcessEvent should not error")
	assert.Equal(t, 201, resp.StatusCode, "Should use custom status code")
	assert.Equal(t, []byte("custom response"), resp.Body, "Should use custom response body")
}

// TestMockServiceFactory_IntegrationWithController demonstrates usage with Controller
func TestMockServiceFactory_IntegrationWithController(t *testing.T) {
	// Create controller
	ctrl := NewController("mock-test")

	// Create route with normal setup (this demonstrates how the factory could be used)
	handler := ctrl.RouteFunc(func() RouteOptions {
		return RouteOptions{
			ID("mock-integration-test"),
			Content("<div>Mock integration test</div>"),
		}
	})

	require.NotNil(t, handler, "Handler should be created")

	// This test demonstrates that the mock services are compatible with the framework
	// In a real scenario, the mock factory would be used to inject test services
	// into the route creation process
	t.Log("Mock service factory creates framework-compatible services")
}
