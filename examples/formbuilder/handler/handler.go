package handler

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/livefir/fir"
)

// NewRoute creates and returns the fir.RouteOptions for the formbuilder.
// contentFile should be the path to the "app.html" template, resolved correctly
// by the fir.Content option (usually relative to the application's working directory
// or a configured template root).
func NewRoute(contentFile string) fir.RouteOptions {
	// Seed random number generator for generating unique keys for new inputs.
	// It's good practice to seed once, ideally in main, but for this example structure,
	// seeding here ensures it's done if this route is used.
	// For more complex apps, consider a global init.
	rand.Seed(time.Now().UnixNano())

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
