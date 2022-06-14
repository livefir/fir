---
layout: page
title: Quickstart
permalink: /quickstart/
---


<details open markdown="block">
  <summary>
    Table of contents
  </summary>
  {: .text-delta }
- TOC
{:toc}
</details>

# Quickstart

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


type CounterView struct {
	fir.DefaultView
	count int32
}

func (c *CounterView) Content() string {
	return "A counter app"
}

func main() {
	controller := fir.NewController("counter_app", fir.DevelopmentMode(true))
	http.Handle("/", controller.Handler(CounterView{}))
	http.ListenAndServe(":9867", nil)
}
```

Run the above code to see the changes at [localhost:9867](http://localhost:9867).

## User interaction

`Fir` has a companion javascript library which lets you send browser events to the server. You can use these events to change server state(in our case: `count int32`) and make partial page updates without a page reload.

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
    <script defer src="https://unpkg.com/@adnaanx/fir@latest/dist/cdn.min.js"></script>
</head>

<body>
    <div class="my-6" style="height: 500px">
        <div class="columns is-mobile is-centered is-vcentered">
            <div x-data class="column is-one-third-desktop has-text-centered is-narrow">
                <div>
                    <div id="count" x-text="$store.fir.count || {{.count}}">{{.count}}</div>
                    <button class="button has-background-primary" @click="$fir.emit('inc')">+
                    </button>
                    <button class="button has-background-primary" @click="$fir.emit('dec')">-
                    </button>
                </div>
            </div>
        </div>
    </div>
</body>

</html>

```
{% endraw %}

</details>

