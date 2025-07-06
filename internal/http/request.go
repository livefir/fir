package http

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"time"
)

// RequestModel represents an abstracted HTTP request that decouples
// business logic from the underlying HTTP implementation.
type RequestModel struct {
	// Basic request information
	Method     string
	URL        *url.URL
	Proto      string
	Header     http.Header
	Body       io.ReadCloser
	Host       string
	RemoteAddr string
	RequestURI string

	// Parsed data
	Form        url.Values
	PostForm    url.Values
	QueryParams url.Values
	PathParams  map[string]string

	// Context and metadata
	Context   context.Context
	Timestamp time.Time

	// Fir-specific fields
	EventID     string
	EventTarget string
	SessionID   string
	ElementKey  string
	IsWebSocket bool
}

// RequestParser is an interface for parsing HTTP requests into RequestModel
type RequestParser interface {
	ParseRequest(*http.Request) (*RequestModel, error)
	ParsePathParams(*http.Request) map[string]string
	ParseEventData(*http.Request) (EventData, error)
}

// EventData represents parsed event information from the request
type EventData struct {
	ID         string
	Target     *string
	ElementKey *string
	Params     map[string]interface{}
}

// RequestAdapter adapts http.Request to RequestModel
type RequestAdapter struct {
	pathParamExtractor func(*http.Request) map[string]string
}

// NewRequestAdapter creates a new RequestAdapter
func NewRequestAdapter(pathParamExtractor func(*http.Request) map[string]string) *RequestAdapter {
	return &RequestAdapter{
		pathParamExtractor: pathParamExtractor,
	}
}

// ParseRequest converts an http.Request to a RequestModel
func (a *RequestAdapter) ParseRequest(r *http.Request) (*RequestModel, error) {
	// Parse form data if not already parsed
	if r.Form == nil {
		if err := r.ParseForm(); err != nil {
			return nil, err
		}
	}

	// Extract path parameters
	var pathParams map[string]string
	if a.pathParamExtractor != nil {
		pathParams = a.pathParamExtractor(r)
	} else {
		pathParams = make(map[string]string)
	}

	return &RequestModel{
		Method:      r.Method,
		URL:         r.URL,
		Proto:       r.Proto,
		Header:      r.Header.Clone(),
		Body:        r.Body,
		Host:        r.Host,
		RemoteAddr:  r.RemoteAddr,
		RequestURI:  r.RequestURI,
		Form:        cloneValues(r.Form),
		PostForm:    cloneValues(r.PostForm),
		QueryParams: cloneValues(r.URL.Query()),
		PathParams:  pathParams,
		Context:     r.Context(),
		Timestamp:   time.Now(),
		IsWebSocket: isWebSocketUpgrade(r),
	}, nil
}

// ParseEventData extracts event-specific data from the request
func (a *RequestAdapter) ParseEventData(r *http.Request) (EventData, error) {
	// Extract event ID from various sources (query, form, header)
	eventID := r.URL.Query().Get("event")
	if eventID == "" {
		eventID = r.FormValue("event")
	}

	// Extract target from form or query
	var target *string
	if t := r.FormValue("target"); t != "" {
		target = &t
	} else if t := r.URL.Query().Get("target"); t != "" {
		target = &t
	}

	// Extract element key
	var elementKey *string
	if ek := r.FormValue("element_key"); ek != "" {
		elementKey = &ek
	} else if ek := r.URL.Query().Get("element_key"); ek != "" {
		elementKey = &ek
	}

	// Extract all form parameters as event params
	params := make(map[string]interface{})
	for key, values := range r.Form {
		if len(values) == 1 {
			params[key] = values[0]
		} else if len(values) > 1 {
			params[key] = values
		}
	}

	return EventData{
		ID:         eventID,
		Target:     target,
		ElementKey: elementKey,
		Params:     params,
	}, nil
}

// ParsePathParams extracts path parameters from the request
func (a *RequestAdapter) ParsePathParams(r *http.Request) map[string]string {
	if a.pathParamExtractor != nil {
		return a.pathParamExtractor(r)
	}
	return make(map[string]string)
}

// Helper functions

func cloneValues(src url.Values) url.Values {
	if src == nil {
		return nil
	}
	dst := make(url.Values)
	for k, vs := range src {
		dst[k] = make([]string, len(vs))
		copy(dst[k], vs)
	}
	return dst
}

func isWebSocketUpgrade(r *http.Request) bool {
	return r.Header.Get("Connection") == "Upgrade" &&
		r.Header.Get("Upgrade") == "websocket"
}
