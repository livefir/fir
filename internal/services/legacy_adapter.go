package services

import (
	"context"
	"fmt"
	"html/template"
	"net/http"

	firHttp "github.com/livefir/fir/internal/http"
	"github.com/livefir/fir/pubsub"
)

// LegacyRenderAdapter adapts the new render services to work with existing route and renderer interfaces
type LegacyRenderAdapter struct {
	renderService   RenderService
	templateService TemplateService
	responseBuilder ResponseBuilder
	legacyRenderer  interface{} // Legacy renderer interface
}

// NewLegacyRenderAdapter creates a new adapter for integrating new render services with legacy code
func NewLegacyRenderAdapter(renderService RenderService, templateService TemplateService, responseBuilder ResponseBuilder, legacyRenderer interface{}) *LegacyRenderAdapter {
	return &LegacyRenderAdapter{
		renderService:   renderService,
		templateService: templateService,
		responseBuilder: responseBuilder,
		legacyRenderer:  legacyRenderer,
	}
}

// RenderRouteTemplateWithNewServices renders a route template using the new service layer
func (a *LegacyRenderAdapter) RenderRouteTemplateWithNewServices(
	ctx context.Context,
	routeID string,
	templatePath string,
	layoutPath string,
	partialPaths []string,
	data interface{},
	funcMap template.FuncMap,
	requestModel *firHttp.RequestModel,
) (*RenderResult, error) {

	renderCtx := RenderContext{
		RouteID:      routeID,
		TemplateType: StandardTemplate,
		Data:         data,
		TemplatePath: templatePath,
		LayoutPath:   layoutPath,
		PartialPaths: partialPaths,
		FuncMap:      funcMap,
		RequestModel: requestModel,
		Context:      ctx,
	}

	return a.renderService.RenderTemplate(renderCtx)
}

// RenderErrorTemplateWithNewServices renders an error template using the new service layer
func (a *LegacyRenderAdapter) RenderErrorTemplateWithNewServices(
	ctx context.Context,
	routeID string,
	templatePath string,
	layoutPath string,
	err error,
	statusCode int,
	errorData map[string]interface{},
	funcMap template.FuncMap,
	requestModel *firHttp.RequestModel,
) (*RenderResult, error) {

	errorCtx := ErrorContext{
		Error:      err,
		StatusCode: statusCode,
		ErrorData:  errorData,
		RenderContext: RenderContext{
			RouteID:      routeID,
			TemplateType: ErrorTemplate,
			TemplatePath: templatePath,
			LayoutPath:   layoutPath,
			FuncMap:      funcMap,
			RequestModel: requestModel,
			Context:      ctx,
		},
	}

	return a.renderService.RenderError(errorCtx)
}

// ConvertPubSubEventsToDOM converts PubSub events to DOM events using the new service layer
func (a *LegacyRenderAdapter) ConvertPubSubEventsToDOM(events []pubsub.Event, routeID string) ([]firHttp.DOMEvent, error) {
	return a.renderService.RenderEvents(events, routeID)
}

// BuildHTTPResponseFromRenderResult creates an HTTP response from a render result
func (a *LegacyRenderAdapter) BuildHTTPResponseFromRenderResult(result *RenderResult, statusCode int) (*firHttp.ResponseModel, error) {
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	return a.responseBuilder.BuildTemplateResponse(result, statusCode)
}

// BuildHTTPResponseFromEventResult creates an HTTP response from an event processing result
func (a *LegacyRenderAdapter) BuildHTTPResponseFromEventResult(result *EventResponse, requestModel *firHttp.RequestModel) (*firHttp.ResponseModel, error) {
	return a.responseBuilder.BuildEventResponse(result, requestModel)
}

// BuildHTTPErrorResponse creates an HTTP error response
func (a *LegacyRenderAdapter) BuildHTTPErrorResponse(err error, statusCode int) (*firHttp.ResponseModel, error) {
	return a.responseBuilder.BuildErrorResponse(err, statusCode)
}

// BuildHTTPRedirectResponse creates an HTTP redirect response
func (a *LegacyRenderAdapter) BuildHTTPRedirectResponse(url string, statusCode int) (*firHttp.ResponseModel, error) {
	return a.responseBuilder.BuildRedirectResponse(url, statusCode)
}

// LoadTemplateWithNewService loads a template using the new template service
func (a *LegacyRenderAdapter) LoadTemplateWithNewService(config TemplateConfig) (*template.Template, error) {
	return a.templateService.LoadTemplate(config)
}

// ParseTemplateWithNewService parses template content using the new template service
func (a *LegacyRenderAdapter) ParseTemplateWithNewService(content, layout string, partials []string, funcMap template.FuncMap) (*template.Template, error) {
	return a.templateService.ParseTemplate(content, layout, partials, funcMap)
}

// GetLegacyRenderer returns the legacy renderer for fallback scenarios
func (a *LegacyRenderAdapter) GetLegacyRenderer() interface{} {
	return a.legacyRenderer
}

// HasNewServices returns true if the adapter has the new render services configured
func (a *LegacyRenderAdapter) HasNewServices() bool {
	return a.renderService != nil && a.templateService != nil && a.responseBuilder != nil
}

// ValidateServices validates that all required services are properly configured
func (a *LegacyRenderAdapter) ValidateServices() error {
	if a.renderService == nil {
		return fmt.Errorf("renderService is required but not set")
	}
	if a.templateService == nil {
		return fmt.Errorf("templateService is required but not set")
	}
	if a.responseBuilder == nil {
		return fmt.Errorf("responseBuilder is required but not set")
	}
	return nil
}

// TemplateConfigFromRouteData creates a TemplateConfig from route data
func (a *LegacyRenderAdapter) TemplateConfigFromRouteData(
	routeID string,
	templatePath string,
	layoutPath string,
	partialPaths []string,
	funcMap template.FuncMap,
	cacheDisabled bool,
) TemplateConfig {
	return TemplateConfig{
		ContentPath:   templatePath,
		LayoutPath:    layoutPath,
		PartialPaths:  partialPaths,
		FuncMap:       funcMap,
		CacheDisabled: cacheDisabled,
		RouteID:       routeID,
	}
}
