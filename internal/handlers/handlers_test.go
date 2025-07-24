package handlers

import (
	"context"
	"errors"
	"html/template"
	"io"
	"net/http"
	"strings"
	"testing"

	firHttp "github.com/livefir/fir/internal/http"
	"github.com/livefir/fir/internal/services"
	"github.com/livefir/fir/pubsub"
)

// Mock services for testing

type mockEventService struct {
	processEventResponse *services.EventResponse
	processEventError    error
}

func (m *mockEventService) ProcessEvent(ctx context.Context, request services.EventRequest) (*services.EventResponse, error) {
	return m.processEventResponse, m.processEventError
}

func (m *mockEventService) RegisterHandler(eventID string, handler services.EventHandler) error {
	return nil
}

func (m *mockEventService) GetEventMetrics() services.EventMetrics {
	return services.EventMetrics{}
}

type mockRenderService struct {
	renderTemplateResponse *services.RenderResult
	renderTemplateError    error
}

func (m *mockRenderService) RenderTemplate(ctx services.RenderContext) (*services.RenderResult, error) {
	return m.renderTemplateResponse, m.renderTemplateError
}

func (m *mockRenderService) RenderError(ctx services.ErrorContext) (*services.RenderResult, error) {
	return nil, errors.New("not implemented")
}

func (m *mockRenderService) RenderEvents(events []pubsub.Event, routeID string) ([]firHttp.DOMEvent, error) {
	return nil, errors.New("not implemented")
}

type mockTemplateService struct{}

func (m *mockTemplateService) LoadTemplate(config services.TemplateConfig) (*template.Template, error) {
	return nil, errors.New("not implemented")
}

func (m *mockTemplateService) ParseTemplate(content, layout string, partials []string, funcMap template.FuncMap) (*template.Template, error) {
	return nil, errors.New("not implemented")
}

func (m *mockTemplateService) GetTemplate(routeID string, templateType services.TemplateType) (*template.Template, error) {
	return nil, errors.New("not implemented")
}

func (m *mockTemplateService) ClearCache() error {
	return nil
}

func (m *mockTemplateService) SetCacheEnabled(enabled bool) {}

type mockResponseBuilder struct {
	buildEventResponse    *firHttp.ResponseModel
	buildTemplateResponse *firHttp.ResponseModel
	buildErrorResponse    *firHttp.ResponseModel
	buildRedirectResponse *firHttp.ResponseModel
	buildError            error
}

func (m *mockResponseBuilder) BuildEventResponse(result *services.EventResponse, request *firHttp.RequestModel) (*firHttp.ResponseModel, error) {
	return m.buildEventResponse, m.buildError
}

func (m *mockResponseBuilder) BuildTemplateResponse(render *services.RenderResult, statusCode int) (*firHttp.ResponseModel, error) {
	return m.buildTemplateResponse, m.buildError
}

func (m *mockResponseBuilder) BuildErrorResponse(err error, statusCode int) (*firHttp.ResponseModel, error) {
	return m.buildErrorResponse, m.buildError
}

func (m *mockResponseBuilder) BuildRedirectResponse(url string, statusCode int) (*firHttp.ResponseModel, error) {
	return m.buildRedirectResponse, m.buildError
}

type mockEventValidator struct {
	validateError error
}

func (m *mockEventValidator) ValidateEvent(req services.EventRequest) error {
	return m.validateError
}

func (m *mockEventValidator) ValidateParams(eventID string, params map[string]interface{}) error {
	return m.validateError
}

// mockReadCloser wraps strings.Reader to implement io.ReadCloser
type mockReadCloser struct {
	*strings.Reader
}

func (m mockReadCloser) Close() error {
	return nil
}

func newMockReadCloser(s string) io.ReadCloser {
	return mockReadCloser{strings.NewReader(s)}
}

// JSONEventHandler Tests

