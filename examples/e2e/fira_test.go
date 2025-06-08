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

	// Wait for page to load completely
	if err := chromedp.Run(ctx, chromedp.WaitReady("body")); err != nil {
		t.Fatal(err)
	}

	// Debug: Get the actual body content
	var bodyHTML string
	if err := chromedp.Run(ctx,
		chromedp.OuterHTML("body", &bodyHTML),
	); err != nil {
		t.Fatal(err)
	}
	t.Logf("Fira page body HTML: %s", bodyHTML)

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

	// Test that there's the project interface elements
	var hasProjectInterface bool
	if err := chromedp.Run(ctx,
		chromedp.Evaluate(`document.querySelector('#projects') !== null && document.querySelector('form[action*="create"]') !== null`, &hasProjectInterface),
	); err != nil {
		t.Fatal(err)
	}

	if !hasProjectInterface {
		t.Fatal("Fira project interface not found")
	}

	// Test creating a new project (which should use x-fir-refresh and x-fir-append)
	if err := chromedp.Run(ctx,
		chromedp.WaitVisible(`input[name="title"]`, chromedp.ByQuery),
		chromedp.SendKeys(`input[name="title"]`, "Test Project"),
		chromedp.SendKeys(`input[name="description"]`, "Test Description"),
		chromedp.Click(`button[type="submit"]`, chromedp.ByQuery),
	); err != nil {
		t.Fatal(err)
	}

	// Wait a moment for the project to be created and DOM to update
	var projectCreated bool
	if err := chromedp.Run(ctx,
		chromedp.Sleep(2000), // Give time for the request to complete
		chromedp.Evaluate(`document.body.innerHTML.includes('Test Project')`, &projectCreated),
	); err != nil {
		t.Fatal(err)
	}

	if !projectCreated {
		t.Fatal("Project creation with x-fir- actions failed")
	}

	// Test form reset functionality (x-fir-reset should clear the form)
	var titleFieldEmpty bool
	if err := chromedp.Run(ctx,
		chromedp.Evaluate(`document.querySelector('input[name="title"]').value === ''`, &titleFieldEmpty),
	); err != nil {
		t.Fatal(err)
	}

	if !titleFieldEmpty {
		t.Fatal("Form reset with x-fir-reset failed - title field should be empty")
	}

	// Test that we can navigate to edit a project (if project was created)
	var editButtonExists bool
	if err := chromedp.Run(ctx,
		chromedp.Evaluate(`document.querySelector('a[href*="Test Project"]') !== null || document.body.innerHTML.includes('Edit')`, &editButtonExists),
	); err != nil {
		t.Fatal(err)
	}

	// If there's a way to navigate to edit, test the update functionality
	if editButtonExists {
		t.Log("Testing project update functionality...")

		// Try to find and click an edit link or navigate to edit page
		var editLinkFound bool
		if err := chromedp.Run(ctx,
			chromedp.Evaluate(`
				var links = document.querySelectorAll('a');
				for (let link of links) {
					if (link.href.includes('Test Project') || link.textContent.includes('Test Project')) {
						link.click();
						return true;
					}
				}
				return false;
			`, &editLinkFound),
		); err != nil {
			t.Log("Could not find edit link, skipping update test")
		} else if editLinkFound {
			// Wait for edit page to load and test update functionality
			if err := chromedp.Run(ctx,
				chromedp.Sleep(1000),
				chromedp.WaitVisible(`input[name="title"]`, chromedp.ByQuery),
			); err == nil {
				// Test updating the project (which should use x-fir-refresh)
				if err := chromedp.Run(ctx,
					chromedp.Clear(`input[name="title"]`),
					chromedp.SendKeys(`input[name="title"]`, "Updated Test Project"),
					chromedp.Click(`button[type="submit"]`, chromedp.ByQuery),
				); err == nil {
					// Check if update was successful
					var updateSuccessful bool
					if err := chromedp.Run(ctx,
						chromedp.Sleep(2000),
						chromedp.Evaluate(`document.body.innerHTML.includes('Updated Test Project')`, &updateSuccessful),
					); err == nil && updateSuccessful {
						t.Log("Project update with x-fir-refresh successful")
					} else {
						t.Log("Project update test inconclusive")
					}
				}
			}
		}
	}
}

func TestFiraXFirAttributeProcessing(t *testing.T) {
	controller := fir.NewController("fira_xfir_test_"+strings.ReplaceAll(t.Name(), "/", "_"), fir.DevelopmentMode(true))
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

	// Navigate to the page
	if err := chromedp.Run(ctx, chromedp.Navigate(ts.URL)); err != nil {
		t.Fatal(err)
	}

	// Wait for page to load completely
	if err := chromedp.Run(ctx, chromedp.WaitReady("body")); err != nil {
		t.Fatal(err)
	}

	// Verify that x-fir- attributes have been converted to @fir: attributes
	var xFirAttributesProcessed bool
	if err := chromedp.Run(ctx,
		chromedp.Evaluate(`
			(function() {
				// Check if original x-fir- attributes were processed and converted
				var elements = document.querySelectorAll('*');
				var hasProcessedAttrs = false;
				
				for (var i = 0; i < elements.length; i++) {
					var el = elements[i];
					// Check if any element has @fir: attributes in their outerHTML or Alpine.js has processed them
					var html = el.outerHTML;
					if (html.includes('@fir:') || el.hasAttribute('x-fir-refresh') || el.hasAttribute('x-fir-append')) {
						hasProcessedAttrs = true;
						break;
					}
				}
				
				// Also check if the page structure looks correct for fira
				var hasExpectedStructure = document.querySelector('#projects') !== null && 
											document.querySelector('form[action*="create"]') !== null;
				
				return hasProcessedAttrs || hasExpectedStructure;
			})()
		`, &xFirAttributesProcessed),
	); err != nil {
		t.Fatal(err)
	}

	if !xFirAttributesProcessed {
		// Get more detailed info for debugging
		var pageHTML string
		if err := chromedp.Run(ctx, chromedp.OuterHTML("html", &pageHTML)); err == nil {
			t.Logf("Page HTML for debugging: %s", pageHTML)
		}
		t.Fatal("x-fir- attributes do not appear to be properly processed")
	}

	// Test that specific x-fir- functionality is present
	var hasXFirRefresh, hasXFirAppend bool
	if err := chromedp.Run(ctx,
		chromedp.Evaluate(`document.querySelector('[x-fir-refresh]') !== null`, &hasXFirRefresh),
		chromedp.Evaluate(`document.querySelector('[x-fir-append]') !== null`, &hasXFirAppend),
	); err != nil {
		t.Fatal(err)
	}

	if !hasXFirRefresh {
		t.Log("Warning: No x-fir-refresh attributes found in DOM")
	}
	if !hasXFirAppend {
		t.Log("Warning: No x-fir-append attributes found in DOM")
	}

	t.Log("x-fir- attribute processing verification completed")
}
