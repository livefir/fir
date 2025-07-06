package services

import (
	"time"

	"github.com/livefir/fir/internal/event"
	"github.com/livefir/fir/pubsub"
)

// ServiceFactory creates and configures all the rendering and template services
type ServiceFactory struct {
	cacheEnabled bool
	cacheConfig  CacheConfig
}

// CacheConfig contains configuration for template caching
type CacheConfig struct {
	DefaultExpiration time.Duration
	CleanupInterval   time.Duration
}

// NewServiceFactory creates a new service factory with default configuration
func NewServiceFactory(cacheEnabled bool) *ServiceFactory {
	return &ServiceFactory{
		cacheEnabled: cacheEnabled,
		cacheConfig: CacheConfig{
			DefaultExpiration: 5 * time.Minute,
			CleanupInterval:   10 * time.Minute,
		},
	}
}

// NewServiceFactoryWithCacheConfig creates a new service factory with custom cache configuration
func NewServiceFactoryWithCacheConfig(cacheEnabled bool, cacheConfig CacheConfig) *ServiceFactory {
	return &ServiceFactory{
		cacheEnabled: cacheEnabled,
		cacheConfig:  cacheConfig,
	}
}

// CreateRenderServices creates all the rendering-related services
func (f *ServiceFactory) CreateRenderServices() (RenderService, TemplateService, ResponseBuilder, TemplateEngine, TemplateCache) {
	// Create template service
	templateService := NewDefaultTemplateService(f.cacheEnabled)

	// Create template engine
	templateEngine := NewGoTemplateEngine()

	// Create template cache
	templateCache := NewInMemoryTemplateCache(f.cacheConfig.DefaultExpiration, f.cacheConfig.CleanupInterval)

	// Create response builder
	responseBuilder := NewDefaultResponseBuilder()

	// Create render service
	renderService := NewDefaultRenderService(templateService, templateEngine, responseBuilder)

	return renderService, templateService, responseBuilder, templateEngine, templateCache
}

// CreateEventServices creates the main event service using existing infrastructure
func (f *ServiceFactory) CreateEventServices(eventRegistry event.EventRegistry, pubsubAdapter pubsub.Adapter) EventValidator {
	// Create event validator (the main component we need for route integration)
	validator := NewDefaultEventValidator()
	return validator
}

// CreateIntegratedServices creates all services and the adapter for legacy integration
func (f *ServiceFactory) CreateIntegratedServices(eventRegistry event.EventRegistry, pubsubAdapter pubsub.Adapter, legacyRenderer interface{}) (*LegacyRenderAdapter, RenderService, TemplateService, ResponseBuilder, EventValidator) {
	// Create render services
	renderService, templateService, responseBuilder, _, _ := f.CreateRenderServices()

	// Create event services
	validator := f.CreateEventServices(eventRegistry, pubsubAdapter)

	// Create legacy adapter
	adapter := NewLegacyRenderAdapter(renderService, templateService, responseBuilder, legacyRenderer)

	return adapter, renderService, templateService, responseBuilder, validator
}

// SetCacheEnabled enables or disables caching for future service creation
func (f *ServiceFactory) SetCacheEnabled(enabled bool) {
	f.cacheEnabled = enabled
}

// SetCacheConfig updates the cache configuration for future service creation
func (f *ServiceFactory) SetCacheConfig(config CacheConfig) {
	f.cacheConfig = config
}

// GetCacheConfig returns the current cache configuration
func (f *ServiceFactory) GetCacheConfig() CacheConfig {
	return f.cacheConfig
}

// IsCacheEnabled returns whether caching is enabled
func (f *ServiceFactory) IsCacheEnabled() bool {
	return f.cacheEnabled
}

// ValidateFactory validates that the factory is properly configured
func (f *ServiceFactory) ValidateFactory() error {
	// Currently no validation needed, but could be extended in the future
	return nil
}
