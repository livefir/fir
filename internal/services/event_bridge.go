package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	firHttp "github.com/livefir/fir/internal/http"
	"github.com/livefir/fir/pubsub"
)

// EventBridge bridges the current route event system with the new EventService
type EventBridge struct {
	eventService EventService
	integration  *EventServiceIntegration
}

// NewEventBridge creates a new bridge between old and new event systems
func NewEventBridge(eventService EventService) *EventBridge {
	return &EventBridge{
		eventService: eventService,
		integration:  NewEventServiceIntegration(eventService),
	}
}

// ProcessLegacyRouteEvent processes a legacy route event using the new service architecture
// This method converts the legacy Event struct to our EventRequest format
func (b *EventBridge) ProcessLegacyRouteEvent(
	ctx context.Context,
	legacyEvent interface{}, // Event from event.go
	request *http.Request,
) (*EventResponse, error) {
	// Extract event data using type assertion (to avoid import cycles)
	eventData := extractEventData(legacyEvent)
	
	// Parse the HTTP request
	requestModel, err := firHttp.NewRequestAdapter(nil).ParseRequest(request)
	if err != nil {
		return nil, err
	}

	// Convert legacy params (json.RawMessage) to map[string]interface{}
	params := make(map[string]interface{})
	if len(eventData.Params) > 0 {
		if err := json.Unmarshal(eventData.Params, &params); err != nil {
			// If it fails to unmarshal as object, store as raw string
			params["raw"] = string(eventData.Params)
		}
	}

	// Build EventRequest
	eventRequest := EventRequest{
		ID:           eventData.ID,
		Target:       eventData.Target,
		ElementKey:   eventData.ElementKey,
		SessionID:    getStringValue(eventData.SessionID),
		Context:      ctx,
		Params:       params,
		RequestModel: requestModel,
	}

	// Process through the new event service
	return b.eventService.ProcessEvent(ctx, eventRequest)
}

// eventData represents the extracted event information
type eventData struct {
	ID         string
	Params     json.RawMessage
	Target     *string
	ElementKey *string
	SessionID  *string
}

// extractEventData extracts event data from legacy Event struct
// Uses reflection-like approach to avoid import cycles
func extractEventData(event interface{}) eventData {
	// This is a simplified extraction - in a real implementation,
	// we'd use type assertion or reflection to extract the fields
	// For now, assuming the event follows the expected structure
	
	if eventMap, ok := event.(map[string]interface{}); ok {
		data := eventData{}
		
		if id, ok := eventMap["ID"].(string); ok {
			data.ID = id
		}
		if params, ok := eventMap["Params"].(json.RawMessage); ok {
			data.Params = params
		}
		if target, ok := eventMap["Target"].(*string); ok {
			data.Target = target
		}
		if elementKey, ok := eventMap["ElementKey"].(*string); ok {
			data.ElementKey = elementKey
		}
		if sessionID, ok := eventMap["SessionID"].(*string); ok {
			data.SessionID = sessionID
		}
		
		return data
	}

	// Fallback: return empty data
	return eventData{}
}

// getStringValue safely extracts string value from pointer
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// CreateEventServiceProvider creates a provider function for the event service
// This can be used to inject the event service into the existing route system
func CreateEventServiceProvider(
	validator EventValidator,
	publisher EventPublisher,
	logger EventLogger,
) func() EventService {
	return func() EventService {
		registry := NewInMemoryEventRegistry()
		return NewDefaultEventService(registry, validator, publisher, logger)
	}
}

// LegacyEventHandlerWrapper wraps a legacy OnEventFunc to work with EventService
type LegacyEventHandlerWrapper struct {
	eventID  string
	routeID  string
	handler  interface{} // OnEventFunc
	bridge   *EventBridge
}

// NewLegacyEventHandlerWrapper creates a wrapper for legacy event handlers
func NewLegacyEventHandlerWrapper(eventID, routeID string, handler interface{}, bridge *EventBridge) *LegacyEventHandlerWrapper {
	return &LegacyEventHandlerWrapper{
		eventID: eventID,
		routeID: routeID,
		handler: handler,
		bridge:  bridge,
	}
}

// Handle implements EventHandler interface for legacy handlers
func (w *LegacyEventHandlerWrapper) Handle(ctx context.Context, req EventRequest) (*EventResponse, error) {
	// This would require calling the legacy handler with a RouteContext
	// Since we can't create a RouteContext here without import cycles,
	// we'll need to handle this in the route package
	return nil, ErrLegacyHandlerRequiresRouteContext
}

// GetEventID returns the event ID
func (w *LegacyEventHandlerWrapper) GetEventID() string {
	return w.eventID
}

// GetRouteID returns the route ID  
func (w *LegacyEventHandlerWrapper) GetRouteID() string {
	return w.routeID
}

// Custom error for legacy handler requirements
var ErrLegacyHandlerRequiresRouteContext = fmt.Errorf("legacy handler requires RouteContext for execution")

// EventServiceConfiguration holds configuration for the event service integration
type EventServiceConfiguration struct {
	EnableDebugLogging bool
	EnableMetrics      bool
	PubSubAdapter      interface{} // pubsub.Adapter
	ValidatorConfig    map[string][]string // eventID -> required fields
}

// ConfigureEventService configures the event service with the provided options
func ConfigureEventService(config EventServiceConfiguration) EventService {
	// Create validator
	validator := NewDefaultEventValidator()
	for eventID, requiredFields := range config.ValidatorConfig {
		validator.SetRequiredFields(eventID, requiredFields)
	}

	// Create logger
	logger := NewDefaultEventLogger(config.EnableDebugLogging)

	// Create publisher (simplified - would need actual pubsub adapter)
	publisher := &NoOpEventPublisher{}

	// Create registry and service
	registry := NewInMemoryEventRegistry()
	return NewDefaultEventService(registry, validator, publisher, logger)
}

// NoOpEventPublisher is a no-op implementation for testing/fallback
type NoOpEventPublisher struct{}

func (p *NoOpEventPublisher) PublishEvent(event pubsub.Event) error {
	return nil
}

func (p *NoOpEventPublisher) PublishEvents(events []pubsub.Event) error {
	return nil
}
