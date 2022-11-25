package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/adnaan/fir"
)

type CountRequest struct {
	Count string `json:"count"`
}

type index struct{}

func (i *index) load(e fir.Event, r fir.RouteRenderer) error {
	return r(fir.M{"total": 0})
}
func (i *index) update(e fir.Event, r fir.PatchRenderer) error {
	req := new(CountRequest)
	if err := e.DecodeParams(req); err != nil {
		return err
	}
	count, err := strconv.Atoi(req.Count)
	if err != nil {
		return err
	}
	return r(fir.Morph("#total", "total", fir.M{"total": count * 10}))
}
func (i *index) Options() []fir.RouteOption {
	return []fir.RouteOption{
		fir.Content("app.html"),
		fir.OnLoad(i.load),
		fir.OnEvent("update", i.update),
	}
}

func main() {
	c := fir.NewController("fir-range", fir.DevelopmentMode(true))
	http.Handle("/", c.Route(&index{}))
	log.Println("listening on http://localhost:9867")
	http.ListenAndServe(":9867", nil)
}
