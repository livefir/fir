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
	return TranslateEventExpression(info.Value, "$fir.replace()")
}

// RemoveActionHandler handles x-fir-remove
type RemoveActionHandler struct{}

func (h *RemoveActionHandler) Name() string    { return "remove" }
func (h *RemoveActionHandler) Precedence() int { return 30 }
func (h *RemoveActionHandler) Translate(info ActionInfo, actionsMap map[string]string) (string, error) {
	// TranslateEventExpression needs the value and the action name ("remove")
	return TranslateEventExpression(info.Value, "$fir.removeEl()")
}

// RemoveParentActionHandler handles x-fir-remove-parent
type RemoveParentActionHandler struct{}

func (h *RemoveParentActionHandler) Name() string    { return "remove-parent" }
func (h *RemoveParentActionHandler) Precedence() int { return 40 }
func (h *RemoveParentActionHandler) Translate(info ActionInfo, actionsMap map[string]string) (string, error) {
	// TranslateEventExpression needs the value and the action name ("remove-parent")
	// Assuming the JS function is $fir.removeParentEl()
	return TranslateEventExpression(info.Value, "$fir.removeParentEl()")
}

// ActionPrefixHandler handles x-fir-action-* (doesn't translate directly, just used for collection)
type ActionPrefixHandler struct{}

func (h *ActionPrefixHandler) Name() string    { return "action-" } // Special prefix identifier
func (h *ActionPrefixHandler) Precedence() int { return 100 }       // Lowest precedence, processed first for collection
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
	RegisterActionHandler(&RemoveParentActionHandler{}) // Register the new handler
	RegisterActionHandler(&ActionPrefixHandler{})       // Register the prefix handler
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
