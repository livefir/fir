package e2e

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/livefir/fir"
	"github.com/livefir/fir/examples/routing"
)

func TestRoutingExampleE2E(t *testing.T) {
	controller := fir.NewController("routing_example_e2e_"+strings.ReplaceAll(t.Name(), "/", "_"), fir.DevelopmentMode(true))
	mux := http.NewServeMux()
	mux.Handle("/", controller.RouteFunc(routing.NewRoute))
	ts := httptest.NewServer(mux)
	defer ts.Close()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
	)

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()

	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(t.Logf))
	defer cancel()

	// Listen for browser console logs and exceptions
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if ev, ok := ev.(*runtime.EventConsoleAPICalled); ok {
			for _, arg := range ev.Args {
				var valStr string
				if arg.Value != nil {
					valStr = string(arg.Value)
				}
				t.Logf("Browser Console (%s): %s", ev.Type, valStr)
			}
		}
		if ev, ok := ev.(*runtime.EventExceptionThrown); ok {
			t.Logf("Browser Exception: %s", ev.ExceptionDetails.Text)
		}
	})

	// Navigate to the page
	if err := chromedp.Run(ctx, chromedp.Navigate(ts.URL)); err != nil {
		t.Fatal(err)
	}

	// Test that the page has loaded with some content
	var hasContent bool
	if err := chromedp.Run(ctx,
		chromedp.Evaluate(`document.body !== null && document.body.innerHTML.trim().length > 0`, &hasContent),
	); err != nil {
		t.Fatal(err)
	}

	if !hasContent {
		t.Fatal("Routing example page has no content")
	}

	// Test that there's navigation or routing elements (links, nav, or title)
	var hasNavigation bool
	if err := chromedp.Run(ctx,
		chromedp.Evaluate(`document.querySelector('nav') !== null || document.querySelector('a') !== null || document.querySelector('h1') !== null || document.body.innerHTML.includes('Home') || document.body.innerHTML.includes('page')`, &hasNavigation),
	); err != nil {
		t.Fatal(err)
	}

	if !hasNavigation {
		t.Fatal("Routing example navigation elements not found")
	}
}
