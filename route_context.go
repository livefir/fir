package fir

import (
	"bytes"
	"encoding/gob"
	"errors"
	"net/http"
	"net/url"
	"reflect"

	"github.com/goccy/go-json"

	"github.com/fatih/structs"

	firErrors "github.com/livefir/fir/internal/errors"
)

type ContextKey int

const (
	// PathParamsKey is the key for the path params in the request context.
	PathParamsKey ContextKey = iota
	// UserKey is the key for the user id/name in the request context. It is used in the default channel function.
	UserKey
)

type PathParams map[string]any

type userStore map[string]any

func init() {
	gob.Register(userStore{})
}

// RouteContext is the context for a route handler.
// Its methods are used to return data or patch operations to the client.
type RouteContext struct {
	event     Event
	request   *http.Request
	response  http.ResponseWriter
	urlValues url.Values
	route     *route
	isOnLoad  bool
}

func (c RouteContext) Event() Event {
	return c.event
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
	pathParams, ok := c.request.Context().Value(PathParamsKey).(PathParams)
	if !ok {
		return nil
	}
	s := structs.New(v)
	for _, field := range s.Fields() {
		if field.IsExported() {
			if v, ok := pathParams[field.Tag("json")]; ok {
				err := field.Set(v)
				if err != nil {
					return err
				}
				continue
			}

			if v, ok := pathParams[field.Tag("json")]; ok {
				err := field.Set(v)
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

// KV is a wrapper for ctx.Data(map[string]any{key: data})
func (c RouteContext) KV(key string, data any) error {
	return buildData(false, map[string]any{key: data})
}

// KV is a wrapper for ctx.State(map[string]any{key: data})
func (c RouteContext) StateKV(key string, data any) error {
	return buildData(true, map[string]any{key: data})
}

// State data is only passed to event receiver without a bound template
// it can be acccessed in the event receiver via $event.detail
// e.g. @fir:myevent:ok="console.log('$event.detail.mykey')"
func (c RouteContext) State(dataset ...any) error {
	return buildData(true, dataset...)
}

// Data sets the data to be hydrated into the route's template or an event's associated template/block action
// It accepts either a map or struct type so that fir can inject utility functions: fir.Error, fir.ActiveRoute etc.
// If the data is a struct, it will be converted to a map using github.com/fatih/structs
// If the data is a pointer to a struct, it will be dereferenced and converted to a map using github.com/fatih/structs
// If the data is a map, it will be used as is
// If the data is a pointer to a map, it will be dereferenced and used as is
// The function will return nil if no data is passed
// The function accepts variadic arguments so that you can pass multiple structs or maps which will be merged
func (c RouteContext) Data(dataset ...any) error {
	return buildData(false, dataset...)
}

// FieldError sets the error message for the given field and can be looked up by {{fir.Error "myevent.field"}}
func (c RouteContext) FieldError(field string, err error) error {
	if err == nil || field == "" {
		return nil
	}
	return &firErrors.Fields{field: firErrors.User(err)}
}

// FieldErrors sets the error messages for the given fields and can be looked up by {{fir.Error "myevent.field"}}
func (c RouteContext) FieldErrors(fields map[string]error) error {
	m := firErrors.Fields{}
	for field, err := range fields {
		if err != nil {
			m[field] = firErrors.User(err)
		}
	}
	return &m
}

func (c RouteContext) Status(code int, err error) error {
	return &firErrors.Status{Code: code, Err: firErrors.User(err)}
}

func (c RouteContext) GetUserFromContext() string {
	return getUserFromRequestContext(c.request)
}

func getUserFromRequestContext(r *http.Request) string {
	user, ok := r.Context().Value(UserKey).(string)
	if !ok {
		return ""
	}
	return user
}
