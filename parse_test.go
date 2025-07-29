package fir

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/livefir/fir/internal/file"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom" // Import atom package
)

//go:embed testdata/public
var testFS embed.FS

// TestDeepMergeEventTemplates tests the deep merging of event templates
func TestDeepMergeEventTemplates(t *testing.T) {
	testCases := []struct {
		evt1           eventTemplates
		evt2           eventTemplates
		expectedResult eventTemplates
	}{
		// Test case 1: Merging two empty eventTemplates
		{
			evt1:           make(eventTemplates),
			evt2:           make(eventTemplates),
			expectedResult: make(eventTemplates),
		},
		// Test case 2: Merging two eventTemplates one empty and one non-empty
		{
			evt1: make(eventTemplates),
			evt2: eventTemplates{
				"event1": eventTemplate{"template3": struct{}{}},
				"event2": eventTemplate{"template4": struct{}{}},
			},
			expectedResult: eventTemplates{
				"event1": eventTemplate{"template3": struct{}{}},
				"event2": eventTemplate{"template4": struct{}{}},
			},
		},
		// Test case 2: Merging two eventTemplates with non-empty values
		{
			evt1: eventTemplates{
				"event1": eventTemplate{"template1": struct{}{}, "template2": struct{}{}},
			},
			evt2: eventTemplates{
				"event1": eventTemplate{"template3": struct{}{}},
				"event2": eventTemplate{"template4": struct{}{}},
			},
			expectedResult: eventTemplates{
				"event1": eventTemplate{"template1": struct{}{}, "template2": struct{}{}, "template3": struct{}{}},
				"event2": eventTemplate{"template4": struct{}{}},
			},
		},
		// Test case 3: Merging two eventTemplates with duplicate values
		{
			evt1: eventTemplates{
				"event1": eventTemplate{"template1": struct{}{}, "template2": struct{}{}},
			},
			evt2: eventTemplates{
				"event1": eventTemplate{"template2": struct{}{}, "template3": struct{}{}},
			},
			expectedResult: eventTemplates{
				"event1": eventTemplate{"template1": struct{}{}, "template2": struct{}{}, "template3": struct{}{}},
			},
		},
		// Test case 4: Merging two eventTemplates with no common keys
		{
			evt1: eventTemplates{
				"event1": eventTemplate{"template1": struct{}{}},
			},
			evt2: eventTemplates{
				"event2": eventTemplate{"template2": struct{}{}},
			},
			expectedResult: eventTemplates{
				"event1": eventTemplate{"template1": struct{}{}},
				"event2": eventTemplate{"template2": struct{}{}},
			},
		},
	}

	for _, tc := range testCases {
		result := deepMergeEventTemplates(tc.evt1, tc.evt2)
		if !reflect.DeepEqual(result, tc.expectedResult) {
			t.Errorf("Expected %v, but got %v", tc.expectedResult, result)
		}
	}
}

