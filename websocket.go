package fir

import (
	"bytes"
	"context"

	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/goccy/go-json"

	"github.com/gorilla/websocket"
	"github.com/livefir/fir/internal/dom"
	"github.com/livefir/fir/internal/logger"
	"github.com/livefir/fir/pubsub"
	"github.com/minio/sha256-simd"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 20 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 55 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 1024

	EventSocketConnected    = "fir_socket_connected"
	EventSocketDisconnected = "fir_socket_disconnected"
)

type SocketStatus struct {
	Connected bool
	User      string
}

// RedirectUnauthorisedWebSocket sends a 4001 close message to the client
// It sends the redirect url in the close message payload
// If the request is not a websocket request or has error upgrading and writing the close message, it returns false
// redirect url must be less than 123 bytes
func RedirectUnauthorisedWebSocket(w http.ResponseWriter, r *http.Request, redirect string) bool {
	if len(redirect) > 123 {
		panic("redirect url is too long: max size 123 bytes")
	}
	if !websocket.IsWebSocketUpgrade(r) {
		return false
	}

	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Errorf("upgrade err: %v", err)
		return false
	}
	err = conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(4001, redirect), time.Now().Add(writeWait))
	if err != nil {
		logger.Errorf("write control err: %v", err)
		return false
	}
	defer conn.Close()

	return true
}

