package fir

import (
	"bytes"
	"context"
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func onWebsocket(w http.ResponseWriter, r *http.Request, route *route) {
	channel := route.channelFunc(r, route.id)
	if channel == nil {
		log.Printf("[onWebsocket] error: channel is empty")
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
			event.request = r
			event.response = w

			onEventFunc, ok := route.onEvents[event.ID]
			if !ok {
				log.Printf("[onWebsocket] err: event %v, event.id not found\n", event)
				continue
			}
			onEventFunc(event, patchSocketRenderer(ctx, conn, *channel, route))

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
			go writePatchOperations(conn, *channel, route.template, patchset)
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
				go writePatchOperations(conn, devReloadChannel, route.template, patchset)
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
			log.Printf("[onWebsocket] err: parsing event, msg %s \n", string(message))
			continue
		}

		if event.ID == "" {
			log.Printf("[onWebsocket] err: event %v, field event.id is required\n", event)
			continue
		}

		event.request = r
		event.response = w

		log.Printf("[onWebsocket] received event: %+v\n", event)
		onEventFunc, ok := route.onEvents[event.ID]
		if !ok {
			log.Printf("[onWebsocket] err: event %v, event.id not found\n", event)
			continue
		}
		err = onEventFunc(event, patchSocketRenderer(ctx, conn, *channel, route))
		if err != nil {
			log.Printf("[onWebsocket] err: event %v, %v\n", event, err)
			continue
		}
	}
}

func writePatchOperations(conn *websocket.Conn, channel string, t *template.Template, patchset []Patch) error {
	message := buildPatchOperations(t, patchset)
	log.Printf("[writePatchOperations] sending patch op to client:%v,  %+v\n", conn.RemoteAddr().String(), string(message))
	err := conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		log.Printf("[writePatchOperations] error: writing message for channel:%v, closing conn with err %v", channel, err)
		conn.Close()
	}
	return err
}
