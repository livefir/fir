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

type QueryRequest struct {
	Query string `json:"query"`
}

type Search struct {
	fir.DefaultView
}

func (s *Search) Content() string {
	return "app.html"
}

func (s *Search) Partials() []string {
	return []string{"cities.html"}
}

func (s *Search) OnPatch(event fir.Event) (fir.Patchset, error) {
	switch event.ID {
	case "search":
		req := new(QueryRequest)
		if err := event.DecodeParams(req); err != nil {
			return nil, err
		}
		return fir.Patchset{fir.Morph{
			Template: "cities",
			Selector: "#cities",
			Data: fir.Data{
				"cities": getCities(req.Query),
			},
		}}, nil
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", event)
	}
	return nil, nil
}

func main() {
	c := fir.NewController("fir-search", fir.DevelopmentMode(true))
	http.Handle("/", c.Handler(&Search{}))
	log.Println("listening on http://localhost:9867")
	http.ListenAndServe(":9867", nil)
}
