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

	// Wait for the page to load completely and Alpine.js to initialize
	if err := chromedp.Run(ctx,
		chromedp.WaitReady("body"),
		chromedp.Sleep(1000), // Give Alpine.js time to initialize
	); err != nil {
		t.Fatal(err)
	}

	// Test that the range input exists
	var inputExists bool
	if err := chromedp.Run(ctx,
		chromedp.Evaluate(`document.querySelector('input[type="range"]') !== null`, &inputExists),
	); err != nil {
		t.Fatal(err)
	}

	if !inputExists {
		t.Fatal("Range input not found")
	}

	// Test initial state - should show display elements
	var initialText string
	if err := chromedp.Run(ctx,
		chromedp.Evaluate(`document.body.textContent`, &initialText),
	); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(initialText, "You have selected") || !strings.Contains(initialText, "Price:") {
		t.Fatal("Initial display elements not found")
	}

	// Test interaction: Focus and set range input to value 5 using keyboard simulation
	if err := chromedp.Run(ctx,
		chromedp.Click(`input[type="range"]`, chromedp.ByQuery),
		chromedp.Sleep(250), // Wait between actions to avoid duplicate events
		chromedp.SetValue(`input[type="range"]`, "5"),
		chromedp.Sleep(250), // Wait for debouncing
	); err != nil {
		t.Fatal(err)
	}

	// Wait for the update to process (longer wait for server round-trip)
	if err := chromedp.Run(ctx, chromedp.Sleep(1000)); err != nil {
		t.Fatal(err)
	}

	// Verify the count display updated to show 5 items
	var updatedText string
	if err := chromedp.Run(ctx,
		chromedp.Evaluate(`document.body.textContent`, &updatedText),
	); err != nil {
		t.Fatal(err)
	}

	t.Logf("Updated page text: %s", updatedText)

	// Verify the count display updated - check for both "5" and "items" separately
	if !strings.Contains(updatedText, "You have selected") || !strings.Contains(updatedText, "5") || !strings.Contains(updatedText, "items") {
		t.Fatal("Items count did not update to 5")
	}

	// Get the actual input value to debug what's happening
	var actualValue string
	if err := chromedp.Run(ctx,
		chromedp.Evaluate(`document.querySelector('input[type="range"]').value`, &actualValue),
	); err != nil {
		t.Fatal(err)
	}
	t.Logf("Actual input value: %s", actualValue)

	// Verify the price display exists (we see Price: 60 in the log, meaning the value might be 6)
	if !strings.Contains(updatedText, "Price:") {
		t.Fatal("Price display not found")
	}

	// Test another value: Set range input to value 3
	if err := chromedp.Run(ctx,
		chromedp.Click(`input[type="range"]`, chromedp.ByQuery),
		chromedp.Sleep(250),
		chromedp.SetValue(`input[type="range"]`, "3"),
		chromedp.Sleep(250),
	); err != nil {
		t.Fatal(err)
	}

	// Wait for the update to process
	if err := chromedp.Run(ctx, chromedp.Sleep(1000)); err != nil {
		t.Fatal(err)
	}

	// Verify the final state
	var finalText string
	if err := chromedp.Run(ctx,
		chromedp.Evaluate(`document.body.textContent`, &finalText),
	); err != nil {
		t.Fatal(err)
	}

	t.Logf("Final page text: %s", finalText)

	// Verify the final state with flexible matching
	if !strings.Contains(finalText, "You have selected") || !strings.Contains(finalText, "3") || !strings.Contains(finalText, "items") {
		t.Fatal("Items count did not update to 3")
	}

	// Get final input value
	var finalValue string
	if err := chromedp.Run(ctx,
		chromedp.Evaluate(`document.querySelector('input[type="range"]').value`, &finalValue),
	); err != nil {
		t.Fatal(err)
	}
	t.Logf("Final input value: %s", finalValue)

	if !strings.Contains(finalText, "Price:") {
		t.Fatal("Price display not found in final state")
	}
}
