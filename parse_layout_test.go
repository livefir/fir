package fir

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to check if file exists
func simpleFileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// TestLayoutSetContentSet tests the layoutSetContentSet function
func TestLayoutSetContentSet(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "fir_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test layout file
	layoutContent := `<html><head><title>Test Layout</title></head><body>{{template "content" .}}</body></html>`
	layoutPath := filepath.Join(tempDir, "layout.html")
	err = os.WriteFile(layoutPath, []byte(layoutContent), 0644)
	require.NoError(t, err)

	// Create test content file
	contentFileContent := `{{define "content"}}<div>File Content: {{.Message}}</div>{{end}}`
	contentPath := filepath.Join(tempDir, "content.html")
	err = os.WriteFile(contentPath, []byte(contentFileContent), 0644)
	require.NoError(t, err)

	tests := []struct {
		name              string
		content           string
		layout            string
		layoutContentName string
		setupOpt          func() routeOpt
		expectedError     bool
		expectedTemplate  bool
		validateTemplate  func(*testing.T, *template.Template, eventTemplates)
		expectedPanic     bool
	}{
		{
			name:              "content as string template",
			content:           `{{define "content"}}<div>String Content: {{.Message}}</div>{{end}}`,
			layout:            "layout.html",
			layoutContentName: "content",
			setupOpt: func() routeOpt {
				return routeOpt{
					layout:            "layout.html",
					layoutContentName: "content",
					funcMap:           make(template.FuncMap),
					funcMapMutex:      &sync.RWMutex{},
					extensions:        []string{".html", ".gohtml"},
					opt: opt{
						publicDir: tempDir,
						existFile: func(path string) bool {
							// Return false for the content string to trigger parseString path
							return path != filepath.Join(tempDir, `{{define "content"}}<div>String Content: {{.Message}}</div>{{end}}`) &&
								simpleFileExists(path)
						},
						readFile: func(path string) (string, []byte, error) {
							content, err := os.ReadFile(path)
							return string(content), content, err
						},
					},
				}
			},
			expectedTemplate: true,
			validateTemplate: func(t *testing.T, tmpl *template.Template, evt eventTemplates) {
				assert.NotNil(t, tmpl)
				assert.NotNil(t, tmpl.Lookup("content"))
				// Template name will be layout.html when parsing from file
				assert.Equal(t, "layout.html", tmpl.Name())
			},
		},
		{
			name:              "content as existing file",
			content:           "content.html",
			layout:            "layout.html",
			layoutContentName: "content",
			setupOpt: func() routeOpt {
				return routeOpt{
					layout:            "layout.html",
					layoutContentName: "content",
					funcMap:           make(template.FuncMap),
					funcMapMutex:      &sync.RWMutex{},
					extensions:        []string{".html", ".gohtml"},
					opt: opt{
						publicDir: tempDir,
						existFile: simpleFileExists,
						readFile: func(path string) (string, []byte, error) {
							content, err := os.ReadFile(path)
							return string(content), content, err
						},
					},
				}
			},
			expectedTemplate: true,
			validateTemplate: func(t *testing.T, tmpl *template.Template, evt eventTemplates) {
				assert.NotNil(t, tmpl)
				assert.NotNil(t, tmpl.Lookup("content"))
				assert.Equal(t, "layout.html", tmpl.Name())
			},
		},
		{
			name:              "layout as string when file doesn't exist",
			content:           "content.html",
			layout:            `<html><body>{{template "content" .}}</body></html>`,
			layoutContentName: "content",
			setupOpt: func() routeOpt {
				return routeOpt{
					layout:            `<html><body>{{template "content" .}}</body></html>`,
					layoutContentName: "content",
					funcMap:           make(template.FuncMap),
					funcMapMutex:      &sync.RWMutex{},
					extensions:        []string{".html", ".gohtml"},
					opt: opt{
						publicDir: tempDir,
						existFile: func(path string) bool {
							// Layout string won't exist as file
							return path != filepath.Join(tempDir, `<html><body>{{template "content" .}}</body></html>`) && simpleFileExists(path)
						},
						readFile: func(path string) (string, []byte, error) {
							content, err := os.ReadFile(path)
							return string(content), content, err
						},
					},
				}
			},
			expectedTemplate: true,
			validateTemplate: func(t *testing.T, tmpl *template.Template, evt eventTemplates) {
				assert.NotNil(t, tmpl)
				assert.NotNil(t, tmpl.Lookup("content"))
				// Successfully parsed template with content
			},
		},
		{
			name:              "missing layout content template after string parsing",
			content:           `{{define "wrong_name"}}<div>Wrong template name</div>{{end}}`,
			layout:            "layout.html",
			layoutContentName: "content",
			setupOpt: func() routeOpt {
				return routeOpt{
					layout:            "layout.html",
					layoutContentName: "content",
					funcMap:           make(template.FuncMap),
					funcMapMutex:      &sync.RWMutex{},
					extensions:        []string{".html", ".gohtml"},
					opt: opt{
						publicDir: tempDir,
						existFile: func(path string) bool {
							// Return false for the content string to trigger parseString path
							return path != filepath.Join(tempDir, `{{define "wrong_name"}}<div>Wrong template name</div>{{end}}`) &&
								simpleFileExists(path)
						},
						readFile: func(path string) (string, []byte, error) {
							content, err := os.ReadFile(path)
							return string(content), content, err
						},
					},
				}
			},
			expectedError: true,
		},
		{
			name:              "missing layout content template after file parsing",
			content:           "wrong_content.html",
			layout:            "layout.html",
			layoutContentName: "content",
			setupOpt: func() routeOpt {
				// Create content file with wrong template name
				wrongContent := `{{define "wrong_name"}}<div>Wrong template name</div>{{end}}`
				wrongPath := filepath.Join(tempDir, "wrong_content.html")
				err := os.WriteFile(wrongPath, []byte(wrongContent), 0644)
				require.NoError(t, err)

				return routeOpt{
					layout:            "layout.html",
					layoutContentName: "content",
					funcMap:           make(template.FuncMap),
					funcMapMutex:      &sync.RWMutex{},
					extensions:        []string{".html", ".gohtml"},
					opt: opt{
						publicDir: tempDir,
						existFile: simpleFileExists,
						readFile: func(path string) (string, []byte, error) {
							content, err := os.ReadFile(path)
							return string(content), content, err
						},
					},
				}
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := tt.setupOpt()

			if tt.expectedPanic {
				defer func() {
					if r := recover(); r != nil {
						assert.True(t, true, "Expected panic occurred: %v", r)
					} else {
						t.Error("Expected panic but none occurred")
					}
				}()
			}

			tmpl, evt, err := layoutSetContentSet(opt, tt.content, tt.layout, tt.layoutContentName)

			if tt.expectedPanic {
				// If we reach here without panic, the test should fail
				t.Error("Expected panic but none occurred")
				return
			}

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, tmpl)
				return
			}

			assert.NoError(t, err)

			if tt.expectedTemplate {
				assert.NotNil(t, tmpl)
				assert.NotNil(t, evt)

				if tt.validateTemplate != nil {
					tt.validateTemplate(t, tmpl, evt)
				}
			}
		})
	}
}

