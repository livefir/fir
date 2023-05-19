package fir

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/livefir/fir/internal/dom"
	"github.com/livefir/fir/pubsub"
	"k8s.io/klog/v2"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func onWebsocket(w http.ResponseWriter, r *http.Request, cntrl *controller) {

	wsConn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		CompressionMode: websocket.CompressionNoContextTakeover,
	})
	if err != nil {
		klog.Errorf("[onWebsocket] error: %v, close status", err, websocket.CloseStatus(err))
		return
	}
	wsConn.SetReadLimit(1024) // 1kb
	defer wsConn.Close(websocket.StatusNormalClosure, "")

	ctx, cancel := context.WithTimeout(r.Context(), time.Minute*10)
	defer cancel()

	done := make(chan struct{})
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
					go renderAndWriteEvent(wsConn, *routeChannel, routeCtx, pubsubEvent)
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
						go writeEvent(wsConn, pubsubEvent)
					}
				}()
			}
			<-done
		}(rt)
	}
loop:
	for {
		var event Event
		err = wsjson.Read(ctx, wsConn, &event)
		if err != nil {
			closeStatus := websocket.CloseStatus(err)
			klog.Errorf("[onWebsocket] err: %v", closeStatus)
			if closeStatus == websocket.StatusInvalidFramePayloadData {
				continue
			} else {
				break loop
			}
		}

		if event.ID == "" {
			klog.Errorf("[onWebsocket] err: event %v, field event.id is required\n", event)
			continue
		}

		if event.SessionID == nil {
			klog.Errorf("[onWebsocket] err: event %v, field event.sessionID is required\n", event)
			continue
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
	wg.Wait()

}

func renderAndWriteEvent(ws *websocket.Conn, channel string, ctx RouteContext, pubsubEvent pubsub.Event) error {
	events := renderDOMEvents(ctx, pubsubEvent)
	klog.Errorf("[writeDOMevents] sending patch op %+v\n", events)
	err := wsjson.Write(ctx.request.Context(), ws, events)
	if err != nil {
		klog.Errorf("[writeDOMevents] error: writing message for channel:%v, closing conn with err %v", channel, err)
		ws.Close(websocket.StatusInternalError, "closed")
	}
	return err
}

func writeEvent(ws *websocket.Conn, pubsubEvent pubsub.Event) error {
	reload := dom.Event{
		Type: pubsubEvent.ID,
	}
	err := wsjson.Write(context.Background(), ws, reload)
	if err != nil {
		klog.Errorf("[writeReloadEvent] error: writing message for channel:%v, closing conn with err %v", devReloadChannel, err)
		ws.Close(websocket.StatusInternalError, "closed")
	}
	return err
}
