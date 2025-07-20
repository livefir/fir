package handlers

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/livefir/fir/internal/firattr"
	"github.com/livefir/fir/internal/services"
	"github.com/patrickmn/go-cache"
)

// TemplateActionValidator provides comprehensive template action validation and processing
type TemplateActionValidator struct {
	// Core dependencies
	templateService services.TemplateService
	cache           *cache.Cache

	// Configuration
	config *TemplateActionConfig

	// Metrics
	metrics *TemplateActionMetrics
}

// TemplateActionConfig holds configuration for template action validation
type TemplateActionConfig struct {
	// Validation settings
	EnableStrictValidation bool     // Enable strict action validation
	MaxActionDepth         int      // Maximum nesting depth for actions
	MaxActionsPerTemplate  int      // Maximum actions per template
	AllowedActions         []string // Whitelist of allowed actions
	ForbiddenActions       []string // Blacklist of forbidden actions

	// Performance settings
	CacheEnabled      bool          // Enable action validation caching
	CacheTTL          time.Duration // Cache time-to-live
	ValidationTimeout time.Duration // Timeout for validation operations

	// Processing settings
	EnableActionOptimization bool // Enable action optimization
	EnableActionRewrite      bool // Enable action rewriting
	PreprocessActions        bool // Preprocess actions during validation
}

// TemplateActionMetrics tracks template action processing metrics
type TemplateActionMetrics struct {
	mu                    sync.RWMutex
	ValidationCount       int64         // Total validations performed
	ValidationErrors      int64         // Total validation errors
	CacheHits             int64         // Cache hits
	CacheMisses           int64         // Cache misses
	ProcessingTime        time.Duration // Total processing time
	AverageProcessingTime time.Duration // Average processing time
	LastValidationTime    time.Time     // Last validation timestamp
}

// TemplateActionResult represents the result of template action validation
type TemplateActionResult struct {
	IsValid          bool                    // Whether actions are valid
	Actions          []ValidatedAction       // Validated actions
	Errors           []TemplateActionError   // Validation errors
	Warnings         []TemplateActionWarning // Validation warnings
	ProcessingTime   time.Duration           // Time taken for validation
	CacheHit         bool                    // Whether result came from cache
	OptimizedActions []ValidatedAction       // Optimized actions (if enabled)
	SecurityRisks    []SecurityRisk          // Identified security risks
}

// ValidatedAction represents a validated template action
type ValidatedAction struct {
	OriginalExpression  string                   // Original fir: expression
	ParsedAttribute     *firattr.ParsedAttribute // Parsed attribute
	ActionType          ActionType               // Type of action
	Parameters          []ActionParameter        // Action parameters
	SecurityLevel       SecurityLevel            // Security assessment
	PerformanceImpact   PerformanceImpact        // Performance impact assessment
	Dependencies        []string                 // Action dependencies
	OptimizedExpression string                   // Optimized expression (if applicable)
}

// ActionType represents different types of template actions
type ActionType string

const (
	ActionTypeReplace ActionType = "replace"
	ActionTypeAppend  ActionType = "append"
	ActionTypePrepend ActionType = "prepend"
	ActionTypeRemove  ActionType = "remove"
	ActionTypeUpdate  ActionType = "update"
	ActionTypeCustom  ActionType = "custom"
	ActionTypeUnknown ActionType = "unknown"
)

// SecurityLevel represents the security assessment of an action
type SecurityLevel string

const (
	SecurityLevelSafe     SecurityLevel = "safe"
	SecurityLevelLow      SecurityLevel = "low"
	SecurityLevelMedium   SecurityLevel = "medium"
	SecurityLevelHigh     SecurityLevel = "high"
	SecurityLevelCritical SecurityLevel = "critical"
)

// PerformanceImpact represents the performance impact of an action
type PerformanceImpact string

const (
	PerformanceImpactLow    PerformanceImpact = "low"
	PerformanceImpactMedium PerformanceImpact = "medium"
	PerformanceImpactHigh   PerformanceImpact = "high"
)

// ActionParameter represents a parameter within a template action
type ActionParameter struct {
	Name     string        // Parameter name
	Value    interface{}   // Parameter value
	Type     ParameterType // Parameter type
	IsValid  bool          // Whether parameter is valid
	ErrorMsg string        // Error message if invalid
}

