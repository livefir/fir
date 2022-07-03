package todos

import (
	"log"
	"net/http"

	"github.com/adnaan/fir"
	"github.com/adnaan/fir/cli/testdata/todos/models"
)

type View struct {
	DB *models.Client
	fir.DefaultView
}

func (v *View) Content() string {
	return "views/todos"
}

func (v *View) Layout() string {
	return "./templates/layouts/index.html"
}

func (v *View) OnGet(w http.ResponseWriter, r *http.Request) fir.Page {
	todos, err := v.DB.Todo.Query().All(r.Context())
	if err != nil {
		log.Printf("error querying todos,: %s\n", err)
		return fir.Status{Code: 200, Message: "Internal Server Error"}, nil
	}

	return fir.Status{Code: http.StatusOK}, fir.Data{"todos,": todos}
}
