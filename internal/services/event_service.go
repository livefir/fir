package services

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DefaultEventService implements EventService
type DefaultEventService struct {
	registry  EventRegistry
	validator EventValidator
	publisher EventPublisher
	logger    EventLogger
	metrics   *eventMetrics
}

// eventMetrics holds thread-safe metrics
type eventMetrics struct {
	totalEvents      int64
	successfulEvents int64
	failedEvents     int64
	totalLatency     time.Duration
	mu               sync.RWMutex
}

// NewDefaultEventService creates a new DefaultEventService
func NewDefaultEventService(
	registry EventRegistry,
	validator EventValidator,
	publisher EventPublisher,
	logger EventLogger,
) *DefaultEventService {
	return &DefaultEventService{
		registry:  registry,
		validator: validator,
		publisher: publisher,
		logger:    logger,
		metrics:   &eventMetrics{},
	}
}

// ProcessEvent processes an event request through the complete pipeline
func (s *DefaultEventService) ProcessEvent(ctx context.Context, req EventRequest) (*EventResponse, error) {
	start := time.Now()
	s.metrics.incrementTotal()

	// Log event start
	if s.logger != nil {
		s.logger.LogEventStart(ctx, req)
	}

	// Validate the event request
	if s.validator != nil {
		if err := s.validator.ValidateEvent(req); err != nil {
			s.metrics.incrementFailed()
			if s.logger != nil {
				s.logger.LogEventError(ctx, req, err)
			}
			return nil, NewEventError(req.ID, "VALIDATION_ERROR", "Event validation failed").WithCause(err)
		}
	}

	// Get the event handler
	handler, exists := s.registry.GetHandler(req.ID)
	if !exists {
		err := fmt.Errorf("no handler registered for event: %s", req.ID)
		s.metrics.incrementFailed()
		if s.logger != nil {
			s.logger.LogEventError(ctx, req, err)
		}
		return nil, NewEventError(req.ID, "HANDLER_NOT_FOUND", "Event handler not found").WithCause(err)
	}

	// Process the event
	resp, err := handler.Handle(ctx, req)
	if err != nil {
		s.metrics.incrementFailed()
		if s.logger != nil {
			s.logger.LogEventError(ctx, req, err)
		}
		return nil, NewEventError(req.ID, "PROCESSING_ERROR", "Event processing failed").WithCause(err)
	}

	// Publish any pubsub events
	if s.publisher != nil && len(resp.PubSubEvents) > 0 {
		if err := s.publisher.PublishEvents(resp.PubSubEvents); err != nil {
			// Log the error but don't fail the entire request
			// TODO: Use proper logger interface once available
			fmt.Printf("Failed to publish events: %v for event_id: %s\n", err, req.ID)
		}
	}

	// Record metrics
	s.metrics.incrementSuccessful()
	s.metrics.addLatency(time.Since(start))

	// Log success
	if s.logger != nil {
		s.logger.LogEventSuccess(ctx, req, resp)
	}

	return resp, nil
}

// RegisterHandler registers an event handler
func (s *DefaultEventService) RegisterHandler(eventID string, handler EventHandler) error {
	return s.registry.RegisterHandler(eventID, handler)
}

// GetEventMetrics returns current event processing metrics
func (s *DefaultEventService) GetEventMetrics() EventMetrics {
	return s.metrics.getMetrics()
}

// eventMetrics methods
func (m *eventMetrics) incrementTotal() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.totalEvents++
}

func (m *eventMetrics) incrementSuccessful() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.successfulEvents++
}

func (m *eventMetrics) incrementFailed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failedEvents++
}

func (m *eventMetrics) addLatency(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.totalLatency += duration
}

func (m *eventMetrics) getMetrics() EventMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var avgLatency float64
	if m.totalEvents > 0 {
		avgLatency = float64(m.totalLatency.Nanoseconds()) / float64(m.totalEvents) / 1000000 // Convert to milliseconds
	}

	return EventMetrics{
		TotalEvents:      m.totalEvents,
		SuccessfulEvents: m.successfulEvents,
		FailedEvents:     m.failedEvents,
		AverageLatency:   avgLatency,
	}
}

// InMemoryEventRegistry implements EventRegistry using in-memory storage
type InMemoryEventRegistry struct {
	handlers map[string]EventHandler
	mu       sync.RWMutex
}

// NewInMemoryEventRegistry creates a new InMemoryEventRegistry
func NewInMemoryEventRegistry() *InMemoryEventRegistry {
	return &InMemoryEventRegistry{
		handlers: make(map[string]EventHandler),
	}
}

// GetHandler retrieves a handler for the given event ID
func (r *InMemoryEventRegistry) GetHandler(eventID string) (EventHandler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	handler, exists := r.handlers[eventID]
	return handler, exists
}

// RegisterHandler registers a handler for the given event ID
func (r *InMemoryEventRegistry) RegisterHandler(eventID string, handler EventHandler) error {
	if eventID == "" {
		return fmt.Errorf("event ID cannot be empty")
	}
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[eventID] = handler
	return nil
}

// ListHandlers returns a list of registered event IDs
func (r *InMemoryEventRegistry) ListHandlers() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	handlers := make([]string, 0, len(r.handlers))
	for eventID := range r.handlers {
		handlers = append(handlers, eventID)
	}
	return handlers
}
