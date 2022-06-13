package main

import (
	"log"
	"net/http"

	"github.com/adnaan/fir"
)

func main() {
	c := fir.NewController("default-fir-app", fir.DevelopmentMode(true))
	http.Handle("/", c.Handler(&fir.DefaultView{}))
	log.Println("listening on http://localhost:9867")
	http.ListenAndServe(":9867", nil)
}
