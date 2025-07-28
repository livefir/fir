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
	defaultroute "github.com/livefir/fir/examples/default_route"
)

func TestDefaultRouteExampleE2E(t *testing.T) {
	controller := fir.NewController("default_route_example_e2e_"+strings.ReplaceAll(t.Name(), "/", "_"), fir.DevelopmentMode(true))
	mux := http.NewServeMux()

// Add static file server for Alpine.js plugin to solve Docker networking issues
if err := SetupStaticFileServer(mux); err != nil {
t.Fatalf("Failed to setup static file server: %v", err)
}
	mux.Handle("/", controller.RouteFunc(defaultroute.NewRoute))
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

	// Test that the page loads (has HTML structure)
	var hasHtml bool
	if err := chromedp.Run(ctx,
		chromedp.Evaluate(`document.documentElement !== null`, &hasHtml),
	); err != nil {
		t.Fatal(err)
	}

	if !hasHtml {
		t.Fatal("Default route page did not load properly")
	}
}
