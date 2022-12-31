package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/livefir/fir"
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

func filterCities(str string) map[string]any {
	if str == "" {
		return nil
	}
	var result []string
	for _, city := range cities {
		if strings.HasPrefix(strings.ToLower(city), strings.ToLower(str)) {
			result = append(result, city)
		}
	}
	return map[string]any{"cities": result, "query": str}
}

type queryRequest struct {
	Query string `json:"query"`
}

func index() fir.RouteOptions {
	query := func(ctx fir.RouteContext) error {
		req := new(queryRequest)
		if err := ctx.Bind(req); err != nil {
			return err
		}
		return ctx.Data(filterCities(req.Query))
	}
	return fir.RouteOptions{
		fir.Content("app.html"),
		fir.OnLoad(query),
		fir.OnEvent("query", query),
	}
}

func main() {
	c := fir.NewController("fir-search", fir.DevelopmentMode(true))
	http.Handle("/", c.RouteFunc(index))
	log.Println("listening on http://localhost:9867")
	http.ListenAndServe(":9867", nil)
}
