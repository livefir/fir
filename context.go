package fir

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/adnaan/fir/patch"
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
}

// DecodeParams decodes the event params into the given struct
func (c Context) DecodeParams(v any) error {
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

// MorphKV is a shortcut for Morph(k, Block(k, v))
func (c Context) MorphKV(name string, value any) error {
	return c.Morph(fmt.Sprintf("#%s", name), patch.Block(name, map[string]any{name: value}))
}

// func (c Context) AppendKV(name string, value any) error {
// 	return c.Append(fmt.Sprintf("#%s", name), Block(name, M{name: value}))
// }

// func (c Context) AfterKV(name string, value any) error {
// 	return c.After(fmt.Sprintf("#%s", name), Block(name, M{name: value}))
// }

// func (c Context) BeforeKV(name string, value any) error {
// 	return c.Before(fmt.Sprintf("#%s", name), Block(name, M{name: value}))
// }

//	func (c Context) PrependKV(name string, value any) error {
//		return c.Prepend(fmt.Sprintf("#%s", name), Block(name, M{name: value}))
//	}
//
// Patch adds a list of patch operations to the response
func (c Context) Patch(patches ...patch.Op) error {
	var pl patch.Set
	for _, p := range patches {
		pl = append(pl, p)
	}
	return &pl
}

// Morph replaces the element at the given selector with the given template
func (c Context) Morph(selector string, t patch.TemplateRenderer) error {
	return c.Patch(patch.Morph(selector, t))
}

// After inserts the given template after the element at the given selector
func (c Context) After(selector string, t patch.TemplateRenderer) error {
	return c.Patch(patch.After(selector, t))
}

// Before inserts the given template before the element at the given selector
func (c Context) Before(selector string, t patch.TemplateRenderer) error {
	return c.Patch(patch.Before(selector, t))
}

// Append inserts the given template after the element at the given selector
func (c Context) Append(selector string, t patch.TemplateRenderer) error {
	return c.Patch(patch.Append(selector, t))
}

// Prepend inserts the given template before the element at the given selector
func (c Context) Prepend(selector string, t patch.TemplateRenderer) error {
	return c.Patch(patch.Prepend(selector, t))
}

// Remove removes the element at the given selector
func (c Context) Remove(selector string) error {
	return c.Patch(patch.Remove(selector))
}

// Navigate navigates the client to the given url
func (c Context) Navigate(url string) error {
	return c.Patch(patch.Navigate(url))
}

// Store updates the named alpinejs store with the given value
func (c Context) Store(name string, data any) error {
	return c.Patch(patch.Store(name, data))
}

// Reload reloads the current page
func (c Context) Reload() error {
	return c.Patch(patch.Reload())
}

// ResetForm resets the form to its initial state
func (c Context) ResetForm(selector string) error {
	return c.Patch(patch.ResetForm(selector))
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
