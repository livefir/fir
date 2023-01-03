package fir

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/golang/glog"
	"github.com/gorilla/websocket"
	"github.com/livefir/fir/pubsub"
)

func onWebsocket(w http.ResponseWriter, r *http.Request, route *route, sessionUserStore userStore) {
	channel := route.channelFunc(r, route.id)
	if channel == nil {
		glog.Errorf("[onWebsocket] error: channel is empty")
		http.Error(w, "channel is empty", http.StatusUnauthorized)
		return
	}

	conn, err := route.websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ctx := context.Background()

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
		for patchset := range subscription.C() {
			go writeDOMevents(conn, *channel, route.template, patchset)
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
			for patchset := range reloadSubscriber.C() {
				go writeDOMevents(conn, devReloadChannel, route.template, patchset)
			}
		}()
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
			glog.Errorf("[onWebsocket] err: parsing event, msg %s \n", string(message))
			continue
		}

		if event.ID == "" {
			glog.Errorf("[onWebsocket] err: event %v, field event.id is required\n", event)
			continue
		}

		eventCtx := RouteContext{
			event:     event,
			request:   r,
			response:  w,
			route:     route,
			userStore: sessionUserStore,
		}

		glog.Errorf("[onWebsocket] received event: %+v\n", event)
		onEventFunc, ok := route.onEvents[strings.ToLower(event.ID)]
		if !ok {
			glog.Errorf("[onWebsocket] err: event %v, event.id not found\n", event)
			continue
		}

		sessionUserStore = handleOnEventResult(onEventFunc(eventCtx), eventCtx, publishEvents(ctx, eventCtx))
	}
}

func writeDOMevents(conn *websocket.Conn, channel string, t *template.Template, pubsubEvents []pubsub.Event) error {
	message := domEvents(t, pubsubEvents)
	if len(message) == 0 {
		err := fmt.Errorf("[writePatchOperations] error: message is empty, channel %s, events %+v", channel, pubsubEvents)
		log.Println(err)
		return err
	}
	glog.Errorf("[writePatchOperations] sending patch op to client:%v,  %+v\n", conn.RemoteAddr().String(), string(message))
	err := conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		glog.Errorf("[writePatchOperations] error: writing message for channel:%v, closing conn with err %v", channel, err)
		conn.Close()
	}
	return err
}
