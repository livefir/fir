package fir

import (
	"embed"
	"testing"
)

//go:embed testdata/public
var testdata embed.FS

func TestIsDir(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		embedfs  *embed.FS
		expected bool
	}{
		{
			name:     "Embedded Directory",
			path:     "testdata/public",
			embedfs:  &testdata,
			expected: true,
		},
		{
			name:     "Non-Embedded Directory",
			path:     "testdata/public",
			embedfs:  nil,
			expected: true,
		},
		{
			name:     "Non-Existing Directory",
			path:     "/path/to/non-existing/directory",
			embedfs:  nil,
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := isDir(test.path, test.embedfs)
			if actual != test.expected {
				t.Errorf("isDir(%s, %v) = %v, expected %v", test.path, test.embedfs, actual, test.expected)
			}
		})
	}
}

func TestFind(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		extensions []string
		embedfs    *embed.FS
		expected   []string
	}{
		{
			name:       "Embedded HTML Files",
			path:       "testdata/public",
			extensions: []string{".html"},
			embedfs:    &testdata,
			expected:   []string{"testdata/public/index.html", "testdata/public/partials/header.html"},
		},
		{
			name:       "Non-Embedded HTML Files",
			path:       "testdata/public",
			extensions: []string{".html"},
			embedfs:    nil,
			expected:   []string{"testdata/public/index.html", "testdata/public/partials/header.html"},
		},
		{
			name:       "Embedded CSS Files don't exist",
			path:       "testdata/public",
			extensions: []string{".css"},
			embedfs:    &testdata,
			expected:   []string{},
		},
		{
			name:       "Embedded CSS Files don't exist",
			path:       "testdata/public",
			extensions: []string{".css"},
			embedfs:    nil,
			expected:   []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := find(test.path, test.extensions, test.embedfs)
			if len(actual) != len(test.expected) {
				t.Errorf("find(%s, %v, %v) returned %v files, expected %v files", test.path, test.extensions, test.embedfs, len(actual), len(test.expected))
			} else {
				for i := range actual {
					if actual[i] != test.expected[i] {
						t.Errorf("find(%s, %v, %v) returned %s, expected %s", test.path, test.extensions, test.embedfs, actual[i], test.expected[i])
					}
				}
			}
		})
	}
}
