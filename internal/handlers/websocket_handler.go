package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	firHttp "github.com/livefir/fir/internal/http"
	"github.com/livefir/fir/internal/services"
)

// WebSocketHandler handles WebSocket upgrade requests and manages WebSocket connections
type WebSocketHandler struct {
	eventService    services.EventService
	responseBuilder services.ResponseBuilder
	config          HandlerConfig
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(
	eventService services.EventService,
	responseBuilder services.ResponseBuilder,
) *WebSocketHandler {
	return &WebSocketHandler{
		eventService:    eventService,
		responseBuilder: responseBuilder,
		config: HandlerConfig{
			Name:        "websocket-handler",
			Priority:    5, // High priority to handle WebSocket upgrades early
			Methods:     []string{http.MethodGet},
			ContentType: "",
			Enabled:     true,
		},
	}
}

// Handle processes a WebSocket upgrade request
func (h *WebSocketHandler) Handle(ctx context.Context, req *firHttp.RequestModel) (*firHttp.ResponseModel, error) {
	// Validate WebSocket upgrade headers
	if err := h.validateWebSocketHeaders(req); err != nil {
		return h.responseBuilder.BuildErrorResponse(
			fmt.Errorf("invalid websocket upgrade request: %w", err),
			http.StatusBadRequest,
		)
	}

	// For WebSocket upgrades, we need to return a special response
	// that indicates the transport layer should handle the upgrade
	resp := &firHttp.ResponseModel{
		StatusCode: http.StatusSwitchingProtocols,
		Headers: map[string]string{
			"Upgrade":    "websocket",
			"Connection": "Upgrade",
		},
		Body: []byte{},
	}

	// Add WebSocket-specific headers
	h.addWebSocketHeaders(resp, req)

	return resp, nil
}

// SupportsRequest determines if this handler can process the given request
func (h *WebSocketHandler) SupportsRequest(req *firHttp.RequestModel) bool {
	// Must be a GET request
	if req.Method != http.MethodGet {
		return false
	}

	// Must have WebSocket upgrade headers
	return h.isWebSocketUpgrade(req)
}

// HandlerName returns the name of this handler
func (h *WebSocketHandler) HandlerName() string {
	return h.config.Name
}

// validateWebSocketHeaders validates the required WebSocket upgrade headers
func (h *WebSocketHandler) validateWebSocketHeaders(req *firHttp.RequestModel) error {
	// Check Upgrade header
	upgrade := req.Header.Get("Upgrade")
	if strings.ToLower(upgrade) != "websocket" {
		return fmt.Errorf("invalid upgrade header: %s", upgrade)
	}

	// Check Connection header
	connection := req.Header.Get("Connection")
	if !strings.Contains(strings.ToLower(connection), "upgrade") {
		return fmt.Errorf("invalid connection header: %s", connection)
	}

	// Check WebSocket version
	version := req.Header.Get("Sec-WebSocket-Version")
	if version != "13" {
		return fmt.Errorf("unsupported websocket version: %s", version)
	}

	// Check WebSocket key
	key := req.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		return fmt.Errorf("missing websocket key")
	}

	return nil
}

// addWebSocketHeaders adds the required WebSocket response headers
func (h *WebSocketHandler) addWebSocketHeaders(resp *firHttp.ResponseModel, req *firHttp.RequestModel) {
	// Generate WebSocket accept key
	key := req.Header.Get("Sec-WebSocket-Key")
	acceptKey := h.generateWebSocketAcceptKey(key)
	resp.Headers["Sec-WebSocket-Accept"] = acceptKey

	// Handle WebSocket protocol if requested
	protocol := req.Header.Get("Sec-WebSocket-Protocol")
	if protocol != "" {
		// For now, accept the first protocol offered
		// This could be enhanced with protocol negotiation
		protocols := strings.Split(protocol, ",")
		if len(protocols) > 0 {
			resp.Headers["Sec-WebSocket-Protocol"] = strings.TrimSpace(protocols[0])
		}
	}

	// Add any custom extensions if needed
	// extensions := req.Header.Get("Sec-WebSocket-Extensions")
	// Handle extensions negotiation here if needed
}

// generateWebSocketAcceptKey generates the WebSocket accept key from the client key
func (h *WebSocketHandler) generateWebSocketAcceptKey(clientKey string) string {
	// In a real implementation, this would use SHA-1 and Base64 encoding
	// with the WebSocket RFC 6455 magic string "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	// For now, return a placeholder - this would need to be implemented properly
	// with crypto/sha1 and encoding/base64
	return clientKey + "-accept-key"
}

// isWebSocketUpgrade checks if this is a WebSocket upgrade request
func (h *WebSocketHandler) isWebSocketUpgrade(req *firHttp.RequestModel) bool {
	upgrade := req.Header.Get("Upgrade")
	connection := req.Header.Get("Connection")

	return strings.ToLower(upgrade) == "websocket" &&
		strings.Contains(strings.ToLower(connection), "upgrade")
}

// ConfigurableWebSocketHandler is a configurable version of WebSocketHandler
type ConfigurableWebSocketHandler struct {
	*WebSocketHandler
	customConfig HandlerConfig
}

// NewConfigurableWebSocketHandler creates a new configurable WebSocket handler
func NewConfigurableWebSocketHandler(
	eventService services.EventService,
	responseBuilder services.ResponseBuilder,
	config HandlerConfig,
) *ConfigurableWebSocketHandler {
	baseHandler := NewWebSocketHandler(eventService, responseBuilder)

	// Override base config with custom config
	if config.Name != "" {
		baseHandler.config.Name = config.Name
	}
	if config.Priority != 0 {
		baseHandler.config.Priority = config.Priority
	}
	if len(config.Methods) > 0 {
		baseHandler.config.Methods = config.Methods
	}
	if config.PathPattern != "" {
		baseHandler.config.PathPattern = config.PathPattern
	}
	baseHandler.config.Enabled = config.Enabled

	return &ConfigurableWebSocketHandler{
		WebSocketHandler: baseHandler,
		customConfig:     config,
	}
}

// SupportsRequest overrides the base implementation to use custom configuration
func (h *ConfigurableWebSocketHandler) SupportsRequest(req *firHttp.RequestModel) bool {
	// First check the base requirements
	if !h.WebSocketHandler.SupportsRequest(req) {
		return false
	}

	// Check custom path pattern if configured
	if h.customConfig.PathPattern != "" {
		// For WebSocket handlers, path pattern matching is important
		// for routing different WebSocket endpoints
		if !strings.HasPrefix(req.URL.Path, h.customConfig.PathPattern) {
			return false
		}
	}

	// Check if handler is enabled
	return h.customConfig.Enabled
}

// HandlerName returns the custom handler name
func (h *ConfigurableWebSocketHandler) HandlerName() string {
	return h.customConfig.Name
}
