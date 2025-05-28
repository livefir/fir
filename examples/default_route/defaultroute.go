package defaultroute

import (
	"fmt"
	"log"
	"net/http"

	"github.com/livefir/fir"
)

func Index() fir.RouteOptions {
	return nil
}

func NewRoute() fir.RouteOptions {
	return Index()
}

func Run(port int) error {
	c := fir.NewController("default", fir.DevelopmentMode(true))
	http.Handle("/", c.RouteFunc(Index))
	log.Printf("Default route example listening on http://localhost:%d", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