// ParameterType represents different types of action parameters
type ParameterType string

const (
	ParameterTypeString   ParameterType = "string"
	ParameterTypeNumber   ParameterType = "number"
	ParameterTypeBoolean  ParameterType = "boolean"
	ParameterTypeArray    ParameterType = "array"
	ParameterTypeObject   ParameterType = "object"
	ParameterTypeFunction ParameterType = "function"
)

// TemplateActionError represents a template action validation error
type TemplateActionError struct {
	Type       ErrorType `json:"type"`
	Message    string    `json:"message"`
	Expression string    `json:"expression"`
	Line       int       `json:"line,omitempty"`
	Column     int       `json:"column,omitempty"`
	Severity   string    `json:"severity"`
	Suggestion string    `json:"suggestion,omitempty"`
}

// TemplateActionWarning represents a template action validation warning
type TemplateActionWarning struct {
	Type       WarningType `json:"type"`
	Message    string      `json:"message"`
	Expression string      `json:"expression"`
	Line       int         `json:"line,omitempty"`
	Column     int         `json:"column,omitempty"`
	Suggestion string      `json:"suggestion,omitempty"`
}

// SecurityRisk represents a security risk identified in template actions
type SecurityRisk struct {
	Type        SecurityRiskType `json:"type"`
	Level       SecurityLevel    `json:"level"`
	Description string           `json:"description"`
	Expression  string           `json:"expression"`
	Mitigation  string           `json:"mitigation"`
}

// ErrorType represents different types of template action errors
type ErrorType string

const (
	ErrorTypeSyntax        ErrorType = "syntax"
	ErrorTypeValidation    ErrorType = "validation"
	ErrorTypeSecurity      ErrorType = "security"
	ErrorTypePerformance   ErrorType = "performance"
	ErrorTypeCompatibility ErrorType = "compatibility"
)

// WarningType represents different types of template action warnings
type WarningType string

const (
	WarningTypeDeprecated    WarningType = "deprecated"
	WarningTypePerformance   WarningType = "performance"
	WarningTypeBestPractice  WarningType = "best_practice"
	WarningTypeCompatibility WarningType = "compatibility"
)

// SecurityRiskType represents different types of security risks
type SecurityRiskType string

const (
	SecurityRiskTypeXSS          SecurityRiskType = "xss"
	SecurityRiskTypeInjection    SecurityRiskType = "injection"
	SecurityRiskTypeUnauthorized SecurityRiskType = "unauthorized"
	SecurityRiskTypeDataExposure SecurityRiskType = "data_exposure"
)

// DefaultTemplateActionConfig provides default configuration
func DefaultTemplateActionConfig() *TemplateActionConfig {
	return &TemplateActionConfig{
		EnableStrictValidation:   true,
		MaxActionDepth:           10,
		MaxActionsPerTemplate:    50, // Lower limit for testing
		AllowedActions:           []string{"replace", "append", "prepend", "remove", "update"},
		ForbiddenActions:         []string{},
		CacheEnabled:             true,
		CacheTTL:                 15 * time.Minute,
		ValidationTimeout:        5 * time.Second,
		EnableActionOptimization: true,
		EnableActionRewrite:      true,
		PreprocessActions:        true,
	}
}

// NewTemplateActionValidator creates a new template action validator
func NewTemplateActionValidator(templateService services.TemplateService, config *TemplateActionConfig) *TemplateActionValidator {
	if config == nil {
		config = DefaultTemplateActionConfig()
	}

	var validationCache *cache.Cache
	if config.CacheEnabled {
		validationCache = cache.New(config.CacheTTL, config.CacheTTL*2)
	}

	return &TemplateActionValidator{
		templateService: templateService,
		cache:           validationCache,
		config:          config,
		metrics:         &TemplateActionMetrics{},
	}
}

