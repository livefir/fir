package main

import (
	"net/http"

	"github.com/adnaan/fir"
	"github.com/adnaan/fir/examples/twitter/routes/app"
)

func main() {

	controller := fir.NewController("app", fir.DevelopmentMode(true))
	http.Handle("/", controller.RouteFunc(app.Route))
	http.ListenAndServe(":9867", nil)
}
