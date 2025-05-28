package routing

import (
	"fmt"
	"log"
	"net/http"

	"github.com/livefir/fir"
)

func home() fir.RouteOptions {
	return fir.RouteOptions{
		fir.Content("routes/page.html"),
		fir.Layout("routes/layout.html"),
		fir.Partials("routes/partials"),
		fir.OnLoad(func(ctx fir.RouteContext) error {
			return ctx.Data(map[string]any{"title": "Home"})
		}),
	}
}

func about() fir.RouteOptions {
	return fir.RouteOptions{
		fir.Content("routes/about"),
		fir.Layout("routes/layout.html"),
		fir.Partials("routes/partials"),
	}
}

func Index() fir.RouteOptions {
	return home()
}

func NewRoute() fir.RouteOptions {
	return Index()
}

func Run(port int) error {
	c := fir.NewController("routing", fir.DevelopmentMode(true))
	http.Handle("/", c.RouteFunc(home))
	http.Handle("/about", c.RouteFunc(about))
	log.Printf("Routing example listening on http://localhost:%d", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