// ValidateTemplateActions validates all actions within a template
func (v *TemplateActionValidator) ValidateTemplateActions(ctx context.Context, templatePath string, templateContent string) (*TemplateActionResult, error) {
	startTime := time.Now()
	defer func() {
		v.updateMetrics(time.Since(startTime))
	}()

	// Check cache first
	if v.config.CacheEnabled {
		cacheKey := v.generateCacheKey(templatePath, templateContent)
		if cached, found := v.cache.Get(cacheKey); found {
			if result, ok := cached.(*TemplateActionResult); ok {
				result.CacheHit = true
				v.incrementCacheHits()
				return result, nil
			}
		}
		v.incrementCacheMisses()
	}

	// Create attribute extractor
	extractor, err := firattr.NewAttributeExtractor()
	if err != nil {
		return nil, fmt.Errorf("failed to create attribute extractor: %w", err)
	}

	// Extract template actions
	attributes, err := extractor.ExtractFromTemplate(templateContent)
	if err != nil {
		return nil, fmt.Errorf("failed to extract template actions: %w", err)
	}

	// Validate extracted actions
	result := &TemplateActionResult{
		IsValid:        true,
		Actions:        make([]ValidatedAction, 0, len(attributes)),
		Errors:         []TemplateActionError{},
		Warnings:       []TemplateActionWarning{},
		ProcessingTime: 0,
		CacheHit:       false,
		SecurityRisks:  []SecurityRisk{},
	}

	// Process each action
	for _, attr := range attributes {
		validatedAction, actionErrors, actionWarnings, securityRisks := v.validateAction(ctx, &attr)

		result.Actions = append(result.Actions, validatedAction)
		result.Errors = append(result.Errors, actionErrors...)
		result.Warnings = append(result.Warnings, actionWarnings...)
		result.SecurityRisks = append(result.SecurityRisks, securityRisks...)

		if len(actionErrors) > 0 {
			result.IsValid = false
		}
	}

	// Perform optimization if enabled
	if v.config.EnableActionOptimization {
		result.OptimizedActions = v.optimizeActions(result.Actions)
	}

	// Validate overall template constraints
	v.validateTemplateConstraints(result)

	result.ProcessingTime = time.Since(startTime)

	// Cache the result
	if v.config.CacheEnabled {
		cacheKey := v.generateCacheKey(templatePath, templateContent)
		v.cache.Set(cacheKey, result, v.config.CacheTTL)
	}

	return result, nil
}

// validateAction validates a single template action
func (v *TemplateActionValidator) validateAction(ctx context.Context, attr *firattr.ParsedAttribute) (ValidatedAction, []TemplateActionError, []TemplateActionWarning, []SecurityRisk) {
	action := ValidatedAction{
		OriginalExpression: attr.ToCanonicalForm(),
		ParsedAttribute:    attr,
		Parameters:         []ActionParameter{},
		Dependencies:       []string{},
	}

	var errors []TemplateActionError
	var warnings []TemplateActionWarning
	var risks []SecurityRisk

	// Determine action type
	action.ActionType = v.determineActionType(attr.Action)

	// Validate action syntax
	if syntaxErrors := v.validateActionSyntax(attr); len(syntaxErrors) > 0 {
		errors = append(errors, syntaxErrors...)
	}

	// Validate action semantics
	if semanticErrors := v.validateActionSemantics(attr); len(semanticErrors) > 0 {
		errors = append(errors, semanticErrors...)
	}

	// Assess security level
	securityLevel, securityRisks := v.assessSecurityLevel(attr)
	action.SecurityLevel = securityLevel
	risks = append(risks, securityRisks...)

	// Assess performance impact
	action.PerformanceImpact = v.assessPerformanceImpact(attr)

	// Check for warnings
	if actionWarnings := v.checkActionWarnings(attr); len(actionWarnings) > 0 {
		warnings = append(warnings, actionWarnings...)
	}

	// Generate optimized expression if enabled
	if v.config.EnableActionRewrite {
		action.OptimizedExpression = v.optimizeActionExpression(attr)
	}

	return action, errors, warnings, risks
}

