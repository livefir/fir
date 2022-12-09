package fir

import (
	"encoding/json"
	"strings"

	"github.com/tidwall/gjson"
)

func newRouteContext(ctx Context, errs map[string]any) *RouteContext {
	return &RouteContext{
		URLPath: ctx.request.URL.Path,
		Name:    ctx.route.appName,
		errors:  errs,
	}
}

// RouteContext is a struct that holds route context data and is passed to the template
type RouteContext struct {
	Name    string
	URLPath string
	errors  map[string]any
}

// ActiveRoute returns the class if the route is active
func (rc *RouteContext) ActiveRoute(path, class string) string {
	if rc.URLPath == path {
		return class
	}
	return ""
}

// NotActive returns the class if the route is not active
func (rc *RouteContext) NotActiveRoute(path, class string) string {
	if rc.URLPath != path {
		return class
	}
	return ""
}

// Error can be used to lookup an error by name
// Example: {{.fir.Error "myevent.field"}} will return the error for the field myevent.field
// Example: {{.fir.Error "myevent" "field"}} will return the error for the event myevent.field
// It can be used in conjunction with ctx.FieldError to get the error for a field
func (rc *RouteContext) Error(paths ...string) any {
	data, _ := json.Marshal(rc.errors)
	return gjson.GetBytes(data, getErrorLookupPath(paths...)).Value()
}
func getErrorLookupPath(paths ...string) string {
	path := ""
	if len(paths) == 0 {
		path = "route"
	} else {
		for _, p := range paths {
			p = strings.Trim(p, ".")
			path += p + "."
		}
	}
	path = strings.Trim(path, ".")
	return path
}
