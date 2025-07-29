package fir

import (
	"html/template"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCheckPageContent tests the checkPageContent function
func TestCheckPageContent(t *testing.T) {
	t.Run("missing layout content", func(t *testing.T) {
		// Create a template without the expected content placeholder
		tmpl := template.Must(template.New("layout").Parse(`<html><body>No content placeholder</body></html>`))

		err := checkPageContent(tmpl, "content")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expects a template named content")
	})

	t.Run("with layout content", func(t *testing.T) {
		// Create a template with the expected content placeholder
		tmpl := template.Must(template.New("layout").Parse(`<html><body>{{template "content" .}}</body></html>`))
		// Add the content template
		tmpl = template.Must(tmpl.New("content").Parse(`<p>Content here</p>`))

		err := checkPageContent(tmpl, "content")
		assert.NoError(t, err)
	})
}
