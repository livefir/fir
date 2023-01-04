package main

import (
	"net/http"

	"github.com/livefir/fir"
)

func index() fir.RouteOptions {
	return fir.RouteOptions{
		fir.ID("formbuilder"),
		fir.Content("app.html"),
		fir.OnLoad(func(ctx fir.RouteContext) error {
			return nil
		}),
		fir.OnEvent("add", func(ctx fir.RouteContext) error {
			return nil
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
