package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/adnaan/fir"
	"github.com/adnaan/fir/pubsub"
	"github.com/golang/glog"
)

type Counter struct {
	count   int32
	updated time.Time
	sync.RWMutex
}

func (c *Counter) Inc(ctx fir.RouteContext) error {
	c.Lock()
	defer c.Unlock()
	c.count += 1
	c.updated = time.Now()
	return ctx.KV("count", c.count)
}

func (c *Counter) Dec(ctx fir.RouteContext) error {
	c.Lock()
	defer c.Unlock()
	c.count -= 1
	c.updated = time.Now()
	return ctx.KV("count", c.count)
}

func (c *Counter) Updated() float64 {
	c.RLock()
	defer c.RUnlock()
	return time.Since(c.updated).Seconds()
}

func (c *Counter) Count() int32 {
	c.RLock()
	defer c.RUnlock()
	return c.count
}

func NewCounterIndex(pubsub pubsub.Adapter) *index {
	c := &index{
		model:       &Counter{},
		pubsub:      pubsub,
		eventSender: make(chan fir.Event),
		id:          "counter",
	}

	ticker := time.NewTicker(time.Second)
	pattern := fmt.Sprintf("*:%s", c.id)

	go func() {
		for ; true; <-ticker.C {
			if !c.pubsub.HasSubscribers(context.Background(), pattern) {
				// if userID:viewID(*:viewID) channel pattern has no subscribers, skip costly operation
				glog.Errorf("channel pattern %s has no subscribers", pattern)
				continue
			}
			c.eventSender <- fir.NewEvent("updated", countUpdate{CountUpdated: c.model.Updated()})
		}
	}()
	return c
}

type countUpdate struct {
	CountUpdated float64 `json:"count_updated"`
}

type index struct {
	model       *Counter
	pubsub      pubsub.Adapter
	eventSender chan fir.Event
	id          string
}

func (i *index) Options() fir.RouteOptions {
	return fir.RouteOptions{
		fir.ID(i.id),
		fir.Content(content),
		fir.Layout(layout),
		fir.OnLoad(i.load),
		fir.OnEvent("inc", i.inc),
		fir.OnEvent("dec", i.dec),
		fir.OnEvent("updated", i.updated),
		fir.EventSender(i.eventSender),
	}
}

func (i *index) load(ctx fir.RouteContext) error {
	return ctx.KV("count", i.model.Count())
}

func (i *index) inc(ctx fir.RouteContext) error {
	return i.model.Inc(ctx)
}

func (i *index) dec(ctx fir.RouteContext) error {
	return i.model.Dec(ctx)
}
func (i *index) updated(ctx fir.RouteContext) error {
	req := &countUpdate{}
	err := ctx.Bind(req)
	if err != nil {
		return err
	}
	return ctx.DOM().Store("fir", req)
}

var content = `
{{define "content" }} 
<div class="my-6" style="height: 500px">
	<div class="columns is-mobile is-centered is-vcentered">
		<div x-data="{patch:{selector:'#count',block:'count'}}" 
			 class="column is-one-third-desktop has-text-centered is-narrow">
			<div>
				<div>Count updated: <span x-text="$store.fir.count_updated || 0"></span> seconds ago</div>
				<hr>
				{{block "count" .}}<div id="count">{{.count}}</div>{{end}}
				<button class="button has-background-primary" @click="$fir.emit('inc',$data)">+
				</button>
				<button class="button has-background-primary" @click="$fir.emit('dec',$data)">-
				</button>
			</div>
		</div>
	</div>
</div>
{{end}}`

var layout = `<!DOCTYPE html>
	<html lang="en">
	
	<head>
		<title>{{.app_name}}</title>
		<meta charset="UTF-8">
		<meta name="description" content="A counter app">
		<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bulma@0.9.4/css/bulma.min.css" />
		<script defer src="http://localhost:8000/cdn.js"></script>
		<!-- <script defer src="https://unpkg.com/@adnaanx/fir@latest/dist/fir.min.js"></script> -->
		<script defer src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"></script>
	</head>
	
	<body>
		{{template "content" .}}
	</body>
	
	</html>`

func main() {
	pubsubAdapter := pubsub.NewInmem()
	controller := fir.NewController("counter_app", fir.DevelopmentMode(true), fir.WithPubsubAdapter(pubsubAdapter))
	http.Handle("/", controller.Route(NewCounterIndex(pubsubAdapter)))
	http.ListenAndServe(":9867", nil)
}
