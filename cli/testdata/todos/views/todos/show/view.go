package show

import (
	"context"
	"log"
	"net/http"

	"github.com/adnaan/fir"
	"github.com/adnaan/fir/cli/testdata/todos/models"
	"github.com/adnaan/fir/cli/testdata/todos/utils"
	"github.com/fatih/structs"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type View struct {
	DB *models.Client
	fir.DefaultView
}

func (v *View) Content() string {
	return "views/todos/show"
}

func (v *View) Layout() string {
	return "./templates/layouts/index.html"
}

func (v *View) OnGet(w http.ResponseWriter, r *http.Request) fir.Page {

	id := chi.URLParam(r, "id")
	uid, err := uuid.Parse(id)
	if err != nil {
		return fir.PageError(err, "invalid todo id")
	}
	todo, err := v.DB.Todo.Get(r.Context(), uid)
	if err != nil {
		return fir.ErrNotFound(err, "todo not found")
	}

	return fir.Page{Data: structs.Map(todo)}
}

func (v *View) OnPost(w http.ResponseWriter, r *http.Request) fir.Page {
	var req updateTodoReq
	err := fir.DecodeForm(&req, r)
	if err != nil {
		return fir.PageError(err, "error decoding request")
	}

	req.id = chi.URLParamFromCtx(r.Context(), "id")

	switch req.Action {
	case "update":
		todo, err := updateTodo(r.Context(), v.DB, req)
		if err != nil {
			return utils.PageFormError(err)
		}
		return fir.Page{Data: structs.Map(todo)}

	case "delete":
		if err := deleteTodo(r.Context(), v.DB, req); err != nil {
			return fir.PageError(err, "error deleting todo")
		}
		http.Redirect(w, r, "/", http.StatusFound)
	default:
		return fir.PageError(err, "unknown request")

	}

	return fir.Page{}
}

func (v *View) OnEvent(event fir.Event) fir.Patchset {
	switch event.ID {
	case "todo-update":
		return onTodoUpdate(v.DB, event)
	case "todo-delete":
		return onTodoDelete(v.DB, event)
	default:
		log.Printf("unknown event: %s\n", event.ID)
		return nil
	}
}

type updateTodoReq struct {
	id          string
	Action      string `json:"action" schema:"action"`
	Title       string `json:"title" schema:"title"`
	Description string `json:"description" schema:"description"`
}

func updateTodo(ctx context.Context, db *models.Client, req updateTodoReq) (*models.Todo, error) {
	id, err := uuid.Parse(req.id)
	if err != nil {
		return nil, err
	}
	todo, err := db.Todo.
		UpdateOneID(id).
		SetTitle(req.Title).
		SetDescription(req.Description).
		Save(ctx)
	return todo, err
}

func deleteTodo(ctx context.Context, db *models.Client, req updateTodoReq) error {
	id, err := uuid.Parse(req.id)
	if err != nil {
		return err
	}
	return db.Todo.DeleteOneID(id).Exec(ctx)
}

func onTodoUpdate(db *models.Client, event fir.Event) fir.Patchset {
	var req updateTodoReq
	if err := event.DecodeFormParams(&req); err != nil {
		return fir.PatchError(err)
	}

	req.id = chi.URLParamFromCtx(event.RequestContext(), "id")

	todo, err := updateTodo(event.RequestContext(), db, req)
	if err != nil {
		return utils.PatchFormError(err)
	}

	patchset := append(
		fir.UnsetFormErrors("title", "description"),
		fir.Morph{
			Selector: "todo",
			Template: &fir.Template{
				Name: "todo",
				Data: structs.Map(todo),
			},
		})

	return patchset
}

func onTodoDelete(db *models.Client, event fir.Event) fir.Patchset {
	id := chi.URLParamFromCtx(event.RequestContext(), "id")
	if err := deleteTodo(event.RequestContext(), db, updateTodoReq{id: id}); err != nil {
		return fir.PatchError(err)
	}
	return fir.Patchset{
		fir.Navigate{
			To: "/",
		},
	}
}
