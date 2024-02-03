package main

import (
	"net/http"

	"github.com/livefir/fir"
)

func home() fir.RouteOptions {
	return fir.RouteOptions{
		fir.Content("./routes/page.html"),
		fir.Layout("./routes/layout.html"),
		fir.OnLoad(func(ctx fir.RouteContext) error {
			return ctx.Data(map[string]any{"title": "Home"})
		}),
	}
}

func about() fir.RouteOptions {
	return fir.RouteOptions{
		fir.Content("./routes/about"),
		fir.Layout("./routes/layout.html"),
	}
}

func main() {
	c := fir.NewController("routing", fir.DevelopmentMode(true))
	http.Handle("/", c.RouteFunc(home))
	http.Handle("/about", c.RouteFunc(about))
	http.ListenAndServe(":9867", nil)
}
