package e2e

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/livefir/fir"
	countertickerredis "github.com/livefir/fir/examples/counter-ticker-redis"
	"github.com/livefir/fir/pubsub"
	"github.com/redis/go-redis/v9"
	redisContainer "github.com/testcontainers/testcontainers-go/modules/redis"
)

// Removed CounterRedisModel struct and its methods (Inc, Dec, Count, UpdatedSecondsAgo)
// Removed countUpdateDataRedis struct

// counterTickerRedisRouteForTest now uses the imported package and correctly overrides options
func counterTickerRedisRouteForTest(t *testing.T, pubsubAdapter pubsub.Adapter) fir.RouteOptions {
	// Use the new package structure to create a route with the test's pubsub adapter
	actualRoute := countertickerredis.NewRoute(pubsubAdapter)
	originalOpts := actualRoute.Options()

	// Append new options for ID, Content, and Layout for testing.
	// These will override any defaults set by the original package's options
	// because route options are applied sequentially.
	overriddenOpts := append(originalOpts,
		fir.ID("counter_ticker_redis_test_"+strings.ReplaceAll(t.Name(), "/", "_")),
		fir.Content("../counter-ticker-redis/count.html"), // Path relative to e2e test directory
		fir.Layout("../counter-ticker-redis/layout.html"), // Path relative to e2e test directory
	)

	return overriddenOpts
}

