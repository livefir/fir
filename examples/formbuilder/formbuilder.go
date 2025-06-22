package formbuilder

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"

	"github.com/livefir/fir"
)

// Index returns the route options for the formbuilder example.
func Index() fir.RouteOptions {
	return newRoute("app.html")
}

// NewRoute creates and returns the fir.RouteOptions for the formbuilder with a custom content file.
// This function is exported for e2e tests that need to specify custom template paths.
func NewRoute(contentFile string) fir.RouteOptions {
	return newRoute(contentFile)
}

// newRoute is the internal implementation that creates the route options.
func newRoute(contentFile string) fir.RouteOptions {
	// Note: As of Go 1.20, global random number generator is automatically seeded
	// and rand.Seed() is deprecated. No need to manually seed anymore.

	return fir.RouteOptions{
		fir.ID("formbuilder"),
		fir.Content(contentFile),
		fir.OnLoad(func(ctx fir.RouteContext) error {
			// Optional: Log onLoad event or perform initial setup
			// fmt.Println("Formbuilder OnLoad event triggered")
			return nil
		}),
		fir.OnEvent("add", func(ctx fir.RouteContext) error {
			// Generate a random key for the new input
			// fmt.Printf("--- Server received add event, generating key ---\n")
			return ctx.KV("key", rand.Intn(1000-1)+1)
		}),
		fir.OnEvent("remove", func(ctx fir.RouteContext) error {
			// The actual removal is handled by x-fir-remove on the client-side.
			// This server-side event is mostly for acknowledgment or server-side cleanup if needed.
			fmt.Println("--- Server received remove event ---")
			return nil
		}),
	}
}

// Run starts the formbuilder example server.
func Run(httpPort int) error {
	controller := fir.NewController("formbuilder", fir.DevelopmentMode(true))
	http.Handle("/", controller.RouteFunc(Index))
	log.Printf("Starting formbuilder server on port %d\n", httpPort)
	return http.ListenAndServe(fmt.Sprintf(":%v", httpPort), nil)
}
