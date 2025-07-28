package counter // Changed from main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/livefir/fir/internal/dev"

	"github.com/livefir/fir"
)

func Index() fir.RouteOptions {
	var count int32
	return fir.RouteOptions{
		fir.ID("counter"),
		fir.Content("count.html"), // Now uses relative path resolution
		fir.OnLoad(func(ctx fir.RouteContext) error {
			return ctx.KV("count", atomic.LoadInt32(&count))
		}),
		fir.OnEvent("inc", func(ctx fir.RouteContext) error {
			return ctx.KV("count", atomic.AddInt32(&count, 1))
		}),
		fir.OnEvent("dec", func(ctx fir.RouteContext) error {
			return ctx.KV("count", atomic.AddInt32(&count, -1))
		}),
	}
}

// Run starts the counter example server.
func Run(httpPort int) error { // Changed from main
	dev.SetupAlpinePluginServer()

	controller := fir.NewController("counter_app", fir.DevelopmentMode(true))
	http.Handle("/", controller.RouteFunc(Index))
	log.Printf("Starting counter server on port %d\n", httpPort)
	return http.ListenAndServe(fmt.Sprintf(":%v", httpPort), nil) // Use parameter and return error
}
