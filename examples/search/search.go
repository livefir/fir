package search

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/livefir/fir/internal/dev"

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

func Index() fir.RouteOptions {
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

func NewRoute() fir.RouteOptions {
	return Index()
}

func Run(port int) error {
	dev.SetupAlpinePluginServer()
	c := fir.NewController("fir-search", fir.DevelopmentMode(true))
	http.Handle("/", c.RouteFunc(Index))
	log.Printf("Search example listening on http://localhost:%d", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
