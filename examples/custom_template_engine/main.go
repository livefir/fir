package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/livefir/fir"
	"github.com/livefir/fir/internal/templateengine"
)

// CustomTemplateEngine demonstrates how to implement a custom template engine
// This example creates a simple template engine that adds custom processing
type CustomTemplateEngine struct {
	baseEngine templateengine.TemplateEngine
	cache      map[string]templateengine.Template
	cacheMutex sync.RWMutex
	prefix     string // Custom prefix for all templates
}

// NewCustomTemplateEngine creates a new custom template engine
func NewCustomTemplateEngine(prefix string) *CustomTemplateEngine {
	return &CustomTemplateEngine{
		baseEngine: templateengine.NewGoTemplateEngine(),
		cache:      make(map[string]templateengine.Template),
		prefix:     prefix,
	}
}

// LoadTemplate implements the TemplateEngine interface with custom processing
func (cte *CustomTemplateEngine) LoadTemplate(config templateengine.TemplateConfig) (templateengine.Template, error) {
	// Add custom prefix to all template content
	if config.ContentTemplate != "" {
		config.ContentTemplate = cte.addCustomPrefix(config.ContentTemplate)
	}

	// Use base engine to load the template
	return cte.baseEngine.LoadTemplate(config)
}

// LoadErrorTemplate implements the TemplateEngine interface
func (cte *CustomTemplateEngine) LoadErrorTemplate(config templateengine.TemplateConfig) (templateengine.Template, error) {
	// Add custom error styling
	if config.ErrorContentTemplate != "" {
		config.ErrorContentTemplate = cte.addErrorStyling(config.ErrorContentTemplate)
	}

	return cte.baseEngine.LoadErrorTemplate(config)
}

// LoadTemplateWithContext implements context-aware loading
func (cte *CustomTemplateEngine) LoadTemplateWithContext(config templateengine.TemplateConfig, ctx templateengine.TemplateContext) (templateengine.Template, error) {
	// Add custom functions to context
	if ctx.FuncMap == nil {
		ctx.FuncMap = make(template.FuncMap)
	}
	ctx.FuncMap["customPrefix"] = func() string { return cte.prefix }
	ctx.FuncMap["toUpper"] = strings.ToUpper
	ctx.FuncMap["customFormat"] = func(s string) string {
		return fmt.Sprintf("[%s] %s", cte.prefix, s)
	}

	// Process template content
	if config.ContentTemplate != "" {
		config.ContentTemplate = cte.addCustomPrefix(config.ContentTemplate)
	}

	return cte.baseEngine.LoadTemplateWithContext(config, ctx)
}

// LoadErrorTemplateWithContext implements context-aware error template loading
func (cte *CustomTemplateEngine) LoadErrorTemplateWithContext(config templateengine.TemplateConfig, ctx templateengine.TemplateContext) (templateengine.Template, error) {
	// Add error-specific functions
	if ctx.FuncMap == nil {
		ctx.FuncMap = make(template.FuncMap)
	}
	ctx.FuncMap["errorStyle"] = func(msg string) string {
		return fmt.Sprintf("<div class='error'>%s</div>", msg)
	}

	return cte.baseEngine.LoadErrorTemplateWithContext(config, ctx)
}

// Render implements template rendering with custom processing
func (cte *CustomTemplateEngine) Render(tmpl templateengine.Template, data interface{}, w io.Writer) error {
	return cte.baseEngine.Render(tmpl, data, w)
}

// RenderWithContext implements context-aware rendering
func (cte *CustomTemplateEngine) RenderWithContext(tmpl templateengine.Template, ctx templateengine.TemplateContext, data interface{}, w io.Writer) error {
	return cte.baseEngine.RenderWithContext(tmpl, ctx, data, w)
}

// ExtractEventTemplates implements event template extraction
func (cte *CustomTemplateEngine) ExtractEventTemplates(tmpl templateengine.Template) (templateengine.EventTemplateMap, error) {
	return cte.baseEngine.ExtractEventTemplates(tmpl)
}

// RenderEventTemplate implements event template rendering
func (cte *CustomTemplateEngine) RenderEventTemplate(tmpl templateengine.Template, eventID string, state string, data interface{}) (string, error) {
	return cte.baseEngine.RenderEventTemplate(tmpl, eventID, state, data)
}

