package counterticker

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/livefir/fir"
	"github.com/livefir/fir/internal/dev"
	"github.com/livefir/fir/pubsub"
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
		fir.OnEvent(fir.EventSocketConnected, func(ctx fir.RouteContext) error {
			var status fir.SocketStatus
			err := ctx.Bind(&status)
			if err != nil {
				return err
			}
			fmt.Printf("onevent: socket connected for user %s\n", status.User)
			return nil
		}),
		fir.OnEvent(fir.EventSocketDisconnected, func(ctx fir.RouteContext) error {
			var status fir.SocketStatus
			err := ctx.Bind(&status)
			if err != nil {
				return err
			}
			fmt.Printf("onevent: socket disconnected for user %s\n", status.User)
			return nil
		}),
		fir.EventSender(i.eventSender),
	}
}

func (i *index) load(ctx fir.RouteContext) error {
	return ctx.Data(map[string]any{"count": i.model.Count(), "updated": i.model.Updated()})
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

func newRoute(pubsubAdapter pubsub.Adapter) fir.Route {
	c := &index{
		model:       &Counter{updated: time.Now()},
		pubsub:      pubsubAdapter,
		eventSender: make(chan fir.Event),
		id:          "counter",
	}

	ticker := time.NewTicker(time.Second * 2)
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

// Index returns the route options for the counter-ticker example.
func Index() fir.RouteOptions {
	pubsubAdapter := pubsub.NewInmem()
	route := newRoute(pubsubAdapter)
	return route.Options()
}

// NewRoute creates a new route with the given pubsub adapter.
// This is mainly used for testing purposes.
func NewRoute(pubsubAdapter pubsub.Adapter) fir.Route {
	return newRoute(pubsubAdapter)
}

// Run starts the counter-ticker example server.

func Run(httpPort int) error {
	dev.SetupAlpinePluginServer()
	pubsubAdapter := pubsub.NewInmem()
	controller := fir.NewController("counter_app",
		fir.DevelopmentMode(true),
		fir.WithPubsubAdapter(pubsubAdapter),
		fir.WithOnSocketConnect(func(userOrSessionID string) error {
			fmt.Printf("socket connected for user %s\n", userOrSessionID)
			return nil
		}),
		fir.WithOnSocketDisconnect(func(userOrSessionID string) {
			fmt.Printf("socket disconnected for user %s\n", userOrSessionID)
		}),
	)
	http.Handle("/", controller.Route(newRoute(pubsubAdapter)))
	log.Printf("Starting counter-ticker server on port %d\n", httpPort)
	return http.ListenAndServe(fmt.Sprintf(":%v", httpPort), nil)
}
