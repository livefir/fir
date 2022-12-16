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
		fir.OnLoad(func(ctx fir.RouteContext) error {
			return ctx.Data(0)
		}),
		fir.OnEvent("update", func(ctx fir.RouteContext) error {
			req := new(countRequest)
			if err := ctx.Bind(req); err != nil {
				return err
			}
			count, err := strconv.Atoi(req.Count)
			if err != nil {
				return err
			}
			return ctx.Data(count * 10)
		}),
	}
}

func main() {
	c := fir.NewController("fir-range", fir.DevelopmentMode(true))
	http.Handle("/", c.RouteFunc(index))
	log.Println("listening on http://localhost:9867")
	http.ListenAndServe(":9867", nil)
}
