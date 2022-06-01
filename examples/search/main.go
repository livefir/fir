package main

import (
	"log"
	"net/http"
	"strings"

	pwc "github.com/adnaan/fir/controller"
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
	pwc.DefaultView
}

func (s *Search) Content() string {
	return "app.html"
}

func (s *Search) Partials() []string {
	return []string{"cities.html"}
}

func (s *Search) OnRequest(_ http.ResponseWriter, _ *http.Request) (pwc.Status, pwc.Data) {
	return pwc.Status{Code: 200}, nil
}

func (s *Search) OnEvent(st pwc.Socket) error {
	switch st.Event().ID {
	case "search":
		req := new(QueryRequest)
		if err := st.Event().DecodeParams(req); err != nil {
			return err
		}
		st.Morph("#cities", "cities", pwc.Data{
			"cities": getCities(req.Query),
		})
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", st.Event())
	}
	return nil
}

func main() {
	glvc := pwc.Websocket("fir-search", pwc.DevelopmentMode(true))
	http.Handle("/", glvc.Handler(&Search{}))
	log.Println("listening on http://localhost:9867")
	http.ListenAndServe(":9867", nil)
}
