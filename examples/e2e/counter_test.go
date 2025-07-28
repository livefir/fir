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
	"github.com/livefir/fir/examples/counter" // Import the counter package
)

func TestCounterExampleE2E(t *testing.T) {
	controller := fir.NewController("counter_example_e2e_"+strings.ReplaceAll(t.Name(), "/", "_"), fir.DevelopmentMode(true))
	// Create a new http.ServeMux for each test to avoid conflicts if running tests in parallel
	mux := http.NewServeMux()

	// Add static file server for Alpine.js plugin to solve Docker networking issues
	if err := SetupStaticFileServer(mux); err != nil {
		t.Fatalf("Failed to setup static file server: %v", err)
	}

	// Use counter.Index from the imported package
	mux.Handle("/", controller.RouteFunc(counter.Index))
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
				// Try to convert arg.Value to string. It's a json.RawMessage.
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
			if ev.ExceptionDetails.StackTrace != nil {
				for _, frame := range ev.ExceptionDetails.StackTrace.CallFrames {
					t.Logf("    at %s (%s:%d:%d)", frame.FunctionName, frame.URL, frame.LineNumber, frame.ColumnNumber)
				}
			}
		}
	})

	timeoutCtx, cancelTimeout := context.WithTimeout(ctx, 30*time.Second)
	defer cancelTimeout()

	var initialCountText, countAfterIncText, countAfterDecText string

	// Updated selector based on x-fir-refresh="inc,dec" which creates @fir:[inc:ok,dec:ok] attributes
	countDisplaySelector := `//div[starts-with(normalize-space(.), 'Count:') and contains(@class, 'fir-inc-ok--fir-') and contains(@class, 'fir-dec-ok--fir-')]`
	incrementButtonSelector := `//button[@formaction="/?event=inc"]`
	decrementButtonSelector := `//button[@formaction="/?event=dec"]`

	err := chromedp.Run(timeoutCtx,
		chromedp.Navigate(ts.URL),

		// Debug: Log the initial HTML to see the actual structure
		chromedp.ActionFunc(func(ctx context.Context) error {
			var bodyHTML string
			if err := chromedp.OuterHTML("body", &bodyHTML, chromedp.ByQuery).Do(ctx); err != nil {
				t.Logf("Error getting body HTML: %v", err)
			} else {
				t.Logf("Initial body HTML:\n%s", bodyHTML)
			}
			return nil
		}),

		chromedp.WaitVisible(countDisplaySelector, chromedp.BySearch),
		chromedp.TextContent(countDisplaySelector, &initialCountText, chromedp.BySearch),

		chromedp.Click(incrementButtonSelector, chromedp.BySearch),
		chromedp.Sleep(500*time.Millisecond), // Give time for the update
		chromedp.TextContent(countDisplaySelector, &countAfterIncText, chromedp.BySearch),

		chromedp.Click(decrementButtonSelector, chromedp.BySearch),
		chromedp.Sleep(500*time.Millisecond), // Give time for the update
		chromedp.TextContent(countDisplaySelector, &countAfterDecText, chromedp.BySearch),
	)

	if err != nil {
		t.Fatalf("Chromedp execution failed: %v", err)
	}

	// Helper to extract the number from "Count:X"
	extractCount := func(text string) string {
		parts := strings.Split(text, ":")
		if len(parts) == 2 {
			return strings.TrimSpace(parts[1])
		}
		return "N/A" // Or handle error
	}

	if initial := extractCount(initialCountText); initial != "0" {
		t.Errorf("expected initial count to be 0, got %s (full text: %q)", initial, initialCountText)
	}
	if afterInc := extractCount(countAfterIncText); afterInc != "1" {
		t.Errorf("expected count after increment to be 1, got %s (full text: %q)", afterInc, countAfterIncText)
	}
	if afterDec := extractCount(countAfterDecText); afterDec != "0" {
		t.Errorf("expected count after decrement to be 0, got %s (full text: %q)", afterDec, countAfterDecText)
	}
}
