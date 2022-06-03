package main

import (
	"log"
	"net/http"
	"sync/atomic"

	"github.com/adnaan/fir"
)

type Counter struct {
	fir.DefaultView
	count int32
}

func (c *Counter) Inc() int32 {
	atomic.AddInt32(&c.count, 1)
	return atomic.LoadInt32(&c.count)
}
func (c *Counter) Dec() int32 {
	atomic.AddInt32(&c.count, -1)
	return atomic.LoadInt32(&c.count)
}

func (c *Counter) Value() int32 {
	return atomic.LoadInt32(&c.count)
}

func (c *Counter) Content() string {
	return "app.html"
}

func (c *Counter) OnRequest(_ http.ResponseWriter, _ *http.Request) (fir.Status, fir.Data) {
	return fir.Status{Code: 200}, fir.Data{
		"count": c.Value(),
	}
}

func (c *Counter) OnEvent(s fir.Socket) error {
	switch s.Event().ID {
	case "inc":
		s.Store().UpdateProp("count", c.Inc())
	case "dec":
		s.Store().UpdateProp("count", c.Dec())
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", s.Event())
	}
	return nil
}

func main() {
	glvc := fir.Websocket("fir-counter", fir.DevelopmentMode(true))
	http.Handle("/", glvc.Handler(&Counter{}))
	log.Println("listening on http://localhost:9867")
	http.ListenAndServe(":9867", nil)
}
