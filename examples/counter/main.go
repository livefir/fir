package main

import (
	"log"
	"net/http"
	"sync/atomic"

	pwc "github.com/adnaan/fir/controller"
)

type Counter struct {
	pwc.DefaultView
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

func (c *Counter) OnMount(_ http.ResponseWriter, _ *http.Request) (pwc.Status, pwc.M) {
	return pwc.Status{Code: 200}, pwc.M{
		"count": c.Value(),
	}
}

func (c *Counter) OnLiveEvent(ctx pwc.Context) error {
	switch ctx.Event().ID {
	case "inc":
		ctx.Store().UpdateProp("count", c.Inc())
	case "dec":
		ctx.Store().UpdateProp("count", c.Dec())
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", ctx.Event())
	}
	return nil
}

func main() {
	glvc := pwc.Websocket("fir-counter", pwc.DevelopmentMode(true))
	http.Handle("/", glvc.Handler(&Counter{}))
	log.Println("listening on http://localhost:9867")
	http.ListenAndServe(":9867", nil)
}
