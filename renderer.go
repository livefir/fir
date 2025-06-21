package fir

import (
	"github.com/livefir/fir/internal/dom"
	"github.com/livefir/fir/pubsub"
)

// Renderer defines the interface for rendering routes and DOM events
type Renderer interface {
	// RenderRoute renders a route with the given data
	RenderRoute(ctx RouteContext, data routeData, useErrorTemplate bool) error

	// RenderDOMEvents renders DOM events from a pubsub event
	RenderDOMEvents(ctx RouteContext, pubsubEvent pubsub.Event) []dom.Event
}

// TemplateRenderer is the default implementation of Renderer that uses Go's html/template
type TemplateRenderer struct{}

// NewTemplateRenderer creates a new TemplateRenderer
func NewTemplateRenderer() *TemplateRenderer {
	return &TemplateRenderer{}
}

// RenderRoute implements the Renderer interface for template-based rendering
func (tr *TemplateRenderer) RenderRoute(ctx RouteContext, data routeData, useErrorTemplate bool) error {
	return renderRoute(ctx, useErrorTemplate)(data)
}

// RenderDOMEvents implements the Renderer interface for DOM event rendering
func (tr *TemplateRenderer) RenderDOMEvents(ctx RouteContext, pubsubEvent pubsub.Event) []dom.Event {
	return renderDOMEvents(ctx, pubsubEvent)
}
