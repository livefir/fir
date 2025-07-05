package templateengine

import (
	"html/template"
)

// RouteFuncMapBuilder creates function maps specific to route contexts.
// This builder understands how to extract information from Fir's RouteContext
// and provides route-specific template functions.
type RouteFuncMapBuilder struct {
	baseProvider FuncMapProvider
}

// NewRouteFuncMapBuilder creates a new route-specific function map builder.
func NewRouteFuncMapBuilder() *RouteFuncMapBuilder {
	return &RouteFuncMapBuilder{
		baseProvider: NewDefaultFuncMapProvider(),
	}
}

// NewRouteFuncMapBuilderWithProvider creates a new route-specific function map builder
// with a custom base provider.
func NewRouteFuncMapBuilderWithProvider(provider FuncMapProvider) *RouteFuncMapBuilder {
	return &RouteFuncMapBuilder{
		baseProvider: provider,
	}
}

// BuildFuncMap implements FuncMapProvider interface.
// This method knows how to extract route context information and convert it
// to the template engine's FuncMapContext format.
func (rfmb *RouteFuncMapBuilder) BuildFuncMap(ctx FuncMapContext) template.FuncMap {
	// Start with base provider functions
	funcMap := rfmb.baseProvider.BuildFuncMap(ctx)

	// Add route-specific functions here
	// For example, route navigation helpers, security functions, etc.
	funcMap["routeInfo"] = func() map[string]interface{} {
		return map[string]interface{}{
			"path":        ctx.URLPath,
			"app":         ctx.AppName,
			"development": ctx.DevelopmentMode,
		}
	}

	return funcMap
}

// GetName implements FuncMapProvider interface.
func (rfmb *RouteFuncMapBuilder) GetName() string {
	return "RouteFuncMapBuilder"
}

// SetBaseProvider sets the base function map provider.
func (rfmb *RouteFuncMapBuilder) SetBaseProvider(provider FuncMapProvider) {
	rfmb.baseProvider = provider
}

// TemplateEngineAdapter provides integration helpers for using the template engine
// with existing Fir framework components.
type TemplateEngineAdapter struct {
	engine   TemplateEngine
	provider FuncMapProvider
}

// NewTemplateEngineAdapter creates a new adapter for integrating with Fir framework.
func NewTemplateEngineAdapter(engine TemplateEngine) *TemplateEngineAdapter {
	return &TemplateEngineAdapter{
		engine:   engine,
		provider: NewRouteFuncMapBuilder(),
	}
}

// NewTemplateEngineAdapterWithProvider creates a new adapter with a custom function map provider.
func NewTemplateEngineAdapterWithProvider(engine TemplateEngine, provider FuncMapProvider) *TemplateEngineAdapter {
	return &TemplateEngineAdapter{
		engine:   engine,
		provider: provider,
	}
}

// ConvertRouteContextToTemplateContext converts a Fir RouteContext to TemplateContext.
// This is a helper function for bridging the gap between the existing Fir framework
// and the new template engine.
func (tea *TemplateEngineAdapter) ConvertRouteContextToTemplateContext(
	routeContext interface{},
	errors map[string]interface{},
	urlPath string,
	appName string,
	developmentMode bool,
) TemplateContext {
	return TemplateContext{
		RouteContext: routeContext,
		Errors:       errors,
		Data: map[string]interface{}{
			"URLPath":         urlPath,
			"AppName":         appName,
			"DevelopmentMode": developmentMode,
		},
	}
}

// LoadTemplateForRoute loads a template with route-specific context.
// This is a convenience method that combines template loading with route context conversion.
func (tea *TemplateEngineAdapter) LoadTemplateForRoute(
	config TemplateConfig,
	routeContext interface{},
	errors map[string]interface{},
	urlPath string,
	appName string,
	developmentMode bool,
) (Template, error) {
	ctx := tea.ConvertRouteContextToTemplateContext(
		routeContext, errors, urlPath, appName, developmentMode,
	)

	// If this is a GoTemplateEngine, set our function map provider
	if goEngine, ok := tea.engine.(*GoTemplateEngine); ok {
		goEngine.SetFuncMapProvider(tea.provider)
		return goEngine.LoadTemplateWithContext(config, ctx)
	}

	// Use context-aware loading if the engine supports it
	if contextEngine, ok := tea.engine.(interface {
		LoadTemplateWithContext(TemplateConfig, TemplateContext) (Template, error)
	}); ok {
		return contextEngine.LoadTemplateWithContext(config, ctx)
	}

	// Fall back to regular loading
	return tea.engine.LoadTemplate(config)
}

// LoadErrorTemplateForRoute loads an error template with route-specific context.
func (tea *TemplateEngineAdapter) LoadErrorTemplateForRoute(
	config TemplateConfig,
	routeContext interface{},
	errors map[string]interface{},
	urlPath string,
	appName string,
	developmentMode bool,
) (Template, error) {
	ctx := tea.ConvertRouteContextToTemplateContext(
		routeContext, errors, urlPath, appName, developmentMode,
	)

	// If this is a GoTemplateEngine, set our function map provider
	if goEngine, ok := tea.engine.(*GoTemplateEngine); ok {
		goEngine.SetFuncMapProvider(tea.provider)
		return goEngine.LoadErrorTemplateWithContext(config, ctx)
	}

	// Use context-aware loading if the engine supports it
	if contextEngine, ok := tea.engine.(interface {
		LoadErrorTemplateWithContext(TemplateConfig, TemplateContext) (Template, error)
	}); ok {
		return contextEngine.LoadErrorTemplateWithContext(config, ctx)
	}

	// Fall back to regular loading
	return tea.engine.LoadErrorTemplate(config)
}

// GetEngine returns the underlying template engine.
func (tea *TemplateEngineAdapter) GetEngine() TemplateEngine {
	return tea.engine
}

// SetFuncMapProvider sets the function map provider for the adapter.
func (tea *TemplateEngineAdapter) SetFuncMapProvider(provider FuncMapProvider) {
	tea.provider = provider
}

// GetFuncMapProvider returns the current function map provider.
func (tea *TemplateEngineAdapter) GetFuncMapProvider() FuncMapProvider {
	return tea.provider
}
