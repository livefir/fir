package stories

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/adnaan/autobahn/models"

	"github.com/adnaan/autobahn/models/board"
	"github.com/adnaan/autobahn/models/predicate"

	"github.com/adnaan/autobahn/models/story"
	"github.com/adnaan/autobahn/utils"
	"github.com/adnaan/fir"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

var defaultPageSize = 5

type View struct {
	DB *models.Client
	fir.DefaultView
}

func (v *View) Content() string {
	return "views/stories/index"
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

	stories, err := storyQuery(v.DB, req).All(r.Context())
	if err != nil {
		return fir.ErrInternalServer(err)
	}

	data := fir.Data{"stories": stories}
	for k, v := range paginationData(req, len(stories)) {
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

	story, err := saveStory(r.Context(), v.DB, req)
	if err != nil {
		return utils.PageFormError(err)
	}

	http.Redirect(w, r, fmt.Sprintf("/%s/stories/%s/show", req.boardID.String(), story.ID.String()), http.StatusFound)

	return fir.Page{}
}

func (v *View) OnEvent(event fir.Event) fir.Patchset {
	switch event.ID {
	case "story-create":
		return onStoryCreate(v.DB, event)
	case "story-query":
		return onStoryQuery(v.DB, event)
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

func storyQuery(db *models.Client, req queryReq) *models.StoryQuery {
	if req.Limit == 0 {
		req.Limit = defaultPageSize
	}
	q := db.Story.Query().Offset(req.Offset).Limit(req.Limit)

	ps := []predicate.Story{story.HasOwnerWith(board.ID(req.boardID))}
	if req.Search != "" {
		ps = append(ps, story.TitleContains(req.Search))
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

func paginationData(req queryReq, storyLen int) fir.Data {
	prev := req.Offset - defaultPageSize
	hasPrevious := true
	if prev < 0 || req.Offset == 0 {
		hasPrevious = false
	}
	next := defaultPageSize + req.Offset
	hasNext := true
	if storyLen < defaultPageSize {
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

func saveStory(ctx context.Context, db *models.Client, req createReq) (*models.Story, error) {
	story, err := db.Story.
		Create().
		SetTitle(req.Title).
		SetDescription(req.Description).
		SetOwnerID(req.boardID).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	return story, nil
}

func onStoryCreate(db *models.Client, event fir.Event) fir.Patchset {
	var req createReq
	if err := event.DecodeFormParams(&req); err != nil {
		return fir.PatchError(err, "error decoding request")
	}

	boardID, err := uuid.Parse(chi.URLParamFromCtx(event.RequestContext(), "boardID"))
	if err != nil {
		return fir.PatchError(err, "error parsing board id")
	}
	req.boardID = boardID

	story, err := saveStory(event.RequestContext(), db, req)
	if err != nil {
		return utils.PatchFormError(err)
	}

	return fir.Patchset{
		fir.Navigate{
			To: fmt.Sprintf("/%s/stories/%s/show", req.boardID.String(), story.ID.String()),
		},
	}
}

func onStoryQuery(db *models.Client, event fir.Event) fir.Patchset {
	var req queryReq
	if err := event.DecodeFormParams(&req); err != nil {
		return fir.PatchError(err, "error decoding request")
	}

	boardID, err := uuid.Parse(chi.URLParamFromCtx(event.RequestContext(), "boardID"))
	if err != nil {
		return fir.PatchError(err, "error parsing board id")
	}
	req.boardID = boardID

	stories, err := storyQuery(db, req).All(event.RequestContext())
	if err != nil {
		return fir.PatchError(err, "error querying stories")
	}

	return fir.Patchset{
		fir.Morph{
			Selector: "#storylist",
			Template: &fir.Template{
				Name: "storylist",
				Data: fir.Data{"stories": stories},
			},
		},
		fir.Morph{
			Selector: "#pagination",
			Template: &fir.Template{
				Name: "pagination",
				Data: paginationData(req, len(stories)),
			},
		},
	}
}
