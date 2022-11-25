package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/adnaan/fir"
)

var cities = []string{
	"Paris",
	"Amsterdam",
	"Berlin",
	"New York",
	"Delhi",
	"Beijing",
	"London",
	"Rome",
	"Athens",
	"Seoul",
}

func getCities(str string) []string {
	if str == "" {
		return nil
	}
	var result []string
	for _, city := range cities {
		if strings.HasPrefix(strings.ToLower(city), strings.ToLower(str)) {
			result = append(result, city)
		}
	}
	return result
}

type queryRequest struct {
	Query string `json:"query"`
}

type search struct{}

func (s *search) Options() []fir.RouteOption {
	return []fir.RouteOption{
		fir.Content("app.html"),
		fir.OnEvent("query", s.query),
		fir.OnLoad(s.onLoad),
	}
}

func (s *search) onLoad(e fir.Event, r fir.RouteRenderer) error {
	return r(fir.M{"cities": cities})
}

func (s *search) query(e fir.Event, r fir.PatchRenderer) error {
	req := new(queryRequest)
	if err := e.DecodeParams(req); err != nil {
		return err
	}
	return r(fir.Morph("#cities", "cities", fir.M{"cities": getCities(req.Query)}))

}

func main() {
	c := fir.NewController("fir-autocomplete", fir.DevelopmentMode(true))
	http.Handle("/", c.Route(&search{}))
	log.Println("listening on http://localhost:9867")
	http.ListenAndServe(":9867", nil)
}
