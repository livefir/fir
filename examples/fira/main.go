package main

import (
	"net/http"
	"sync/atomic"

	"github.com/adnaan/fir"
)

func index() fir.RouteOptions {
	var value int32

	load := func(ctx fir.Context) error {
		return ctx.KV("count", atomic.LoadInt32(&value))
	}

	inc := func(ctx fir.Context) error {
		return ctx.MorphKV("count", atomic.AddInt32(&value, 1))
	}

	dec := func(ctx fir.Context) error {
		return ctx.MorphKV("count", atomic.AddInt32(&value, -1))
	}

	return fir.RouteOptions{
		fir.ID("counter"),
		fir.Content("app.html"),
		fir.OnLoad(load),
		fir.OnEvent("inc", inc),
		fir.OnEvent("dec", dec),
	}
}

func main() {
	controller := fir.NewController("app", fir.DevelopmentMode(true))
	http.Handle("/", controller.RouteFunc(index))
	http.ListenAndServe(":9867", nil)
}