package file

import (
	"embed"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
		{
			name:     "Embedded File (not directory)",
			path:     "testdata/public/index.html",
			embedfs:  &testdata,
			expected: false,
		},
		{
			name:     "Non-Embedded File (not directory)",
			path:     "testdata/public/index.html",
			embedfs:  nil,
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Use IsDirWithExistFile with appropriate existFile function
			var existFile ExistFileFunc
			if test.embedfs != nil {
				existFile = ExistFileFS(*test.embedfs)
			} else {
				existFile = ExistFileOS
			}

			actual := IsDirWithExistFile(test.path, existFile, test.embedfs)
			if actual != test.expected {
				t.Errorf("IsDirWithExistFile(%s, existFunc, %v) = %v, expected %v", test.path, test.embedfs, actual, test.expected)
			}
		})
	}
}

func TestIsDirWithExistFile(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		embedfs   *embed.FS
		existFile ExistFileFunc
		expected  bool
	}{
		{
			name:      "Embedded Directory - exists",
			path:      "testdata/public",
			embedfs:   &testdata,
			existFile: ExistFileFS(testdata),
			expected:  true,
		},
		{
			name:      "Embedded Directory - does not exist",
			path:      "testdata/nonexistent",
			embedfs:   &testdata,
			existFile: ExistFileFS(testdata),
			expected:  false,
		},
		{
			name:      "Embedded File - exists but not directory",
			path:      "testdata/public/index.html",
			embedfs:   &testdata,
			existFile: ExistFileFS(testdata),
			expected:  false,
		},
		{
			name:      "Embedded File - does not exist",
			path:      "testdata/public/nonexistent.html",
			embedfs:   &testdata,
			existFile: ExistFileFS(testdata),
			expected:  false,
		},
		{
			name:      "OS Directory - exists",
			path:      "testdata/public",
			embedfs:   nil,
			existFile: ExistFileOS,
			expected:  true,
		},
		{
			name:      "OS Directory - does not exist",
			path:      "/path/to/non-existing/directory",
			embedfs:   nil,
			existFile: ExistFileOS,
			expected:  false,
		},
		{
			name:      "OS File - exists but not directory",
			path:      "testdata/public/index.html",
			embedfs:   nil,
			existFile: ExistFileOS,
			expected:  false,
		},
		{
			name:      "OS File - does not exist",
			path:      "/path/to/nonexistent.html",
			embedfs:   nil,
			existFile: ExistFileOS,
			expected:  false,
		},
		{
			name:      "Subdirectory in embedded filesystem",
			path:      "testdata/public/partials",
			embedfs:   &testdata,
			existFile: ExistFileFS(testdata),
			expected:  true,
		},
		{
			name:      "Subdirectory on OS filesystem",
			path:      "testdata/public/partials",
			embedfs:   nil,
			existFile: ExistFileOS,
			expected:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := IsDirWithExistFile(test.path, test.existFile, test.embedfs)
			if actual != test.expected {
				t.Errorf("IsDirWithExistFile(%s, existFunc, %v) = %v, expected %v", test.path, test.embedfs, actual, test.expected)
			}
		})
	}
}

func TestExistFileOS(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "Existing directory",
			path:     "testdata/public",
			expected: true,
		},
		{
			name:     "Existing file",
			path:     "testdata/public/index.html",
			expected: true,
		},
		{
			name:     "Non-existing file",
			path:     "/path/to/nonexistent.txt",
			expected: false,
		},
		{
			name:     "Non-existing directory",
			path:     "/path/to/nonexistent",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := ExistFileOS(test.path)
			if actual != test.expected {
				t.Errorf("ExistFileOS(%s) = %v, expected %v", test.path, actual, test.expected)
			}
		})
	}
}

func TestExistFileFS(t *testing.T) {
	existFunc := ExistFileFS(testdata)

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "Existing directory in embed",
			path:     "testdata/public",
			expected: true,
		},
		{
			name:     "Existing file in embed",
			path:     "testdata/public/index.html",
			expected: true,
		},
		{
			name:     "Non-existing file in embed",
			path:     "testdata/public/nonexistent.html",
			expected: false,
		},
		{
			name:     "Non-existing directory in embed",
			path:     "testdata/nonexistent",
			expected: false,
		},
		{
			name:     "Subdirectory in embed",
			path:     "testdata/public/partials",
			expected: true,
		},
		{
			name:     "File in subdirectory in embed",
			path:     "testdata/public/partials/header.html",
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := existFunc(test.path)
			if actual != test.expected {
				t.Errorf("ExistFileFS(%s) = %v, expected %v", test.path, actual, test.expected)
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
			name:       "Non-Embedded CSS Files don't exist",
			path:       "testdata/public",
			extensions: []string{".css"},
			embedfs:    nil,
			expected:   []string{},
		},
		{
			name:       "Single file - embedded",
			path:       "testdata/public/index.html",
			extensions: []string{".html"},
			embedfs:    &testdata,
			expected:   []string{"testdata/public/index.html"},
		},
		{
			name:       "Single file - OS",
			path:       "testdata/public/index.html",
			extensions: []string{".html"},
			embedfs:    nil,
			expected:   []string{"testdata/public/index.html"},
		},
		{
			name:       "Non-existing path - embedded",
			path:       "testdata/nonexistent",
			extensions: []string{".html"},
			embedfs:    &testdata,
			expected:   []string{},
		},
		{
			name:       "Non-existing path - OS",
			path:       "/path/to/nonexistent",
			extensions: []string{".html"},
			embedfs:    nil,
			expected:   []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := Find(test.path, test.extensions, test.embedfs)
			if len(actual) != len(test.expected) {
				t.Errorf("Find(%s, %v, %v) returned %v files, expected %v files", test.path, test.extensions, test.embedfs, len(actual), len(test.expected))
			} else {
				for i := range actual {
					if actual[i] != test.expected[i] {
						t.Errorf("Find(%s, %v, %v) returned %s, expected %s", test.path, test.extensions, test.embedfs, actual[i], test.expected[i])
					}
				}
			}
		})
	}
}

