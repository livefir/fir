package handlers

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	firHttp "github.com/livefir/fir/internal/http"
)

// DefaultHandlerChain is the default implementation of HandlerChain
type DefaultHandlerChain struct {
	handlers []RequestHandler
	logger   HandlerLogger
	metrics  HandlerMetrics
	mutex    sync.RWMutex
}

// NewDefaultHandlerChain creates a new default handler chain
func NewDefaultHandlerChain(logger HandlerLogger, metrics HandlerMetrics) *DefaultHandlerChain {
	return &DefaultHandlerChain{
		handlers: make([]RequestHandler, 0),
		logger:   logger,
		metrics:  metrics,
		mutex:    sync.RWMutex{},
	}
}

// Handle processes a request through the handler chain
func (c *DefaultHandlerChain) Handle(ctx context.Context, req *firHttp.RequestModel) (*firHttp.ResponseModel, error) {
	startTime := time.Now()

	c.mutex.RLock()
	handlers := make([]RequestHandler, len(c.handlers))
	copy(handlers, c.handlers)
	c.mutex.RUnlock()

	if len(handlers) == 0 {
		return nil, fmt.Errorf("no handlers configured in chain")
	}

	// Find the first handler that supports this request
	var selectedHandler RequestHandler
	for _, handler := range handlers {
		if handler.SupportsRequest(req) {
			selectedHandler = handler
			break
		}
	}

	if selectedHandler == nil {
		return nil, fmt.Errorf("no handler found for request: %s %s", req.Method, req.URL.Path)
	}

	// Log handler selection
	if c.logger != nil {
		c.logger.LogHandlerSelection(selectedHandler.HandlerName(), req)
		c.logger.LogRequest(selectedHandler.HandlerName(), req)
	}

	// Record metrics
	if c.metrics != nil {
		c.metrics.RecordRequest(selectedHandler.HandlerName(), req.Method)
	}

	// Process the request
	response, err := selectedHandler.Handle(ctx, req)

	duration := time.Since(startTime).Milliseconds()

	// Log response or error
	if c.logger != nil {
		if err != nil {
			c.logger.LogError(selectedHandler.HandlerName(), err, req)
		} else if response != nil {
			c.logger.LogResponse(selectedHandler.HandlerName(), response, duration)
		}
	}

	// Record metrics
	if c.metrics != nil {
		if err != nil {
			c.metrics.RecordError(selectedHandler.HandlerName(), err)
		} else if response != nil {
			c.metrics.RecordResponse(selectedHandler.HandlerName(), response.StatusCode, duration)
		}
	}

	return response, err
}

// AddHandler adds a handler to the chain
func (c *DefaultHandlerChain) AddHandler(handler RequestHandler) {
	if handler == nil {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Check if handler with this name already exists
	for i, existing := range c.handlers {
		if existing.HandlerName() == handler.HandlerName() {
			// Replace existing handler
			c.handlers[i] = handler
			return
		}
	}

	// Add new handler
	c.handlers = append(c.handlers, handler)
}

// RemoveHandler removes a handler from the chain by name
func (c *DefaultHandlerChain) RemoveHandler(handlerName string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for i, handler := range c.handlers {
		if handler.HandlerName() == handlerName {
			// Remove handler by swapping with last element and truncating
			c.handlers[i] = c.handlers[len(c.handlers)-1]
			c.handlers = c.handlers[:len(c.handlers)-1]
			return true
		}
	}

	return false
}

// GetHandlers returns all handlers in the chain
func (c *DefaultHandlerChain) GetHandlers() []RequestHandler {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	handlers := make([]RequestHandler, len(c.handlers))
	copy(handlers, c.handlers)
	return handlers
}

// ClearHandlers removes all handlers from the chain
func (c *DefaultHandlerChain) ClearHandlers() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.handlers = c.handlers[:0] // Clear slice but keep capacity
}

// PriorityHandlerChain is a handler chain that respects handler priority
type PriorityHandlerChain struct {
	*DefaultHandlerChain
	handlerConfigs map[string]HandlerConfig
}

// NewPriorityHandlerChain creates a new priority-based handler chain
func NewPriorityHandlerChain(logger HandlerLogger, metrics HandlerMetrics) *PriorityHandlerChain {
	return &PriorityHandlerChain{
		DefaultHandlerChain: NewDefaultHandlerChain(logger, metrics),
		handlerConfigs:      make(map[string]HandlerConfig),
	}
}

