package fir

import (
	"fmt"
	"sort" // Import sort
	"strings"

	"github.com/livefir/fir/internal/logger"
	"golang.org/x/net/html"
)

type ActionType int

const (
	TypeLive ActionType = iota
	TypeRefresh
	TypeRemove
	TypeAppend       // Add new type
	TypePrepend      // Add new type
	TypeRemoveParent // Add new type
	TypeActionPrefix // Represents x-fir-action-* attributes
	TypeUnknown
)

// ActionInfo holds parsed data from an x-fir-* attribute.
type ActionInfo struct {
	AttrName   string   // Original full attribute name (e.g., "x-fir-live", "x-fir-action-doSave")
	ActionName string   // Parsed action name (e.g., "live", "refresh", "action-doSave")
	Params     []string // Parsed parameters like ["loading", "visible"]
	Value      string   // The attribute's value
}

// ActionHandler defines the interface for processing a specific x-fir-* action.
type ActionHandler interface {
	// Name returns the core action name (e.g., "live", "refresh") or prefix identifier ("action-").
	Name() string
	// Precedence returns the processing priority (lower value means higher priority).
	Precedence() int
	// Translate processes the action's value and returns the translated attribute string(s).
	// It receives the specific ActionInfo for the attribute and the map of x-fir-action-* values found on the node.
	Translate(info ActionInfo, actionsMap map[string]string) (string, error)
}

// actionRegistry holds registered action handlers.
var actionRegistry = make(map[string]ActionHandler)

// RegisterActionHandler adds an action handler to the registry.
// Panics if the handler name is already registered.
func RegisterActionHandler(handler ActionHandler) {
	name := handler.Name()
	if _, exists := actionRegistry[name]; exists {
		// Allow registering handlers with the same base name if they handle prefixes (like append:)
		// The lookup logic will need to handle this.
		// For now, we keep the simple check, assuming base names are unique for direct handlers.
		// Update: Reverted to simple check as prefix handling is managed by lookup.
		panic(fmt.Sprintf("action handler already registered for name: %s", name))
	}
	actionRegistry[name] = handler
}

// --- Concrete Handler Implementations ---

// LiveActionHandler handles x-fir-live
type LiveActionHandler struct{}

func (h *LiveActionHandler) Name() string    { return "live" }
func (h *LiveActionHandler) Precedence() int { return 10 } // Highest precedence
func (h *LiveActionHandler) Translate(info ActionInfo, actionsMap map[string]string) (string, error) {
	// TranslateRenderExpression needs the value and the collected actionsMap
	return TranslateRenderExpression(info.Value, actionsMap)
}

// RefreshActionHandler handles x-fir-refresh
type RefreshActionHandler struct{}

func (h *RefreshActionHandler) Name() string    { return "refresh" }
func (h *RefreshActionHandler) Precedence() int { return 20 }
func (h *RefreshActionHandler) Translate(info ActionInfo, actionsMap map[string]string) (string, error) {
	// TranslateEventExpression needs the value and the action name ("refresh")
	return TranslateEventExpression(info.Value, "$fir.replace()", "")
}

// RemoveActionHandler handles x-fir-remove
type RemoveActionHandler struct{}

func (h *RemoveActionHandler) Name() string    { return "remove" }
func (h *RemoveActionHandler) Precedence() int { return 30 }
func (h *RemoveActionHandler) Translate(info ActionInfo, actionsMap map[string]string) (string, error) {
	// TranslateEventExpression needs the value and the action name ("remove")
	return TranslateEventExpression(info.Value, "$fir.removeEl()", "", "nohtml")
}

// AppendActionHandler handles x-fir-append:target
type AppendActionHandler struct{}

func (h *AppendActionHandler) Name() string    { return "append" } // Base name
func (h *AppendActionHandler) Precedence() int { return 50 }
func (h *AppendActionHandler) Translate(info ActionInfo, actionsMap map[string]string) (string, error) {
	// Expect the templateValue to be the first parameter
	if len(info.Params) == 0 {
		return "", fmt.Errorf("missing target template name parameter for append action: '%s'", info.AttrName)
	}
	templateValue := info.Params[0]
	if templateValue == "" {
		// This check might be redundant if the parser ensures non-empty params, but good for safety.
		return "", fmt.Errorf("empty target template name parameter for append action: '%s'", info.AttrName)
	}

	// TranslateEventExpression needs the value, the JS action, and the templateValue
	return TranslateEventExpression(info.Value, "$fir.appendEl()", templateValue)
}

// PrependActionHandler handles x-fir-prepend:target
type PrependActionHandler struct{}

