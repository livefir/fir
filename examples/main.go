package main

import (
	"log"

	"github.com/livefir/fir/examples/counter" // Import the counter example
)

func main() {
	// Run the counter example
	port := 9867 // Default port
	log.Printf("Attempting to run counter example on port %d", port)
	err := counter.Run(port)
	if err != nil {
		log.Fatalf("Error running counter example: %v", err)
	}
}
