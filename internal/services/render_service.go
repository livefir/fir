package services

import (
	"bytes"
	"fmt"
	"time"

	firHttp "github.com/livefir/fir/internal/http"
	"github.com/livefir/fir/pubsub"
)

// DefaultRenderService is the default implementation of RenderService
type DefaultRenderService struct {
	templateService TemplateService
	templateEngine  TemplateEngine
	responseBuilder ResponseBuilder
}

// NewDefaultRenderService creates a new default render service
func NewDefaultRenderService(templateService TemplateService, templateEngine TemplateEngine, responseBuilder ResponseBuilder) *DefaultRenderService {
	return &DefaultRenderService{
		templateService: templateService,
		templateEngine:  templateEngine,
		responseBuilder: responseBuilder,
	}
}

// RenderTemplate renders a template with the given context
func (s *DefaultRenderService) RenderTemplate(ctx RenderContext) (*RenderResult, error) {
	startTime := time.Now()

	// Build template configuration from context
	config := TemplateConfig{
		ContentPath:       ctx.TemplatePath,
		LayoutPath:        ctx.LayoutPath,
		PartialPaths:      ctx.PartialPaths,
		Extensions:        ctx.Extensions,
		FuncMap:           ctx.FuncMap,
		CacheDisabled:     ctx.CacheDisabled,
		RouteID:           ctx.RouteID,
	}

	// Load template
	tmpl, err := s.templateService.LoadTemplate(config)
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}

	// Execute template
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, ctx.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	result := &RenderResult{
		HTML:           buf.Bytes(),
		TemplateUsed:   ctx.TemplatePath,
		RenderDuration: time.Since(startTime).Milliseconds(),
		CacheHit:       false, // TODO: determine cache hit from template service
	}

	return result, nil
}

// RenderError renders an error template
func (s *DefaultRenderService) RenderError(ctx ErrorContext) (*RenderResult, error) {
	startTime := time.Now()

	// Prepare error data for template
	errorData := map[string]interface{}{
		"Error":      ctx.Error.Error(),
		"StatusCode": ctx.StatusCode,
	}

	// Merge with additional error data
	for k, v := range ctx.ErrorData {
		errorData[k] = v
	}

	// Build template configuration for error template
	config := TemplateConfig{
		ContentPath:       ctx.TemplatePath,
		LayoutPath:        ctx.LayoutPath,
		PartialPaths:      ctx.PartialPaths,
		Extensions:        ctx.Extensions,
		FuncMap:           ctx.FuncMap,
		CacheDisabled:     ctx.CacheDisabled,
		RouteID:           ctx.RouteID,
	}

	// Load error template
	tmpl, err := s.templateService.LoadTemplate(config)
	if err != nil {
		// Fallback to simple error rendering if template fails
		return s.renderSimpleError(ctx.Error, ctx.StatusCode), nil
	}

	// Execute error template
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, errorData)
	if err != nil {
		// Fallback to simple error rendering if execution fails
		return s.renderSimpleError(ctx.Error, ctx.StatusCode), nil
	}

	result := &RenderResult{
		HTML:           buf.Bytes(),
		TemplateUsed:   ctx.TemplatePath,
		RenderDuration: time.Since(startTime).Milliseconds(),
		CacheHit:       false,
	}

	return result, nil
}

// RenderEvents converts PubSub events to DOM events
func (s *DefaultRenderService) RenderEvents(events []pubsub.Event, routeID string) ([]firHttp.DOMEvent, error) {
	if len(events) == 0 {
		return nil, nil
	}

	domEvents := make([]firHttp.DOMEvent, 0, len(events))

	for _, event := range events {
		domEvent, err := s.convertPubSubEventToDOMEvent(event, routeID)
		if err != nil {
			target := "unknown"
			if event.Target != nil {
				target = *event.Target
			}
			return nil, fmt.Errorf("failed to convert event with target %s: %w", target, err)
		}
		
		if domEvent != nil {
			domEvents = append(domEvents, *domEvent)
		}
	}

	return domEvents, nil
}

// renderSimpleError renders a simple error message when template rendering fails
func (s *DefaultRenderService) renderSimpleError(err error, statusCode int) *RenderResult {
	errorHTML := fmt.Sprintf(`
<div style="padding: 20px; border: 1px solid #ff0000; background-color: #ffe6e6; color: #cc0000;">
	<h3>Error %d</h3>
	<p>%s</p>
</div>`, statusCode, err.Error())

	return &RenderResult{
		HTML:           []byte(errorHTML),
		TemplateUsed:   "internal-error",
		RenderDuration: 0,
		CacheHit:       false,
	}
}

// convertPubSubEventToDOMEvent converts a PubSub event to a DOM event
func (s *DefaultRenderService) convertPubSubEventToDOMEvent(event pubsub.Event, routeID string) (*firHttp.DOMEvent, error) {
	// Convert based on the event state
	var eventType string
	switch event.State {
	case "ok":
		eventType = "update"
	case "error":
		eventType = "error"
	case "pending":
		eventType = "pending"
	case "done":
		eventType = "update"
	default:
		eventType = "update" // Default to update
	}

	// Extract target selector
	target := ""
	if event.Target != nil {
		target = *event.Target
	}

	// Extract element key
	elementKey := ""
	if event.ElementKey != nil {
		elementKey = *event.ElementKey
	}

	// Build data map from detail
	data := make(map[string]interface{})
	if event.Detail != nil {
		// Convert detail to data map (this depends on your dom.Detail structure)
		data["detail"] = event.Detail
	}

	domEvent := &firHttp.DOMEvent{
		Type:       eventType,
		Target:     target,
		ElementKey: elementKey,
		Data:       data,
	}

	// Add ID if available
	if event.ID != nil {
		domEvent.ID = *event.ID
	}

	return domEvent, nil
}
