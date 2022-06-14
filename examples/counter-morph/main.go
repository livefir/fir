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
