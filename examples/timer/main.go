package main

import (
	"log"
	"net/http"
	"time"

	"github.com/adnaan/fir"
)

func NewTimer() *Timer {
	timerCh := make(chan fir.Event)
	ticker := time.NewTicker(time.Second)
	go func() {
		for ; true; <-ticker.C {
			timerCh <- fir.Event{ID: "tick"}
		}
	}()
	return &Timer{ch: timerCh}
}

type Timer struct {
	fir.DefaultView
	ch chan fir.Event
}

func (t *Timer) Content() string {
	return "app.html"
}

func (t *Timer) OnRequest(_ http.ResponseWriter, _ *http.Request) (fir.Status, fir.Data) {
	return fir.Status{Code: 200}, fir.Data{
		"ts": time.Now().String(),
	}
}

func (t *Timer) OnEvent(s fir.Socket) error {
	switch s.Event().ID {
	case "tick":
		s.Store("").UpdateProp("ts", time.Now().String())
		return nil
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", s.Event())
	}
	return nil
}

func (t *Timer) EventReceiver() <-chan fir.Event {
	return t.ch
}

func main() {
	glvc := fir.Websocket("fir-timer", fir.DevelopmentMode(true))
	http.Handle("/", glvc.Handler(NewTimer()))
	log.Println("listening on http://localhost:9867")
	http.ListenAndServe(":9867", nil)
}