func TestJSONEventHandler_SupportsRequest(t *testing.T) {
	handler := NewJSONEventHandler(
		&mockEventService{},
		&mockRenderService{},
		&mockResponseBuilder{},
		&mockEventValidator{},
	)

	tests := []struct {
		name     string
		request  *firHttp.RequestModel
		expected bool
	}{
		{
			name: "supports JSON POST request with X-FIR-MODE event header",
			request: &firHttp.RequestModel{
				Method: http.MethodPost,
				Header: http.Header{
					"Content-Type": []string{"application/json"},
					"X-Fir-Mode":   []string{"event"},
				},
				Body: newMockReadCloser(`{"event": "test"}`),
			},
			expected: true,
		},
		{
			name: "supports JSON POST with charset and X-FIR-MODE event header",
			request: &firHttp.RequestModel{
				Method: http.MethodPost,
				Header: http.Header{
					"Content-Type": []string{"application/json; charset=utf-8"},
					"X-Fir-Mode":   []string{"event"},
				},
				Body: newMockReadCloser(`{"event": "test"}`),
			},
			expected: true,
		},
		{
			name: "does not support GET request",
			request: &firHttp.RequestModel{
				Method: http.MethodGet,
				Header: http.Header{
					"Content-Type": []string{"application/json"},
					"X-Fir-Mode":   []string{"event"},
				},
			},
			expected: false,
		},
		{
			name: "does not support non-JSON content type",
			request: &firHttp.RequestModel{
				Method: http.MethodPost,
				Header: http.Header{
					"Content-Type": []string{"text/html"},
					"X-Fir-Mode":   []string{"event"},
				},
			},
			expected: false,
		},
		{
			name: "does not support missing content type",
			request: &firHttp.RequestModel{
				Method: http.MethodPost,
				Header: http.Header{
					"X-Fir-Mode": []string{"event"},
				},
			},
			expected: false,
		},
		{
			name: "does not support POST without X-FIR-MODE header",
			request: &firHttp.RequestModel{
				Method: http.MethodPost,
				Header: http.Header{
					"Content-Type": []string{"application/json"},
				},
				Body: newMockReadCloser(`{"event": "test"}`),
			},
			expected: false,
		},
		{
			name: "does not support POST with wrong X-FIR-MODE value",
			request: &firHttp.RequestModel{
				Method: http.MethodPost,
				Header: http.Header{
					"Content-Type": []string{"application/json"},
					"X-Fir-Mode":   []string{"other"},
				},
				Body: newMockReadCloser(`{"event": "test"}`),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.SupportsRequest(tt.request)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestJSONEventHandler_HandlerName(t *testing.T) {
	handler := NewJSONEventHandler(
		&mockEventService{},
		&mockRenderService{},
		&mockResponseBuilder{},
		&mockEventValidator{},
	)

	name := handler.HandlerName()
	if name != "json-event-handler" {
		t.Errorf("expected 'json-event-handler', got '%s'", name)
	}
}

// GetHandler Tests

func TestGetHandler_SupportsRequest(t *testing.T) {
	handler := NewGetHandler(
		&mockRenderService{},
		&mockTemplateService{},
		&mockResponseBuilder{},
		&mockEventService{}, // Added for onLoad support
	)

	tests := []struct {
		name     string
		request  *firHttp.RequestModel
		expected bool
	}{
		{
			name: "supports GET request",
			request: &firHttp.RequestModel{
				Method: http.MethodGet,
				URL:    createTestURL("/test"),
			},
			expected: true,
		},
		{
			name: "supports HEAD request",
			request: &firHttp.RequestModel{
				Method: http.MethodHead,
				URL:    createTestURL("/test"),
			},
			expected: true,
		},
		{
			name: "does not support POST request",
			request: &firHttp.RequestModel{
				Method: http.MethodPost,
				URL:    createTestURL("/test"),
			},
			expected: false,
		},
		{
			name: "does not support WebSocket upgrade",
			request: &firHttp.RequestModel{
				Method: http.MethodGet,
				URL:    createTestURL("/test"),
				Header: http.Header{
					"Upgrade":    []string{"websocket"},
					"Connection": []string{"upgrade"},
				},
			},
			expected: false,
		},
		{
			name: "does not support API paths",
			request: &firHttp.RequestModel{
				Method: http.MethodGet,
				URL:    createTestURL("/api/test"),
			},
			expected: false,
		},
		{
			name: "does not support static assets",
			request: &firHttp.RequestModel{
				Method: http.MethodGet,
				URL:    createTestURL("/static/style.css"),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.SupportsRequest(tt.request)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetHandler_Handle(t *testing.T) {
	mockRender := &mockRenderService{
		renderTemplateResponse: &services.RenderResult{
			HTML: []byte("<html>test</html>"),
		},
	}
	mockBuilder := &mockResponseBuilder{
		buildTemplateResponse: &firHttp.ResponseModel{
			StatusCode: 200,
			Headers:    map[string]string{"Content-Type": "text/html"},
			Body:       []byte("<html>test</html>"),
		},
	}

	handler := NewGetHandler(
		mockRender,
		&mockTemplateService{},
		mockBuilder,
		&mockEventService{}, // Added for onLoad support
	)

	request := &firHttp.RequestModel{
		Method: http.MethodGet,
		URL:    createTestURL("/test"),
	}

	resp, err := handler.Handle(context.Background(), request)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestGetHandler_HandlerName(t *testing.T) {
	handler := NewGetHandler(
		&mockRenderService{},
		&mockTemplateService{},
		&mockResponseBuilder{},
		&mockEventService{}, // Added for onLoad support
	)

	name := handler.HandlerName()
	if name != "get-handler" {
		t.Errorf("expected 'get-handler', got '%s'", name)
	}
}

// WebSocketHandler Tests

func TestWebSocketHandler_SupportsRequest(t *testing.T) {
	handler := NewWebSocketHandler(
		&mockEventService{},
		&mockResponseBuilder{},
	)

	tests := []struct {
		name     string
		request  *firHttp.RequestModel
		expected bool
	}{
		{
			name: "supports WebSocket upgrade request",
			request: &firHttp.RequestModel{
				Method: http.MethodGet,
				Header: http.Header{
					"Upgrade":    []string{"websocket"},
					"Connection": []string{"upgrade"},
				},
			},
			expected: true,
		},
		{
			name: "supports WebSocket upgrade with mixed case",
			request: &firHttp.RequestModel{
				Method: http.MethodGet,
				Header: http.Header{
					"Upgrade":    []string{"WebSocket"},
					"Connection": []string{"Upgrade"},
				},
			},
			expected: true,
		},
		{
			name: "does not support non-GET request",
			request: &firHttp.RequestModel{
				Method: http.MethodPost,
				Header: http.Header{
					"Upgrade":    []string{"websocket"},
					"Connection": []string{"upgrade"},
				},
			},
			expected: false,
		},
		{
			name: "does not support missing upgrade header",
			request: &firHttp.RequestModel{
				Method: http.MethodGet,
				Header: http.Header{
					"Connection": []string{"upgrade"},
				},
			},
			expected: false,
		},
		{
			name: "does not support regular GET request",
			request: &firHttp.RequestModel{
				Method: http.MethodGet,
				Header: http.Header{},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.SupportsRequest(tt.request)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestWebSocketHandler_HandlerName(t *testing.T) {
	handler := NewWebSocketHandler(
		&mockEventService{},
		&mockResponseBuilder{},
	)

	name := handler.HandlerName()
	if name != "websocket-handler" {
		t.Errorf("expected 'websocket-handler', got '%s'", name)
	}
}

// FormHandler Tests

func TestFormHandler_SupportsRequest(t *testing.T) {
	handler := NewFormHandler(
		&mockEventService{},
		&mockRenderService{},
		&mockResponseBuilder{},
		&mockEventValidator{},
	)

	tests := []struct {
		name     string
		request  *firHttp.RequestModel
		expected bool
	}{
		{
			name: "supports form POST request",
			request: &firHttp.RequestModel{
				Method: http.MethodPost,
				Header: http.Header{
					"Content-Type": []string{"application/x-www-form-urlencoded"},
				},
			},
			expected: true,
		},
		{
			name: "supports multipart form request",
			request: &firHttp.RequestModel{
				Method: http.MethodPost,
				Header: http.Header{
					"Content-Type": []string{"multipart/form-data; boundary=----WebKitFormBoundary"},
				},
			},
			expected: true,
		},
		{
			name: "does not support GET request",
			request: &firHttp.RequestModel{
				Method: http.MethodGet,
				Header: http.Header{
					"Content-Type": []string{"application/x-www-form-urlencoded"},
				},
			},
			expected: false,
		},
		{
			name: "does not support JSON content type",
			request: &firHttp.RequestModel{
				Method: http.MethodPost,
				Header: http.Header{
					"Content-Type": []string{"application/json"},
				},
			},
			expected: false,
		},
		{
			name: "does not support form POST with X-FIR-MODE event header (should be handled by JSON event handler)",
			request: &firHttp.RequestModel{
				Method: http.MethodPost,
				Header: http.Header{
					"Content-Type": []string{"application/x-www-form-urlencoded"},
					"X-Fir-Mode":   []string{"event"},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.SupportsRequest(tt.request)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFormHandler_HandlerName(t *testing.T) {
	handler := NewFormHandler(
		&mockEventService{},
		&mockRenderService{},
		&mockResponseBuilder{},
		&mockEventValidator{},
	)

	name := handler.HandlerName()
	if name != "form-handler" {
		t.Errorf("expected 'form-handler', got '%s'", name)
	}
}
