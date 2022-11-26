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

func index() []fir.RouteOption {
	return []fir.RouteOption{
		fir.Content("app.html"),
		fir.OnEvent("query", func(e fir.Event, r fir.PatchRenderer) error {
			req := new(queryRequest)
			if err := e.DecodeParams(req); err != nil {
				return err
			}
			cities := fir.M{"cities": getCities(req.Query)}
			return r(fir.Morph("#cities", fir.Block("cities", cities)))
		}),
	}
}

func main() {
	c := fir.NewController("fir-search", fir.DevelopmentMode(true))
	http.Handle("/", c.RouteFunc(index))
	log.Println("listening on http://localhost:9867")
	http.ListenAndServe(":9867", nil)
}