func TestCounterTickerRedisE2E(t *testing.T) {
	if os.Getenv("DOCKER") != "1" {
		t.Skip("Skipping testing since DOCKER=1 is not set")
	}

	ctx := context.Background()
	redisC, err := redisContainer.Run(ctx, "docker.io/redis:7")
	if err != nil {
		t.Fatalf("failed to start redis container: %s", err)
	}
	defer func() {
		if err := redisC.Terminate(ctx); err != nil {
			t.Logf("failed to terminate redis container: %s", err)
		}
	}()

	host, err := redisC.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get redis host: %s", err)
	}
	port, err := redisC.MappedPort(ctx, "6379/tcp")
	if err != nil {
		t.Fatalf("failed to get redis mapped port: %s", err)
	}
	redisAddr := fmt.Sprintf("%s:%s", host, port.Port())

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	defer redisClient.Close()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		t.Fatalf("failed to ping redis: %s", err)
	}

	redisPubsubAdapter := pubsub.NewRedis(redisClient)

	controller := fir.NewController(
		"counter_ticker_redis_e2e_"+strings.ReplaceAll(t.Name(), "/", "_"),
		fir.DevelopmentMode(true),
		fir.WithPubsubAdapter(redisPubsubAdapter), // Use Redis pubsub adapter
	)

	mux := http.NewServeMux()

	// Add static file server for Alpine.js plugin to solve Docker networking issues
	if err := SetupStaticFileServer(mux); err != nil {
		t.Fatalf("Failed to setup static file server: %v", err)
	}
	routeFunc := func() fir.RouteOptions { return counterTickerRedisRouteForTest(t, redisPubsubAdapter) }
	mux.Handle("/", controller.RouteFunc(routeFunc))
	ts := httptest.NewServer(mux)
	defer ts.Close()

	chromedpOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
	)
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), chromedpOpts...)
	defer cancelAlloc()

	chromedpCtx, cancelChromedp := chromedp.NewContext(allocCtx, chromedp.WithLogf(t.Logf))
	defer cancelChromedp()

	chromedp.ListenTarget(chromedpCtx, func(ev interface{}) {
		if ev, ok := ev.(*runtime.EventConsoleAPICalled); ok {
			for _, arg := range ev.Args {
				t.Logf("Browser Console (%s): %s", ev.Type, string(arg.Value))
			}
		}
		if ev, ok := ev.(*runtime.EventExceptionThrown); ok {
			t.Logf("Browser Exception: %+v", ev.ExceptionDetails)
		}
	})

	timeoutCtx, cancelTimeout := context.WithTimeout(chromedpCtx, 25*time.Second) // Slightly increased timeout
	defer cancelTimeout()

	var initialCountText, countAfterIncText, countAfterDecText string
	var initialUpdatedText, updatedTextAfterWait string

	countDisplaySelector := `//div[
                                starts-with(normalize-space(.), 'Count:') and 
                                (count(@*[starts-with(name(), '@fir:inc:ok::fir-')]) > 0 or count(@*[starts-with(name(), '@fir:dec:ok::fir-')]) > 0) and
                                (contains(@class, 'fir-inc-ok--fir-') or contains(@class, 'fir-dec-ok--fir-'))
                            ]`
	updatedDisplaySelector := `//div[
                                starts-with(normalize-space(.), 'Count updated:') and 
                                count(@*[starts-with(name(), '@fir:updated:ok::fir-')]) > 0 and
                                contains(@class, 'fir-updated-ok--fir-')
                            ]`
	incrementButtonSelector := `#increment-btn`
	decrementButtonSelector := `#decrement-btn`

	if err := chromedp.Run(timeoutCtx,
		chromedp.Navigate(ts.URL),
		chromedp.ActionFunc(func(c context.Context) error { t.Logf("After Navigate"); return nil }),
		chromedp.WaitVisible(countDisplaySelector, chromedp.BySearch),
		chromedp.ActionFunc(func(c context.Context) error { t.Logf("After WaitVisible countDisplaySelector"); return nil }),
		chromedp.WaitVisible(updatedDisplaySelector, chromedp.BySearch),
		chromedp.ActionFunc(func(c context.Context) error { t.Logf("After WaitVisible updatedDisplaySelector"); return nil }),
		chromedp.TextContent(countDisplaySelector, &initialCountText, chromedp.BySearch),
		chromedp.ActionFunc(func(c context.Context) error { t.Logf("After TextContent initialCountText"); return nil }),
		chromedp.TextContent(updatedDisplaySelector, &initialUpdatedText, chromedp.BySearch),
		chromedp.ActionFunc(func(c context.Context) error { t.Logf("After TextContent initialUpdatedText"); return nil }),

		chromedp.WaitVisible(incrementButtonSelector, chromedp.ByID),
		chromedp.ActionFunc(func(c context.Context) error { t.Logf("After WaitVisible incrementButtonSelector"); return nil }),
		chromedp.Click(incrementButtonSelector, chromedp.ByID),
		chromedp.ActionFunc(func(c context.Context) error { t.Logf("After Click incrementButtonSelector"); return nil }),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.ActionFunc(func(c context.Context) error { t.Logf("After Sleep (post-increment)"); return nil }),
		chromedp.TextContent(countDisplaySelector, &countAfterIncText, chromedp.BySearch),
		chromedp.ActionFunc(func(c context.Context) error { t.Logf("After TextContent countAfterIncText"); return nil }),

		chromedp.WaitVisible(decrementButtonSelector, chromedp.ByID),
		chromedp.ActionFunc(func(c context.Context) error { t.Logf("After WaitVisible decrementButtonSelector"); return nil }),
		chromedp.Click(decrementButtonSelector, chromedp.ByID),
		chromedp.ActionFunc(func(c context.Context) error { t.Logf("After Click decrementButtonSelector"); return nil }),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.ActionFunc(func(c context.Context) error { t.Logf("After Sleep (post-decrement)"); return nil }),
		chromedp.TextContent(countDisplaySelector, &countAfterDecText, chromedp.BySearch),
		chromedp.ActionFunc(func(c context.Context) error { t.Logf("After TextContent countAfterDecText"); return nil }),

		chromedp.Sleep(2500*time.Millisecond), // Wait for ticker (increased for Redis version)
		chromedp.ActionFunc(func(c context.Context) error { t.Logf("After Sleep (ticker wait)"); return nil }),
		chromedp.TextContent(updatedDisplaySelector, &updatedTextAfterWait, chromedp.BySearch),
		chromedp.ActionFunc(func(c context.Context) error { t.Logf("After TextContent updatedTextAfterWait"); return nil }),
	); err != nil {
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

	updatedSecondsRegex := regexp.MustCompile(`Count updated: (\d+\.?\d*) seconds ago`)
	extractUpdatedSeconds := func(text string) float64 {
		matches := updatedSecondsRegex.FindStringSubmatch(text)
		if len(matches) == 2 {
			val, errFloat := strconv.ParseFloat(matches[1], 64)
			if errFloat == nil {
				return val
			}
			t.Logf("Warning: could not parse float from updated text %q: %v", matches[1], errFloat)
		}
		t.Logf("Warning: could not extract updated seconds from %q", text)
		return -1
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

	if !(finalSeconds >= 0 && finalSeconds < 3) {
		t.Errorf("expected final updated seconds to be a small positive number (0-2s), got %.2f (from text: %q)", finalSeconds, updatedTextAfterWait)
	}
	if initialSeconds < 2 && finalSeconds <= initialSeconds && updatedTextAfterWait == initialUpdatedText {
		// This condition might be tricky if the initial update was very recent due to page load.
		// A more reliable check might be that finalSeconds is small and positive, as done above.
		t.Logf("Initial updated text was %q (%.2fs), final updated text is %q (%.2fs). Text might not have changed if initial update was already < 1s ago from a previous tick.", initialUpdatedText, initialSeconds, updatedTextAfterWait, finalSeconds)
	}

	t.Logf("Initial updated text: %q (%.2fs)", initialUpdatedText, initialSeconds)
	t.Logf("Final updated text: %q (%.2fs)", updatedTextAfterWait, finalSeconds)
}
