package show

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/adnaan/fir"
	"github.com/adnaan/fir/cli/testdata/todos/models"
	"github.com/adnaan/fir/cli/testdata/todos/models/todo"
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
	id, err := uuid.Parse(chi.URLParam(r, "todoID"))
	if err != nil {
		return fir.PageError(err, "invalid todo id")
	}
	todo, err := v.DB.Todo.Query().
		Where(todo.ID(id)).
		WithOwner().
		Only(r.Context())
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

	id, err := uuid.Parse(chi.URLParam(r, "todoID"))
	if err != nil {
		return fir.PageError(err, "invalid todo id")
	}
	req.id = id

	switch req.Action {
	case "update":
		todo, err := updateTodo(r.Context(), v.DB, req)
		if err != nil {
			return utils.PageFormError(err)
		}
		return fir.Page{Data: structs.Map(todo)}

	case "delete":
		if err := deleteTodo(r.Context(), v.DB, id); err != nil {
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
	id          uuid.UUID
	Action      string `json:"action" schema:"action"`
	Title       string `json:"title" schema:"title"`
	Description string `json:"description" schema:"description"`
}

func updateTodo(ctx context.Context, db *models.Client, req updateTodoReq) (*models.Todo, error) {
	todo, err := db.Todo.
		UpdateOneID(req.id).
		SetTitle(req.Title).
		SetDescription(req.Description).
		Save(ctx)
	return todo, err
}

func deleteTodo(ctx context.Context, db *models.Client, id uuid.UUID) error {
	return db.Todo.DeleteOneID(id).Exec(ctx)
}

func onTodoUpdate(db *models.Client, event fir.Event) fir.Patchset {
	var req updateTodoReq
	if err := event.DecodeFormParams(&req); err != nil {
		return fir.PatchError(err)
	}

	id, err := uuid.Parse(chi.URLParamFromCtx(event.RequestContext(), "todoID"))
	if err != nil {
		return fir.PatchError(err, "invalid todo id")
	}
	req.id = id

	todo, err := updateTodo(event.RequestContext(), db, req)
	if err != nil {
		return utils.PatchFormError(err)
	}

	patchset := append(
		fir.UnsetFormErrors("title", "description"),
		fir.Morph{
			Selector: "todo",
			HTML: &fir.Render{
				Template: "todo",
				Data:     structs.Map(todo),
			},
		})

	return patchset
}

func onTodoDelete(db *models.Client, event fir.Event) fir.Patchset {
	boardID, err := uuid.Parse(chi.URLParamFromCtx(event.RequestContext(), "boardID"))
	if err != nil {
		return fir.PatchError(err, "invalid board id")
	}

	id, err := uuid.Parse(chi.URLParamFromCtx(event.RequestContext(), "todoID"))
	if err != nil {
		return fir.PatchError(err, "invalid todo id")
	}

	if err := deleteTodo(event.RequestContext(), db, id); err != nil {
		return fir.PatchError(err)
	}
	return fir.Patchset{
		fir.Navigate{
			To: fmt.Sprintf("/%s/todos", boardID),
		},
	}
}
