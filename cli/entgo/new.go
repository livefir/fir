package entgo

import (
	"bytes"
	"embed"
	"go/format"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/yosssi/gohtml"
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

func execTextTemplate(inFile, outFile string, vars map[string]any) {
	inFileData, err := simpleAssets.ReadFile(inFile)
	checkerr(err)

	t := template.Must(
		template.New("gen").
			Delims("[[", "]]").
			Funcs(template.FuncMap(sprig.FuncMap())).
			Parse(string(inFileData)))
	buf := new(bytes.Buffer)
	err = t.ExecuteTemplate(buf, "gen", vars)
	checkerr(err)

	if filepath.Ext(outFile) == ".go" {
		inFileData, err = format.Source(buf.Bytes())
		checkerr(err)
	} else if filepath.Ext(outFile) == ".html" {
		inFileData = gohtml.FormatBytes(buf.Bytes())
	} else {
		inFileData = buf.Bytes()
	}

	err = ioutil.WriteFile(outFile, inFileData, 0644)
	checkerr(err)
	buf.Reset()
}

func checkerr(err error) {
	if err != nil {
		panic(err)
	}
}
