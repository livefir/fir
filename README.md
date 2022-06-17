# Fir

A Go library to build reactive apps.

- Suitable for Go developers who want to build moderately complex apps, internal tools, prototypes etc.
- Skills needed: Go, HTML, CSS, Alpine.js.
- Focuses only on the view layer.

The library is a result of a series of experiments to build reactive apps in Go: [gomodest-template](https://github.com/adnaan/gomodest-template).

## A counter app.

See the complete code in [examples/counter](./examples/counter)

`main.go`

```go
...

func (c *CounterView) OnPatch(event fir.Event) (fir.Patchset, error) {
	switch event.ID {
	case "inc":
		return fir.Patchset{
			fir.Morph{
				Selector: "#count",
				Template: "count",
				Data:     fir.Data{"count": c.Inc()}}}, nil

	case "dec":
		return fir.Patchset{
			fir.Morph{
				Selector: "#count",
				Template: "count",
				Data:     fir.Data{"count": c.Dec()}}}, nil
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", event)
	}

	return nil, nil
}

...
```

`app.html`

```html
<head>
    ...
    <script defer src="https://unpkg.com/@adnaanx/fir@latest/dist/cdn.min.js"></script>
</head>
<div>
  <div id="count" x-text="$store.fir.count || {{.count}}">{{.count}}</div>
  <button @click="$fir.emit('inc')">
    +
  </button>
  <button @click="$fir.emit('dec')">
    -
  </button>
</div>

...
```

See a more real world example in [examples/starter](./examples/starter/) which is also deployed here: [https://fir-starter.fly.dev/](https://fir-starter.fly.dev/)

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

Work in progress.
