package projects

import (
	"github.com/livefir/fir"
	"github.com/livefir/fir/examples/fira/ent"
)

func Show(db *ent.Client) fir.RouteFunc {
	return func() fir.RouteOptions {
		return fir.RouteOptions{
			fir.ID("project"),
			fir.Content("routes/projects/show.html"),
			fir.Layout("routes/layout.html"),
			fir.Partials("routes/partials"),
		}
	}
}
