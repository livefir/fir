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
	<script defer src="http://localhost:8000/cdn.js"></script>
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

type counter struct {
	value int32
}

func (c *counter) onLoad(e fir.Event, r fir.RouteRenderer) error {
	return r(fir.M{"count": atomic.LoadInt32(&c.value)})
}

func (c *counter) onInc(e fir.Event, r fir.PatchRenderer) error {
	return r(
		fir.Morph(
			"#count",
			"count",
			fir.M{"count": atomic.AddInt32(&c.value, 1)},
		))
}

func (c *counter) onDec(e fir.Event, r fir.PatchRenderer) error {
	return r(
		fir.Morph(
			"#count",
			"count",
			fir.M{"count": atomic.AddInt32(&c.value, -1)},
		))
}

func (c *counter) opts() []fir.RouteOption {
	return []fir.RouteOption{
		fir.ID("counter"),
		fir.Content(content),
		fir.OnLoad(c.onLoad),
		fir.OnEvent("inc", c.onInc),
		fir.OnEvent("dec", c.onDec),
	}
}

func main() {
	controller := fir.NewController("counter_app", fir.DevelopmentMode(true))
	c := &counter{}
	http.Handle("/", controller.Route(c.opts()...))
	http.ListenAndServe(":9867", nil)
}
