package dom

import (
	"bytes"
	"html/template"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNodes(t *testing.T) {
	// Test cases
	tests := []struct {
		description     string
		templateString  string
		expectedDynamic []string
		expectedStatic  []string
	}{
		{
			description:     "Template with action node",
			templateString:  "Hello {{.Name}} World",
			expectedDynamic: []string{"{{.Name}}"},
			expectedStatic:  []string{"Hello ", " World"},
		},
		{
			description:     "Template with text node",
			templateString:  "Hello, World!",
			expectedDynamic: nil,
			expectedStatic:  []string{"Hello, World!"},
		},
		{
			description:     "Template with conditional nodes",
			templateString:  "{{if .Condition}}Hello{{else}}Goodbye{{end}}",
			expectedDynamic: []string{"{{if .Condition}}Hello{{else}}Goodbye{{end}}"},
			expectedStatic:  nil,
		},
		{
			description:     "Template with nested action nodes",
			templateString:  "{{if .Condition}}{{.Name}}{{else}}{{.Age}}{{end}}",
			expectedDynamic: []string{"{{if .Condition}}{{.Name}}{{else}}{{.Age}}{{end}}"},
			expectedStatic:  nil,
		},
		{
			description:     "Template from todos file",
			templateString:  readTestFile("testdata/todos.html"),
			expectedDynamic: []string{`{{.Planet}}`, "{{.Name}}", "{{.Age}}", "{{.City}}"},
			expectedStatic:  []string{"Hello, On ", "Planet ", " ", "! You are ", " years old.\n", "You live in city ", "\n", "\n"},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			templateObj, err := template.New("").Parse(test.templateString)
			if err != nil {
				t.Fatalf("Failed to parse template (%s): %v", test.description, err)
			}

			got := CalcDiff(templateObj)
			if len(got.Dynamic) != len(test.expectedDynamic) {
				t.Errorf("Dynamic length mismatch in test (%s). Expected: %d, Got: %d", test.description, len(test.expectedDynamic), len(got.Dynamic))
			}
			if diff := cmp.Diff(got.Dynamic, test.expectedDynamic); diff != "" {
				t.Errorf("Dynamic mismatch in test (%s): (-got +want)\n%s", test.description, diff)
			}
			if len(got.Static) != len(test.expectedStatic) {
				t.Errorf("Static length mismatch in test (%s). Expected: %d, Got: %d", test.description, len(test.expectedStatic), len(got.Static))
			}
			if diff := cmp.Diff(got.Static, test.expectedStatic); diff != "" {
				t.Errorf("Static mismatch in test (%s): (-got +want)\n%s", test.description, diff)
			}
		})
	}
}

func readTestFile(filePath string) string {
	data, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	return string(bytes.TrimSpace(data))
}
