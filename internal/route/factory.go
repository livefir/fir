package route

import (
	"github.com/livefir/fir/internal/handlers"
	"github.com/livefir/fir/internal/routeservices"
	"github.com/livefir/fir/internal/services"
)

// RouteServiceFactory creates and configures the services needed by routes
type RouteServiceFactory struct {
	services *routeservices.RouteServices
}

// NewRouteServiceFactory creates a new route service factory
func NewRouteServiceFactory(services *routeservices.RouteServices) *RouteServiceFactory {
	return &RouteServiceFactory{
		services: services,
	}
}

// CreateEventService returns the configured event service
func (f *RouteServiceFactory) CreateEventService() services.EventService {
	return f.services.EventService
}

// CreateRenderService returns the configured render service
func (f *RouteServiceFactory) CreateRenderService() services.RenderService {
	return f.services.RenderService
}

// CreateTemplateService returns the configured template service
func (f *RouteServiceFactory) CreateTemplateService() services.TemplateService {
	return f.services.TemplateService
}

// CreateResponseBuilder returns the configured response builder
func (f *RouteServiceFactory) CreateResponseBuilder() services.ResponseBuilder {
	return f.services.ResponseBuilder
}

// CreateHandlerChain creates and configures a handler chain for request processing
func (f *RouteServiceFactory) CreateHandlerChain() handlers.HandlerChain {
	// If the services already has a handler chain, return it
	if f.services.HandlerChain != nil {
		if chain, ok := f.services.HandlerChain.(handlers.HandlerChain); ok {
			return chain
		}
	}

	// Otherwise, create a new default handler chain
	return handlers.SetupDefaultHandlerChain(f.services)
}

// GetServices returns the underlying RouteServices
func (f *RouteServiceFactory) GetServices() *routeservices.RouteServices {
	return f.services
}
