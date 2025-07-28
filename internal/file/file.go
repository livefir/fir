package file

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"slices"
)

type ReadFileFunc func(string) (string, []byte, error)
type ExistFileFunc func(string) bool

// Find is the original function that returns empty slice when files don't exist
// This maintains backward compatibility for existing code and tests
func Find(path string, extensions []string, embedfs *embed.FS) []string {
	files, _ := FindWithError(path, extensions, embedfs)
	return files
}

// FindWithError returns files and an error, allowing callers to handle missing files
func FindWithError(path string, extensions []string, embedfs *embed.FS) ([]string, error) {
	var files []string
	var fi fs.FileInfo
	var err error

	if embedfs != nil {
		fi, err = fs.Stat(*embedfs, path)
		if err != nil {
			return files, fmt.Errorf("file or directory not found: %s", path)
		}
	} else {
		fi, err = os.Stat(path)
		if err != nil {
			return files, fmt.Errorf("file or directory not found: %s", path)
		}
	}

	if !fi.IsDir() {
		if !slices.Contains(extensions, filepath.Ext(path)) {
			return files, nil
		}
		files = append(files, path)
		return files, nil
	}

	if embedfs != nil {
		err = fs.WalkDir(*embedfs, path, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if slices.Contains(extensions, filepath.Ext(d.Name())) {
				files = append(files, path)
			}
			return nil
		})

		if err != nil {
			return files, fmt.Errorf("cannot access path: %s - %v", path, err)
		}

	} else {
		err = filepath.WalkDir(path, func(fpath string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if slices.Contains(extensions, filepath.Ext(d.Name())) {
				files = append(files, fpath)
			}
			return nil
		})

		if err != nil {
			return files, fmt.Errorf("cannot access path: %s - %v", path, err)
		}
	}

	return files, nil
}

// FindOrExit exits with error code 1 if files are not found
// Use this function when you want the application to exit on missing files
func FindOrExit(path string, extensions []string, embedfs *embed.FS) []string {
	files, err := FindWithError(path, extensions, embedfs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	return files
}

// IsDirWithExistFile checks if a path is a directory using the provided existFile function
// and filesystem configuration. This should be used instead of isDir when the existFile
// function is available from the route options.
func IsDirWithExistFile(path string, existFile ExistFileFunc, embedfs *embed.FS) bool {
	// First check if the path exists
	if !existFile(path) {
		return false
	}

	// Then check if it's a directory using the appropriate filesystem
	if embedfs != nil {
		fileInfo, err := fs.Stat(*embedfs, path)
		if err != nil {
			return false
		}
		return fileInfo.IsDir()
	}
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}

	return fileInfo.IsDir()
}

func ReadFileOS(file string) (name string, b []byte, err error) {
	name = filepath.Base(file)
	b, err = os.ReadFile(file)
	return
}

func ReadFileFS(fsys fs.FS) func(string) (string, []byte, error) {
	return func(file string) (name string, b []byte, err error) {
		name = path.Base(file)
		b, err = fs.ReadFile(fsys, file)
		return
	}
}

func ExistFileOS(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
}

func ExistFileFS(fsys fs.FS) func(string) bool {
	return func(path string) bool {
		if _, err := fs.Stat(fsys, path); err != nil {
			return false
		}
		return true
	}
}
