package main

import (
	"log"
	"net/http"
	"time"

	pwc "github.com/adnaan/fir/controller"
)

type Loading struct {
	pwc.DefaultView
}

func (l *Loading) Content() string {
	return "app.html"
}

func (l *Loading) OnLiveEvent(ctx pwc.Context) error {
	switch ctx.Event().ID {
	case "loading":
		// "" defaults to "fir" store
		ctx.Store("loader").Update(true)
		defer func() {
			ctx.Store("loader").Update(false)
		}()

		// some work
		time.Sleep(time.Second * 2)
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", ctx.Event())
	}
	return nil
}

func main() {
	glvc := pwc.Websocket("fir-counter", pwc.DevelopmentMode(true))
	http.Handle("/", glvc.Handler(&Loading{}))
	log.Println("listening on http://localhost:9867")
	http.ListenAndServe(":9867", nil)
}