// TestExtractTemplates tests the extraction of templates from content
func TestExtractTemplates(t *testing.T) {
	testCases := []struct {
		name           string
		content        []byte
		expectedResult []byte
		expectedBlocks map[string]string
		expectedError  error
	}{
		// Empty content
		{
			name:           "Empty content",
			content:        []byte{},
			expectedResult: []byte{},
			expectedBlocks: map[string]string{},
			expectedError:  nil,
		},
		// Content with no @fir attributes
		{
			name:           "Content with no @fir attributes",
			content:        []byte("<html><head></head><body><div>Hello, World!</div></body></html>"),
			expectedResult: []byte("<html><head></head><body><div>Hello, World!</div></body></html>"),
			expectedBlocks: map[string]string{},
			expectedError:  nil,
		},
		// Content with @fir attributes and valid templates
		{
			name: "Content with @fir:event:ok attributes and valid templates",
			content: []byte(`
				<div @fir:event:ok="template1">
					<h1>{{ .Title }}</h1>
					<p>{{ .Content }}</p>
				</div>
				<div x-on:fir:event:ok="template2">
					<h2>{{ .Subtitle }}</h2>
					<p>{{ .Description }}</p>
				</div>
			`),
			expectedResult: []byte(`
				<div @fir:event:ok::fir-5885da529fe19205="template1">
					<h1>{{ .Title }}</h1>
					<p>{{ .Content }}</p>
				</div>
				<div x-on:fir:event:ok::fir-7e8373b562556ff8="template2">
					<h2>{{ .Subtitle }}</h2>
					<p>{{ .Description }}</p>
				</div>
			`),
			expectedBlocks: map[string]string{
				"fir-5885da529fe19205": `
					<h1>{{ .Title }}</h1>
					<p>{{ .Content }}</p>
				`,
				"fir-7e8373b562556ff8": `
					<h2>{{ .Subtitle }}</h2>
					<p>{{ .Description }}</p>
				`,
			},
			expectedError: nil,
		},
		{
			name: "Content with @fir:[event1:ok,event2:ok] attributes and valid templates",
			content: []byte(`
				<div @fir:[event1:ok,event2:ok]="template1">
					<h1>{{ .Title }}</h1>
					<p>{{ .Content }}</p>
				</div>
				<div x-on:fir:event:ok="template2">
					<h2>{{ .Subtitle }}</h2>
					<p>{{ .Description }}</p>
				</div>
			`),
			expectedResult: []byte(`
				<div @fir:[event1:ok,event2:ok]::fir-5885da529fe19205="template1">
					<h1>{{ .Title }}</h1>
					<p>{{ .Content }}</p>
				</div>
				<div x-on:fir:event:ok::fir-7e8373b562556ff8="template2">
					<h2>{{ .Subtitle }}</h2>
					<p>{{ .Description }}</p>
				</div>
			`),
			expectedBlocks: map[string]string{
				"fir-5885da529fe19205": `
					<h1>{{ .Title }}</h1>
					<p>{{ .Content }}</p>
				`,
				"fir-7e8373b562556ff8": `
					<h2>{{ .Subtitle }}</h2>
					<p>{{ .Description }}</p>
				`,
			},
			expectedError: nil,
		},

		// Content with @fir attributes and invalid templates
		{
			name: "Content with @fir attributes and invalid templates",
			content: []byte(`
				<div @fir:event:ok="template1">
					<h1>Hello, World!</h1>
					<p>This is not a template</p>
				</div>
				<div x-on:fir:event:ok="template2">
					<h2>{{ .Subtitle }}</h2>
					<p>{{ .Description }}</p>
				</div>
			`),
			expectedResult: []byte(`
				<div @fir:event:ok="template1">
					<h1>Hello, World!</h1>
					<p>This is not a template</p>
				</div>
				<div x-on:fir:event:ok::fir-7e8373b562556ff8="template2">
					<h2>{{ .Subtitle }}</h2>
					<p>{{ .Description }}</p>
				</div>
			`),
			expectedBlocks: map[string]string{
				"fir-7e8373b562556ff8": `
					<h2>{{ .Subtitle }}</h2>
					<p>{{ .Description }}</p>
				`,
			},
			expectedError: nil,
		},
		// TODO: figure out how to test nested templates since hash is generated randomly

		// // Content with @fir attributes and nested templates
		// {
		// 	name: "Content with @fir:event:ok attributes and nested templates",
		// 	content: []byte(`
		// 				<div @fir:event:ok="template1">
		// 					<h1>{{ .Title }}</h1>
		// 					<p>{{ .Content }}</p>
		// 					<div x-on:fir:event:ok="template2">
		// 						<h2>{{ .Subtitle }}</h2>
		// 						<p>{{ .Description }}</p>
		// 					</div>
		// 				</div>

		// 			`),
		// 	expectedResult: []byte(`
		// 				<div @fir:event:ok::fir-e864ccdcce27f3be="template1">
		// 					<h1>{{ .Title }}</h1>
		// 					<p>{{ .Content }}</p>
		// 					<div x-on:fir:event:ok::fir-fir-7e8373b562556ff8="template2">
		// 						<h2>{{ .Subtitle }}</h2>
		// 						<p>{{ .Description }}</p>
		// 					</div>
		// 				</div>
		// 			`),
		// 	expectedBlocks: map[string]string{
		// 		"fir-e864ccdcce27f3be": `
		// 					<h1>{{ .Title }}</h1>
		// 					<p>{{ .Content }}</p>
		// 					<div x-on:fir:event:ok::fir-7e8373b562556ff8="template2">
		// 						<h2>{{ .Subtitle }}</h2>
		// 						<p>{{ .Description }}</p>
		// 					</div>
		// 				`,
		// 		"fir-7e8373b562556ff8": `
		// 					<h2>{{ .Subtitle }}</h2>
		// 					<p>{{ .Description }}</p>
		// 				`,
		// 	},
		// 	expectedError: nil,
		// },
	}

	for _, tc := range testCases {
		result, blocks, err := extractTemplates(tc.content)

		if !reflect.DeepEqual(removeSpace(string(result)), removeSpace(string(tc.expectedResult))) {
			t.Errorf("%v: Expected result %s, but got %s", tc.name, tc.expectedResult, result)
		}

		if !reflect.DeepEqual(blocks, tc.expectedBlocks) {
			t.Errorf("%v: Expected blocks %v, but got %v", tc.name, tc.expectedBlocks, blocks)
		}

		if err != tc.expectedError {
			t.Errorf("%v: Expected error %v, but got %v", tc.name, tc.expectedError, err)
		}
	}
}

// Helper function to normalize HTML for comparison
// This removes insignificant whitespace between tags and normalizes attribute order.
func normalizeHTML(t *testing.T, htmlBytes []byte) string {
	if len(htmlBytes) == 0 {
		return ""
	}
	// Use html.ParseFragment to handle fragments correctly, especially with comments
	// Provide a body context for parsing isolated elements.
	nodes, err := html.ParseFragment(bytes.NewReader(htmlBytes), &html.Node{
		Type:     html.ElementNode,
		Data:     "body",    // Context node type
		DataAtom: atom.Body, // Use the correct atom for body
	})
	require.NoError(t, err, "Failed to parse HTML fragment for normalization")

	var buf bytes.Buffer
	// Render each top-level node returned by ParseFragment
	for _, node := range nodes {
		err = html.Render(&buf, node)
		require.NoError(t, err, "Failed to render HTML fragment node for normalization")
	}

	// Basic whitespace cleanup
	s := buf.String()
	// Collapse multiple spaces and remove spaces around tags
	s = strings.Join(strings.Fields(s), " ")
	s = strings.ReplaceAll(s, " >", ">")
	s = strings.ReplaceAll(s, "< /", "</")
	return s
}