// determineActionType determines the type of action from the expression
func (v *TemplateActionValidator) determineActionType(actionExpr string) ActionType {
	// Extract action name from $fir.actionName() format
	if strings.HasPrefix(actionExpr, "$fir.") {
		// Extract the action name between $fir. and (
		re := regexp.MustCompile(`\$fir\.([a-zA-Z][a-zA-Z0-9]*)\(`)
		matches := re.FindStringSubmatch(actionExpr)
		if len(matches) > 1 {
			actionName := matches[1]
			switch actionName {
			case "replace":
				return ActionTypeReplace
			case "append":
				return ActionTypeAppend
			case "prepend":
				return ActionTypePrepend
			case "remove":
				return ActionTypeRemove
			case "update":
				return ActionTypeUpdate
			default:
				return ActionTypeCustom
			}
		}
	}

	// Fallback to old logic for non-$fir actions
	switch {
	case strings.Contains(actionExpr, "replace"):
		return ActionTypeReplace
	case strings.Contains(actionExpr, "append"):
		return ActionTypeAppend
	case strings.Contains(actionExpr, "prepend"):
		return ActionTypePrepend
	case strings.Contains(actionExpr, "remove"):
		return ActionTypeRemove
	case strings.Contains(actionExpr, "update"):
		return ActionTypeUpdate
	default:
		return ActionTypeUnknown
	}
}

