package fir

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

const UserIDKey = "key_user_id"

func onWebsocket(w http.ResponseWriter, r *http.Request, v *viewHandler) {
	channel := *v.cntrl.channelFunc(r, v.view.ID())

	conn, err := v.cntrl.websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// publisher
	go func() {
		for patch := range v.streamCh {
			err := v.cntrl.pubsub.Publish(r.Context(), channel, patch)
			if err != nil {
				log.Printf("[onWebsocket] error publishing patch: %v\n", err)
			}
		}
	}()

	// subscriber
	subscription, err := v.cntrl.pubsub.Subscribe(r.Context(), channel)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer subscription.Close()

	go func() {
		for patch := range subscription.C() {
			go func(patch Patch) {
				log.Printf("[onWebsocket] sending patch to client:%v,  %+v\n", conn.RemoteAddr().String(), patch)
				operation, err := buildOperation(v.viewTemplate, patch)
				if err != nil {
					if strings.ContainsAny("fir-error", err.Error()) {
						return
					}
					log.Printf("[onWebsocket] buildOperation error: %v\n", err)
					return
				}

				err = conn.WriteJSON([]Operation{operation})
				if err != nil {
					log.Printf("[onWebsocket] error: writing message for channel:%v, closing conn with err %v", channel, err)
					conn.Close()
				}
			}(patch)
		}
	}()

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
		for _, patch := range patchset {
			err := v.cntrl.pubsub.Publish(r.Context(), channel, patch)
			if err != nil {
				log.Printf("[onPatchEvent] error publishing patch: %v\n", err)
			}
		}
	}
}
