package handlers

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	firHttp "github.com/livefir/fir/internal/http"
	"github.com/livefir/fir/internal/services"
)

// GetHandler handles GET requests for pages and static content
type GetHandler struct {
	renderService   services.RenderService
	templateService services.TemplateService
	responseBuilder services.ResponseBuilder
	eventService    services.EventService // Added for onLoad support
	config          HandlerConfig
}

// NewGetHandler creates a new GET request handler
func NewGetHandler(
	renderService services.RenderService,
	templateService services.TemplateService,
	responseBuilder services.ResponseBuilder,
	eventService services.EventService, // Added for onLoad support
) *GetHandler {
	return &GetHandler{
		renderService:   renderService,
		templateService: templateService,
		responseBuilder: responseBuilder,
		eventService:    eventService, // Added for onLoad support
		config: HandlerConfig{
			Name:     "get-handler",
			Priority: 50, // Lower priority than event handlers
			Methods:  []string{http.MethodGet, http.MethodHead},
			Enabled:  true,
		},
	}
}

// Handle processes a GET request
func (h *GetHandler) Handle(ctx context.Context, req *firHttp.RequestModel) (*firHttp.ResponseModel, error) {
	// Extract the path from the request
	path := req.URL.Path
	if path == "" || path == "/" {
		path = "index"
	}

	// Clean the path to prevent directory traversal
	path = filepath.Clean(path)
	if strings.HasPrefix(path, "../") || strings.Contains(path, "/../") {
		return h.responseBuilder.BuildErrorResponse(
			fmt.Errorf("invalid path: %s", path),
			http.StatusBadRequest,
		)
	}

	// Remove leading slash for template resolution
	path = strings.TrimPrefix(path, "/")

	// If path is empty after cleaning, default to index
	if path == "" {
		path = "index"
	}

	// Try to find a template for this path
	templatePath := h.resolveTemplatePath(path)

	// Extract query parameters and form data for template context
	templateData := h.buildTemplateData(req)

	// Process onLoad event if EventService is available and route has onLoad handler
	if h.eventService != nil {
		// Extract route ID from the request (this would need to be passed in context or header)
		// For now, we'll use the path as a route identifier
		routeID := path
		if routeID == "" {
			routeID = "index"
		}

		// Try to process onLoad event
		err := h.processOnLoadEvent(ctx, req, routeID, templateData)
		if err != nil {
			// Log the error but don't fail the request
			// onLoad errors shouldn't prevent page rendering
			// TODO: Add proper logging
		}
	}

	// Build render context
	renderCtx := services.RenderContext{
		TemplatePath: templatePath,
		Data:         templateData,
		RequestModel: req,
		Context:      ctx,
		TemplateType: services.StandardTemplate,
	}

	// Render the template
	renderResp, err := h.renderService.RenderTemplate(renderCtx)
	if err != nil {
		return h.responseBuilder.BuildErrorResponse(
			fmt.Errorf("failed to render template: %w", err),
			http.StatusInternalServerError,
		)
	}

	// Build response using ResponseBuilder
	return h.responseBuilder.BuildTemplateResponse(renderResp, http.StatusOK)
}

// SupportsRequest determines if this handler can process the given request
func (h *GetHandler) SupportsRequest(req *firHttp.RequestModel) bool {
	// Check if this is a GET or HEAD request
	method := req.Method
	if method != http.MethodGet && method != http.MethodHead {
		return false
	}

	// Check if this is not a WebSocket upgrade request
	if h.isWebSocketUpgrade(req) {
		return false
	}

	// Check if this is not an API request (by convention, API paths start with /api/)
	if strings.HasPrefix(req.URL.Path, "/api/") {
		return false
	}

	// Check if this is not a static asset request (by extension)
	if h.isStaticAsset(req.URL.Path) {
		return false
	}

	return true
}

// HandlerName returns the name of this handler
func (h *GetHandler) HandlerName() string {
	return h.config.Name
}

// resolveTemplatePath resolves the template path from the request path
func (h *GetHandler) resolveTemplatePath(path string) string {
	// Add .html extension if not present
	if !strings.Contains(path, ".") {
		path += ".html"
	}
	return path
}

