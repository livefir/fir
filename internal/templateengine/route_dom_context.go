package templateengine

import (
	"strings"

	"github.com/goccy/go-json"
	"github.com/tidwall/gjson"
)

// RouteDOMContext is a struct that holds route context data and is passed to the template.
// This is a decoupled version of the original RouteDOMContext that doesn't depend on
// the full route infrastructure.
type RouteDOMContext struct {
	Name        string
	Development bool
	URLPath     string
	errors      map[string]interface{}
}

// NewRouteDOMContext creates a new RouteDOMContext from a FuncMapContext.
func NewRouteDOMContext(ctx FuncMapContext) *RouteDOMContext {
	errors := ctx.Errors
	if errors == nil {
		errors = make(map[string]interface{})
	}

	return &RouteDOMContext{
		URLPath:     ctx.URLPath,
		Name:        ctx.AppName,
		Development: ctx.DevelopmentMode,
		errors:      errors,
	}
}

// ActiveRoute returns the class if the route is active
func (rc *RouteDOMContext) ActiveRoute(path, class string) string {
	if rc.URLPath == path {
		return class
	}
	return ""
}

// NotActiveRoute returns the class if the route is not active
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
func (rc *RouteDOMContext) Error(paths ...string) interface{} {
	if len(rc.errors) == 0 {
		return nil
	}
	data, _ := json.Marshal(rc.errors)
	val := gjson.GetBytes(data, getErrorLookupPath(paths...)).Value()
	_, ok := val.(map[string]interface{})
	if ok {
		return nil
	}
	return val
}

// getErrorLookupPath constructs the path for error lookup
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
