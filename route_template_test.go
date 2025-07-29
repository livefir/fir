package fir

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRoute_GetErrorTemplate tests the getErrorTemplate method
func TestRoute_GetErrorTemplate(t *testing.T) {
	tests := []struct {
		name           string
		setupRoute     func() *route
		expectedResult *template.Template
	}{
		{
			name: "returns nil when no error template is set",
			setupRoute: func() *route {
				return &route{
					errorTemplate: nil,
					RWMutex:       sync.RWMutex{},
				}
			},
			expectedResult: nil,
		},
		{
			name: "returns error template when set",
			setupRoute: func() *route {
				tmpl := template.Must(template.New("error").Parse("<div>Error: {{.Message}}</div>"))
				return &route{
					errorTemplate: tmpl,
					RWMutex:       sync.RWMutex{},
				}
			},
			expectedResult: func() *template.Template {
				return template.Must(template.New("error").Parse("<div>Error: {{.Message}}</div>"))
			}(),
		},
		{
			name: "returns complex error template",
			setupRoute: func() *route {
				tmpl := template.Must(template.New("complex_error").Parse(`
					<html>
						<head><title>Error</title></head>
						<body>
							<h1>An error occurred</h1>
							<p>{{.ErrorMessage}}</p>
							<p>Status: {{.StatusCode}}</p>
						</body>
					</html>
				`))
				return &route{
					errorTemplate: tmpl,
					RWMutex:       sync.RWMutex{},
				}
			},
			expectedResult: func() *template.Template {
				return template.Must(template.New("complex_error").Parse(`
					<html>
						<head><title>Error</title></head>
						<body>
							<h1>An error occurred</h1>
							<p>{{.ErrorMessage}}</p>
							<p>Status: {{.StatusCode}}</p>
						</body>
					</html>
				`))
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := tt.setupRoute()
			result := route.getErrorTemplate()

			if tt.expectedResult == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedResult.Name(), result.Name())
				// Verify template functionality by comparing their text
				expectedText, _ := tt.expectedResult.Tree.Root.String(), ""
				resultText, _ := result.Tree.Root.String(), ""
				assert.Equal(t, expectedText, resultText)
			}
		})
	}
}

// TestRoute_SetAndGetErrorTemplate tests the setErrorTemplate and getErrorTemplate methods together
func TestRoute_SetAndGetErrorTemplate(t *testing.T) {
	tests := []struct {
		name         string
		templateName string
		templateText string
	}{
		{
			name:         "simple error template",
			templateName: "simple_error",
			templateText: "<div>Error: {{.Message}}</div>",
		},
		{
			name:         "error template with conditional logic",
			templateName: "conditional_error",
			templateText: `{{if .IsAuthenticated}}<p>Authenticated error</p>{{else}}<p>Please login</p>{{end}}`,
		},
		{
			name:         "error template with range",
			templateName: "list_error",
			templateText: `<ul>{{range .Errors}}<li>{{.}}</li>{{end}}</ul>`,
		},
		{
			name:         "empty template",
			templateName: "empty",
			templateText: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a route
			route := &route{
				RWMutex: sync.RWMutex{},
			}

			// Create and set a template
			tmpl := template.Must(template.New(tt.templateName).Parse(tt.templateText))
			route.setErrorTemplate(tmpl)

			// Get the template back
			result := route.getErrorTemplate()

			// Verify the template
			assert.NotNil(t, result)
			assert.Equal(t, tt.templateName, result.Name())

			// Verify template content by comparing tree structure
			expectedText := tmpl.Tree.Root.String()
			resultText := result.Tree.Root.String()
			assert.Equal(t, expectedText, resultText)
		})
	}
}

// TestRoute_ErrorTemplateNilSafety tests nil safety of error template methods
func TestRoute_ErrorTemplateNilSafety(t *testing.T) {
	route := &route{
		RWMutex: sync.RWMutex{},
	}

	// Initially should return nil
	result := route.getErrorTemplate()
	assert.Nil(t, result)

	// Set to nil explicitly
	route.setErrorTemplate(nil)
	result = route.getErrorTemplate()
	assert.Nil(t, result)

	// Set a template then set back to nil
	tmpl := template.Must(template.New("test").Parse("<div>Test</div>"))
	route.setErrorTemplate(tmpl)
	result = route.getErrorTemplate()
	assert.NotNil(t, result)

	route.setErrorTemplate(nil)
	result = route.getErrorTemplate()
	assert.Nil(t, result)
}

