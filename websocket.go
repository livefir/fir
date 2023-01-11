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

	"github.com/golang/glog"
	"github.com/gorilla/websocket"
	"github.com/livefir/fir/internal/dom"
	"github.com/livefir/fir/pubsub"
)

func onWebsocket(w http.ResponseWriter, r *http.Request, cntrl *controller, sessionUserStore userStore) {
	conn, err := cntrl.websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	wsConn := &websocketConn{conn: conn}
	ctx := context.Background()
	done := make(chan struct{})
	wg := &sync.WaitGroup{}
	wg.Add(len(cntrl.routes))

	for _, rt := range cntrl.routes {
		go func(route *route) {
			defer wg.Done()
			channel := route.channelFunc(r, route.id)
			if channel == nil {
				glog.Errorf("[onWebsocket] error: channel is empty")
				http.Error(w, "channel is empty", http.StatusUnauthorized)
				return
			}

			// eventSender
			go func() {
				for event := range route.eventSender {
					eventCtx := RouteContext{
						event:    event,
						request:  r,
						response: w,
						route:    route,
						// ignore user store for server events because you don't want to affect user state from a non-user event
						userStore: make(map[string]any),
					}
					glog.Errorf("[onWebsocket] received server event: %+v\n", event)
					onEventFunc, ok := route.onEvents[strings.ToLower(event.ID)]
					if !ok {
						glog.Errorf("[onWebsocket] err: event %v, event.id not found\n", event)
						continue
					}

					// ignore user store for server events
					_ = handleOnEventResult(onEventFunc(eventCtx), eventCtx, publishEvents(ctx, eventCtx))
				}
			}()

			// subscribers
			subscription, err := route.pubsub.Subscribe(ctx, *channel)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer subscription.Close()

			go func() {
				for pubsubEvent := range subscription.C() {
					routeCtx := RouteContext{
						request:   r,
						response:  w,
						route:     route,
						userStore: sessionUserStore,
					}
					go renderAndWriteEvent(wsConn, *channel, routeCtx, pubsubEvent)
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
						go writeEvent(wsConn, pubsubEvent)
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
			log.Println("[onWebsocket] c.readMessage error: ", err)
			break loop
		}

		var event Event
		err = json.NewDecoder(bytes.NewReader(message)).Decode(&event)
		if err != nil {
			glog.Errorf("[onWebsocket] err: %v, \n parsing event, msg %s \n", err, string(message))
			continue
		}

		if event.ID == "" {
			glog.Errorf("[onWebsocket] err: event %v, field event.id is required\n", event)
			continue
		}

		eventRoute := cntrl.routes[*event.RouteID]

		eventCtx := RouteContext{
			event:     event,
			request:   r,
			response:  w,
			route:     eventRoute,
			userStore: sessionUserStore,
		}

		glog.Errorf("[onWebsocket] route %v received event: %+v\n", eventRoute.id, event)
		onEventFunc, ok := eventRoute.onEvents[strings.ToLower(event.ID)]
		if !ok {
			glog.Errorf("[onWebsocket] err: event %v, event.id not found\n", event)
			continue
		}

		sessionUserStore = handleOnEventResult(onEventFunc(eventCtx), eventCtx, publishEvents(ctx, eventCtx))
	}
	close(done)
	wg.Wait()
}

type websocketConn struct {
	conn *websocket.Conn
	sync.Mutex
}

func renderAndWriteEvent(ws *websocketConn, channel string, ctx RouteContext, pubsubEvent pubsub.Event) (userStore, error) {
	ws.Lock()
	defer ws.Unlock()
	events := renderDOMEvents(ctx, pubsubEvent)
	events, ctx.userStore = unsetErrors(ctx.userStore, events)
	eventsData, err := json.Marshal(events)
	if err != nil {
		glog.Errorf("[writeDOMevents] error: marshaling events %+v, err %v", events, err)
		return ctx.userStore, err
	}
	if len(eventsData) == 0 {
		err := fmt.Errorf("[writeDOMevents] error: message is empty, channel %s, events %+v", channel, pubsubEvent)
		log.Println(err)
		return ctx.userStore, err
	}
	glog.Errorf("[writeDOMevents] sending patch op to client:%v,  %+v\n", ws.conn.RemoteAddr().String(), string(eventsData))
	err = ws.conn.WriteMessage(websocket.TextMessage, eventsData)
	if err != nil {
		glog.Errorf("[writeDOMevents] error: writing message for channel:%v, closing conn with err %v", channel, err)
		ws.conn.Close()
	}
	return ctx.userStore, err
}

func writeEvent(ws *websocketConn, pubsubEvent pubsub.Event) error {
	ws.Lock()
	defer ws.Unlock()
	reload := dom.Event{
		Type: pubsubEvent.ID,
	}
	reloadData, err := json.Marshal(reload)
	if err != nil {
		glog.Errorf("[writeReloadEvent] error: marshaling reload event %+v, err %v", reload, err)
		return err
	}
	err = ws.conn.WriteMessage(websocket.TextMessage, reloadData)
	if err != nil {
		glog.Errorf("[writeReloadEvent] error: writing message for channel:%v, closing conn with err %v", devReloadChannel, err)
		ws.conn.Close()
	}
	return err
}
