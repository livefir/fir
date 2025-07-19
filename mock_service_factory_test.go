package fir

import (
	"context"
	"html/template"
	"net/http"

	firHttp "github.com/livefir/fir/internal/http"
	"github.com/livefir/fir/internal/routeservices"
	"github.com/livefir/fir/internal/services"
	"github.com/livefir/fir/pubsub"
)

// MockServiceFactory provides centralized mock service creation for testing
// Implements Step 3.2 of the migration guide: Mock Service Creation
type MockServiceFactory struct{}

// NewMockServiceFactory creates a new mock service factory
func NewMockServiceFactory() *MockServiceFactory {
	return &MockServiceFactory{}
}

// CreateTestRouteServices creates a complete RouteServices setup for testing
func (f *MockServiceFactory) CreateTestRouteServices() *routeservices.RouteServices {
	return &routeservices.RouteServices{
		EventService:    f.CreateMockEventService(),
		RenderService:   f.CreateMockRenderService(),
		TemplateService: f.CreateMockTemplateService(),
		ResponseBuilder: f.CreateMockResponseBuilder(),
		Options: &routeservices.Options{
			DisableTemplateCache: false,
			DisableWebsocket:     false,
		},
	}
}

// CreateMockEventService creates a mock event service with default behaviors
func (f *MockServiceFactory) CreateMockEventService() services.EventService {
	return &TestMockEventService{
		processFunc: func(ctx context.Context, req services.EventRequest) (*services.EventResponse, error) {
			return &services.EventResponse{
				StatusCode:   http.StatusOK,
				Headers:      make(map[string]string),
				Body:         []byte("mock response"),
				Events:       []firHttp.DOMEvent{},
				PubSubEvents: []pubsub.Event{},
			}, nil
		},
	}
}

// CreateMockRenderService creates a mock render service
func (f *MockServiceFactory) CreateMockRenderService() services.RenderService {
	return &TestMockRenderService{}
}

// CreateMockTemplateService creates a mock template service
func (f *MockServiceFactory) CreateMockTemplateService() services.TemplateService {
	return &TestMockTemplateService{}
}

// CreateMockResponseBuilder creates a mock response builder
func (f *MockServiceFactory) CreateMockResponseBuilder() services.ResponseBuilder {
	return &TestMockResponseBuilder{}
}

// CreateMockEventServiceWithBehavior creates a mock event service with custom behavior
func (f *MockServiceFactory) CreateMockEventServiceWithBehavior(
	processFunc func(ctx context.Context, req services.EventRequest) (*services.EventResponse, error),
) services.EventService {
	return &TestMockEventService{
		processFunc: processFunc,
	}
}

// TestMockEventService implements services.EventService for testing
type TestMockEventService struct {
	processFunc func(ctx context.Context, req services.EventRequest) (*services.EventResponse, error)
	metrics     services.EventMetrics
}

func (m *TestMockEventService) ProcessEvent(ctx context.Context, req services.EventRequest) (*services.EventResponse, error) {
	if m.processFunc != nil {
		return m.processFunc(ctx, req)
	}
	return &services.EventResponse{
		StatusCode:   http.StatusOK,
		Headers:      make(map[string]string),
		Body:         []byte("default mock response"),
		Events:       []firHttp.DOMEvent{},
		PubSubEvents: []pubsub.Event{},
	}, nil
}

func (m *TestMockEventService) RegisterHandler(eventID string, handler services.EventHandler) error {
	return nil
}

func (m *TestMockEventService) GetEventMetrics() services.EventMetrics {
	return m.metrics
}

// TestMockRenderService implements services.RenderService for testing
type TestMockRenderService struct{}

func (m *TestMockRenderService) RenderTemplate(ctx services.RenderContext) (*services.RenderResult, error) {
	return &services.RenderResult{
		HTML:   []byte("<div>mock render</div>"),
		Events: []firHttp.DOMEvent{},
	}, nil
}

func (m *TestMockRenderService) RenderError(ctx services.ErrorContext) (*services.RenderResult, error) {
	return &services.RenderResult{
		HTML:   []byte("<div>mock error</div>"),
		Events: []firHttp.DOMEvent{},
	}, nil
}

func (m *TestMockRenderService) RenderEvents(events []pubsub.Event, routeID string) ([]firHttp.DOMEvent, error) {
	return []firHttp.DOMEvent{}, nil
}

// TestMockTemplateService implements services.TemplateService for testing
type TestMockTemplateService struct{}

func (m *TestMockTemplateService) LoadTemplate(config services.TemplateConfig) (*template.Template, error) {
	return template.New("mock"), nil
}

func (m *TestMockTemplateService) ParseTemplate(content, layout string, partials []string, funcMap template.FuncMap) (*template.Template, error) {
	return template.New("mock"), nil
}

func (m *TestMockTemplateService) GetTemplate(routeID string, templateType services.TemplateType) (*template.Template, error) {
	return template.New("mock"), nil
}

func (m *TestMockTemplateService) ClearCache() error {
	return nil
}

func (m *TestMockTemplateService) SetCacheEnabled(enabled bool) {}

// TestMockResponseBuilder implements services.ResponseBuilder for testing
type TestMockResponseBuilder struct{}

func (m *TestMockResponseBuilder) BuildEventResponse(result *services.EventResponse, request *firHttp.RequestModel) (*firHttp.ResponseModel, error) {
	return &firHttp.ResponseModel{
		StatusCode: http.StatusOK,
		Headers:    make(map[string]string),
		Body:       []byte("mock event response"),
	}, nil
}

func (m *TestMockResponseBuilder) BuildTemplateResponse(render *services.RenderResult, statusCode int) (*firHttp.ResponseModel, error) {
	return &firHttp.ResponseModel{
		StatusCode: statusCode,
		Headers:    make(map[string]string),
		Body:       render.HTML,
	}, nil
}

func (m *TestMockResponseBuilder) BuildErrorResponse(err error, statusCode int) (*firHttp.ResponseModel, error) {
	return &firHttp.ResponseModel{
		StatusCode: statusCode,
		Headers:    make(map[string]string),
		Body:       []byte("mock error response"),
	}, nil
}

func (m *TestMockResponseBuilder) BuildRedirectResponse(url string, statusCode int) (*firHttp.ResponseModel, error) {
	return &firHttp.ResponseModel{
		StatusCode: statusCode,
		Headers:    map[string]string{"Location": url},
		Body:       []byte{},
	}, nil
}