// AddHandlerWithConfig adds a handler with configuration to the chain
func (c *PriorityHandlerChain) AddHandlerWithConfig(handler RequestHandler, config HandlerConfig) {
	if handler == nil {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Store config
	c.handlerConfigs[handler.HandlerName()] = config

	// Check if handler with this name already exists
	for i, existing := range c.handlers {
		if existing.HandlerName() == handler.HandlerName() {
			// Replace existing handler
			c.handlers[i] = handler
			c.sortHandlersByPriority()
			return
		}
	}

	// Add new handler
	c.handlers = append(c.handlers, handler)
	c.sortHandlersByPriority()
}

// RemoveHandler removes a handler and its config
func (c *PriorityHandlerChain) RemoveHandler(handlerName string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Remove from config map
	delete(c.handlerConfigs, handlerName)

	// Remove from handlers slice
	for i, handler := range c.handlers {
		if handler.HandlerName() == handlerName {
			c.handlers[i] = c.handlers[len(c.handlers)-1]
			c.handlers = c.handlers[:len(c.handlers)-1]
			c.sortHandlersByPriority()
			return true
		}
	}

	return false
}

// GetHandlerConfig returns the configuration for a handler
func (c *PriorityHandlerChain) GetHandlerConfig(handlerName string) (HandlerConfig, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	config, exists := c.handlerConfigs[handlerName]
	return config, exists
}

// sortHandlersByPriority sorts handlers by their priority (lower number = higher priority)
func (c *PriorityHandlerChain) sortHandlersByPriority() {
	sort.Slice(c.handlers, func(i, j int) bool {
		configI, existsI := c.handlerConfigs[c.handlers[i].HandlerName()]
		configJ, existsJ := c.handlerConfigs[c.handlers[j].HandlerName()]

		// Default priority is 100 if no config exists
		priorityI := 100
		priorityJ := 100

		if existsI {
			priorityI = configI.Priority
		}
		if existsJ {
			priorityJ = configJ.Priority
		}

		return priorityI < priorityJ
	})
}

// Handle override to respect enabled/disabled handlers
func (c *PriorityHandlerChain) Handle(ctx context.Context, req *firHttp.RequestModel) (*firHttp.ResponseModel, error) {
	startTime := time.Now()

	c.mutex.RLock()
	handlers := make([]RequestHandler, 0, len(c.handlers))

	// Only include enabled handlers
	for _, handler := range c.handlers {
		if config, exists := c.handlerConfigs[handler.HandlerName()]; exists {
			if config.Enabled {
				handlers = append(handlers, handler)
			}
		} else {
			// Default to enabled if no config
			handlers = append(handlers, handler)
		}
	}
	c.mutex.RUnlock()

	if len(handlers) == 0 {
		return nil, fmt.Errorf("no enabled handlers configured in chain")
	}

	// Find the first handler that supports this request
	var selectedHandler RequestHandler
	for _, handler := range handlers {
		if handler.SupportsRequest(req) {
			selectedHandler = handler
			break
		}
	}

	if selectedHandler == nil {
		return nil, fmt.Errorf("no enabled handler found for request: %s %s", req.Method, req.URL.Path)
	}

	// Log handler selection
	if c.logger != nil {
		c.logger.LogHandlerSelection(selectedHandler.HandlerName(), req)
		c.logger.LogRequest(selectedHandler.HandlerName(), req)
	}

	// Record metrics
	if c.metrics != nil {
		c.metrics.RecordRequest(selectedHandler.HandlerName(), req.Method)
	}

	// Process the request
	response, err := selectedHandler.Handle(ctx, req)

	duration := time.Since(startTime).Milliseconds()

	// Log response or error
	if c.logger != nil {
		if err != nil {
			c.logger.LogError(selectedHandler.HandlerName(), err, req)
		} else if response != nil {
			c.logger.LogResponse(selectedHandler.HandlerName(), response, duration)
		}
	}

	// Record metrics
	if c.metrics != nil {
		if err != nil {
			c.metrics.RecordError(selectedHandler.HandlerName(), err)
		} else if response != nil {
			c.metrics.RecordResponse(selectedHandler.HandlerName(), response.StatusCode, duration)
		}
	}

	return response, err
}
