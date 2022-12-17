package main

import (
	"net/http"
	"sync/atomic"

	"github.com/livefir/fir"
)

var content = `<!DOCTYPE html>
<html lang="en">
<head>
	<script defer src="https://unpkg.com/@adnaanx/fir@latest/dist/fir.min.js"></script>
	<script defer src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"></script>
</head>

<body>
	<div x-data>
		{{block "count" .}}
			<div id="count" @inc.window="$fir.replaceEl()" @dec.window="$fir.replaceEl()">
				{{.count}}
			</div>
		{{end}}
		<button @click="$dispatch('inc')">+</button>
		<button @click="$dispatch('dec')">-</button>
	</div>
</body>
</html>`

func index() fir.RouteOptions {
	var value int32

	load := func(ctx fir.RouteContext) error {
		return ctx.Data(map[string]any{"count": atomic.LoadInt32(&value)})
	}

	inc := func(ctx fir.RouteContext) error {
		return ctx.Data(map[string]any{"count": atomic.AddInt32(&value, 1)})
	}

	dec := func(ctx fir.RouteContext) error {
		return ctx.Data(map[string]any{"count": atomic.AddInt32(&value, -1)})
	}

	return fir.RouteOptions{
		fir.ID("counter"),
		fir.Content(content),
		fir.OnLoad(load),
		fir.OnEvent("inc", inc),
		fir.OnEvent("dec", dec),
	}
}

func main() {
	controller := fir.NewController("counter_app", fir.DevelopmentMode(true))
	http.Handle("/", controller.RouteFunc(index))
	http.ListenAndServe(":9867", nil)
}
