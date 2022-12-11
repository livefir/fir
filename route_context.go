package fir

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"reflect"

	"github.com/adnaan/fir/dom"
	"github.com/fatih/structs"
)

// RouteContext is the context for a route handler.
// Its methods are used to return data or patch operations to the client.
type RouteContext struct {
	event     Event
	request   *http.Request
	response  http.ResponseWriter
	urlValues url.Values
	route     *route
	isOnLoad  bool
	dom       dom.Patcher
}

func (c RouteContext) Event() Event {
	return c.event
}

func (c RouteContext) DOM() dom.Patcher {
	return c.dom
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
	if c.event.IsForm {
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

// Data sets the data to be hydrated into the route's template
func (c RouteContext) Data(data map[string]any) error {
	m := routeData{}
	for k, v := range data {
		m[k] = v
	}
	return &m
}

// KV sets a key value pair to be hydrated into the route's template
func (c RouteContext) KV(k string, v any) error {
	return &routeData{k: v}
}

// FieldError sets the error message for the given field and can be looked up by {{.fir.Error "myevent.field"}}
func (c RouteContext) FieldError(field string, err error) error {
	if err == nil || field == "" {
		return nil
	}
	return &fieldErrors{field: userError(c, err)}
}

// FieldErrors sets the error messages for the given fields and can be looked up by {{.fir.Error "myevent.field"}}
func (c RouteContext) FieldErrors(fields map[string]error) error {
	m := fieldErrors{}
	for field, err := range fields {
		if err != nil {
			m[field] = userError(c, err)
		}
	}
	return &m
}

// RenderTemplate renders a partial template on the server
func (c *RouteContext) RenderTemplate(name string, data any) dom.TemplateRenderer {
	return dom.NewTemplateRenderer(name, data)
}

// RenderBlock renders a partial template on the server and is an alias for RenderTemplate(...)
func (c *RouteContext) RenderBlock(name string, data any) dom.TemplateRenderer {
	return c.RenderTemplate(name, data)
}

// RenderHTML is a utility function for rendering raw html on the server
func (c *RouteContext) RenderHTML(html string) dom.TemplateRenderer {
	return c.RenderTemplate("_fir_html", html)
}
