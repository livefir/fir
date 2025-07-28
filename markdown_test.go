package fir

import (
	"bytes"
	"html/template"
	"strings"
	"testing"

	"github.com/livefir/fir/internal/file"
	"golang.org/x/net/html"
)

func TestMarkdownTemplate(t *testing.T) {
	tmpl := `{{ markdown "./testdata/snippet_input.md" "marker" }}`
	expected := `<p>Snippet Content</p>`
	md := markdown(file.ReadFileOS, file.ExistFileOS)
	var buf bytes.Buffer
	template.Must(template.New("test").Funcs(template.FuncMap{
		"markdown": md,
	}).Parse(tmpl)).Execute(&buf, nil)
	actual := html.UnescapeString(buf.String())
	actualNode, err := html.Parse(strings.NewReader(actual))
	if err != nil {
		t.Fatalf("failed to parse actual HTML: %v", err)
	}
	expectedNode, err := html.Parse(strings.NewReader(expected))
	if err != nil {
		t.Fatalf("failed to parse expected HTML: %v", err)
	}
	if err := areNodesDeepEqual(actualNode, expectedNode); err != nil {
		t.Errorf("Expected:\n%s\n\nGot:\n%s\n, err: %v \n", expected, actual, err)
	}
}

func TestMarkdown(t *testing.T) {
	type markdownTestCase struct {
		name        string
		input       string
		markers     []string
		expected    string
		expectError bool
	}

	markdownTestCases := []markdownTestCase{
		{
			name:     "Test empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "Test basic Markdown conversion",
			input:    "## Heading",
			expected: "<h2 id=\"heading\">Heading</h2>\n",
		},
		{
			name:     "Test snippet with single marker",
			input:    "<!-- start marker -->\nSnippet Content\n<!-- end marker -->",
			markers:  []string{"marker"},
			expected: "<p>Snippet Content</p>",
		},
		{
			name:     "Test snippet with multiple markers",
			input:    "<!-- start marker1 -->\nSnippet Content 1\n<!-- end marker1 -->\nSome Content\n<!-- start marker2 -->\nSnippet Content 2\n<!-- end marker2 -->",
			markers:  []string{"marker1", "marker2"},
			expected: `<p>Snippet Content 1<br />Snippet Content 2</p>`,
		},
		{
			name:        "Test basic Markdown conversion",
			input:       "./testdata/basic_input.md",
			expected:    "<h2 id=\"heading\">Heading</h2>\n",
			expectError: false,
		},
		{
			name:        "Test snippet with single marker",
			input:       "./testdata/snippet_input.md",
			markers:     []string{"marker"},
			expected:    "<p>Snippet Content</p>",
			expectError: false,
		},
		{
			name:    "Test snippet with multiple markers, both",
			input:   "./testdata/snippet_input.md",
			markers: []string{"marker1", "marker2"},
			expected: `<p>Snippet Content 1<br />
            Snippet Content 2</p>`,
			expectError: false,
		},
		{
			name:        "Test snippet with multiple markers, single",
			input:       "./testdata/snippet_input.md",
			markers:     []string{"marker"},
			expected:    `<p>Snippet Content</p>`,
			expectError: false,
		},
		// Add more test cases as needed
	}

	for _, tc := range markdownTestCases {
		t.Run(tc.name, func(t *testing.T) {
			var inputData []byte
			var err error
			if file.ExistFileOS(tc.input) {
				_, inputData, err = file.ReadFileOS(tc.input)
				if err != nil && !tc.expectError {
					t.Errorf("%v: \n Error reading input file: %s: %s", tc.name, tc.input, err)
				}
			} else {
				inputData = []byte(tc.input)
			}

			md := markdown(file.ReadFileOS, file.ExistFileOS)
			actual := md(string(inputData), tc.markers...)

			if tc.expectError {
				if actual != "" {
					t.Errorf("%v: \n  Expected an error, but got a result: %s", tc.name, actual)
				}
			} else {
				actualNode, err := html.Parse(strings.NewReader(string(actual)))
				if err != nil {
					t.Fatalf("%v: \n failed to parse actual HTML: %v", tc.name, err)
				}
				expectedNode, err := html.Parse(strings.NewReader(tc.expected))
				if err != nil {
					t.Fatalf("%v: \n failed to parse expected HTML: %v", tc.name, err)
				}
				if err := areNodesDeepEqual(actualNode, expectedNode); err != nil {
					t.Errorf("%v: \nExpected:\n%s\n\nGot:\n%s\n, err: %v \n", tc.name, tc.expected, actual, err)
				}
			}
		})
	}
}

