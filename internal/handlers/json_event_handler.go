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
		// For Phase 5: If the event service is not implemented (like our noOpEventService),
		// propagate the error to cause handler chain failure and fallback to legacy
		return nil, fmt.Errorf("failed to process JSON event: %w", err)
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

	// Check for the X-FIR-MODE header with value "event" (matches legacy logic)
	// Legacy: r.Header.Get("X-FIR-MODE") == "event" && r.Method == http.MethodPost
	if req.Header.Get("X-FIR-MODE") != "event" {
		return false
	}

	// Check Content-Type header for JSON
	contentType := req.Header.Get("Content-Type")
	return strings.HasPrefix(strings.ToLower(contentType), "application/json")
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
		Params:       make(map[string]interface{}), // Will be populated after extracting control fields
	}

	// Extract event ID (required) - accept both "event_id" (legacy/Alpine.js) and "id" (modern)
	if id, ok := eventData["event_id"].(string); ok {
		eventReq.ID = id
		delete(eventData, "event_id") // Remove from params
	} else if id, ok := eventData["id"].(string); ok {
		eventReq.ID = id
		delete(eventData, "id") // Remove from params
	} else {
		return nil, fmt.Errorf("event ID is required")
	}

	// Handle legacy Event format where params are nested under "params" field
	if params, ok := eventData["params"]; ok {
		// Use the nested params as the actual parameters
		if paramsMap, ok := params.(map[string]interface{}); ok {
			eventReq.Params = paramsMap
		} else {
			eventReq.Params = map[string]interface{}{"params": params}
		}
		delete(eventData, "params") // Remove from eventData
	}

	// Extract target (optional)
	if target, ok := eventData["target"].(string); ok {
		eventReq.Target = &target
		delete(eventData, "target") // Remove from params
	}

	// Extract element key (optional)
	if elementKey, ok := eventData["element_key"].(string); ok {
		eventReq.ElementKey = &elementKey
		delete(eventData, "element_key") // Remove from params
	}

	// Extract session ID (optional, could come from params or headers)
	if sessionID, ok := eventData["session_id"].(string); ok {
		eventReq.SessionID = sessionID
		delete(eventData, "session_id") // Remove from params
	}

	// Extract timestamp (optional, remove from params)
	delete(eventData, "ts") // Remove from params (safe even if key doesn't exist)

	// If params weren't nested (modern format), use remaining eventData as params
	if eventReq.Params == nil {
		eventReq.Params = eventData
	}

	// If session ID wasn't in the JSON, try to get it from headers
	if eventReq.SessionID == "" {
		if sessionID := req.Header.Get("X-Session-ID"); sessionID != "" {
			eventReq.SessionID = sessionID
		}
	}

	return eventReq, nil
}

// ConfigurableJSONEventHandler extends JSONEventHandler with more configuration options
type ConfigurableJSONEventHandler struct {
	*JSONEventHandler
	eventPatterns   []string
	requiredHeaders []string
	optionalHeaders []string
	validateBody    bool
	maxBodySize     int64
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
