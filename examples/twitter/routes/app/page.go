package app

import "github.com/adnaan/fir"

func Route() []fir.RouteOption {
	return []fir.RouteOption{
		fir.Content("./routes/app/page.html"),
		fir.Layout("./routes/layout.html"),
	}
}
