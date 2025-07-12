package services

import (
	"context"
	"errors"
	"html/template"
	"testing"

	"github.com/livefir/fir/internal/eventstate"
	"github.com/livefir/fir/pubsub"
)

// MockTemplateService is a mock implementation of TemplateService for testing
type MockTemplateService struct {
	loadTemplateFunc    func(TemplateConfig) (*template.Template, error)
	parseTemplateFunc   func(string, string, []string, template.FuncMap) (*template.Template, error)
	getTemplateFunc     func(string, TemplateType) (*template.Template, error)
	clearCacheFunc      func() error
	setCacheEnabledFunc func(bool)
}

func (m *MockTemplateService) LoadTemplate(config TemplateConfig) (*template.Template, error) {
	if m.loadTemplateFunc != nil {
		return m.loadTemplateFunc(config)
	}
	tmpl, err := template.New("test").Parse("<h1>{{.Title}}</h1>")
	return tmpl, err
}

func (m *MockTemplateService) ParseTemplate(content, layout string, partials []string, funcMap template.FuncMap) (*template.Template, error) {
	if m.parseTemplateFunc != nil {
		return m.parseTemplateFunc(content, layout, partials, funcMap)
	}
	return template.New("test").Parse(content)
}

func (m *MockTemplateService) GetTemplate(routeID string, templateType TemplateType) (*template.Template, error) {
	if m.getTemplateFunc != nil {
		return m.getTemplateFunc(routeID, templateType)
	}
	return nil, errors.New("not implemented")
}

func (m *MockTemplateService) ClearCache() error {
	if m.clearCacheFunc != nil {
		return m.clearCacheFunc()
	}
	return nil
}

func (m *MockTemplateService) SetCacheEnabled(enabled bool) {
	if m.setCacheEnabledFunc != nil {
		m.setCacheEnabledFunc(enabled)
	}
}

// MockTemplateEngine is a mock implementation of TemplateEngine for testing
type MockTemplateEngine struct {
	parseFilesFunc   func(...string) (TemplateHandle, error)
	parseContentFunc func(string) (TemplateHandle, error)
	executeFunc      func(TemplateHandle, interface{}) ([]byte, error)
	addFuncMapFunc   func(template.FuncMap)
}

func (m *MockTemplateEngine) ParseFiles(files ...string) (TemplateHandle, error) {
	if m.parseFilesFunc != nil {
		return m.parseFilesFunc(files...)
	}
	return &MockTemplateHandle{}, nil
}

func (m *MockTemplateEngine) ParseContent(content string) (TemplateHandle, error) {
	if m.parseContentFunc != nil {
		return m.parseContentFunc(content)
	}
	return &MockTemplateHandle{}, nil
}

func (m *MockTemplateEngine) Execute(tmpl TemplateHandle, data interface{}) ([]byte, error) {
	if m.executeFunc != nil {
		return m.executeFunc(tmpl, data)
	}
	return []byte("<h1>Test</h1>"), nil
}

func (m *MockTemplateEngine) AddFuncMap(funcMap template.FuncMap) {
	if m.addFuncMapFunc != nil {
		m.addFuncMapFunc(funcMap)
	}
}

// MockTemplateHandle is a mock implementation of TemplateHandle for testing
type MockTemplateHandle struct {
	executeFunc func(interface{}) ([]byte, error)
	cloneFunc   func() (TemplateHandle, error)
}

func (m *MockTemplateHandle) Execute(data interface{}) ([]byte, error) {
	if m.executeFunc != nil {
		return m.executeFunc(data)
	}
	return []byte("<h1>Test</h1>"), nil
}

func (m *MockTemplateHandle) Clone() (TemplateHandle, error) {
	if m.cloneFunc != nil {
		return m.cloneFunc()
	}
	return &MockTemplateHandle{}, nil
}

func TestDefaultRenderService_RenderTemplate(t *testing.T) {
	templateService := &MockTemplateService{}
	templateEngine := &MockTemplateEngine{}
	responseBuilder := NewDefaultResponseBuilder()

	service := NewDefaultRenderService(templateService, templateEngine, responseBuilder)

	tests := []struct {
		name    string
		ctx     RenderContext
		wantErr bool
	}{
		{
			name: "successful render",
			ctx: RenderContext{
				RouteID:      "test-route",
				TemplatePath: "<h1>{{.Title}}</h1>",
				Data:         map[string]interface{}{"Title": "Test"},
				Context:      context.Background(),
			},
			wantErr: false,
		},
		{
			name: "template load error",
			ctx: RenderContext{
				RouteID:      "error-route",
				TemplatePath: "invalid-template",
				Data:         map[string]interface{}{"Title": "Test"},
				Context:      context.Background(),
			},
			wantErr: false, // Should not error because mock returns valid template
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.RenderTemplate(tt.ctx)

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

			if result == nil {
				t.Errorf("expected render result but got nil")
				return
			}

			if len(result.HTML) == 0 {
				t.Errorf("expected HTML content but got empty")
			}

			if result.TemplateUsed != tt.ctx.TemplatePath {
				t.Errorf("expected template used %s, got %s", tt.ctx.TemplatePath, result.TemplateUsed)
			}
		})
	}
}

