package rangecounter

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/livefir/fir"
)

type countRequest struct {
	Count string `json:"count"`
}

func Index() fir.RouteOptions {
	return fir.RouteOptions{
		fir.Content("app.html"),
		fir.OnLoad(func(ctx fir.RouteContext) error {
			return ctx.Data(map[string]any{"total": 0})
		}),
		fir.OnEvent("update", func(ctx fir.RouteContext) error {
			req := new(countRequest)
			if err := ctx.Bind(req); err != nil {
				return err
			}
			count, err := strconv.Atoi(req.Count)
			if err != nil {
				return err
			}
			return ctx.Data(map[string]any{"total": count * 10})
		}),
	}
}

func NewRoute() fir.RouteOptions {
	return Index()
}

func Run(port int) error {
	c := fir.NewController("fir-range", fir.DevelopmentMode(true))
	http.Handle("/", c.RouteFunc(Index))
	log.Printf("Range example listening on http://localhost:%d", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
