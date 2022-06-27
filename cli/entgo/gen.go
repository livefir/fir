package entgo

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	htmlTemplate "html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	pluralize "github.com/gertd/go-pluralize"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func checkerr(err error) {
	if err != nil {
		panic(err)
	}
}

var templateAssetsPath = "./template_assets"

//go:embed template_assets/*
var templateAssets embed.FS

func Generate(projectPath, modelsPkg string) {
	schemaPath, err := filepath.Abs(filepath.Join(projectPath, "schema"))
	checkerr(err)
	g, err := entc.LoadGraph(schemaPath, &gen.Config{})
	checkerr(err)
	for _, node := range g.Nodes {
		buildTemplates(node, projectPath, modelsPkg)
	}
}

func buildTemplates(node *gen.Type, projectPath, modelsPkg string) {
	viewsPath := filepath.Join(projectPath, "views")
	modelName := strings.ToLower(node.Name)
	pluralize := pluralize.NewClient()
	pluralizedModelName := pluralize.Plural(modelName)

	fmt.Println("generating views for: ", modelName)
	if err := os.MkdirAll(fmt.Sprintf("%s/%s", viewsPath, modelName), os.ModePerm); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(fmt.Sprintf("%s/%s", viewsPath, pluralizedModelName), os.ModePerm); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(fmt.Sprintf("%s/%s/partials", viewsPath, pluralizedModelName), os.ModePerm); err != nil {
		panic(err)
	}

	// views/model/index.go
	execTextTemplate(
		filepath.Join(templateAssetsPath, "model_index_go.str"),
		filepath.Join(viewsPath, modelName, "index.go"),
		map[string]any{
			"pkgName":   modelName,
			"content":   filepath.Join("./views", modelName),
			"layout":    "./templates/layouts/index.html",
			"modelsPkg": modelsPkg,
		},
	)

	// views/models/index.go
	execTextTemplate(
		filepath.Join(templateAssetsPath, "models_index_go.str"),
		filepath.Join(viewsPath, pluralizedModelName, "index.go"),
		map[string]any{
			"pkgName":         pluralizedModelName,
			"content":         filepath.Join("./views", pluralizedModelName),
			"layout":          "./templates/layouts/index.html",
			"modelNamePlural": pluralizedModelName,
			"modelNameTitled": cases.Title(language.AmericanEnglish).String(modelName),
			"modelsPkg":       modelsPkg,
		},
	)

	replaceVars := map[string]string{
		"$MODEL_NAME":        modelName,
		"$MODEL_PLURAL_NAME": pluralizedModelName,
	}
	// views/model/index.html
	genHtmlTemplate(
		filepath.Join(templateAssetsPath, "model_index_html.str"),
		filepath.Join(viewsPath, modelName, "index.html"),
		replaceVars,
	)

	// views/models/index.html
	genHtmlTemplate(
		filepath.Join(templateAssetsPath, "models_index_html.str"),
		filepath.Join(viewsPath, pluralizedModelName, "index.html"),
		replaceVars,
	)

	// views/models/partials/model.html
	genHtmlTemplate(
		filepath.Join(templateAssetsPath, "model_partials_html.str"),
		filepath.Join(viewsPath, pluralizedModelName, "partials", modelName+".html"),
		replaceVars,
	)

	// views/models/partials/new_model.html
	// build form
	elements := loadDefaultElements(filepath.Join(templateAssetsPath, "elements.html"))
	var fields []string
	var buf bytes.Buffer
	for _, field := range node.Fields {
		switch field.Type.String() {
		case "string":
			err := elements.ExecuteTemplate(&buf, "bulma:form-input", map[string]any{
				"label":       cases.Title(language.AmericanEnglish).String(field.Name),
				"name":        strings.ToLower(field.Name),
				"type":        "text",
				"placeholder": field.Name,
			})
			fields = append(fields, buf.String())
			buf.Reset()
			checkerr(err)
		default:
		}
	}
	elements.ExecuteTemplate(&buf, "bulma:form", map[string]any{
		"name":   modelName,
		"action": "new",
		"fields": htmlTemplate.HTML(strings.Join(fields, "\n")),
	})
	replaceVars["$FORM"] = buf.String()
	// generate template
	genHtmlTemplate(
		filepath.Join(templateAssetsPath, "new_model_partials_html.str"),
		filepath.Join(viewsPath, pluralizedModelName, "partials", "new_"+modelName+".html"),
		replaceVars,
	)
}

func loadDefaultElements(elementsPath string) *htmlTemplate.Template {
	data, err := templateAssets.ReadFile(elementsPath)
	checkerr(err)
	return htmlTemplate.Must(htmlTemplate.New("").Parse(string(data)))
}

func genHtmlTemplate(inFile, outFile string, vars map[string]string) {
	inFileData, err := templateAssets.ReadFile(inFile)
	checkerr(err)
	for k, v := range vars {
		inFileData = bytes.ReplaceAll(inFileData, []byte(k), []byte(v))
	}
	err = ioutil.WriteFile(outFile, inFileData, 0644)
	checkerr(err)
}

func execTextTemplate(inFile, outFile string, vars map[string]any) {
	inFileData, err := templateAssets.ReadFile(inFile)
	checkerr(err)

	t := template.Must(template.New("gen").Parse(string(inFileData)))
	buf := new(bytes.Buffer)
	err = t.ExecuteTemplate(buf, "gen", vars)
	checkerr(err)

	inFileData, err = format.Source(buf.Bytes())
	checkerr(err)

	err = ioutil.WriteFile(outFile, inFileData, 0644)
	checkerr(err)
	buf.Reset()
}
