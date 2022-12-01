package show

import (
	"context"
	"net/http"

	"github.com/adnaan/fir"
	"github.com/adnaan/fir/cli/testdata/todos/models"
	"github.com/adnaan/fir/cli/testdata/todos/models/board"
	"github.com/adnaan/fir/cli/testdata/todos/utils"
	"github.com/fatih/structs"
	"github.com/go-chi/chi/v5"
	"github.com/golang/glog"
	"github.com/google/uuid"
)

type View struct {
	DB *models.Client
	fir.DefaultView
}

func (v *View) Content() string {
	return "views/boards/show"
}

func (v *View) Layout() string {
	return "./templates/layouts/index.html"
}

func (v *View) OnGet(w http.ResponseWriter, r *http.Request) fir.Pagedata {
	id, err := uuid.Parse(chi.URLParam(r, "boardID"))
	if err != nil {
		return fir.PageError(err, "invalid board id")
	}
	board, err := v.DB.Board.Query().
		Where(board.ID(id)).
		Only(r.Context())
	if err != nil {
		return fir.ErrNotFound(err, "board not found")
	}

	return fir.Pagedata{Data: structs.Map(board)}
}

func (v *View) OnPost(w http.ResponseWriter, r *http.Request) fir.Pagedata {
	var req updateBoardReq
	err := fir.DecodeForm(&req, r)
	if err != nil {
		return fir.PageError(err, "error decoding request")
	}

	id, err := uuid.Parse(chi.URLParam(r, "boardID"))
	if err != nil {
		return fir.PageError(err, "invalid board id")
	}
	req.id = id

	switch req.Action {
	case "update":
		board, err := updateBoard(r.Context(), v.DB, req)
		if err != nil {
			return utils.PageFormError(err)
		}
		return fir.Pagedata{Data: structs.Map(board)}

	case "delete":
		if err := deleteBoard(r.Context(), v.DB, id); err != nil {
			return fir.PageError(err, "error deleting board")
		}
		http.Redirect(w, r, "/", http.StatusFound)
	default:
		return fir.PageError(err, "unknown request")

	}

	return fir.Pagedata{}
}

func (v *View) OnEvent(event fir.Event) fir.Patchset {
	switch event.ID {
	case "board-update":
		return onBoardUpdate(v.DB, event)
	case "board-delete":
		return onBoardDelete(v.DB, event)
	default:
		glog.Errorf("unknown event: %s\n", event.ID)
		return nil
	}
}

type updateBoardReq struct {
	id          uuid.UUID
	Action      string `json:"action" schema:"action"`
	Title       string `json:"title" schema:"title"`
	Description string `json:"description" schema:"description"`
}

func updateBoard(ctx context.Context, db *models.Client, req updateBoardReq) (*models.Board, error) {
	board, err := db.Board.
		UpdateOneID(req.id).
		SetTitle(req.Title).
		SetDescription(req.Description).
		Save(ctx)
	return board, err
}

func deleteBoard(ctx context.Context, db *models.Client, id uuid.UUID) error {
	return db.Board.DeleteOneID(id).Exec(ctx)
}

func onBoardUpdate(db *models.Client, event fir.Event) fir.Patchset {
	var req updateBoardReq
	if err := event.DecodeFormParams(&req); err != nil {
		return fir.PatchError(err)
	}

	id, err := uuid.Parse(chi.URLParamFromCtx(event.RequestContext(), "boardID"))
	if err != nil {
		return fir.PatchError(err, "invalid board id")
	}
	req.id = id

	board, err := updateBoard(event.RequestContext(), db, req)
	if err != nil {
		return utils.PatchFormError(err)
	}

	patchset := append(
		fir.UnsetFormErrors("title", "description"),
		fir.Morph{
			Selector: "board",
			HTML: &fir.Render{
				Template: "board",
				Data:     structs.Map(board),
			},
		})

	return patchset
}

func onBoardDelete(db *models.Client, event fir.Event) fir.Patchset {

	id, err := uuid.Parse(chi.URLParamFromCtx(event.RequestContext(), "boardID"))
	if err != nil {
		return fir.PatchError(err, "invalid board id")
	}

	if err := deleteBoard(event.RequestContext(), db, id); err != nil {
		return fir.PatchError(err)
	}
	return fir.Patchset{
		fir.Navigate{
			To: "/",
		},
	}
}
