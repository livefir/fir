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

func (c *Counter) Count() int32 {
	c.RLock()
	defer c.RUnlock()
	return c.count
}

func NewCounterView() *CounterView {
	stream := make(chan fir.Patch)
	ticker := time.NewTicker(time.Second)
	ticker.Stop()
	c := &CounterView{stream: stream, model: &Counter{}, ticker: ticker}

	go func() {
		for ; true; <-ticker.C {
			log.Println("Tick")
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
	ticker *time.Ticker
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

func (c *CounterView) OnTopicCreated(topic_name string) {
	log.Printf("Topic %s created\n", topic_name)
	c.ticker.Reset(time.Second)
}
func (c *CounterView) OnTopicDestroyed(topic_name string) {
	log.Printf("Topic %s destroyed\n", topic_name)
	c.ticker.Stop()
}

func (c *CounterView) OnGet(_ http.ResponseWriter, _ *http.Request) fir.Page {
	return fir.Page{
		Data: fir.Data{
			"count": c.model.Count(),
		}}
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
