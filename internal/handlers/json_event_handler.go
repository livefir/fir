package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	firHttp "github.com/livefir/fir/internal/http"
	"github.com/livefir/fir/internal/services"
)

// JSONEventHandler handles JSON-based event requests
type JSONEventHandler struct {
	eventService    services.EventService
	renderService   services.RenderService
	responseBuilder services.ResponseBuilder
	validator       services.EventValidator
	config          HandlerConfig
}

// NewJSONEventHandler creates a new JSON event handler
func NewJSONEventHandler(
	eventService services.EventService,
	renderService services.RenderService,
	responseBuilder services.ResponseBuilder,
	validator services.EventValidator,
) *JSONEventHandler {
	return &JSONEventHandler{
		eventService:    eventService,
		renderService:   renderService,
		responseBuilder: responseBuilder,
		validator:       validator,
		config: HandlerConfig{
			Name:        "json-event-handler",
			Priority:    10,
			Methods:     []string{http.MethodPost},
			ContentType: "application/json",
			Enabled:     true,
		},
	}
}

// Handle processes a JSON event request
func (h *JSONEventHandler) Handle(ctx context.Context, req *firHttp.RequestModel) (*firHttp.ResponseModel, error) {
	// Parse JSON event from request body
	eventReq, err := h.parseJSONEvent(req)
	if err != nil {
		return h.responseBuilder.BuildErrorResponse(
			fmt.Errorf("failed to parse JSON event: %w", err),
			http.StatusBadRequest,
		)
	}

	// Validate the event request
	if h.validator != nil {
		if err := h.validator.ValidateEvent(*eventReq); err != nil {
			return h.responseBuilder.BuildErrorResponse(
				fmt.Errorf("event validation failed: %w", err),
				http.StatusBadRequest,
			)
		}
	}

	// Process the event through the event service
	eventResp, err := h.eventService.ProcessEvent(ctx, *eventReq)
	if err != nil {
		return h.responseBuilder.BuildErrorResponse(
			fmt.Errorf("failed to process event: %w", err),
			http.StatusInternalServerError,
		)
	}

	// Build HTTP response from event response
	return h.responseBuilder.BuildEventResponse(eventResp, req)
}

// SupportsRequest determines if this handler can process the given request
func (h *JSONEventHandler) SupportsRequest(req *firHttp.RequestModel) bool {
	// Check if this is a POST request
	if req.Method != http.MethodPost {
		return false
	}

	// Check Content-Type header
	contentType := req.Header.Get("Content-Type")
	if !strings.HasPrefix(strings.ToLower(contentType), "application/json") {
		return false
	}

	// Check if this looks like an event request
	// This could be enhanced with more sophisticated detection
	return h.looksLikeEventRequest(req)
}

// HandlerName returns the name of this handler
func (h *JSONEventHandler) HandlerName() string {
	return h.config.Name
}

// GetConfig returns the handler configuration
func (h *JSONEventHandler) GetConfig() HandlerConfig {
	return h.config
}

// SetConfig updates the handler configuration
func (h *JSONEventHandler) SetConfig(config HandlerConfig) {
	h.config = config
}

// parseJSONEvent parses a JSON event request from the HTTP request body
func (h *JSONEventHandler) parseJSONEvent(req *firHttp.RequestModel) (*services.EventRequest, error) {
	if req.Body == nil {
		return nil, fmt.Errorf("request body is empty")
	}

	// Parse JSON from body
	var eventData map[string]interface{}
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&eventData); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	// Extract event information
	eventReq := &services.EventRequest{
		Context:      req.Context,
		RequestModel: req,
		Params:       eventData,
	}

	// Extract event ID (required)
	if id, ok := eventData["id"].(string); ok {
		eventReq.ID = id
	} else {
		return nil, fmt.Errorf("event ID is required")
	}

	// Extract target (optional)
	if target, ok := eventData["target"].(string); ok {
		eventReq.Target = &target
	}

	// Extract element key (optional)
	if elementKey, ok := eventData["element_key"].(string); ok {
		eventReq.ElementKey = &elementKey
	}

	// Extract session ID (optional, could come from params or headers)
	if sessionID, ok := eventData["session_id"].(string); ok {
		eventReq.SessionID = sessionID
	} else if sessionID := req.Header.Get("X-Session-ID"); sessionID != "" {
		eventReq.SessionID = sessionID
	}

	return eventReq, nil
}

