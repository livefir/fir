package handlers

import (
	"context"
	"errors"

	"github.com/livefir/fir/internal/http"
)

// POCHandler is a minimal proof-of-concept handler for Phase 0
type POCHandler struct{}

// NewPOCHandler creates a new POCHandler instance
func NewPOCHandler() *POCHandler {
	return &POCHandler{}
}

// SupportsRequest returns true only for GET /poc requests
func (h *POCHandler) SupportsRequest(req *http.RequestModel) bool {
	return req.Method == "GET" && req.URL.Path == "/poc"
}

// Handle processes the request and returns a simple "POC Working" response
func (h *POCHandler) Handle(ctx context.Context, req *http.RequestModel) (*http.ResponseModel, error) {
	if !h.SupportsRequest(req) {
		return nil, errors.New("request not supported by POCHandler")
	}

	return &http.ResponseModel{
		StatusCode: 200,
		Headers:    make(map[string]string),
		Body:       []byte("POC Working"),
	}, nil
}

// HandlerName returns the name of this handler
func (h *POCHandler) HandlerName() string {
	return "POCHandler"
}
