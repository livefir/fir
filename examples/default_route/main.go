package main

import (
	"net/http"

	"github.com/adnaan/fir"
)

func main() {
	c := fir.NewController("default", fir.DevelopmentMode(true))
	http.Handle("/", c.RouteFunc(func() fir.RouteOptions { return nil }))
	http.ListenAndServe(":9867", nil)
}