// extractActionName extracts the action name from an action expression
func (v *TemplateActionValidator) extractActionName(actionExpr string) string {
	// Extract action name from $fir.actionName() format
	if strings.HasPrefix(actionExpr, "$fir.") {
		re := regexp.MustCompile(`\$fir\.([a-zA-Z][a-zA-Z0-9]*)\(`)
		matches := re.FindStringSubmatch(actionExpr)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	// For non-$fir actions, return the action type as string
	actionType := v.determineActionType(actionExpr)
	return string(actionType)
}

// validateActionSyntax validates the syntax of a template action
func (v *TemplateActionValidator) validateActionSyntax(attr *firattr.ParsedAttribute) []TemplateActionError {
	var errors []TemplateActionError

	// Check for empty action
	if attr.Action == "" {
		errors = append(errors, TemplateActionError{
			Type:       ErrorTypeSyntax,
			Message:    "Action expression cannot be empty",
			Expression: attr.ToCanonicalForm(),
			Severity:   "error",
			Suggestion: "Provide a valid action expression like $fir.replace()",
		})
	}

	// Validate action expression format
	if !v.isValidActionFormat(attr.Action) {
		errors = append(errors, TemplateActionError{
			Type:       ErrorTypeSyntax,
			Message:    "Invalid action expression format",
			Expression: attr.Action,
			Severity:   "error",
			Suggestion: "Use valid $fir.action() format",
		})
	}

	// Validate event format
	for _, event := range attr.Events {
		if !v.isValidEventFormat(event.Name) {
			errors = append(errors, TemplateActionError{
				Type:       ErrorTypeSyntax,
				Message:    fmt.Sprintf("Invalid event format: %s", event.Name),
				Expression: attr.ToCanonicalForm(),
				Severity:   "error",
				Suggestion: "Use valid event naming conventions",
			})
		}
	}

	return errors
}

// validateActionSemantics validates the semantic correctness of actions
func (v *TemplateActionValidator) validateActionSemantics(attr *firattr.ParsedAttribute) []TemplateActionError {
	var errors []TemplateActionError

	// Extract actual action name for validation
	actionName := v.extractActionName(attr.Action)

	// Check allowed actions
	if v.config.EnableStrictValidation && len(v.config.AllowedActions) > 0 {
		if !v.isActionAllowed(actionName) {
			errors = append(errors, TemplateActionError{
				Type:       ErrorTypeValidation,
				Message:    fmt.Sprintf("Action '%s' is not allowed", actionName),
				Expression: attr.Action,
				Severity:   "error",
				Suggestion: fmt.Sprintf("Use one of the allowed actions: %v", v.config.AllowedActions),
			})
		}
	}

	// Check forbidden actions
	if len(v.config.ForbiddenActions) > 0 {
		if v.isActionForbidden(actionName) {
			errors = append(errors, TemplateActionError{
				Type:       ErrorTypeValidation,
				Message:    fmt.Sprintf("Action '%s' is forbidden", actionName),
				Expression: attr.Action,
				Severity:   "error",
				Suggestion: "Use an allowed action type",
			})
		}
	}

	return errors
}

// assessSecurityLevel assesses the security level of an action
func (v *TemplateActionValidator) assessSecurityLevel(attr *firattr.ParsedAttribute) (SecurityLevel, []SecurityRisk) {
	var risks []SecurityRisk
	level := SecurityLevelSafe

	// Check for potential XSS risks
	if v.hasXSSRisk(attr.Action) {
		level = SecurityLevelHigh
		risks = append(risks, SecurityRisk{
			Type:        SecurityRiskTypeXSS,
			Level:       SecurityLevelHigh,
			Description: "Action may introduce XSS vulnerability",
			Expression:  attr.Action,
			Mitigation:  "Ensure proper output escaping and input validation",
		})
	}

	// Check for injection risks
	if v.hasInjectionRisk(attr.Action) {
		level = SecurityLevelMedium
		risks = append(risks, SecurityRisk{
			Type:        SecurityRiskTypeInjection,
			Level:       SecurityLevelMedium,
			Description: "Action may be vulnerable to injection attacks",
			Expression:  attr.Action,
			Mitigation:  "Validate and sanitize all user inputs",
		})
	}

	return level, risks
}

// assessPerformanceImpact assesses the performance impact of an action
func (v *TemplateActionValidator) assessPerformanceImpact(attr *firattr.ParsedAttribute) PerformanceImpact {
	// Count total events across all event types
	eventCount := len(attr.Events)

	// Performance assessment based on event count and action type
	// Replace actions have higher performance impact due to DOM replacement
	isReplaceAction := strings.Contains(attr.Action, "replace")

	if eventCount > 5 {
		if isReplaceAction {
			return PerformanceImpactHigh
		} else {
			return PerformanceImpactMedium // Non-replace actions with many events are medium impact
		}
	}
	// Medium impact: more than 2 events
	if eventCount > 2 {
		return PerformanceImpactMedium
	}
	return PerformanceImpactLow
}

// checkActionWarnings checks for potential warnings in actions
func (v *TemplateActionValidator) checkActionWarnings(attr *firattr.ParsedAttribute) []TemplateActionWarning {
	var warnings []TemplateActionWarning

	// Check for deprecated patterns
	if v.isDeprecatedPattern(attr.Action) {
		warnings = append(warnings, TemplateActionWarning{
			Type:       WarningTypeDeprecated,
			Message:    "This action pattern is deprecated",
			Expression: attr.Action,
			Suggestion: "Use the newer action syntax",
		})
	}

	// Check for performance concerns
	performanceImpact := v.assessPerformanceImpact(attr)
	if performanceImpact == PerformanceImpactHigh {
		warnings = append(warnings, TemplateActionWarning{
			Type:       WarningTypePerformance,
			Message:    "This action may have high performance impact",
			Expression: attr.Action,
			Suggestion: "Consider optimizing the action or reducing event handlers",
		})
	}

	return warnings
}

// Helper methods for validation

func (v *TemplateActionValidator) isValidActionFormat(action string) bool {
	// Basic regex for $fir.action() format
	pattern := `^\$fir\.[a-zA-Z][a-zA-Z0-9]*\(\)$`
	matched, _ := regexp.MatchString(pattern, action)
	return matched
}

func (v *TemplateActionValidator) isValidEventFormat(event string) bool {
	// Basic validation for event names
	pattern := `^[a-zA-Z][a-zA-Z0-9_-]*$`
	matched, _ := regexp.MatchString(pattern, event)
	return matched
}

func (v *TemplateActionValidator) isActionAllowed(actionType string) bool {
	for _, allowed := range v.config.AllowedActions {
		if allowed == actionType {
			return true
		}
	}
	return false
}

func (v *TemplateActionValidator) isActionForbidden(actionType string) bool {
	for _, forbidden := range v.config.ForbiddenActions {
		if forbidden == actionType {
			return true
		}
	}
	return false
}

func (v *TemplateActionValidator) hasXSSRisk(action string) bool {
	// Simple pattern matching for XSS risks
	riskPatterns := []string{"innerHTML", "outerHTML", "insertAdjacentHTML", "dangerousAction"}
	for _, pattern := range riskPatterns {
		if strings.Contains(action, pattern) {
			return true
		}
	}
	return false
}

func (v *TemplateActionValidator) hasInjectionRisk(action string) bool {
	// Simple pattern matching for injection risks
	riskPatterns := []string{"eval", "Function", "setTimeout", "setInterval"}
	for _, pattern := range riskPatterns {
		if strings.Contains(action, pattern) {
			return true
		}
	}
	return false
}

func (v *TemplateActionValidator) isDeprecatedPattern(action string) bool {
	// Check for deprecated patterns
	deprecatedPatterns := []string{"$fir.old", "$legacy"}
	for _, pattern := range deprecatedPatterns {
		if strings.Contains(action, pattern) {
			return true
		}
	}
	return false
}

// optimizeActions optimizes a collection of validated actions
func (v *TemplateActionValidator) optimizeActions(actions []ValidatedAction) []ValidatedAction {
	if !v.config.EnableActionOptimization {
		return actions
	}

	optimized := make([]ValidatedAction, 0, len(actions))

	// Simple optimization: remove duplicate actions
	seen := make(map[string]bool)
	for _, action := range actions {
		key := action.OriginalExpression
		if !seen[key] {
			seen[key] = true
			optimized = append(optimized, action)
		}
	}

	return optimized
}

// optimizeActionExpression optimizes a single action expression
func (v *TemplateActionValidator) optimizeActionExpression(attr *firattr.ParsedAttribute) string {
	if !v.config.EnableActionRewrite {
		return attr.Action
	}

	// Simple optimization: remove unnecessary whitespace
	optimized := strings.TrimSpace(attr.Action)
	optimized = regexp.MustCompile(`\s+`).ReplaceAllString(optimized, " ")

	return optimized
}

// validateTemplateConstraints validates overall template constraints
func (v *TemplateActionValidator) validateTemplateConstraints(result *TemplateActionResult) {
	// Check maximum actions per template
	if len(result.Actions) > v.config.MaxActionsPerTemplate {
		result.Errors = append(result.Errors, TemplateActionError{
			Type:       ErrorTypeValidation,
			Message:    fmt.Sprintf("Template exceeds maximum allowed actions (%d > %d)", len(result.Actions), v.config.MaxActionsPerTemplate),
			Severity:   "error",
			Suggestion: "Reduce the number of actions or increase the limit",
		})
		result.IsValid = false
	}
}

// Utility methods

func (v *TemplateActionValidator) generateCacheKey(templatePath, content string) string {
	// Simple cache key generation - in production, use a proper hash
	return fmt.Sprintf("template_actions:%s:%d", templatePath, len(content))
}

func (v *TemplateActionValidator) updateMetrics(duration time.Duration) {
	v.metrics.mu.Lock()
	defer v.metrics.mu.Unlock()

	v.metrics.ValidationCount++
	v.metrics.ProcessingTime += duration
	v.metrics.AverageProcessingTime = time.Duration(int64(v.metrics.ProcessingTime) / v.metrics.ValidationCount)
	v.metrics.LastValidationTime = time.Now()
}

func (v *TemplateActionValidator) incrementCacheHits() {
	v.metrics.mu.Lock()
	defer v.metrics.mu.Unlock()
	v.metrics.CacheHits++
}

func (v *TemplateActionValidator) incrementCacheMisses() {
	v.metrics.mu.Lock()
	defer v.metrics.mu.Unlock()
	v.metrics.CacheMisses++
}

// GetMetrics returns current validation metrics
func (v *TemplateActionValidator) GetMetrics() TemplateActionMetrics {
	v.metrics.mu.RLock()
	defer v.metrics.mu.RUnlock()

	// Return a copy without the mutex
	return TemplateActionMetrics{
		ValidationCount:       v.metrics.ValidationCount,
		ValidationErrors:      v.metrics.ValidationErrors,
		CacheHits:             v.metrics.CacheHits,
		CacheMisses:           v.metrics.CacheMisses,
		ProcessingTime:        v.metrics.ProcessingTime,
		AverageProcessingTime: v.metrics.AverageProcessingTime,
		LastValidationTime:    v.metrics.LastValidationTime,
	}
}

// ClearCache clears the validation cache
func (v *TemplateActionValidator) ClearCache() {
	if v.cache != nil {
		v.cache.Flush()
	}
}
