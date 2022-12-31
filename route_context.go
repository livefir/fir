package fir

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"reflect"

	"github.com/alexedwards/scs/v2"
	"github.com/fatih/structs"
	"github.com/livefir/fir/internal/dom"
	firErrors "github.com/livefir/fir/internal/errors"
)

// RouteContext is the context for a route handler.
// Its methods are used to return data or patch operations to the client.
type RouteContext struct {
	event      Event
	request    *http.Request
	response   http.ResponseWriter
	urlValues  url.Values
	route      *route
	isOnLoad   bool
	domPatcher dom.Patcher
	session    *scs.SessionManager
}

func (c RouteContext) Event() Event {
	return c.event
}

func (c RouteContext) dom() dom.Patcher {
	return c.domPatcher
}

// Bind decodes the event params into the given struct
func (c RouteContext) Bind(v any) error {
	if v == nil {
		return errors.New("bind value cannot be nil")
	}
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return errors.New("bind value must be a pointer to a struct")
	}
	if val.Kind() == reflect.Ptr {
		val = val.Elem() // dereference the pointer
		if val.Kind() != reflect.Struct {
			return errors.New("bind value must be a pointer to a struct")
		}
	}
	if err := c.BindPathParams(v); err != nil {
		return err
	}

	if err := c.BindQueryParams(v); err != nil {
		return err
	}

	return c.BindEventParams(v)
}

func (c RouteContext) BindPathParams(v any) error {
	if v == nil {
		return nil // nothing to bind
	}
	m, ok := v.(map[string]any)
	if ok {
		for k := range m {
			if value := c.request.Context().Value(k); value != nil {
				m[k] = value
			}
		}
		v = m
		return nil
	}
	s := structs.New(v)
	for _, field := range s.Fields() {
		if field.IsExported() {
			if value := c.request.Context().Value(field.Tag("json")); value != nil && !reflect.ValueOf(v).IsZero() {
				err := field.Set(value)
				if err != nil {
					return err
				}
				continue
			}

			if value := c.request.Context().Value(field.Name()); value != nil && !reflect.ValueOf(v).IsZero() {
				err := field.Set(value)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (c RouteContext) BindQueryParams(v any) error {
	return c.route.formDecoder.Decode(v, c.request.URL.Query())
}

func (c RouteContext) BindEventParams(v any) error {
	if c.event.Params == nil {
		return nil
	}
	if c.event.FormID != nil {
		if len(c.urlValues) == 0 {
			var urlValues url.Values
			if err := json.NewDecoder(bytes.NewReader(c.event.Params)).Decode(&urlValues); err != nil {
				return err
			}
			c.urlValues = urlValues
		}
		return c.route.formDecoder.Decode(v, c.urlValues)
	}

	return json.NewDecoder(bytes.NewReader(c.event.Params)).Decode(v)
}

// Request returns the http.Request for the current context
func (c RouteContext) Request() *http.Request {
	return c.request
}

// Response returns the http.ResponseWriter for the current context
func (c RouteContext) Response() http.ResponseWriter {
	return c.response
}

// Redirect redirects the client to the given url
func (c RouteContext) Redirect(url string, status int) error {
	if url == "" {
		return errors.New("url is required")
	}
	if status < 300 || status > 308 {
		return errors.New("status code must be between 300 and 308")
	}
	http.Redirect(c.response, c.request, url, status)
	return nil
}

// KV is a wrapper for ctx.Data(map[string]any{key: data})
func (c RouteContext) KV(key string, data any) error {
	return c.Data(map[string]any{key: data})
}

// Data sets the data to be hydrated into the route's template or an event's associated template/block action
// It accepts either a map or struct type so that fir can inject utility functions: .fir.Error, .fir.ActiveRoute etc.
// If the data is a struct, it will be converted to a map using github.com/fatih/structs
// If the data is a pointer to a struct, it will be dereferenced and converted to a map using github.com/fatih/structs
// If the data is a map, it will be used as is
// If the data is a pointer to a map, it will be dereferenced and used as is
// The function will return nil if no data is passed
// The function accepts variadic arguments so that you can pass multiple structs or maps which will be merged
func (c RouteContext) Data(dataset ...any) error {
	if len(dataset) == 0 {
		return nil
	}
	m := routeData{}
	for _, data := range dataset {
		val := reflect.ValueOf(data)
		if val.Kind() == reflect.Ptr {
			el := val.Elem() // dereference the pointer
			if el.Kind() == reflect.Struct {
				for k, v := range structs.Map(data) {
					m[k] = v
				}
			}
		} else if val.Kind() == reflect.Struct {
			for k, v := range structs.Map(data) {
				m[k] = v
			}
		} else if val.Kind() == reflect.Map {
			ms, ok := data.(map[string]any)
			if !ok {
				return errors.New("data must be a map[string]any , struct or pointer to a struct")
			}

			for k, v := range ms {
				m[k] = v
			}
		} else {
			return errors.New("data must be a map[string]any , struct or pointer to a struct")
		}
	}

	return &m
}

// FieldError sets the error message for the given field and can be looked up by {{.fir.Error "myevent.field"}}
func (c RouteContext) FieldError(field string, err error) error {
	if err == nil || field == "" {
		return nil
	}
	return &firErrors.Fields{field: firErrors.User(err)}
}

// FieldErrors sets the error messages for the given fields and can be looked up by {{.fir.Error "myevent.field"}}
func (c RouteContext) FieldErrors(fields map[string]error) error {
	m := firErrors.Fields{}
	for field, err := range fields {
		if err != nil {
			m[field] = firErrors.User(err)
		}
	}
	return &m
}

// renderTemplate renders a partial template on the server
func (c *RouteContext) renderTemplate(name string, data any) dom.TemplateRenderer {
	return dom.NewTemplateRenderer(name, data)
}

// renderBlock renders a partial template on the server and is an alias for RenderTemplate(...)
func (c *RouteContext) renderBlock(name string, data any) dom.TemplateRenderer {
	return c.renderTemplate(name, data)
}

// renderHTML is a utility function for rendering raw html on the server
func (c *RouteContext) renderHTML(html string) dom.TemplateRenderer {
	return c.renderTemplate("_fir_html", html)
}
