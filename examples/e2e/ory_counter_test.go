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
	orycounter "github.com/livefir/fir/examples/ory-counter"
)

func TestOryCounterExampleE2E(t *testing.T) {
	controller := fir.NewController("ory_counter_example_e2e_"+strings.ReplaceAll(t.Name(), "/", "_"), fir.DevelopmentMode(true))
	mux := http.NewServeMux()

// Add static file server for Alpine.js plugin to solve Docker networking issues
if err := SetupStaticFileServer(mux); err != nil {
t.Fatalf("Failed to setup static file server: %v", err)
}
	mux.Handle("/", controller.RouteFunc(orycounter.NewRoute))
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

	// Test that the counter interface exists (buttons and count display)
	var hasCounterInterface bool
	if err := chromedp.Run(ctx,
		chromedp.Evaluate(`document.querySelector('button') !== null && (document.querySelector('form') !== null || document.body.innerHTML.includes('Count:'))`, &hasCounterInterface),
	); err != nil {
		t.Fatal(err)
	}

	if !hasCounterInterface {
		t.Fatal("Ory Counter interface (buttons and count display) not found")
	}
}
