package todos

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/adnaan/fir/cli/testdata/todos/models"

	"github.com/adnaan/fir/cli/testdata/todos/models/board"
	"github.com/adnaan/fir/cli/testdata/todos/models/predicate"

	"github.com/adnaan/fir"
	"github.com/adnaan/fir/cli/testdata/todos/models/todo"
	"github.com/adnaan/fir/cli/testdata/todos/utils"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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

	boardID, err := uuid.Parse(chi.URLParam(r, "boardID"))
	if err != nil {
		return fir.PageError(err, "error parsing board id")
	}
	req.boardID = boardID

	todos, err := todoQuery(v.DB, req).All(r.Context())
	if err != nil {
		return fir.ErrInternalServer(err)
	}

	data := map[string]any{"todos": todos}
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

	boardID, err := uuid.Parse(chi.URLParam(r, "boardID"))
	if err != nil {
		return fir.PageError(err, "error parsing board id")
	}
	req.boardID = boardID

	todo, err := saveTodo(r.Context(), v.DB, req)
	if err != nil {
		return utils.PageFormError(err)
	}

	http.Redirect(w, r, fmt.Sprintf("/%s/todos/%s/show", req.boardID.String(), todo.ID.String()), http.StatusFound)

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
	boardID uuid.UUID

	Order  string `json:"order" schema:"order"`
	Search string `json:"search" schema:"search"`
	Offset int    `json:"offset" schema:"offset"`
	Limit  int    `json:"limit" schema:"limit"`
}

type createReq struct {
	boardID uuid.UUID

	Title       string `json:"title" schema:"title,required"`
	Description string `json:"description" schema:"description,required"`
}

func todoQuery(db *models.Client, req queryReq) *models.TodoQuery {
	if req.Limit == 0 {
		req.Limit = defaultPageSize
	}
	q := db.Todo.Query().Offset(req.Offset).Limit(req.Limit)

	ps := []predicate.Todo{todo.HasOwnerWith(board.ID(req.boardID))}
	if req.Search != "" {
		ps = append(ps, todo.TitleContains(req.Search))
	}
	q = q.Where(ps...)
	q = q.WithOwner()

	if req.Order == "oldest" {
		q = q.Order(models.Desc("create_time"))
	} else {
		q = q.Order(models.Asc("create_time"))
	}

	return q
}

func paginationData(req queryReq, todoLen int) map[string]any {
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
	return map[string]any{
		"prev":        prev,
		"next":        next,
		"hasPrevious": hasPrevious,
		"hasNext":     hasNext,
		"search":      req.Search,
	}
}

func saveTodo(ctx context.Context, db *models.Client, req createReq) (*models.Todo, error) {
	todo, err := db.Todo.
		Create().
		SetTitle(req.Title).
		SetDescription(req.Description).
		SetOwnerID(req.boardID).
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

	boardID, err := uuid.Parse(chi.URLParamFromCtx(event.RequestContext(), "boardID"))
	if err != nil {
		return fir.PatchError(err, "error parsing board id")
	}
	req.boardID = boardID

	todo, err := saveTodo(event.RequestContext(), db, req)
	if err != nil {
		return utils.PatchFormError(err)
	}

	return fir.Patchset{
		fir.Navigate{
			To: fmt.Sprintf("/%s/todos/%s/show", req.boardID.String(), todo.ID.String()),
		},
	}
}

func onTodoQuery(db *models.Client, event fir.Event) fir.Patchset {
	var req queryReq
	if err := event.DecodeFormParams(&req); err != nil {
		return fir.PatchError(err, "error decoding request")
	}

	boardID, err := uuid.Parse(chi.URLParamFromCtx(event.RequestContext(), "boardID"))
	if err != nil {
		return fir.PatchError(err, "error parsing board id")
	}
	req.boardID = boardID

	todos, err := todoQuery(db, req).All(event.RequestContext())
	if err != nil {
		return fir.PatchError(err, "error querying todos")
	}

	return fir.Patchset{
		fir.Morph{
			Selector: "#todolist",
			Template: &fir.Template{
				Name: "todolist",
				Data: map[string]any{"todos": todos},
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
