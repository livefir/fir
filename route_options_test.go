package fir

import (
	"html/template"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouteOptions_Layout(t *testing.T) {
	// Test Layout option
	layoutOpt := Layout("main.html")

	opts := RouteOptions{}
	opts = append(opts, layoutOpt)

	// Create a routeOpt to test the option
	routeOpt := &routeOpt{}
	for _, opt := range opts {
		opt(routeOpt)
	}

	assert.Equal(t, "main.html", routeOpt.layout)
}

func TestRouteOptions_LayoutContentName(t *testing.T) {
	// Test LayoutContentName option
	contentNameOpt := LayoutContentName("main-content")

	opts := RouteOptions{}
	opts = append(opts, contentNameOpt)

	routeOpt := &routeOpt{}
	for _, opt := range opts {
		opt(routeOpt)
	}

	assert.Equal(t, "main-content", routeOpt.layoutContentName)
}

func TestRouteOptions_ErrorLayout(t *testing.T) {
	// Test ErrorLayout option
	errorLayoutOpt := ErrorLayout("error.html")

	opts := RouteOptions{}
	opts = append(opts, errorLayoutOpt)

	routeOpt := &routeOpt{}
	for _, opt := range opts {
		opt(routeOpt)
	}

	assert.Equal(t, "error.html", routeOpt.errorLayout)
}

func TestRouteOptions_ErrorContent(t *testing.T) {
	// Test ErrorContent option
	errorContentOpt := ErrorContent("<div>Error occurred</div>")

	opts := RouteOptions{}
	opts = append(opts, errorContentOpt)

	routeOpt := &routeOpt{}
	for _, opt := range opts {
		opt(routeOpt)
	}

	assert.Equal(t, "<div>Error occurred</div>", routeOpt.errorContent)
}

func TestRouteOptions_ErrorLayoutContentName(t *testing.T) {
	// Test ErrorLayoutContentName option
	errorContentNameOpt := ErrorLayoutContentName("error-content")

	opts := RouteOptions{}
	opts = append(opts, errorContentNameOpt)

	routeOpt := &routeOpt{}
	for _, opt := range opts {
		opt(routeOpt)
	}

	assert.Equal(t, "error-content", routeOpt.errorLayoutContentName)
}

func TestRouteOptions_Partials(t *testing.T) {
	// Test Partials option
	partials := []string{"header.html", "footer.html", "sidebar.html"}
	partialsOpt := Partials(partials...)

	opts := RouteOptions{}
	opts = append(opts, partialsOpt)

	routeOpt := &routeOpt{}
	for _, opt := range opts {
		opt(routeOpt)
	}

	assert.Equal(t, partials, routeOpt.partials)
}

func TestRouteOptions_Extensions(t *testing.T) {
	// Test Extensions option
	extensions := []string{".html", ".gohtml", ".tmpl"}
	extensionsOpt := Extensions(extensions...)

	opts := RouteOptions{}
	opts = append(opts, extensionsOpt)

	routeOpt := &routeOpt{}
	for _, opt := range opts {
		opt(routeOpt)
	}

	assert.Equal(t, extensions, routeOpt.extensions)
}

func TestRouteOptions_FuncMap(t *testing.T) {
	// Test FuncMap option
	customFuncs := template.FuncMap{
		"upper": func(s string) string {
			return "UPPER_" + s
		},
		"double": func(i int) int {
			return i * 2
		},
	}
	funcMapOpt := FuncMap(customFuncs)

	opts := RouteOptions{}
	opts = append(opts, funcMapOpt)

	// Initialize funcMap and mutex
	routeOpt := &routeOpt{
		funcMap:      make(template.FuncMap),
		funcMapMutex: &sync.RWMutex{},
	}
	for _, opt := range opts {
		opt(routeOpt)
	}

	assert.NotNil(t, routeOpt.funcMap)
	// Check if custom functions were merged into funcMap
	assert.Contains(t, routeOpt.funcMap, "upper")
	assert.Contains(t, routeOpt.funcMap, "double")
}

func TestRouteOptions_EventSender(t *testing.T) {
	// Test EventSender option
	eventChan := make(chan Event, 1)
	eventSenderOpt := EventSender(eventChan)

	opts := RouteOptions{}
	opts = append(opts, eventSenderOpt)

	routeOpt := &routeOpt{}
	for _, opt := range opts {
		opt(routeOpt)
	}

	assert.Equal(t, eventChan, routeOpt.eventSender)

	// Clean up
	close(eventChan)
}

func TestRouteOptions_MultipleOptions(t *testing.T) {
	// Test combining multiple options
	customFuncs := template.FuncMap{
		"test": func() string { return "test" },
	}

	opts := RouteOptions{
		Layout("layout.html"),
		LayoutContentName("content"),
		ErrorLayout("error.html"),
		ErrorContent("<div>Error</div>"),
		ErrorLayoutContentName("error-content"),
		Partials("header.html", "footer.html"),
		Extensions(".html", ".gohtml"),
		FuncMap(customFuncs),
	}

	// Initialize routeOpt properly
	routeOpt := &routeOpt{
		funcMap:      make(template.FuncMap),
		funcMapMutex: &sync.RWMutex{},
	}
	for _, opt := range opts {
		opt(routeOpt)
	}

	// Verify all options were applied
	assert.Equal(t, "layout.html", routeOpt.layout)
	assert.Equal(t, "content", routeOpt.layoutContentName)
	assert.Equal(t, "error.html", routeOpt.errorLayout)
	assert.Equal(t, "<div>Error</div>", routeOpt.errorContent)
	assert.Equal(t, "error-content", routeOpt.errorLayoutContentName)
	assert.Equal(t, []string{"header.html", "footer.html"}, routeOpt.partials)
	assert.Equal(t, []string{".html", ".gohtml"}, routeOpt.extensions)
	assert.Contains(t, routeOpt.funcMap, "test")
}

func TestRouteOptions_EmptyOptions(t *testing.T) {
	// Test with no options
	opts := RouteOptions{}

	routeOpt := &routeOpt{}
	for _, opt := range opts {
		opt(routeOpt)
	}

	// Verify defaults/empty values
	assert.Equal(t, "", routeOpt.layout)
	assert.Equal(t, "", routeOpt.layoutContentName)
	assert.Equal(t, "", routeOpt.errorLayout)
	assert.Equal(t, "", routeOpt.errorContent)
	assert.Equal(t, "", routeOpt.errorLayoutContentName)
	assert.Nil(t, routeOpt.partials)
	assert.Nil(t, routeOpt.extensions)
	assert.Nil(t, routeOpt.funcMap)
}