func (h *PrependActionHandler) Name() string    { return "prepend" } // Base name
func (h *PrependActionHandler) Precedence() int { return 60 }
func (h *PrependActionHandler) Translate(info ActionInfo, actionsMap map[string]string) (string, error) {
	// Expect the templateValue to be the first parameter
	if len(info.Params) == 0 {
		return "", fmt.Errorf("missing target template name parameter for prepend action: '%s'", info.AttrName)
	}
	templateValue := info.Params[0]
	if templateValue == "" {
		return "", fmt.Errorf("empty target template name parameter for prepend action: '%s'", info.AttrName)
	}

	// TranslateEventExpression needs the value, the JS action, and the templateValue
	return TranslateEventExpression(info.Value, "$fir.prependEl()", templateValue)
}

// RemoveParentActionHandler handles x-fir-remove-parent
type RemoveParentActionHandler struct{}

func (h *RemoveParentActionHandler) Name() string    { return "remove-parent" }
func (h *RemoveParentActionHandler) Precedence() int { return 40 }
func (h *RemoveParentActionHandler) Translate(info ActionInfo, actionsMap map[string]string) (string, error) {
	// TranslateEventExpression needs the value and the action name ("remove-parent")
	// Assuming the JS function is $fir.removeParentEl()
	return TranslateEventExpression(info.Value, "$fir.removeParentEl()", "", "nohtml")
}

// ResetActionHandler handles x-fir-reset
type ResetActionHandler struct{}

func (h *ResetActionHandler) Name() string    { return "reset" }
func (h *ResetActionHandler) Precedence() int { return 35 }
func (h *ResetActionHandler) Translate(info ActionInfo, actionsMap map[string]string) (string, error) {
	// TranslateEventExpression needs the value and the action
	// For reset, we use $el.reset() and force nohtml modifier
	return TranslateEventExpression(info.Value, "$el.reset()", "", "nohtml")
}

// ToggleDisabledActionHandler handles x-fir-toggle-disabled
type ToggleDisabledActionHandler struct{}

func (h *ToggleDisabledActionHandler) Name() string    { return "toggle-disabled" }
func (h *ToggleDisabledActionHandler) Precedence() int { return 34 }
func (h *ToggleDisabledActionHandler) Translate(info ActionInfo, actionsMap map[string]string) (string, error) {
	// For toggle-disabled, we use $fir.toggleDisabled() and force nohtml modifier
	// The toggleDisabled function automatically handles enabling/disabling based on event state
	return TranslateEventExpression(info.Value, "$fir.toggleDisabled()", "", "nohtml")
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

	// For dispatch, we use the built dispatch call and force nohtml modifier
	return TranslateEventExpression(info.Value, dispatchCall, template, "nohtml")
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
	parser, err := getRenderExpressionParser()
	if err != nil {
		return "", fmt.Errorf("error creating parser: %w", err)
	}

	parsed, err := parseRenderExpression(parser, input)
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

	// Use TranslateEventExpression to translate the events, forcing nohtml modifier
	return TranslateEventExpression(info.Value, actionValue, "", "nohtml")
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

// Register default handlers
func init() {
	RegisterActionHandler(&LiveActionHandler{})
	RegisterActionHandler(&RefreshActionHandler{})
	RegisterActionHandler(&RemoveActionHandler{})
	RegisterActionHandler(&AppendActionHandler{})  // Register the append handler
	RegisterActionHandler(&PrependActionHandler{}) // Register the prepend handler
	RegisterActionHandler(&RemoveParentActionHandler{})
	RegisterActionHandler(&ResetActionHandler{}) // Register the reset handler
	RegisterActionHandler(&ToggleDisabledActionHandler{})
	RegisterActionHandler(&TriggerActionHandler{})  // Register the trigger handler
	RegisterActionHandler(&DispatchActionHandler{}) // Register the dispatch handler
	RegisterActionHandler(&ActionPrefixHandler{})   // Register the prefix handler
}

// Helper struct for processing within processRenderAttributes
type collectedAction struct {
	Handler ActionHandler
	Info    ActionInfo
}

// Helper to parse the multi-line string potentially returned by translators.
func parseTranslatedString(translated string) []html.Attribute {
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
		} else {
			logger.Warnf("Skipping malformed translated attribute line: %s", line)
		}
	}
	return attrs
}

// Sorts collected actions by precedence
func sortActionsByPrecedence(actions []collectedAction) {
	sort.Slice(actions, func(i, j int) bool {
		return actions[i].Handler.Precedence() < actions[j].Handler.Precedence()
	})
}
