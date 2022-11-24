package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/adnaan/fir"
	"github.com/go-redis/redis/v8"
)

type Counter struct {
	count   int32
	updated time.Time
	sync.RWMutex
}

func morphCount(c int32) fir.Patch {
	return fir.Morph{
		Selector: "#count",
		HTML: &fir.Render{
			Template: "count",
			Data:     map[string]any{"count": c},
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
		Data: map[string]any{
			"count_updated": time.Since(c.updated).Seconds(),
		},
	}, nil
}

func (c *Counter) Count() int32 {
	c.RLock()
	defer c.RUnlock()
	return c.count
}

func NewCounterView(pubsubAdapter fir.PubsubAdapter) *CounterView {
	publisher := make(chan fir.Patchset)
	ticker := time.NewTicker(time.Second)
	c := &CounterView{publisher: publisher, model: &Counter{}, pubsubAdapter: pubsubAdapter}
	pattern := fmt.Sprintf("*:%s", c.ID())

	go func() {
		for ; true; <-ticker.C {
			if !c.pubsubAdapter.HasSubscribers(context.Background(), pattern) {
				// if userID:viewID(*:viewID) channel pattern has no subscribers, skip costly operation
				log.Printf("channel pattern %s has no subscribers", pattern)
				continue
			}
			patch, err := c.model.Updated()
			if err != nil {
				continue
			}
			publisher <- fir.Patchset{patch}
		}
	}()
	return c
}

type CounterView struct {
	fir.DefaultView
	model         *Counter
	publisher     chan fir.Patchset
	pubsubAdapter fir.PubsubAdapter
	sync.RWMutex
}

func (c *CounterView) ID() string {
	return "counter"
}

func (c *CounterView) Publisher() <-chan fir.Patchset {
	return c.publisher
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
		<!-- <script defer src="http://localhost:8000/cdn.js"></script> -->
		<script defer src="https://unpkg.com/@adnaanx/fir@latest/dist/fir.min.js"></script>
		<script defer src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"></script>
	</head>
	
	<body>
		{{template "content" .}}
	</body>
	
	</html>`
}

func (c *CounterView) OnGet(_ http.ResponseWriter, _ *http.Request) fir.Pagedata {
	return fir.Pagedata{
		Data: map[string]any{
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
	port := flag.String("port", "9867", "port to listen on")

	pubsubAdapter := fir.NewPubsubRedis(
		redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "", // no password set
			DB:       0,  // use default DB
		}),
	)
	controller := fir.NewController("counter_app",
		fir.DevelopmentMode(true),
		fir.WithPubsubAdapter(pubsubAdapter))
	http.Handle("/", controller.Handler(NewCounterView(pubsubAdapter)))
	http.ListenAndServe(fmt.Sprintf(":%s", *port), nil)
}
