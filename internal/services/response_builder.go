package services

import (
	"fmt"
	"net/http"

	firHttp "github.com/livefir/fir/internal/http"
)

// DefaultResponseBuilder is the default implementation of ResponseBuilder
type DefaultResponseBuilder struct{}

// NewDefaultResponseBuilder creates a new default response builder
func NewDefaultResponseBuilder() *DefaultResponseBuilder {
	return &DefaultResponseBuilder{}
}

// BuildEventResponse builds a response from an event processing result
func (b *DefaultResponseBuilder) BuildEventResponse(result *EventResponse, request *firHttp.RequestModel) (*firHttp.ResponseModel, error) {
	if result == nil {
		return nil, fmt.Errorf("event response result cannot be nil")
	}

	response := &firHttp.ResponseModel{
		StatusCode: result.StatusCode,
		Headers:    make(map[string]string),
		Events:     result.Events,
	}

	// Copy headers from result
	for k, v := range result.Headers {
		response.Headers[k] = v
	}

	// Handle different response types
	switch {
	case result.Redirect != nil:
		response.StatusCode = result.Redirect.StatusCode
		response.Headers["Location"] = result.Redirect.URL
		return response, nil

	case len(result.Body) > 0:
		response.Body = result.Body
		if response.Headers["Content-Type"] == "" {
			response.Headers["Content-Type"] = "text/html; charset=utf-8"
		}
		return response, nil

	case len(result.Events) > 0:
		// For event-only responses, we might want to return minimal HTML or JSON
		response.Headers["Content-Type"] = "application/json"
		// The events will be handled by the HTTP adapter
		return response, nil

	default:
		// Empty successful response
		response.StatusCode = http.StatusNoContent
		return response, nil
	}
}

// BuildTemplateResponse builds a response from a template render result
func (b *DefaultResponseBuilder) BuildTemplateResponse(render *RenderResult, statusCode int) (*firHttp.ResponseModel, error) {
	if render == nil {
		return nil, fmt.Errorf("render result cannot be nil")
	}

	response := &firHttp.ResponseModel{
		StatusCode: statusCode,
		Headers:    make(map[string]string),
		Body:       render.HTML,
		Events:     render.Events,
	}

	// Set content type
	response.Headers["Content-Type"] = "text/html; charset=utf-8"

	return response, nil
}

// BuildErrorResponse builds an error response
func (b *DefaultResponseBuilder) BuildErrorResponse(err error, statusCode int) (*firHttp.ResponseModel, error) {
	if err == nil {
		return nil, fmt.Errorf("error cannot be nil")
	}

	// Default status code if not provided
	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}

	response := &firHttp.ResponseModel{
		StatusCode: statusCode,
		Headers:    make(map[string]string),
		Body:       []byte(err.Error()),
	}

	// Set content type
	response.Headers["Content-Type"] = "text/plain; charset=utf-8"

	return response, nil
}

// BuildRedirectResponse builds a redirect response
func (b *DefaultResponseBuilder) BuildRedirectResponse(url string, statusCode int) (*firHttp.ResponseModel, error) {
	if url == "" {
		return nil, fmt.Errorf("redirect URL cannot be empty")
	}

	// Default to temporary redirect if no status code provided
	if statusCode == 0 {
		statusCode = http.StatusTemporaryRedirect
	}

	// Validate redirect status codes
	if statusCode < 300 || statusCode >= 400 {
		return nil, fmt.Errorf("invalid redirect status code: %d", statusCode)
	}

	response := &firHttp.ResponseModel{
		StatusCode: statusCode,
		Headers:    make(map[string]string),
		Redirect: &firHttp.RedirectInfo{
			URL:        url,
			StatusCode: statusCode,
		},
	}

	// Set Location header
	response.Headers["Location"] = url

	return response, nil
}
