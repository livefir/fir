package services

import (
	"errors"
	"net/http"
	"testing"

	firHttp "github.com/livefir/fir/internal/http"
)

func TestDefaultResponseBuilder_BuildEventResponse(t *testing.T) {
	builder := NewDefaultResponseBuilder()

	tests := []struct {
		name           string
		result         *EventResponse
		request        *firHttp.RequestModel
		expectedStatus int
		expectedType   string
		wantErr        bool
	}{
		{
			name:    "nil result",
			result:  nil,
			wantErr: true,
		},
		{
			name: "redirect response",
			result: &EventResponse{
				StatusCode: http.StatusOK,
				Redirect: &firHttp.RedirectInfo{
					URL:        "/test",
					StatusCode: http.StatusFound,
				},
			},
			expectedStatus: http.StatusFound,
			wantErr:        false,
		},
		{
			name: "HTML response",
			result: &EventResponse{
				StatusCode: http.StatusOK,
				Body:       []byte("<h1>Test</h1>"),
			},
			expectedStatus: http.StatusOK,
			expectedType:   "text/html; charset=utf-8",
			wantErr:        false,
		},
		{
			name: "events only response",
			result: &EventResponse{
				StatusCode: http.StatusOK,
				Events: []firHttp.DOMEvent{
					{
						Type:   "update",
						Target: "#test",
					},
				},
			},
			expectedStatus: http.StatusOK,
			expectedType:   "application/json",
			wantErr:        false,
		},
		{
			name: "empty response",
			result: &EventResponse{
				StatusCode: http.StatusOK,
			},
			expectedStatus: http.StatusNoContent,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := builder.BuildEventResponse(tt.result, tt.request)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if response.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, response.StatusCode)
			}

			if tt.expectedType != "" {
				if ct := response.Headers["Content-Type"]; ct != tt.expectedType {
					t.Errorf("expected content type %s, got %s", tt.expectedType, ct)
				}
			}
		})
	}
}

func TestDefaultResponseBuilder_BuildTemplateResponse(t *testing.T) {
	builder := NewDefaultResponseBuilder()

	tests := []struct {
		name           string
		render         *RenderResult
		statusCode     int
		expectedStatus int
		wantErr        bool
	}{
		{
			name:    "nil render result",
			render:  nil,
			wantErr: true,
		},
		{
			name: "successful render",
			render: &RenderResult{
				HTML: []byte("<h1>Test</h1>"),
			},
			statusCode:     http.StatusOK,
			expectedStatus: http.StatusOK,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := builder.BuildTemplateResponse(tt.render, tt.statusCode)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if response.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, response.StatusCode)
			}

			if ct := response.Headers["Content-Type"]; ct != "text/html; charset=utf-8" {
				t.Errorf("expected HTML content type, got %s", ct)
			}
		})
	}
}

func TestDefaultResponseBuilder_BuildErrorResponse(t *testing.T) {
	builder := NewDefaultResponseBuilder()

	tests := []struct {
		name           string
		err            error
		statusCode     int
		expectedStatus int
		wantErr        bool
	}{
		{
			name:    "nil error",
			err:     nil,
			wantErr: true,
		},
		{
			name:           "error with status code",
			err:            errors.New("test error"),
			statusCode:     http.StatusBadRequest,
			expectedStatus: http.StatusBadRequest,
			wantErr:        false,
		},
		{
			name:           "error with default status code",
			err:            errors.New("test error"),
			statusCode:     0,
			expectedStatus: http.StatusInternalServerError,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := builder.BuildErrorResponse(tt.err, tt.statusCode)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if response.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, response.StatusCode)
			}

			if ct := response.Headers["Content-Type"]; ct != "text/plain; charset=utf-8" {
				t.Errorf("expected plain text content type, got %s", ct)
			}
		})
	}
}

func TestDefaultResponseBuilder_BuildRedirectResponse(t *testing.T) {
	builder := NewDefaultResponseBuilder()

	tests := []struct {
		name           string
		url            string
		statusCode     int
		expectedStatus int
		wantErr        bool
	}{
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
		},
		{
			name:           "valid redirect with status",
			url:            "/test",
			statusCode:     http.StatusMovedPermanently,
			expectedStatus: http.StatusMovedPermanently,
			wantErr:        false,
		},
		{
			name:           "valid redirect with default status",
			url:            "/test",
			statusCode:     0,
			expectedStatus: http.StatusTemporaryRedirect,
			wantErr:        false,
		},
		{
			name:       "invalid status code",
			url:        "/test",
			statusCode: http.StatusOK,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := builder.BuildRedirectResponse(tt.url, tt.statusCode)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if response.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, response.StatusCode)
			}

			if location := response.Headers["Location"]; location != tt.url {
				t.Errorf("expected location %s, got %s", tt.url, location)
			}

			if response.Redirect == nil {
				t.Errorf("expected redirect info but got nil")
			} else {
				if response.Redirect.URL != tt.url {
					t.Errorf("expected redirect URL %s, got %s", tt.url, response.Redirect.URL)
				}
				if response.Redirect.StatusCode != tt.expectedStatus {
					t.Errorf("expected redirect status %d, got %d", tt.expectedStatus, response.Redirect.StatusCode)
				}
			}
		})
	}
}
