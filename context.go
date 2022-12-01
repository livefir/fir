package fir

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

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

func (c Context) Request() *http.Request {
	return c.request
}

func (c Context) Response() http.ResponseWriter {
	return c.response
}

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

func (c Context) Data(data map[string]any) error {
	m := routeData{}
	for k, v := range data {
		m[k] = v
	}
	return &m
}

func (c Context) KV(k string, v any) error {
	return &routeData{k: v}
}

func (c Context) MorphKV(name string, value any) error {
	return c.Morph(fmt.Sprintf("#%s", name), Block(name, M{name: value}))
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

// func (c Context) PrependKV(name string, value any) error {
// 	return c.Prepend(fmt.Sprintf("#%s", name), Block(name, M{name: value}))
// }

func (c Context) Patch(patch ...Patch) error {
	var pl patchlist
	for _, p := range patch {
		pl = append(pl, p)
	}
	return &pl

}

func (c Context) Morph(selector string, t TemplateRenderer) error {
	return c.Patch(Morph(selector, t))
}
func (c Context) After(selector string, t TemplateRenderer) error {
	return c.Patch(After(selector, t))
}
func (c Context) Before(selector string, t TemplateRenderer) error {
	return c.Patch(Before(selector, t))
}
func (c Context) Append(selector string, t TemplateRenderer) error {
	return c.Patch(Append(selector, t))
}
func (c Context) Prepend(selector string, t TemplateRenderer) error {
	return c.Patch(Prepend(selector, t))
}
func (c Context) Remove(selector string) error {
	return c.Patch(Remove(selector))
}
func (c Context) Navigate(url string) error {
	return c.Patch(Navigate(url))
}
func (c Context) Store(name string, data any) error {
	return c.Patch(Store(name, data))
}
func (c Context) Reload() error {
	return c.Patch(Reload())
}
func (c Context) ResetForm(selector string) error {
	return c.Patch(ResetForm(selector))
}

func (c Context) FieldError(field string, err error) error {
	if err == nil || field == "" {
		return nil
	}
	return &fieldErrors{field: UserError(c, err)}
}

func (c Context) FieldErrors(fields map[string]error) error {
	m := fieldErrors{}
	for field, err := range fields {
		if err != nil {
			m[field] = UserError(c, err)
		}
	}
	return &m
}
