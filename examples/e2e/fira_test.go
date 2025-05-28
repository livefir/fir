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
	"github.com/livefir/fir/examples/fira"
)

func TestFiraExampleE2E(t *testing.T) {
	controller := fir.NewController("fira_example_e2e_"+strings.ReplaceAll(t.Name(), "/", "_"), fir.DevelopmentMode(true))
	mux := http.NewServeMux()
	mux.Handle("/", controller.RouteFunc(fira.NewRoute))
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

	// Test that the fira project manager page loads
	var pageLoaded bool
	if err := chromedp.Run(ctx,
		chromedp.Evaluate(`document.body !== null && document.body.innerHTML.length > 0`, &pageLoaded),
	); err != nil {
		t.Fatal(err)
	}

	if !pageLoaded {
		t.Fatal("Fira project manager page did not load properly")
	}

	// Test that there's some form of project interface (could be a table, form, or list)
	var hasProjectInterface bool
	if err := chromedp.Run(ctx,
		chromedp.Evaluate(`document.querySelector('#projects') !== null || document.querySelector('.column') !== null || document.body.innerHTML.includes('projects')`, &hasProjectInterface),
	); err != nil {
		t.Fatal(err)
	}

	if !hasProjectInterface {
		t.Fatal("Fira project interface not found")
	}
}
