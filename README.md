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

func (c *Counter) OnEvent(s pwc.Socket) error {
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

