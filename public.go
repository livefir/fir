package fir

import (
	"io/fs"
	"os"
	"path/filepath"

	gitignore "github.com/sabhiram/go-gitignore"
)

type publicOpt struct {
	inputDir   string
	outDir     string
	extensions []string
}

type PublicOption func(*publicOpt)

func InputDir(path string) PublicOption {
	return func(o *publicOpt) {
		o.inputDir = path
	}
}

func OutDir(path string) PublicOption {
	return func(o *publicOpt) {
		o.outDir = path
	}
}

func Extensions(extensions []string) PublicOption {
	return func(o *publicOpt) {
		o.extensions = extensions
	}
}

func GeneratePublic(options ...PublicOption) error {
	opt := &publicOpt{
		inputDir:   ".",
		outDir:     "./public",
		extensions: []string{".html"},
	}

	for _, option := range options {
		option(opt)
	}

	if err := os.MkdirAll(opt.outDir, os.ModePerm); err != nil {
		return err
	}

	ignore, err := gitignore.CompileIgnoreFile(".gitignore")
	if err != nil {
		return err
	}

	err = filepath.WalkDir(opt.inputDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if d.Name() == filepath.Clean(opt.outDir) {
				return filepath.SkipDir
			}
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			if ignore.MatchesPath(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		if ignore.MatchesPath(path) {
			return nil
		}

		if !contains(opt.extensions, filepath.Ext(path)) {
			return nil
		}

		outPath := filepath.Join(opt.outDir, path)
		if err := os.MkdirAll(filepath.Dir(outPath), os.ModePerm); err != nil {
			return err
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(outPath, data, os.ModePerm)
	})

	return err
}