// CacheTemplate implements template caching with custom key generation
func (cte *CustomTemplateEngine) CacheTemplate(id string, tmpl templateengine.Template) {
	cte.cacheMutex.Lock()
	defer cte.cacheMutex.Unlock()

	// Add prefix to cache key
	key := cte.prefix + "-" + id
	cte.cache[key] = tmpl

	// Also cache in base engine
	cte.baseEngine.CacheTemplate(key, tmpl)
}

// GetCachedTemplate implements cached template retrieval
func (cte *CustomTemplateEngine) GetCachedTemplate(id string) (templateengine.Template, bool) {
	cte.cacheMutex.RLock()
	defer cte.cacheMutex.RUnlock()

	// Try with prefix
	key := cte.prefix + "-" + id
	if tmpl, found := cte.cache[key]; found {
		return tmpl, true
	}

	// Fallback to base engine
	return cte.baseEngine.GetCachedTemplate(key)
}

// ClearCache implements cache clearing
func (cte *CustomTemplateEngine) ClearCache() {
	cte.cacheMutex.Lock()
	defer cte.cacheMutex.Unlock()

	cte.cache = make(map[string]templateengine.Template)
	cte.baseEngine.ClearCache()
}

// Custom processing methods

func (cte *CustomTemplateEngine) addCustomPrefix(content string) string {
	// Add custom header to all templates
	header := fmt.Sprintf(`<div class="custom-header">%s App</div>`, cte.prefix)
	return header + "\n" + content
}

func (cte *CustomTemplateEngine) addErrorStyling(content string) string {
	// Add error styling wrapper
	return `<div class="error-container">` + content + `</div>`
}

// Example route using the custom template engine
func customEngineRoute() fir.RouteOptions {
	customEngine := NewCustomTemplateEngine("CUSTOM")

	return fir.RouteOptions{
		fir.ID("custom-engine-example"),
		fir.Content(`
			<style>
				.custom-header { background: #007acc; color: white; padding: 10px; margin-bottom: 20px; }
				.counter-section { border: 2px solid #007acc; padding: 20px; border-radius: 8px; }
				.error-container { border: 2px solid #ff4444; background: #ffeeee; padding: 10px; }
			</style>
			<div class="counter-section">
				<h1>{{customFormat "Custom Template Engine Demo"}}</h1>
				<p>This template is processed by a custom template engine!</p>
				<p>Current count: <strong>{{.count}}</strong></p>
				<p>Formatted message: {{customFormat "Hello World"}}</p>
				<p>Uppercase test: {{toUpper "this will be uppercase"}}</p>
				<div>
					<button fir-click="increment">Increment</button>
					<button fir-click="decrement">Decrement</button>
					<button fir-click="reset">Reset</button>
				</div>
				<template @fir:increment:ok>
					<p>Count increased to: {{.count}}</p>
				</template>
				<template @fir:decrement:ok>
					<p>Count decreased to: {{.count}}</p>
				</template>
				<template @fir:reset:ok>
					<p>Count reset to: {{.count}}</p>
				</template>
			</div>
		`),
		fir.TemplateEngine(customEngine),
		fir.OnLoad(func(ctx fir.RouteContext) error {
			return ctx.KV("count", 0)
		}),
		fir.OnEvent("increment", func(ctx fir.RouteContext) error {
			// Simple increment for demonstration
			return ctx.KV("count", 1)
		}),
		fir.OnEvent("decrement", func(ctx fir.RouteContext) error {
			return ctx.KV("count", 0)
		}),
		fir.OnEvent("reset", func(ctx fir.RouteContext) error {
			return ctx.KV("count", 0)
		}),
	}
}

func main() {
	controller := fir.NewController("custom_template_engine_app", fir.DevelopmentMode(true))
	http.Handle("/", controller.RouteFunc(customEngineRoute))

	log.Println("Custom Template Engine Example Server")
	log.Println("Starting server on :3001")
	log.Println("Visit http://localhost:3001 to see the custom template engine in action")
	log.Fatal(http.ListenAndServe(":3001", nil))
}
