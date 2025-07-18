package actions

import (
	"fmt"
	"sort"
	"strings"

	"github.com/livefir/fir/internal/firattr"
	"golang.org/x/net/html"
)

// ActionInfo holds parsed data from an x-fir-* attribute.
type ActionInfo struct {
	AttrName   string   // Original full attribute name (e.g., "x-fir-refresh", "x-fir-action-doSave")
	ActionName string   // Parsed action name (e.g., "refresh", "action-doSave")
	Params     []string // Parsed parameters like ["loading", "visible"]
	Value      string   // The attribute's value
}

// ActionHandler defines the interface for processing a specific x-fir-* action.
type ActionHandler interface {
	// Name returns the core action name (e.g., "refresh") or prefix identifier ("action-").
	Name() string
	// Precedence returns the processing priority (lower value means higher priority).
	Precedence() int
	// Translate processes the action's value and returns the translated attribute string(s).
	// It receives the specific ActionInfo for the attribute and the map of x-fir-action-* values found on the node.
	Translate(info ActionInfo, actionsMap map[string]string) (string, error)
}

// Registry holds registered action handlers.
var Registry = make(map[string]ActionHandler)

// RegisterHandler adds an action handler to the registry.
// Panics if the handler name is already registered.
func RegisterHandler(handler ActionHandler) {
	name := handler.Name()
	if _, exists := Registry[name]; exists {
		// Allow registering handlers with the same base name if they handle prefixes (like append:)
		// The lookup logic will need to handle this.
		// For now, we keep the simple check, assuming base names are unique for direct handlers.
		// Update: Reverted to simple check as prefix handling is managed by lookup.
		panic(fmt.Sprintf("action handler already registered for name: %s", name))
	}
	Registry[name] = handler
}

// --- Concrete Handler Implementations ---

// RefreshActionHandler handles x-fir-refresh
type RefreshActionHandler struct{}

func (h *RefreshActionHandler) Name() string    { return "refresh" }
func (h *RefreshActionHandler) Precedence() int { return 20 }
func (h *RefreshActionHandler) Translate(info ActionInfo, actionsMap map[string]string) (string, error) {
	// TranslateEventExpression needs the value and the action name ("refresh")
	return firattr.TranslateEventExpression(info.Value, "$fir.replace()", "")
}

// RemoveActionHandler handles x-fir-remove
type RemoveActionHandler struct{}

func (h *RemoveActionHandler) Name() string    { return "remove" }
func (h *RemoveActionHandler) Precedence() int { return 30 }
func (h *RemoveActionHandler) Translate(info ActionInfo, actionsMap map[string]string) (string, error) {
	// TranslateEventExpression needs the value and the action name ("remove")
	return firattr.TranslateEventExpression(info.Value, "$fir.removeEl()", "")
}

// AppendActionHandler handles x-fir-append:target
type AppendActionHandler struct{}

func (h *AppendActionHandler) Name() string    { return "append" } // Base name
func (h *AppendActionHandler) Precedence() int { return 50 }
func (h *AppendActionHandler) Translate(info ActionInfo, actionsMap map[string]string) (string, error) {
	// Use the first parameter as template if provided, otherwise use empty string to allow extracted template
	templateValue := ""
	if len(info.Params) > 0 && info.Params[0] != "" {
		templateValue = info.Params[0]
	}

	// TranslateEventExpression needs the value, the JS action, and the templateValue
	return firattr.TranslateEventExpression(info.Value, "$fir.appendEl()", templateValue)
}

// PrependActionHandler handles x-fir-prepend:target
type PrependActionHandler struct{}

func (h *PrependActionHandler) Name() string    { return "prepend" } // Base name
func (h *PrependActionHandler) Precedence() int { return 60 }
func (h *PrependActionHandler) Translate(info ActionInfo, actionsMap map[string]string) (string, error) {
	// Use the first parameter as template if provided, otherwise use empty string to allow extracted template
	templateValue := ""
	if len(info.Params) > 0 && info.Params[0] != "" {
		templateValue = info.Params[0]
	}

	// TranslateEventExpression needs the value, the JS action, and the templateValue
	return firattr.TranslateEventExpression(info.Value, "$fir.prependEl()", templateValue)
}

// RemoveParentActionHandler handles x-fir-remove-parent
type RemoveParentActionHandler struct{}

func (h *RemoveParentActionHandler) Name() string    { return "remove-parent" }
func (h *RemoveParentActionHandler) Precedence() int { return 40 }
func (h *RemoveParentActionHandler) Translate(info ActionInfo, actionsMap map[string]string) (string, error) {
	// TranslateEventExpression needs the value and the action name ("remove-parent")
	// Assuming the JS function is $fir.removeParentEl()
	return firattr.TranslateEventExpression(info.Value, "$fir.removeParentEl()", "")
}