// buildTemplateData builds template data from request parameters
func (h *GetHandler) buildTemplateData(req *firHttp.RequestModel) map[string]interface{} {
	data := make(map[string]interface{})

	// Add query parameters
	if req.URL != nil && req.URL.RawQuery != "" {
		queryParams := make(map[string]string)
		for key, values := range req.URL.Query() {
			if len(values) > 0 {
				queryParams[key] = values[0] // Take first value
			}
		}
		data["query"] = queryParams
	}

	// Add path parameters if any (this would be enhanced with routing context)
	data["path"] = req.URL.Path

	// Add request metadata
	data["method"] = req.Method
	data["headers"] = req.Header

	return data
}

// isWebSocketUpgrade checks if this is a WebSocket upgrade request
func (h *GetHandler) isWebSocketUpgrade(req *firHttp.RequestModel) bool {
	upgrade := req.Header.Get("Upgrade")
	connection := req.Header.Get("Connection")

	return strings.ToLower(upgrade) == "websocket" &&
		strings.Contains(strings.ToLower(connection), "upgrade")
}

// isStaticAsset checks if this is a request for a static asset
func (h *GetHandler) isStaticAsset(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	staticExtensions := []string{
		".css", ".js", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico",
		".woff", ".woff2", ".ttf", ".eot", ".map", ".json", ".xml",
	}

	for _, staticExt := range staticExtensions {
		if ext == staticExt {
			return true
		}
	}

	return false
}

// ConfigurableGetHandler is a configurable version of GetHandler
type ConfigurableGetHandler struct {
	*GetHandler
	customConfig HandlerConfig
}

// NewConfigurableGetHandler creates a new configurable GET handler
func NewConfigurableGetHandler(
	renderService services.RenderService,
	templateService services.TemplateService,
	responseBuilder services.ResponseBuilder,
	eventService services.EventService, // Added for onLoad support
	config HandlerConfig,
) *ConfigurableGetHandler {
	baseHandler := NewGetHandler(renderService, templateService, responseBuilder, eventService)

	// Override base config with custom config
	if config.Name != "" {
		baseHandler.config.Name = config.Name
	}
	if config.Priority != 0 {
		baseHandler.config.Priority = config.Priority
	}
	if len(config.Methods) > 0 {
		baseHandler.config.Methods = config.Methods
	}
	if config.PathPattern != "" {
		baseHandler.config.PathPattern = config.PathPattern
	}
	baseHandler.config.Enabled = config.Enabled

	return &ConfigurableGetHandler{
		GetHandler:   baseHandler,
		customConfig: config,
	}
}

// SupportsRequest overrides the base implementation to use custom configuration
func (h *ConfigurableGetHandler) SupportsRequest(req *firHttp.RequestModel) bool {
	// First check the base requirements
	if !h.GetHandler.SupportsRequest(req) {
		return false
	}

	// Check custom path pattern if configured
	if h.customConfig.PathPattern != "" {
		matched, err := filepath.Match(h.customConfig.PathPattern, req.URL.Path)
		if err != nil || !matched {
			return false
		}
	}

	// Check if handler is enabled
	return h.customConfig.Enabled
}

// HandlerName returns the custom handler name
func (h *ConfigurableGetHandler) HandlerName() string {
	return h.customConfig.Name
}

// processOnLoadEvent handles onLoad event processing for GET requests
func (h *GetHandler) processOnLoadEvent(ctx context.Context, req *firHttp.RequestModel, routeID string, templateData map[string]interface{}) error {
	// Create event request for onLoad
	eventReq := services.EventRequest{
		ID:           "load",
		Target:       nil, // onLoad doesn't have a specific target
		ElementKey:   nil, // onLoad doesn't have an element key
		SessionID:    "",  // TODO: Extract session ID from request if available
		Context:      ctx,
		Params:       make(map[string]interface{}),
		RequestModel: req,
	}

	// Process the onLoad event through EventService
	eventResp, err := h.eventService.ProcessEvent(ctx, eventReq)
	if err != nil {
		// If no onLoad handler is registered, that's not an error
		if isNoHandlerError(err) {
			return nil
		}
		return fmt.Errorf("failed to process onLoad event: %w", err)
	}

	// If the event processing returned DOM events, we could potentially
	// use them to modify the template data, but for now we'll just
	// ensure the event was processed successfully
	_ = eventResp

	return nil
}

// isNoHandlerError checks if the error indicates no handler was found
// This is expected behavior when a route doesn't have an onLoad handler
func isNoHandlerError(err error) bool {
	// This would need to match the specific error returned by EventService
	// when no handler is found for an event
	return err != nil && (strings.Contains(err.Error(), "no handler") ||
		strings.Contains(err.Error(), "not found") ||
		strings.Contains(err.Error(), "not registered"))
}
