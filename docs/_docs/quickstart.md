---
title: Quickstart
permalink: /docs/quickstart/
redirect_from:
  - /docs/
---


Lets spend the next 15 minutes creating a new `reactive` counter app. If you want to skip ahead and look at final code, its here: [examples/counter/main.go](https://github.com/adnaan/fir/blob/main/examples/counter/main.go)

## Prerequisites

Have you installed [Go](https://go.dev/doc/install) ? If yes, we are good to go.

## Creating a new app

The `fir` library concerns itself with only the view controller so starting off is as easy as mounting a view on the `fir` controller:


```go
package main

import (
	"log"
	"net/http"

	"github.com/adnaan/fir"
)

func main() {
	controller := fir.NewController("counter_app", fir.DevelopmentMode(true))
	http.Handle("/", controller.Handler(&fir.DefaultView{}))
	http.ListenAndServe(":9867", nil)
}

```

Copy the above snippet in a `main.go` file and run `go run main.go`. Open [localhost:9867](http://localhost:9867) to see the running app.

We have created a controller and registered a `DefaultView` by calling `controller.Handler(&fir.HelloView{})`. The `contoller.Handler` method accepts a [View](https://pkg.go.dev/github.com/adnaan/fir#View) interface. `fir.DefaultView` satisfies the methods for the `View` interface with default values.

The fir library doesn't manage routing so you can bring your favorite routing library to actually route requests to the view. Here we keep it simple and mount the `http.HandlerFunc` returned by `controller.Handler` on the `/` route: `http.Handle("/", c.Handler(&fir.DefaultView{}))`

## Creating a new view

We want to build a counter app. To do this we want to create a new view and replace `DefaultView`.

This is how we do that:

```go
type CounterView struct {
	fir.DefaultView
	count int32
}

func (c *CounterView) Content() string {
	return "A counter app"
}

```

In the above snippet we have created a new struct, `CounterView` and embedded a `fir.DefaultView` type in it to satisfy the `View` interface.

```go

package main

import (
	"log"
	"net/http"

	"github.com/adnaan/fir"
)

type Counter struct {
	count int32
}

type CounterView struct {
	fir.DefaultView
	model *Counter
}

func (c *CounterView) Content() string {
	return "A counter app"
}

func main() {
	controller := fir.NewController("counter_app", fir.DevelopmentMode(true))
	http.Handle("/", controller.Handler(&CounterView{model: &Counter{}}))
	http.ListenAndServe(":9867", nil)
}
```

Run the above code to see the changes at [localhost:9867](http://localhost:9867).

## User interaction

`Fir` has a companion javascript library which lets you send browser events to the server. You can use these events to change server state(in our case: `model *Counter`) and make partial page updates without a page reload.

{% raw %}
```html
<div>
    <button class="button has-background-primary" @click="$fir.emit('inc')">+
    </button>
    <button class="button has-background-primary" @click="$fir.emit('dec')">-
    </button>
</div>
```
{% endraw %}   

In the above snippet, we use the custom Alpinejs magic function, `$fir.emit` to send an event to the server on a button click. Shortly we will see how to handle this event to change state on the server, followed by updating a count on the web page.

## Render view

Before we go ahead, lets expand the above snippet to a full html page.

<details markdown="block">
  <summary>
    Expand html page
  </summary>

{% raw %}
```html
<!DOCTYPE html>
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
					<button class="button has-background-primary" @click="$fir.emit('inc')">+</button>
					<button class="button has-background-primary" @click="$fir.emit('dec')">-</button>
				</div>
			</div>
		</div>
	</div>
</body>

</html>

```
{% endraw %}

</details>

The html page includes the `fir` JS library which helps you add tiny bits of interactivity to the page. The library is an  [Alpinejs](https://alpinejs.dev) plugin and ships with extra direcitives(x-* thingy) and magic functions($ thingy).

Let's add the above html page to the `Content` method of our view. The `Content` method can return either a valid filename or html.

<details markdown="block">
  <summary>
    Expand main.go
  </summary>

{% raw %}
```go

package main

import (
	"log"
	"net/http"

	"github.com/adnaan/fir"
)

type Counter struct {
	count int32
}

type CounterView struct {
	fir.DefaultView
}

func (c *CounterView) Content() string {
	return `<!DOCTYPE html>
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
}


func main() {
	controller := fir.NewController("counter_app", fir.DevelopmentMode(true))
	http.Handle("/", controller.Handler(&CounterView{model: &Counter{}}))
	http.ListenAndServe(":9867", nil)
}
```
{% endraw %}

</details>

Running the above code, show render two buttons but nothing else. We want to show an initial count on the page. To do this, we use Go's `html/template` to hydrate some data into our page by overriding the `OnGet` method of the `View` interface.

```go
func (c *CounterView) OnGet(_ http.ResponseWriter, _ *http.Request) (fir.Status, fir.Data) {
	return fir.Status{Code: 200}, fir.Data{
		"count": c.Value(),
	}
}
```

By default, `fir.Data` was zero value. After overriding `OnGet` we are initialising it with a `count` value. The page is then passed through `html/template` and rendered. This is the standard way of rendering html templates in Go so this should be recongisable.

{% raw %}
```html
{{block "count" .}}<div id="count">{{.count}}</div>{{end}}
```

`block` is a `html/template` built-in shorthand for defining and using a template.

This : 
```go 
{{block "count" .}}<div id="count">{{.count}}</div>{{end}}
```
is same as:
```go
{{define "count"}}<div id="count">{{.count}}</div>{{end}}
{{ template "count"}}
```
{% endraw %}


## Update parts of the view

In response to user interaction we want to update a part of our web page to display a result. `Fir` allows you to `patch` targeteted areas of the DOM without a page reload. Lets see it in action.

{% raw %}
```html
<div>
	{{block "count" .}}
		<div id="count">{{.count}}</div>
	{{end}}
	<button 
		class="button has-background-primary" 
		@click="$fir.emit('inc')">
		+
	</button>
	<button 
		class="button has-background-primary" 
		@click="$fir.emit('dec')">
		-
	</button>
</div>
```
{% endraw %}

When the `+` button is clicked, an event `inc` is sent to the server which sends backs a `patch` instruction back to the page.

```go

func (c *Counter) Inc() fir.Patch {
	return fir.Morph{
		Selector: "#count",
		Template: "count",
		Data:     fir.Data{"count": atomic.AddInt32(&c.count, 1)},
	}
}

func (c *Counter) Dec() fir.Patch {
	return fir.Morph{
		Selector: "#count",
		Template: "count",
		Data:     fir.Data{"count": atomic.AddInt32(&c.count, -1)},
	}
}

func (c *CounterView) OnEvent(event fir.Event) fir.Patchset {
	switch event.ID {
	case "inc":
		return fir.Patchset{c.model.Inc()}
	case "dec":
		return fir.Patchset{c.model.Dec()}
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", event)
	}
	return nil
}
```

`fir.Morph` is a `patch` which hydrates the new count value to the template `count`(i.e. {% raw %} `{{block "count" .}}`{%endraw%}) on the server and instructs the javascript client library to update(morph) the `<div id="count">`. 


The updated `main.go` should now be fully working counter example.

<details markdown="block">
  <summary>
    Expand main.go
  </summary>

{% raw %}
```go

package main

import (
	"log"
	"net/http"
	"sync/atomic"

	"github.com/adnaan/fir"
)

type Counter struct {
	count int32
}

func (c *Counter) Inc() fir.Patch {
	return fir.Morph{
		Selector: "#count",
		Template: "count",
		Data:     fir.Data{"count": atomic.AddInt32(&c.count, 1)},
	}
}

func (c *Counter) Dec() fir.Patch {
	return fir.Morph{
		Selector: "#count",
		Template: "count",
		Data:     fir.Data{"count": atomic.AddInt32(&c.count, -1)},
	}
}

func (c *Counter) Value() int32 {
	return atomic.LoadInt32(&c.count)
}

type CounterView struct {
	fir.DefaultView
	model *Counter
}

func (c *CounterView) Content() string {
	return `<!DOCTYPE html>
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
}

func (c *CounterView) OnGet(_ http.ResponseWriter, _ *http.Request) (fir.Status, fir.Data) {
	return fir.Status{Code: 200}, fir.Data{
		"count": c.model.Value(),
	}
}

func (c *CounterView) OnEvent(event fir.Event) fir.Patchset {
	switch event.ID {
	case "inc":
		return fir.Patchset{c.model.Inc()}
	case "dec":
		return fir.Patchset{c.model.Dec()}
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", event)
	}
	return nil
}

func main() {
	controller := fir.NewController("counter_app", fir.DevelopmentMode(true))
	http.Handle("/", controller.Handler(&CounterView{model: &Counter{}}))
	http.ListenAndServe(":9867", nil)
}


```
{% endraw %}

</details>

---

If you want to skip ahead and look at final code for the optionals, its here: [examples/counter-ticker/main.go](https://github.com/adnaan/fir/blob/main/examples/counter-ticker/main.go)

## Optional: Layouts

Right now are, our html page is one big blob. We might want to separate out the layout from the content for reusability. To do this we need to override the `Layout` method of the `View` interface.

{% raw %}
```go
type LayoutView struct {
	fir.DefaultView
}

func (l *LayoutView) Content() string {
	return `{{define "content"}}<div>world</div>{{ end }}`
}

func (l *LayoutView) Layout() string {
	return `<div>Hello: {{template "content" .}}</div>`
}
```

Notice the `{{template "content" .}}`. `Fir` looks for an equivalent defined template in `Content` which here is: `{{define "content"}}<div>world</div>{{ end }}`. By default it looks for a template named `content` but this can be overriden by returned a different layout name in `LayoutContentName() string` 

{% endraw %}

## Optional: Live Ticker

On user interaction events, `OnEvent` sends back a `patchset` which patches the interesting parts of the page. It would be nice to update the page when something changes for a user on the server(e.g. notifications, stock ticker, chat message etc.). Using the `fir` library its possible to `stream` a `patch` over websockets or server-sent events(SSE).

Override the `Stream` method of the `View` interface to return a receive only channel(`<- chan Patch`). When a `patch` is sent to this channel its sent to the client library where its executed to update the page.

```go

type CounterView struct {
	fir.DefaultView
	model  *Counter
	stream chan fir.Patch
	sync.RWMutex
}

func (c *CounterView) Stream() <-chan fir.Patch {
	return c.stream
}

...

http.Handle("/", controller.Handler(&CounterView{stream: make(chan fir.Patch)}))
```

We can send a `patch` to the stream.

```go
c.stream <- fir.Morph{...}
```

Lets expand the `counter` example to add a last updated ticker to the page. The ticker should update every second and tell us when was count last updated.

For this example, we use a different `patch` type: `fir.Store{}`. 


```go
func NewCounterView() *CounterView {
	stream := make(chan fir.Patch)
	ticker := time.NewTicker(time.Second)
	c := &CounterView{stream: stream, model: &Counter{}}

	go func() {
		for ; true; <-ticker.C {
			patch, err := c.model.Updated()
			if err != nil {
				continue
			}
			stream <- patch
		}
	}()
	return c
}
```

```html
<div>Count updated: <span x-text="$store.fir.count_updated || 0"></span> seconds ago</div>
```

`fir.Store{}` updates the global [alpinejs $store](https://alpinejs.dev/globals/alpine-store). Since its reactive, the above html snippet automatically updates.

See the complete working example:

<details markdown="block">
  <summary>
    Expand main.go
  </summary>

{% raw %}

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/adnaan/fir"
)

type Counter struct {
	count   int32
	updated time.Time
	sync.RWMutex
}

func (c *Counter) Inc() fir.Patch {
	c.Lock()
	defer c.Unlock()
	c.count += 1
	c.updated = time.Now()
	return fir.Morph{
		Selector: "#count",
		Template: "count",
		Data:     fir.Data{"count": c.count},
	}
}

func (c *Counter) Dec() fir.Patch {
	c.Lock()
	defer c.Unlock()
	c.count -= 1
	c.updated = time.Now()
	return fir.Morph{
		Selector: "#count",
		Template: "count",
		Data:     fir.Data{"count": c.count},
	}
}

func (c *Counter) Updated() (fir.Patch, error) {
	c.RLock()
	defer c.RUnlock()
	if c.updated.IsZero() {
		return nil, fmt.Errorf("time is zero")
	}
	return fir.Store{
		Name: "fir",
		Data: map[string]any{"count_updated": time.Since(c.updated).Seconds()},
	}, nil
}

func (c *Counter) Count() int32 {
	c.RLock()
	defer c.RUnlock()
	return c.count
}

func NewCounterView() *CounterView {
	stream := make(chan fir.Patch)
	ticker := time.NewTicker(time.Second)
	c := &CounterView{stream: stream, model: &Counter{}}

	go func() {
		for ; true; <-ticker.C {
			patch, err := c.model.Updated()
			if err != nil {
				continue
			}
			stream <- patch
		}
	}()
	return c
}

type CounterView struct {
	fir.DefaultView
	model  *Counter
	stream chan fir.Patch
	sync.RWMutex
}

func (c *CounterView) Stream() <-chan fir.Patch {
	return c.stream
}

func (c *CounterView) Content() string {
	return `
{{define "content" }} 
<div class="my-6" style="height: 500px">
	<div class="columns is-mobile is-centered is-vcentered">
		<div x-data class="column is-one-third-desktop has-text-centered is-narrow">
			<div>
				<div>Count updated: <span x-text="$store.fir.count_updated || 0"></span> seconds ago</div>
				<hr>
				{{block "count" .}}<div id="count">{{.count}}</div>{{end}}
				<button class="button has-background-primary" @click="$fir.emit('inc')">+
				</button>
				<button class="button has-background-primary" @click="$fir.emit('dec')">-
				</button>
			</div>
		</div>
	</div>
</div>
{{end}}`
}

func (c *CounterView) Layout() string {
	return `<!DOCTYPE html>
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
		{{template "content" .}}
	</body>
	
	</html>`
}

func (c *CounterView) OnGet(_ http.ResponseWriter, _ *http.Request) (fir.Status, fir.Data) {
	return fir.Status{Code: 200}, fir.Data{
		"count": c.model.Count(),
	}
}

func (c *CounterView) OnEvent(event fir.Event) fir.Patchset {
	switch event.ID {
	case "inc":
		return fir.Patchset{c.model.Inc()}
	case "dec":
		return fir.Patchset{c.model.Dec()}
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", event)
	}

	return nil
}

func main() {
	controller := fir.NewController("counter_app", fir.DevelopmentMode(true))
	http.Handle("/", controller.Handler(NewCounterView()))
	http.ListenAndServe(":9867", nil)
}
	
```

{% endraw %}

</details>

Run the above main.go and go to [localhost:9867](http://localhost:9867/). Incrementing or decrementing the count should update the ticker.
