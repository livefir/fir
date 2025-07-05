package templateengine

import (
	"html/template"
)

// RouteTemplateEngineBuilder creates template engines specifically configured for routes
type RouteTemplateEngineBuilder struct {
	config           TemplateConfig
	funcMapProviders []FuncMapProvider
	eventEngine      EventTemplateEngine
}

// NewRouteTemplateEngineBuilder creates a new builder for route-specific template engines
func NewRouteTemplateEngineBuilder() *RouteTemplateEngineBuilder {
	return &RouteTemplateEngineBuilder{
		config:           TemplateConfig{DisableTemplateCache: false},
		funcMapProviders: make([]FuncMapProvider, 0),
	}
}

// WithConfig sets the template configuration
func (b *RouteTemplateEngineBuilder) WithConfig(config TemplateConfig) *RouteTemplateEngineBuilder {
	b.config = config
	return b
}

// WithFuncMapProvider adds a function map provider
func (b *RouteTemplateEngineBuilder) WithFuncMapProvider(provider FuncMapProvider) *RouteTemplateEngineBuilder {
	b.funcMapProviders = append(b.funcMapProviders, provider)
	return b
}

// WithBaseFuncMap adds base template functions
func (b *RouteTemplateEngineBuilder) WithBaseFuncMap(funcMap template.FuncMap) *RouteTemplateEngineBuilder {
	if len(funcMap) > 0 {
		provider := &StaticFuncMapProvider{funcMap: funcMap}
		b.funcMapProviders = append(b.funcMapProviders, provider)
	}
	return b
}

// WithEventEngine sets the event template engine
func (b *RouteTemplateEngineBuilder) WithEventEngine(engine EventTemplateEngine) *RouteTemplateEngineBuilder {
	b.eventEngine = engine
	return b
}

// WithRouteConfig configures the builder with route-specific settings
func (b *RouteTemplateEngineBuilder) WithRouteConfig(routeID, content, layout, layoutContentName string, disableCache bool) *RouteTemplateEngineBuilder {
	// Map route parameters to TemplateConfig fields
	if content != "" {
		b.config.ContentPath = content
		b.config.ContentTemplate = content
	}
	if layout != "" {
		b.config.LayoutPath = layout
		b.config.LayoutContent = layout
	}
	if layoutContentName != "" {
		b.config.LayoutContentName = layoutContentName
	}
	b.config.DisableTemplateCache = disableCache
	return b
}

// Build creates a new template engine instance
func (b *RouteTemplateEngineBuilder) Build() (TemplateEngine, error) {
	// Create event engine if not provided
	eventEngine := b.eventEngine
	if eventEngine == nil {
		registry := NewInMemoryEventTemplateRegistry()
		extractor := NewHTMLEventTemplateExtractor()
		eventEngine = NewDefaultEventTemplateEngine()
		// Set the registry and extractor after creation
		if defaultEngine, ok := eventEngine.(*DefaultEventTemplateEngine); ok {
			defaultEngine.registry = registry
			defaultEngine.extractor = extractor
		}
	}

	// Create function map provider that combines all providers
	var combinedProvider FuncMapProvider
	if len(b.funcMapProviders) == 0 {
		// Use a default empty provider
		combinedProvider = &StaticFuncMapProvider{funcMap: template.FuncMap{}}
	} else if len(b.funcMapProviders) == 1 {
		combinedProvider = b.funcMapProviders[0]
	} else {
		// Combine multiple providers
		combinedProvider = NewCompositeFuncMapProvider("route-builder", b.funcMapProviders...)
	}

	// Create the template engine
	engine := NewGoTemplateEngine()
	engine.cache = NewInMemoryTemplateCache()
	engine.enableCache = !b.config.DisableTemplateCache
	engine.funcMapProvider = combinedProvider
	engine.eventEngine = eventEngine

	return engine, nil
}

// StaticFuncMapProvider provides a static function map
type StaticFuncMapProvider struct {
	funcMap template.FuncMap
}

// BuildFuncMap implements FuncMapProvider interface
func (s *StaticFuncMapProvider) BuildFuncMap(ctx FuncMapContext) template.FuncMap {
	return s.funcMap
}

// GetName implements FuncMapProvider interface
func (s *StaticFuncMapProvider) GetName() string {
	return "static-funcmap"
}

// RouteTemplateEngineFactory creates template engines for routes with sensible defaults
type RouteTemplateEngineFactory struct {
	defaultFuncMap template.FuncMap
	defaultConfig  TemplateConfig
}

// NewRouteTemplateEngineFactory creates a new factory with default settings
func NewRouteTemplateEngineFactory(defaultFuncMap template.FuncMap) *RouteTemplateEngineFactory {
	return &RouteTemplateEngineFactory{
		defaultFuncMap: defaultFuncMap,
		defaultConfig:  TemplateConfig{DisableTemplateCache: false},
	}
}

// CreateEngine creates a template engine for a route
func (f *RouteTemplateEngineFactory) CreateEngine(routeID, content, layout, layoutContentName string, disableCache bool) (TemplateEngine, error) {
	builder := NewRouteTemplateEngineBuilder().
		WithBaseFuncMap(f.defaultFuncMap).
		WithRouteConfig(routeID, content, layout, layoutContentName, disableCache)

	return builder.Build()
}

// CreateEngineWithConfig creates a template engine with custom configuration
func (f *RouteTemplateEngineFactory) CreateEngineWithConfig(config TemplateConfig, funcMapProviders []FuncMapProvider) (TemplateEngine, error) {
	builder := NewRouteTemplateEngineBuilder().
		WithConfig(config).
		WithBaseFuncMap(f.defaultFuncMap)

	for _, provider := range funcMapProviders {
		builder.WithFuncMapProvider(provider)
	}

	return builder.Build()
}
