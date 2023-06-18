package fir

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"golang.org/x/exp/slices"
	"k8s.io/klog/v2"
)

type readFileFunc func(string) (string, []byte, error)
type existFileFunc func(string) bool

func find(opt routeOpt, path string, extensions []string) []string {
	var files []string
	var fi fs.FileInfo
	var err error

	if opt.hasEmbedFS {
		fi, err = fs.Stat(opt.embedFS, path)
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

	if opt.hasEmbedFS {
		err = fs.WalkDir(opt.embedFS, path, func(path string, d fs.DirEntry, err error) error {
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

func isDir(path string, opt routeOpt) bool {
	if opt.hasEmbedFS {
		fileInfo, err := fs.Stat(opt.embedFS, path)
		if err != nil {
			klog.Warningf("[warning]isDir warn: ", err)
			return false
		}
		return fileInfo.IsDir()
	}
	fileInfo, err := os.Stat(path)
	if err != nil {
		klog.Warningf("[warning]isDir error: ", err)
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
