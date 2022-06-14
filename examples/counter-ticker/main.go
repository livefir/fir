package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/adnaan/fir"
)

func NewCounterView() *CounterView {
	timerCh := make(chan fir.Event)
	ticker := time.NewTicker(time.Second)
	go func() {
		for ; true; <-ticker.C {
			timerCh <- fir.Event{ID: "tick"}
		}
	}()
	return &CounterView{ch: timerCh}
}

type CounterView struct {
	fir.DefaultView
	count   int32
	updated time.Time
	ch      chan fir.Event
	sync.RWMutex
}

func (c *CounterView) EventReceiver() <-chan fir.Event {
	return c.ch
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
	return `{{define "content" }} 
<div class="my-6" style="height: 500px">
					<div class="columns is-mobile is-centered is-vcentered">
						<div x-data class="column is-one-third-desktop has-text-centered is-narrow">
							<div>
								<div>Count updated: <span x-text="$store.fir.count_updated || 0"></span> seconds ago</div>
								<hr>
								<div id="count" x-text="$store.fir.count || {{.count}}">{{.count}}</div>
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
		<script defer src="https://unpkg.com/@adnaanx/fir@latest/dist/cdn.min.js"></script>
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

func (c *CounterView) OnEvent(s fir.Socket) error {
	switch s.Event().ID {
	case "tick":
		updated := c.Updated()
		if updated.IsZero() {
			return nil
		}
		s.Store().UpdateProp("count_updated", time.Since(updated).Seconds())
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
	controller := fir.NewController("counter_app", fir.DevelopmentMode(true))
	http.Handle("/", controller.Handler(NewCounterView()))
	http.ListenAndServe(":9867", nil)
}
