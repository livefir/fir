package e2e

import (
	"context"
	// Removed sync import as Counter struct with its RWMutex is removed
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/livefir/fir"
	counterticker "github.com/livefir/fir/examples/counter-ticker"
	"github.com/livefir/fir/pubsub"
)

// Removed Counter struct and its methods (Inc, Dec, Count, Updated)
// Removed countUpdateData struct

// counterTickerRouteForTest now uses the imported package and correctly overrides options
func counterTickerRouteForTest(t *testing.T, pubsubAdapter pubsub.Adapter) fir.RouteOptions {
	// Use the new package structure to create a route with the test's pubsub adapter
	actualHandlerRoute := counterticker.NewRoute(pubsubAdapter)
	originalOpts := actualHandlerRoute.Options()

	// Append new options for ID, Content, and Layout.
	// fir.RouteOptions is a slice of functions (fir.RouteOption).
	// Direct field assignment (e.g., opts.ID = ...) is incorrect.
	// Instead, we append new functional options. These will override any defaults
	// set by the original handler's options because route options are applied sequentially,
	// and later ones for the same underlying route properties take precedence.
	overriddenOpts := append(originalOpts,
		fir.ID("counter_ticker_test_"+strings.ReplaceAll(t.Name(), "/", "_")),
		fir.Content("../counter-ticker/count.html"), // Path relative to e2e test directory
		fir.Layout("../counter-ticker/layout.html"), // Path relative to e2e test directory
	)

	// Note: The EventSender, OnLoad, OnEvent handlers are from the imported package's originalOpts.
	// The ticker goroutine started by counterticker.NewRoute is part of the imported package's logic.

	return overriddenOpts
}