func TestSnippets(t *testing.T) {
	testCases := []struct {
		description string
		in          []byte
		markers     []string
		expected    []byte
	}{
		// Test case 1: Single marker with snippet content
		{
			description: "Single marker with snippet content",
			in:          []byte("Line 1\n<!-- start marker -->\nSnippet content\n<!-- end marker -->\nLine 4"),
			markers:     []string{"marker"},
			expected:    []byte("Snippet content"),
		},

		// Test case 2: Multiple markers with snippets
		{
			description: "Multiple markers with snippets",
			in:          []byte("Line 1\n<!-- start marker1 -->\nSnippet 1\n<!-- end marker1 -->\nLine 4\n<!-- start marker2 -->\nSnippet 2\n<!-- end marker2 -->\nLine 7"),
			markers:     []string{"marker1", "marker2"},
			expected:    []byte("Snippet 1\nSnippet 2"),
		},

		// Test case 3: No markers present, return original input
		{
			description: "No markers present, return original input",
			in:          []byte("No markers here"),
			markers:     []string{"marker"},
			expected:    []byte("No markers here"),
		},

		// Test case 4: Empty snippet content between markers, return empty byte slice
		{
			description: "Empty snippet content between markers, return empty byte slice",
			in:          []byte("Line 1\n<!-- start marker -->\n\n<!-- end marker -->\nLine 4"),
			markers:     []string{"marker"},
			expected:    []byte{},
		},

		// Test case 5: Start marker with no end marker, return until end of input
		{
			description: "Start marker with no end marker, return until end of input",
			in:          []byte("Line 1\n<!-- start marker -->\nSnippet\nLine 4"),
			markers:     []string{"marker"},
			expected:    []byte("Snippet\nLine 4"),
		},

		// Test case 6: Invalid input - empty input byte slice, return original input
		{
			description: "Invalid input - empty input byte slice, return original input",
			in:          []byte{},
			markers:     []string{"marker"},
			expected:    []byte{},
		},

		// Test case 7: Invalid input - empty markers slice, return original input
		{
			description: "Invalid input - empty markers slice, return original input",
			in:          []byte("Line 1\n<!-- start marker -->\nSnippet\n"),
			markers:     []string{},
			expected:    []byte("Line 1\n<!-- start marker -->\nSnippet\n"),
		},

		// Test case 8: Invalid input - no start marker, return original input
		{
			description: "Invalid input - no start marker, return original input",
			in:          []byte("Line 1\nSnippet\n<!-- end marker -->\nLine 4"),
			markers:     []string{"marker"},
			expected:    []byte("Line 1\nSnippet\n<!-- end marker -->\nLine 4"),
		},

		// Test case 9: Invalid input - start and end markers swapped, return original input
		{
			description: "Invalid input - start and end markers swapped, return original input",
			in:          []byte("Line 1\n<!-- end marker -->\nSnippet\n<!-- start marker -->\nLine 4"),
			markers:     []string{"marker"},
			expected:    []byte("Line 1\n<!-- end marker -->\nSnippet\n<!-- start marker -->\nLine 4"),
		},
	}

	for _, testCase := range testCases {
		result := snippets(testCase.in, testCase.markers)
		if !bytes.Equal(result, testCase.expected) {
			t.Errorf("Test Case: %s\nInput: %s\nMarkers: %v\nExpected: %s\nGot: %s", testCase.description, testCase.in, testCase.markers, testCase.expected, result)
		}
	}
}
