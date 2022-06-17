package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/adnaan/fir"
)

func NewCounterView() *CounterView {
	stream := make(chan fir.Patch)
	ticker := time.NewTicker(time.Second)
	c := &CounterView{stream: stream}
	go func() {
		for ; true; <-ticker.C {
			updated := c.Updated()
			if updated.IsZero() {
				continue
			}
			stream <- fir.Store{
				Name: "fir",
				Data: map[string]any{"count_updated": time.Since(updated).Seconds()},
			}
		}
	}()
	return c
}

type CounterView struct {
	fir.DefaultView
	count   int32
	updated time.Time
	stream  chan fir.Patch
	sync.RWMutex
}

func (c *CounterView) Stream() <-chan fir.Patch {
	return c.stream
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
		<script defer src="http://localhost:8000/cdn.js"></script>
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

func (c *CounterView) OnPatchEvent(event fir.Event) (fir.Patchset, error) {
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

func main() {
	controller := fir.NewController("counter_app", fir.DevelopmentMode(true))
	http.Handle("/", controller.Handler(NewCounterView()))
	http.ListenAndServe(":9867", nil)
}
