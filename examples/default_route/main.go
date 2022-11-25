package main

import (
	"log"
	"net/http"

	"github.com/adnaan/fir"
)

type IndexRoute struct {
}

func (i *IndexRoute) Options() []fir.RouteOption {
	return []fir.RouteOption{}
}

func main() {
	c := fir.NewController("default-fir-app", fir.DevelopmentMode(true))
	http.Handle("/", c.Route(&IndexRoute{}))
	log.Println("listening on http://localhost:9867")
	http.ListenAndServe(":9867", nil)
}
