# Fir

A Go toolkit to build reactive web interfaces using: [Go](https://go.dev/), [html/template](https://pkg.go.dev/html/template) and [alpinejs](https://alpinejs.dev/). 

**Status**: This is a work in progress. Checkout examples to see what works today: [examples](./examples/)

## Example

```go
package main

import (
	"net/http"
	"sync/atomic"

	"github.com/livefir/fir"
)

var content = `
<!DOCTYPE html>
<html lang="en">
<head>
	<script defer src="https://unpkg.com/@adnaanx/fir@latest/dist/fir.min.js"></script>
	<script defer src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"></script>
</head>

<body>
	<div x-data>
		{{block "count" .}}
			<div id="count"
                 @inc.window="$fir.replaceEl()"
                @dec.window="$fir.replaceEl()">
				{{.count}}
			</div>
		{{end}}
		<button @click="$dispatch('inc')">+</button>
		<button @click="$dispatch('dec')">-</button>
	</div>
</body>
</html>`

func index() fir.RouteOptions {
	var count int32
	return fir.RouteOptions{
		fir.ID("counter"),
		fir.Content(content),
		fir.OnLoad(func(ctx fir.RouteContext) error {
			return ctx.KV("count", atomic.LoadInt32(&count))
		}),
		fir.OnEvent("inc", func(ctx fir.RouteContext) error {
			return ctx.KV("count", atomic.AddInt32(&count, 1))
		}),
		fir.OnEvent("dec", func(ctx fir.RouteContext) error {
			return ctx.KV("count", atomic.AddInt32(&count, -1))
		}),
	}
}

func main() {
	controller := fir.NewController("counter_app", fir.DevelopmentMode(true))
	http.Handle("/", controller.RouteFunc(index))
	http.ListenAndServe(":9867", nil)
}
```
