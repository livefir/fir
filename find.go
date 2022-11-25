package fir

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"golang.org/x/exp/slices"
)

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
			fmt.Println("[warning]isDir warn: ", err)
			return false
		}
		return fileInfo.IsDir()
	}
	fileInfo, err := os.Stat(path)
	if err != nil {
		fmt.Println("[warning]isDir error: ", err)
		return false
	}

	return fileInfo.IsDir()
}

func isFileOrString(path string, opt routeOpt) bool {
	if opt.hasEmbedFS {
		if _, err := fs.Stat(opt.embedFS, path); err != nil {
			return true
		}
		return false
	}
	if _, err := os.Stat(path); err != nil {
		return true
	}
	return false
}