// looksLikeEventRequest uses heuristics to determine if this looks like an event request
func (h *JSONEventHandler) looksLikeEventRequest(req *firHttp.RequestModel) bool {
	// Check if URL is available
	if req.URL == nil {
		// If no URL available, default to true for JSON POST requests
		return true
	}
	
	// Check URL path patterns that typically indicate events
	path := req.URL.Path
	
	// Common event endpoint patterns
	eventPatterns := []string{
		"/events",
		"/event",
		"/_event",
		"/_events",
	}

	for _, pattern := range eventPatterns {
		if strings.Contains(path, pattern) {
			return true
		}
	}

	// Check for event-related headers
	if req.Header.Get("X-Event-ID") != "" {
		return true
	}

	if req.Header.Get("X-FIR-Event") != "" {
		return true
	}

	// Check if the path ends with an event-like suffix
	if strings.HasSuffix(path, "/fire") || strings.HasSuffix(path, "/trigger") {
		return true
	}

	// Default to true for JSON POST requests if no other patterns match
	// This can be made more restrictive based on the application's needs
	return true
}

// ConfigurableJSONEventHandler extends JSONEventHandler with more configuration options
type ConfigurableJSONEventHandler struct {
	*JSONEventHandler
	eventPatterns    []string
	requiredHeaders  []string
	optionalHeaders  []string
	validateBody     bool
	maxBodySize      int64
}

// NewConfigurableJSONEventHandler creates a configurable JSON event handler
func NewConfigurableJSONEventHandler(
	eventService services.EventService,
	renderService services.RenderService,
	responseBuilder services.ResponseBuilder,
	validator services.EventValidator,
	config JSONEventHandlerConfig,
) *ConfigurableJSONEventHandler {
	base := NewJSONEventHandler(eventService, renderService, responseBuilder, validator)
	
	return &ConfigurableJSONEventHandler{
		JSONEventHandler: base,
		eventPatterns:    config.EventPatterns,
		requiredHeaders:  config.RequiredHeaders,
		optionalHeaders:  config.OptionalHeaders,
		validateBody:     config.ValidateBody,
		maxBodySize:      config.MaxBodySize,
	}
}

// JSONEventHandlerConfig contains configuration for the JSON event handler
type JSONEventHandlerConfig struct {
	EventPatterns   []string // URL patterns that indicate event requests
	RequiredHeaders []string // Headers that must be present
	OptionalHeaders []string // Headers that may be present
	ValidateBody    bool     // Whether to validate the request body structure
	MaxBodySize     int64    // Maximum size of request body in bytes
}

// SupportsRequest override with configurable patterns
func (h *ConfigurableJSONEventHandler) SupportsRequest(req *firHttp.RequestModel) bool {
	// Check if this is a POST request
	if req.Method != http.MethodPost {
		return false
	}

	// Check Content-Type header
	contentType := req.Header.Get("Content-Type")
	if !strings.HasPrefix(strings.ToLower(contentType), "application/json") {
		return false
	}

	// Check required headers
	for _, header := range h.requiredHeaders {
		if req.Header.Get(header) == "" {
			return false
		}
	}

	// Check URL patterns
	if len(h.eventPatterns) > 0 {
		path := req.URL.Path
		for _, pattern := range h.eventPatterns {
			if strings.Contains(path, pattern) {
				return true
			}
		}
		return false // No patterns matched
	}

	// Fall back to parent logic if no patterns configured
	return h.JSONEventHandler.SupportsRequest(req)
}
