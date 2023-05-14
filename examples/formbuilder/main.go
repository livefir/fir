package main

import (
	"net/http"
	"time"

	"math/rand"

	"github.com/livefir/fir"
)

func index() fir.RouteOptions {
	rand.Seed(time.Now().UnixNano())
	return fir.RouteOptions{
		fir.ID("formbuilder"),
		fir.Content("app.html"),
		fir.OnLoad(func(ctx fir.RouteContext) error {
			return nil
		}),
		fir.OnEvent("add", func(ctx fir.RouteContext) error {
			return ctx.KV("key", rand.Intn(1000-1)+1)
		}),
		fir.OnEvent("remove", func(ctx fir.RouteContext) error {
			return nil
		}),
	}
}

func main() {
	controller := fir.NewController("formbuilder", fir.DevelopmentMode(true))
	http.Handle("/", controller.RouteFunc(index))
	http.ListenAndServe(":9867", nil)
}
