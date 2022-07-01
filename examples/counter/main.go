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
	return morphCount(atomic.AddInt32(&c.count, 1))
}

func (c *Counter) Dec() fir.Patch {
	return morphCount(atomic.AddInt32(&c.count, -1))
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

func (c *CounterView) OnRequest(_ http.ResponseWriter, _ *http.Request) (fir.Status, fir.Data) {
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
