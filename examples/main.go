package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/livefir/fir/examples/counter"
	counterticker "github.com/livefir/fir/examples/counter-ticker"
	countertickerredis "github.com/livefir/fir/examples/counter-ticker-redis"
	"github.com/livefir/fir/examples/formbuilder"
)

func main() {
	examples := map[int]struct {
		name string
		run  func(int) error
	}{
		1: {"Counter", counter.Run},
		2: {"Counter with Ticker", counterticker.Run},
		3: {"Form Builder", formbuilder.Run},
		4: {"Counter with Ticker (Redis)", countertickerredis.Run},
	}

	fmt.Println("Available Fir Examples:")
	fmt.Println("======================")
	for i, example := range examples {
		fmt.Printf("%d. %s\n", i, example.name)
	}
	fmt.Print("\nSelect an example to run (1-4): ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Error reading input: %v", err)
	}

	choice, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil {
		log.Fatalf("Invalid input: please enter a number")
	}

	example, exists := examples[choice]
	if !exists {
		log.Fatalf("Invalid choice: please select a number between 1 and %d", len(examples))
	}

	port := 9867 // Default port
	fmt.Printf("\nStarting %s example on port %d\n", example.name, port)
	fmt.Printf("Open http://localhost:%d in your browser\n\n", port)

	err = example.run(port)
	if err != nil {
		log.Fatalf("Error running %s example: %v", example.name, err)
	}
}
