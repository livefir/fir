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

func (s *Search) OnEvent(event fir.Event) fir.Patchset {
	switch event.ID {
	case "query":
		req := new(QueryRequest)
		if err := event.DecodeParams(req); err != nil {
			return fir.PatchError(err, "failed to decode query request")
		}
		return fir.Patchset{
			fir.Morph{
				Selector: "#list_cities",
				HTML: &fir.Render{
					Template: "cities",
					Data: map[string]any{
						"cities": getCities(req.Query),
					},
				},
			},
		}
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", event)
	}
	return nil
}

func main() {
	c := fir.NewController("fir-autocomplete", fir.DevelopmentMode(true))
	http.Handle("/", c.Handler(&Search{}))
	log.Println("listening on http://localhost:9867")
	http.ListenAndServe(":9867", nil)
}
