package main

import (
	"log"
	"net/http"
	"strconv"

	pwc "github.com/adnaan/pineview/controller"
)

type CountRequest struct {
	Count string `json:"count"`
}

type Range struct {
	pwc.DefaultView
}

func (r *Range) Content() string {
	return "app.html"
}

func (r *Range) OnMount(_ http.ResponseWriter, _ *http.Request) (pwc.Status, pwc.M) {
	return pwc.Status{Code: 200}, pwc.M{
		"total": 0,
	}
}

func (r *Range) OnLiveEvent(ctx pwc.Context) error {
	switch ctx.Event().ID {
	case "update":
		req := new(CountRequest)
		if err := ctx.Event().DecodeParams(req); err != nil {
			return err
		}
		count, err := strconv.Atoi(req.Count)
		if err != nil {
			return err
		}
		ctx.Store().UpdateProp("total", count*10)
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", ctx.Event())
	}
	return nil
}

func main() {
	glvc := pwc.Websocket("goliveview-range", pwc.DevelopmentMode(true))
	http.Handle("/", glvc.Handler(&Range{}))
	log.Println("listening on http://localhost:9867")
	http.ListenAndServe(":9867", nil)
}
