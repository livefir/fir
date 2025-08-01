package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	actionstest "github.com/livefir/fir/examples/actions-test"
	"github.com/livefir/fir/examples/autocomplete"
	"github.com/livefir/fir/examples/chirper"
	"github.com/livefir/fir/examples/counter"
	counterticker "github.com/livefir/fir/examples/counter-ticker"
	countertickerredis "github.com/livefir/fir/examples/counter-ticker-redis"
	defaultroute "github.com/livefir/fir/examples/default_route"
	"github.com/livefir/fir/examples/fira"
	"github.com/livefir/fir/examples/formbuilder"
	orycounter "github.com/livefir/fir/examples/ory-counter"
	rangecounter "github.com/livefir/fir/examples/range"
	"github.com/livefir/fir/examples/routing"
	"github.com/livefir/fir/examples/search"
	"github.com/livefir/fir/examples/todo"
)

func main() {
	examples := map[int]struct {
		name string
		run  func(int) error
	}{
		1:  {"Counter", counter.Run},
		2:  {"Counter with Ticker", counterticker.Run},
		3:  {"Form Builder", formbuilder.Run},
		4:  {"Counter with Ticker (Redis)", countertickerredis.Run},
		5:  {"Autocomplete", autocomplete.Run},
		6:  {"Chirper (Social Media)", chirper.Run},
		7:  {"Default Route", defaultroute.Run},
		8:  {"Fira (Project Manager)", fira.Run},
		9:  {"Ory Counter (Auth)", orycounter.Run},
		10: {"Range Counter", rangecounter.Run},
		11: {"Routing", routing.Run},
		12: {"Search", search.Run},
		13: {"Todo", todo.Run},
		14: {"Actions Test (WebSocket Enabled)", actionstest.Run},
		15: {"Actions Test (HTTP-Only Mode)", actionstest.RunHTTPOnly},
	}

	fmt.Println("Available Fir Examples:")
	fmt.Println("======================")
	for i := 1; i <= len(examples); i++ {
		example := examples[i]
		fmt.Printf("%d. %s\n", i, example.name)
	}
	fmt.Printf("\nSelect an example to run (1-%d): ", len(examples))

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
