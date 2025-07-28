package fir

import (
	"github.com/livefir/fir/internal/file"
	internalMarkdown "github.com/livefir/fir/internal/markdown"
)

// Markdown creates a markdown processing function for use in templates.
// This is a public API function that provides backward compatibility wrapper for the internal markdown functionality.
func Markdown(readFile file.ReadFileFunc, existFile file.ExistFileFunc) func(in string, markers ...string) string {
	return internalMarkdown.New(readFile, existFile)
}
