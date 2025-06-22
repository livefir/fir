package fir

import (
	"testing"

	"github.com/livefir/fir/internal/dom"
	"github.com/livefir/fir/pubsub"
)

// MockRenderer is a mock implementation of the Renderer interface for testing
type MockRenderer struct {
	renderRouteCalled     bool
	renderDOMEventsCalled bool
	lastRouteData         routeData
	lastUseErrorTemplate  bool
}

func (mr *MockRenderer) RenderRoute(ctx RouteContext, data routeData, useErrorTemplate bool) error {
	mr.renderRouteCalled = true
	mr.lastRouteData = data
	mr.lastUseErrorTemplate = useErrorTemplate
	return nil
}

func (mr *MockRenderer) RenderDOMEvents(ctx RouteContext, pubsubEvent pubsub.Event) []dom.Event {
	mr.renderDOMEventsCalled = true
	return []dom.Event{}
}

func TestRendererInterface(t *testing.T) {
	// Test that the default TemplateRenderer implements the interface
	var renderer Renderer = NewTemplateRenderer()
	_ = renderer // Verify it implements the interface

	// Test that our mock implements the interface
	var mockRenderer Renderer = &MockRenderer{}
	_ = mockRenderer // Verify it implements the interface
}

func TestCustomRenderer(t *testing.T) {
	mockRenderer := &MockRenderer{}

	// Create a controller with a custom renderer
	cntrl := NewController("test", WithRenderer(mockRenderer))

	// Verify the renderer is set in the controller options
	controller := cntrl.(*controller)
	if controller.opt.renderer != mockRenderer {
		t.Fatal("Custom renderer should be set in controller options")
	}
}

func TestDefaultRenderer(t *testing.T) {
	// Create a controller without specifying a renderer
	cntrl := NewController("test")

	// Verify the default renderer is used
	controller := cntrl.(*controller)
	if controller.opt.renderer != nil {
		t.Fatal("Default controller should not have a renderer set in options (it's created in newRoute)")
	}
}
