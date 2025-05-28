package countertickerredis

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/livefir/fir"
	"github.com/livefir/fir/pubsub"
	"github.com/redis/go-redis/v9"
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
	return ctx.Data(map[string]any{"count": c.count})
}

func (c *Counter) Dec(ctx fir.RouteContext) error {
	c.Lock()
	defer c.Unlock()
	c.count -= 1
	c.updated = time.Now()
	return ctx.Data(map[string]any{"count": c.count})
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

type countUpdate struct {
	CountUpdated float64
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
		fir.Content("count.html"),
		fir.Layout("layout.html"),
		fir.OnLoad(i.load),
		fir.OnEvent("inc", i.inc),
		fir.OnEvent("dec", i.dec),
		fir.OnEvent("updated", i.updated),
		fir.EventSender(i.eventSender),
	}
}

func (i *index) load(ctx fir.RouteContext) error {
	return ctx.Data(map[string]any{"count": i.model.Count()})
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
	return ctx.Data(map[string]any{"updated": req.CountUpdated})
}

func newCounterIndex(pubsubAdapter pubsub.Adapter) *index {
	c := &index{
		model:       &Counter{updated: time.Now()},
		pubsub:      pubsubAdapter,
		eventSender: make(chan fir.Event),
		id:          "counter",
	}

	ticker := time.NewTicker(time.Second)
	pattern := fmt.Sprintf("*:%s", c.id)

	go func() {
		for ; true; <-ticker.C {
			if !c.pubsub.HasSubscribers(context.Background(), pattern) {
				log.Printf("channel pattern %s has no subscribers", pattern)
				continue
			}
			c.eventSender <- fir.NewEvent("updated", countUpdate{CountUpdated: c.model.Updated()})
		}
	}()
	return c
}

// Index returns the route options for the counter-ticker-redis example.
// This uses an in-memory pubsub adapter for simple usage.
func Index() fir.RouteOptions {
	pubsubAdapter := pubsub.NewInmem()
	route := newCounterIndex(pubsubAdapter)
	return route.Options()
}

// NewRoute creates and returns a fir.Route for the counter-ticker-redis example with a custom pubsub adapter.
// This function is exported for e2e tests and scenarios that need Redis pubsub.
func NewRoute(pubsubAdapter pubsub.Adapter) fir.Route {
	return newCounterIndex(pubsubAdapter)
}

// Run starts the counter-ticker-redis example server with Redis pubsub.
func Run(httpPort int) error {
	pubsubAdapter := pubsub.NewRedis(
		redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "", // no password set
			DB:       0,  // use default DB
		}),
	)
	controller := fir.NewController("counter_app",
		fir.DevelopmentMode(true),
		fir.WithPubsubAdapter(pubsubAdapter),
	)
	http.Handle("/", controller.Route(newCounterIndex(pubsubAdapter)))
	log.Printf("Starting counter-ticker-redis server on port %d\n", httpPort)
	log.Printf("Note: This example requires Redis running on localhost:6379\n")
	return http.ListenAndServe(fmt.Sprintf(":%v", httpPort), nil)
}
