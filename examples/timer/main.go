package main

import (
	"log"
	"net/http"
	"time"

	"github.com/adnaan/fir"
)

func NewTimer() *Timer {
	publisher := make(chan fir.Patchset)
	ticker := time.NewTicker(time.Second)
	t := &Timer{publisher: publisher}
	go func() {
		for ; true; <-ticker.C {
			publisher <- fir.Patchset{fir.Store{
				Name: "fir",
				Data: map[string]any{"ts": time.Now().String()},
			}}
		}
	}()
	return t
}

type Timer struct {
	fir.DefaultView
	publisher chan fir.Patchset
}

func (t *Timer) Content() string {
	return "app.html"
}

func (t *Timer) OnGet(_ http.ResponseWriter, _ *http.Request) fir.Page {
	return fir.Page{
		Data: map[string]any{"ts": time.Now().String()},
	}
}

func (t *Timer) Publisher() <-chan fir.Patchset {
	return t.publisher
}

func main() {
	c := fir.NewController("fir-timer", fir.DevelopmentMode(true))
	http.Handle("/", c.Handler(NewTimer()))
	log.Println("listening on http://localhost:9867")
	http.ListenAndServe(":9867", nil)
}