// TestFileExistenceIntegration tests the integration between file existence checking and directory detection
// This ensures that the proper existFile function is used in real scenarios
func TestFileExistenceIntegration(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := t.TempDir()

	// Create test files and directories
	testDir := filepath.Join(tempDir, "testdir")
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	testFile := filepath.Join(tempDir, "testfile.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test with OS filesystem
	t.Run("OS filesystem", func(t *testing.T) {
		existFile := ExistFileOS

		// Test existing directory
		if !IsDirWithExistFile(testDir, existFile, nil) {
			t.Errorf("Expected %s to be detected as directory", testDir)
		}

		// Test existing file (should not be directory)
		if IsDirWithExistFile(testFile, existFile, nil) {
			t.Errorf("Expected %s to not be detected as directory", testFile)
		}

		// Test non-existing path
		nonExistentPath := filepath.Join(tempDir, "nonexistent")
		if IsDirWithExistFile(nonExistentPath, existFile, nil) {
			t.Errorf("Expected %s to not be detected as directory", nonExistentPath)
		}
	})

	// Test with embedded filesystem
	t.Run("Embedded filesystem", func(t *testing.T) {
		existFile := ExistFileFS(testdata)

		// Test existing directory in embed
		if !IsDirWithExistFile("testdata/public", existFile, &testdata) {
			t.Error("Expected testdata/public to be detected as directory in embedded fs")
		}

		// Test existing file in embed (should not be directory)
		if IsDirWithExistFile("testdata/public/index.html", existFile, &testdata) {
			t.Error("Expected testdata/public/index.html to not be detected as directory in embedded fs")
		}

		// Test non-existing path in embed
		if IsDirWithExistFile("testdata/nonexistent", existFile, &testdata) {
			t.Error("Expected testdata/nonexistent to not be detected as directory in embedded fs")
		}
	})
}

// TestFindWithError tests the FindWithError function for proper error handling
func TestFindWithError(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		extensions  []string
		embedfs     *embed.FS
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Existing directory - OS filesystem",
			path:        "testdata/public",
			extensions:  []string{".html"},
			embedfs:     nil,
			expectError: false,
		},
		{
			name:        "Existing directory - embedded filesystem",
			path:        "testdata/public",
			extensions:  []string{".html"},
			embedfs:     &testdata,
			expectError: false,
		},
		{
			name:        "Existing file - OS filesystem",
			path:        "testdata/public/index.html",
			extensions:  []string{".html"},
			embedfs:     nil,
			expectError: false,
		},
		{
			name:        "Existing file - embedded filesystem",
			path:        "testdata/public/index.html",
			extensions:  []string{".html"},
			embedfs:     &testdata,
			expectError: false,
		},
		{
			name:        "Non-existing path - OS filesystem",
			path:        "/path/to/nonexistent",
			extensions:  []string{".html"},
			embedfs:     nil,
			expectError: true,
			errorMsg:    "file or directory not found",
		},
		{
			name:        "Non-existing path - embedded filesystem",
			path:        "testdata/nonexistent",
			extensions:  []string{".html"},
			embedfs:     &testdata,
			expectError: true,
			errorMsg:    "file or directory not found",
		},
		{
			name:        "Non-existing content file",
			path:        "nonexistent-content.html",
			extensions:  []string{".html"},
			embedfs:     nil,
			expectError: true,
			errorMsg:    "file or directory not found",
		},
		{
			name:        "Non-existing layout file",
			path:        "layouts/nonexistent-layout.html",
			extensions:  []string{".html"},
			embedfs:     nil,
			expectError: true,
			errorMsg:    "file or directory not found",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			files, err := FindWithError(test.path, test.extensions, test.embedfs)

			if test.expectError {
				if err == nil {
					t.Errorf("Expected error for path %s, but got none", test.path)
				} else if !strings.Contains(err.Error(), test.errorMsg) {
					t.Errorf("Expected error message to contain '%s', but got: %v", test.errorMsg, err)
				}
				if len(files) != 0 {
					t.Errorf("Expected empty files slice on error, but got: %v", files)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for path %s, but got: %v", test.path, err)
				}
			}
		})
	}
}

