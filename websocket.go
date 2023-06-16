package fir

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/livefir/fir/internal/dom"
	"github.com/livefir/fir/pubsub"
	"k8s.io/klog/v2"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 55 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 1024
)

func onWebsocket(w http.ResponseWriter, r *http.Request, cntrl *controller) {
	conn, err := cntrl.websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		klog.Errorf("[onWebsocket] upgrade err: %v\n", err)
		return
	}

	conn.SetReadLimit(maxMessageSize)
	conn.EnableWriteCompression(true)
	conn.SetCompressionLevel(5)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		klog.Errorf("[onWebsocket] pong from %v\n", conn.RemoteAddr())
		return conn.SetReadDeadline(time.Now().Add(pongWait))
	})
	defer conn.Close()
	ctx := context.Background()
	done := make(chan struct{})
	send := make(chan []byte)
	go writePump(conn, send)
	wg := &sync.WaitGroup{}
	wg.Add(len(cntrl.routes))

	for _, rt := range cntrl.routes {
		go func(route *route) {
			defer wg.Done()
			routeChannel := route.channelFunc(r, route.id)
			if routeChannel == nil {
				klog.Errorf("[onWebsocket] error: channel is empty")
				http.Error(w, "channel is empty", http.StatusUnauthorized)
				return
			}

			// subscribers
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
					go renderAndWriteEvent(send, *routeChannel, routeCtx, pubsubEvent)
				}
			}()

			// eventSender
			go func() {
				for event := range route.eventSender {
					eventCtx := RouteContext{
						event:    event,
						request:  r,
						response: w,
						route:    route,
					}
					klog.Errorf("[onWebsocket] received server event: %+v\n", event)
					onEventFunc, ok := route.onEvents[strings.ToLower(event.ID)]
					if !ok {
						klog.Errorf("[onWebsocket] err: event %v, event.id not found\n", event)
						continue
					}

					// ignore user store for server events
					handleOnEventResult(onEventFunc(eventCtx), eventCtx, publishEvents(ctx, eventCtx))
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
			<-done
		}(rt)
	}

loop:
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("[onWebsocket] error: %v", err)
			}
			break loop
		}

		var event Event
		err = json.NewDecoder(bytes.NewReader(message)).Decode(&event)
		if err != nil {
			klog.Errorf("[onWebsocket] err: %v, \n parsing event, msg %s \n", err, string(message))
			continue
		}

		if event.ID == "" {
			klog.Errorf("[onWebsocket] err: event %v, field event.id is required\n", event)
			continue
		}

		if event.SessionID == nil {
			klog.Errorf("[onWebsocket] err: event %v, field event.	ID is required, closing connection\n", event)
			break loop
		}

		// var routeID string
		// if err = cntrl.secureCookie.Decode(cntrl.cookieName, *event.SessionID, &routeID); err != nil {
		// 	klog.Errorf("[onWebsocket] err: event %v, cookie decode error: %v\n", event, err)
		// 	continue
		// }

		eventRoute := cntrl.routes[*event.SessionID]

		eventCtx := RouteContext{
			event:    event,
			request:  r,
			response: w,
			route:    eventRoute,
		}

		klog.Errorf("[onWebsocket] route %v received event: %+v\n", eventRoute.id, event)
		onEventFunc, ok := eventRoute.onEvents[strings.ToLower(event.ID)]
		if !ok {
			klog.Errorf("[onWebsocket] err: event %v, event.id not found\n", event)
			continue
		}

		go handleOnEventResult(onEventFunc(eventCtx), eventCtx, publishEvents(ctx, eventCtx))
	}
	close(done)
	close(send)
	wg.Wait()
}

func renderAndWriteEvent(send chan []byte, channel string, ctx RouteContext, pubsubEvent pubsub.Event) error {
	events := renderDOMEvents(ctx, pubsubEvent)
	eventsData, err := json.Marshal(events)
	if err != nil {
		klog.Errorf("[writeDOMevents] error: marshaling events %+v, err %v", events, err)
		return err
	}
	if len(eventsData) == 0 {
		err := fmt.Errorf("[writeDOMevents] error: message is empty, channel %s, events %+v", channel, pubsubEvent)
		log.Println(err)
		return err
	}
	send <- eventsData
	return err
}

func writeEvent(send chan []byte, pubsubEvent pubsub.Event) error {
	domEvent := dom.Event{
		Type: pubsubEvent.ID,
	}
	eventsData, err := json.Marshal([]dom.Event{domEvent})
	if err != nil {
		klog.Errorf("[writeReloadEvent] error: marshaling dom event %+v, err %v", domEvent, err)
		return err
	}
	send <- eventsData
	return err
}

func writeConn(conn *websocket.Conn, mt int, payload []byte) error {
	conn.SetWriteDeadline(time.Now().Add(writeWait))
	return conn.WriteMessage(mt, payload)
}

func writePump(conn *websocket.Conn, send chan []byte) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()
	for {
		select {
		case message, ok := <-send:
			if !ok {
				// The hub closed the channel.
				writeConn(conn, websocket.CloseMessage, []byte{})
				return
			}

			conn.SetWriteDeadline(time.Now().Add(writeWait))
			w, err := conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			go func() {
				klog.Errorf("[writeDOMevents] sending patch op to client:%v,  %+v\n", conn.RemoteAddr(), string(message))
			}()
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(send)
			for i := 0; i < n; i++ {
				w.Write(<-send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			klog.Errorf("ping to client: %v\n", conn.RemoteAddr())
			if err := writeConn(conn, websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}
