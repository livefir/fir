package fir

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/golang/glog"
	gitignore "github.com/sabhiram/go-gitignore"
	"golang.org/x/exp/slices"
)

type publicOpt struct {
	inDir      string
	outDir     string
	extensions []string
}

// PublicDirOption is a function that can be used to configure generation of public directory using GeneratePublic.
type PublicDirOption func(*publicOpt)

// InDir sets the input directory for the public directory.
func InDir(path string) PublicDirOption {
	return func(o *publicOpt) {
		o.inDir = path
	}
}

// OutputDir sets the output directory for the public directory.
func OutDir(path string) PublicDirOption {
	return func(o *publicOpt) {
		o.outDir = path
	}
}

// Extension adds an extension to the list of extensions that will be copied over.
func PublicFileExtensions(extensions []string) PublicDirOption {
	return func(o *publicOpt) {
		for _, ext := range extensions {
			if !slices.Contains(o.extensions, ext) {
				o.extensions = append(o.extensions, ext)
			}
		}
	}
}

// GeneratePublicDir generates the public directory which can be then embedded into the binary.
func GeneratePublicDir(options ...PublicDirOption) error {
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
		glog.Errorf("[warning] failed to compile .gitignore: %v\n", err)
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

		if !slices.Contains(opt.extensions, filepath.Ext(path)) {
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