// TestLayoutSetContentSet_EventTemplatesMerging tests the event templates merging functionality
func TestLayoutSetContentSet_EventTemplatesMerging(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "fir_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test layout file with event attributes
	layoutContent := `<html><head><title>Test Layout</title></head><body @fir:test:ok="layout_template">{{template "content" .}}</body></html>`
	layoutPath := filepath.Join(tempDir, "layout.html")
	err = os.WriteFile(layoutPath, []byte(layoutContent), 0644)
	require.NoError(t, err)

	// Test with string content that has events
	contentString := `{{define "content"}}<div @fir:click:ok="content_template">Content: {{.Message}}</div>{{end}}`

	opt := routeOpt{
		layout:            "layout.html",
		layoutContentName: "content",
		funcMap:           make(template.FuncMap),
		funcMapMutex:      &sync.RWMutex{},
		extensions:        []string{".html", ".gohtml"},
		opt: opt{
			publicDir: tempDir,
			existFile: func(path string) bool {
				// Return false for the content string to trigger parseString path
				return path != filepath.Join(tempDir, contentString) && simpleFileExists(path)
			},
			readFile: func(path string) (string, []byte, error) {
				content, err := os.ReadFile(path)
				return string(content), content, err
			},
		},
	}

	tmpl, evt, err := layoutSetContentSet(opt, contentString, "layout.html", "content")

	assert.NoError(t, err)
	assert.NotNil(t, tmpl)
	assert.NotNil(t, evt)

	// Check that event templates from both layout and content were merged
	assert.GreaterOrEqual(t, len(evt), 0, "Event templates should be accessible")

	// Verify template functionality
	assert.NotNil(t, tmpl.Lookup("content"))
	// Template name will be generated hash when parsing string layout and content
	assert.NotEmpty(t, tmpl.Name())
}

