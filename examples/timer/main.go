package main

import (
	"log"
	"net/http"
	"time"

	"github.com/adnaan/fir"
)

func NewTimer() *Timer {
	stream := make(chan fir.Patch)
	ticker := time.NewTicker(time.Second)
	t := &Timer{stream: stream}
	go func() {
		for ; true; <-ticker.C {
			stream <- fir.Store{
				Name: "fir",
				Data: map[string]any{"ts": time.Now().String()},
			}
		}
	}()
	return t
}

type Timer struct {
	fir.DefaultView
	stream chan fir.Patch
}

func (t *Timer) Content() string {
	return "app.html"
}

func (t *Timer) OnRequest(_ http.ResponseWriter, _ *http.Request) (fir.Status, fir.Data) {
	return fir.Status{Code: 200}, fir.Data{
		"ts": time.Now().String(),
	}
}

func (t *Timer) Stream() <-chan fir.Patch {
	return t.stream
}

func main() {
	c := fir.NewController("fir-timer", fir.DevelopmentMode(true))
	http.Handle("/", c.Handler(NewTimer()))
	log.Println("listening on http://localhost:9867")
	http.ListenAndServe(":9867", nil)
}
