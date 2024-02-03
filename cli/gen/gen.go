package gen

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/yosssi/gohtml"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var simpleAssetsPath = "simple_assets"
var routesAssetsPath = "routes_assets"

//go:embed simple_assets/*
var simpleAssets embed.FS

//go:embed routes_assets/*
var routesAssets embed.FS

func NewProject(projectName string) {
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

		execTextTemplate(path, outPath, simpleAssets, map[string]any{"projectName": projectName})
		return nil
	})

	fmt.Printf("Generated project: %v\n", projectName)
	fmt.Printf(`
To run the project:

cd %v
go run main.go

`, projectName)

}

func NewRoute(routeName string) {
	routeName = strings.TrimSpace(routeName)
	routeNameTitle := cases.Title(language.English).String(routeName)
	routeNameLower := cases.Lower(language.English).String(routeName)

	fs.WalkDir(routesAssets, routesAssetsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Fatal(err)
		}

		outPath := strings.TrimPrefix(path, routesAssetsPath)
		outPath = strings.TrimSuffix(outPath, ".str")
		outPath = filepath.Join("routes", outPath)
		outPath = strings.Replace(outPath, "route.", routeNameLower+".", -1)

		if d.IsDir() {
			if err := os.MkdirAll(outPath, os.ModePerm); err != nil {
				return err
			}
			return nil
		}

		execTextTemplate(path, outPath, routesAssets, map[string]any{
			"route":      routeNameTitle,
			"routeLower": routeNameLower,
		})
		return nil
	})

	// add route to the http handler
	fmt.Printf(`
Generated route: %v
To add the route to the http handler, add the following line to the http handler:

http.Handle("/%v", controller.RouteFunc(routes.%v))

`,
		routeNameLower, routeNameLower, routeNameTitle)

}

func execTextTemplate(inFile, outFile string, embedfs embed.FS, vars map[string]any) {
	inFileData, err := embedfs.ReadFile(inFile)
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

	// don't overwrite existing files
	if _, err := os.Stat(outFile); err == nil {
		fmt.Printf("file already exists: %s,  skipping...\n", outFile)
		return
	}

	err = os.WriteFile(outFile, inFileData, 0644)
	checkerr(err)
	buf.Reset()
}

func checkerr(err error) {
	if err != nil {
		panic(err)
	}
}
