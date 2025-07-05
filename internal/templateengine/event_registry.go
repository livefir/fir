package templateengine

import (
	"sync"
)

// InMemoryEventTemplateRegistry provides an in-memory implementation of EventTemplateRegistry.
// This registry is thread-safe and suitable for most use cases.
type InMemoryEventTemplateRegistry struct {
	templates map[string]EventTemplateState
	mutex     sync.RWMutex
}

// NewInMemoryEventTemplateRegistry creates a new in-memory event template registry.
func NewInMemoryEventTemplateRegistry() *InMemoryEventTemplateRegistry {
	return &InMemoryEventTemplateRegistry{
		templates: make(map[string]EventTemplateState),
	}
}

// Register implements EventTemplateRegistry interface.
func (reg *InMemoryEventTemplateRegistry) Register(eventID string, state string, templateName string) {
	reg.mutex.Lock()
	defer reg.mutex.Unlock()

	if reg.templates[eventID] == nil {
		reg.templates[eventID] = make(EventTemplateState)
	}
	reg.templates[eventID][state] = struct{}{}
}

// Get implements EventTemplateRegistry interface.
func (reg *InMemoryEventTemplateRegistry) Get(eventID string) map[string][]string {
	reg.mutex.RLock()
	defer reg.mutex.RUnlock()

	result := make(map[string][]string)
	if stateMap, exists := reg.templates[eventID]; exists {
		for state := range stateMap {
			result[state] = []string{eventID + "_" + state} // Simplified template naming
		}
	}
	return result
}

// GetByState implements EventTemplateRegistry interface.
func (reg *InMemoryEventTemplateRegistry) GetByState(eventID string, state string) []string {
	reg.mutex.RLock()
	defer reg.mutex.RUnlock()

	if stateMap, exists := reg.templates[eventID]; exists {
		if _, stateExists := stateMap[state]; stateExists {
			return []string{eventID + "_" + state} // Simplified template naming
		}
	}
	return nil
}

// Clear implements EventTemplateRegistry interface.
func (reg *InMemoryEventTemplateRegistry) Clear() {
	reg.mutex.Lock()
	defer reg.mutex.Unlock()

	reg.templates = make(map[string]EventTemplateState)
}

// GetAll implements EventTemplateRegistry interface.
func (reg *InMemoryEventTemplateRegistry) GetAll() EventTemplateMap {
	reg.mutex.RLock()
	defer reg.mutex.RUnlock()

	result := make(EventTemplateMap)
	for eventID, stateMap := range reg.templates {
		result[eventID] = make(EventTemplateState)
		for state := range stateMap {
			result[eventID][state] = struct{}{}
		}
	}
	return result
}

// Merge implements EventTemplateRegistry interface.
func (reg *InMemoryEventTemplateRegistry) Merge(other EventTemplateRegistry) {
	reg.mutex.Lock()
	defer reg.mutex.Unlock()

	otherTemplates := other.GetAll()
	for eventID, stateMap := range otherTemplates {
		if reg.templates[eventID] == nil {
			reg.templates[eventID] = make(EventTemplateState)
		}
		for state := range stateMap {
			reg.templates[eventID][state] = struct{}{}
		}
	}
}

// Size returns the number of registered events.
func (reg *InMemoryEventTemplateRegistry) Size() int {
	reg.mutex.RLock()
	defer reg.mutex.RUnlock()

	return len(reg.templates)
}

// GetEventIDs returns all registered event IDs.
func (reg *InMemoryEventTemplateRegistry) GetEventIDs() []string {
	reg.mutex.RLock()
	defer reg.mutex.RUnlock()

	eventIDs := make([]string, 0, len(reg.templates))
	for eventID := range reg.templates {
		eventIDs = append(eventIDs, eventID)
	}
	return eventIDs
}

// GetStates returns all states for a given event ID.
func (reg *InMemoryEventTemplateRegistry) GetStates(eventID string) []string {
	reg.mutex.RLock()
	defer reg.mutex.RUnlock()

	var states []string
	if stateMap, exists := reg.templates[eventID]; exists {
		states = make([]string, 0, len(stateMap))
		for state := range stateMap {
			states = append(states, state)
		}
	}
	return states
}
