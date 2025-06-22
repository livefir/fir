package event

import (
	"fmt"
	"sync"
)

// EventRegistry provides a centralized registry for managing route events
// with thread-safe operations and debug introspection capabilities
type EventRegistry interface {
	// Register adds an event handler for a specific route and event ID
	Register(routeID, eventID string, handler interface{}) error

	// Get retrieves an event handler for a specific route and event ID
	Get(routeID, eventID string) (interface{}, bool)

	// GetRouteEvents returns all event handlers for a specific route
	GetRouteEvents(routeID string) map[string]interface{}

	// GetAllEvents returns all registered events organized by route
	// This is primarily for debug introspection
	GetAllEvents() map[string]map[string]interface{}

	// Remove removes an event handler for a specific route and event ID
	Remove(routeID, eventID string) bool

	// RemoveRoute removes all event handlers for a specific route
	RemoveRoute(routeID string) bool

	// ListRoutes returns all route IDs that have registered events
	ListRoutes() []string

	// ListEventIDs returns all event IDs for a specific route
	ListEventIDs(routeID string) []string
}

// registry is the internal implementation of EventRegistry
type registry struct {
	// events maps routeID -> eventID -> handler
	events map[string]map[string]interface{}
	mu     sync.RWMutex
}

// NewEventRegistry creates a new thread-safe event registry
func NewEventRegistry() EventRegistry {
	return &registry{
		events: make(map[string]map[string]interface{}),
	}
}

// Register adds an event handler for a specific route and event ID
func (r *registry) Register(routeID, eventID string, handler interface{}) error {
	if routeID == "" {
		return fmt.Errorf("routeID cannot be empty")
	}
	if eventID == "" {
		return fmt.Errorf("eventID cannot be empty")
	}
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.events[routeID] == nil {
		r.events[routeID] = make(map[string]interface{})
	}

	r.events[routeID][eventID] = handler
	return nil
}

// Get retrieves an event handler for a specific route and event ID
func (r *registry) Get(routeID, eventID string) (interface{}, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	routeEvents, exists := r.events[routeID]
	if !exists {
		return nil, false
	}

	handler, exists := routeEvents[eventID]
	return handler, exists
}

// GetRouteEvents returns all event handlers for a specific route
func (r *registry) GetRouteEvents(routeID string) map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	routeEvents, exists := r.events[routeID]
	if !exists {
		return make(map[string]interface{})
	}

	// Return a copy to prevent external modification
	result := make(map[string]interface{})
	for eventID, handler := range routeEvents {
		result[eventID] = handler
	}
	return result
}

// GetAllEvents returns all registered events organized by route
func (r *registry) GetAllEvents() map[string]map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a deep copy to prevent external modification
	result := make(map[string]map[string]interface{})
	for routeID, routeEvents := range r.events {
		result[routeID] = make(map[string]interface{})
		for eventID, handler := range routeEvents {
			result[routeID][eventID] = handler
		}
	}
	return result
}

// Remove removes an event handler for a specific route and event ID
func (r *registry) Remove(routeID, eventID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	routeEvents, exists := r.events[routeID]
	if !exists {
		return false
	}

	_, exists = routeEvents[eventID]
	if !exists {
		return false
	}

	delete(routeEvents, eventID)

	// Clean up empty route map
	if len(routeEvents) == 0 {
		delete(r.events, routeID)
	}

	return true
}

// RemoveRoute removes all event handlers for a specific route
func (r *registry) RemoveRoute(routeID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.events[routeID]
	if !exists {
		return false
	}

	delete(r.events, routeID)
	return true
}

// ListRoutes returns all route IDs that have registered events
func (r *registry) ListRoutes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	routes := make([]string, 0, len(r.events))
	for routeID := range r.events {
		routes = append(routes, routeID)
	}
	return routes
}

// ListEventIDs returns all event IDs for a specific route
func (r *registry) ListEventIDs(routeID string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	routeEvents, exists := r.events[routeID]
	if !exists {
		return []string{}
	}

	eventIDs := make([]string, 0, len(routeEvents))
	for eventID := range routeEvents {
		eventIDs = append(eventIDs, eventID)
	}
	return eventIDs
}
