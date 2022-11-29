package main

import (
	"net/http"
	"sync/atomic"

	"github.com/adnaan/fir"
)

func app() fir.RouteOptions {
	var value int32

	load := func(ctx fir.Context) error {
		return ctx.KV("count", atomic.LoadInt32(&value))
	}

	inc := func(ctx fir.Context) error {
		return ctx.Morph(
			"#count",
			fir.Block("count", fir.M{"count": atomic.AddInt32(&value, 1)}),
		)
	}

	dec := func(ctx fir.Context) error {
		return ctx.Morph(
			"#count",
			fir.Block("count", fir.M{"count": atomic.AddInt32(&value, -1)}),
		)
	}

	return fir.RouteOptions{
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
