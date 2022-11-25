package main

import (
	"net/http"
	"sync/atomic"

	"github.com/adnaan/fir"
)

var content = `<!DOCTYPE html>
<html lang="en">

<head>
	<title>{{.app_name}}</title>
	<meta charset="UTF-8">
	<meta name="description" content="A counter app">
	<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bulma@0.9.4/css/bulma.min.css" />
	<script defer src="https://unpkg.com/@adnaanx/fir@latest/dist/fir.min.js"></script>
	<script defer src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"></script>
</head>

<body>
	<div class="my-6" style="height: 500px">
		<div class="columns is-mobile is-centered is-vcentered">
			<div x-data class="column is-one-third-desktop has-text-centered is-narrow">
				<div>
					{{block "count" .}}<div id="count">{{.count}}</div>{{end}}
					<button class="button has-background-primary" @click="$fir.emit('inc')">+
					</button>
					<button class="button has-background-primary" @click="$fir.emit('dec')">-
					</button>
				</div>
			</div>
		</div>
	</div>
</body>

</html>`

type index struct {
	value int32
}

func (i *index) load(e fir.Event, r fir.RouteRenderer) error {
	return r(fir.M{"count": atomic.LoadInt32(&i.value)})
}

func (i *index) inc(e fir.Event, r fir.PatchRenderer) error {
	return r(
		fir.Morph(
			"#count",
			"count",
			fir.M{"count": atomic.AddInt32(&i.value, 1)},
		))
}

func (i *index) dec(e fir.Event, r fir.PatchRenderer) error {
	return r(
		fir.Morph(
			"#count",
			"count",
			fir.M{"count": atomic.AddInt32(&i.value, -1)},
		))
}

func (i *index) Options() []fir.RouteOption {
	return []fir.RouteOption{
		fir.ID("counter"),
		fir.Content(content),
		fir.OnLoad(i.load),
		fir.OnEvent("inc", i.inc),
		fir.OnEvent("dec", i.dec),
	}
}

func main() {
	controller := fir.NewController("counter_app", fir.DevelopmentMode(true))
	http.Handle("/", controller.Route(&index{}))
	http.ListenAndServe(":9867", nil)
}
