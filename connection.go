package fir

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"github.com/gorilla/websocket"
	"github.com/livefir/fir/internal/dom"
	"github.com/livefir/fir/internal/logger"
	"github.com/livefir/fir/internal/routeservices"
	"github.com/livefir/fir/pubsub"
)

// Type alias for convenience
type RouteInterface = routeservices.RouteInterface

// Connection represents a WebSocket connection with its associated state and behavior
type Connection struct {
	conn          *websocket.Conn
	wsServices    routeservices.WebSocketServices // WebSocketServices interface (replaces controller)
	controller    *controller                     // Legacy controller (will be removed)
	request       *http.Request
	response      http.ResponseWriter
	sessionID     string
	routeID       string
	user          string
	send          chan []byte
	writePumpDone chan struct{}
	lastEvent     Event
	ctx           context.Context
	cancel        context.CancelFunc
	mu            sync.Mutex
	closed        bool
}

// NewConnection creates a new WebSocket connection
func NewConnection(w http.ResponseWriter, r *http.Request, cntrl *controller) (*Connection, error) {
	// Validate session and extract connection info
	cookie, err := r.Cookie(cntrl.cookieName)
	if err != nil {
		logger.Errorf("cookie err: %v", err)
		RedirectUnauthorisedWebSocket(w, r, "/")
		return nil, err
	}
	if cookie.Value == "" {
		logger.Errorf("cookie err: empty")
		RedirectUnauthorisedWebSocket(w, r, "/")
		return nil, fmt.Errorf("empty cookie")
	}

	sessionID, routeID, err := decodeSession(*cntrl.secureCookie, cntrl.cookieName, cookie.Value)
	if err != nil {
		logger.Errorf("decode session err: %v", err)
		RedirectUnauthorisedWebSocket(w, r, "/")
		return nil, err
	}

	if sessionID == "" {
		logger.Errorf("err: sessionID is empty, routeID is: %s", routeID)
		RedirectUnauthorisedWebSocket(w, r, "/")
		return nil, fmt.Errorf("empty sessionID")
	}

	if routeID == "" {
		logger.Errorf("routeID: is empty")
		RedirectUnauthorisedWebSocket(w, r, "/")
		return nil, fmt.Errorf("empty routeID")
	}

	user := getUserFromRequestContext(r)
	connectedUser := user
	if user == "" {
		connectedUser = sessionID
	}

	// Call socket connect handler if exists
	if cntrl.onSocketConnect != nil {
		err := cntrl.onSocketConnect(connectedUser)
		if err != nil {
			return nil, err
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	conn := &Connection{
		controller:    cntrl,
		request:       r,
		response:      w,
		sessionID:     sessionID,
		routeID:       routeID,
		user:          user,
		send:          make(chan []byte, 100),
		writePumpDone: make(chan struct{}),
		lastEvent:     Event{SessionID: &sessionID},
		ctx:           ctx,
		cancel:        cancel,
	}

	return conn, nil
}

// NewConnectionWithServices creates a new WebSocket connection using WebSocketServices
func NewConnectionWithServices(w http.ResponseWriter, r *http.Request, wsServices routeservices.WebSocketServices) (*Connection, error) {
	// Validate session and extract connection info
	cookie, err := r.Cookie(wsServices.GetCookieName())
	if err != nil {
		logger.Errorf("cookie err: %v", err)
		RedirectUnauthorisedWebSocket(w, r, "/")
		return nil, err
	}
	if cookie.Value == "" {
		logger.Errorf("cookie err: empty")
		RedirectUnauthorisedWebSocket(w, r, "/")
		return nil, fmt.Errorf("empty cookie")
	}

	sessionID, routeID, err := wsServices.DecodeSession(cookie.Value)
	if err != nil {
		logger.Errorf("decode session err: %v", err)
		RedirectUnauthorisedWebSocket(w, r, "/")
		return nil, err
	}

	if sessionID == "" {
		logger.Errorf("err: sessionID is empty, routeID is: %s", routeID)
		RedirectUnauthorisedWebSocket(w, r, "/")
		return nil, fmt.Errorf("empty sessionID")
	}

	if routeID == "" {
		logger.Errorf("routeID: is empty")
		RedirectUnauthorisedWebSocket(w, r, "/")
		return nil, fmt.Errorf("empty routeID")
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Connection{
		conn:          nil, // Will be set during Upgrade
		wsServices:    wsServices,
		controller:    nil, // No controller when using WebSocketServices
		request:       r,
		response:      w,
		sessionID:     sessionID,
		routeID:       routeID,
		user:          sessionID,
		send:          make(chan []byte, 256),
		writePumpDone: make(chan struct{}, 1),
		ctx:           ctx,
		cancel:        cancel,
	}, nil
}

// Upgrade upgrades the HTTP connection to WebSocket
func (c *Connection) Upgrade() error {
	var upgrader *websocket.Upgrader

	if c.wsServices != nil {
		upgrader = c.wsServices.GetWebSocketUpgrader()
	} else if c.controller != nil {
		upgrader = &c.controller.websocketUpgrader
	} else {
		return fmt.Errorf("no WebSocketServices or controller available for upgrade")
	}

	conn, err := upgrader.Upgrade(c.response, c.request, nil)
	if err != nil {
		logger.Errorf("upgrade err: %v", err)
		return err
	}

	c.conn = conn
	c.configureConnection()
	return nil
}

// configureConnection sets up WebSocket connection parameters
func (c *Connection) configureConnection() {
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Set close handler to avoid noisy logs
	c.conn.SetCloseHandler(func(code int, text string) error {
		message := websocket.FormatCloseMessage(code, "")
		c.conn.WriteControl(websocket.CloseMessage, message, time.Now().Add(writeWait))
		return nil
	})
}

// StartPubSubListeners starts listening for pubsub events for all routes
func (c *Connection) StartPubSubListeners() error {
	// Use WebSocketServices if available, fallback to controller
	if c.wsServices != nil {
		return c.startPubSubListenersWithServices()
	} else if c.controller != nil {
		return c.startPubSubListenersWithController()
	} else {
		logger.Errorf("no WebSocketServices or controller available")
		return fmt.Errorf("no WebSocketServices or controller available")
	}
}

// startPubSubListenersWithServices uses WebSocketServices interface
func (c *Connection) startPubSubListenersWithServices() error {
	routes := c.wsServices.GetRoutes()

	for _, rt := range routes {
		routeChannel := rt.ChannelFunc()(c.request, rt.ID())
		if routeChannel == nil {
			logger.Errorf("error: channel is empty")
			http.Error(c.response, "channel is empty", http.StatusUnauthorized)
			return fmt.Errorf("channel is empty")
		}

		// Subscribe to pubsub events
		subscription, err := rt.PubSub().Subscribe(c.ctx, *routeChannel)
		if err != nil {
			http.Error(c.response, err.Error(), http.StatusInternalServerError)
			return err
		}

		go func(routeIface RouteInterface, sub pubsub.Subscription) {
			defer sub.Close()
			for pubsubEvent := range sub.C() {
				// Create a route context for WebSocketServices mode
				routeCtx := RouteContext{
					request:        c.request,
					response:       c.response,
					route:          nil,        // No legacy route object in WebSocketServices mode
					routeInterface: routeIface, // Provide RouteInterface for WebSocketServices mode
				}
				go c.renderAndWriteEvent(*routeChannel, routeCtx, pubsubEvent)
			}
		}(rt, subscription)

		// Handle server events
		go func(routeIface RouteInterface) {
			if eventSenderIface := routeIface.EventSender(); eventSenderIface != nil {
				if eventSenderChan, ok := eventSenderIface.(chan Event); ok {
					for event := range eventSenderChan {
						c.handleServerEventWithServices(routeIface, event)
					}
				}
			}
		}(rt)

		// Handle development mode reload events
		if rt.DevelopmentMode() {
			reloadSubscriber, err := rt.PubSub().Subscribe(c.ctx, devReloadChannel)
			if err != nil {
				http.Error(c.response, err.Error(), http.StatusInternalServerError)
				return err
			}

			go func(sub pubsub.Subscription) {
				defer sub.Close()
				for pubsubEvent := range sub.C() {
					go c.writeEvent(pubsubEvent)
				}
			}(reloadSubscriber)
		}
	}

	return nil
}

// startPubSubListenersWithController uses legacy controller approach
func (c *Connection) startPubSubListenersWithController() error {
	for _, rt := range c.controller.routes {
		routeChannel := rt.channelFunc(c.request, rt.id)
		if routeChannel == nil {
			logger.Errorf("error: channel is empty")
			http.Error(c.response, "channel is empty", http.StatusUnauthorized)
			return fmt.Errorf("channel is empty")
		}

		// Subscribe to pubsub events
		subscription, err := rt.pubsub.Subscribe(c.ctx, *routeChannel)
		if err != nil {
			http.Error(c.response, err.Error(), http.StatusInternalServerError)
			return err
		}

		go func(r *route, sub pubsub.Subscription) {
			defer sub.Close()
			for pubsubEvent := range sub.C() {
				routeCtx := RouteContext{
					request:  c.request,
					response: c.response,
					route:    r,
				}
				go c.renderAndWriteEvent(*routeChannel, routeCtx, pubsubEvent)
			}
		}(rt, subscription)

		// Handle server events
		go func(r *route) {
			for event := range r.eventSender {
				c.handleServerEvent(r, event)
			}
		}(rt)

		// Handle development mode reload events
		if rt.developmentMode {
			reloadSubscriber, err := rt.pubsub.Subscribe(c.ctx, devReloadChannel)
			if err != nil {
				http.Error(c.response, err.Error(), http.StatusInternalServerError)
				return err
			}

			go func(sub pubsub.Subscription) {
				defer sub.Close()
				for pubsubEvent := range sub.C() {
					go c.writeEvent(pubsubEvent)
				}
			}(reloadSubscriber)
		}
	}

	return nil
}

// handleServerEvent processes server-sent events
func (c *Connection) handleServerEvent(route *route, event Event) {
	eventCtx := RouteContext{
		event:    event,
		request:  c.request,
		response: c.response,
		route:    route,
	}

	withEventLogger := logger.GetGlobalLogger().WithFields(map[string]any{
		"route_id":   route.id,
		"event_id":   event.ID,
		"session_id": c.sessionID,
		"transport":  "websocket",
	})

	startTime := time.Now()
	withEventLogger.Info("received server event")

	if logger.GetGlobalLogger().IsDebugEnabled() {
		withEventLogger.Debug("processing server event",
			"params", event.Params,
			"timestamp", startTime.Format(time.RFC3339),
		)
	}

	handlerInterface, ok := route.services.EventRegistry.Get(route.id, strings.ToLower(event.ID))
	if !ok {
		logger.Errorf("err: event %v, event.id not found", event)
		return
	}

	onEventFunc, ok := handlerInterface.(OnEventFunc)
	if !ok {
		logger.Errorf("invalid event handler type for event %v", event)
		return
	}

	// Update request context with user
	eventCtx.request = eventCtx.request.WithContext(context.WithValue(context.Background(), UserKey, c.user))
	channel := *route.channelFunc(eventCtx.request, route.id)

	// Time the event handler execution
	handlerStartTime := time.Now()
	result := onEventFunc(eventCtx)
	handlerDuration := time.Since(handlerStartTime)

	if logger.GetGlobalLogger().IsDebugEnabled() {
		withEventLogger.Debug("event handler completed",
			"handler_duration_ms", handlerDuration.Milliseconds(),
		)
	}

	renderStartTime := time.Now()
	errorEvent := handleOnEventResult(result, eventCtx, publishEvents(c.ctx, eventCtx, channel))
	renderDuration := time.Since(renderStartTime)
	totalDuration := time.Since(startTime)

	if logger.GetGlobalLogger().IsDebugEnabled() {
		withEventLogger.Debug("server event processing complete",
			"handler_duration_ms", handlerDuration.Milliseconds(),
			"render_duration_ms", renderDuration.Milliseconds(),
			"total_duration_ms", totalDuration.Milliseconds(),
		)
	}

	if errorEvent != nil {
		c.renderAndWriteEvent(channel, eventCtx, *errorEvent)
	}
}

// handleServerEventWithServices processes server events using WebSocketServices
func (c *Connection) handleServerEventWithServices(routeInterface RouteInterface, event Event) {
	eventCtx := RouteContext{
		event:    event,
		request:  c.request,
		response: c.response,
		route:    nil, // No legacy route in WebSocketServices mode
	}

	withEventLogger := logger.GetGlobalLogger().WithFields(map[string]any{
		"route_id":   routeInterface.ID(),
		"event_id":   event.ID,
		"session_id": c.sessionID,
		"transport":  "websocket",
	})

	startTime := time.Now()
	withEventLogger.Info("received server event")

	if logger.GetGlobalLogger().IsDebugEnabled() {
		withEventLogger.Debug("processing server event",
			"params", event.Params,
			"timestamp", startTime.Format(time.RFC3339),
		)
	}

	handlerInterface, ok := c.wsServices.GetEventRegistry().Get(routeInterface.ID(), strings.ToLower(event.ID))
	if !ok {
		logger.Errorf("err: event %v, event.id not found", event)
		return
	}

	onEventFunc, ok := handlerInterface.(OnEventFunc)
	if !ok {
		logger.Errorf("invalid event handler type for event %v", event)
		return
	}

	// Update request context with user
	eventCtx.request = eventCtx.request.WithContext(context.WithValue(context.Background(), UserKey, c.user))
	channel := *routeInterface.ChannelFunc()(eventCtx.request, routeInterface.ID())

	// Time the event handler execution
	handlerStartTime := time.Now()
	result := onEventFunc(eventCtx)
	handlerDuration := time.Since(handlerStartTime)

	if logger.GetGlobalLogger().IsDebugEnabled() {
		withEventLogger.Debug("event handler completed",
			"handler_duration_ms", handlerDuration.Milliseconds(),
		)
	}

	renderStartTime := time.Now()
	errorEvent := handleOnEventResult(result, eventCtx, publishEvents(c.ctx, eventCtx, channel))
	renderDuration := time.Since(renderStartTime)
	totalDuration := time.Since(startTime)

	if logger.GetGlobalLogger().IsDebugEnabled() {
		withEventLogger.Debug("server event processing complete",
			"handler_duration_ms", handlerDuration.Milliseconds(),
			"render_duration_ms", renderDuration.Milliseconds(),
			"total_duration_ms", totalDuration.Milliseconds(),
		)
	}

	if errorEvent != nil {
		c.renderAndWriteEvent(channel, eventCtx, *errorEvent)
	}
}

// SendConnectedEvent sends socket connected events to all routes
func (c *Connection) SendConnectedEvent() {
	// Skip if no controller available (WebSocketServices mode doesn't need this yet)
	if c.controller == nil {
		return
	}

	connectedUser := c.user
	if c.user == "" {
		connectedUser = c.sessionID
	}

	for _, rt := range c.controller.routes {
		_, hasConnectedHandler := c.controller.eventRegistry.Get(rt.id, EventSocketConnected)
		if !hasConnectedHandler {
			continue
		}

		connectedParams := SocketStatus{
			Connected: true,
			User:      connectedUser,
		}
		paramBytes, err := json.Marshal(connectedParams)
		if err != nil {
			logger.Errorf("error: marshaling connectedParams %+v, err %v", connectedParams, err)
			continue
		}

		connectedEvent := Event{
			ID:        EventSocketConnected,
			SessionID: &c.sessionID,
			Params:    paramBytes,
			Timestamp: time.Now().UTC().UnixMilli(),
		}

		go func(ev Event, r *route) {
			for {
				select {
				case r.eventSender <- ev:
					return
				default:
					time.Sleep(10 * time.Millisecond)
				}
			}
		}(connectedEvent, rt)
	}
}

// SendDisconnectedEvent sends socket disconnected events to all routes
func (c *Connection) SendDisconnectedEvent() {
	// Skip if no controller available (WebSocketServices mode doesn't need this yet)
	if c.controller == nil {
		return
	}

	connectedUser := c.user
	if c.user == "" {
		connectedUser = c.sessionID
	}

	for _, rt := range c.controller.routes {
		_, hasDisconnectedHandler := c.controller.eventRegistry.Get(rt.id, EventSocketDisconnected)
		if !hasDisconnectedHandler {
			continue
		}

		connectedParams := SocketStatus{
			Connected: false,
			User:      connectedUser,
		}
		paramBytes, err := json.Marshal(connectedParams)
		if err != nil {
			logger.Errorf("error: marshaling connectedParams %+v, err %v", connectedParams, err)
			continue
		}

		rt.eventSender <- Event{
			ID:        EventSocketDisconnected,
			SessionID: &c.sessionID,
			Params:    paramBytes,
			Timestamp: time.Now().UTC().UnixMilli(),
		}
	}
}

// StartWritePump starts the write pump for sending messages
func (c *Connection) StartWritePump() {
	go c.writePump()
}

// ReadLoop handles incoming WebSocket messages
func (c *Connection) ReadLoop() {
	defer func() {
		c.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure) && websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				logger.Errorf("read: %v, %v", c.conn.RemoteAddr().String(), err)
			}
			break
		}

		if err := c.handleMessage(message); err != nil {
			logger.Errorf("handle message error: %v", err)
			break
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (c *Connection) handleMessage(message []byte) error {
	var event Event
	err := json.NewDecoder(bytes.NewReader(message)).Decode(&event)
	if err != nil {
		logger.Errorf("err: %v, parsing event, msg %s", err, string(message))
		return nil // Continue processing other messages
	}

	if event.ID == "" {
		logger.Errorf("err: event %v, field event.id is required", event)
		return nil
	}

	// Handle heartbeat
	if event.ID == "heartbeat" {
		go func() {
			c.send <- []byte(`{"event_id":"heartbeat_ack"}`)
		}()
		return nil
	}

	// Validate session
	if event.SessionID == nil {
		logger.Errorf("err: event %v, field session.ID is required, closing connection", event)
		return fmt.Errorf("session ID required")
	}

	// Check for duplicate events
	if c.isDuplicateEvent(event) {
		logger.Errorf("err: dropped duplicate event in last 250ms, event %v", event)
		return nil
	}

	c.lastEvent = event

	// Validate session authorization
	var eventSessionID, eventRouteID string

	if c.wsServices != nil {
		eventSessionID, eventRouteID, err = c.wsServices.DecodeSession(*event.SessionID)
	} else if c.controller != nil {
		eventSessionID, eventRouteID, err = decodeSession(*c.controller.secureCookie, c.controller.cookieName, *event.SessionID)
	} else {
		return fmt.Errorf("no WebSocketServices or controller available for session decoding")
	}

	if err != nil {
		logger.Errorf("err: %v, decoding session, closing connection", err)
		return err
	}

	if eventSessionID != c.sessionID || eventRouteID != c.routeID {
		logger.Errorf("err: event %v, unauthorised session", event)
		return fmt.Errorf("unauthorized session")
	}

	// Process the event
	go c.processEvent(event, eventRouteID)
	return nil
}

// isDuplicateEvent checks if the event is a duplicate of the last event
func (c *Connection) isDuplicateEvent(event Event) bool {
	if c.lastEvent.ID == event.ID &&
		*c.lastEvent.SessionID == *event.SessionID &&
		c.lastEvent.ElementKey == event.ElementKey {

		lastEventTime := toUnixTime(c.lastEvent.Timestamp)
		eventTime := toUnixTime(event.Timestamp)

		var dropInterval time.Duration
		if c.wsServices != nil {
			dropInterval = c.wsServices.GetDropDuplicateInterval()
		} else if c.controller != nil {
			dropInterval = c.controller.dropDuplicateInterval
		} else {
			// Default interval if neither is available
			dropInterval = 100 * time.Millisecond
		}

		if lastEventTime.Add(dropInterval).After(eventTime) {
			return eqBytesHash(c.lastEvent.Params, event.Params)
		}
	}
	return false
}

// processEvent processes a validated event
func (c *Connection) processEvent(event Event, eventRouteID string) {
	if c.wsServices != nil {
		c.processEventWithServices(event, eventRouteID)
	} else if c.controller != nil {
		c.processEventWithController(event, eventRouteID)
	} else {
		logger.Errorf("no WebSocketServices or controller available for event processing")
	}
}

// processEventWithServices processes events using WebSocketServices
func (c *Connection) processEventWithServices(event Event, eventRouteID string) {
	routes := c.wsServices.GetRoutes()
	routeInterface, exists := routes[eventRouteID]
	if !exists {
		logger.Errorf("route %s not found in WebSocketServices", eventRouteID)
		return
	}

	eventRegistry := c.wsServices.GetEventRegistry()

	eventCtx := RouteContext{
		event:          event,
		request:        c.request,
		response:       c.response,
		route:          nil,                          // No legacy route object in WebSocketServices mode
		formDecoder:    routeInterface.FormDecoder(), // Provide form decoder for binding
		routeInterface: routeInterface,               // Provide RouteInterface for WebSocketServices mode
	}

	withEventLogger := logger.GetGlobalLogger().WithFields(map[string]any{
		"route_id":    eventRouteID,
		"event_id":    event.ID,
		"session_id":  c.sessionID,
		"element_key": event.ElementKey,
		"transport":   "websocket",
	})

	startTime := time.Now()
	withEventLogger.Debug("received user event")

	if logger.GetGlobalLogger().IsDebugEnabled() {
		withEventLogger.Debug("processing user event",
			"params", event.Params,
			"timestamp", startTime.Format(time.RFC3339),
		)
	}

	handlerInterface, ok := eventRegistry.Get(eventRouteID, strings.ToLower(event.ID))
	if !ok {
		logger.Errorf("err: event %v, event.id not found", event)
		return
	}

	onEventFunc, ok := handlerInterface.(OnEventFunc)
	if !ok {
		logger.Errorf("invalid event handler type for event %v", event)
		return
	}

	// Update request context with user
	eventCtx.request = eventCtx.request.WithContext(context.WithValue(context.Background(), UserKey, c.user))
	channel := *routeInterface.ChannelFunc()(eventCtx.request, eventRouteID)

	// Time the event handler execution
	handlerStartTime := time.Now()
	result := onEventFunc(eventCtx)
	handlerDuration := time.Since(handlerStartTime)

	if logger.GetGlobalLogger().IsDebugEnabled() {
		withEventLogger.Debug("event handler completed",
			"handler_duration_ms", handlerDuration.Milliseconds(),
		)
	}

	renderStartTime := time.Now()
	errorEvent := handleOnEventResult(result, eventCtx, publishEventsWithServices(c.ctx, routeInterface.PubSub(), channel))
	renderDuration := time.Since(renderStartTime)
	totalDuration := time.Since(startTime)

	if logger.GetGlobalLogger().IsDebugEnabled() {
		withEventLogger.Debug("user event processing complete",
			"handler_duration_ms", handlerDuration.Milliseconds(),
			"render_duration_ms", renderDuration.Milliseconds(),
			"total_duration_ms", totalDuration.Milliseconds(),
		)
	}

	if errorEvent != nil {
		c.renderAndWriteEvent(channel, eventCtx, *errorEvent)
	}
}

// processEventWithController processes events using legacy controller
func (c *Connection) processEventWithController(event Event, eventRouteID string) {

	eventRoute := c.controller.routes[eventRouteID]

	eventCtx := RouteContext{
		event:    event,
		request:  c.request,
		response: c.response,
		route:    eventRoute,
	}

	withEventLogger := logger.GetGlobalLogger().WithFields(map[string]any{
		"route_id":    eventRoute.id,
		"event_id":    event.ID,
		"session_id":  c.sessionID,
		"element_key": event.ElementKey,
		"transport":   "websocket",
	})

	startTime := time.Now()
	withEventLogger.Debug("received user event")

	if logger.GetGlobalLogger().IsDebugEnabled() {
		withEventLogger.Debug("processing user event",
			"params", event.Params,
			"timestamp", startTime.Format(time.RFC3339),
		)
	}

	handlerInterface, ok := eventRoute.services.EventRegistry.Get(eventRoute.id, strings.ToLower(event.ID))
	if !ok {
		logger.Errorf("err: event %v, event.id not found", event)
		return
	}

	onEventFunc, ok := handlerInterface.(OnEventFunc)
	if !ok {
		logger.Errorf("invalid event handler type for event %v", event)
		return
	}

	channel := *eventRoute.channelFunc(eventCtx.request, eventRoute.id)

	// Time the event handler execution
	handlerStartTime := time.Now()
	result := onEventFunc(eventCtx)
	handlerDuration := time.Since(handlerStartTime)

	if logger.GetGlobalLogger().IsDebugEnabled() {
		withEventLogger.Debug("event handler completed",
			"handler_duration_ms", handlerDuration.Milliseconds(),
		)
	}

	renderStartTime := time.Now()
	errorEvent := handleOnEventResult(result, eventCtx, publishEvents(c.ctx, eventCtx, channel))
	renderDuration := time.Since(renderStartTime)
	totalDuration := time.Since(startTime)

	if logger.GetGlobalLogger().IsDebugEnabled() {
		withEventLogger.Debug("user event processing complete",
			"handler_duration_ms", handlerDuration.Milliseconds(),
			"render_duration_ms", renderDuration.Milliseconds(),
			"total_duration_ms", totalDuration.Milliseconds(),
		)
	}

	if errorEvent != nil {
		c.renderAndWriteEvent(channel, eventCtx, *errorEvent)
	}
}

// renderAndWriteEvent renders and writes an event to the WebSocket
func (c *Connection) renderAndWriteEvent(channel string, ctx RouteContext, pubsubEvent pubsub.Event) error {
	var events []dom.Event
	var err error

	// Handle both WebSocketServices mode and legacy mode
	if c.wsServices != nil {
		// WebSocketServices mode: use renderer from routes map
		routes := c.wsServices.GetRoutes()
		routeIface, exists := routes[c.routeID]
		if !exists {
			return fmt.Errorf("route not found for routeID: %s", c.routeID)
		}
		rendererIface := routeIface.GetRenderer()
		if rendererIface == nil {
			return fmt.Errorf("renderer not found for route: %s", c.routeID)
		}
		// Type assert to Renderer
		renderer, ok := rendererIface.(Renderer)
		if !ok {
			return fmt.Errorf("renderer is not of type Renderer for route: %s", c.routeID)
		}
		// Use the renderer with route interface
		if tr, ok := renderer.(*TemplateRenderer); ok {
			events = tr.RenderDOMEventsWithRoute(ctx, pubsubEvent, routeIface)
		} else {
			// Fallback to regular method if not TemplateRenderer
			events = renderer.RenderDOMEvents(ctx, pubsubEvent)
		}
	} else {
		// Legacy mode: use renderer from route
		if ctx.route == nil || ctx.route.renderer == nil {
			return fmt.Errorf("legacy route or renderer is nil")
		}
		events = ctx.route.renderer.RenderDOMEvents(ctx, pubsubEvent)
	}

	eventsData, err := json.Marshal(events)
	if err != nil {
		logger.Errorf("error: marshaling events %+v, err %v", events, err)
		return err
	}
	if len(eventsData) == 0 {
		err := fmt.Errorf("error: message is empty, channel %s, events %+v", channel, pubsubEvent)
		logger.Errorf("%v", err)
		return err
	}
	c.send <- eventsData
	return nil
}

// writeEvent writes a simple event to the WebSocket
func (c *Connection) writeEvent(pubsubEvent pubsub.Event) error {
	domEvent := dom.Event{
		Type: pubsubEvent.ID,
	}
	eventsData, err := json.Marshal([]dom.Event{domEvent})
	if err != nil {
		logger.Errorf("error: marshaling dom event %+v, err %v", domEvent, err)
		return err
	}
	c.send <- eventsData
	return nil
}

// writePump handles writing messages to the WebSocket connection
func (c *Connection) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				err := c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
					logger.Errorf("write close err: %v", err)
				}
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				logger.Errorf("next writer err: %v", err)
				return
			}

			_, err = w.Write(message)
			if err != nil {
				logger.Errorf("write err: %v", err)
				return
			}

			if err := w.Close(); err != nil {
				logger.Errorf("close err: %v", err)
				return
			}

		case <-c.writePumpDone:
			return
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				logger.Errorf("ping err: %v", err)
				return
			}
		}
	}
}

// Close closes the connection and cleans up resources
func (c *Connection) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if already closed
	if c.closed {
		return
	}

	// Mark as closed
	c.closed = true

	// Cancel context to stop all goroutines
	c.cancel()

	// Close write pump channel safely
	select {
	case <-c.writePumpDone:
		// Channel already closed
	default:
		close(c.writePumpDone)
	}

	// Close WebSocket connection
	if c.conn != nil {
		c.conn.Close()
	}

	// Send disconnected event
	c.SendDisconnectedEvent()

	// Call socket disconnect handler if exists
	if c.wsServices != nil {
		connectedUser := c.user
		if c.user == "" {
			connectedUser = c.sessionID
		}
		c.wsServices.OnSocketDisconnect(connectedUser)
	} else if c.controller != nil && c.controller.onSocketDisconnect != nil {
		connectedUser := c.user
		if c.user == "" {
			connectedUser = c.sessionID
		}
		c.controller.onSocketDisconnect(connectedUser)
	}
}
