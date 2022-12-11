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

func index() fir.RouteOptions {
	return fir.RouteOptions{
		fir.Content("app.html"),
		fir.OnEvent("query", func(ctx fir.RouteContext) error {
			req := new(queryRequest)
			if err := ctx.Bind(req); err != nil {
				return err
			}
			cities := map[string]any{"cities": getCities(req.Query)}
			return ctx.DOM().ReplaceKV("cities", cities)
		}),
	}
}

func main() {
	c := fir.NewController("fir-search", fir.DevelopmentMode(true))
	http.Handle("/", c.RouteFunc(index))
	log.Println("listening on http://localhost:9867")
	http.ListenAndServe(":9867", nil)
}
