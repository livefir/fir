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
	load := func(ctx fir.Context) error {
		return ctx.KV("cities", cities)
	}

	query := func(ctx fir.Context) error {
		req := new(queryRequest)
		if err := ctx.Bind(req); err != nil {
			return err
		}
		data := map[string]any{"cities": getCities(req.Query)}
		return ctx.ReplaceKV("cities", data)
	}

	return fir.RouteOptions{
		fir.Content("app.html"),
		fir.OnEvent("query", query),
		fir.OnLoad(load),
	}
}

func main() {
	c := fir.NewController("autocomplete", fir.DevelopmentMode(true))
	http.Handle("/", c.RouteFunc(index))
	log.Println("listening on http://localhost:9867")
	http.ListenAndServe(":9867", nil)
}
