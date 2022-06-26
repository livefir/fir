package todo

import (
	"github.com/adnaan/fir"
	"github.com/adnaan/fir/cli/testdata/todos/models"
)

type View struct {
	DB *models.Client
	fir.DefaultView
}

func (v *View) Content() string {
	return "views/todo"
}

func (v *View) Layout() string {
	return "./templates/layouts/index.html"
}
