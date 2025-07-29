package fir

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_resolveTemplatePath_AbsolutePaths(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "test_template*.html")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	absolutePath := tmpFile.Name()

	// Test with existing absolute path
	resolvedPath, isValidFile := resolveTemplatePath(absolutePath, 2)
	require.Equal(t, absolutePath, resolvedPath)
	require.True(t, isValidFile)

	// Test with non-existing absolute path
	nonExistentPath := "/non/existent/path.html"
	resolvedPath, isValidFile = resolveTemplatePath(nonExistentPath, 2)
	require.Equal(t, nonExistentPath, resolvedPath)
	require.False(t, isValidFile)
}

func Test_resolveTemplatePath_InlineContent(t *testing.T) {
	// Test with inline HTML content (not a file path)
	inlineHTML := "<div>Hello {{.name}}</div>"
	resolvedPath, isValidFile := resolveTemplatePath(inlineHTML, 2)
	require.Equal(t, inlineHTML, resolvedPath)
	require.False(t, isValidFile)
}

func Test_resolveTemplatePath_NonExistentRelativePath(t *testing.T) {
	// Test relative path that doesn't exist
	resolvedPath, isValidFile := resolveTemplatePath("nonexistent.html", 2)
	require.Equal(t, "nonexistent.html", resolvedPath)
	require.False(t, isValidFile)
}

func Test_resolveTemplatePath_CallerDepthVariations(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "depth_test*.html")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	absolutePath := tmpFile.Name()

	// Test with different caller depths
	for depth := 0; depth <= 5; depth++ {
		resolvedPath, isValidFile := resolveTemplatePath(absolutePath, depth)
		require.Equal(t, absolutePath, resolvedPath, "Should work with caller depth %d", depth)
		require.True(t, isValidFile, "Should be valid file with caller depth %d", depth)
	}
}

func Test_resolveTemplatePath_PathTypes(t *testing.T) {
	// Test with various path formats that don't require file system setup
	testCases := []struct {
		name     string
		input    string
		expected string
		isValid  bool
	}{
		{
			name:     "whitespace around non-existent path",
			input:    "  template.html  ",
			expected: "  template.html  ",
			isValid:  false,
		},
		{
			name:     "inline HTML content",
			input:    "<div>test</div>",
			expected: "<div>test</div>",
			isValid:  false,
		},
		{
			name:     "simple template name",
			input:    "template.html",
			expected: "template.html",
			isValid:  false,
		},
		{
			name:     "template with go syntax",
			input:    "{{.user}}",
			expected: "{{.user}}",
			isValid:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resolvedPath, isValidFile := resolveTemplatePath(tc.input, 2)
			require.Equal(t, tc.expected, resolvedPath)
			require.Equal(t, tc.isValid, isValidFile)
		})
	}
}

func Test_resolveTemplatePath_WithExistingFile(t *testing.T) {
	// Simple test that doesn't change working directory
	// Create a temporary file with absolute path
	tmpFile, err := os.CreateTemp("", "test_resolve*.html")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Test with absolute path (always works)
	resolvedPath, isValidFile := resolveTemplatePath(tmpFile.Name(), 2)
	require.True(t, isValidFile)
	require.Equal(t, tmpFile.Name(), resolvedPath)
}

func Test_resolveTemplatePath_NestedPaths(t *testing.T) {
	// Test with a simpler approach using absolute paths
	tmpDir, err := os.MkdirTemp("", "fir_nested_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create nested directories
	nestedDir := filepath.Join(tmpDir, "templates", "partials")
	err = os.MkdirAll(nestedDir, 0755)
	require.NoError(t, err)

	// Create file in nested directory
	nestedFile := filepath.Join(nestedDir, "nested.html")
	err = os.WriteFile(nestedFile, []byte("<div></div>"), 0644)
	require.NoError(t, err)

	// Test with absolute path to nested file (always works)
	resolvedPath, isValidFile := resolveTemplatePath(nestedFile, 2)
	require.True(t, isValidFile)
	require.Equal(t, nestedFile, resolvedPath)
}

func Test_resolveTemplatePath_EmptyString(t *testing.T) {
	// Test with empty string - this will resolve to caller's directory
	resolvedPath, isValidFile := resolveTemplatePath("", 2)
	// Empty string results in some resolved path (caller directory)
	require.NotEmpty(t, resolvedPath)
	// The resolved directory path will exist, so isValidFile will be true
	require.True(t, isValidFile)
}

func Test_fileExists(t *testing.T) {
	// Test with existing file
	tmpFile, err := os.CreateTemp("", "exists_test*.html")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	require.True(t, fileExists(tmpFile.Name()))

	// Test with non-existing file
	require.False(t, fileExists("/path/that/does/not/exist"))

	// Test with directory
	tmpDir, err := os.MkdirTemp("", "exists_dir_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	require.True(t, fileExists(tmpDir))

	// Test with empty path
	require.False(t, fileExists(""))
}

func Test_resolveTemplatePath_RuntimeCallerHandling(t *testing.T) {
	// Test the function's runtime.Caller handling with different depths
	// This tests the code paths that handle caller depth resolution

	// Create a temp file to test with
	tmpFile, err := os.CreateTemp("", "caller_test*.html")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Test with absolute path (doesn't depend on caller resolution)
	absolutePath := tmpFile.Name()
	resolvedPath, isValidFile := resolveTemplatePath(absolutePath, 10) // high depth
	require.Equal(t, absolutePath, resolvedPath)
	require.True(t, isValidFile)

	// Test with non-file paths
	resolvedPath, isValidFile = resolveTemplatePath("{{template}}", 10)
	require.Equal(t, "{{template}}", resolvedPath)
	require.False(t, isValidFile)
}
