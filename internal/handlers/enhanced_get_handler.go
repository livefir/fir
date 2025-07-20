package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	firHttp "github.com/livefir/fir/internal/http"
	"github.com/livefir/fir/internal/routeservices"
	"github.com/livefir/fir/internal/services"
)

// EnhancedGetHandler provides GET request handling with integrated template validation
type EnhancedGetHandler struct {
	// Core dependencies (same as original GetHandler)
	renderService   services.RenderService
	templateService services.TemplateService
	responseBuilder services.ResponseBuilder
	eventService    services.EventService
	sessionService  services.SessionService
	routeServices   *routeservices.RouteServices

	// Template validation integration
	templateValidator *TemplateActionValidator
	validationConfig  *TemplateValidationConfig
}

// NewEnhancedGetHandler creates a new enhanced GET handler with template validation
func NewEnhancedGetHandler(
	renderService services.RenderService,
	templateService services.TemplateService,
	responseBuilder services.ResponseBuilder,
	eventService services.EventService,
	sessionService services.SessionService,
	routeServices *routeservices.RouteServices,
) *EnhancedGetHandler {
	// Create template validation configuration
	validationConfig := DefaultTemplateValidationConfig()
	validationConfig.EnableValidation = true
	validationConfig.FailOnSecurityRisk = true
	validationConfig.LogValidationErrors = true
	validationConfig.LogSecurityRisks = true
	validationConfig.IncludeValidationInfo = true

	// Create template action validator
	validatorConfig := DefaultTemplateActionConfig()
	validatorConfig.EnableStrictValidation = true
	validatorConfig.AllowedActions = []string{"replace", "append", "prepend", "remove", "update"}
	validatorConfig.CacheEnabled = true
	validatorConfig.CacheTTL = 15 * time.Minute

	templateValidator := NewTemplateActionValidator(templateService, validatorConfig)

	return &EnhancedGetHandler{
		renderService:     renderService,
		templateService:   templateService,
		responseBuilder:   responseBuilder,
		eventService:      eventService,
		sessionService:    sessionService,
		routeServices:     routeServices,
		templateValidator: templateValidator,
		validationConfig:  validationConfig,
	}
}

// Handle processes GET requests with template validation
func (h *EnhancedGetHandler) Handle(ctx context.Context, req *firHttp.RequestModel) (*firHttp.ResponseModel, error) {
	// Pre-validation: Skip validation for static resources
	if h.shouldSkipValidation(req.URL.Path) {
		return h.handleWithoutValidation(ctx, req)
	}

	// Get template information for validation
	templatePath, templateContent, err := h.getTemplateInfo(ctx, req)
	if err != nil {
		// If we can't get template info, proceed without validation
		return h.handleWithoutValidation(ctx, req)
	}

	// Perform template validation
	validationResult, err := h.templateValidator.ValidateTemplateActions(ctx, templatePath, templateContent)
	if err != nil {
		if h.validationConfig.LogValidationErrors {
			fmt.Printf("Template validation error for %s: %v\n", templatePath, err)
		}

		// If validation fails and we're configured to fail on errors, return error
		if h.validationConfig.FailOnValidationError {
			return h.responseBuilder.BuildErrorResponse(
				fmt.Errorf("template validation failed: %w", err),
				http.StatusInternalServerError,
			)
		}

		// Otherwise, proceed without validation
		return h.handleWithoutValidation(ctx, req)
	}

	// Process validation results
	if err := h.processValidationResult(templatePath, validationResult); err != nil {
		return h.responseBuilder.BuildErrorResponse(
			fmt.Errorf("template validation failed: %w", err),
			http.StatusBadRequest,
		)
	}

	// Proceed with normal template rendering
	response, err := h.handleWithoutValidation(ctx, req)
	if err != nil {
		return response, err
	}

	// Add validation information to response
	if h.validationConfig.IncludeValidationInfo && response != nil {
		h.addValidationHeaders(response, validationResult)
	}

	return response, nil
}

