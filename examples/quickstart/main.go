package main

import (
	"net/http"
	"sync/atomic"

	"github.com/adnaan/fir"
)

func app() []fir.RouteOption {
	var value int32

	load := func(e fir.Event, r fir.RouteRenderer) error {
		return r(fir.M{"count": atomic.LoadInt32(&value)})
	}

	inc := func(e fir.Event, r fir.PatchRenderer) error {
		return r(
			fir.Morph(
				"#count",
				fir.Block("count", fir.M{"count": atomic.AddInt32(&value, 1)}),
			))
	}

	dec := func(e fir.Event, r fir.PatchRenderer) error {
		return r(
			fir.Morph(
				"#count",
				fir.Block("count", fir.M{"count": atomic.AddInt32(&value, -1)}),
			))
	}

	return []fir.RouteOption{
		fir.ID("app"),
		fir.Content("app.html"),
		fir.OnLoad(load),
		fir.OnEvent("inc", inc),
		fir.OnEvent("dec", dec),
	}
}

func main() {
	controller := fir.NewController("app", fir.DevelopmentMode(true))
	http.Handle("/", controller.RouteFunc(app))
	http.ListenAndServe(":9867", nil)
}