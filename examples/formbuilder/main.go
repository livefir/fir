package main

import (
	"net/http"

	"github.com/livefir/fir"
	"github.com/livefir/fir/examples/formbuilder/handler"
)

func main() {
	controller := fir.NewController("formbuilder", fir.DevelopmentMode(true))
	// Pass a function literal that calls handler.NewRoute
	http.Handle("/", controller.RouteFunc(func() fir.RouteOptions {
		return handler.NewRoute("app.html")
	}))
	http.ListenAndServe(":9867", nil)
}
