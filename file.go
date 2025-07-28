package fir

import (
	"embed"
	"io/fs"

	"github.com/livefir/fir/internal/file"
)

// Type aliases for backward compatibility
type readFileFunc = file.ReadFileFunc
type existFileFunc = file.ExistFileFunc

// Function aliases for backward compatibility
func find(path string, extensions []string, embedfs *embed.FS) []string {
	return file.Find(path, extensions, embedfs)
}

func findWithError(path string, extensions []string, embedfs *embed.FS) ([]string, error) {
	return file.FindWithError(path, extensions, embedfs)
}

func findOrExit(path string, extensions []string, embedfs *embed.FS) []string {
	return file.FindOrExit(path, extensions, embedfs)
}

func isDirWithExistFile(path string, existFile existFileFunc, embedfs *embed.FS) bool {
	return file.IsDirWithExistFile(path, existFile, embedfs)
}

func readFileOS(filename string) (name string, b []byte, err error) {
	return file.ReadFileOS(filename)
}

func readFileFS(fsys fs.FS) func(string) (string, []byte, error) {
	return file.ReadFileFS(fsys)
}

func existFileOS(path string) bool {
	return file.ExistFileOS(path)
}

func existFileFS(fsys fs.FS) func(string) bool {
	return file.ExistFileFS(fsys)
}
