package todos

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/adnaan/autobahn/utils"
	"github.com/adnaan/fir"
	"github.com/adnaan/fir/cli/testdata/todos/models"
	"github.com/adnaan/fir/cli/testdata/todos/models/todo"
)

var defaultPageSize = 5

type View struct {
	DB *models.Client
	fir.DefaultView
}

func (v *View) Content() string {
	return "views/todos/index"
}

func (v *View) Layout() string {
	return "./templates/layouts/index.html"
}

func (v *View) OnGet(w http.ResponseWriter, r *http.Request) fir.Page {
	var req queryReq
	err := fir.DecodeURLValues(&req, r)
	if err != nil {
		return fir.PageError(err, "error decoding query params")
	}

	todos, err := todoQuery(v.DB, req).All(r.Context())
	if err != nil {
		return fir.ErrInternalServer(err)
	}

	data := fir.Data{"todos": todos}
	for k, v := range paginationData(req, len(todos)) {
		data[k] = v
	}

	return fir.Page{Data: data}
}

func (v *View) OnPost(w http.ResponseWriter, r *http.Request) fir.Page {
	var req createReq
	err := fir.DecodeForm(&req, r)
	if err != nil {
		return fir.PageError(err)
	}

	todo, err := saveTodo(r.Context(), v.DB, req)
	if err != nil {
		return utils.PageFormError(err)
	}

	http.Redirect(w, r, fmt.Sprintf("/%s", todo.ID.String()), http.StatusFound)

	return fir.Page{}
}

func (v *View) OnEvent(event fir.Event) fir.Patchset {
	switch event.ID {
	case "todo-create":
		return onTodoCreate(v.DB, event)
	case "todo-query":
		return onTodoQuery(v.DB, event)
	default:
		log.Printf("unknown event: %s\n", event.ID)
		return nil
	}
}

type queryReq struct {
	Order  string `json:"order" schema:"order"`
	Search string `json:"search" schema:"search"`
	Offset int    `json:"offset" schema:"offset"`
	Limit  int    `json:"limit" schema:"limit"`
}

type createReq struct {
	Title       string `json:"title" schema:"title,required"`
	Description string `json:"description" schema:"description,required"`
}

func todoQuery(db *models.Client, req queryReq) *models.TodoQuery {
	if req.Limit == 0 {
		req.Limit = defaultPageSize
	}

	q := db.Todo.Query().Offset(req.Offset).Limit(req.Limit)
	if req.Search != "" {
		q = q.Where(todo.TitleContains(req.Search))
	}
	if req.Order == "oldest" {
		q = q.Order(models.Desc("create_time"))
	} else {
		q = q.Order(models.Asc("create_time"))
	}

	return q
}

func paginationData(req queryReq, todoLen int) fir.Data {
	prev := req.Offset - defaultPageSize
	hasPrevious := true
	if prev < 0 || req.Offset == 0 {
		hasPrevious = false
	}
	next := defaultPageSize + req.Offset
	hasNext := true
	if todoLen < defaultPageSize {
		hasNext = false
	}
	return fir.Data{
		"prev":        prev,
		"next":        next,
		"hasPrevious": hasPrevious,
		"hasNext":     hasNext,
	}
}

func saveTodo(ctx context.Context, db *models.Client, req createReq) (*models.Todo, error) {
	todo, err := db.Todo.
		Create().
		SetTitle(req.Title).
		SetDescription(req.Description).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	return todo, nil
}

func onTodoCreate(db *models.Client, event fir.Event) fir.Patchset {
	var req createReq
	if err := event.DecodeFormParams(&req); err != nil {
		return fir.PatchError(err, "error decoding request")
	}

	todo, err := saveTodo(event.RequestContext(), db, req)
	if err != nil {
		return utils.PatchFormError(err)
	}

	return fir.Patchset{
		fir.Navigate{
			To: fmt.Sprintf("/%s", todo.ID.String()),
		},
	}
}

func onTodoQuery(db *models.Client, event fir.Event) fir.Patchset {
	var req queryReq
	if err := event.DecodeFormParams(&req); err != nil {
		return fir.PatchError(err, "error decoding request")
	}
	todos, err := todoQuery(db, req).All(event.RequestContext())
	if err != nil {
		return fir.PatchError(err, "error querying todos")
	}

	return fir.Patchset{
		fir.Morph{
			Selector: "#todolist",
			Template: &fir.Template{
				Name: "todolist",
				Data: fir.Data{"todos": todos},
			},
		},
		fir.Morph{
			Selector: "#pagination",
			Template: &fir.Template{
				Name: "pagination",
				Data: paginationData(req, len(todos)),
			},
		},
	}
}
