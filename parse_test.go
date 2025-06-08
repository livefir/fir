package fir

import (
	"bytes"
	"embed"
	"html/template"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

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
		{
			name:         "Simple render, no action",
			inputHTML:    `<div x-fir-live="click">Click Me</div>`,
			expectedHTML: `<div @fir:click:ok="$fir.replace()">Click Me</div>`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "Render with state, no action",
			inputHTML:    `<div x-fir-live="submit:pending">Submitting...</div>`,
			expectedHTML: `<div @fir:submit:pending="$fir.replace()">Submitting...</div>`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "Render with template target, no action",
			inputHTML:    `<form x-fir-live="submit->myform">Form</form>`,
			expectedHTML: `<form @fir:submit:ok::myform="$fir.replace()">Form</form>`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "Render with action target, action defined",
			inputHTML:    `<button x-fir-live="click=>doClick" x-fir-js:doClick="handleMyClick()">Click</button>`,
			expectedHTML: `<button @fir:click:ok="handleMyClick()">Click</button>`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "Render with action target, action NOT defined",
			inputHTML:    `<button x-fir-live="click=>doClick">Click</button>`,
			expectedHTML: `<button @fir:click:ok="doClick">Click</button>`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "Render with template and action target, action defined",
			inputHTML:    `<form x-fir-live="submit->myForm=>doSubmit" x-fir-js:doSubmit="submitTheForm()">Submit</form>`,
			expectedHTML: `<form @fir:submit:ok::myForm="submitTheForm()">Submit</form>`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "Render with template and action target, action NOT defined",
			inputHTML:    `<form x-fir-live="submit->myForm=>doSubmit">Submit</form>`,
			expectedHTML: `<form @fir:submit:ok::myForm="doSubmit">Submit</form>`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "Multiple events, template and action target, action defined",
			inputHTML:    `<div x-fir-live="create:ok,update:ok->item=>saveItem" x-fir-js:saveItem="save()">Save</div>`,
			expectedHTML: `<div @fir:[create:ok,update:ok]::item="save()">Save</div>`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "Multiple expressions, actions defined",
			inputHTML:    `<div x-fir-live="save=>saveData;load=>loadData" x-fir-js:saveData="doSave()" x-fir-js:loadData="doLoad()">Data</div>`,
			expectedHTML: `<div @fir:save:ok="doSave()" @fir:load:ok="doLoad()">Data</div>`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "Multiple expressions, one action defined",
			inputHTML:    `<div x-fir-live="save=>saveData;load=>loadData" x-fir-js:saveData="doSave()">Data</div>`,
			expectedHTML: `<div @fir:save:ok="doSave()" @fir:load:ok="loadData">Data</div>`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "Render with modifier, no action",
			inputHTML:    `<div x-fir-live="click.debounce">Debounced Click</div>`,
			expectedHTML: `<div @fir:click:ok.debounce="$fir.replace()">Debounced Click</div>`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "Render with modifier and action, action defined",
			inputHTML:    `<button x-fir-live="click.throttle=>handleClick" x-fir-js:handleClick="throttledClick()">Click</button>`,
			expectedHTML: `<button @fir:click:ok.throttle="throttledClick()">Click</button>`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "Complex mix with modifiers and actions",
			inputHTML:    `<div x-fir-live="create:ok.nohtml,delete:error->todo=>replaceIt;update:pending.debounce->done=>archiveIt" x-fir-js:replaceIt="doReplace()" x-fir-js:archiveIt="doArchive()">Complex</div>`,
			expectedHTML: `<div @fir:[create:ok,delete:error]::todo.nohtml="doReplace()" @fir:update:pending::done.debounce="doArchive()">Complex</div>`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "Retain other attributes",
			inputHTML:    `<button id="myBtn" class="btn" x-fir-live="click=>doClick" x-fir-js:doClick="handleClick()">Click</button>`,
			expectedHTML: `<button id="myBtn" class="btn" @fir:click:ok="handleClick()">Click</button>`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "Nested elements with render attributes",
			inputHTML:    `<div><div x-fir-live="load=>loadOuter" x-fir-js:loadOuter="outer()"><button x-fir-live="click=>clickInner" x-fir-js:clickInner="inner()">Inner</button></div></div>`,
			expectedHTML: `<div><div @fir:load:ok="outer()"><button @fir:click:ok="inner()">Inner</button></div></div>`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "Multiple actions on one node",
			inputHTML:    `<div x-fir-live="save=>saveIt; load=>loadIt" x-fir-js:saveIt="doSave()" x-fir-js:loadIt="doLoad()">Data</div>`,
			expectedHTML: `<div @fir:save:ok="doSave()" @fir:load:ok="doLoad()">Data</div>`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "Action key with hyphen",
			inputHTML:    `<button x-fir-live="click=>my-action" x-fir-js:my-action="handleIt()">Click</button>`,
			expectedHTML: `<button @fir:click:ok="handleIt()">Click</button>`, // Not checked on error
			wantErr:      false,
		},

		// --- Dispatch Action Tests ---
		{
			name:         "Single dispatch event",
			inputHTML:    `<button x-fir-dispatch:[switch-tab]="update-now:ok">Dispatch</button>`,
			expectedHTML: `<button @fir:update-now:ok.nohtml="$dispatch('switch-tab')">Dispatch</button>`,
			wantErr:      false,
		},
		{
			name:         "Multiple dispatch events",
			inputHTML:    `<button x-fir-dispatch:[switch-tab,now]="update-now:ok">Dispatch Multiple</button>`,
			expectedHTML: `<button @fir:update-now:ok.nohtml="$dispatch('switch-tab','now')">Dispatch Multiple</button>`,
			wantErr:      false,
		},
		{
			name:         "Dispatch with different trigger event",
			inputHTML:    `<div x-fir-dispatch:[modal-open]="click">Open Modal</div>`,
			expectedHTML: `<div @fir:click:ok.nohtml="$dispatch('modal-open')">Open Modal</div>`,
			wantErr:      false,
		},
		{
			name:         "Dispatch with state and template",
			inputHTML:    `<form x-fir-dispatch:[form-submit,validation]="submit:pending->form">Submit</form>`,
			expectedHTML: `<form @fir:submit:pending::form.nohtml="$dispatch('form-submit','validation')">Submit</form>`,
			wantErr:      false,
		},

		// --- Error Cases ---
		{
			name:         "Invalid render expression - bad state",
			inputHTML:    `<div x-fir-live="click:badstate">Error</div>`,
			expectedHTML: "", // Not checked on error
			wantErr:      true,
		},
		{
			name:         "Invalid render expression - bad syntax",
			inputHTML:    `<div x-fir-live="click->">Error</div>`,
			expectedHTML: "", // Not checked on error
			wantErr:      true,
		},
		{
			name:         "Invalid render expression - empty",
			inputHTML:    `<div x-fir-live="">Error</div>`,
			expectedHTML: "", // Not checked on error
			wantErr:      true,
		},
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
		{
			name:         "Void element",
			inputHTML:    `<input x-fir-live="change=>updateValue" x-fir-js:updateValue="val()">`,
			expectedHTML: `<input @fir:change:ok="val()">`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "HTML with comments",
			inputHTML:    `<!-- comment --><div x-fir-live="load=>init" x-fir-js:init="doInit()">Load</div><!-- another -->`,
			expectedHTML: `<!-- comment --><div @fir:load:ok="doInit()">Load</div><!-- another -->`, // No wrapper
			wantErr:      false,
		},

		// --- New Test Cases ---
		{
			name:         "Whitespace variations in render attribute",
			inputHTML:    `<button x-fir-live="  click => doClick  ;  submit -> myForm => doSubmit " x-fir-js:doClick="clickAction()" x-fir-js:doSubmit="submitAction()">WS</button>`,
			expectedHTML: `<button @fir:click:ok="clickAction()" @fir:submit:ok::myForm="submitAction()">WS</button>`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "Action value with quotes",
			inputHTML:    `<div x-fir-live="load=>loadData" x-fir-js:loadData="load('item')">Load</div>`,
			expectedHTML: `<div @fir:load:ok="load('item')">Load</div>`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "Action value with escaped quotes",
			inputHTML:    `<div x-fir-live="load=>loadData" x-fir-js:loadData="load(&#34;item&#34;)">Load</div>`,
			expectedHTML: `<div @fir:load:ok="load(&#34;item&#34;)">Load</div>`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "Multiple states combined",
			inputHTML:    `<div x-fir-live="save:ok,delete:error=>handleResult" x-fir-js:handleResult="process()">Handle</div>`,
			expectedHTML: `<div @fir:[save:ok,delete:error]="process()">Handle</div>`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "Multiple states with template",
			inputHTML:    `<div x-fir-live="save:ok,delete:error->result=>handleResult" x-fir-js:handleResult="process()">Handle</div>`,
			expectedHTML: `<div @fir:[save:ok,delete:error]::result="process()">Handle</div>`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "Unused x-fir-action attribute",
			inputHTML:    `<div x-fir-live="click=>doClick" x-fir-js:doClick="clickAction()" x-fir-js:unused="unused()">Click</div>`,
			expectedHTML: `<div @fir:click:ok="clickAction()">Click</div>`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "x-fir-action attribute without corresponding render action key",
			inputHTML:    `<div x-fir-live="click" x-fir-js:doClick="clickAction()">Click</div>`,
			expectedHTML: `<div @fir:click:ok="$fir.replace()">Click</div>`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "Render expression with only modifier",
			inputHTML:    `<div x-fir-live=".debounce">Click</div>`,
			expectedHTML: "", // Invalid syntax
			wantErr:      true,
		},
		{
			name:         "Render expression with only target",
			inputHTML:    `<div x-fir-live="->myTemplate">Click</div>`,
			expectedHTML: "", // Invalid syntax
			wantErr:      true,
		},
		{
			name:         "Render expression with only action key",
			inputHTML:    `<div x-fir-live="=>myAction" x-fir-js:myAction="act()">Click</div>`,
			expectedHTML: "", // Invalid syntax
			wantErr:      true,
		},

		// --- Fir Action Rule Tests ---
		{
			name:         "Render with Fir Action",
			inputHTML:    `<div x-fir-live="click => $fir.ActionX()">Click</div>`, // Changed X to ActionX
			expectedHTML: `<div @fir:click:ok="$fir.ActionX()">Click</div>`,       // Changed X to ActionX
			wantErr:      false,
		},
		{
			name:         "Render with state and Fir Action",
			inputHTML:    `<div x-fir-live="submit:pending => $fir.ActionY()">Submitting...</div>`, // Changed Y to ActionY
			expectedHTML: `<div @fir:submit:pending="$fir.ActionY()">Submitting...</div>`,          // Changed Y to ActionY
			wantErr:      false,
		},
		{
			name:         "Render with template and Fir Action",
			inputHTML:    `<form x-fir-live="submit->myForm => $fir.ActionZ()">Form</form>`, // Changed Z to ActionZ
			expectedHTML: `<form @fir:submit:ok::myForm="$fir.ActionZ()">Form</form>`,       // Changed Z to ActionZ
			wantErr:      false,
		},
		{
			name:         "Multiple events with Fir Action",
			inputHTML:    `<div x-fir-live="create:ok,update:ok => $fir.ActionA()">Save</div>`, // Changed A to ActionA
			expectedHTML: `<div @fir:[create:ok,update:ok]="$fir.ActionA()">Save</div>`,        // Changed A to ActionA
			wantErr:      false,
		},
		{
			name:         "Multiple expressions with Fir Actions",
			inputHTML:    `<div x-fir-live="save => $fir.Save(); load => $fir.Load()">Data</div>`, // Already multi-letter
			expectedHTML: `<div @fir:save:ok="$fir.Save()" @fir:load:ok="$fir.Load()">Data</div>`, // Already multi-letter
			wantErr:      false,
		},
		{
			name:         "Mixed standard and Fir Actions",
			inputHTML:    `<div x-fir-live="save => saveData; load => $fir.Load()" x-fir-js:saveData="doSave()">Data</div>`, // Already multi-letter
			expectedHTML: `<div @fir:save:ok="doSave()" @fir:load:ok="$fir.Load()">Data</div>`,                              // Already multi-letter
			wantErr:      false,
		},
		{
			name:         "Render with modifier and Fir Action",
			inputHTML:    `<button x-fir-live="click.debounce => $fir.ActionB()">Click</button>`, // Changed B to ActionB
			expectedHTML: `<button @fir:click:ok.debounce="$fir.ActionB()">Click</button>`,       // Changed B to ActionB
			wantErr:      false,
		},
		{
			name:      "Complex mix with Fir Action",
			inputHTML: `<div x-fir-live="create:ok.nohtml->todo=>replaceIt;update:pending.debounce->done=>$fir.Archive()" x-fir-js:replaceIt="doReplace()">Complex</div>`, // Already multi-letter
			// Corrected expectedHTML: removed delete:error, adjusted attribute structure
			expectedHTML: `<div @fir:create:ok::todo.nohtml="doReplace()" @fir:update:pending::done.debounce="$fir.Archive()">Complex</div>`, // Already multi-letter
			wantErr:      false,
		},
		{
			name:         "Invalid: Modifier after Fir Action in render",
			inputHTML:    `<div x-fir-live="click => $fir.ActionX().mod">Error</div>`, // Changed X to ActionX
			expectedHTML: "",                                                          // Not checked on error
			wantErr:      true,                                                        // Parser should reject this based on lexer changes
		},
		{
			name:         "Invalid: Fir Action with incorrect format in render",
			inputHTML:    `<div x-fir-live="click => $fir.1()">Error</div>`, // Format error, no change needed
			expectedHTML: "",                                                // Not checked on error
			wantErr:      true,                                              // Parser should reject this
		},

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
			inputHTML:    `<div x-fir-remove="delete:ok.nohtml">Item</div>`,
			expectedHTML: `<div @fir:delete:ok.nohtml="$fir.removeEl()">Item</div>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-remove multiple events",
			inputHTML:    `<div x-fir-remove="clear:ok, reset:done">Clear</div>`,
			expectedHTML: `<div @fir:[clear:ok,reset:done].nohtml="$fir.removeEl()">Clear</div>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-remove ignores target",
			inputHTML:    `<div x-fir-remove="delete=>doDelete">Delete</div>`,
			expectedHTML: `<div @fir:delete:ok.nohtml="$fir.removeEl()">Delete</div>`,
			wantErr:      false,
		},
		// --- x-fir-remove-parent tests ---
		{
			name:         "Only x-fir-remove-parent",
			inputHTML:    `<div x-fir-remove-parent="delete:ok.nohtml">Item</div>`,
			expectedHTML: `<div @fir:delete:ok.nohtml="$fir.removeParentEl()">Item</div>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-remove-parent multiple events",
			inputHTML:    `<div x-fir-remove-parent="clear:ok, reset:done">Clear</div>`,
			expectedHTML: `<div @fir:[clear:ok,reset:done].nohtml="$fir.removeParentEl()">Clear</div>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-remove-parent ignores target and action",
			inputHTML:    `<div x-fir-remove-parent="delete->target=>doDelete">Delete</div>`,
			expectedHTML: `<div @fir:delete:ok.nohtml="$fir.removeParentEl()">Delete</div>`,
			wantErr:      false,
		},
		// --- x-fir-reset tests ---
		{
			name:         "Only x-fir-reset",
			inputHTML:    `<form x-fir-reset="create-chirp">Submit</form>`,
			expectedHTML: `<form @fir:create-chirp:ok.nohtml="$el.reset()">Submit</form>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-reset with state",
			inputHTML:    `<form x-fir-reset="submit:ok">Form</form>`,
			expectedHTML: `<form @fir:submit:ok.nohtml="$el.reset()">Form</form>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-reset multiple events",
			inputHTML:    `<form x-fir-reset="create:ok, update:done">Form</form>`,
			expectedHTML: `<form @fir:[create:ok,update:done].nohtml="$el.reset()">Form</form>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-reset ignores target",
			inputHTML:    `<form x-fir-reset="submit->myForm">Form</form>`,
			expectedHTML: `<form @fir:submit:ok.nohtml="$el.reset()">Form</form>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-reset ignores action target",
			inputHTML:    `<form x-fir-reset="submit=>doSubmit">Form</form>`,
			expectedHTML: `<form @fir:submit:ok.nohtml="$el.reset()">Form</form>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-reset with modifier (nohtml is added to existing modifiers)",
			inputHTML:    `<form x-fir-reset="submit.debounce">Form</form>`,
			expectedHTML: `<form @fir:submit:ok.debounce.nohtml="$el.reset()">Form</form>`,
			wantErr:      false,
		},
		// --- x-fir-toggle-disabled tests ---
		{
			name:         "Only x-fir-toggle-disabled",
			inputHTML:    `<button x-fir-toggle-disabled="submit">Submit</button>`,
			expectedHTML: `<button @fir:submit:ok.nohtml="$fir.toggleDisabled()">Submit</button>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-toggle-disabled with state",
			inputHTML:    `<button x-fir-toggle-disabled="submit:pending">Submit</button>`,
			expectedHTML: `<button @fir:submit:pending.nohtml="$fir.toggleDisabled()">Submit</button>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-toggle-disabled multiple events",
			inputHTML:    `<button x-fir-toggle-disabled="submit:pending, submit:ok">Submit</button>`,
			expectedHTML: `<button @fir:[submit:pending,submit:ok].nohtml="$fir.toggleDisabled()">Submit</button>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-toggle-disabled ignores target",
			inputHTML:    `<button x-fir-toggle-disabled="submit->form">Submit</button>`,
			expectedHTML: `<button @fir:submit:ok.nohtml="$fir.toggleDisabled()">Submit</button>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-toggle-disabled ignores action target",
			inputHTML:    `<button x-fir-toggle-disabled="submit=>doSubmit">Submit</button>`,
			expectedHTML: `<button @fir:submit:ok.nohtml="$fir.toggleDisabled()">Submit</button>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-toggle-disabled with modifier (nohtml is added to existing modifiers)",
			inputHTML:    `<button x-fir-toggle-disabled="submit.prevent">Submit</button>`,
			expectedHTML: `<button @fir:submit:ok.nohtml.prevent="$fir.toggleDisabled()">Submit</button>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-toggle-disabled with complex events",
			inputHTML:    `<button x-fir-toggle-disabled="submit:pending, submit:ok, submit:error">Submit</button>`,
			expectedHTML: `<button @fir:[submit:pending,submit:ok,submit:error].nohtml="$fir.toggleDisabled()">Submit</button>`,
			wantErr:      false,
		},
		// --- x-fir-live tests ---
		{
			name:         "Only x-fir-live (no actions)",
			inputHTML:    `<div x-fir-live="save->task">Task</div>`,
			expectedHTML: `<div @fir:save:ok::task="$fir.replace()">Task</div>`, // Assuming TranslateRenderExpression default
			wantErr:      false,
		},
		{
			name:         "x-fir-live with x-fir-action",
			inputHTML:    `<div x-fir-live="save=>doSave" x-fir-js:doSave="saveData()">Data</div>`,
			expectedHTML: `<div @fir:save:ok="saveData()">Data</div>`, // Action is replaced
			wantErr:      false,
		},
		{
			name:         "x-fir-live with multiple x-fir-action",
			inputHTML:    `<div x-fir-live="save=>doSave; load=>doLoad" x-fir-js:doSave="saveData()" x-fir-js:doLoad="loadData()">Data</div>`,
			expectedHTML: `<div @fir:save:ok="saveData()" @fir:load:ok="loadData()">Data</div>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-live with unused x-fir-action",
			inputHTML:    `<div x-fir-live="click=>doClick" x-fir-js:doClick="clickAction()" x-fir-js:unused="unused()">Click</div>`,
			expectedHTML: `<div @fir:click:ok="clickAction()">Click</div>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-live with x-fir-action not matching action key",
			inputHTML:    `<div x-fir-live="click" x-fir-js:doClick="clickAction()">Click</div>`,
			expectedHTML: `<div @fir:click:ok="$fir.replace()">Click</div>`, // Action map doesn't apply
			wantErr:      false,
		},
		// --- Precedence tests ---
		{
			name:         "Precedence: live > refresh > remove > remove-parent",
			inputHTML:    `<div x-fir-live="a=>act" x-fir-refresh="b" x-fir-remove="c" x-fir-remove-parent="d" x-fir-js:act="doAct()">Live</div>`,
			expectedHTML: `<div @fir:a:ok="doAct()">Live</div>`, // Only live is processed
			wantErr:      false,
		},
		{
			name:         "Precedence: refresh > remove > remove-parent",
			inputHTML:    `<div x-fir-refresh="b" x-fir-remove="c" x-fir-remove-parent="d">Refresh</div>`,
			expectedHTML: `<div @fir:b:ok="$fir.replace()">Refresh</div>`, // Only refresh is processed
			wantErr:      false,
		},
		{
			name:         "Precedence: remove > remove-parent",
			inputHTML:    `<div x-fir-remove="c" x-fir-remove-parent="d">Remove</div>`,
			expectedHTML: `<div @fir:c:ok.nohtml="$fir.removeEl()">Remove</div>`, // Only remove is processed
			wantErr:      false,
		},
		{
			name:         "Precedence: remove > remove-parent > append > prepend",
			inputHTML:    `<div x-fir-remove="c" x-fir-remove-parent="d" x-fir-append:t1="e" x-fir-prepend:t2="f">Remove</div>`,
			expectedHTML: `<div @fir:c:ok.nohtml="$fir.removeEl()">Remove</div>`, // Only remove is processed
			wantErr:      false,
		},
		{
			name:         "Precedence: remove-parent > append > prepend",
			inputHTML:    `<div x-fir-remove-parent="d" x-fir-append:t1="e" x-fir-prepend:t2="f">Remove Parent</div>`,
			expectedHTML: `<div @fir:d:ok.nohtml="$fir.removeParentEl()">Remove Parent</div>`, // Only remove-parent is processed
			wantErr:      false,
		},
		{
			name:         "Precedence: append > prepend",
			inputHTML:    `<div x-fir-append:t1="e" x-fir-prepend:t2="f">Append</div>`,
			expectedHTML: `<div @fir:e:ok::t1="$fir.appendEl()">Append</div>`, // Only append is processed
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
			expectedHTML: `<div @fir:a:ok="$fir.replace()">A</div><div @fir:b:ok.nohtml="$fir.removeEl()">B</div>`,
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
			name:         "Error: Invalid x-fir-live expression",
			inputHTML:    `<div x-fir-live="click:badstate">Error</div>`,
			expectedHTML: "", // Not checked on error
			wantErr:      true,
		},
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
			name:         "Error: Empty x-fir-live",
			inputHTML:    `<div x-fir-live="">Error</div>`,
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
			name:         "Alpine.js directive: x-fir-live with complex modifiers",
			inputHTML:    `<div x-fir-live.prevent.stop.outside="click=>doClick" x-fir-js:doClick="handleClick()">Click</div>`,
			expectedHTML: `<div @fir:click:ok="handleClick()">Click</div>`,
			wantErr:      false,
		},
		{
			name:         "Alpine.js directive: x-fir-remove with attribute filter modifier",
			inputHTML:    `<div x-fir-remove.child-list.attribute-filter:class,id="delete">Delete</div>`,
			expectedHTML: `<div @fir:delete:ok.nohtml="$fir.removeEl()">Delete</div>`,
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
		opt.existFile = existFileOS
		opt.readFile = readFileOS
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
		opt.existFile = existFileOS
		opt.readFile = readFileOS
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
		opt.existFile = existFileOS
		opt.readFile = readFileOS
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
		opt.existFile = existFileFS(testFS)
		opt.readFile = readFileFS(testFS)
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
		opt.existFile = existFileFS(testFS)
		opt.readFile = readFileFS(testFS)
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
		opt.existFile = existFileFS(testFS)
		opt.readFile = readFileFS(testFS)
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
