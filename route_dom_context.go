package fir

import (
	"strings"
	"text/template"

	"github.com/goccy/go-json"

	"github.com/tidwall/gjson"
)

func newFirFuncMap(ctx RouteContext, errs map[string]any) template.FuncMap {
	return template.FuncMap{
		"fir": func() *RouteDOMContext {
			return newRouteDOMContext(ctx, errs)
		},
	}
}

func newRouteDOMContext(ctx RouteContext, errs map[string]any) *RouteDOMContext {
	var urlPath string
	var name string
	var developmentMode bool

	if ctx.request != nil {
		urlPath = ctx.request.URL.Path
	}

	// Use routeInterface for both legacy and WebSocketServices modes
	name = ctx.routeInterface.GetAppName()
	developmentMode = ctx.routeInterface.DevelopmentMode()

	if errs == nil {
		errs = make(map[string]any)
	}
	return &RouteDOMContext{
		URLPath:     urlPath,
		Name:        name,
		Development: developmentMode,
		errors:      errs,
	}
}

// RouteDOMContext is a struct that holds route context data and is passed to the template
type RouteDOMContext struct {
	Name        string
	Development bool
	URLPath     string
	errors      map[string]any
}

// ActiveRoute returns the class if the route is active
func (rc *RouteDOMContext) ActiveRoute(path, class string) string {
	if rc.URLPath == path {
		return class
	}
	return ""
}

// NotActive returns the class if the route is not active
func (rc *RouteDOMContext) NotActiveRoute(path, class string) string {
	if rc.URLPath != path {
		return class
	}
	return ""
}

// Error can be used to lookup an error by name
// Example: {{fir.Error "myevent.field"}} will return the error for the field myevent.field
// Example: {{fir.Error "myevent" "field"}} will return the error for the event myevent.field
// It can be used in conjunction with ctx.FieldError to get the error for a field
func (rc *RouteDOMContext) Error(paths ...string) any {
	if len(rc.errors) == 0 {
		return nil
	}
	data, _ := json.Marshal(rc.errors)
	val := gjson.GetBytes(data, getErrorLookupPath(paths...)).Value()
	_, ok := val.(map[string]any)
	if ok {
		return nil
	}
	return val
}
func getErrorLookupPath(paths ...string) string {
	path := ""
	if len(paths) == 0 {
		path = "default"
	} else {
		for _, p := range paths {
			p = strings.Trim(p, ".")
			path += p + "."
		}
	}
	path = strings.Trim(path, ".")
	return path
}
