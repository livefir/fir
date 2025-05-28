package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/livefir/fir"
	"github.com/livefir/fir/examples/formbuilder"
)

// listenForConsoleMessages sets up a listener for JavaScript console messages and exceptions.
// It returns a cancel function to stop listening.
func listenForConsoleMessages(ctx context.Context, t *testing.T) context.CancelFunc {
	logCtx, cancel := context.WithCancel(ctx)
	chromedp.ListenTarget(logCtx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *runtime.EventConsoleAPICalled:
			var parts []string
			for _, arg := range ev.Args {
				// arg.Value is a json.RawMessage. Convert to string for logging.
				// For simple values, this will be like "string" or 123.
				// For objects, it will be the JSON representation.
				parts = append(parts, string(arg.Value))
			}
			t.Logf("JS CONSOLE [%s]: %s", ev.Type, strings.Join(parts, " "))
		case *runtime.EventExceptionThrown:
			t.Logf("JS EXCEPTION: %s", ev.ExceptionDetails.Text)
			if ev.ExceptionDetails.Exception != nil {
				t.Logf("  Description: %s", ev.ExceptionDetails.Exception.Description)
				// Exception object details can be fetched using runtime.GetProperties if needed
			}
			if ev.ExceptionDetails.StackTrace != nil {
				for _, frame := range ev.ExceptionDetails.StackTrace.CallFrames {
					t.Logf("    at %s (%s:%d:%d)", frame.FunctionName, frame.URL, frame.LineNumber, frame.ColumnNumber)
				}
			}
		}
	})
	return cancel
}

