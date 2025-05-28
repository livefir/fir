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
	rangecounter "github.com/livefir/fir/examples/range"
)

func TestRangeExampleE2E(t *testing.T) {
	controller := fir.NewController("range_example_e2e_"+strings.ReplaceAll(t.Name(), "/", "_"), fir.DevelopmentMode(true))
	mux := http.NewServeMux()
	mux.Handle("/", controller.RouteFunc(rangecounter.NewRoute))
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

	// Test that the range input exists
	var inputExists bool
	if err := chromedp.Run(ctx,
		chromedp.Evaluate(`document.querySelector('input[name="count"]') !== null || document.querySelector('input[type="range"]') !== null || document.querySelector('input[type="number"]') !== null`, &inputExists),
	); err != nil {
		t.Fatal(err)
	}

	if !inputExists {
		t.Fatal("Range counter input not found")
	}

	// Test that there's a total display element
	var totalExists bool
	if err := chromedp.Run(ctx,
		chromedp.Evaluate(`document.body.innerHTML.includes('Price:') || document.body.innerHTML.includes('total') || document.body.innerHTML.includes('Total')`, &totalExists),
	); err != nil {
		t.Fatal(err)
	}

	if !totalExists {
		t.Fatal("Range counter total display not found")
	}
}