The html page includes the `fir` JS library which helps you add tiny bits of interactivity to the page.. The library bundles [Alpinejs](https://alpinejs.dev) and ships with extra direcitives(x-* thingy) and magic functions($ thingy).

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

type CounterView struct {
	fir.DefaultView
	count int32
}

func (c *CounterView) Content() string {
	return `<!DOCTYPE html>
	<html lang="en">
	
	<head>
		<title>{{.app_name}}</title>
		<meta charset="UTF-8">
		<meta name="description" content="A counter app">
		<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bulma@0.9.4/css/bulma.min.css" />
		<script defer src="https://unpkg.com/@adnaanx/fir@latest/dist/cdn.min.js"></script>
	</head>
	
	<body>
		<div class="my-6" style="height: 500px">
			<div class="columns is-mobile is-centered is-vcentered">
				<div x-data class="column is-one-third-desktop has-text-centered is-narrow">
					<div>
						<div id="count" x-text="$store.fir.count || {{.count}}">{{.count}}</div>
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
	http.Handle("/", controller.Handler(CounterView{}))
	http.ListenAndServe(":9867", nil)
}
```
{% endraw %}

</details>

Running the above code, show render two buttons but nothing else. We want to show an initial count on the page. To do this, we use Go's `html/template` to hydrate some date into our page. We override the `OnRequest` method of the `View` interface to do custom rendering.

```go
func (c *CounterView) OnRequest(_ http.ResponseWriter, _ *http.Request) (fir.Status, fir.Data) {
	return fir.Status{Code: 200}, fir.Data{
		"count": c.Value(),
	}
}
```

By default, `fir.Data` was empty. After overriding `OnRequest` we are initialising it with a `count` value. The page is then passed through `html/template` and rendered. This is the standard way of rendering html templates in Go so this should be recongisable. 

{% raw %}
```html
<div id="count" x-text="$store.fir.count || {{.count}}">{{.count}}</div>
```

Here the `{{.count}}` is replaced by `count` set in `fir.Data`. We will come to the `x-text` part shortly. 
{% endraw %}

Since we want to count concurrently, we have used an atomic counter and added a few extra methods(`Inc`, `Dec`) to make our life easier. See the updated code below.

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

type CounterView struct {
	fir.DefaultView
	count int32
}

func (c *CounterView) Inc() int32 {
	atomic.AddInt32(&c.count, 1)
	return atomic.LoadInt32(&c.count)
}

func (c *CounterView) Dec() int32 {
	atomic.AddInt32(&c.count, -1)
	return atomic.LoadInt32(&c.count)
}

func (c *CounterView) Value() int32 {
	return atomic.LoadInt32(&c.count)
}

func (c *CounterView) Content() string {
	return `<!DOCTYPE html>
	<html lang="en">
	
	<head>
		<title>{{.app_name}}</title>
		<meta charset="UTF-8">
		<meta name="description" content="A counter app">
		<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bulma@0.9.4/css/bulma.min.css" />
		<script defer src="https://unpkg.com/@adnaanx/fir@latest/dist/cdn.min.js"></script>
	</head>
	
	<body>
		<div class="my-6" style="height: 500px">
			<div class="columns is-mobile is-centered is-vcentered">
				<div x-data class="column is-one-third-desktop has-text-centered is-narrow">
					<div>
						<div id="count" x-text="$store.fir.count || {{.count}}">{{.count}}</div>
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

func (c *CounterView) OnRequest(_ http.ResponseWriter, _ *http.Request) (fir.Status, fir.Data) {
	return fir.Status{Code: 200}, fir.Data{
		"count": c.Value(),
	}
}

func main() {
	controller := fir.NewController("counter_app", fir.DevelopmentMode(true))
	http.Handle("/", controller.Handler(CounterView{}))
	http.ListenAndServe(":9867", nil)
}
```
{% endraw %}

</details>

Running the above code doesn't do anything new. We need a way to handle events emitted on clicking the `+`, `-` buttons.


## Handling events

Now that we have a way to send events to the server on user interaction, lets handle them to change state on the server. We override the `OnEvent` method of `View` interface.

```go
func (c *CounterView) OnEvent(s fir.Socket) error {
	switch s.Event().ID {
	case "inc":
		s.Store().UpdateProp("count", c.Inc())
	case "dec":
		s.Store().UpdateProp("count", c.Dec())
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", s.Event())
	}
	return nil
}
```

We instruct the client fir library to update [$store.fir](https://alpinejs.dev/globals/alpine-store) with `count` value. To show the updated value, we use [x-text](https://alpinejs.dev/directives/text).

{% raw %}
```html
<div id="count" x-text="$store.fir.count || {{.count}}">{{.count}}</div>
```

Since the initial value of `$store.fir.count` is `undefined` when the page is first rendered, we `or` with `{{.count}}`.
{% endraw %}

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

type CounterView struct {
	fir.DefaultView
	count int32
}

func (c *CounterView) Inc() int32 {
	atomic.AddInt32(&c.count, 1)
	return atomic.LoadInt32(&c.count)
}

func (c *CounterView) Dec() int32 {
	atomic.AddInt32(&c.count, -1)
	return atomic.LoadInt32(&c.count)
}

func (c *CounterView) Value() int32 {
	return atomic.LoadInt32(&c.count)
}

func (c *CounterView) Content() string {
	return `<!DOCTYPE html>
	<html lang="en">
	
	<head>
		<title>{{.app_name}}</title>
		<meta charset="UTF-8">
		<meta name="description" content="A counter app">
		<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bulma@0.9.4/css/bulma.min.css" />
		<script defer src="https://unpkg.com/@adnaanx/fir@latest/dist/cdn.min.js"></script>
	</head>
	
	<body>
		<div class="my-6" style="height: 500px">
			<div class="columns is-mobile is-centered is-vcentered">
				<div x-data class="column is-one-third-desktop has-text-centered is-narrow">
					<div>
						<div id="count" x-text="$store.fir.count || {{.count}}">{{.count}}</div>
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

func (c *CounterView) OnRequest(_ http.ResponseWriter, _ *http.Request) (fir.Status, fir.Data) {
	return fir.Status{Code: 200}, fir.Data{
		"count": c.Value(),
	}
}

func (c *CounterView) OnEvent(s fir.Socket) error {
	switch s.Event().ID {
	case "inc":
		s.Store().UpdateProp("count", c.Inc())
	case "dec":
		s.Store().UpdateProp("count", c.Dec())
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", s.Event())
	}
	return nil
}


func main() {
	controller := fir.NewController("counter_app", fir.DevelopmentMode(true))
	http.Handle("/", controller.Handler(CounterView{}))
	http.ListenAndServe(":9867", nil)
}
```
{% endraw %}

</details>

## Optional: Server only rendering


If you have noticed, the above rendering of `count` html has a problem. We are updating the `<div id="count" ...></div>` twice. Once when we execute the template using `html/template` which replaces `{{.count}}` and after the page has loaded using alpinejs directive `x-text="$store.fir.count"`. We do this because we want to make tiny partial updates to our page by calling, `socket.Store().Update(...)` when an event comes in. 

Instead we can [morph](https://alpinejs.dev/plugins/morph) our target html element. 

{% raw %}
```go

<div>
	{{define "count"}}<div id="count">{{.count}}</div>{{end}}
	{{ template "count" .}}
	<button class="button has-background-primary" @click="$fir.emit('inc')">+
	</button>
	<button class="button has-background-primary" @click="$fir.emit('dec')">-
	</button>
</div>

...


case "inc":
	s.Morph("#count", "count", fir.Data{"count": c.Inc()})
case "dec":
	s.Morph("#count", "count", fir.Data{"count": c.Dec()})
		
```
{% endraw %}

In the above approach, after the first render of the page, alpinejs update has no role. We use `socket.Morph` to send a partial html snippet back to the client `fir` library which then morphs the target element. This approach is heavier on the wire but a lot simpler to reason about and provides a better page rendering performance.

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

type CounterView struct {
	fir.DefaultView
	count int32
}

func (c *CounterView) Inc() int32 {
	atomic.AddInt32(&c.count, 1)
	return atomic.LoadInt32(&c.count)
}

func (c *CounterView) Dec() int32 {
	atomic.AddInt32(&c.count, -1)
	return atomic.LoadInt32(&c.count)
}

func (c *CounterView) Value() int32 {
	return atomic.LoadInt32(&c.count)
}

func (c *CounterView) Content() string {
	return `
<!DOCTYPE html>
<html lang="en">

<head>
	<title>{{.app_name}}</title>
	<meta charset="UTF-8">
	<meta name="description" content="A counter app">
	<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bulma@0.9.4/css/bulma.min.css" />
	<script defer src="https://unpkg.com/@adnaanx/fir@latest/dist/cdn.min.js"></script>
</head>

<body>
	<div class="my-6" style="height: 500px">
		<div class="columns is-mobile is-centered is-vcentered">
			<div x-data class="column is-one-third-desktop has-text-centered is-narrow">
				<div>
					{{define "count"}}<div id="count">{{.count}}</div>{{end}}
					{{ template "count" .}}
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

func (c *CounterView) OnRequest(_ http.ResponseWriter, _ *http.Request) (fir.Status, fir.Data) {
	return fir.Status{Code: 200}, fir.Data{
		"count": c.Value(),
	}
}

func (c *CounterView) OnEvent(s fir.Socket) error {
	switch s.Event().ID {
	case "inc":
		s.Morph("#count", "count", fir.Data{"count": c.Inc()})
	case "dec":
		s.Morph("#count", "count", fir.Data{"count": c.Dec()})
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", s.Event())
	}
	return nil
}

func main() {
	controller := fir.NewController("counter_app", fir.DevelopmentMode(true))
	http.Handle("/", controller.Handler(&CounterView{}))
	http.ListenAndServe(":9867", nil)
}


```
{% endraw %}

</details>

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

Lets update our code to split layout and content.


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

type CounterView struct {
	fir.DefaultView
	count int32
}

func (c *CounterView) Inc() int32 {
	atomic.AddInt32(&c.count, 1)
	return atomic.LoadInt32(&c.count)
}

func (c *CounterView) Dec() int32 {
	atomic.AddInt32(&c.count, -1)
	return atomic.LoadInt32(&c.count)
}

func (c *CounterView) Value() int32 {
	return atomic.LoadInt32(&c.count)
}

func (c *CounterView) Content() string {
	return `{{define "content" }} 
		<div class="my-6" style="height: 500px">
					<div class="columns is-mobile is-centered is-vcentered">
						<div x-data class="column is-one-third-desktop has-text-centered is-narrow">
							<div>
								<div id="count" x-text="$store.fir.count || {{.count}}">{{.count}}</div>
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
		<script defer src="https://unpkg.com/@adnaanx/fir@latest/dist/cdn.min.js"></script>
	</head>
	
	<body>
		{{template "content" .}}
	</body>
	
	</html>`
}

func (c *CounterView) OnRequest(_ http.ResponseWriter, _ *http.Request) (fir.Status, fir.Data) {
	return fir.Status{Code: 200}, fir.Data{
		"count": c.Value(),
	}
}

func (c *CounterView) OnEvent(s fir.Socket) error {
	switch s.Event().ID {
	case "inc":
		s.Store().UpdateProp("count", c.Inc())
	case "dec":
		s.Store().UpdateProp("count", c.Dec())
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", s.Event())
	}
	return nil
}

func main() {
	controller := fir.NewController("counter_app", fir.DevelopmentMode(true))
	http.Handle("/", controller.Handler(&CounterView{}))
	http.ListenAndServe(":9867", nil)
}
```
{% endraw %}

</details>

## Optional: Live Ticker

We changed the state(`count`) on events(`inc`, `dec`) coming from the browser. It would be quite useful to change the state from events emanating from the server itself. To show this behaviour lets add a live ticker which shows how long ago the count was updated.

Override the `EventReceiver` method of the `View` interface to return a receive only channel(`<- chan Event`). The channel is used  by `Fir` to dispatch the event to `OnEvent` where it can be handled.

```go
type CounterView struct {
	fir.DefaultView
	count int32
	ch    chan fir.Event
}

func (c *CounterView) EventReceiver() <-chan fir.Event {
	return c.ch
}

...

http.Handle("/", controller.Handler(&CounterView{ch: make(chan fir.Event)}))
```

Since there are a bunch of changes, please see the full code:

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
	"sync"
	"time"

	"github.com/adnaan/fir"
)

func NewCounterView() *CounterView {
	timerCh := make(chan fir.Event)
	ticker := time.NewTicker(time.Second)
	go func() {
		for ; true; <-ticker.C {
			timerCh <- fir.Event{ID: "tick"}
		}
	}()
	return &CounterView{ch: timerCh}
}

type CounterView struct {
	fir.DefaultView
	count   int32
	updated time.Time
	ch      chan fir.Event
	sync.RWMutex
}

func (c *CounterView) EventReceiver() <-chan fir.Event {
	return c.ch
}

func (c *CounterView) Inc() int32 {
	c.Lock()
	defer c.Unlock()
	c.count += 1
	c.updated = time.Now()
	return c.count
}

func (c *CounterView) Dec() int32 {
	c.Lock()
	defer c.Unlock()
	c.count -= 1
	c.updated = time.Now()
	return c.count
}

func (c *CounterView) Count() int32 {
	c.RLock()
	defer c.RUnlock()
	return c.count
}

func (c *CounterView) Updated() time.Time {
	c.RLock()
	defer c.RUnlock()
	return c.updated
}

func (c *CounterView) Content() string {
	return `{{define "content" }} 
		<div class="my-6" style="height: 500px">
					<div class="columns is-mobile is-centered is-vcentered">
						<div x-data class="column is-one-third-desktop has-text-centered is-narrow">
							<div>
								<div>Count updated: <span x-text="$store.fir.count_updated || 0"></span> seconds ago</div>
								<hr>
								<div id="count" x-text="$store.fir.count || {{.count}}">{{.count}}</div>
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
		<script defer src="https://unpkg.com/@adnaanx/fir@latest/dist/cdn.min.js"></script>
	</head>
	
	<body>
		{{template "content" .}}
	</body>
	
	</html>`
}

func (c *CounterView) OnRequest(_ http.ResponseWriter, _ *http.Request) (fir.Status, fir.Data) {
	return fir.Status{Code: 200}, fir.Data{
		"count": c.Count(),
	}
}

func (c *CounterView) OnEvent(s fir.Socket) error {
	switch s.Event().ID {
	case "tick":
		updated := c.Updated()
		if updated.IsZero() {
			return nil
		}
		s.Store().UpdateProp("count_updated", time.Since(updated).Seconds())
	case "inc":
		s.Store().UpdateProp("count", c.Inc())
	case "dec":
		s.Store().UpdateProp("count", c.Dec())
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", s.Event())
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

Final code with optional steps is also here [examples/counter-ticker](https://github.com/adnaan/fir/blob/main/examples/counter-ticker/main.go).