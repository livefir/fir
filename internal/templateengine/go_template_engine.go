package templateengine

import (
	"fmt"
	"html/template"
	"io"
)

// GoTemplateEngine implements TemplateEngine using Go's html/template package.
type GoTemplateEngine struct {
	cache           TemplateCache
	enableCache     bool
	funcMapProvider FuncMapProvider
	eventEngine     EventTemplateEngine
}

// NewGoTemplateEngine creates a new GoTemplateEngine.
func NewGoTemplateEngine() *GoTemplateEngine {
	return &GoTemplateEngine{
		cache:           NewInMemoryTemplateCache(),
		enableCache:     true,
		funcMapProvider: NewDefaultFuncMapProvider(),
		eventEngine:     NewDefaultEventTemplateEngine(),
	}
}

// NewGoTemplateEngineWithProvider creates a new GoTemplateEngine with a custom function map provider.
func NewGoTemplateEngineWithProvider(provider FuncMapProvider) *GoTemplateEngine {
	return &GoTemplateEngine{
		cache:           NewInMemoryTemplateCache(),
		enableCache:     true,
		funcMapProvider: provider,
		eventEngine:     NewDefaultEventTemplateEngine(),
	}
}

// LoadTemplate implements TemplateEngine interface.
func (gte *GoTemplateEngine) LoadTemplate(config TemplateConfig) (Template, error) {
	// Use empty context for backward compatibility
	return gte.LoadTemplateWithContext(config, TemplateContext{})
}

// LoadErrorTemplate implements TemplateEngine interface.
func (gte *GoTemplateEngine) LoadErrorTemplate(config TemplateConfig) (Template, error) {
	// Use empty context for backward compatibility
	return gte.LoadErrorTemplateWithContext(config, TemplateContext{})
}

// LoadTemplateWithContext implements TemplateEngine interface.
func (gte *GoTemplateEngine) LoadTemplateWithContext(config TemplateConfig, ctx TemplateContext) (Template, error) {
	// Build function map from provider and config
	funcMapCtx := gte.buildFuncMapContext(ctx)
	funcMap := gte.buildFuncMap(config, funcMapCtx)

	// For now, create a simple template based on the configuration
	if config.ContentTemplate != "" {
		// Inline content template
		return gte.createTemplate(config.ContentTemplate, config.LayoutContentName, funcMap)
	} else if config.LayoutContent != "" && config.ContentTemplate != "" {
		// Both layout and content inline
		return gte.createLayoutAndContentTemplate(config.LayoutContent, config.ContentTemplate, config.LayoutContentName, funcMap)
	} else {
		// Default template
		return gte.createDefaultTemplate(false)
	}
}

// LoadErrorTemplateWithContext implements TemplateEngine interface.
func (gte *GoTemplateEngine) LoadErrorTemplateWithContext(config TemplateConfig, ctx TemplateContext) (Template, error) {
	// Build function map from provider and config
	funcMapCtx := gte.buildFuncMapContext(ctx)
	funcMap := gte.buildFuncMap(config, funcMapCtx)

	// Check error template configuration first
	if config.ErrorContentTemplate != "" {
		return gte.createTemplate(config.ErrorContentTemplate, config.ErrorLayoutContentName, funcMap)
	} else if config.ErrorLayoutContent != "" && config.ErrorContentTemplate != "" {
		return gte.createLayoutAndContentTemplate(config.ErrorLayoutContent, config.ErrorContentTemplate, config.ErrorLayoutContentName, funcMap)
	} else {
		// Fall back to regular template configuration
		return gte.LoadTemplateWithContext(config, ctx)
	}
}

// Render implements TemplateEngine interface.
func (gte *GoTemplateEngine) Render(template Template, data interface{}, w io.Writer) error {
	if template == nil {
		return ErrInvalidTemplate
	}
	return template.Execute(w, data)
}

// RenderWithContext implements TemplateEngine interface.
func (gte *GoTemplateEngine) RenderWithContext(template Template, ctx TemplateContext, data interface{}, w io.Writer) error {
	if template == nil {
		return ErrInvalidTemplate
	}

	// Clone template and apply context function map
	cloned, err := template.Clone()
	if err != nil {
		return err
	}

	if ctx.FuncMap != nil {
		cloned = cloned.Funcs(ctx.FuncMap)
	}

	return cloned.Execute(w, data)
}

// ExtractEventTemplates implements TemplateEngine interface.
func (gte *GoTemplateEngine) ExtractEventTemplates(template Template) (EventTemplateMap, error) {
	return gte.eventEngine.ExtractEventTemplates(template)
}

// RenderEventTemplate implements TemplateEngine interface.
func (gte *GoTemplateEngine) RenderEventTemplate(template Template, eventID string, state string, data interface{}) (string, error) {
	return gte.eventEngine.RenderEventTemplate(template, eventID, state, data)
}

