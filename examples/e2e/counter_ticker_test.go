package e2e

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/livefir/fir"
	"github.com/livefir/fir/pubsub"
)

// CounterTickerModel is a simplified version of the Counter model in counter-ticker/main.go
type CounterTickerModel struct {
	count   int32
	updated time.Time
	sync.RWMutex
}

func (c *CounterTickerModel) Inc() int32 {
	c.Lock()
	defer c.Unlock()
	c.count++
	c.updated = time.Now()
	return c.count
}

func (c *CounterTickerModel) Dec() int32 {
	c.Lock()
	defer c.Unlock()
	c.count--
	c.updated = time.Now()
	return c.count
}

func (c *CounterTickerModel) Count() int32 {
	c.RLock()
	defer c.RUnlock()
	return c.count
}

func (c *CounterTickerModel) UpdatedSecondsAgo() float64 {
	c.RLock()
	defer c.RUnlock()
	if c.updated.IsZero() { // Handle case where it hasn't been updated yet
		return 0
	}
	return time.Since(c.updated).Seconds()
}

type countUpdateData struct {
	CountUpdated float64
}

func counterTickerRouteForTest(t *testing.T) fir.RouteOptions {
	model := &CounterTickerModel{updated: time.Now()} // Initialize updated time
	eventSender := make(chan fir.Event)
	// Use a unique ID for the route to avoid conflicts if other tests use the same ID
	routeID := "counter_ticker_test_" + strings.ReplaceAll(t.Name(), "/", "_")

	// Start a ticker similar to the one in counter-ticker/main.go
	// For testing, we might want a slightly faster ticker or a way to control it.
	// Let's use a 1-second ticker for the test.
	ticker := time.NewTicker(1 * time.Second)
	// Create a context that can be cancelled to stop the ticker goroutine
	ctx, cancel := context.WithCancel(context.Background())

	// It's important to stop the ticker and cancel the context when the test/server is done.
	// Since httptest.Server doesn't have a direct hook for this, we'll rely on test completion.
	// For a real app, you'd manage this lifecycle more carefully.
	// In this test setup, the goroutine will run as long as the test server is up.
	// We'll call cancel() when the test function ends.
	t.Cleanup(func() {
		ticker.Stop()
		cancel()
		close(eventSender)
	})

	go func() {
		for {
			select {
			case <-ticker.C:
				// In a real app, you'd check for subscribers like in the example.
				// For this test, we'll always send.
				eventSender <- fir.NewEvent("updated", countUpdateData{CountUpdated: model.UpdatedSecondsAgo()})
			case <-ctx.Done(): // Stop if the context is cancelled
				return
			}
		}
	}()

	return fir.RouteOptions{
		fir.ID(routeID),
		fir.Content("../counter-ticker/count.html"), // Adjust path as needed
		fir.Layout("../counter-ticker/layout.html"), // Adjust path as needed
		fir.EventSender(eventSender),
		fir.OnLoad(func(rc fir.RouteContext) error {
			return rc.Data(map[string]any{
				"count":   model.Count(),
				"updated": model.UpdatedSecondsAgo(),
			})
		}),
		fir.OnEvent("inc", func(rc fir.RouteContext) error {
			model.Inc()
			return rc.Data(map[string]any{"count": model.Count()})
		}),
		fir.OnEvent("dec", func(rc fir.RouteContext) error {
			model.Dec()
			return rc.Data(map[string]any{"count": model.Count()})
		}),
		fir.OnEvent("updated", func(rc fir.RouteContext) error {
			req := &countUpdateData{}
			if err := rc.Bind(req); err != nil {
				return err
			}
			return rc.Data(map[string]any{"updated": req.CountUpdated})
		}),
	}
}

func TestCounterTickerExampleE2E(t *testing.T) {
	// The counter-ticker example uses an in-memory pubsub adapter.
	pubsubAdapter := pubsub.NewInmem()
	controller := fir.NewController(
		"counter_ticker_e2e_"+strings.ReplaceAll(t.Name(), "/", "_"),
		fir.DevelopmentMode(true),
		fir.WithPubsubAdapter(pubsubAdapter), // Add pubsub adapter as in the example
	)

	mux := http.NewServeMux()
	// Pass `t` to the route function so it can use t.Cleanup
	routeFunc := func() fir.RouteOptions { return counterTickerRouteForTest(t) }
	mux.Handle("/", controller.RouteFunc(routeFunc))
	ts := httptest.NewServer(mux)
	defer ts.Close()

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
		chromedp.Sleep(1500*time.Millisecond), // Ticker is 1s, wait a bit longer
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
	// Since the ticker runs every second, and we waited 1.5s,
	// finalSeconds should be roughly initialSeconds + 1.5 (or more, due to processing).
	// A simpler check is that finalSeconds > initialSeconds, or that it's a small positive number.
	// The value from model.UpdatedSecondsAgo() resets on each event, so it should be small.
	if !(finalSeconds >= 0 && finalSeconds < 3) { // Expecting it to be around 0-2 seconds after the last tick
		t.Errorf("expected final updated seconds to be a small positive number (0-2s), got %.2f (from text: %q)", finalSeconds, updatedTextAfterWait)
	}
	// More robust check: ensure it changed from the initial state if initial was also small
	if initialSeconds < 2 && finalSeconds <= initialSeconds && updatedTextAfterWait == initialUpdatedText {
		t.Errorf("expected updated text to change after waiting, initial: %q, final: %q", initialUpdatedText, updatedTextAfterWait)
	}

	t.Logf("Initial updated text: %q (%.2fs)", initialUpdatedText, initialSeconds)
	t.Logf("Final updated text: %q (%.2fs)", updatedTextAfterWait, finalSeconds)
}
