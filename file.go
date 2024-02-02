package fir

import (
	"embed"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"slices"

	"github.com/livefir/fir/internal/logger"
)

type readFileFunc func(string) (string, []byte, error)
type existFileFunc func(string) bool

func find(path string, extensions []string, embedfs *embed.FS) []string {
	var files []string
	var fi fs.FileInfo
	var err error

	if embedfs != nil {
		fi, err = fs.Stat(*embedfs, path)
		if err != nil {
			return files
		}
	} else {
		fi, err = os.Stat(path)
		if err != nil {
			return files
		}
	}

	if !fi.IsDir() {
		if !slices.Contains(extensions, filepath.Ext(path)) {
			return files
		}
		files = append(files, path)
		return files
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
			panic(err)
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
			panic(err)
		}

	}

	return files
}

func isDir(path string, embedfs *embed.FS) bool {
	if embedfs != nil {
		fileInfo, err := fs.Stat(*embedfs, path)
		if err != nil {
			logger.Warnf("isDir: %v", err)
			return false
		}
		return fileInfo.IsDir()
	}
	fileInfo, err := os.Stat(path)
	if err != nil {
		logger.Warnf("isDir: %v", err)
		return false
	}

	return fileInfo.IsDir()
}

func readFileOS(file string) (name string, b []byte, err error) {
	name = filepath.Base(file)
	b, err = os.ReadFile(file)
	return
}

func readFileFS(fsys fs.FS) func(string) (string, []byte, error) {
	return func(file string) (name string, b []byte, err error) {
		name = path.Base(file)
		b, err = fs.ReadFile(fsys, file)
		return
	}
}

func existFileOS(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
}

func existFileFS(fsys fs.FS) func(string) bool {
	return func(path string) bool {
		if _, err := fs.Stat(fsys, path); err != nil {
			return false
		}
		return true
	}
}