// ResetActionHandler handles x-fir-reset
type ResetActionHandler struct{}

func (h *ResetActionHandler) Name() string    { return "reset" }
func (h *ResetActionHandler) Precedence() int { return 35 }
func (h *ResetActionHandler) Translate(info ActionInfo, actionsMap map[string]string) (string, error) {
	// TranslateEventExpression needs the value and the action
	// For reset, we use $el.reset()
	return firattr.TranslateEventExpression(info.Value, "$el.reset()", "")
}

// ToggleDisabledActionHandler handles x-fir-toggle-disabled
type ToggleDisabledActionHandler struct{}

func (h *ToggleDisabledActionHandler) Name() string    { return "toggle-disabled" }
func (h *ToggleDisabledActionHandler) Precedence() int { return 34 }
func (h *ToggleDisabledActionHandler) Translate(info ActionInfo, actionsMap map[string]string) (string, error) {
	// For toggle-disabled, we use $fir.toggleDisabled()
	// The toggleDisabled function automatically handles enabling/disabling based on event state
	return firattr.TranslateEventExpression(info.Value, "$fir.toggleDisabled()", "")
}

// ToggleClassActionHandler handles x-fir-toggleClass:class or x-fir-toggleClass:[class1,class2]
type ToggleClassActionHandler struct{}

func (h *ToggleClassActionHandler) Name() string    { return "toggleClass" }
func (h *ToggleClassActionHandler) Precedence() int { return 33 }
func (h *ToggleClassActionHandler) Translate(info ActionInfo, actionsMap map[string]string) (string, error) {
	var classNames []string

	// Parse class names from the Params field
	if len(info.Params) > 0 {
		classNames = info.Params
	}

	if len(classNames) == 0 {
		return "", fmt.Errorf("no class names specified for toggleClass action: '%s'", info.AttrName)
	}

	// Build the JavaScript function call
	var jsArgs []string
	for _, className := range classNames {
		jsArgs = append(jsArgs, fmt.Sprintf("'%s'", className))
	}
	jsAction := fmt.Sprintf("$fir.toggleClass(%s)", strings.Join(jsArgs, ","))

	// TranslateEventExpression since we're just toggling classes
	return firattr.TranslateEventExpression(info.Value, jsAction, "")
}

// DispatchActionHandler handles x-fir-dispatch:[param1,param2,...]
type DispatchActionHandler struct{}

func (h *DispatchActionHandler) Name() string    { return "dispatch" }
func (h *DispatchActionHandler) Precedence() int { return 33 }
func (h *DispatchActionHandler) Translate(info ActionInfo, actionsMap map[string]string) (string, error) {
	// Check that we have parameters for dispatch
	if len(info.Params) == 0 {
		return "", fmt.Errorf("dispatch requires at least one parameter")
	}

	// Check for empty parameters
	for i, param := range info.Params {
		if strings.TrimSpace(param) == "" {
			return "", fmt.Errorf("empty parameter found in dispatch at position %d", i)
		}
	}

	// Build the $dispatch function call
	dispatchCall := h.buildDispatchCall(info.Params)

	// Parse the expression to extract template
	template, err := h.extractTemplate(info.Value)
	if err != nil {
		return "", fmt.Errorf("error parsing dispatch expression: %w", err)
	}

	// For dispatch, we use the built dispatch call
	return firattr.TranslateEventExpression(info.Value, dispatchCall, template)
}

// buildDispatchCall creates the $dispatch() function call with quoted parameters
func (h *DispatchActionHandler) buildDispatchCall(params []string) string {
	var quotedParams []string
	for _, param := range params {
		quotedParams = append(quotedParams, "'"+param+"'")
	}
	return "$dispatch(" + strings.Join(quotedParams, ",") + ")"
}

// extractTemplate parses the expression and extracts the template from the binding target
func (h *DispatchActionHandler) extractTemplate(input string) (string, error) {
	parsed, err := firattr.ParseRenderExpression(input)
	if err != nil {
		return "", fmt.Errorf("error parsing render expression: %w", err)
	}

	// Look for template in any binding's target
	for _, expr := range parsed.Expressions {
		for _, binding := range expr.Bindings {
			if binding.Target != nil && binding.Target.Template != "" {
				return binding.Target.Template, nil
			}
		}
	}

	return "", nil // No template found
}

// TriggerActionHandler handles x-fir-runjs:actionName
type TriggerActionHandler struct{}

