package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	firHttp "github.com/livefir/fir/internal/http"
	"github.com/livefir/fir/internal/services"
)

// FormHandler handles form-based requests (POST with form data)
type FormHandler struct {
	eventService    services.EventService
	renderService   services.RenderService
	responseBuilder services.ResponseBuilder
	validator       services.EventValidator
	config          HandlerConfig
}

// NewFormHandler creates a new form handler
func NewFormHandler(
	eventService services.EventService,
	renderService services.RenderService,
	responseBuilder services.ResponseBuilder,
	validator services.EventValidator,
) *FormHandler {
	return &FormHandler{
		eventService:    eventService,
		renderService:   renderService,
		responseBuilder: responseBuilder,
		validator:       validator,
		config: HandlerConfig{
			Name:        "form-handler",
			Priority:    20,
			Methods:     []string{http.MethodPost},
			ContentType: "application/x-www-form-urlencoded",
			Enabled:     true,
		},
	}
}

// Handle processes a form request
func (h *FormHandler) Handle(ctx context.Context, req *firHttp.RequestModel) (*firHttp.ResponseModel, error) {
	// Parse form data
	if err := h.parseFormData(req); err != nil {
		return h.responseBuilder.BuildErrorResponse(
			fmt.Errorf("failed to parse form data: %w", err),
			http.StatusBadRequest,
		)
	}

	// Determine the action to take based on form data
	action, err := h.determineFormAction(req)
	if err != nil {
		return h.responseBuilder.BuildErrorResponse(
			fmt.Errorf("failed to determine form action: %w", err),
			http.StatusBadRequest,
		)
	}

	switch action.Type {
	case "event":
		return h.handleFormEvent(ctx, req, action)
	case "submit":
		return h.handleFormSubmit(ctx, req, action)
	default:
		return h.responseBuilder.BuildErrorResponse(
			fmt.Errorf("unknown form action type: %s", action.Type),
			http.StatusBadRequest,
		)
	}
}

// SupportsRequest determines if this handler can process the given request
func (h *FormHandler) SupportsRequest(req *firHttp.RequestModel) bool {
	// Check if this is a POST request
	if req.Method != http.MethodPost {
		return false
	}

	// Exclude requests that are JSON events (should be handled by JSONEventHandler)
	// This ensures proper handler separation according to the legacy logic
	if req.Header.Get("X-FIR-MODE") == "event" {
		return false
	}

	// Check Content-Type header for form data
	contentType := req.Header.Get("Content-Type")
	return strings.HasPrefix(strings.ToLower(contentType), "application/x-www-form-urlencoded") ||
		strings.HasPrefix(strings.ToLower(contentType), "multipart/form-data")
}

// HandlerName returns the name of this handler
func (h *FormHandler) HandlerName() string {
	return h.config.Name
}

// GetConfig returns the handler configuration
func (h *FormHandler) GetConfig() HandlerConfig {
	return h.config
}

// SetConfig updates the handler configuration
func (h *FormHandler) SetConfig(config HandlerConfig) {
	h.config = config
}

// FormAction represents the action to be taken based on form data
type FormAction struct {
	Type       string                 // "event" or "submit"
	EventID    string                 // Event ID for event actions
	Target     string                 // Target element for events
	ElementKey string                 // Element key for events
	Data       map[string]interface{} // Form data
	Redirect   string                 // Redirect URL for submit actions
}

// parseFormData parses form data from the request
func (h *FormHandler) parseFormData(req *firHttp.RequestModel) error {
	if req.Form != nil {
		// Form already parsed
		return nil
	}

	// Parse form data based on content type
	contentType := req.Header.Get("Content-Type")

	if strings.HasPrefix(strings.ToLower(contentType), "multipart/form-data") {
		// Handle multipart form data
		// Note: This would typically require the original *http.Request
		// For now, we'll return an error and handle this in the future
		return fmt.Errorf("multipart form data not yet supported in handler layer")
	}

	// Parse URL-encoded form data
	if req.Body == nil {
		return fmt.Errorf("request body is empty")
	}

	// Read body and parse as form data
	// Note: In a real implementation, we'd need to properly handle the body
	// For now, we'll assume it's already parsed in req.Form
	if req.Form == nil {
		req.Form = make(url.Values)
	}

	return nil
}