// CacheTemplate implements TemplateEngine interface.
func (gte *GoTemplateEngine) CacheTemplate(id string, template Template) {
	if gte.cache != nil {
		gte.cache.Set(id, template)
	}
}

// GetCachedTemplate implements TemplateEngine interface.
func (gte *GoTemplateEngine) GetCachedTemplate(id string) (Template, bool) {
	if gte.cache == nil {
		return nil, false
	}
	return gte.cache.Get(id)
}

// ClearCache implements TemplateEngine interface.
func (gte *GoTemplateEngine) ClearCache() {
	if gte.cache != nil {
		gte.cache.Clear()
	}
}

// SetFuncMapProvider sets the function map provider for this engine.
func (gte *GoTemplateEngine) SetFuncMapProvider(provider FuncMapProvider) {
	gte.funcMapProvider = provider
}

// GetFuncMapProvider returns the current function map provider.
func (gte *GoTemplateEngine) GetFuncMapProvider() FuncMapProvider {
	return gte.funcMapProvider
}

// SetEventEngine sets the event template engine for this template engine.
func (gte *GoTemplateEngine) SetEventEngine(engine EventTemplateEngine) {
	gte.eventEngine = engine
}

// GetEventEngine returns the current event template engine.
func (gte *GoTemplateEngine) GetEventEngine() EventTemplateEngine {
	return gte.eventEngine
}

// GetEventTemplateRegistry returns the event template registry from the event engine.
func (gte *GoTemplateEngine) GetEventTemplateRegistry() EventTemplateRegistry {
	if gte.eventEngine != nil {
		return gte.eventEngine.GetEventTemplateRegistry()
	}
	return NewInMemoryEventTemplateRegistry()
}

// Helper methods

func (gte *GoTemplateEngine) createDefaultTemplate(isError bool) (Template, error) {
	content := `<div style="text-align:center">This is a default page.</div>`
	if isError {
		content = `<div style="text-align:center">This is a default error page.</div>`
	}

	tmpl, err := template.New("default").Parse(content)
	if err != nil {
		return nil, err
	}

	return NewGoTemplate(tmpl), nil
}

func (gte *GoTemplateEngine) createTemplate(content, name string, funcMap template.FuncMap) (Template, error) {
	if name == "" {
		name = "content"
	}

	// Create template with a proper name and functions
	tmpl := template.New(name).Funcs(funcMap)

	_, err := tmpl.Parse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return NewGoTemplate(tmpl), nil
}

func (gte *GoTemplateEngine) createLayoutAndContentTemplate(layout, content, contentName string, funcMap template.FuncMap) (Template, error) {
	if contentName == "" {
		contentName = "content"
	}

	// Create base template with layout
	tmpl, err := template.New("layout").Funcs(funcMap).Parse(layout)
	if err != nil {
		return nil, fmt.Errorf("failed to parse layout: %w", err)
	}

	// Add content template
	_, err = tmpl.New(contentName).Parse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse content: %w", err)
	}

	return NewGoTemplate(tmpl), nil
}

// buildFuncMap creates a merged function map from the config and provider.
// Config function maps take precedence over provider function maps.
func (gte *GoTemplateEngine) buildFuncMap(config TemplateConfig, ctx FuncMapContext) template.FuncMap {
	result := make(template.FuncMap)

	// Start with provider functions if available
	if gte.funcMapProvider != nil {
		providerFuncs := gte.funcMapProvider.BuildFuncMap(ctx)
		for name, fn := range providerFuncs {
			result[name] = fn
		}
	}

	// Override with config-specific functions
	if config.FuncMap != nil {
		for name, fn := range config.FuncMap {
			result[name] = fn
		}
	}

	return result
}

// buildFuncMapContext converts TemplateContext to FuncMapContext for the provider.
func (gte *GoTemplateEngine) buildFuncMapContext(ctx TemplateContext) FuncMapContext {
	// Extract context information - handle the case where RouteContext might be nil
	var urlPath, appName string
	var developmentMode bool

	// Try to extract information from the route context interface
	if ctx.RouteContext != nil {
		// This would need to be adapted based on the actual RouteContext structure
		// For now, use reasonable defaults
		urlPath = ""
		appName = ""
		developmentMode = false
	}

	// Check for URL path and app name in additional data
	if ctx.Data != nil {
		if path, ok := ctx.Data["URLPath"].(string); ok {
			urlPath = path
		}
		if name, ok := ctx.Data["AppName"].(string); ok {
			appName = name
		}
		if dev, ok := ctx.Data["DevelopmentMode"].(bool); ok {
			developmentMode = dev
		}
	}

	return FuncMapContext{
		RouteContext:    ctx.RouteContext,
		Errors:          ctx.Errors,
		URLPath:         urlPath,
		AppName:         appName,
		DevelopmentMode: developmentMode,
		CustomData:      ctx.Data,
	}
}