// TestFindOrExitWithExitCode tests that FindOrExit actually calls os.Exit(1) when files don't exist
// This test uses a subprocess approach to test the exit behavior
func TestFindOrExitWithExitCode(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") == "1" {
		// This is the subprocess that will call FindOrExit
		// It should exit with code 1 when files don't exist
		FindOrExit("/nonexistent/path", []string{".html"}, nil)
		return
	}

	// Test that FindOrExit calls os.Exit(1) for non-existing files
	t.Run("Non-existing file should exit with code 1", func(t *testing.T) {
		cmd := exec.Command(os.Args[0], "-test.run=TestFindOrExitWithExitCode")
		cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")

		err := cmd.Run()
		if err == nil {
			t.Fatal("Expected subprocess to exit with non-zero code, but it succeeded")
		}

		exitError, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("Expected ExitError, got %T: %v", err, err)
		}

		if exitError.ExitCode() != 1 {
			t.Errorf("Expected exit code 1, got %d", exitError.ExitCode())
		}
	})

	// Test that FindOrExit succeeds for existing files
	t.Run("Existing file should not exit", func(t *testing.T) {
		// This should not exit
		files := FindOrExit("testdata/public", []string{".html"}, nil)
		if len(files) == 0 {
			t.Error("Expected to find files, but got empty slice")
		}
	})

	// Test with embedded filesystem
	t.Run("Existing file in embedded fs should not exit", func(t *testing.T) {
		files := FindOrExit("testdata/public", []string{".html"}, &testdata)
		if len(files) == 0 {
			t.Error("Expected to find files in embedded fs, but got empty slice")
		}
	})
}

// TestFindOrExitNonExistingContent tests the exit behavior specifically for content files
func TestFindOrExitNonExistingContent(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") == "1" {
		// Test different content file scenarios that should cause exit
		switch os.Getenv("TEST_SCENARIO") {
		case "missing_content":
			FindOrExit("missing-content.html", []string{".html"}, nil)
		case "missing_layout":
			FindOrExit("layouts/missing-layout.html", []string{".html"}, nil)
		case "missing_directory":
			FindOrExit("missing-directory", []string{".html"}, nil)
		case "missing_embedded_content":
			FindOrExit("missing-embedded.html", []string{".html"}, &testdata)
		}
		return
	}

	testCases := []struct {
		name     string
		scenario string
		desc     string
	}{
		{
			name:     "Missing content file should exit with code 1",
			scenario: "missing_content",
			desc:     "When a content file doesn't exist",
		},
		{
			name:     "Missing layout file should exit with code 1",
			scenario: "missing_layout",
			desc:     "When a layout file doesn't exist",
		},
		{
			name:     "Missing directory should exit with code 1",
			scenario: "missing_directory",
			desc:     "When a content directory doesn't exist",
		},
		{
			name:     "Missing embedded content should exit with code 1",
			scenario: "missing_embedded_content",
			desc:     "When content doesn't exist in embedded filesystem",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=TestFindOrExitNonExistingContent")
			cmd.Env = append(os.Environ(),
				"GO_WANT_HELPER_PROCESS=1",
				"TEST_SCENARIO="+tc.scenario,
			)

			err := cmd.Run()
			if err == nil {
				t.Fatalf("%s: Expected subprocess to exit with non-zero code, but it succeeded", tc.desc)
			}

			exitError, ok := err.(*exec.ExitError)
			if !ok {
				t.Fatalf("%s: Expected ExitError, got %T: %v", tc.desc, err, err)
			}

			if exitError.ExitCode() != 1 {
				t.Errorf("%s: Expected exit code 1, got %d", tc.desc, exitError.ExitCode())
			}
		})
	}
}

// TestFindOrExitErrorOutput tests that FindOrExit writes error messages to stderr
func TestFindOrExitErrorOutput(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") == "1" {
		FindOrExit("/completely/nonexistent/path/file.html", []string{".html"}, nil)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestFindOrExitErrorOutput")
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")

	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	stderrOutput := make([]byte, 1024)
	n, _ := stderr.Read(stderrOutput)
	stderr.Close()

	if err := cmd.Wait(); err == nil {
		t.Fatal("Expected subprocess to exit with error")
	}

	errorOutput := string(stderrOutput[:n])
	if !strings.Contains(errorOutput, "Error:") {
		t.Errorf("Expected error output to contain 'Error:', got: %s", errorOutput)
	}
	if !strings.Contains(errorOutput, "file or directory not found") {
		t.Errorf("Expected error output to contain 'file or directory not found', got: %s", errorOutput)
	}
}
