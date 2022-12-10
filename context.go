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

// Context is the context for a route handler.
// Its methods are used to return data or patch operations to the client.
type Context struct {
	event     Event
	request   *http.Request
	response  http.ResponseWriter
	urlValues url.Values
	route     *route
	isOnLoad  bool
	DOM       dom.Patcher
}

// Bind decodes the event params into the given struct
func (c Context) Bind(v any) error {
	if reflect.ValueOf(v).Kind() != reflect.Ptr {
		return errors.New("bind value must be a pointer")
	}
	if err := c.BindPathParams(v); err != nil {
		return err
	}

	if err := c.BindQueryParams(v); err != nil {
		return err
	}

	return c.BindEventParams(v)
}

func (c Context) BindEventParams(v any) error {
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

func (c Context) BindQueryParams(v any) error {
	return c.route.formDecoder.Decode(v, c.request.URL.Query())
}

func (c Context) BindPathParams(v any) error {
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

// Request returns the http.Request for the current context
func (c Context) Request() *http.Request {
	return c.request
}

// Response returns the http.ResponseWriter for the current context
func (c Context) Response() http.ResponseWriter {
	return c.response
}

// Redirect redirects the client to the given url
func (c Context) Redirect(url string, status int) error {
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
func (c Context) Data(data map[string]any) error {
	m := routeData{}
	for k, v := range data {
		m[k] = v
	}
	return &m
}

// KV sets a key value pair to be hydrated into the route's template
func (c Context) KV(k string, v any) error {
	return &routeData{k: v}
}

// FieldError sets the error message for the given field and can be looked up by {{.fir.Error "myevent.field"}}
func (c Context) FieldError(field string, err error) error {
	if err == nil || field == "" {
		return nil
	}
	return &fieldErrors{field: userError(c, err)}
}

// FieldErrors sets the error messages for the given fields and can be looked up by {{.fir.Error "myevent.field"}}
func (c Context) FieldErrors(fields map[string]error) error {
	m := fieldErrors{}
	for field, err := range fields {
		if err != nil {
			m[field] = userError(c, err)
		}
	}
	return &m
}

// RenderTemplate renders a partial template on the server
func (c *Context) RenderTemplate(name string, data any) dom.TemplateRenderer {
	return dom.NewTemplateRenderer(name, data)
}

// RenderBlock renders a partial template on the server and is an alias for RenderTemplate(...)
func (c *Context) RenderBlock(name string, data any) dom.TemplateRenderer {
	return c.RenderTemplate(name, data)
}

// RenderHTML is a utility function for rendering raw html on the server
func (c *Context) RenderHTML(html string) dom.TemplateRenderer {
	return c.RenderTemplate("_fir_html", html)
}
