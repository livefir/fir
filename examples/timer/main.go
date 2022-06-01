package main

import (
	"log"
	"net/http"
	"time"

	pwc "github.com/adnaan/fir/controller"
)

func NewTimer() *Timer {
	timerCh := make(chan pwc.Event)
	ticker := time.NewTicker(time.Second)
	go func() {
		for ; true; <-ticker.C {
			timerCh <- pwc.Event{ID: "tick"}
		}
	}()
	return &Timer{ch: timerCh}
}

type Timer struct {
	pwc.DefaultView
	ch chan pwc.Event
}

func (t *Timer) Content() string {
	return "app.html"
}

func (t *Timer) OnRequest(_ http.ResponseWriter, _ *http.Request) (pwc.Status, pwc.Data) {
	return pwc.Status{Code: 200}, pwc.Data{
		"ts": time.Now().String(),
	}
}

func (t *Timer) OnEvent(s pwc.Socket) error {
	switch s.Event().ID {
	case "tick":
		s.Store("").UpdateProp("ts", time.Now().String())
		return nil
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", s.Event())
	}
	return nil
}

func (t *Timer) EventReceiver() <-chan pwc.Event {
	return t.ch
}

func main() {
	glvc := pwc.Websocket("fir-timer", pwc.DevelopmentMode(true))
	http.Handle("/", glvc.Handler(NewTimer()))
	log.Println("listening on http://localhost:9867")
	http.ListenAndServe(":9867", nil)
}
