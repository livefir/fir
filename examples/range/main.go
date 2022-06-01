package main

import (
	"log"
	"net/http"
	"strconv"

	pwc "github.com/adnaan/fir/controller"
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

func (r *Range) OnRequest(_ http.ResponseWriter, _ *http.Request) (pwc.Status, pwc.Data) {
	return pwc.Status{Code: 200}, pwc.Data{
		"total": 0,
	}
}

func (r *Range) OnEvent(s pwc.Socket) error {
	switch s.Event().ID {
	case "update":
		req := new(CountRequest)
		if err := s.Event().DecodeParams(req); err != nil {
			return err
		}
		count, err := strconv.Atoi(req.Count)
		if err != nil {
			return err
		}
		s.Store().UpdateProp("total", count*10)
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", s.Event())
	}
	return nil
}

func main() {
	glvc := pwc.Websocket("fir-range", pwc.DevelopmentMode(true))
	http.Handle("/", glvc.Handler(&Range{}))
	log.Println("listening on http://localhost:9867")
	http.ListenAndServe(":9867", nil)
}
