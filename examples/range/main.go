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

func index() []fir.RouteOption {
	return []fir.RouteOption{
		fir.Content("app.html"),
		fir.OnLoad(func(e fir.Event, r fir.RouteRenderer) error {
			return r(fir.M{"total": 0})
		}),
		fir.OnEvent("update", func(e fir.Event, r fir.PatchRenderer) error {
			req := new(countRequest)
			if err := e.DecodeParams(req); err != nil {
				return err
			}
			count, err := strconv.Atoi(req.Count)
			if err != nil {
				return err
			}
			return r(fir.Store("fir", fir.M{"total": count * 10}))
		}),
	}
}

func main() {
	c := fir.NewController("fir-range", fir.DevelopmentMode(true))
	http.Handle("/", c.RouteFunc(index))
	log.Println("listening on http://localhost:9867")
	http.ListenAndServe(":9867", nil)
}
