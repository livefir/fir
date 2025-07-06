package services

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	firHttp "github.com/livefir/fir/internal/http"
	"github.com/livefir/fir/pubsub"
)

// RouteEventServiceAdapter adapts the new EventService to work with the existing route system
type RouteEventServiceAdapter struct {
	eventService EventService
	requestAdapter *firHttp.RequestAdapter
}

// NewRouteEventServiceAdapter creates a new adapter for integrating EventService with routes
func NewRouteEventServiceAdapter(eventService EventService) *RouteEventServiceAdapter {
	return &RouteEventServiceAdapter{
		eventService: eventService,
		requestAdapter: firHttp.NewRequestAdapter(nil), // PathParams extraction handled elsewhere
	}
}

// ProcessRouteEvent processes an event using the new event service architecture
func (a *RouteEventServiceAdapter) ProcessRouteEvent(
	ctx context.Context,
	eventID string,
	sessionID string,
	target *string,
	elementKey *string,
	params map[string]interface{},
	r *http.Request,
) (*EventResponse, error) {
	// Parse the HTTP request into our abstracted model
	requestModel, err := a.requestAdapter.ParseRequest(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}

	// Build the event request
	eventRequest := &EventRequest{
		ID:           strings.ToLower(eventID),
		Target:       target,
		ElementKey:   elementKey,
		SessionID:    sessionID,
		Context:      ctx,
		Params:       params,
		RequestModel: requestModel,
	}

	// Process the event through the service layer
	return a.eventService.ProcessEvent(ctx, *eventRequest)
}

// LegacyOnEventFuncAdapter adapts a legacy OnEventFunc to work with the new EventHandler interface
type LegacyOnEventFuncAdapter struct {
	eventID   string
	routeID   string
	onEventFunc interface{} // OnEventFunc - using interface{} to avoid import cycle
}

// NewLegacyOnEventFuncAdapter creates an adapter for legacy OnEventFunc handlers
func NewLegacyOnEventFuncAdapter(eventID, routeID string, onEventFunc interface{}) *LegacyOnEventFuncAdapter {
	return &LegacyOnEventFuncAdapter{
		eventID:     eventID,
		routeID:     routeID,
		onEventFunc: onEventFunc,
	}
}

// Handle implements the EventHandler interface
func (a *LegacyOnEventFuncAdapter) Handle(ctx context.Context, req EventRequest) (*EventResponse, error) {
	// This will be implemented when we have access to RouteContext
	// For now, return a basic response indicating that legacy handlers need special handling
	return nil, fmt.Errorf("legacy handler adaptation requires RouteContext - use ProcessLegacyEvent instead")
}

// GetEventID returns the event ID this handler processes
func (a *LegacyOnEventFuncAdapter) GetEventID() string {
	return a.eventID
}

// GetRouteID returns the route ID this handler belongs to
func (a *LegacyOnEventFuncAdapter) GetRouteID() string {
	return a.routeID
}

// EventServiceIntegration provides integration helpers for the route system
type EventServiceIntegration struct {
	eventService EventService
	adapter      *RouteEventServiceAdapter
}

// NewEventServiceIntegration creates a new integration helper
func NewEventServiceIntegration(eventService EventService) *EventServiceIntegration {
	return &EventServiceIntegration{
		eventService: eventService,
		adapter:      NewRouteEventServiceAdapter(eventService),
	}
}

// RegisterLegacyHandler registers a legacy OnEventFunc with the new event service
func (i *EventServiceIntegration) RegisterLegacyHandler(eventID, routeID string, onEventFunc interface{}) error {
	// Create an adapter for the legacy handler
	adapter := NewLegacyOnEventFuncAdapter(eventID, routeID, onEventFunc)
	
	// Register it with the event service
	// Note: The actual handler execution will need to be handled specially
	// since it requires RouteContext
	return i.eventService.RegisterHandler(fmt.Sprintf("%s:%s", routeID, eventID), adapter)
}

// GetAdapter returns the route event service adapter
func (i *EventServiceIntegration) GetAdapter() *RouteEventServiceAdapter {
	return i.adapter
}

// GetEventService returns the underlying event service
func (i *EventServiceIntegration) GetEventService() EventService {
	return i.eventService
}

// ProcessLegacyEvent processes an event using legacy OnEventFunc handlers
// This method provides backward compatibility during the migration
func (i *EventServiceIntegration) ProcessLegacyEvent(
	ctx context.Context,
	eventID, routeID string,
	onEventFunc interface{}, // OnEventFunc
	routeContext interface{}, // RouteContext - using interface{} to avoid import cycle
) (*EventResponse, error) {
	// This method would handle the legacy event processing
	// It would call the OnEventFunc and convert the results to EventResponse
	// Implementation would depend on having access to the route types
	
	// For now, return an error indicating this needs to be implemented in the route package
	return nil, fmt.Errorf("legacy event processing not yet implemented - requires route package integration")
}

// ConvertPubSubEventsToResponse converts pubsub events to our EventResponse format
func ConvertPubSubEventsToResponse(pubsubEvents []pubsub.Event) *EventResponse {
	response := NewEventResponseBuilder().
		WithStatus(http.StatusOK).
		Build()

	// Add all pubsub events to the response
	response.PubSubEvents = append(response.PubSubEvents, pubsubEvents...)

	return response
}

// ConvertEventResponseToPubSubEvents extracts pubsub events from an EventResponse
func ConvertEventResponseToPubSubEvents(response *EventResponse) []pubsub.Event {
	if response == nil {
		return nil
	}
	return response.PubSubEvents
}