// determineFormAction analyzes form data to determine the action to take
func (h *FormHandler) determineFormAction(req *firHttp.RequestModel) (*FormAction, error) {
	if req.Form == nil {
		return nil, fmt.Errorf("form data not parsed")
	}

	action := &FormAction{
		Data: make(map[string]interface{}),
	}

	// Convert form values to data map
	for key, values := range req.Form {
		if len(values) == 1 {
			action.Data[key] = values[0]
		} else {
			action.Data[key] = values
		}
	}

	// Check URL query parameters for event parameter first
	if eventID := req.QueryParams.Get("event"); eventID != "" {
		action.Type = "event"
		action.EventID = eventID
		action.Target = req.QueryParams.Get("target")
		action.ElementKey = req.QueryParams.Get("element_key")
		return action, nil
	}

	// Check for event-related form fields
	if eventID := req.Form.Get("_event"); eventID != "" {
		action.Type = "event"
		action.EventID = eventID
		action.Target = req.Form.Get("_target")
		action.ElementKey = req.Form.Get("_element_key")
		return action, nil
	}

	// Check for fir-specific event fields
	if eventID := req.Form.Get("fir-event"); eventID != "" {
		action.Type = "event"
		action.EventID = eventID
		action.Target = req.Form.Get("fir-target")
		action.ElementKey = req.Form.Get("fir-element-key")
		return action, nil
	}

	// Check for submit action
	if redirectURL := req.Form.Get("_redirect"); redirectURL != "" {
		action.Type = "submit"
		action.Redirect = redirectURL
		return action, nil
	}

	// Default to submit action
	action.Type = "submit"
	return action, nil
}

// handleFormEvent processes a form-based event
func (h *FormHandler) handleFormEvent(ctx context.Context, req *firHttp.RequestModel, action *FormAction) (*firHttp.ResponseModel, error) {
	// Create event request from form action
	eventReq := &services.EventRequest{
		ID:           action.EventID,
		Context:      ctx,
		RequestModel: req,
		Params:       action.Data,
		SessionID:    h.extractSessionID(req),
	}

	if action.Target != "" {
		eventReq.Target = &action.Target
	}

	if action.ElementKey != "" {
		eventReq.ElementKey = &action.ElementKey
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
		// For Phase 3: Return proper error response instead of causing fallback
		return h.responseBuilder.BuildErrorResponse(
			fmt.Errorf("failed to process form event: %w", err),
			http.StatusInternalServerError,
		)
	}

	// Build HTTP response from event response
	return h.responseBuilder.BuildEventResponse(eventResp, req)
}

// handleFormSubmit processes a regular form submission
func (h *FormHandler) handleFormSubmit(ctx context.Context, req *firHttp.RequestModel, action *FormAction) (*firHttp.ResponseModel, error) {
	// For now, we'll treat form submits as a special type of event
	// In a more complex system, this might involve different processing

	eventReq := &services.EventRequest{
		ID:           "form_submit",
		Context:      ctx,
		RequestModel: req,
		Params:       action.Data,
		SessionID:    h.extractSessionID(req),
	}

	// Process through event service
	eventResp, err := h.eventService.ProcessEvent(ctx, *eventReq)
	if err != nil {
		// If event processing fails, handle as a regular form submit
		return h.handleRegularFormSubmit(ctx, req, action)
	}

	// Handle redirect if specified
	if action.Redirect != "" && eventResp.Redirect == nil {
		return h.responseBuilder.BuildRedirectResponse(action.Redirect, http.StatusSeeOther)
	}

	return h.responseBuilder.BuildEventResponse(eventResp, req)
}

// handleRegularFormSubmit handles form submissions that don't trigger events
func (h *FormHandler) handleRegularFormSubmit(ctx context.Context, req *firHttp.RequestModel, action *FormAction) (*firHttp.ResponseModel, error) {
	// Handle redirect if specified
	if action.Redirect != "" {
		return h.responseBuilder.BuildRedirectResponse(action.Redirect, http.StatusSeeOther)
	}

	// Return a simple success response
	resp, err := h.responseBuilder.BuildTemplateResponse(&services.RenderResult{
		HTML: []byte("<p>Form submitted successfully</p>"),
	}, http.StatusOK)

	if err != nil {
		// For Phase 3: Return proper error response instead of bubbling up error
		return h.responseBuilder.BuildErrorResponse(
			fmt.Errorf("failed to build form submit response: %w", err),
			http.StatusInternalServerError,
		)
	}

	return resp, nil
}

// extractSessionID extracts session ID from request
func (h *FormHandler) extractSessionID(req *firHttp.RequestModel) string {
	// Try form field first
	if req.Form != nil {
		if sessionID := req.Form.Get("_session_id"); sessionID != "" {
			return sessionID
		}
	}

	// Try header
	if sessionID := req.Header.Get("X-Session-ID"); sessionID != "" {
		return sessionID
	}

	// Try cookie (would need to parse cookies from header)
	// This is a simplified implementation
	return ""
}