func TestProcessRenderAttributes(t *testing.T) {
	tests := []struct {
		name         string
		inputHTML    string
		expectedHTML string // Expected HTML *after* processing (before normalization)
		wantErr      bool
	}{
		// Existing tests (Corrected expectedHTML based on function change)
		{
			name:         "No fir attributes",
			inputHTML:    `<div id="test">Content</div>`,
			expectedHTML: `<div id="test">Content</div>`, // No wrapper
			wantErr:      false,
		},

		// --- Dispatch Action Tests ---
		{
			name:         "Single dispatch event",
			inputHTML:    `<button x-fir-dispatch:[switch-tab]="update-now:ok">Dispatch</button>`,
			expectedHTML: `<button @fir:update-now:ok="$dispatch('switch-tab')">Dispatch</button>`,
			wantErr:      false,
		},
		{
			name:         "Multiple dispatch events",
			inputHTML:    `<button x-fir-dispatch:[switch-tab,now]="update-now:ok">Dispatch Multiple</button>`,
			expectedHTML: `<button @fir:update-now:ok="$dispatch('switch-tab','now')">Dispatch Multiple</button>`,
			wantErr:      false,
		},
		{
			name:         "Dispatch with different trigger event",
			inputHTML:    `<div x-fir-dispatch:[modal-open]="click">Open Modal</div>`,
			expectedHTML: `<div @fir:click:ok="$dispatch('modal-open')">Open Modal</div>`,
			wantErr:      false,
		},
		{
			name:         "Dispatch with state and template",
			inputHTML:    `<form x-fir-dispatch:[form-submit,validation]="submit:pending->form">Submit</form>`,
			expectedHTML: `<form @fir:submit:pending::form="$dispatch('form-submit','validation')">Submit</form>`,
			wantErr:      false,
		},

		// --- Error Cases ---

		{
			name:         "Empty input content",
			inputHTML:    ``,
			expectedHTML: ``, // Function returns empty for empty input
			wantErr:      false,
		},
		{
			name:         "Dispatch with no parameters",
			inputHTML:    `<button x-fir-dispatch:[]="click">Error</button>`,
			expectedHTML: "", // Not checked on error
			wantErr:      true,
		},
		{
			name:         "Dispatch with empty parameter",
			inputHTML:    `<button x-fir-dispatch:[,valid]="click">Error</button>`,
			expectedHTML: "", // Not checked on error
			wantErr:      true,
		},

		// --- New Test Cases ---

		// --- x-fir-refresh tests ---
		{
			name:         "Only x-fir-refresh",
			inputHTML:    `<div x-fir-refresh="inc; dec:ok">Count: {{.}}</div>`,
			expectedHTML: `<div @fir:inc:ok="$fir.replace()" @fir:dec:ok="$fir.replace()">Count: {{.}}</div>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-refresh with modifier",
			inputHTML:    `<div x-fir-refresh="update.debounce">Update</div>`,
			expectedHTML: `<div @fir:update:ok.debounce="$fir.replace()">Update</div>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-refresh ignores target",
			inputHTML:    `<div x-fir-refresh="load->data">Load</div>`,
			expectedHTML: `<div @fir:load:ok="$fir.replace()">Load</div>`,
			wantErr:      false,
		},
		// --- x-fir-remove tests ---
		{
			name:         "Only x-fir-remove",
			inputHTML:    `<div x-fir-remove="delete:ok">Item</div>`,
			expectedHTML: `<div @fir:delete:ok="$fir.removeEl()">Item</div>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-remove multiple events",
			inputHTML:    `<div x-fir-remove="clear:ok, reset:done">Clear</div>`,
			expectedHTML: `<div @fir:[clear:ok,reset:done]="$fir.removeEl()">Clear</div>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-remove ignores target",
			inputHTML:    `<div x-fir-remove="delete=>doDelete">Delete</div>`,
			expectedHTML: `<div @fir:delete:ok="$fir.removeEl()">Delete</div>`,
			wantErr:      false,
		},
		// --- x-fir-remove-parent tests ---
		{
			name:         "Only x-fir-remove-parent",
			inputHTML:    `<div x-fir-remove-parent="delete:ok">Item</div>`,
			expectedHTML: `<div @fir:delete:ok="$fir.removeParentEl()">Item</div>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-remove-parent multiple events",
			inputHTML:    `<div x-fir-remove-parent="clear:ok, reset:done">Clear</div>`,
			expectedHTML: `<div @fir:[clear:ok,reset:done]="$fir.removeParentEl()">Clear</div>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-remove-parent ignores target and action",
			inputHTML:    `<div x-fir-remove-parent="delete->target=>doDelete">Delete</div>`,
			expectedHTML: `<div @fir:delete:ok="$fir.removeParentEl()">Delete</div>`,
			wantErr:      false,
		},
		// --- x-fir-reset tests ---
		{
			name:         "Only x-fir-reset",
			inputHTML:    `<form x-fir-reset="create-chirp">Submit</form>`,
			expectedHTML: `<form @fir:create-chirp:ok="$el.reset()">Submit</form>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-reset with state",
			inputHTML:    `<form x-fir-reset="submit:ok">Form</form>`,
			expectedHTML: `<form @fir:submit:ok="$el.reset()">Form</form>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-reset multiple events",
			inputHTML:    `<form x-fir-reset="create:ok, update:done">Form</form>`,
			expectedHTML: `<form @fir:[create:ok,update:done]="$el.reset()">Form</form>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-reset ignores target",
			inputHTML:    `<form x-fir-reset="submit->myForm">Form</form>`,
			expectedHTML: `<form @fir:submit:ok="$el.reset()">Form</form>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-reset ignores action target",
			inputHTML:    `<form x-fir-reset="submit=>doSubmit">Form</form>`,
			expectedHTML: `<form @fir:submit:ok="$el.reset()">Form</form>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-reset with existing modifiers",
			inputHTML:    `<form x-fir-reset="submit.debounce">Form</form>`,
			expectedHTML: `<form @fir:submit:ok.debounce="$el.reset()">Form</form>`,
			wantErr:      false,
		},
		// --- x-fir-toggle-disabled tests ---
		{
			name:         "Only x-fir-toggle-disabled",
			inputHTML:    `<button x-fir-toggle-disabled="submit">Submit</button>`,
			expectedHTML: `<button @fir:submit:ok="$fir.toggleDisabled()">Submit</button>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-toggle-disabled with state",
			inputHTML:    `<button x-fir-toggle-disabled="submit:pending">Submit</button>`,
			expectedHTML: `<button @fir:submit:pending="$fir.toggleDisabled()">Submit</button>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-toggle-disabled multiple events",
			inputHTML:    `<button x-fir-toggle-disabled="submit:pending, submit:ok">Submit</button>`,
			expectedHTML: `<button @fir:[submit:pending,submit:ok]="$fir.toggleDisabled()">Submit</button>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-toggle-disabled ignores target",
			inputHTML:    `<button x-fir-toggle-disabled="submit->form">Submit</button>`,
			expectedHTML: `<button @fir:submit:ok="$fir.toggleDisabled()">Submit</button>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-toggle-disabled ignores action target",
			inputHTML:    `<button x-fir-toggle-disabled="submit=>doSubmit">Submit</button>`,
			expectedHTML: `<button @fir:submit:ok="$fir.toggleDisabled()">Submit</button>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-toggle-disabled with existing modifiers",
			inputHTML:    `<button x-fir-toggle-disabled="submit.prevent">Submit</button>`,
			expectedHTML: `<button @fir:submit:ok.prevent="$fir.toggleDisabled()">Submit</button>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-toggle-disabled with complex events",
			inputHTML:    `<button x-fir-toggle-disabled="submit:pending, submit:ok, submit:error">Submit</button>`,
			expectedHTML: `<button @fir:[submit:pending,submit:ok,submit:error]="$fir.toggleDisabled()">Submit</button>`,
			wantErr:      false,
		},

		// --- x-fir-redirect tests ---
		{
			name:         "Only x-fir-redirect (default to root)",
			inputHTML:    `<button x-fir-redirect="delete:ok">Delete</button>`,
			expectedHTML: `<button @fir:delete:ok="$fir.redirect(&#39;/&#39;)">Delete</button>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-redirect with URL parameter",
			inputHTML:    `<button x-fir-redirect:home="delete:ok">Delete</button>`,
			expectedHTML: `<button @fir:delete:ok="$fir.redirect(&#39;/home&#39;)">Delete</button>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-redirect with dashboard parameter",
			inputHTML:    `<button x-fir-redirect:dashboard="delete:ok">Delete</button>`,
			expectedHTML: `<button @fir:delete:ok="$fir.redirect(&#39;/dashboard&#39;)">Delete</button>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-redirect multiple events",
			inputHTML:    `<button x-fir-redirect="delete:ok, cancel:done">Action</button>`,
			expectedHTML: `<button @fir:[delete:ok,cancel:done]="$fir.redirect(&#39;/&#39;)">Action</button>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-redirect with state specified",
			inputHTML:    `<button x-fir-redirect="submit:pending">Submit</button>`,
			expectedHTML: `<button @fir:submit:pending="$fir.redirect(&#39;/&#39;)">Submit</button>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-redirect ignores target",
			inputHTML:    `<button x-fir-redirect="delete->form">Delete</button>`,
			expectedHTML: `<button @fir:delete:ok="$fir.redirect(&#39;/&#39;)">Delete</button>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-redirect ignores action target",
			inputHTML:    `<button x-fir-redirect="delete=>doDelete">Delete</button>`,
			expectedHTML: `<button @fir:delete:ok="$fir.redirect(&#39;/&#39;)">Delete</button>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-redirect with existing modifiers",
			inputHTML:    `<button x-fir-redirect="delete.prevent">Delete</button>`,
			expectedHTML: `<button @fir:delete:ok.prevent="$fir.redirect(&#39;/&#39;)">Delete</button>`,
			wantErr:      false,
		},

		// --- No longer precedence-based, all non-duplicate actions are processed ---
		{
			name:         "Multiple actions with different expressions",
			inputHTML:    `<div x-fir-refresh="b" x-fir-remove="c" x-fir-remove-parent="d">Refresh</div>`,
			expectedHTML: `<div @fir:b:ok="$fir.replace()" @fir:c:ok="$fir.removeEl()" @fir:d:ok="$fir.removeParentEl()">Refresh</div>`, // All actions processed
			wantErr:      false,
		},
		{
			name:         "Multiple actions with different expressions (2)",
			inputHTML:    `<div x-fir-remove="c" x-fir-remove-parent="d">Remove</div>`,
			expectedHTML: `<div @fir:c:ok="$fir.removeEl()" @fir:d:ok="$fir.removeParentEl()">Remove</div>`, // Both actions processed
			wantErr:      false,
		},
		{
			name:         "Multiple actions with different expressions (3)",
			inputHTML:    `<div x-fir-remove="c" x-fir-remove-parent="d" x-fir-append:t1="e" x-fir-prepend:t2="f">Remove</div>`,
			expectedHTML: `<div @fir:c:ok="$fir.removeEl()" @fir:d:ok="$fir.removeParentEl()" @fir:e:ok::t1="$fir.appendEl()" @fir:f:ok::t2="$fir.prependEl()">Remove</div>`, // All actions processed
			wantErr:      false,
		},
		{
			name:         "Multiple actions with different expressions (4)",
			inputHTML:    `<div x-fir-remove-parent="d" x-fir-append:t1="e" x-fir-prepend:t2="f">Remove Parent</div>`,
			expectedHTML: `<div @fir:d:ok="$fir.removeParentEl()" @fir:e:ok::t1="$fir.appendEl()" @fir:f:ok::t2="$fir.prependEl()">Remove Parent</div>`, // All actions processed
			wantErr:      false,
		},
		{
			name:         "Multiple actions with different expressions (5)",
			inputHTML:    `<div x-fir-append:t1="e" x-fir-prepend:t2="f">Append</div>`,
			expectedHTML: `<div @fir:e:ok::t1="$fir.appendEl()" @fir:f:ok::t2="$fir.prependEl()">Append</div>`, // Both actions processed
			wantErr:      false,
		},
		// --- Duplicate translated expression filtering test ---
		{
			name:         "Duplicate translated expressions filtered out",
			inputHTML:    `<div x-fir-refresh="update" x-fir-refresh="update">Content</div>`,
			expectedHTML: `<div @fir:update:ok="$fir.replace()">Content</div>`, // Only one instance should remain
			wantErr:      false,
		},
		// --- General tests ---
		{
			name:         "Attributes preserved",
			inputHTML:    `<div id="myDiv" class="p-4" x-fir-refresh="update">Content</div>`,
			expectedHTML: `<div id="myDiv" class="p-4" @fir:update:ok="$fir.replace()">Content</div>`,
			wantErr:      false,
		},
		{
			name:         "Nested elements",
			inputHTML:    `<div><span x-fir-refresh="count">Nested {{.}}</span></div>`,
			expectedHTML: `<div><span @fir:count:ok="$fir.replace()">Nested {{.}}</span></div>`,
			wantErr:      false,
		},
		{
			name:         "Multiple elements",
			inputHTML:    `<div x-fir-refresh="a">A</div><div x-fir-remove="b">B</div>`,
			expectedHTML: `<div @fir:a:ok="$fir.replace()">A</div><div @fir:b:ok="$fir.removeEl()">B</div>`,
			wantErr:      false,
		},
		{
			name:         "Void element",
			inputHTML:    `<input x-fir-refresh="change">`,
			expectedHTML: `<input @fir:change:ok="$fir.replace()">`,
			wantErr:      false,
		},
		{
			name:         "HTML with comments",
			inputHTML:    `<!-- comment --><div x-fir-refresh="load">Load</div><!-- another -->`,
			expectedHTML: `<!-- comment --><div @fir:load:ok="$fir.replace()">Load</div><!-- another -->`,
			wantErr:      false,
		},
		{
			name:         "Empty input content",
			inputHTML:    ``,
			expectedHTML: ``,
			wantErr:      false,
		},
		// --- Error Cases ---

		{
			name:         "Error: Invalid x-fir-refresh expression",
			inputHTML:    `<div x-fir-refresh="click->">Error</div>`,
			expectedHTML: "", // Not checked on error
			wantErr:      true,
		},
		{
			name:         "Error: Invalid x-fir-remove expression",
			inputHTML:    `<div x-fir-remove=".mod">Error</div>`,
			expectedHTML: "", // Not checked on error
			wantErr:      true,
		},

		{
			name:         "Error: Empty x-fir-refresh",
			inputHTML:    `<div x-fir-refresh="">Error</div>`,
			expectedHTML: "", // Not checked on error
			wantErr:      true,
		},
		{
			name:         "Error: Empty x-fir-remove",
			inputHTML:    `<div x-fir-remove="">Error</div>`,
			expectedHTML: "", // Not checked on error
			wantErr:      true,
		},
		{
			name:         "Error: Invalid x-fir-remove-parent expression",
			inputHTML:    `<div x-fir-remove-parent=".mod">Error</div>`,
			expectedHTML: "", // Not checked on error
			wantErr:      true,
		},
		{
			name:         "Error: Empty x-fir-remove-parent",
			inputHTML:    `<div x-fir-remove-parent="">Error</div>`,
			expectedHTML: "", // Not checked on error
			wantErr:      true,
		},
		// --- Alpine.js Directive Modifier Tests ---
		{
			name:         "Alpine.js directive: x-fir-mutation-observer with modifiers",
			inputHTML:    `<div x-fir-mutation-observer.child-list.subtree="observe">Observer</div>`,
			expectedHTML: `<div x-fir-mutation-observer.child-list.subtree="observe">Observer</div>`, // Should remain unchanged
			wantErr:      false,
		},
		{
			name:         "Alpine.js directive: x-fir-refresh with modifiers",
			inputHTML:    `<div x-fir-refresh.once.passive="load">Load</div>`,
			expectedHTML: `<div @fir:load:ok="$fir.replace()">Load</div>`,
			wantErr:      false,
		},

		{
			name:         "Alpine.js directive: x-fir-remove with attribute filter modifier",
			inputHTML:    `<div x-fir-remove.child-list.attribute-filter:class,id="delete">Delete</div>`,
			expectedHTML: `<div @fir:delete:ok="$fir.removeEl()">Delete</div>`,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBytes, err := processRenderAttributes([]byte(tt.inputHTML))

			if tt.wantErr {
				require.Error(t, err, "Expected an error but got none")
			} else {
				require.NoError(t, err, "Got unexpected error")
				// Normalize both expected and actual HTML before comparing
				// Note: normalizeHTML handles the case where expectedHTML might have the wrapper
				// but gotBytes (the actual output) does not.
				normalizedGot := normalizeHTML(t, gotBytes)
				normalizedExpected := normalizeHTML(t, []byte(tt.expectedHTML))
				require.Equal(t, normalizedExpected, normalizedGot, "HTML output mismatch")
			}
		})
	}
}

