package entgo

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	htmlTemplate "html/template"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	"github.com/Masterminds/sprig"
	pluralize "github.com/gertd/go-pluralize"
	"github.com/yosssi/gohtml"
	"golang.org/x/exp/slices"
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
	modelName := strings.ToLower(node.Name)
	pluralize := pluralize.NewClient()
	pluralizedModelName := pluralize.Plural(modelName)

	fmt.Println(">> generating pages for schema: ", modelName)
	nodeHasParent := false
	nodeParentName := ""
	var children []string
	for _, edge := range node.Edges {
		// if owner is the same as modelName
		if edge.Owner.Name == node.Name {
			// edge is a child
			children = append(children, edge.Name)
			continue
		} else {
			// edge is a parent
			if edge.Unique {
				fmt.Printf("node has a parent: %+v\n", edge.Owner.Name)
				nodeHasParent = true
				nodeParentName = strings.ToLower(edge.Owner.Name)
			}
		}
	}

	annotationError := fmt.Sprintf(`
error: fir page annotations not found for: %s

The generator requires the three annotations to be present in the entgo schema.
e.g.

func (Todo) Annotations() []schema.Annotation {
	return []schema.Annotation{
		fir.CreateForm{
			Fields: []string{"title", "description"},
		},
		fir.UpdateForm{
			Fields: []string{"title", "description"},
		},
		fir.ListItem{
			Fields: []string{"title"},
		},
	}
}

See the fir documentation for more information.
	
`, modelName)

	if len(node.Annotations) == 0 {
		fmt.Printf("  >> skip generating pages for schema: %s. page annotations not found. see cli docs for details.\n", modelName)
		return
	}

	if node.Annotations["CreateForm"] == nil || node.Annotations["UpdateForm"] == nil || node.Annotations["ListItem"] == nil {
		fmt.Print(annotationError)
		return
	}

	var createFormAnnotation []string
	for _, v := range (node.Annotations["CreateForm"].(map[string]any))["Fields"].([]interface{}) {
		createFormAnnotation = append(createFormAnnotation, v.(string))
	}

	var updateFormAnnotation []string
	for _, v := range (node.Annotations["UpdateForm"].(map[string]any))["Fields"].([]interface{}) {
		updateFormAnnotation = append(updateFormAnnotation, v.(string))
	}

	var listItemAnnotation []string
	for _, v := range (node.Annotations["ListItem"].(map[string]any))["Fields"].([]interface{}) {
		listItemAnnotation = append(listItemAnnotation, v.(string))
	}

	elements := loadDefaultElements(filepath.Join(templateAssetsPath, "elements.html"))
	var createFormFields []*gen.Field
	var createFormFieldsHtml []string
	var buf bytes.Buffer

	for _, field := range node.Fields {
		if !slices.Contains(createFormAnnotation, field.Name) {
			continue
		}
		createFormFields = append(createFormFields, field)
		switch field.Type.String() {
		case "string":
			err := elements.ExecuteTemplate(&buf, "bulma:form-input",
				map[string]any{
					"label":       cases.Title(language.AmericanEnglish).String(field.Name),
					"name":        field.Name,
					"type":        "text",
					"placeholder": field.StructField(),
					"value":       "",
				})
			createFormFieldsHtml = append(createFormFieldsHtml, buf.String())
			buf.Reset()
			checkerr(err)
		default:
		}
	}
	var updateFormFieldsHtml []string
	var updateFormFields []*gen.Field
	for _, field := range node.Fields {
		if !slices.Contains(updateFormAnnotation, field.Name) {
			continue
		}
		updateFormFields = append(updateFormFields, field)
		switch field.Type.String() {
		case "string":
			err := elements.ExecuteTemplate(&buf, "bulma:form-input",
				map[string]any{
					"label":       field.StructField(),
					"name":        field.Name,
					"type":        "text",
					"placeholder": field.StructField(),
					"value":       fmt.Sprintf("{{.%s}}", field.StructField()),
				})
			updateFormFieldsHtml = append(updateFormFieldsHtml, buf.String())
			buf.Reset()
			checkerr(err)
		default:
		}
	}

	var listItemFields []*gen.Field
	for _, field := range node.Fields {
		if !slices.Contains(listItemAnnotation, field.Name) {
			continue
		}
		listItemFields = append(listItemFields, field)
	}

	data := map[string]any{
		"pkgName":               pluralizedModelName,
		"content":               filepath.Join("./pages", pluralizedModelName),
		"layout":                "./templates/layouts/index.html",
		"modelName":             modelName,
		"modelNamePlural":       pluralizedModelName,
		"modelNameTitled":       cases.Title(language.AmericanEnglish).String(modelName),
		"modelNameTitledPlural": cases.Title(language.AmericanEnglish).String(pluralizedModelName),
		"modelsPkg":             modelsPkg,
		"utilsPkg":              strings.Replace(modelsPkg, "models", "utils", 1),
		"createFormFieldsHtml":  htmlTemplate.HTML(strings.Join(createFormFieldsHtml, "\n")),
		"updateFormFieldsHtml":  htmlTemplate.HTML(strings.Join(updateFormFieldsHtml, "\n")),
		"createFormFields":      createFormFields,
		"updateFormFields":      updateFormFields,
		"listItemFields":        listItemFields,
		"hasParent":             nodeHasParent,
		"nodeParentName":        nodeParentName,
		"children":              children,
	}

	//glog.Errorf("%+v\n", data)

	pkgPrefix := "template_assets/package"

	fs.WalkDir(templateAssets, pkgPrefix, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		outPath := strings.TrimPrefix(path, pkgPrefix)
		outPath = strings.TrimSuffix(outPath, ".str")
		outPath = strings.Replace(outPath, "pages/models", "pages/"+pluralizedModelName, -1)
		outPath = strings.Replace(outPath, "model", modelName, -1)
		outPath = filepath.Join(projectPath, outPath)

		if d.IsDir() {
			if err := os.MkdirAll(outPath, os.ModePerm); err != nil {
				return err
			}
			return nil
		}

		execTextTemplate(path, outPath, data)
		return nil
	})

}

func loadDefaultElements(elementsPath string) *template.Template {
	data, err := templateAssets.ReadFile(elementsPath)
	checkerr(err)
	return template.Must(template.New("").Delims("[[", "]]").Parse(string(data)))
}

func execTextTemplate(inFile, outFile string, vars map[string]any) {
	inFileData, err := templateAssets.ReadFile(inFile)
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
