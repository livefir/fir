package fir

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"golang.org/x/exp/slices"
)

func find(opt opt, p string, extensions []string) []string {
	var files []string
	var fi fs.FileInfo
	var err error

	if opt.hasEmbedFS {
		fi, err = fs.Stat(opt.embedFS, p)
		if err != nil {
			return files
		}
	} else {
		fi, err = os.Stat(p)
		if err != nil {
			return files
		}
	}

	if !fi.IsDir() {
		if !slices.Contains(extensions, filepath.Ext(p)) {
			return files
		}
		files = append(files, p)
		return files
	}

	if opt.hasEmbedFS {
		err = fs.WalkDir(opt.embedFS, p, func(path string, d fs.DirEntry, err error) error {
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

		err = filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
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

	}

	return files
}

func isDir(path string, opt opt) bool {
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

func isFileHTML(path string, opt opt) bool {
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
