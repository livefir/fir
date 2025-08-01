package actionstest

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/livefir/fir"
	"github.com/livefir/fir/internal/dev"
)

// Item represents a simple item with a text value
type Item struct {
	Text string
}

func Route() fir.RouteOptions {
	var counter int32
	var appendItems = []Item{
		{Text: "Append Item 1"},
		{Text: "Append Item 2"},
		{Text: "Append Item 3"},
	}
	var prependItems = []Item{
		{Text: "Prepend Item 1"},
		{Text: "Prepend Item 2"},
		{Text: "Prepend Item 3"},
	}
	var message = "Welcome to Actions Test"
	var isDisabled bool
	var highlightActive bool
	var formValue string

	return fir.RouteOptions{
		fir.ID("actions-test"),
		fir.Content("actions.html"),
		fir.OnLoad(func(ctx fir.RouteContext) error {
			return ctx.Data(map[string]interface{}{
				"counter":         atomic.LoadInt32(&counter),
				"appendItems":     appendItems,
				"prependItems":    prependItems,
				"message":         message,
				"isDisabled":      isDisabled,
				"highlightActive": highlightActive,
				"formValue":       formValue,
			})
		}),
		fir.OnEvent("increment", func(ctx fir.RouteContext) error {
			return ctx.KV("counter", atomic.AddInt32(&counter, 1))
		}),
		fir.OnEvent("decrement", func(ctx fir.RouteContext) error {
			return ctx.KV("counter", atomic.AddInt32(&counter, -1))
		}),
		fir.OnEvent("reset", func(ctx fir.RouteContext) error {
			atomic.StoreInt32(&counter, 0)
			formValue = ""
			if err := ctx.KV("counter", atomic.LoadInt32(&counter)); err != nil {
				return err
			}
			return ctx.KV("formValue", formValue)
		}),
		fir.OnEvent("add-item", func(ctx fir.RouteContext) error {
			newItem := Item{Text: fmt.Sprintf("New Append Item %d", len(appendItems)+1)}
			appendItems = append(appendItems, newItem)
			return ctx.Data(newItem)
		}),
		fir.OnEvent("prepend-item", func(ctx fir.RouteContext) error {
			newItem := Item{Text: fmt.Sprintf("New Prepend Item %d", len(prependItems)+1)}
			prependItems = append([]Item{newItem}, prependItems...)
			return ctx.Data(newItem)
		}),
		fir.OnEvent("remove-item", func(ctx fir.RouteContext) error {
			if len(appendItems) > 0 {
				appendItems = appendItems[1:]
			}
			return ctx.KV("appendItems", appendItems)
		}),
		fir.OnEvent("remove-parent", func(ctx fir.RouteContext) error {
			return ctx.KV("removeParent", true)
		}),
		fir.OnEvent("toggle-disabled", func(ctx fir.RouteContext) error {
			// Toggle the disabled state of the input field
			isDisabled = !isDisabled
			message = fmt.Sprintf("Input disabled state toggled to: %t", isDisabled)
			if err := ctx.KV("isDisabled", isDisabled); err != nil {
				return err
			}
			return ctx.KV("message", message)
		}),
		fir.OnEvent("toggle-highlight", func(ctx fir.RouteContext) error {
			highlightActive = !highlightActive
			return ctx.KV("highlightActive", highlightActive)
		}),
		fir.OnEvent("dispatch-event", func(ctx fir.RouteContext) error {
			message = "Dispatch event triggered!"
			return ctx.KV("message", message)
		}),
		fir.OnEvent("js-action", func(ctx fir.RouteContext) error {
			message = "JavaScript action executed!"
			return ctx.KV("message", message)
		}),
		fir.OnEvent("form-submit", func(ctx fir.RouteContext) error {
			formValue = ctx.Request().FormValue("test_input")
			message = fmt.Sprintf("Form submitted with value: %s", formValue)
			if err := ctx.KV("formValue", formValue); err != nil {
				return err
			}
			return ctx.KV("message", message)
		}),
		fir.OnEvent("redirect-action", func(ctx fir.RouteContext) error {
			return ctx.Redirect("/", 302)
		}),
	}
}

// Index returns the route for the actions test page
func Index() fir.RouteOptions {
	return Route()
}

// Run starts the actions test server with WebSocket enabled
func Run(httpPort int) error {
	dev.SetupAlpinePluginServer()

	controller := fir.NewController("actions_test", fir.DevelopmentMode(true))
	http.Handle("/", controller.RouteFunc(Index))
	log.Printf("Starting actions test server with WebSocket ENABLED on port %d\n", httpPort)
	return http.ListenAndServe(fmt.Sprintf(":%v", httpPort), nil)
}

// RunHTTPOnly starts the actions test server with WebSocket disabled (HTTP-only mode)
func RunHTTPOnly(httpPort int) error {
	dev.SetupAlpinePluginServer()

	controller := fir.NewController("actions_test",
		fir.DevelopmentMode(true),
		fir.WithDisableWebsocket(), // Disable WebSocket
	)
	http.Handle("/", controller.RouteFunc(Index))
	log.Printf("Starting actions test server with WebSocket DISABLED (HTTP-only) on port %d\n", httpPort)
	return http.ListenAndServe(fmt.Sprintf(":%v", httpPort), nil)
}