func TestFormBuilderExampleE2E(t *testing.T) {
	controller := fir.NewController("formbuilder_e2e_"+strings.ReplaceAll(t.Name(), "/", "_"), fir.DevelopmentMode(true))
	mux := http.NewServeMux()
	// Use the NewRoute function from the imported formbuilder package
	// Wrap the call in a function literal to match RouteFunc signature
	mux.Handle("/", controller.RouteFunc(func() fir.RouteOptions {
		return formbuilder.NewRoute("../formbuilder/app.html")
	}))
	ts := httptest.NewServer(mux)
	defer ts.Close()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		//chromedp.Flag("headless", false), // Set to true for headless mode
	)
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Listen for console logs and exceptions from the browser
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *runtime.EventConsoleAPICalled:
			for _, arg := range ev.Args {
				var val interface{}
				if err := json.Unmarshal(arg.Value, &val); err == nil {
					t.Logf("chrome console.%s: %v", ev.Type, val)
				} else {
					t.Logf("chrome console.%s: %s (raw: %s)", ev.Type, string(arg.Value), err)
				}
			}
		case *runtime.EventExceptionThrown:
			t.Logf("chrome exception: %s", ev.ExceptionDetails.Text)
			if ev.ExceptionDetails.Exception != nil && ev.ExceptionDetails.Exception.Description != "" {
				t.Logf("chrome exception details: %s", ev.ExceptionDetails.Exception.Description)
			}
		}
	})

	timeoutCtx, cancelTimeout := context.WithTimeout(ctx, 30*time.Second)
	defer cancelTimeout()

	addButtonSelector := `//button[@formaction="/?event=add"]`
	// This XPath selector targets the div where new inputs are appended,
	// using its class name, which is derived from the fir event and template.
	inputContainerXPath := `//div[@class="fir-add-ok--input-text"]` // Corrected XPath string to use class
	// This XPath selector targets the input items themselves.
	inputItemsXPath := fmt.Sprintf(`%s/div[@fir-key]`, inputContainerXPath)

	var initialInputCount, countAfterAdd, countAfterRemove int

	err := chromedp.Run(timeoutCtx,
		chromedp.ActionFunc(func(ctx context.Context) error { t.Logf("Navigating to %s", ts.URL); return nil }),
		chromedp.Navigate(ts.URL),
		chromedp.ActionFunc(func(ctx context.Context) error { t.Logf("Navigation complete."); return nil }),

		// Log the body's outerHTML to inspect the DOM
		chromedp.ActionFunc(func(ctx context.Context) error {
			t.Logf("Attempting to get body outerHTML...")
			var bodyHTML string
			err := chromedp.OuterHTML("body", &bodyHTML, chromedp.ByQuery).Do(ctx)
			if err != nil {
				t.Logf("Error getting body outerHTML: %v", err)
				// Don't fail the test here, allow other steps to proceed if possible, or let them timeout.
				// But this log is crucial.
			}
			t.Logf("Body outerHTML:\n%s", bodyHTML)
			return nil
		}),

		// 1. Wait for the input container to be present and log its attributes for verification
		chromedp.ActionFunc(func(ctx context.Context) error {
			t.Logf("Waiting for input container to be present: %s", inputContainerXPath)
			return nil
		}),
		chromedp.WaitReady(inputContainerXPath, chromedp.BySearch), // Changed from WaitVisible to WaitReady
		chromedp.ActionFunc(func(ctx context.Context) error {
			t.Logf("Input container is ready. Verifying uniqueness and attributes...")
			var nodes []*cdp.Node
			if err := chromedp.Nodes(inputContainerXPath, &nodes, chromedp.BySearch).Do(ctx); err != nil {
				t.Logf("Error finding input container node: %v", err)
				return fmt.Errorf("error finding input container node: %w", err)
			}
			if len(nodes) == 0 {
				t.Logf("Input container node NOT FOUND with XPath: %s", inputContainerXPath)
				return fmt.Errorf("input container node not found with XPath: %s", inputContainerXPath)
			}
			if len(nodes) > 1 {
				t.Logf("Multiple input container nodes (%d) found with XPath: %s", len(nodes), inputContainerXPath)
				return fmt.Errorf("multiple input container nodes found with XPath: %s", inputContainerXPath)
			}
			t.Logf("Input container node found and is unique.")
			// Optionally log attributes if needed for debugging
			// var attrs map[string]string
			// if err := chromedp.Attributes(inputContainerXPath, &attrs, chromedp.BySearch).Do(ctx); err == nil {
			// 	t.Logf("Input container attributes: %+v", attrs)
			// }
			return nil
		}),

		// 2. Check initial state: no inputs. Focus on existence first.
		chromedp.ActionFunc(func(ctx context.Context) error {
			t.Logf("Checking initial input items existence with XPath: %s", inputItemsXPath)
			var nodes []*cdp.Node
			// We expect this to potentially find no nodes, which is not an error for the initial state.
			err := chromedp.Nodes(inputItemsXPath, &nodes, chromedp.BySearch, chromedp.AtLeast(0)).Do(ctx)
			if err != nil {
				// This error is unexpected if AtLeast(0) is used, unless it's a malformed XPath or context error.
				t.Logf("Initial check: Error querying for input items (even with AtLeast(0)): %v", err)
				return err
			}
			initialInputCount = len(nodes)
			t.Logf("Initial check: Found %d input items.", initialInputCount)
			return nil
		}),
		chromedp.ActionFunc(func(ctx context.Context) error { t.Logf("Initial input count: %d", initialInputCount); return nil }),

		// Click "Add input" button
		chromedp.ActionFunc(func(ctx context.Context) error { t.Logf("Waiting for add button: %s", addButtonSelector); return nil }),
		chromedp.WaitVisible(addButtonSelector, chromedp.BySearch),
		chromedp.ActionFunc(func(ctx context.Context) error { t.Logf("Clicking add button."); return nil }),
		chromedp.Click(addButtonSelector, chromedp.BySearch),
		chromedp.ActionFunc(func(ctx context.Context) error { t.Logf("Clicked add button"); return nil }),

		// Wait for the new input to appear and verify count
		chromedp.ActionFunc(func(ctx context.Context) error {
			t.Logf("Waiting for new input item to appear: %s", inputItemsXPath)
			return nil
		}),
		chromedp.WaitVisible(inputItemsXPath, chromedp.BySearch),
		chromedp.ActionFunc(func(ctx context.Context) error { t.Logf("New input item visible. Counting inputs..."); return nil }),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var nodes []*cdp.Node
			err := chromedp.Nodes(inputItemsXPath, &nodes, chromedp.BySearch).Do(ctx)
			if err != nil {
				t.Logf("Count after add: Error querying nodes: %v", err)
				return err
			}
			countAfterAdd = len(nodes)
			t.Logf("Count after add: Found %d items.", countAfterAdd)
			return nil
		}),
		chromedp.ActionFunc(func(ctx context.Context) error { t.Logf("Input count after add: %d", countAfterAdd); return nil }),

		// Get the key of the added input to target its remove button
		chromedp.ActionFunc(func(ctx context.Context) error {
			t.Logf("Getting fir-key of added input...")
			var addedInputKey string
			err := chromedp.AttributeValue(inputItemsXPath, "fir-key", &addedInputKey, nil, chromedp.BySearch).Do(ctx)
			if err != nil {
				t.Logf("Failed to get fir-key: %v", err)
				return fmt.Errorf("failed to get fir-key of added input: %w", err)
			}
			if addedInputKey == "" {
				t.Logf("Added input fir-key is empty.")
				return fmt.Errorf("added input fir-key is empty")
			}
			t.Logf("Added input key: %s", addedInputKey)

			// Log the HTML of the input container for debugging
			var containerHTML string
			if err := chromedp.OuterHTML(inputContainerXPath, &containerHTML, chromedp.BySearch).Do(ctx); err != nil {
				t.Logf("Error getting input container HTML: %v", err)
			} else {
				t.Logf("Input container HTML after add:\n%s", containerHTML)
			}

			removeButtonSelector := fmt.Sprintf(`//div[@fir-key='%s']/form[@fir-key='%s']/button[@formaction='/?event=remove']`, addedInputKey, addedInputKey)
			t.Logf("Waiting for remove button: %s", removeButtonSelector)
			if err := chromedp.WaitReady(removeButtonSelector, chromedp.BySearch).Do(ctx); err != nil { // Changed WaitVisible to WaitReady
				t.Logf("Failed to find remove button with WaitReady: %v", err)
				return err
			}
			t.Logf("Clicking remove button for key %s", addedInputKey)
			if err := chromedp.Click(removeButtonSelector, chromedp.BySearch).Do(ctx); err != nil {
				t.Logf("Failed to click remove button: %v", err)
				return err
			}
			t.Logf("Clicked remove button for key %s", addedInputKey)

			// Start listening for console messages right after the click
			cancelConsoleListener := listenForConsoleMessages(ctx, t)
			defer cancelConsoleListener()

			return nil
		}),

		// Wait for the input to be removed (disappear) and verify count
		chromedp.ActionFunc(func(ctx context.Context) error {
			t.Logf("Waiting for input item(s) to be removed (using WaitNotPresent on %s, timeout 10s)...", inputItemsXPath)

			// Create a context with a specific timeout for WaitNotPresent.
			// This timeout is separate from the overall test timeout (timeoutCtx).
			waitCtx, waitCancel := context.WithTimeout(ctx, 10*time.Second)
			defer waitCancel()

			// Attempt to wait for the element(s) to disappear.
			// WaitNotPresent is an action, so it needs .Do(waitCtx) when used this way.
			errWait := chromedp.WaitNotPresent(inputItemsXPath, chromedp.BySearch).Do(waitCtx)

			if errWait != nil {
				// WaitNotPresent timed out or encountered another error.
				// This means the element(s) matching inputItemsXPath are likely still present.
				t.Logf("WaitNotPresent for %s failed after 10s: %v. Querying current count of items.", inputItemsXPath, errWait)

				var nodes []*cdp.Node
				// Query the current number of nodes. Use the original `ctx` (from the ActionFunc parameter),
				// as `waitCtx` might be done or cancelled.
				// AtLeast(0) ensures it doesn't error if no nodes are found (though we expect some if WaitNotPresent failed).
				if queryErr := chromedp.Nodes(inputItemsXPath, &nodes, chromedp.BySearch, chromedp.AtLeast(0)).Do(ctx); queryErr != nil {
					t.Logf("Error querying nodes for %s after WaitNotPresent failure: %v", inputItemsXPath, queryErr)
					// Both WaitNotPresent and the fallback Nodes query failed. This is a more complex error state.
					// Return an error that encompasses both failures.
					return fmt.Errorf("WaitNotPresent for %s failed (%v) and subsequent node query also failed (%v)", inputItemsXPath, errWait, queryErr)
				}

				countAfterRemove = len(nodes)
				t.Logf("After WaitNotPresent failure, %s matched %d item(s).", inputItemsXPath, countAfterRemove)

				// Return an error to indicate that the removal was not confirmed within the timeout.
				// This will cause the parent chromedp.Run to fail.
				return fmt.Errorf("element(s) matching %s were still present after 10s (count: %d), WaitNotPresent error: %w", inputItemsXPath, countAfterRemove, errWait)
			}

			// WaitNotPresent succeeded. This means no elements currently match inputItemsXPath.
			t.Logf("Input item(s) at %s confirmed removed by WaitNotPresent.", inputItemsXPath)
			countAfterRemove = 0
			return nil
		}),
		chromedp.ActionFunc(func(ctx context.Context) error { t.Logf("Input count after remove: %d", countAfterRemove); return nil }),
	)

	if err != nil {
		t.Fatalf("Chromedp execution failed: %v", err)
	}

	if initialInputCount != 0 {
		t.Errorf("expected initial input count to be 0, got %d", initialInputCount)
	}
	if countAfterAdd != 1 {
		t.Errorf("expected input count after add to be 1, got %d", countAfterAdd)
	}
	if countAfterRemove != 0 {
		t.Errorf("expected input count after remove to be 0, got %d", countAfterRemove)
	}
}
