package entgo

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

func PrepareTemplates(schemaPath, viewsPath, templateAssetsPath string) string {
	g, err := entc.LoadGraph(schemaPath, &gen.Config{})
	checkerr(err)
	tmpDir, err := ioutil.TempDir("", "")
	checkerr(err)
	for _, node := range g.Nodes {
		buildTemplates(strings.ToLower(node.Name), tmpDir, viewsPath, templateAssetsPath)
	}
	return tmpDir
}

func buildTemplates(viewName, templatesPath, viewsPath, templateAssetsPath string) {
	fmt.Println("building templates for", viewName)
	if err := os.MkdirAll(fmt.Sprintf("%s/%s", viewsPath, viewName), os.ModePerm); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(fmt.Sprintf("%s/%ss", viewsPath, viewName), os.ModePerm); err != nil {
		panic(err)
	}
	// model/index.go
	modelIndexGoTempl, err := ioutil.ReadFile(filepath.Join(templateAssetsPath, "model_index_go.str"))
	checkerr(err)
	modelIndexGoTempl = bytes.ReplaceAll(modelIndexGoTempl, []byte("$VIEW_NAME"), []byte(viewName))
	err = ioutil.WriteFile(filepath.Join(templatesPath, "model_index_go.tmpl"), modelIndexGoTempl, 0644)
	checkerr(err)

	// models/index.go
	modelsIndexGoTempl, err := ioutil.ReadFile(filepath.Join(templateAssetsPath, "models_index_go.str"))
	checkerr(err)
	modelsIndexGoTempl = bytes.ReplaceAll(modelsIndexGoTempl, []byte("$VIEW_NAME"), []byte(viewName))
	err = ioutil.WriteFile(filepath.Join(templatesPath, "models_index_go.tmpl"), modelsIndexGoTempl, 0644)
	checkerr(err)

	// views/model/index.html
	modelIndexHtmlStr, err := ioutil.ReadFile(filepath.Join(templateAssetsPath, "model_index_html.str"))
	checkerr(err)
	indexHtml := bytes.ReplaceAll(modelIndexHtmlStr, []byte("$VIEW_NAME"), []byte(viewName))
	err = ioutil.WriteFile(fmt.Sprintf("%s/%s/index.html", viewsPath, viewName), indexHtml, 0644)
	checkerr(err)

	// views/models/index.html
	modelsIndexHtmlStr, err := ioutil.ReadFile(filepath.Join(templateAssetsPath, "models_index_html.str"))
	checkerr(err)
	indexHtml = bytes.ReplaceAll(modelsIndexHtmlStr, []byte("$VIEW_NAME"), []byte(viewName))
	err = ioutil.WriteFile(fmt.Sprintf("%s/%ss/index.html", viewsPath, viewName), indexHtml, 0644)
	checkerr(err)
}
