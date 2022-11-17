package boards

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/adnaan/autobahn/models"

	"github.com/adnaan/autobahn/models/board"
	"github.com/adnaan/autobahn/utils"
	"github.com/adnaan/fir"
)

var defaultPageSize = 5

type View struct {
	DB *models.Client
	fir.DefaultView
}

func (v *View) Content() string {
	return "views/boards/index"
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

	boards, err := boardQuery(v.DB, req).All(r.Context())
	if err != nil {
		return fir.ErrInternalServer(err)
	}

	data := fir.Data{"boards": boards}
	for k, v := range paginationData(req, len(boards)) {
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

	board, err := saveBoard(r.Context(), v.DB, req)
	if err != nil {
		return utils.PageFormError(err)
	}

	http.Redirect(w, r, fmt.Sprintf("/%s/show", board.ID.String()), http.StatusFound)

	return fir.Page{}
}

func (v *View) OnEvent(event fir.Event) fir.Patchset {
	switch event.ID {
	case "board-create":
		return onBoardCreate(v.DB, event)
	case "board-query":
		return onBoardQuery(v.DB, event)
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

func boardQuery(db *models.Client, req queryReq) *models.BoardQuery {
	if req.Limit == 0 {
		req.Limit = defaultPageSize
	}
	q := db.Board.Query().Offset(req.Offset).Limit(req.Limit)

	if req.Search != "" {
		q = q.Where(board.TitleContains(req.Search))
	}

	if req.Order == "oldest" {
		q = q.Order(models.Desc("create_time"))
	} else {
		q = q.Order(models.Asc("create_time"))
	}

	return q
}

func paginationData(req queryReq, boardLen int) fir.Data {
	prev := req.Offset - defaultPageSize
	hasPrevious := true
	if prev < 0 || req.Offset == 0 {
		hasPrevious = false
	}
	next := defaultPageSize + req.Offset
	hasNext := true
	if boardLen < defaultPageSize {
		hasNext = false
	}
	return fir.Data{
		"prev":        prev,
		"next":        next,
		"hasPrevious": hasPrevious,
		"hasNext":     hasNext,
		"search":      req.Search,
	}
}

func saveBoard(ctx context.Context, db *models.Client, req createReq) (*models.Board, error) {
	board, err := db.Board.
		Create().
		SetTitle(req.Title).
		SetDescription(req.Description).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	return board, nil
}

func onBoardCreate(db *models.Client, event fir.Event) fir.Patchset {
	var req createReq
	if err := event.DecodeFormParams(&req); err != nil {
		return fir.PatchError(err, "error decoding request")
	}

	board, err := saveBoard(event.RequestContext(), db, req)
	if err != nil {
		return utils.PatchFormError(err)
	}

	return fir.Patchset{
		fir.Navigate{
			To: fmt.Sprintf("/%s/show", board.ID.String()),
		},
	}
}

func onBoardQuery(db *models.Client, event fir.Event) fir.Patchset {
	var req queryReq
	if err := event.DecodeFormParams(&req); err != nil {
		return fir.PatchError(err, "error decoding request")
	}

	boards, err := boardQuery(db, req).All(event.RequestContext())
	if err != nil {
		return fir.PatchError(err, "error querying boards")
	}

	return fir.Patchset{
		fir.Morph{
			Selector: "#boardlist",
			Template: &fir.Template{
				Name: "boardlist",
				Data: fir.Data{"boards": boards},
			},
		},
		fir.Morph{
			Selector: "#pagination",
			Template: &fir.Template{
				Name: "pagination",
				Data: paginationData(req, len(boards)),
			},
		},
	}
}
