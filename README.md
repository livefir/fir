# Fir

A Go library which makes building reactive UIs easy. 

- Suitable for Go developers who want to build moderately complex apps, internal tools, prototypes etc. 
- Skills needed: Go, HTML, CSS, Alpine.js.
- Focuses only on the view layer.

## A counter app.

See full example in [examples/counter](./examples/counter)

`main.go`

```go
...

func (c *Counter) OnEvent(s fir.Socket) error {
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

...
```

`app.html`

```html

<div>
    <div id="count" x-text="$store.fir.count || {{.count}}">{{.count}}</div>
    <button class="button has-background-primary" @click="$fir.emit('inc')">+</button>
    <button class="button has-background-primary" @click="$fir.emit('dec')">-</button>
</div>
 
...
```

## Principles

- **Library** and not a framework. It’s a Go **library** to build reactive user interfaces.
- **Nothing crazy tech**: It is built on nothing crazy tech: Go, html/template and Alpinejs. It’s just plain old html templates sprinkled with a bit of javascript.
- Keep Go code free of html/css: Use `html/template` to hydrate html pages.
- Keep Javascript to the minimum: Alpinejs provides declarative constructs to wire up moderately complex logic. The `fir` JS client provides additional Alpinejs functions and directives to achieve this goal.
- Have a simple lifecycle:
    - Stages: Render html page -> Handle UI change events → Update parts of the html.
- Be SEO friendly: First page render is done fully on the server side. Real-time interaction is done once the page has been rendered.
- Have a low learning curve: For a Go user the only new thing to learn would be Alpinejs. And yes: HTML & CSS
- No custom template engine: Writing our own template engine can enable in-memory html diffing and minimal change partial for the client,  but it also means maintaining a new non standard template engine.

