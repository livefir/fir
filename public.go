package fir

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"

	gitignore "github.com/sabhiram/go-gitignore"
)

type publicOpt struct {
	inDir      string
	outDir     string
	extensions []string
}

type PublicOption func(*publicOpt)

func InDir(path string) PublicOption {
	return func(o *publicOpt) {
		o.inDir = path
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
		inDir:      ".",
		outDir:     "./public",
		extensions: []string{".html"},
	}

	for _, option := range options {
		option(opt)
	}

	if err := os.MkdirAll(opt.outDir, os.ModePerm); err != nil {
		return err
	}

	ignore, err := gitignore.CompileIgnoreFile(filepath.Join(opt.inDir, ".gitignore"))
	if err != nil {
		log.Printf("[warning] failed to compile .gitignore: %v\n", err)
	}

	err = filepath.WalkDir(opt.inDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if d.Name() == filepath.Base(opt.outDir) {
				return filepath.SkipDir
			}
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			if ignore != nil && ignore.MatchesPath(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		if ignore != nil && ignore.MatchesPath(path) {
			return nil
		}

		if !contains(opt.extensions, filepath.Ext(path)) {
			return nil
		}

		relpath, err := filepath.Rel(opt.inDir, path)
		if err != nil {
			return err
		}

		outPath := filepath.Join(opt.outDir, relpath)
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
