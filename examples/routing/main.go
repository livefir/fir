package main

import (
	"net/http"

	"github.com/adnaan/fir"
)

func home() []fir.RouteOption {
	return []fir.RouteOption{
		fir.Content("./routes/page.html"),
		fir.Layout("./routes/layout.html"),
	}
}

func about() []fir.RouteOption {
	return []fir.RouteOption{
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
