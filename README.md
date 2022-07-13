# Fir

A Go library to build reactive apps .

## A live counter app.

See the complete code in [examples/counter-ticker](./examples/counter-ticker)

`app.html`

```html
<head>
    ...
    <script defer src="https://unpkg.com/@adnaanx/fir@latest/dist/fir.min.js"></script>
	<script defer src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"></script>
</head>
<div>
  <div>Count updated: <span x-text="$store.fir.count_updated || 0"></span> seconds ago</div>
  {{block "count" .}}<div id="count">{{.count}}</div>{{end}}
  <button @click="$fir.emit('inc')">+</button>
  <button @click="$fir.emit('dec')">-</button>
</div>

...
```

`main.go`

```go
...

type Counter struct {
	count   int32
	updated time.Time
	sync.RWMutex
}

func morphCount(c int32) fir.Patch {
	return fir.Morph{
		Selector: "#count",
		Template: &fir.Template{
			Name: "count",
			Data: fir.Data{"count": c},
		},
	}
}

func (c *Counter) Inc() fir.Patch {
	c.Lock()
	defer c.Unlock()
	c.count += 1
	c.updated = time.Now()
	return morphCount(c.count)
}

func (c *Counter) Dec() fir.Patch {
	c.Lock()
	defer c.Unlock()
	c.count -= 1
	c.updated = time.Now()
	return morphCount(c.count)
}

func (c *Counter) Updated() (fir.Patch, error) {
	c.RLock()
	defer c.RUnlock()
	if c.updated.IsZero() {
		return nil, fmt.Errorf("time is zero")
	}
	return fir.Store{
		Name: "fir",
		Data: fir.Data{
			"count_updated": time.Since(c.updated).Seconds(),
		},
	}, nil
}
...
```



See a more real world example in [examples/starter](./examples/starter/) which is also deployed here: [https://fir-starter.fly.dev/](https://fir-starter.fly.dev/)

**Why does it exist ?**

Wants to provide a way to build moderately complex reactive apps for folks who are comfortable with Go.

The library is a result of a series of experiments to build reactive apps in Go: [gomodest-template](https://github.com/adnaan/gomodest-template). It works by `patching` the DOM on user events using [morphdom](https://github.com/patrick-steele-idem/morphdom).

**What is it ?**
- A Go library
- Focuses only on the view layer.
- Ships with an Alpinejs plugin for user interactions(click, submit, navigate ) etc.

**Who is it for ?**
- Is also the why
- Suitable for Go developers who want to build moderately complex apps, internal tools, prototypes etc.
- Skills needed: Go, HTML, CSS, Alpine.js.

**What can you do with it ?**
- Update parts of the web page on user interaction without reloading the page over regular http: clicks, form submits etc.
- Stream page updates over a persistent connection(WS, SSE): notifications, live tickers, chat messages etc.

**Is it like hotwire or is it like phoenix liveview ?**

It borrows the idea of patching DOM on user interaction events from [phoenix live view](https://hex.pm/packages/phoenix_live_view). But instead of streaming DOM diffs over websocket and sticthing it back on the client, it takes the [hotwire](https://hotwired.dev/) approach of re-rendering html templates on the server and sending back a patch DOM operation to the javascript client. 

Live patching of the DOM(over websockets, sse) is also available but only for server driven DOM patching.(notifications, live ticker etc.)



## Principles

- **Library** and not a framework. It’s a Go **library** to build reactive user interfaces.
- **Nothing crazy tech**: It is built on nothing crazy tech: Go, html/template and Alpinejs. It’s just plain old html templates sprinkled with a bit of javascript.
- Keep Go code free of html/css: Use `html/template` to hydrate html pages.
- Keep Javascript to the minimum: Alpinejs provides declarative constructs to wire up moderately complex logic. The `fir` JS client provides additional Alpinejs functions and directives to achieve this goal.
- Have a simple lifecycle:
  - Stages: Render html page -> Handle UI change events → Update parts of the html.
- Be SEO friendly: First page render is done fully on the server side. Real-time interaction is done once the page has been rendered.
- Have a low learning curve: For a Go user the only new thing to learn would be Alpinejs. And yes: HTML & CSS
- No custom template engine: Writing our own template engine can enable in-memory html diffing and minimal change partial for the client, but it also means maintaining a new non standard template engine.


## Status

Work in progress. The current focus is to get to a developer experience which is acceptable to the community. Roadmap to v1.0.0 is still uncertain.