func onWebsocket(w http.ResponseWriter, r *http.Request, cntrl *controller) {
	wsLogger := logger.Logger().With("client_ip", r.RemoteAddr, "user_agent", r.Header.Get("User-Agent"))
	wsLogger.Debug("websocket connection attempt")

	cookie, err := r.Cookie(cntrl.cookieName)
	if err != nil {
		wsLogger.Error("failed to get session cookie", "error", err, "cookie_name", cntrl.cookieName)
		RedirectUnauthorisedWebSocket(w, r, "/")
		return
	}
	if cookie.Value == "" {
		wsLogger.Error("session cookie is empty")
		RedirectUnauthorisedWebSocket(w, r, "/")
		return
	}
	sessionID, routeID, err := decodeSession(*cntrl.secureCookie, cntrl.cookieName, cookie.Value)
	if err != nil {
		wsLogger.Error("failed to decode session", "error", err)
		RedirectUnauthorisedWebSocket(w, r, "/")
		return
	}

	if sessionID == "" {
		wsLogger.Error("session ID is empty", "route_id", routeID)
		RedirectUnauthorisedWebSocket(w, r, "/")
		return
	}

	if routeID == "" {
		wsLogger.Error("route ID is empty", "session_id", sessionID)
		RedirectUnauthorisedWebSocket(w, r, "/")
		return
	}

	// Add session and route context to logger
	wsLogger = wsLogger.With("session_id", sessionID, "route_id", routeID)
	wsLogger.Debug("session decoded successfully")

	user := getUserFromRequestContext(r)

	connectedUser := user
	if user == "" {
		connectedUser = sessionID
	}
	wsLogger = wsLogger.With("user", connectedUser)

	if cntrl.onSocketConnect != nil {
		wsLogger.Debug("calling onSocketConnect handler")
		err := cntrl.onSocketConnect(connectedUser)
		if err != nil {
			wsLogger.Error("onSocketConnect handler failed", "error", err)
			return
		}
	}
	if cntrl.onSocketDisconnect != nil {
		defer func() {
			wsLogger.Debug("calling onSocketDisconnect handler")
			cntrl.onSocketDisconnect(connectedUser)
		}()
	}

	send := make(chan []byte, 100)

	ctx := context.Background()

	for _, route := range cntrl.routes {
		routeChannel := route.channelFunc(r, route.id)
		if routeChannel == nil {
			logger.Errorf("error: channel is empty")
			http.Error(w, "channel is empty", http.StatusUnauthorized)
			return
		}

		// subscribers: subscribe to pubsub events
		subscription, err := route.pubsub.Subscribe(ctx, *routeChannel)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer subscription.Close()

		go func() {
			for pubsubEvent := range subscription.C() {
				routeCtx := RouteContext{
					request:  r,
					response: w,
					route:    route,
				}

				go renderAndWriteEventWS(send, *routeChannel, routeCtx, pubsubEvent)
			}
		}()

		// eventSenders: handle server events
		go func() {
			for event := range route.eventSender {
				eventCtx := RouteContext{
					event:    event,
					request:  r,
					response: w,
					route:    route,
				}

				withEventLogger := logger.Logger().
					With(
						"route_id", route.id,
						"event_id", event.ID,
						"session_id", sessionID,
					)
				withEventLogger.Info("received server event")
				onEventFunc, ok := route.onEvents[strings.ToLower(event.ID)]
				if !ok {
					withEventLogger.Error("event handler not found",
						"available_events", func() []string {
							keys := make([]string, 0, len(route.onEvents))
							for k := range route.onEvents {
								keys = append(keys, k)
							}
							return keys
						}())
					logger.Errorf("err: event %v, event.id not found", event)
					continue
				}

				// update request context with user
				eventCtx.request = eventCtx.request.WithContext(context.WithValue(context.Background(), UserKey, user))
				channel := *route.channelFunc(eventCtx.request, route.id)
				errorEvent := handleOnEventResult(onEventFunc(eventCtx), eventCtx, publishEvents(ctx, eventCtx, channel))
				if errorEvent != nil {
					renderAndWriteEventWS(send, channel, eventCtx, *errorEvent)
				}
			}
		}()

		if route.developmentMode {
			// subscriber for reload operations in development mode. see watch.go
			reloadSubscriber, err := route.pubsub.Subscribe(ctx, devReloadChannel)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer reloadSubscriber.Close()

			go func() {
				for pubsubEvent := range reloadSubscriber.C() {
					go writeEvent(send, pubsubEvent)
				}
			}()
		}

	}

	conn, err := cntrl.websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Errorf("upgrade err: %v", err)
		return
	}

	conn.SetReadLimit(maxMessageSize)
	// disabled compression since its too noisy: https://github.com/gorilla/websocket/issues/859
	//conn.EnableWriteCompression(true)
	//conn.SetCompressionLevel(4)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		//logger.Infof("pong from %v", conn.RemoteAddr())
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// workaround for noisy logs due to close handler printing a useless error
	//  https://github.com/gorilla/websocket/issues/880
	conn.SetCloseHandler(func(code int, text string) error {
		message := websocket.FormatCloseMessage(code, "")
		conn.WriteControl(websocket.CloseMessage, message, time.Now().Add(writeWait))
		return nil
	})

	for _, route := range cntrl.routes {
		if route.onEvents[EventSocketConnected] == nil {
			continue
		}

		connectedParams := SocketStatus{
			Connected: true,
			User:      connectedUser,
		}
		paramBytes, err := json.Marshal(connectedParams)
		if err != nil {
			logger.Errorf("error: marshaling connectedParams %+v, err %v", connectedParams, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		connectedEvent := Event{
			ID:        EventSocketConnected,
			SessionID: &sessionID,
			Params:    paramBytes,
			Timestamp: time.Now().UTC().UnixMilli(),
		}

		go func(ev Event) {
			for {
				select {
				case route.eventSender <- ev:
					return
				default:
					time.Sleep(10 * time.Millisecond)
				}
			}
		}(connectedEvent)
	}

	writePumpDone := make(chan struct{})
	go writePump(conn, writePumpDone, send)

	sid := ""
	lastEvent := Event{
		SessionID: &sid,
	}

loop:

	for {

		_, message, err := conn.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure) && websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				logger.Errorf("read: %v, %v", conn.RemoteAddr().String(), err)
			}

			break loop
		}
		// start := time.Now()

		var event Event
		err = json.NewDecoder(bytes.NewReader(message)).Decode(&event)
		if err != nil {
			logger.Errorf("err: %v,  parsing event, msg %s ", err, string(message))
			continue
		}

		if event.ID == "" {
			logger.Errorf("err: event %v, field event.id is required", event)
			continue
		}

		logger.Infof("WEBSOCKET DEBUG: received event ID: %s", event.ID)

		// logger.Infof("received event: %+v took %v ", event, time.Since(start))

		if event.ID == "heartbeat" && conn != nil {
			// err := conn.WriteMessage(websocket.TextMessage, []byte(`{"event_id":"heartbeat_ack"}`))
			// if err != nil {
			// 	logger.Errorf("write heartbeat err: %v, ", err)
			// 	break loop
			// }
			go func() {
				send <- []byte(`{"event_id":"heartbeat_ack"}`)
			}()
			// logger.Errorf("wrote heartbeat: %+v took %v ", event, time.Since(start))
			continue
		}

		if event.SessionID == nil {
			logger.Errorf("err: event %v, field session.ID is required, closing connection", event)
			break loop
		}

		if lastEvent.ID == event.ID && *lastEvent.SessionID == *event.SessionID && lastEvent.ElementKey == event.ElementKey {
			lastEventTime := toUnixTime(lastEvent.Timestamp)
			eventTime := toUnixTime(event.Timestamp)
			if lastEventTime.Add(cntrl.dropDuplicateInterval).After(eventTime) {
				if eqBytesHash(lastEvent.Params, event.Params) {
					logger.Errorf("err: dropped duplicate event in last 250ms, event %v ", event)
					continue
				}
			}
		}

		lastEvent = event

		eventSessionID, eventRouteID, err := decodeSession(*cntrl.secureCookie, cntrl.cookieName, *event.SessionID)
		if err != nil {
			logger.Errorf("err: %v,  decoding session, closing connection", err)
			break loop
		}

		if eventSessionID != sessionID || eventRouteID != routeID {
			logger.Errorf("err: event %v, unauthorised session", event)
			break loop
		}

		eventRoute := cntrl.routes[eventRouteID]

		eventCtx := RouteContext{
			event:    event,
			request:  r,
			response: w,
			route:    eventRoute,
		}

		withEventLogger := logger.Logger().
			With(
				"route_id", eventRoute.id,
				"event_id", event.ID,
				"session_id", eventSessionID,
				"element_key", event.ElementKey,
			)
		withEventLogger.Debug("received user event")
		onEventFunc, ok := eventRoute.onEvents[strings.ToLower(event.ID)]
		if !ok {
			withEventLogger.Error("event handler not found")
			logger.Errorf("err: event %v, event.id not found", event)
			continue
		}

		withEventLogger.Debug("executing event handler")
		go func() {
			eventStart := time.Now()
			channel := *eventRoute.channelFunc(eventCtx.request, eventRoute.id)
			withEventLogger.Debug("got channel for event", "channel", channel)

			// Execute the event handler and log the result
			handlerResult := onEventFunc(eventCtx)
			withEventLogger.Debug("event handler completed",
				"duration_ms", time.Since(eventStart).Milliseconds(),
				"has_error", handlerResult != nil,
			)

			if handlerResult != nil {
				withEventLogger.Error("event handler returned error", "error", handlerResult)
			}

			// errors are only sent to current local connection and not published
			errorEvent := handleOnEventResult(handlerResult, eventCtx, publishEvents(ctx, eventCtx, channel))
			if errorEvent != nil {
				withEventLogger.Debug("sending error event to client")
				renderAndWriteEventWS(send, channel, eventCtx, *errorEvent)
			} else {
				withEventLogger.Debug("event processing completed successfully")
			}
		}()

	}

	close(writePumpDone)
	conn.Close()
	for _, route := range cntrl.routes {

		if route.onEvents[EventSocketDisconnected] == nil {
			continue
		}
		connectedParams := SocketStatus{
			Connected: false,
			User:      connectedUser,
		}
		paramBytes, err := json.Marshal(connectedParams)
		if err != nil {
			logger.Errorf("error: marshaling connectedParams %+v, err %v", connectedParams, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		route.eventSender <- Event{
			ID:        EventSocketDisconnected,
			SessionID: &sessionID,
			Params:    paramBytes,
			Timestamp: time.Now().UTC().UnixMilli(),
		}
	}

}

func renderAndWriteEventWS(send chan []byte, channel string, ctx RouteContext, pubsubEvent pubsub.Event) error {
	wsLogger := logger.WithRoute(ctx.route.id).With(
		"event_id", ctx.event.ID,
		"pubsub_event_id", *pubsubEvent.ID,
		"channel", channel,
	)
	wsLogger.Debug("rendering and writing event to websocket")

	renderStart := time.Now()
	events := renderDOMEvents(ctx, pubsubEvent)
	wsLogger.Debug("DOM events rendered",
		"event_count", len(events),
		"render_ms", time.Since(renderStart).Milliseconds(),
	)

	if len(events) == 0 {
		wsLogger.Warn("no DOM events generated for pubsub event")
	} else {
		for i, event := range events {
			var eventType, target, htmlLength string
			if event.Type != nil {
				eventType = *event.Type
			}
			if event.Target != nil {
				target = *event.Target
			}
			if event.Detail != nil && event.Detail.HTML != "" {
				htmlLength = fmt.Sprintf("%d", len(event.Detail.HTML))
			}
			wsLogger.Debug("generated DOM event",
				"index", i,
				"type", eventType,
				"target", target,
				"html_length", htmlLength,
				"has_detail", event.Detail != nil,
			)
		}
	}

	eventsData, err := json.Marshal(events)
	if err != nil {
		wsLogger.Error("error marshaling events for websocket", "error", err, "events", events)
		logger.Errorf("error: marshaling events %+v, err %v", events, err)
		return err
	}
	if len(eventsData) == 0 {
		err := fmt.Errorf("error: message is empty, channel %s, events %+v", channel, pubsubEvent)
		wsLogger.Error("empty message generated", "error", err)
		logger.Errorf("%v", err)
		return err
	}
	wsLogger.Debug("sending events to websocket",
		"payload_size_bytes", len(eventsData),
		"payload_preview", string(eventsData[:min(200, len(eventsData))]),
	)
	send <- eventsData
	wsLogger.Debug("events sent to websocket successfully")
	return err
}

func writeEvent(send chan []byte, pubsubEvent pubsub.Event) error {
	domEvent := dom.Event{
		Type: pubsubEvent.ID,
	}
	eventsData, err := json.Marshal([]dom.Event{domEvent})
	if err != nil {
		logger.Errorf("error: marshaling dom event %+v, err %v", domEvent, err)
		return err
	}
	send <- eventsData
	return err
}

func writeConn(conn *websocket.Conn, mt int, payload []byte) error {
	return conn.WriteMessage(mt, payload)
}

func writePump(conn *websocket.Conn, closeWritePump chan struct{}, send chan []byte) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()
loop:
	for {
		select {
		case message, ok := <-send:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				err := writeConn(conn, websocket.CloseMessage, []byte{})
				if err != nil {
					logger.Errorf("write close err: %v", err)
				}
				return
			}

			w, err := conn.NextWriter(websocket.TextMessage)
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

		case <-closeWritePump:
			break loop
		case <-ticker.C:
			//logger.Infof("ping to client: %v", conn.RemoteAddr())
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := writeConn(conn, websocket.PingMessage, []byte{}); err != nil {
				logger.Errorf("ping err: %v", err)
				return
			}
		}
	}
}

func eqBytesHash(a, b []byte) bool {
	w := sha256.New()
	w.Write(a)
	aHash := w.Sum(nil)
	w.Reset()
	w.Write(b)
	bHash := w.Sum(nil)
	return bytes.Equal(aHash, bHash)
}
