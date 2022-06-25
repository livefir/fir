package entgo

import (
	"path/filepath"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

type ViewExtension struct {
	entc.DefaultExtension
	TemplatesPath string
}

func (v *ViewExtension) Templates() []*gen.Template {
	return []*gen.Template{
		gen.MustParse(gen.NewTemplate("modelindexgo").ParseFiles(filepath.Join(v.TemplatesPath, "model_index_go.tmpl"))),
		gen.MustParse(gen.NewTemplate("modelsindexgo").ParseFiles(filepath.Join(v.TemplatesPath, "models_index_go.tmpl"))),
	}
}

func checkerr(err error) {
	if err != nil {
		panic(err)
	}
}
