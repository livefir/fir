package main

import (
	"log"
	"net/http"

	"github.com/livefir/fir"
	"github.com/livefir/fir/internal/templateengine"
)

// Example showing how to use the new template engine with routes
func templateEngineRoute() fir.RouteOptions {
	// Create a custom template engine
	engine := templateengine.NewGoTemplateEngine()

	return fir.RouteOptions{
		fir.ID("template-engine-example"),
		fir.Content(`<h1>Template Engine Example</h1><p>Count: {{.count}}</p><button fir-click="increment">+</button>`),
		fir.TemplateEngine(engine), // Use the template engine
		fir.OnLoad(func(ctx fir.RouteContext) error {
			return ctx.KV("count", 0)
		}),
		fir.OnEvent("increment", func(ctx fir.RouteContext) error {
			// Simple counter increment for demonstration
			return ctx.KV("count", 1)
		}),
	}
}

func main() {
	controller := fir.NewController("template_engine_app", fir.DevelopmentMode(true))
	http.Handle("/", controller.RouteFunc(templateEngineRoute))

	log.Println("Starting template engine example server on :3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
