package entgo

import (
	"bytes"
	"fmt"
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

func Generate(schemaPath, viewsPath, templateAssetsPath, modelsPkg string) {
	g, err := entc.LoadGraph(schemaPath, &gen.Config{})
	checkerr(err)
	tmpDir, err := ioutil.TempDir("", "")
	checkerr(err)
	for _, node := range g.Nodes {
		buildTemplates(node, tmpDir, viewsPath, templateAssetsPath, modelsPkg)
	}
}

func buildTemplates(node *gen.Type, templatesPath, viewsPath, templateAssetsPath, modelsPkg string) {
	modelName := strings.ToLower(node.Name)
	pluralize := pluralize.NewClient()
	pluralizedModelName := pluralize.Plural(modelName)

	fmt.Println("building templates for", modelName)
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
	elements := loadDefaultElements(templateAssetsPath)
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

func loadDefaultElements(templateAssetsPath string) *htmlTemplate.Template {
	return htmlTemplate.Must(htmlTemplate.ParseFiles(filepath.Join(templateAssetsPath, "elements.html")))
}

func genHtmlTemplate(inFile, outFile string, vars map[string]string) {
	outData, err := ioutil.ReadFile(filepath.Join(inFile))
	checkerr(err)
	for k, v := range vars {
		outData = bytes.ReplaceAll(outData, []byte(k), []byte(v))
	}
	err = ioutil.WriteFile(outFile, outData, 0644)
	checkerr(err)
}

func execTextTemplate(inFile, outFile string, vars map[string]any) {
	outData, err := ioutil.ReadFile(filepath.Join(inFile))
	checkerr(err)

	t := template.Must(template.New("gen").Parse(string(outData)))
	buf := new(bytes.Buffer)
	err = t.ExecuteTemplate(buf, "gen", vars)
	checkerr(err)

	err = ioutil.WriteFile(outFile, buf.Bytes(), 0644)
	checkerr(err)
}
