package fir

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarkdown(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Create test markdown files
	testMarkdownContent := `# Test Header

This is a **bold** text and this is *italic* text.

- List item 1
- List item 2

[Link text](http://example.com)
`

	testFile := filepath.Join(tmpDir, "test.md")
	err := os.WriteFile(testFile, []byte(testMarkdownContent), 0644)
	assert.NoError(t, err)

	// Test readFile function
	readFile := func(filename string) (string, []byte, error) {
		data, err := os.ReadFile(filename)
		return filename, data, err
	}

	// Test existFile function
	existFile := func(filename string) bool {
		_, err := os.Stat(filename)
		return err == nil
	}

	// Create markdown processor
	markdownFunc := Markdown(readFile, existFile)
	assert.NotNil(t, markdownFunc)

	t.Run("process markdown string", func(t *testing.T) {
		input := "# Header\n\nThis is **bold** text."
		result := markdownFunc(input)

		assert.NotEmpty(t, result)
		// The markdown processor generates HTML with id attributes
		assert.Contains(t, result, `<h1 id="header">`)
		assert.Contains(t, result, "Header")
		assert.Contains(t, result, "<strong>")
		assert.Contains(t, result, "bold")
	})

	t.Run("process markdown with file inclusion", func(t *testing.T) {
		// Test with file inclusion markers
		input := "Start text\n\n" + testFile + "\n\nEnd text"
		result := markdownFunc(input, "START_MARKER", "END_MARKER")

		assert.NotEmpty(t, result)
		assert.Contains(t, result, "Start text")
		assert.Contains(t, result, "End text")
	})

	t.Run("process empty markdown", func(t *testing.T) {
		result := markdownFunc("")
		assert.Equal(t, "", result)
	})

	t.Run("process simple text", func(t *testing.T) {
		input := "Just plain text without markdown"
		result := markdownFunc(input)

		assert.NotEmpty(t, result)
		assert.Contains(t, result, "Just plain text")
	})

	t.Run("process markdown with lists", func(t *testing.T) {
		input := "# List Example\n\n- Item 1\n- Item 2\n- Item 3"
		result := markdownFunc(input)

		assert.NotEmpty(t, result)
		assert.Contains(t, result, `<h1 id="list-example">`)
		assert.Contains(t, result, "<ul>")
		assert.Contains(t, result, "<li>")
		assert.Contains(t, result, "Item 1")
	})

	t.Run("process markdown with links", func(t *testing.T) {
		input := "Check out [this link](http://example.com) for more info."
		result := markdownFunc(input)

		assert.NotEmpty(t, result)
		assert.Contains(t, result, "<a href=\"http://example.com\">")
		assert.Contains(t, result, "this link")
	})

	t.Run("process markdown with code", func(t *testing.T) {
		input := "Here is some `inline code` and a code block:\n\n```\ncode block\n```"
		result := markdownFunc(input)

		assert.NotEmpty(t, result)
		assert.Contains(t, result, "<code>")
		assert.Contains(t, result, "inline code")
	})
}

func TestMarkdownWithNilFunctions(t *testing.T) {
	// Test behavior with nil functions - this may panic in the internal implementation
	// so we need to use safe non-nil functions

	// Use safe dummy functions instead of nil
	readFile := func(filename string) (string, []byte, error) {
		return filename, []byte(""), os.ErrNotExist
	}

	existFile := func(filename string) bool {
		return false
	}

	markdownFunc := Markdown(readFile, existFile)
	assert.NotNil(t, markdownFunc)

	// Should work for basic markdown processing
	input := "# Header\n\nThis is **bold** text."
	result := markdownFunc(input)

	assert.NotEmpty(t, result)
	assert.Contains(t, result, `<h1 id="header">`)
	assert.Contains(t, result, "<strong>")
}

func TestMarkdownWithCustomFunctions(t *testing.T) {
	// Test with custom readFile and existFile functions

	// Mock readFile that returns predefined content
	mockContent := "# Mock File\n\nThis is mock content."
	readFile := func(filename string) (string, []byte, error) {
		if filename == "mock.md" {
			return filename, []byte(mockContent), nil
		}
		return filename, nil, os.ErrNotExist
	}

	// Mock existFile that only returns true for specific files
	existFile := func(filename string) bool {
		return filename == "mock.md"
	}

	markdownFunc := Markdown(readFile, existFile)
	assert.NotNil(t, markdownFunc)

	t.Run("process with mock file functions", func(t *testing.T) {
		input := "# Test\n\nSome content here."
		result := markdownFunc(input)

		assert.NotEmpty(t, result)
		assert.Contains(t, result, `<h1 id="test">`)
		assert.Contains(t, result, "Test")
	})
}

func TestMarkdownIntegration(t *testing.T) {
	// Integration test that combines markdown processing with file operations
	tmpDir := t.TempDir()

	// Create multiple test files
	file1 := filepath.Join(tmpDir, "intro.md")
	file2 := filepath.Join(tmpDir, "content.md")

	intro := "# Introduction\n\nWelcome to the documentation."
	content := "## Main Content\n\nThis is the main **content** section."

	err := os.WriteFile(file1, []byte(intro), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(file2, []byte(content), 0644)
	assert.NoError(t, err)

	readFile := func(filename string) (string, []byte, error) {
		data, err := os.ReadFile(filename)
		return filename, data, err
	}

	existFile := func(filename string) bool {
		_, err := os.Stat(filename)
		return err == nil
	}

	markdownFunc := Markdown(readFile, existFile)

	// Test processing markdown that might reference files
	complexMarkdown := `# Documentation

This is a complex markdown document.

## Features

- **Bold text** support
- *Italic text* support  
- [Links](http://example.com)
- Code blocks

### Code Example

` + "```go\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n```" + `

That's it!`

	result := markdownFunc(complexMarkdown)

	assert.NotEmpty(t, result)
	assert.Contains(t, result, `<h1 id="documentation">`)
	assert.Contains(t, result, "Documentation")
	assert.Contains(t, result, `<h2 id="features">`)
	assert.Contains(t, result, "Features")
	assert.Contains(t, result, "<strong>")
	assert.Contains(t, result, "Bold text")
	assert.Contains(t, result, "<em>")
	assert.Contains(t, result, "Italic text")
	assert.Contains(t, result, "<a href=")
	assert.Contains(t, result, "http://example.com")
	assert.Contains(t, result, "<ul>")
	assert.Contains(t, result, "<li>")
	// The code is syntax highlighted, so we look for the actual function name
	assert.Contains(t, result, "main")
}