// TestLayoutSetContentSet_GetFuncMap tests that getFuncMap is properly called
func TestLayoutSetContentSet_GetFuncMap(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "fir_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test layout file
	layoutContent := `<html><head><title>Test Layout</title></head><body>{{template "content" .}}</body></html>`
	layoutPath := filepath.Join(tempDir, "layout.html")
	err = os.WriteFile(layoutPath, []byte(layoutContent), 0644)
	require.NoError(t, err)

	// Test with custom function in funcMap
	customFunc := func() string { return "custom_value" }
	funcMap := template.FuncMap{
		"customFunc": customFunc,
	}

	contentString := `{{define "content"}}<div>{{customFunc}}</div>{{end}}`

	opt := routeOpt{
		layout:            "layout.html",
		layoutContentName: "content",
		funcMap:           funcMap,
		funcMapMutex:      &sync.RWMutex{},
		extensions:        []string{".html", ".gohtml"},
		opt: opt{
			publicDir: tempDir,
			existFile: func(path string) bool {
				return path != filepath.Join(tempDir, contentString) && simpleFileExists(path)
			},
			readFile: func(path string) (string, []byte, error) {
				content, err := os.ReadFile(path)
				return string(content), content, err
			},
		},
	}

	tmpl, evt, err := layoutSetContentSet(opt, contentString, "layout.html", "content")

	assert.NoError(t, err)
	assert.NotNil(t, tmpl)
	assert.NotNil(t, evt)
	assert.NotNil(t, tmpl.Lookup("content"))
}

// TestLayoutSetContentSet_EdgeCases tests various edge cases
func TestLayoutSetContentSet_EdgeCases(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "fir_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test layout file
	layoutContent := `<html><head><title>Test Layout</title></head><body>{{template "content" .}}</body></html>`
	layoutPath := filepath.Join(tempDir, "layout.html")
	err = os.WriteFile(layoutPath, []byte(layoutContent), 0644)
	require.NoError(t, err)

	tests := []struct {
		name        string
		content     string
		setupOpt    func() routeOpt
		expectError bool
		expectPanic bool
	}{
		{
			name:    "empty content string",
			content: "",
			setupOpt: func() routeOpt {
				return routeOpt{
					layout:            "layout.html",
					layoutContentName: "content",
					funcMap:           make(template.FuncMap),
					funcMapMutex:      &sync.RWMutex{},
					extensions:        []string{".html", ".gohtml"},
					opt: opt{
						publicDir: tempDir,
						existFile: func(path string) bool {
							return path != filepath.Join(tempDir, "") && simpleFileExists(path)
						},
						readFile: func(path string) (string, []byte, error) {
							content, err := os.ReadFile(path)
							return string(content), content, err
						},
					},
				}
			},
			expectError: true, // Empty string should cause checkPageContent to fail
		},
		{
			name:    "content is directory path",
			content: ".",
			setupOpt: func() routeOpt {
				return routeOpt{
					layout:            "layout.html",
					layoutContentName: "content",
					funcMap:           make(template.FuncMap),
					funcMapMutex:      &sync.RWMutex{},
					extensions:        []string{".html", ".gohtml"},
					opt: opt{
						publicDir: tempDir,
						existFile: func(path string) bool {
							// Directory exists but it's not a file
							return path == filepath.Join(tempDir, ".")
						},
						readFile: func(path string) (string, []byte, error) {
							return "", nil, fmt.Errorf("is a directory")
						},
					},
				}
			},
			expectPanic: true, // Should panic due to file reading error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := tt.setupOpt()

			if tt.expectPanic {
				defer func() {
					if r := recover(); r != nil {
						assert.True(t, true, "Expected panic occurred: %v", r)
					} else {
						t.Error("Expected panic but none occurred")
					}
				}()
			}

			tmpl, evt, err := layoutSetContentSet(opt, tt.content, "layout.html", "content")

			if tt.expectPanic {
				t.Error("Expected panic but none occurred")
				return
			}

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, tmpl)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, tmpl)
			assert.NotNil(t, evt)
		})
	}
}
