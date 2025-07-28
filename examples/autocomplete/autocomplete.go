package autocomplete

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
	return map[string]any{"cities": result}
}

type queryRequest struct {
	Query string `json:"query"`
}

func Index() fir.RouteOptions {
	load := func(ctx fir.RouteContext) error {
		return ctx.Data(cities)
	}

	query := func(ctx fir.RouteContext) error {
		req := new(queryRequest)
		if err := ctx.Bind(req); err != nil {
			return err
		}
		return ctx.Data(filterCities(req.Query))
	}

	return fir.RouteOptions{
		fir.Content("app.html"),
		fir.OnLoad(load),
		fir.OnEvent("query", query),
	}
}

func NewRoute() fir.RouteOptions {
	return Index()
}

func Run(port int) error {
	dev.SetupAlpinePluginServer()
	c := fir.NewController("autocomplete", fir.DevelopmentMode(true))
	http.Handle("/", c.RouteFunc(Index))
	log.Printf("Autocomplete example listening on http://localhost:%d", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
