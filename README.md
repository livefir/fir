# Fir

A Go library which makes building reactive UIs easy.

## Counter app example



`main.go`

```go
package main

import (
	"log"
	"net/http"
	"sync/atomic"

	pwc "github.com/adnaan/fir/controller"
)

type Counter struct {
	pwc.DefaultView
	count int32
}

func (c *Counter) Inc() int32 {
	atomic.AddInt32(&c.count, 1)
	return atomic.LoadInt32(&c.count)
}
func (c *Counter) Dec() int32 {
	atomic.AddInt32(&c.count, -1)
	return atomic.LoadInt32(&c.count)
}

func (c *Counter) Value() int32 {
	return atomic.LoadInt32(&c.count)
}

func (c *Counter) Content() string {
	return "app.html"
}

func (c *Counter) OnRequest(_ http.ResponseWriter, _ *http.Request) (pwc.Status, pwc.Data) {
	return pwc.Status{Code: 200}, pwc.Data{
		"count": c.Value(),
	}
}

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

func main() {
	glvc := pwc.Websocket("fir-counter", pwc.DevelopmentMode(true))
	http.Handle("/", glvc.Handler(&Counter{}))
	http.ListenAndServe(":9867", nil)
}
```

`app.html`

```html
<!DOCTYPE html>
<html lang="en">

<head>
    <title>{{.app_name}}</title>
    <meta charset="UTF-8">
    <meta name="description" content="A fir counter app">
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=5.0, minimum-scale=1.0">
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bulma@0.9.4/css/bulma.min.css" />
    <script defer src="https://unpkg.com/@adnaanx/fir@latest/dist/cdn.min.js"></script>
    <script defer src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"></script>
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
    <footer class="footer">
        <div class="content has-text-centered">
            <p>
                {{.app_name}}, 2022
            </p>
        </div>
    </footer>
</body>

</html>

```