func TestCounterTickerExampleE2E(t *testing.T) {
	pubsubAdapter := pubsub.NewInmem()
	controller := fir.NewController(
		"counter_ticker_e2e_"+strings.ReplaceAll(t.Name(), "/", "_"),
		fir.DevelopmentMode(true),
		fir.WithPubsubAdapter(pubsubAdapter),
	)

	mux := http.NewServeMux()
	// Update the call to pass pubsubAdapter
	routeFunc := func() fir.RouteOptions { return counterTickerRouteForTest(t, pubsubAdapter) }
	mux.Handle("/", controller.RouteFunc(routeFunc))
	ts := httptest.NewServer(mux)
	defer ts.Close()

	// ...existing code...
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		//chromedp.Flag("headless", false), // Set to true for headless mode
	)
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()

	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(t.Logf))
	defer cancel()

	// Listen for browser console logs and exceptions
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if ev, ok := ev.(*runtime.EventConsoleAPICalled); ok {
			for _, arg := range ev.Args {
				t.Logf("Browser Console (%s): %s", ev.Type, string(arg.Value))
			}
		}
		// You can also listen for exceptions
		if ev, ok := ev.(*runtime.EventExceptionThrown); ok {
			t.Logf("Browser Exception: %+v", ev.ExceptionDetails)
		}
	})

	timeoutCtx, cancelTimeout := context.WithTimeout(ctx, 20*time.Second)
	defer cancelTimeout()

	var initialCountText, countAfterIncText, countAfterDecText string
	var initialUpdatedText, updatedTextAfterWait string

	// Selectors based on counter-ticker/count.html
	// Dynamic XPath for the count display div
	countDisplaySelector := `//div[
                                starts-with(normalize-space(.), 'Count:') and 
                                (count(@*[starts-with(name(), '@fir:inc:ok::fir-')]) > 0 or count(@*[starts-with(name(), '@fir:dec:ok::fir-')]) > 0) and
                                (contains(@class, 'fir-inc-ok--fir-') or contains(@class, 'fir-dec-ok--fir-'))
                            ]`
	// Dynamic XPath for the "updated" div
	updatedDisplaySelector := `//div[
                                starts-with(normalize-space(.), 'Count updated:') and 
                                count(@*[starts-with(name(), '@fir:updated:ok::fir-')]) > 0 and
                                contains(@class, 'fir-updated-ok--fir-')
                            ]`
	// Buttons use $fir.emit - now using IDs
	incrementButtonSelector := `#increment-btn` // Changed to ID selector
	decrementButtonSelector := `#decrement-btn` // Changed to ID selector

	err := chromedp.Run(timeoutCtx,
		chromedp.Navigate(ts.URL),
		chromedp.ActionFunc(func(ctx context.Context) error { t.Logf("After Navigate"); return nil }),
		chromedp.WaitVisible(countDisplaySelector, chromedp.BySearch),
		chromedp.ActionFunc(func(ctx context.Context) error { t.Logf("After WaitVisible countDisplaySelector"); return nil }),
		chromedp.WaitVisible(updatedDisplaySelector, chromedp.BySearch),
		chromedp.ActionFunc(func(ctx context.Context) error { t.Logf("After WaitVisible updatedDisplaySelector"); return nil }),
		chromedp.TextContent(countDisplaySelector, &initialCountText, chromedp.BySearch), // Keep BySearch for countDisplaySelector
		chromedp.ActionFunc(func(ctx context.Context) error { t.Logf("After TextContent initialCountText"); return nil }),
		chromedp.TextContent(updatedDisplaySelector, &initialUpdatedText, chromedp.BySearch),
		chromedp.ActionFunc(func(ctx context.Context) error { t.Logf("After TextContent initialUpdatedText"); return nil }),

		// Wait for and click buttons using ByID
		chromedp.WaitVisible(incrementButtonSelector, chromedp.ByID), // Changed to ByID
		chromedp.ActionFunc(func(ctx context.Context) error { t.Logf("After WaitVisible incrementButtonSelector"); return nil }),
		chromedp.Click(incrementButtonSelector, chromedp.ByID), // Changed to ByID
		chromedp.ActionFunc(func(ctx context.Context) error { t.Logf("After Click incrementButtonSelector"); return nil }),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.ActionFunc(func(ctx context.Context) error { t.Logf("After Sleep (post-increment)"); return nil }),
		chromedp.TextContent(countDisplaySelector, &countAfterIncText, chromedp.BySearch), // Keep BySearch for countDisplaySelector
		chromedp.ActionFunc(func(ctx context.Context) error { t.Logf("After TextContent countAfterIncText"); return nil }),

		chromedp.WaitVisible(decrementButtonSelector, chromedp.ByID), // Changed to ByID
		chromedp.ActionFunc(func(ctx context.Context) error { t.Logf("After WaitVisible decrementButtonSelector"); return nil }),
		chromedp.Click(decrementButtonSelector, chromedp.ByID), // Changed to ByID
		chromedp.ActionFunc(func(ctx context.Context) error { t.Logf("After Click decrementButtonSelector"); return nil }),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.ActionFunc(func(ctx context.Context) error { t.Logf("After Sleep (post-decrement)"); return nil }),
		chromedp.TextContent(countDisplaySelector, &countAfterDecText, chromedp.BySearch), // Keep BySearch for countDisplaySelector
		chromedp.ActionFunc(func(ctx context.Context) error { t.Logf("After TextContent countAfterDecText"); return nil }),

		// Wait for the ticker to fire and update the "updated" message
		// The handler's ticker is 2 seconds. Wait a bit longer.
		chromedp.Sleep(2500*time.Millisecond),
		chromedp.ActionFunc(func(ctx context.Context) error { t.Logf("After Sleep (ticker wait)"); return nil }),
		chromedp.TextContent(updatedDisplaySelector, &updatedTextAfterWait, chromedp.BySearch),
		chromedp.ActionFunc(func(ctx context.Context) error { t.Logf("After TextContent updatedTextAfterWait"); return nil }),
	)

	if err != nil {
		t.Fatalf("Chromedp execution failed: %v", err)
	}

	extractCountVal := func(text string) string {
		parts := strings.Split(text, ":")
		if len(parts) == 2 {
			return strings.TrimSpace(parts[1])
		}
		t.Logf("Warning: could not extract count from %q", text)
		return "N/A"
	}

	// Regex to extract the number of seconds from "Count updated: X.XXX seconds ago"
	updatedSecondsRegex := regexp.MustCompile(`Count updated: (\d+\.?\d*) seconds ago`)
	extractUpdatedSeconds := func(text string) float64 {
		matches := updatedSecondsRegex.FindStringSubmatch(text)
		if len(matches) == 2 {
			val, err := strconv.ParseFloat(matches[1], 64)
			if err == nil {
				return val
			}
			t.Logf("Warning: could not parse float from updated text %q: %v", matches[1], err)
		}
		t.Logf("Warning: could not extract updated seconds from %q", text)
		return -1 // Indicate failure
	}

	if initial := extractCountVal(initialCountText); initial != "0" {
		t.Errorf("expected initial count to be 0, got %s (full text: %q)", initial, initialCountText)
	}
	if afterInc := extractCountVal(countAfterIncText); afterInc != "1" {
		t.Errorf("expected count after increment to be 1, got %s (full text: %q)", afterInc, countAfterIncText)
	}
	if afterDec := extractCountVal(countAfterDecText); afterDec != "0" {
		t.Errorf("expected count after decrement to be 0, got %s (full text: %q)", afterDec, countAfterDecText)
	}

	initialSeconds := extractUpdatedSeconds(initialUpdatedText)
	finalSeconds := extractUpdatedSeconds(updatedTextAfterWait)

	if initialSeconds < 0 {
		t.Errorf("could not parse initial updated seconds: %q", initialUpdatedText)
	}
	if finalSeconds < 0 {
		t.Errorf("could not parse final updated seconds: %q", updatedTextAfterWait)
	}

	// Check that the "updated" time has progressed.
	// The handler's ticker is 2s. We waited 2.5s.
	// The value from model.Updated() (which the handler uses) should be small after an event.
	if !(finalSeconds >= 0 && finalSeconds < 3) {
		t.Errorf("expected final updated seconds to be a small positive number (0-2.x s), got %.2f (from text: %q)", finalSeconds, updatedTextAfterWait)
	}

	if initialSeconds < 3 && finalSeconds <= initialSeconds && updatedTextAfterWait == initialUpdatedText {
		// This condition might be tricky if the initial update was very recent due to page load.
		// A more reliable check might be that finalSeconds is small and positive, as done above.
		t.Logf("Initial updated text was %q (%.2fs), final updated text is %q (%.2fs). Text might not have changed if initial update was already < 2s ago from a previous tick.", initialUpdatedText, initialSeconds, updatedTextAfterWait, finalSeconds)
	}

	t.Logf("Initial updated text: %q (%.2fs)", initialUpdatedText, initialSeconds)
	t.Logf("Final updated text: %q (%.2fs)", updatedTextAfterWait, finalSeconds)
}
