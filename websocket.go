package fir

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
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
			v.cntrl.pubsub.Publish(r.Context(), channel, patch)
		}
	}()

	// subscriber
	subscription, closeSubscription := v.cntrl.pubsub.Subscribe(r.Context(), channel)
	defer closeSubscription()

	go func() {
		for patch := range subscription {
			operation, err := buildOperation(v.viewTemplate, patch)
			if err != nil {
				continue
			}

			err = conn.WriteJSON([]Operation{operation})
			if err != nil {
				log.Printf("error: writing message for channel:%v, closing conn with err %v", channel, err)
				conn.Close()
			}
		}
	}()

loop:
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("c.readMessage error: ", err)
			break loop
		}

		event := new(Event)
		err = json.NewDecoder(bytes.NewReader(message)).Decode(event)
		if err != nil {
			log.Printf("err: parsing event, msg %s \n", string(message))
			continue
		}

		if event.ID == "" {
			log.Printf("err: event %v, field event.id is required\n", event)
			continue
		}

		v.reloadTemplates()
	}
}