func (h *TriggerActionHandler) Name() string    { return "runjs" }
func (h *TriggerActionHandler) Precedence() int { return 32 }
func (h *TriggerActionHandler) Translate(info ActionInfo, actionsMap map[string]string) (string, error) {
	// Check that we have parameters for runjs (the action name)
	if len(info.Params) == 0 {
		return "", fmt.Errorf("runjs requires exactly one parameter (action name)")
	}
	if len(info.Params) > 1 {
		return "", fmt.Errorf("runjs accepts only one parameter (action name), got %d", len(info.Params))
	}

	actionName := strings.TrimSpace(info.Params[0])
	if actionName == "" {
		return "", fmt.Errorf("runjs action name parameter cannot be empty")
	}

	// Look up the action in the actionsMap
	actionValue, exists := actionsMap[actionName]
	if !exists {
		return "", fmt.Errorf("runjs action '%s' not found in actions map", actionName)
	}

	if strings.TrimSpace(actionValue) == "" {
		return "", fmt.Errorf("runjs action '%s' value cannot be empty", actionName)
	}

	// Use TranslateEventExpression to translate the events
	return firattr.TranslateEventExpression(info.Value, actionValue, "")
}

// ActionPrefixHandler handles x-fir-js:* (doesn't translate directly, just used for collection)
type ActionPrefixHandler struct{}

func (h *ActionPrefixHandler) Name() string    { return "js" } // Special prefix identifier for JavaScript actions
func (h *ActionPrefixHandler) Precedence() int { return 100 }  // Lowest precedence, processed first for collection
func (h *ActionPrefixHandler) Translate(info ActionInfo, actionsMap map[string]string) (string, error) {
	// This handler doesn't produce translated attributes itself.
	// Its presence is mainly for identification during the collection phase and removal.
	return "", nil
}

// RedirectActionHandler handles x-fir-redirect
type RedirectActionHandler struct{}

func (h *RedirectActionHandler) Name() string    { return "redirect" }
func (h *RedirectActionHandler) Precedence() int { return 90 } // Higher precedence than js actions
func (h *RedirectActionHandler) Translate(info ActionInfo, actionsMap map[string]string) (string, error) {
	// Extract URL from first parameter, default to '/' if not provided
	var url = "'/'"
	if len(info.Params) > 0 && strings.TrimSpace(info.Params[0]) != "" {
		paramUrl := strings.TrimSpace(info.Params[0])

		// Convert parameter to URL path
		// If it doesn't start with '/', add it to make it a proper path
		if !strings.HasPrefix(paramUrl, "/") {
			paramUrl = "/" + paramUrl
		}

		// Ensure the URL is properly quoted for JavaScript
		url = fmt.Sprintf("'%s'", paramUrl)
	}

	// Create the redirect function call with the URL
	jsAction := fmt.Sprintf("$fir.redirect(%s)", url)

	// Use TranslateEventExpression to translate the events
	return firattr.TranslateEventExpression(info.Value, jsAction, "")
}

// Helper struct for processing within processRenderAttributes
type CollectedAction struct {
	Handler ActionHandler
	Info    ActionInfo
}

// Helper to parse the multi-line string potentially returned by translators.
func ParseTranslatedString(translated string) []html.Attribute {
	var attrs []html.Attribute
	lines := strings.Split(translated, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			// Remove surrounding quotes, be careful not to remove internal quotes
			val := strings.TrimPrefix(parts[1], `"`)
			val = strings.TrimSuffix(val, `"`)
			attrs = append(attrs, html.Attribute{Key: key, Val: val})
		}
	}
	return attrs
}

// SortActionsByPrecedence sorts collected actions by precedence
func SortActionsByPrecedence(actions []CollectedAction) {
	sort.Slice(actions, func(i, j int) bool {
		return actions[i].Handler.Precedence() < actions[j].Handler.Precedence()
	})
}

// Initialize default handlers
func init() {
	RegisterHandler(&RefreshActionHandler{})
	RegisterHandler(&RemoveActionHandler{})
	RegisterHandler(&AppendActionHandler{})  // Register the append handler
	RegisterHandler(&PrependActionHandler{}) // Register the prepend handler
	RegisterHandler(&RemoveParentActionHandler{})
	RegisterHandler(&ResetActionHandler{}) // Register the reset handler
	RegisterHandler(&ToggleDisabledActionHandler{})
	RegisterHandler(&ToggleClassActionHandler{}) // Register the toggleClass handler
	RegisterHandler(&RedirectActionHandler{})    // Register the redirect handler
	RegisterHandler(&TriggerActionHandler{})     // Register the trigger handler
	RegisterHandler(&DispatchActionHandler{})    // Register the dispatch handler
	RegisterHandler(&ActionPrefixHandler{})      // Register the prefix handler
}
