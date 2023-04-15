# Fir

[![Go Reference](https://pkg.go.dev/badge/github.com/livefir/fir.svg)](https://pkg.go.dev/github.com/livefir/fir) 
[![npm version](https://badge.fury.io/js/@livefir%2Ffir.svg)](https://badge.fury.io/js/@livefir%2Ffir)

**A Go toolkit to build reactive web interfaces using: [Go](https://go.dev/), [html/template](https://pkg.go.dev/html/template) and [alpinejs](https://alpinejs.dev/).**

**Status**: This is a work in progress. Checkout examples to see what works today: [examples](./examples/)


## Example

Using fir's alpinejs plugin, the page below has been progressively enhanced to a real-time single page app. Open two tabs to see the count update in both. Disable javascript in your browser to see it still work without the enhancements.

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
            <div
                @fir:inc:ok:count.window="$fir.replace()"
                @fir:dec:ok:count.window="$fir.replace()">
                {{ block "count" . }}
                    <div>Count: {{ .count }}</div>
                {{ end }}
            </div>
            <form method="post" @submit.prevent="$fir.submit()">
                <button formaction="/?event=inc" type="submit">+</button>
                <button formaction="/?event=dec" type="submit">-</button>
            </form>
        </div>
    </body>
</html>
```

In the above example, `@fir:inc:ok:count.window="$fir.replace()"` marks content of that div to be replaced by the content of `block count`. Event namespacing is used to indicate the server renderer which `block` to re-render and send on a successful event response(`ok`). The allowed event namespace format is `@fir:<event-name>:<ok|error>:<block-name>`. 

`$fir.submit` is a helper which prevents the form submission and dispatches browser events `inc` and `dec` which is then captured by `@fir:inc:ok:count.window, @fir:dec:ok:count.window` listeners. It also captures any form data and attaches it to the event before its sent to the server by `$fir.replace`. In this example we don't have any form data.

## About Fir

Fir is a toolkit for building server-rendered HTML applications and progressively enhancing them to enable real-time user experiences. It is intended for developers who want to build real-time web apps using Go, server-rendered HTML (html/template), CSS, and sprinkles of declarative javascript (Alpine.js). The toolkit can be used to build a completely server-rendered web application with zero javascript, and the same app can then be progressively enhanced on the client to a real-time dynamic app with little bits of javascript while still using Go's html/template engine on the server. Fir can be used to build various types of web applications, including static websites like landing pages or blogs, interactive CRUD apps like ticket helpdesks, and real-time apps like metrics dashboards or social media streams.

Fir enhances a standard `html/template` web page with `alpine.js` allowing predefined parts of the page updatable on user interaction. A `html/template` page is decomposed into updatable parts using the `block` action. This allows Fir to re-compile the targeted block on the server and update the web page over the *wire*(http & websocket) and without page reloads. 
The HTML itself is largely quite standard i.e. its free of magics or special attributes except what’s exposed by the alpinejs plugin. The plugin exposes a small API and is designed to be unobtrusive and easy to remove incase one wants to migrate the HTML to another framework. 