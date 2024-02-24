package fir

import (
	"reflect"
	"testing"
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
		{
			name: "Content with @fir:event:ok.nohtml attribute and valid templates",
			content: []byte(`
				<div @fir:event:ok.nohtml="template1">
					<h1>{{ .Title }}</h1>
					<p>{{ .Content }}</p>
				</div>
				<div x-on:fir:event:ok="template2">
					<h2>{{ .Subtitle }}</h2>
					<p>{{ .Description }}</p>
				</div>
			`),
			expectedResult: []byte(`
				<div @fir:event:ok::fir-5885da529fe19205.nohtml="template1">
					<h1>{{ .Title }}</h1>
					<p>{{ .Content }}</p>
				</div>
				<div x-on:fir:event:ok::fir-7e8373b562556ff8="template2">
					<h2>{{ .Subtitle }}</h2>
					<p>{{ .Description }}</p>
				</div>
			`),
			expectedBlocks: map[string]string{
				"fir-5885da529fe19205.nohtml": `
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
			name: "Content with @fir:[event1:ok,event2:ok].nohtml attribute and valid templates",
			content: []byte(`
				<div @fir:[event1:ok,event2:ok].nohtml="template1">
					<h1>{{ .Title }}</h1>
					<p>{{ .Content }}</p>
				</div>
				<div x-on:fir:event:ok="template2">
					<h2>{{ .Subtitle }}</h2>
					<p>{{ .Description }}</p>
				</div>
			`),
			expectedResult: []byte(`
				<div @fir:[event1:ok,event2:ok]::fir-5885da529fe19205.nohtml="template1">
					<h1>{{ .Title }}</h1>
					<p>{{ .Content }}</p>
				</div>
				<div x-on:fir:event:ok::fir-7e8373b562556ff8="template2">
					<h2>{{ .Subtitle }}</h2>
					<p>{{ .Description }}</p>
				</div>
			`),
			expectedBlocks: map[string]string{
				"fir-5885da529fe19205.nohtml": `
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
