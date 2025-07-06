package handlers

import (
	"context"
	"net/url"
	"testing"

	firHttp "github.com/livefir/fir/internal/http"
)

// mockLogger implements HandlerLogger for testing
type mockLogger struct{}

func (m *mockLogger) LogRequest(handlerName string, req *firHttp.RequestModel)                       {}
func (m *mockLogger) LogResponse(handlerName string, resp *firHttp.ResponseModel, duration int64)   {}
func (m *mockLogger) LogError(handlerName string, err error, req *firHttp.RequestModel)             {}
func (m *mockLogger) LogHandlerSelection(selectedHandler string, req *firHttp.RequestModel)         {}

// mockMetrics implements HandlerMetrics for testing
type mockMetrics struct{}

func (m *mockMetrics) RecordRequest(handlerName string, method string)                {}
func (m *mockMetrics) RecordResponse(handlerName string, statusCode int, duration int64) {}
func (m *mockMetrics) RecordError(handlerName string, err error)                      {}

// Helper function to create test URL
func createTestURL(path string) *url.URL {
	u, _ := url.Parse("http://example.com" + path)
	return u
}

// mockHandler is a test helper that implements RequestHandler
type mockHandler struct {
	name             string
	supportsRequest  bool
	handleResponse   *firHttp.ResponseModel
	handleError      error
	priority         int
	handledRequests  []*firHttp.RequestModel
}

func newMockHandler(name string, priority int) *mockHandler {
	return &mockHandler{
		name:     name,
		priority: priority,
		supportsRequest: true,
		handleResponse: &firHttp.ResponseModel{
			StatusCode: 200,
			Headers:    map[string]string{"Content-Type": "text/plain"},
			Body:       []byte("mock response from " + name),
		},
	}
}

func (m *mockHandler) Handle(ctx context.Context, req *firHttp.RequestModel) (*firHttp.ResponseModel, error) {
	m.handledRequests = append(m.handledRequests, req)
	return m.handleResponse, m.handleError
}

func (m *mockHandler) SupportsRequest(req *firHttp.RequestModel) bool {
	return m.supportsRequest
}

func (m *mockHandler) HandlerName() string {
	return m.name
}

func (m *mockHandler) setSupportsRequest(supports bool) *mockHandler {
	m.supportsRequest = supports
	return m
}

func TestDefaultHandlerChain_Handle(t *testing.T) {
	tests := []struct {
		name           string
		handlers       []RequestHandler
		request        *firHttp.RequestModel
		expectedStatus int
		expectedBody   string
		expectError    bool
	}{
		{
			name: "single handler supports request",
			handlers: []RequestHandler{
				newMockHandler("handler1", 10),
			},
			request: &firHttp.RequestModel{
				Method: "GET",
				URL:    createTestURL("/test"),
			},
			expectedStatus: 200,
			expectedBody:   "mock response from handler1",
			expectError:    false,
		},
		{
			name: "first handler doesn't support, second does",
			handlers: []RequestHandler{
				newMockHandler("handler1", 10).setSupportsRequest(false),
				newMockHandler("handler2", 20),
			},
			request: &firHttp.RequestModel{
				Method: "GET",
				URL:    createTestURL("/test"),
			},
			expectedStatus: 200,
			expectedBody:   "mock response from handler2",
			expectError:    false,
		},
		{
			name: "no handlers support request",
			handlers: []RequestHandler{
				newMockHandler("handler1", 10).setSupportsRequest(false),
				newMockHandler("handler2", 20).setSupportsRequest(false),
			},
			request: &firHttp.RequestModel{
				Method: "GET",
				URL:    createTestURL("/test"),
			},
			expectedStatus: 0,
			expectedBody:   "",
			expectError:    true,
		},
		{
			name:     "empty handler chain",
			handlers: []RequestHandler{},
			request: &firHttp.RequestModel{
				Method: "GET",
				URL:    createTestURL("/test"),
			},
			expectedStatus: 0,
			expectedBody:   "",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chain := NewDefaultHandlerChain(&mockLogger{}, &mockMetrics{})
			
			// Add handlers to chain
			for _, handler := range tt.handlers {
				chain.AddHandler(handler)
			}

			// Execute the chain
			resp, err := chain.Handle(context.Background(), tt.request)

			// Check error expectation
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Check response
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if string(resp.Body) != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, string(resp.Body))
			}
		})
	}
}