// TestLayoutSetContentEmptyWithFileExistence tests that layoutSetContentEmpty properly checks
// file existence before determining if a layout is a directory
func TestLayoutSetContentEmptyWithFileExistence(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test layout directory (this should cause an error)
	layoutDir := filepath.Join(tempDir, "layout_dir")
	err := os.MkdirAll(layoutDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create layout directory: %v", err)
	}

	// Create a test layout file
	layoutFile := filepath.Join(tempDir, "layout.html")
	layoutContent := "<html><body>{{template \"content\" .}}</body></html>"
	err = os.WriteFile(layoutFile, []byte(layoutContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create layout file: %v", err)
	}

	// Test with OS filesystem
	t.Run("OS filesystem - directory layout should error", func(t *testing.T) {
		opt := routeOpt{}
		opt.publicDir = tempDir
		opt.existFile = file.ExistFileOS
		opt.readFile = file.ReadFileOS
		opt.embedfs = nil
		opt.funcMapMutex = &sync.RWMutex{}
		opt.funcMap = make(template.FuncMap)

		_, _, err := layoutSetContentEmpty(opt, filepath.Base(layoutDir))
		if err == nil {
			t.Error("Expected error when layout is a directory, but got none")
		}
		if !strings.Contains(err.Error(), "is a directory but must be a file") {
			t.Errorf("Expected directory error message, got: %v", err)
		}
	})

	t.Run("OS filesystem - file layout should work", func(t *testing.T) {
		opt := routeOpt{}
		opt.publicDir = tempDir
		opt.existFile = file.ExistFileOS
		opt.readFile = file.ReadFileOS
		opt.embedfs = nil
		opt.funcMapMutex = &sync.RWMutex{}
		opt.funcMap = make(template.FuncMap)

		_, _, err := layoutSetContentEmpty(opt, filepath.Base(layoutFile))
		if err != nil {
			t.Errorf("Expected no error for file layout, but got: %v", err)
		}
	})

	t.Run("OS filesystem - non-existent layout should treat as content", func(t *testing.T) {
		opt := routeOpt{}
		opt.publicDir = tempDir
		opt.existFile = file.ExistFileOS
		opt.readFile = file.ReadFileOS
		opt.embedfs = nil
		opt.funcMapMutex = &sync.RWMutex{}
		opt.funcMap = make(template.FuncMap)

		// This should parse the layout string as HTML content instead of trying to read a file
		tmpl, _, err := layoutSetContentEmpty(opt, "nonexistent.html")
		if err != nil {
			t.Errorf("Expected no error for non-existent layout (should be treated as content), but got: %v", err)
		}
		if tmpl == nil {
			t.Error("Expected template to be created from content string")
		}
	})
}

// TestLayoutSetContentEmptyWithEmbeddedFS tests file existence checking with embedded filesystem
func TestLayoutSetContentEmptyWithEmbeddedFS(t *testing.T) {
	t.Run("Embedded filesystem - existing file should work", func(t *testing.T) {
		opt := routeOpt{}
		opt.publicDir = "testdata/public"
		opt.existFile = file.ExistFileFS(testFS)
		opt.readFile = file.ReadFileFS(testFS)
		opt.embedfs = &testFS
		opt.funcMapMutex = &sync.RWMutex{}
		opt.funcMap = make(template.FuncMap)

		// Use a file that exists in the embedded filesystem
		_, _, err := layoutSetContentEmpty(opt, "index.html")
		if err != nil {
			t.Errorf("Expected no error for existing embedded file, but got: %v", err)
		}
	})

	t.Run("Embedded filesystem - directory should error", func(t *testing.T) {
		opt := routeOpt{}
		opt.publicDir = "testdata"
		opt.existFile = file.ExistFileFS(testFS)
		opt.readFile = file.ReadFileFS(testFS)
		opt.embedfs = &testFS
		opt.funcMapMutex = &sync.RWMutex{}
		opt.funcMap = make(template.FuncMap)

		// Try to use a directory as layout
		_, _, err := layoutSetContentEmpty(opt, "public")
		if err == nil {
			t.Error("Expected error when layout is a directory in embedded fs, but got none")
		}
		if !strings.Contains(err.Error(), "is a directory but must be a file") {
			t.Errorf("Expected directory error message, got: %v", err)
		}
	})

	t.Run("Embedded filesystem - non-existent file should treat as content", func(t *testing.T) {
		opt := routeOpt{}
		opt.publicDir = "testdata/public"
		opt.existFile = file.ExistFileFS(testFS)
		opt.readFile = file.ReadFileFS(testFS)
		opt.embedfs = &testFS
		opt.funcMapMutex = &sync.RWMutex{}
		opt.funcMap = make(template.FuncMap)

		// This should parse the layout string as HTML content instead of trying to read a file
		tmpl, _, err := layoutSetContentEmpty(opt, "nonexistent.html")
		if err != nil {
			t.Errorf("Expected no error for non-existent layout in embedded fs, but got: %v", err)
		}
		if tmpl == nil {
			t.Error("Expected template to be created for non-existent layout")
		}
	})
}

func TestFixCorruptedGoTemplateSyntax(t *testing.T) {
	t.Run("BasicCorruptedSyntax", func(t *testing.T) {
		input := `{{="" if="" eq="" .field="" "value"="" }}=""`
		expected := `{{ if eq .field "value" }}`
		result := fixCorruptedGoTemplateSyntax(input)
		require.Equal(t, expected, result, "Should fix basic corrupted template syntax")
	})

	t.Run("MultipleCorruptedPatterns", func(t *testing.T) {
		input := `<div>{{="" if="" eq="" .user.name="" "john"="" }}="" Hello {{="" .user.name="" }}=""</div>`
		expected := `<div>{{ if eq .user.name "john" }} Hello {{ .user.name }}</div>`
		result := fixCorruptedGoTemplateSyntax(input)
		require.Equal(t, expected, result, "Should fix multiple corrupted patterns in same content")
	})

	t.Run("ComplexExpression", func(t *testing.T) {
		input := `{{="" range="" $i,="" $item="" :="" .items="" }}="" {{="" $item.name="" }}="" {{="" end="" }}=""`
		expected := `{{ range $i, $item : .items }} {{ $item.name }} {{ end }}`
		result := fixCorruptedGoTemplateSyntax(input)
		require.Equal(t, expected, result, "Should fix complex range expressions")
	})

	t.Run("NestedConditions", func(t *testing.T) {
		input := `{{="" if="" and="" (eq="" .status="" "active")="" (gt="" .count="" 0)="" }}="" Active {{="" else="" }}="" Inactive {{="" end="" }}=""`
		expected := `{{ if and (eq .status "active") (gt .count 0) }} Active {{ else }} Inactive {{ end }}`
		result := fixCorruptedGoTemplateSyntax(input)
		require.Equal(t, expected, result, "Should fix nested conditional expressions")
	})

	t.Run("EmptyString", func(t *testing.T) {
		input := ""
		expected := ""
		result := fixCorruptedGoTemplateSyntax(input)
		require.Equal(t, expected, result, "Should handle empty string input")
	})

	t.Run("NoCorruptedSyntax", func(t *testing.T) {
		input := `<div>{{ if eq .name "test" }}Hello {{ .name }}{{ end }}</div>`
		expected := `<div>{{ if eq .name "test" }}Hello {{ .name }}{{ end }}</div>`
		result := fixCorruptedGoTemplateSyntax(input)
		require.Equal(t, expected, result, "Should leave valid template syntax unchanged")
	})

	t.Run("MixedValidAndCorrupted", func(t *testing.T) {
		input := `<div>{{ .validField }}{{="" corrupted="" field="" }}="" more content</div>`
		expected := `<div>{{ .validField }}{{ corrupted field }} more content</div>`
		result := fixCorruptedGoTemplateSyntax(input)
		require.Equal(t, expected, result, "Should only fix corrupted parts while leaving valid syntax alone")
	})

	t.Run("CorruptedWithSpecialCharacters", func(t *testing.T) {
		input := `{{="" printf="" "%s=%d"="" .key="" .value="" }}=""`
		expected := `{{ printf "%s=%d" .key .value }}`
		result := fixCorruptedGoTemplateSyntax(input)
		require.Equal(t, expected, result, "Should handle special characters in corrupted syntax")
	})

	t.Run("MultilineCorrupted", func(t *testing.T) {
		input := `{{="" if="" 
		eq="" 
		.multiline="" 
		"test"="" }}=""`
		expected := `{{ if eq .multiline "test" }}`
		result := fixCorruptedGoTemplateSyntax(input)
		require.Equal(t, expected, result, "Should handle multiline corrupted syntax")
	})

	t.Run("CorruptedWithExtraSpaces", func(t *testing.T) {
		input := `{{=""   if=""    eq=""   .field=""   "value"=""   }}=""`
		expected := `{{ if eq .field "value" }}`
		result := fixCorruptedGoTemplateSyntax(input)
		require.Equal(t, expected, result, "Should handle extra spaces in corrupted syntax")
	})

	t.Run("EmptyCorruptedPattern", func(t *testing.T) {
		input := `{{="" empty="" }}=""`
		expected := `{{ empty }}`
		result := fixCorruptedGoTemplateSyntax(input)
		require.Equal(t, expected, result, "Should handle minimal corrupted pattern")
	})

	t.Run("PartiallyCorrupted", func(t *testing.T) {
		input := `{{="" partial="" corruption with normal text`
		expected := `{{="" partial="" corruption with normal text`
		result := fixCorruptedGoTemplateSyntax(input)
		require.Equal(t, expected, result, "Should not modify partially corrupted patterns that don't match full pattern")
	})

	t.Run("NonMatchingPattern", func(t *testing.T) {
		input := `{{="" }}=""`    // This doesn't match because there's nothing between the quotes
		expected := `{{="" }}=""` // Should remain unchanged
		result := fixCorruptedGoTemplateSyntax(input)
		require.Equal(t, expected, result, "Should leave patterns that don't match regex unchanged")
	})

	t.Run("CorruptedFunctionCalls", func(t *testing.T) {
		input := `{{="" call="" .method="" "arg1"="" "arg2"="" }}=""`
		expected := `{{ call .method "arg1" "arg2" }}`
		result := fixCorruptedGoTemplateSyntax(input)
		require.Equal(t, expected, result, "Should fix corrupted function calls")
	})

	t.Run("CorruptedPipeOperations", func(t *testing.T) {
		input := `{{="" .value="" |="" upper="" |="" printf="" "Value:="" %s"="" }}=""`
		expected := `{{ .value | upper | printf "Value: %s" }}`
		result := fixCorruptedGoTemplateSyntax(input)
		require.Equal(t, expected, result, "Should fix corrupted pipe operations")
	})

	t.Run("CorruptedVariableAssignment", func(t *testing.T) {
		input := `{{="" $var="" :="" .field="" }}="" {{="" $var="" }}=""`
		expected := `{{ $var : .field }} {{ $var }}`
		result := fixCorruptedGoTemplateSyntax(input)
		require.Equal(t, expected, result, "Should fix corrupted variable assignment")
	})

	t.Run("ComplexRealWorldExample", func(t *testing.T) {
		input := `<div class="user-card">
		{{="" if="" .user="" }}=""
			<h2>{{="" .user.name="" }}=""</h2>
			<p>Email: {{="" .user.email="" }}=""</p>
			{{="" if="" gt="" .user.age="" 18="" }}=""
				<span class="adult">Adult</span>
			{{="" else="" }}=""
				<span class="minor">Minor</span>
			{{="" end="" }}=""
		{{="" end="" }}=""
	</div>`
		expected := `<div class="user-card">
		{{ if .user }}
			<h2>{{ .user.name }}</h2>
			<p>Email: {{ .user.email }}</p>
			{{ if gt .user.age 18 }}
				<span class="adult">Adult</span>
			{{ else }}
				<span class="minor">Minor</span>
			{{ end }}
		{{ end }}
	</div>`
		result := fixCorruptedGoTemplateSyntax(input)
		require.Equal(t, expected, result, "Should fix complex real-world corrupted template")
	})

	t.Run("CorruptedWithRange", func(t *testing.T) {
		input := `{{="" range="" .items="" }}="" <li>{{="" .name="" }}=""</li> {{="" end="" }}=""`
		expected := `{{ range .items }} <li>{{ .name }}</li> {{ end }}`
		result := fixCorruptedGoTemplateSyntax(input)
		require.Equal(t, expected, result, "Should fix corrupted range loops")
	})

	t.Run("CorruptedWithWith", func(t *testing.T) {
		input := `{{="" with="" .user="" }}="" Hello {{="" .name="" }}=""! {{="" end="" }}=""`
		expected := `{{ with .user }} Hello {{ .name }}! {{ end }}`
		result := fixCorruptedGoTemplateSyntax(input)
		require.Equal(t, expected, result, "Should fix corrupted with statements")
	})

	t.Run("LargeContent", func(t *testing.T) {
		// Test performance with larger content - using regular syntax
		var input strings.Builder
		var expected strings.Builder

		// Generate content with regular Go template syntax
		for i := 0; i < 100; i++ {
			line := fmt.Sprintf("<div>{{ if eq .field%d \"value\" }} Content {{ end }}</div>\n", i)
			input.WriteString(line)
			expected.WriteString(line) // Should remain unchanged
		}

		result := fixCorruptedGoTemplateSyntax(input.String())
		require.Equal(t, expected.String(), result, "Should handle large content efficiently")
	})

	t.Run("CorruptedWithComments", func(t *testing.T) {
		input := `{{="" /*="" comment="" */="" if="" .field="" }}="" content {{="" end="" }}=""`
		expected := `{{ /* comment */ if .field }} content {{ end }}`
		result := fixCorruptedGoTemplateSyntax(input)
		require.Equal(t, expected, result, "Should fix corrupted syntax with comments")
	})

	t.Run("EdgeCaseEmptyParts", func(t *testing.T) {
		// Test with content that doesn't match the regex pattern
		input := `{{=""=""=""=""}}=""`
		expected := `{{=""=""=""=""}}=""` // Should remain unchanged since it doesn't match the regex
		result := fixCorruptedGoTemplateSyntax(input)
		require.Equal(t, expected, result, "Should leave non-matching patterns unchanged")
	})

	t.Run("CorruptedWithDefine", func(t *testing.T) {
		input := `{{="" define="" "template-name"="" }}="" content {{="" end="" }}=""`
		expected := `{{ define "template-name" }} content {{ end }}`
		result := fixCorruptedGoTemplateSyntax(input)
		require.Equal(t, expected, result, "Should fix corrupted define statements")
	})

	t.Run("CorruptedWithTemplate", func(t *testing.T) {
		input := `{{="" template="" "partial"="" .data="" }}=""`
		expected := `{{ template "partial" .data }}`
		result := fixCorruptedGoTemplateSyntax(input)
		require.Equal(t, expected, result, "Should fix corrupted template includes")
	})

	t.Run("CorruptedWithBlock", func(t *testing.T) {
		input := `{{="" block="" "main"="" .="" }}="" default content {{="" end="" }}=""`
		expected := `{{ block "main" . }} default content {{ end }}`
		result := fixCorruptedGoTemplateSyntax(input)
		require.Equal(t, expected, result, "Should fix corrupted block statements")
	})
}
