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

type Range struct {
	fir.DefaultView
}

func (r *Range) Content() string {
	return "app.html"
}

func (r *Range) OnGet(_ http.ResponseWriter, _ *http.Request) fir.Page {
	return fir.Page{
		Data: map[string]any{
			"total": 0,
		}}
}

func (r *Range) OnEvent(event fir.Event) fir.Patchset {
	switch event.ID {
	case "update":
		req := new(CountRequest)
		if err := event.DecodeParams(req); err != nil {
			return fir.PatchError(err, "failed to decode update request")
		}
		count, err := strconv.Atoi(req.Count)
		if err != nil {
			return fir.PatchError(err, "failed to parse count")
		}
		return fir.Patchset{
			fir.Store{Name: "fir", Data: map[string]any{"total": count * 10}},
		}
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", event)
	}
	return nil
}

func main() {
	c := fir.NewController("fir-range", fir.DevelopmentMode(true))
	http.Handle("/", c.Handler(&Range{}))
	log.Println("listening on http://localhost:9867")
	http.ListenAndServe(":9867", nil)
}
