package main

import (
	"fmt"
	"net/http"

	"github.com/livefir/fir"
	"github.com/livefir/fir/examples/counter-ticker/handler"
	"github.com/livefir/fir/pubsub"
)

func main() {
	pubsubAdapter := pubsub.NewInmem()
	controller := fir.NewController("counter_app",
		fir.DevelopmentMode(true),
		fir.WithPubsubAdapter(pubsubAdapter),
		fir.WithOnSocketConnect(func(userOrSessionID string) error {
			fmt.Printf("socket connected for user %s\n", userOrSessionID)
			return nil
		}),
		fir.WithOnSocketDisconnect(func(userOrSessionID string) {
			fmt.Printf("socket disconnected for user %s\n", userOrSessionID)
		}),
	)
	http.Handle("/", controller.Route(handler.NewRoute(pubsubAdapter)))
	http.ListenAndServe(":9867", nil)
}
