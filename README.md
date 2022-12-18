# Fir

[![Go Reference](https://pkg.go.dev/badge/github.com/livefir/fir.svg)](https://pkg.go.dev/github.com/livefir/fir) 
[![npm version](https://badge.fury.io/js/@livefir%2Ffir.svg)](https://badge.fury.io/js/@livefir%2Ffir)

**A Go toolkit to build reactive web interfaces using: [Go](https://go.dev/), [html/template](https://pkg.go.dev/html/template) and [alpinejs](https://alpinejs.dev/).**


Fir is a toolkit for building server-rendered HTML applications and progressively enhancing them to enable real-time user experiences. It is intended for developers who want to build real-time web apps using Go, server-rendered HTML (html/template), CSS, and sprinkles of declarative javascript (Alpine.js). The toolkit can be used to build a completely server-rendered web application with zero javascript, and the same app can then be progressively enhanced on the client to a real-time dynamic app with little bits of javascript while still using Go's html/template engine on the server. Fir can be used to build various types of web applications, including static websites like landing pages or blogs, interactive CRUD apps like ticket helpdesks, and real-time apps like metrics dashboards or social media streams.


**Status**: This is a work in progress. Checkout examples to see what works today: [examples](./examples/)

## Example

```go
package main

import (
	"net/http"
	"sync/atomic"

	"github.com/livefir/fir"
)

func index() fir.RouteOptions {
	var count int32
	return fir.RouteOptions{
		fir.ID("counter"),
		fir.Content("count.html"),
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

```html
<!DOCTYPE html>
<html lang="en">
    <head>
        <script
            defer
            src="https://unpkg.com/@livefir/fir@latest/dist/fir.min.js"></script>
        <script
            defer
            src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"></script>
    </head>

    <body>
        <div x-data>
            {{ block "count" . }}
                <div
                    id="count"
                    @inc.window="$fir.replaceEl()"
                    @dec.window="$fir.replaceEl()">
                    {{ .count }}
                </div>
            {{ end }}
            <button @click="$dispatch('inc')">+</button>
            <button @click="$dispatch('dec')">-</button>
        </div>
    </body>
</html>
```
