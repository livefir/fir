package fir

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom" // Import atom package
)

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
			inputHTML:    `<button x-fir-live="click=>doClick" x-fir-action-doClick="handleMyClick()">Click</button>`,
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
			inputHTML:    `<form x-fir-live="submit->myForm=>doSubmit" x-fir-action-doSubmit="submitTheForm()">Submit</form>`,
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
			inputHTML:    `<div x-fir-live="create:ok,update:ok->item=>saveItem" x-fir-action-saveItem="save()">Save</div>`,
			expectedHTML: `<div @fir:[create:ok,update:ok]::item="save()">Save</div>`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "Multiple expressions, actions defined",
			inputHTML:    `<div x-fir-live="save=>saveData;load=>loadData" x-fir-action-saveData="doSave()" x-fir-action-loadData="doLoad()">Data</div>`,
			expectedHTML: `<div @fir:save:ok="doSave()" @fir:load:ok="doLoad()">Data</div>`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "Multiple expressions, one action defined",
			inputHTML:    `<div x-fir-live="save=>saveData;load=>loadData" x-fir-action-saveData="doSave()">Data</div>`,
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
			inputHTML:    `<button x-fir-live="click.throttle=>handleClick" x-fir-action-handleClick="throttledClick()">Click</button>`,
			expectedHTML: `<button @fir:click:ok.throttle="throttledClick()">Click</button>`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "Complex mix with modifiers and actions",
			inputHTML:    `<div x-fir-live="create:ok.nohtml,delete:error->todo=>replaceIt;update:pending.debounce->done=>archiveIt" x-fir-action-replaceIt="doReplace()" x-fir-action-archiveIt="doArchive()">Complex</div>`,
			expectedHTML: `<div @fir:[create:ok,delete:error]::todo.nohtml="doReplace()" @fir:update:pending::done.debounce="doArchive()">Complex</div>`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "Retain other attributes",
			inputHTML:    `<button id="myBtn" class="btn" x-fir-live="click=>doClick" x-fir-action-doClick="handleClick()">Click</button>`,
			expectedHTML: `<button id="myBtn" class="btn" @fir:click:ok="handleClick()">Click</button>`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "Nested elements with render attributes",
			inputHTML:    `<div><div x-fir-live="load=>loadOuter" x-fir-action-loadOuter="outer()"><button x-fir-live="click=>clickInner" x-fir-action-clickInner="inner()">Inner</button></div></div>`,
			expectedHTML: `<div><div @fir:load:ok="outer()"><button @fir:click:ok="inner()">Inner</button></div></div>`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "Multiple actions on one node",
			inputHTML:    `<div x-fir-live="save=>saveIt; load=>loadIt" x-fir-action-saveIt="doSave()" x-fir-action-loadIt="doLoad()">Data</div>`,
			expectedHTML: `<div @fir:save:ok="doSave()" @fir:load:ok="doLoad()">Data</div>`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "Action key with hyphen",
			inputHTML:    `<button x-fir-live="click=>my-action" x-fir-action-my-action="handleIt()">Click</button>`,
			expectedHTML: ``,   // Not checked on error
			wantErr:      true, // Expecting parser error for "my-action"
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
			name:         "Void element",
			inputHTML:    `<input x-fir-live="change=>updateValue" x-fir-action-updateValue="val()">`,
			expectedHTML: `<input @fir:change:ok="val()">`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "HTML with comments",
			inputHTML:    `<!-- comment --><div x-fir-live="load=>init" x-fir-action-init="doInit()">Load</div><!-- another -->`,
			expectedHTML: `<!-- comment --><div @fir:load:ok="doInit()">Load</div><!-- another -->`, // No wrapper
			wantErr:      false,
		},

		// --- New Test Cases ---
		{
			name:         "Whitespace variations in render attribute",
			inputHTML:    `<button x-fir-live="  click => doClick  ;  submit -> myForm => doSubmit " x-fir-action-doClick="clickAction()" x-fir-action-doSubmit="submitAction()">WS</button>`,
			expectedHTML: `<button @fir:click:ok="clickAction()" @fir:submit:ok::myForm="submitAction()">WS</button>`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "Action value with quotes",
			inputHTML:    `<div x-fir-live="load=>loadData" x-fir-action-loadData="load('item')">Load</div>`,
			expectedHTML: `<div @fir:load:ok="load('item')">Load</div>`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "Action value with escaped quotes",
			inputHTML:    `<div x-fir-live="load=>loadData" x-fir-action-loadData="load(&#34;item&#34;)">Load</div>`,
			expectedHTML: `<div @fir:load:ok="load(&#34;item&#34;)">Load</div>`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "Multiple states combined",
			inputHTML:    `<div x-fir-live="save:ok,delete:error=>handleResult" x-fir-action-handleResult="process()">Handle</div>`,
			expectedHTML: `<div @fir:[save:ok,delete:error]="process()">Handle</div>`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "Multiple states with template",
			inputHTML:    `<div x-fir-live="save:ok,delete:error->result=>handleResult" x-fir-action-handleResult="process()">Handle</div>`,
			expectedHTML: `<div @fir:[save:ok,delete:error]::result="process()">Handle</div>`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "Unused x-fir-action attribute",
			inputHTML:    `<div x-fir-live="click=>doClick" x-fir-action-doClick="clickAction()" x-fir-action-unused="unused()">Click</div>`,
			expectedHTML: `<div @fir:click:ok="clickAction()">Click</div>`, // No wrapper
			wantErr:      false,
		},
		{
			name:         "x-fir-action attribute without corresponding render action key",
			inputHTML:    `<div x-fir-live="click" x-fir-action-doClick="clickAction()">Click</div>`,
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
			inputHTML:    `<div x-fir-live="=>myAction" x-fir-action-myAction="act()">Click</div>`,
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
			inputHTML:    `<div x-fir-live="save => saveData; load => $fir.Load()" x-fir-action-saveData="doSave()">Data</div>`, // Already multi-letter
			expectedHTML: `<div @fir:save:ok="doSave()" @fir:load:ok="$fir.Load()">Data</div>`,                                  // Already multi-letter
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
			inputHTML: `<div x-fir-live="create:ok.nohtml->todo=>replaceIt;update:pending.debounce->done=>$fir.Archive()" x-fir-action-replaceIt="doReplace()">Complex</div>`, // Already multi-letter
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
			inputHTML:    `<div x-fir-remove-parent="delete:ok.nohtml">Item</div>`,
			expectedHTML: `<div @fir:delete:ok.nohtml="$fir.removeParentEl()">Item</div>`,
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
		// --- x-fir-live tests ---
		{
			name:         "Only x-fir-live (no actions)",
			inputHTML:    `<div x-fir-live="save->task">Task</div>`,
			expectedHTML: `<div @fir:save:ok::task="$fir.replace()">Task</div>`, // Assuming TranslateRenderExpression default
			wantErr:      false,
		},
		{
			name:         "x-fir-live with x-fir-action",
			inputHTML:    `<div x-fir-live="save=>doSave" x-fir-action-doSave="saveData()">Data</div>`,
			expectedHTML: `<div @fir:save:ok="saveData()">Data</div>`, // Action is replaced
			wantErr:      false,
		},
		{
			name:         "x-fir-live with multiple x-fir-action",
			inputHTML:    `<div x-fir-live="save=>doSave; load=>doLoad" x-fir-action-doSave="saveData()" x-fir-action-doLoad="loadData()">Data</div>`,
			expectedHTML: `<div @fir:save:ok="saveData()" @fir:load:ok="loadData()">Data</div>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-live with unused x-fir-action",
			inputHTML:    `<div x-fir-live="click=>doClick" x-fir-action-doClick="clickAction()" x-fir-action-unused="unused()">Click</div>`,
			expectedHTML: `<div @fir:click:ok="clickAction()">Click</div>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-live with x-fir-action not matching action key",
			inputHTML:    `<div x-fir-live="click" x-fir-action-doClick="clickAction()">Click</div>`,
			expectedHTML: `<div @fir:click:ok="$fir.replace()">Click</div>`, // Action map doesn't apply
			wantErr:      false,
		},
		// --- Precedence tests ---
		{
			name:         "Precedence: live > refresh > remove > remove-parent",
			inputHTML:    `<div x-fir-live="a=>act" x-fir-refresh="b" x-fir-remove="c" x-fir-remove-parent="d" x-fir-action-act="doAct()">Live</div>`,
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
			expectedHTML: `<div @fir:c:ok="$fir.removeEl()">Remove</div>`, // Only remove is processed
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
