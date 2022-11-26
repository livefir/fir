package entgo

import (
	"embed"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var simpleAssetsPath = "simple_assets"

//go:embed simple_assets/*
var simpleAssets embed.FS

func New(projectName string) {
	fs.WalkDir(simpleAssets, simpleAssetsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		outPath := strings.TrimPrefix(path, simpleAssetsPath)
		outPath = strings.TrimSuffix(outPath, ".str")
		outPath = filepath.Join(projectName, outPath)

		if d.IsDir() {
			if err := os.MkdirAll(outPath, os.ModePerm); err != nil {
				return err
			}
			return nil
		}

		execTextTemplate(path, outPath, map[string]any{"projectName": projectName})
		return nil
	})

}
