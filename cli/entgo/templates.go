package entgo

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	pluralize "github.com/gertd/go-pluralize"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func PrepareTemplates(schemaPath, viewsPath, templateAssetsPath string) string {
	g, err := entc.LoadGraph(schemaPath, &gen.Config{})
	checkerr(err)
	tmpDir, err := ioutil.TempDir("", "")
	checkerr(err)
	for _, node := range g.Nodes {
		buildTemplates(node, tmpDir, viewsPath, templateAssetsPath)
	}
	return tmpDir
}

func buildTemplates(node *gen.Type, templatesPath, viewsPath, templateAssetsPath string) {
	viewName := strings.ToLower(node.Name)
	pluralize := pluralize.NewClient()
	pluralizedViewName := pluralize.Plural(viewName)

	fmt.Println("building templates for", viewName)
	if err := os.MkdirAll(fmt.Sprintf("%s/%s", viewsPath, viewName), os.ModePerm); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(fmt.Sprintf("%s/%s", viewsPath, pluralizedViewName), os.ModePerm); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(fmt.Sprintf("%s/%s/partials", viewsPath, pluralizedViewName), os.ModePerm); err != nil {
		panic(err)
	}

	replaceVars := map[string]string{
		"$VIEW_NAME":        viewName,
		"$VIEW_PLURAL_NAME": pluralizedViewName,
	}
	// model/index.go
	generateTemplateWithReplace(
		filepath.Join(templateAssetsPath, "model_index_go.str"),
		filepath.Join(templatesPath, "model_index_go.tmpl"),
		replaceVars,
	)

	// models/index.go
	generateTemplateWithReplace(
		filepath.Join(templateAssetsPath, "models_index_go.str"),
		filepath.Join(templatesPath, "models_index_go.tmpl"),
		replaceVars,
	)

	// views/model/index.html
	generateTemplateWithReplace(
		filepath.Join(templateAssetsPath, "model_index_html.str"),
		filepath.Join(viewsPath, viewName, "index.html"),
		replaceVars,
	)

	// views/models/index.html
	generateTemplateWithReplace(
		filepath.Join(templateAssetsPath, "models_index_html.str"),
		filepath.Join(viewsPath, pluralizedViewName, "index.html"),
		replaceVars,
	)

	// views/models/partials/model.html
	generateTemplateWithReplace(
		filepath.Join(templateAssetsPath, "model_partials_html.str"),
		filepath.Join(viewsPath, pluralizedViewName, "partials", viewName+".html"),
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
		"name":   viewName,
		"action": "new",
		"fields": template.HTML(strings.Join(fields, "\n")),
	})
	replaceVars["$FORM"] = buf.String()
	// generate template
	generateTemplateWithReplace(
		filepath.Join(templateAssetsPath, "new_model_partials_html.str"),
		filepath.Join(viewsPath, pluralizedViewName, "partials", "new_"+viewName+".html"),
		replaceVars,
	)
}

func loadDefaultElements(templateAssetsPath string) *template.Template {
	return template.Must(template.ParseFiles(filepath.Join(templateAssetsPath, "elements.html")))
}

func generateTemplateWithReplace(inFile, outFile string, vars map[string]string) {
	outData, err := ioutil.ReadFile(filepath.Join(inFile))
	checkerr(err)
	for k, v := range vars {
		outData = bytes.ReplaceAll(outData, []byte(k), []byte(v))
	}
	err = ioutil.WriteFile(outFile, outData, 0644)
	checkerr(err)
}