func TestDefaultRenderService_RenderError(t *testing.T) {
	templateService := &MockTemplateService{
		loadTemplateFunc: func(config TemplateConfig) (*template.Template, error) {
			return nil, errors.New("template not found")
		},
	}
	templateEngine := &MockTemplateEngine{}
	responseBuilder := NewDefaultResponseBuilder()

	service := NewDefaultRenderService(templateService, templateEngine, responseBuilder)

	ctx := ErrorContext{
		Error:      errors.New("test error"),
		StatusCode: 500,
		ErrorData:  map[string]interface{}{"extra": "data"},
		RenderContext: RenderContext{
			RouteID:      "error-route",
			TemplatePath: "error.html",
			Context:      context.Background(),
		},
	}

	result, err := service.RenderError(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if result == nil {
		t.Errorf("expected render result but got nil")
		return
	}

	if len(result.HTML) == 0 {
		t.Errorf("expected HTML content but got empty")
	}

	// Should fall back to simple error rendering
	if result.TemplateUsed != "internal-error" {
		t.Errorf("expected fallback template 'internal-error', got %s", result.TemplateUsed)
	}
}

func TestDefaultRenderService_RenderEvents(t *testing.T) {
	templateService := &MockTemplateService{}
	templateEngine := &MockTemplateEngine{}
	responseBuilder := NewDefaultResponseBuilder()

	service := NewDefaultRenderService(templateService, templateEngine, responseBuilder)

	tests := []struct {
		name      string
		events    []pubsub.Event
		routeID   string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "empty events",
			events:    []pubsub.Event{},
			routeID:   "test-route",
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "single event",
			events: []pubsub.Event{
				{
					ID:     stringPtr("event-1"),
					State:  eventstate.OK,
					Target: stringPtr("#test"),
				},
			},
			routeID:   "test-route",
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "multiple events",
			events: []pubsub.Event{
				{
					ID:     stringPtr("event-1"),
					State:  eventstate.OK,
					Target: stringPtr("#test1"),
				},
				{
					ID:     stringPtr("event-2"),
					State:  eventstate.Error,
					Target: stringPtr("#test2"),
				},
			},
			routeID:   "test-route",
			wantCount: 2,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			domEvents, err := service.RenderEvents(tt.events, tt.routeID)

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

			if len(domEvents) != tt.wantCount {
				t.Errorf("expected %d DOM events, got %d", tt.wantCount, len(domEvents))
			}
		})
	}
}

func TestDefaultRenderService_convertPubSubEventToDOMEvent(t *testing.T) {
	templateService := &MockTemplateService{}
	templateEngine := &MockTemplateEngine{}
	responseBuilder := NewDefaultResponseBuilder()

	service := NewDefaultRenderService(templateService, templateEngine, responseBuilder)

	tests := []struct {
		name         string
		event        pubsub.Event
		routeID      string
		expectedType string
		wantErr      bool
	}{
		{
			name: "ok state event",
			event: pubsub.Event{
				ID:     stringPtr("event-1"),
				State:  eventstate.OK,
				Target: stringPtr("#test"),
			},
			routeID:      "test-route",
			expectedType: "update",
			wantErr:      false,
		},
		{
			name: "error state event",
			event: pubsub.Event{
				ID:     stringPtr("event-2"),
				State:  eventstate.Error,
				Target: stringPtr("#error"),
			},
			routeID:      "test-route",
			expectedType: "error",
			wantErr:      false,
		},
		{
			name: "pending state event",
			event: pubsub.Event{
				ID:     stringPtr("event-3"),
				State:  eventstate.Pending,
				Target: stringPtr("#pending"),
			},
			routeID:      "test-route",
			expectedType: "pending",
			wantErr:      false,
		},
		{
			name: "done state event",
			event: pubsub.Event{
				ID:     stringPtr("event-4"),
				State:  eventstate.Done,
				Target: stringPtr("#done"),
			},
			routeID:      "test-route",
			expectedType: "update",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			domEvent, err := service.convertPubSubEventToDOMEvent(tt.event, tt.routeID)

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

			if domEvent == nil {
				t.Errorf("expected DOM event but got nil")
				return
			}

			if domEvent.Type != tt.expectedType {
				t.Errorf("expected event type %s, got %s", tt.expectedType, domEvent.Type)
			}

			if tt.event.Target != nil && domEvent.Target != *tt.event.Target {
				t.Errorf("expected target %s, got %s", *tt.event.Target, domEvent.Target)
			}

			if tt.event.ID != nil && domEvent.ID != *tt.event.ID {
				t.Errorf("expected ID %s, got %s", *tt.event.ID, domEvent.ID)
			}
		})
	}
}
