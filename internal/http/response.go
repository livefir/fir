package http

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ResponseModel represents an abstracted HTTP response that decouples
// business logic from the underlying HTTP implementation.
type ResponseModel struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte
	Events     []DOMEvent
	Redirect   *RedirectInfo
}

// DOMEvent represents a DOM manipulation event to be sent to the client
type DOMEvent struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Target     string                 `json:"target,omitempty"`
	ElementKey string                 `json:"element_key,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty"`
	HTML       string                 `json:"html,omitempty"`
}

// RedirectInfo contains redirect information
type RedirectInfo struct {
	URL        string
	StatusCode int
}

// ResponseWriter is an interface for writing HTTP responses
type ResponseWriter interface {
	WriteResponse(ResponseModel) error
	WriteError(statusCode int, message string) error
	WriteRedirect(url string, statusCode int) error
	WriteJSON(data interface{}) error
	WriteHTML(html string) error
}

// ResponseAdapter adapts ResponseModel to http.ResponseWriter
type ResponseAdapter struct {
	writer  http.ResponseWriter
	written bool
}

// NewResponseAdapter creates a new ResponseAdapter
func NewResponseAdapter(w http.ResponseWriter) *ResponseAdapter {
	return &ResponseAdapter{
		writer:  w,
		written: false,
	}
}

// WriteResponse writes a ResponseModel to the underlying http.ResponseWriter
func (a *ResponseAdapter) WriteResponse(resp ResponseModel) error {
	if a.written {
		return fmt.Errorf("response already written")
	}

	// Handle redirect
	if resp.Redirect != nil {
		return a.WriteRedirect(resp.Redirect.URL, resp.Redirect.StatusCode)
	}

	// Set headers
	for key, value := range resp.Headers {
		a.writer.Header().Set(key, value)
	}

	// Handle DOM events (for JSON responses)
	if len(resp.Events) > 0 {
		a.writer.Header().Set("Content-Type", "application/json")
		if resp.StatusCode > 0 {
			a.writer.WriteHeader(resp.StatusCode)
		}

		eventsJSON, err := json.Marshal(resp.Events)
		if err != nil {
			return fmt.Errorf("failed to marshal events: %w", err)
		}

		_, err = a.writer.Write(eventsJSON)
		a.written = true
		return err
	}

	// Set status code
	if resp.StatusCode > 0 {
		a.writer.WriteHeader(resp.StatusCode)
	}

	// Write body
	if len(resp.Body) > 0 {
		_, err := a.writer.Write(resp.Body)
		a.written = true
		return err
	}

	a.written = true
	return nil
}

// WriteError writes an error response
func (a *ResponseAdapter) WriteError(statusCode int, message string) error {
	if a.written {
		return fmt.Errorf("response already written")
	}

	http.Error(a.writer, message, statusCode)
	a.written = true
	return nil
}

// WriteRedirect writes a redirect response
func (a *ResponseAdapter) WriteRedirect(url string, statusCode int) error {
	if a.written {
		return fmt.Errorf("response already written")
	}

	a.writer.Header().Set("Location", url)
	a.writer.WriteHeader(statusCode)
	a.written = true
	return nil
}

// WriteJSON writes a JSON response
func (a *ResponseAdapter) WriteJSON(data interface{}) error {
	if a.written {
		return fmt.Errorf("response already written")
	}

	a.writer.Header().Set("Content-Type", "application/json")

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	_, err = a.writer.Write(jsonData)
	a.written = true
	return err
}

// WriteHTML writes an HTML response
func (a *ResponseAdapter) WriteHTML(html string) error {
	if a.written {
		return fmt.Errorf("response already written")
	}

	a.writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err := a.writer.Write([]byte(html))
	a.written = true
	return err
}

// IsWritten returns true if the response has been written
func (a *ResponseAdapter) IsWritten() bool {
	return a.written
}

// ResponseBuilder helps build ResponseModel instances
type ResponseBuilder struct {
	model ResponseModel
}

// NewResponseBuilder creates a new ResponseBuilder
func NewResponseBuilder() *ResponseBuilder {
	return &ResponseBuilder{
		model: ResponseModel{
			Headers: make(map[string]string),
			Events:  make([]DOMEvent, 0),
		},
	}
}

// WithStatus sets the response status code
func (b *ResponseBuilder) WithStatus(code int) *ResponseBuilder {
	b.model.StatusCode = code
	return b
}

// WithHeader adds a header to the response
func (b *ResponseBuilder) WithHeader(key, value string) *ResponseBuilder {
	b.model.Headers[key] = value
	return b
}

// WithBody sets the response body
func (b *ResponseBuilder) WithBody(body []byte) *ResponseBuilder {
	b.model.Body = body
	return b
}

// WithHTML sets the response body as HTML
func (b *ResponseBuilder) WithHTML(html string) *ResponseBuilder {
	b.model.Body = []byte(html)
	b.model.Headers["Content-Type"] = "text/html; charset=utf-8"
	return b
}

// WithJSON sets the response body as JSON
func (b *ResponseBuilder) WithJSON(data interface{}) *ResponseBuilder {
	jsonData, err := json.Marshal(data)
	if err != nil {
		// For builder pattern, we store the error in a way that can be checked later
		b.model.Body = []byte(fmt.Sprintf(`{"error": "failed to marshal JSON: %s"}`, err.Error()))
		b.model.StatusCode = http.StatusInternalServerError
	} else {
		b.model.Body = jsonData
	}
	b.model.Headers["Content-Type"] = "application/json"
	return b
}

// WithRedirect sets up a redirect response
func (b *ResponseBuilder) WithRedirect(url string, statusCode int) *ResponseBuilder {
	b.model.Redirect = &RedirectInfo{
		URL:        url,
		StatusCode: statusCode,
	}
	return b
}

// WithEvent adds a DOM event to the response
func (b *ResponseBuilder) WithEvent(event DOMEvent) *ResponseBuilder {
	b.model.Events = append(b.model.Events, event)
	return b
}

// Build returns the constructed ResponseModel
func (b *ResponseBuilder) Build() ResponseModel {
	return b.model
}