func TestDefaultHandlerChain_AddHandler(t *testing.T) {
	chain := NewDefaultHandlerChain(&mockLogger{}, &mockMetrics{})
	handler1 := newMockHandler("handler1", 10)
	handler2 := newMockHandler("handler2", 20)

	// Initially empty
	if len(chain.GetHandlers()) != 0 {
		t.Errorf("expected empty chain, got %d handlers", len(chain.GetHandlers()))
	}

	// Add first handler
	chain.AddHandler(handler1)
	handlers := chain.GetHandlers()
	if len(handlers) != 1 {
		t.Errorf("expected 1 handler, got %d", len(handlers))
	}
	if handlers[0].HandlerName() != "handler1" {
		t.Errorf("expected handler1, got %s", handlers[0].HandlerName())
	}

	// Add second handler
	chain.AddHandler(handler2)
	handlers = chain.GetHandlers()
	if len(handlers) != 2 {
		t.Errorf("expected 2 handlers, got %d", len(handlers))
	}
}

func TestDefaultHandlerChain_RemoveHandler(t *testing.T) {
	chain := NewDefaultHandlerChain(&mockLogger{}, &mockMetrics{})
	handler1 := newMockHandler("handler1", 10)
	handler2 := newMockHandler("handler2", 20)

	// Add handlers
	chain.AddHandler(handler1)
	chain.AddHandler(handler2)

	// Remove existing handler
	removed := chain.RemoveHandler("handler1")
	if !removed {
		t.Errorf("expected handler to be removed")
	}

	handlers := chain.GetHandlers()
	if len(handlers) != 1 {
		t.Errorf("expected 1 handler after removal, got %d", len(handlers))
	}
	if handlers[0].HandlerName() != "handler2" {
		t.Errorf("expected handler2 to remain, got %s", handlers[0].HandlerName())
	}

	// Try to remove non-existent handler
	removed = chain.RemoveHandler("nonexistent")
	if removed {
		t.Errorf("expected false when removing non-existent handler")
	}
}

func TestDefaultHandlerChain_ClearHandlers(t *testing.T) {
	chain := NewDefaultHandlerChain(&mockLogger{}, &mockMetrics{})
	chain.AddHandler(newMockHandler("handler1", 10))
	chain.AddHandler(newMockHandler("handler2", 20))

	// Clear all handlers
	chain.ClearHandlers()

	handlers := chain.GetHandlers()
	if len(handlers) != 0 {
		t.Errorf("expected empty chain after clear, got %d handlers", len(handlers))
	}
}

func TestPriorityHandlerChain_Handle(t *testing.T) {
	tests := []struct {
		name         string
		handlers     []RequestHandler
		expectedOrder []string
	}{
		{
			name: "handlers execute in priority order",
			handlers: []RequestHandler{
				newMockHandler("high-priority", 5),
				newMockHandler("medium-priority", 10),
				newMockHandler("low-priority", 20),
			},
			expectedOrder: []string{"high-priority", "medium-priority", "low-priority"},
		},
		{
			name: "same priority handlers execute in addition order",
			handlers: []RequestHandler{
				newMockHandler("first", 10),
				newMockHandler("second", 10),
				newMockHandler("third", 10),
			},
			expectedOrder: []string{"first", "second", "third"},
		},
		{
			name: "mixed priorities",
			handlers: []RequestHandler{
				newMockHandler("medium", 10),
				newMockHandler("high", 5),
				newMockHandler("low", 15),
				newMockHandler("highest", 1),
			},
			expectedOrder: []string{"highest", "high", "medium", "low"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chain := NewPriorityHandlerChain(&mockLogger{}, &mockMetrics{})
			
			// Add handlers to chain
			for _, handler := range tt.handlers {
				config := HandlerConfig{
					Name:     handler.HandlerName(),
					Priority: getMockHandlerPriority(handler),
					Enabled:  true, // Make sure handlers are enabled
				}
				chain.AddHandlerWithConfig(handler, config)
			}

			// Test that only the first handler processes the request
			request := &firHttp.RequestModel{
				Method: "GET",
				URL:    createTestURL("/test"),
			}

			resp, err := chain.Handle(context.Background(), request)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Should get response from the highest priority handler
			expectedBody := "mock response from " + tt.expectedOrder[0]
			if string(resp.Body) != expectedBody {
				t.Errorf("expected response from %s, got %q", tt.expectedOrder[0], string(resp.Body))
			}

			// Verify handlers are in correct order
			handlers := chain.GetHandlers()
			for i, expectedName := range tt.expectedOrder {
				if i >= len(handlers) {
					t.Errorf("missing handler at position %d", i)
					continue
				}
				if handlers[i].HandlerName() != expectedName {
					t.Errorf("handler at position %d: expected %s, got %s", 
						i, expectedName, handlers[i].HandlerName())
				}
			}
		})
	}
}

// Helper function to get priority from mock handler
func getMockHandlerPriority(handler RequestHandler) int {
	if mock, ok := handler.(*mockHandler); ok {
		return mock.priority
	}
	return 100 // default priority
}
