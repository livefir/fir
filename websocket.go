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

const UserIDKey = "key_user_id"

func onWebsocket(w http.ResponseWriter, r *http.Request, v *viewHandler) {
	channel := *v.cntrl.channelFunc(r, v.view.ID())

	conn, err := v.cntrl.websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ctx := context.Background()

	// publisher
	go func() {
		for patchset := range v.streamCh {
			err = v.cntrl.pubsub.Publish(ctx, channel, patchset)
			if err != nil {
				log.Printf("[onWebsocket] error publishing patch: %v\n", err)
			}
		}
	}()

	// subscriber
	subscription, err := v.cntrl.pubsub.Subscribe(ctx, channel)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer subscription.Close()

	go func() {
		for patchset := range subscription.C() {
			go writePatchOperations(*conn, channel, v.viewTemplate, patchset)
		}
	}()

	if v.cntrl.opt.developmentMode {
		// subscriber for reload operations in development mode. see watch.go
		reloadSubscriber, err := v.cntrl.pubsub.Subscribe(ctx, devReloadChannel)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer reloadSubscriber.Close()

		go func() {
			for patchset := range reloadSubscriber.C() {
				go writePatchOperations(*conn, devReloadChannel, v.viewTemplate, patchset)
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

		event := new(Event)
		err = json.NewDecoder(bytes.NewReader(message)).Decode(event)
		if err != nil {
			log.Printf("[onWebsocket] err: parsing event, msg %s \n", string(message))
			continue
		}

		if event.ID == "" {
			log.Printf("[onWebsocket] err: event %v, field event.id is required\n", event)
			continue
		}

		v.reloadTemplates()

		log.Printf("[onWebsocket] received event: %+v\n", event)

		patchset := getEventPatchset(*event, v.view)
		err = v.cntrl.pubsub.Publish(ctx, channel, patchset)
		if err != nil {
			log.Printf("[onWebsocket][getEventPatchset] error publishing patch: %v\n", err)
		}
	}
}

func writePatchOperations(conn websocket.Conn, channel string, t *template.Template, patchset Patchset) error {
	message := buildPatchOperations(t, patchset)
	log.Printf("[writePatchOperations] sending patch op to client:%v,  %+v\n", conn.RemoteAddr().String(), string(message))
	err := conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		log.Printf("[writePatchOperations] error: writing message for channel:%v, closing conn with err %v", channel, err)
		conn.Close()
	}
	return err
}
