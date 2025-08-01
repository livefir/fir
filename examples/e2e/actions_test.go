package e2e

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/livefir/fir"
	actionstest "github.com/livefir/fir/examples/actions-test"
)

func TestActionsE2E(t *testing.T) {
	controller := fir.NewController("actions_e2e_"+strings.ReplaceAll(t.Name(), "/", "_"), fir.DevelopmentMode(true))
	mux := http.NewServeMux()

	// Add static file server for Alpine.js plugin to solve Docker networking issues
	if err := SetupStaticFileServer(mux); err != nil {
		t.Fatalf("Failed to setup static file server: %v", err)
	}

	// Use actions test example
	mux.Handle("/", controller.RouteFunc(func() fir.RouteOptions {
		return actionstest.Index()
	}))
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
			if ev.ExceptionDetails.Exception != nil {
				t.Logf("Browser Exception (details): %s", ev.ExceptionDetails.Exception.Description)
			}
		}
	})

	timeoutCtx, cancelTimeout := context.WithTimeout(ctx, 60*time.Second)
	defer cancelTimeout()

	// Test basic connectivity - more comprehensive tests can be added later
	t.Run("BasicConnectivity", func(t *testing.T) {
		err := chromedp.Run(timeoutCtx,
			chromedp.Navigate(ts.URL),
			chromedp.WaitVisible("#connection-status"),
		)
		if err != nil {
			t.Fatalf("Failed to load actions test page: %v", err)
		}

		// Get connection status without requiring WebSocket
		var connectionStatus string
		err = chromedp.Run(timeoutCtx,
			chromedp.TextContent("#connection-status span", &connectionStatus),
		)
		if err != nil {
			t.Logf("Failed to get connection status, but page loaded successfully: %v", err)
		} else {
			t.Logf("Connection status: %s", connectionStatus)
		}

		t.Log("Actions test page loaded successfully")
	})
}
