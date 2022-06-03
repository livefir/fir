package main

import (
	"log"
	"net/http"
	"time"

	"github.com/adnaan/fir"
)

type Loading struct {
	fir.DefaultView
}

func (l *Loading) Content() string {
	return "app.html"
}

func (l *Loading) OnEvent(s fir.Socket) error {
	switch s.Event().ID {
	case "loading":
		// "" defaults to "fir" store
		s.Store("loader").Update(true)
		defer func() {
			s.Store("loader").Update(false)
		}()

		// some work
		time.Sleep(time.Second * 2)
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", s.Event())
	}
	return nil
}

func main() {
	c := fir.NewController("fir-counter", fir.DevelopmentMode(true))
	http.Handle("/", c.Handler(&Loading{}))
	log.Println("listening on http://localhost:9867")
	http.ListenAndServe(":9867", nil)
}
