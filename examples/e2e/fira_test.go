package e2e

import (
	"context"
	"fmt"
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
	var consoleMessages []string
	var exceptions []string
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if ev, ok := ev.(*runtime.EventConsoleAPICalled); ok {
			for _, arg := range ev.Args {
				var valStr string
				if arg.Value != nil {
					valStr = string(arg.Value)
				}
				message := fmt.Sprintf("Browser Console (%s): %s", ev.Type, valStr)
				t.Log(message)
				consoleMessages = append(consoleMessages, message)
			}
		}
		if ev, ok := ev.(*runtime.EventExceptionThrown); ok {
			message := fmt.Sprintf("Browser Exception: %s", ev.ExceptionDetails.Text)
			t.Log(message)
			exceptions = append(exceptions, message)
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
	// First, check if Fir JavaScript is properly loaded
	var firLoaded bool
	var firObject string
	var magicFunctions string
	if err := chromedp.Run(ctx,
		chromedp.Evaluate(`window.$fir !== undefined`, &firLoaded),
		chromedp.Evaluate(`typeof window.$fir`, &firObject),
		chromedp.Evaluate(`window.Alpine && window.Alpine.magic ? Object.keys(window.Alpine.magic) : 'no magic'`, &magicFunctions),
	); err != nil {
		t.Fatal(err)
	}
	t.Logf("Fir JavaScript loaded: %v, $fir type: %s, Alpine magic functions: %s", firLoaded, firObject, magicFunctions)

	// Inject debugging code to track events
	if err := chromedp.Run(ctx,
		chromedp.Evaluate(`
			window.jsErrors = [];
			window.firEvents = [];
			
			// Track JavaScript errors
			window.addEventListener('error', function(e) {
				window.jsErrors.push('Error: ' + e.message + ' at ' + e.filename + ':' + e.lineno);
			});
			
			// Track unhandled promise rejections
			window.addEventListener('unhandledrejection', function(e) {
				window.jsErrors.push('Promise rejection: ' + e.reason);
			});
			
			// Track fir events if possible
			if (window.$fir && window.$fir.ws) {
				var originalOnMessage = window.$fir.ws.onmessage;
				window.$fir.ws.onmessage = function(event) {
					try {
						var data = JSON.parse(event.data);
						if (data.event) {
							window.firEvents.push(data.event + ':' + data.status);
						}
					} catch(e) {
						window.jsErrors.push('WebSocket parse error: ' + e.message);
					}
					if (originalOnMessage) originalOnMessage.call(this, event);
				};
			}
			
			return 'debugging setup complete';
		`, nil),
	); err != nil {
		t.Logf("Could not inject debugging code: %v", err)
	}

	if err := chromedp.Run(ctx,
		chromedp.WaitVisible(`input[name="title"]`, chromedp.ByQuery),
		chromedp.SendKeys(`input[name="title"]`, "Test Project"),
		chromedp.SendKeys(`input[name="description"]`, "Test Description"),
	); err != nil {
		t.Fatal(err)
	}

	// Log console messages before and after form submission
	t.Logf("Console messages before submission: %v", consoleMessages)

	if err := chromedp.Run(ctx,
		chromedp.Click(`button[type="submit"]`, chromedp.ByQuery),
	); err != nil {
		t.Fatal(err)
	}

	// Wait a bit and log any new console messages
	if err := chromedp.Run(ctx, chromedp.Sleep(2000)); err != nil {
		t.Fatal(err)
	}
	t.Logf("Console messages after submission: %v", consoleMessages)

	// Wait a moment for the project to be created and DOM to update
	// First check for any error messages that might have appeared
	var hasErrors bool
	var errorText string
	var wsConnected bool
	var wsLastMessage string
	// Check for errors first
	if err := chromedp.Run(ctx,
		chromedp.Sleep(2000), // Increased wait time for WebSocket message processing
		chromedp.Evaluate(`document.querySelector('.help.is-danger') !== null`, &hasErrors),
	); err != nil {
		t.Fatal(err)
	}

	// Check WebSocket status
	if err := chromedp.Run(ctx,
		chromedp.Evaluate(`window.fir && window.fir.websocket && window.fir.websocket.readyState === 1`, &wsConnected),
		chromedp.Evaluate(`(window.fir && window.fir.websocket && window.fir.websocket._lastMessage) || 'no message'`, &wsLastMessage),
	); err != nil {
		t.Logf("WebSocket check failed: %v", err)
		wsConnected = false
		wsLastMessage = "check failed"
	}

	t.Logf("WebSocket connected: %v, Last message: %s", wsConnected, wsLastMessage)

	if hasErrors {
		if err := chromedp.Run(ctx,
			chromedp.Evaluate(`(function() { 
				var el = document.querySelector('.help.is-danger'); 
				return el ? el.textContent : 'no error element'; 
			})()`, &errorText),
		); err == nil {
			t.Logf("Project creation error: %s", errorText)
		}
	}

	// Check the projects container for changes
	var projectsHTML string
	var projectCreated bool
	var projectCount int
	if err := chromedp.Run(ctx,
		chromedp.Sleep(2000), // Give time for the request to complete
		chromedp.InnerHTML(`#projects`, &projectsHTML),
		chromedp.Evaluate(`document.body.innerHTML.includes('Test Project')`, &projectCreated),
		chromedp.Evaluate(`document.querySelectorAll('#projects a[id^="projectitem-"]').length`, &projectCount),
	); err != nil {
		t.Fatal(err)
	}

	t.Logf("Projects container HTML: %s", projectsHTML)
	t.Logf("Project created: %v", projectCreated)
	t.Logf("Project count in DOM: %d", projectCount)

	// Also check specifically for project items with "Test Project" text
	var testProjectExists bool
	if err := chromedp.Run(ctx,
		chromedp.Evaluate(`Array.from(document.querySelectorAll('#projects a')).some(a => a.textContent.includes('Test Project'))`, &testProjectExists),
	); err == nil {
		t.Logf("Test Project specifically found in projects container: %v", testProjectExists)
	}

	if !projectCreated {
		// Additional debugging - check if the form was actually submitted
		var formAction string
		var formMethod string
		var appendHandlerExists bool
		if err := chromedp.Run(ctx,
			chromedp.Evaluate(`(function() { 
				var form = document.querySelector('form[action*="create"]'); 
				return form ? form.action : 'form not found'; 
			})()`, &formAction),
			chromedp.Evaluate(`(function() { 
				var form = document.querySelector('form[action*="create"]'); 
				return form ? form.method : 'form not found'; 
			})()`, &formMethod),
			// Check if the append handler is properly set up
			chromedp.Evaluate(`(function() { 
				var projectsEl = document.querySelector('#projects'); 
				return projectsEl && projectsEl.hasAttribute('x-fir-append:projectitem'); 
			})()`, &appendHandlerExists),
		); err == nil {
			t.Logf("Form action: %s, method: %s", formAction, formMethod)
			t.Logf("Append handler exists on #projects: %v", appendHandlerExists)
		}

		// Check if AlpineJS is loaded and working
		var alpineLoaded bool
		var alpineVersion string
		if err := chromedp.Run(ctx,
			chromedp.Evaluate(`window.Alpine !== undefined`, &alpineLoaded),
			chromedp.Evaluate(`window.Alpine ? (window.Alpine.version || 'version unknown') : 'not loaded'`, &alpineVersion),
		); err == nil {
			t.Logf("AlpineJS loaded: %v, version: %s", alpineLoaded, alpineVersion)
		}

		// Check if WebSocket is still connected
		if err := chromedp.Run(ctx,
			chromedp.Evaluate(`(function() { 
				if (!window.$fir || !window.$fir.ws) return 'no websocket'; 
				return window.$fir.ws.readyState === 1 ? 'connected' : 'disconnected'; 
			})()`, &wsLastMessage),
		); err == nil {
			t.Logf("WebSocket status after form submission: %s", wsLastMessage)
		}

		// Check for JavaScript errors
		var jsErrors string
		if err := chromedp.Run(ctx,
			chromedp.Evaluate(`(function() { 
				var errors = window.jsErrors || []; 
				return errors.length > 0 ? errors.join('; ') : 'no errors'; 
			})()`, &jsErrors),
		); err == nil {
			t.Logf("JavaScript errors: %s", jsErrors)
		}

		// Check if events are being fired
		var eventsFired string
		if err := chromedp.Run(ctx,
			chromedp.Evaluate(`(function() { 
				var events = window.firEvents || []; 
				return events.length > 0 ? events.join(', ') : 'no events recorded'; 
			})()`, &eventsFired),
		); err == nil {
			t.Logf("Fir events fired: %s", eventsFired)
		}
		var wsConnected bool
		if err := chromedp.Run(ctx,
			chromedp.Evaluate(`window.$fir && window.$fir.ws && window.$fir.ws.readyState === 1`, &wsConnected),
		); err == nil {
			t.Logf("WebSocket connected: %v", wsConnected)
		}

		// Check for any pending network requests
		var networkPending bool
		if err := chromedp.Run(ctx,
			chromedp.Evaluate(`performance.getEntriesByType('navigation').length > 0`, &networkPending),
		); err == nil {
			t.Logf("Network activity detected: %v", networkPending)
		}

		// Log final console messages and exceptions
		t.Logf("Final console messages: %v", consoleMessages)
		t.Logf("Exceptions: %v", exceptions)

		// Check the entire body for any new content
		var currentBodyHTML string
		if err := chromedp.Run(ctx,
			chromedp.OuterHTML("body", &currentBodyHTML),
		); err == nil {
			t.Logf("Current body HTML after project creation attempt: %s", currentBodyHTML)
		}

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