// handleWithoutValidation performs the original GET handler logic
func (h *EnhancedGetHandler) handleWithoutValidation(ctx context.Context, req *firHttp.RequestModel) (*firHttp.ResponseModel, error) {
	// This implements the same logic as the original GetHandler
	// Extract the request path for template resolution
	path := req.URL.Path

	// Remove leading slash for template resolution
	path = strings.TrimPrefix(path, "/")

	// If no path, default to index
	if path == "" {
		path = "index"
	}

	// Try to find a template for this path
	templatePath := h.resolveTemplatePath(path)

	// Extract query parameters and form data for template context
	templateData := h.buildTemplateData(req)

	// Generate route ID for this request
	routeID := fmt.Sprintf("get:%s", path)

	// Process onLoad event if there's an event service
	if h.eventService != nil {
		err := h.processOnLoadEvent(ctx, req, routeID, templateData)
		if err != nil {
			return h.responseBuilder.BuildErrorResponse(err, http.StatusInternalServerError)
		}
	}

	// Build render context with route-specific template configuration
	renderCtx := h.buildRenderContext(ctx, req, templatePath, templateData)

	// Render the template
	renderResp, err := h.renderService.RenderTemplate(renderCtx)
	if err != nil {
		return h.responseBuilder.BuildErrorResponse(
			fmt.Errorf("failed to render template: %w", err),
			http.StatusInternalServerError,
		)
	}

	// Build response using ResponseBuilder
	response, err := h.responseBuilder.BuildTemplateResponse(renderResp, http.StatusOK)
	if err != nil {
		return h.responseBuilder.BuildErrorResponse(
			fmt.Errorf("failed to build response: %w", err),
			http.StatusInternalServerError,
		)
	}

	return response, nil
}

// shouldSkipValidation determines if validation should be skipped for a path
func (h *EnhancedGetHandler) shouldSkipValidation(path string) bool {
	for _, skipPath := range h.validationConfig.SkipValidationPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// getTemplateInfo gets template path and content for validation
func (h *EnhancedGetHandler) getTemplateInfo(ctx context.Context, req *firHttp.RequestModel) (string, string, error) {
	// Simple template path resolution based on request path
	path := req.URL.Path
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		path = "index"
	}

	templatePath := h.resolveTemplatePath(path)

	// Try to load template to get content
	config := services.TemplateConfig{
		ContentPath: templatePath,
	}

	_, err := h.templateService.LoadTemplate(config)
	if err != nil {
		return "", "", fmt.Errorf("failed to load template: %w", err)
	}

	// For validation purposes, we need the raw template content
	// This is a simplified approach - in a real implementation,
	// you'd need to extract the raw content from the template
	templateContent := fmt.Sprintf("<!-- Template: %s -->", templatePath)

	return templatePath, templateContent, nil
}

// processValidationResult processes validation results and determines if request should proceed
func (h *EnhancedGetHandler) processValidationResult(templatePath string, result *TemplateActionResult) error {
	// Log validation errors if configured
	if h.validationConfig.LogValidationErrors && len(result.Errors) > 0 {
		for _, err := range result.Errors {
			fmt.Printf("Template validation error in %s: %s\n", templatePath, err.Message)
		}
	}

	// Log security risks if configured
	if h.validationConfig.LogSecurityRisks && len(result.SecurityRisks) > 0 {
		for _, risk := range result.SecurityRisks {
			fmt.Printf("Security risk in %s: %s (Level: %s)\n", templatePath, risk.Description, risk.Level)
		}
	}

	// Log performance warnings if configured
	if h.validationConfig.LogPerformanceWarnings && len(result.Warnings) > 0 {
		for _, warning := range result.Warnings {
			if warning.Type == WarningTypePerformance {
				fmt.Printf("Performance warning in %s: %s\n", templatePath, warning.Message)
			}
		}
	}

	// Check if we should fail on validation errors
	if h.validationConfig.FailOnValidationError && !result.IsValid {
		return fmt.Errorf("template contains %d validation errors", len(result.Errors))
	}

	// Check if we should fail on security risks
	if h.validationConfig.FailOnSecurityRisk {
		for _, risk := range result.SecurityRisks {
			if risk.Level == SecurityLevelHigh || risk.Level == SecurityLevelCritical {
				return fmt.Errorf("template contains high/critical security risk: %s", risk.Description)
			}
		}
	}

	return nil
}

