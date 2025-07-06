package http

import (
	"net/http"
)

// HTTPAdapter provides a unified interface for HTTP request/response handling
// that abstracts away the underlying HTTP implementation details.
type HTTPAdapter interface {
	RequestParser
	ResponseWriter
}

// StandardHTTPAdapter implements HTTPAdapter using standard HTTP components
type StandardHTTPAdapter struct {
	*RequestAdapter
	*ResponseAdapter
}

// NewStandardHTTPAdapter creates a new StandardHTTPAdapter
func NewStandardHTTPAdapter(w http.ResponseWriter, r *http.Request, pathParamExtractor func(*http.Request) map[string]string) *StandardHTTPAdapter {
	return &StandardHTTPAdapter{
		RequestAdapter:  NewRequestAdapter(pathParamExtractor),
		ResponseAdapter: NewResponseAdapter(w),
	}
}

// RequestResponsePair holds a parsed request and response writer for processing
type RequestResponsePair struct {
	Request  *RequestModel
	Response ResponseWriter
}

// NewRequestResponsePair creates a new RequestResponsePair from HTTP primitives
func NewRequestResponsePair(w http.ResponseWriter, r *http.Request, pathParamExtractor func(*http.Request) map[string]string) (*RequestResponsePair, error) {
	adapter := NewStandardHTTPAdapter(w, r, pathParamExtractor)

	request, err := adapter.ParseRequest(r)
	if err != nil {
		return nil, err
	}

	return &RequestResponsePair{
		Request:  request,
		Response: adapter,
	}, nil
}

// HandleHTTPRequest is a convenience function that creates a RequestResponsePair
// and executes a handler function with proper error handling
func HandleHTTPRequest(w http.ResponseWriter, r *http.Request, pathParamExtractor func(*http.Request) map[string]string, handler func(*RequestResponsePair) error) {
	pair, err := NewRequestResponsePair(w, r, pathParamExtractor)
	if err != nil {
		http.Error(w, "Failed to parse request", http.StatusBadRequest)
		return
	}

	if err := handler(pair); err != nil {
		// Only write error if response hasn't been written yet
		if adapter, ok := pair.Response.(*ResponseAdapter); ok && !adapter.IsWritten() {
			pair.Response.WriteError(http.StatusInternalServerError, err.Error())
		} else {
			// Fallback: try to write error anyway
			pair.Response.WriteError(http.StatusInternalServerError, err.Error())
		}
	}
}