// TestRoute_ErrorTemplateConcurrency tests concurrent access to error template methods
func TestRoute_ErrorTemplateConcurrency(t *testing.T) {
	route := &route{
		RWMutex: sync.RWMutex{},
	}

	// Number of goroutines for concurrent testing
	numGoroutines := 10

	// Create templates for setting
	templates := make([]*template.Template, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		tmpl := template.Must(template.New(fmt.Sprintf("error_%d", i)).Parse(fmt.Sprintf("<div>Error %d: {{.Message}}</div>", i)))
		templates[i] = tmpl
	}

	// Test concurrent setting and getting
	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2) // Both setters and getters

	// Concurrent setters
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()
			route.setErrorTemplate(templates[index])
		}(i)
	}

	// Concurrent getters
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			result := route.getErrorTemplate()
			// Should not panic and should return either nil or a valid template
			if result != nil {
				assert.NotEmpty(t, result.Name())
			}
		}()
	}

	wg.Wait()

	// Final verification - should have one of the templates set
	finalResult := route.getErrorTemplate()
	if finalResult != nil {
		assert.NotEmpty(t, finalResult.Name())
		assert.Contains(t, finalResult.Name(), "error_")
	}
}

// TestRoute_ErrorTemplateOverwrite tests overwriting error templates
func TestRoute_ErrorTemplateOverwrite(t *testing.T) {
	route := &route{
		RWMutex: sync.RWMutex{},
	}

	// Set first template
	tmpl1 := template.Must(template.New("error1").Parse("<div>First error template</div>"))
	route.setErrorTemplate(tmpl1)

	result1 := route.getErrorTemplate()
	assert.NotNil(t, result1)
	assert.Equal(t, "error1", result1.Name())

	// Overwrite with second template
	tmpl2 := template.Must(template.New("error2").Parse("<div>Second error template</div>"))
	route.setErrorTemplate(tmpl2)

	result2 := route.getErrorTemplate()
	assert.NotNil(t, result2)
	assert.Equal(t, "error2", result2.Name())
	assert.NotEqual(t, result1.Name(), result2.Name())

	// Verify the content changed
	assert.NotEqual(t, result1.Tree.Root.String(), result2.Tree.Root.String())
}

// TestRoute_ErrorTemplateComplexScenarios tests complex template scenarios
func TestRoute_ErrorTemplateComplexScenarios(t *testing.T) {
	tests := []struct {
		name         string
		templateText string
		shouldPanic  bool
	}{
		{
			name:         "template with functions",
			templateText: `<div>{{printf "Error: %s" .Message}}</div>`,
			shouldPanic:  false,
		},
		{
			name:         "template with blocks",
			templateText: `{{define "error_block"}}<span>{{.}}</span>{{end}}{{template "error_block" .Message}}`,
			shouldPanic:  false,
		},
		{
			name:         "template with complex nesting",
			templateText: `{{range .Errors}}{{if .Critical}}<strong>{{.Message}}</strong>{{else}}<em>{{.Message}}</em>{{end}}{{end}}`,
			shouldPanic:  false,
		},
		{
			name:         "invalid template syntax should panic during creation",
			templateText: `{{invalid syntax`,
			shouldPanic:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := &route{
				RWMutex: sync.RWMutex{},
			}

			if tt.shouldPanic {
				assert.Panics(t, func() {
					tmpl := template.Must(template.New("test").Parse(tt.templateText))
					route.setErrorTemplate(tmpl)
				})
			} else {
				assert.NotPanics(t, func() {
					tmpl := template.Must(template.New("test").Parse(tt.templateText))
					route.setErrorTemplate(tmpl)

					result := route.getErrorTemplate()
					assert.NotNil(t, result)
					assert.Equal(t, "test", result.Name())
				})
			}
		})
	}
}

// TestRoute_ErrorTemplateExecution tests that retrieved templates can be executed
func TestRoute_ErrorTemplateExecution(t *testing.T) {
	route := &route{
		RWMutex: sync.RWMutex{},
	}

	// Create and set a template
	tmpl := template.Must(template.New("executable_error").Parse("<div>Error: {{.Message}}</div>"))
	route.setErrorTemplate(tmpl)

	// Get the template
	result := route.getErrorTemplate()
	assert.NotNil(t, result)

	// Test template execution
	var buf strings.Builder
	data := map[string]interface{}{
		"Message": "Test error message",
	}

	err := result.Execute(&buf, data)
	assert.NoError(t, err)
	assert.Equal(t, "<div>Error: Test error message</div>", buf.String())
}

// TestRoute_ErrorTemplateIntegration tests error template methods with real route creation
func TestRoute_ErrorTemplateIntegration(t *testing.T) {
	// Create a controller
	controller := NewController("test-controller")

	// Create a RouteFunc that uses error templates
	routeFunc := func() RouteOptions {
		return RouteOptions{
			ID("test-route"),
			Content("<div>Content</div>"),
			ErrorContent("<div>Error: {{.Message}}</div>"),
		}
	}

	// Test that we can get a handler from this
	handler := controller.RouteFunc(routeFunc)
	assert.NotNil(t, handler)

	// The route should be created internally by the controller
	// We can't directly access it, but we can test that the setup works
	assert.IsType(t, handler, http.HandlerFunc(nil))
}