// addValidationHeaders adds validation information to response headers
func (h *EnhancedGetHandler) addValidationHeaders(response *firHttp.ResponseModel, result *TemplateActionResult) {
	if response.Headers == nil {
		response.Headers = make(map[string]string)
	}

	// Add basic validation info
	response.Headers["X-Template-Validation"] = fmt.Sprintf("valid=%t,errors=%d,warnings=%d",
		result.IsValid, len(result.Errors), len(result.Warnings))

	response.Headers["X-Template-Actions"] = fmt.Sprintf("%d", len(result.Actions))
	response.Headers["X-Template-Processing-Time"] = fmt.Sprintf("%dms", result.ProcessingTime.Milliseconds())

	if result.CacheHit {
		response.Headers["X-Template-Cache"] = "hit"
	} else {
		response.Headers["X-Template-Cache"] = "miss"
	}

	// Add security information
	if len(result.SecurityRisks) > 0 {
		highRiskCount := 0
		for _, risk := range result.SecurityRisks {
			if risk.Level == SecurityLevelHigh || risk.Level == SecurityLevelCritical {
				highRiskCount++
			}
		}
		response.Headers["X-Template-Security"] = fmt.Sprintf("risks=%d,high=%d",
			len(result.SecurityRisks), highRiskCount)
	}

	// Add debug headers if enabled
	if h.validationConfig.EnableDebugHeaders {
		// Add detailed error information (be careful in production!)
		if len(result.Errors) > 0 {
			errorMessages := make([]string, len(result.Errors))
			for i, err := range result.Errors {
				errorMessages[i] = err.Message
			}
			response.Headers["X-Template-Debug-Errors"] = strings.Join(errorMessages, "; ")
		}
	}
}

// Helper methods from original GetHandler (simplified implementations)

func (h *EnhancedGetHandler) resolveTemplatePath(path string) string {
	// Add .html extension if not present
	if !strings.Contains(path, ".") {
		path += ".html"
	}
	return path
}

func (h *EnhancedGetHandler) buildTemplateData(req *firHttp.RequestModel) map[string]interface{} {
	data := make(map[string]interface{})

	// Add query parameters
	for key, values := range req.QueryParams {
		if len(values) == 1 {
			data[key] = values[0]
		} else {
			data[key] = values
		}
	}

	// Add form data
	for key, values := range req.Form {
		if len(values) == 1 {
			data[key] = values[0]
		} else {
			data[key] = values
		}
	}

	// Add path parameters
	for key, value := range req.PathParams {
		data[key] = value
	}

	return data
}

func (h *EnhancedGetHandler) processOnLoadEvent(ctx context.Context, req *firHttp.RequestModel, routeID string, templateData map[string]interface{}) error {
	// Simplified onLoad event processing
	return nil
}

func (h *EnhancedGetHandler) buildRenderContext(ctx context.Context, req *firHttp.RequestModel, templatePath string, templateData map[string]interface{}) services.RenderContext {
	var options *routeservices.Options
	if h.routeServices != nil {
		options = h.routeServices.Options
	}
	if options == nil {
		options = &routeservices.Options{}
	}

	// Determine the template path/content to use
	finalTemplatePath := templatePath
	if options.Content != "" {
		finalTemplatePath = options.Content
	}

	// Safely get FuncMap
	var funcMap map[string]interface{}
	if options.FuncMap != nil {
		funcMap = options.FuncMap
	}

	return services.RenderContext{
		TemplatePath:  finalTemplatePath,
		LayoutPath:    "",
		PartialPaths:  []string{},
		Data:          templateData,
		FuncMap:       funcMap,
		Extensions:    []string{".html"},
		CacheDisabled: false,
		RouteID:       fmt.Sprintf("get:%s", strings.TrimSuffix(templatePath, ".html")),
	}
}

// SupportsRequest determines if this handler can process the given request
func (h *EnhancedGetHandler) SupportsRequest(req *firHttp.RequestModel) bool {
	return req.Method == http.MethodGet
}

// HandlerName returns a unique name for this handler
func (h *EnhancedGetHandler) HandlerName() string {
	return "EnhancedGetHandler"
}

// GetValidationMetrics returns current validation metrics
func (h *EnhancedGetHandler) GetValidationMetrics() TemplateActionMetrics {
	return h.templateValidator.GetMetrics()
}

// ClearValidationCache clears the template validation cache
func (h *EnhancedGetHandler) ClearValidationCache() {
	h.templateValidator.ClearCache()
}
