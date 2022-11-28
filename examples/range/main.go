package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/adnaan/fir"
)

type countRequest struct {
	Count string `json:"count"`
}

func index() fir.RouteOptions {
	return fir.RouteOptions{
		fir.Content("app.html"),
		fir.OnLoad(func(ctx fir.Context) error {
			return &fir.M{"total": 0}
		}),
		fir.OnEvent("update", func(ctx fir.Context) error {
			req := new(countRequest)
			if err := ctx.DecodeParams(req); err != nil {
				return err
			}
			count, err := strconv.Atoi(req.Count)
			if err != nil {
				return err
			}
			return ctx.Patch(fir.Store("fir", fir.M{"total": count * 10}))
		}),
	}
}

func main() {
	c := fir.NewController("fir-range", fir.DevelopmentMode(true))
	http.Handle("/", c.RouteFunc(index))
	log.Println("listening on http://localhost:9867")
	http.ListenAndServe(":9867", nil)
}
