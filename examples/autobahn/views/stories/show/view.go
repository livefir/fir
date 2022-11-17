package show

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/adnaan/autobahn/models"
	"github.com/adnaan/autobahn/models/story"
	"github.com/adnaan/autobahn/utils"
	"github.com/adnaan/fir"
	"github.com/fatih/structs"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type View struct {
	DB *models.Client
	fir.DefaultView
}

func (v *View) Content() string {
	return "views/stories/show"
}

func (v *View) Layout() string {
	return "./templates/layouts/index.html"
}

func (v *View) OnGet(w http.ResponseWriter, r *http.Request) fir.Page {
	id, err := uuid.Parse(chi.URLParam(r, "storyID"))
	if err != nil {
		return fir.PageError(err, "invalid story id")
	}
	story, err := v.DB.Story.Query().
		Where(story.ID(id)).
		WithOwner().
		Only(r.Context())
	if err != nil {
		return fir.ErrNotFound(err, "story not found")
	}

	return fir.Page{Data: structs.Map(story)}
}

func (v *View) OnPost(w http.ResponseWriter, r *http.Request) fir.Page {
	var req updateStoryReq
	err := fir.DecodeForm(&req, r)
	if err != nil {
		return fir.PageError(err, "error decoding request")
	}

	id, err := uuid.Parse(chi.URLParam(r, "storyID"))
	if err != nil {
		return fir.PageError(err, "invalid story id")
	}
	req.id = id

	switch req.Action {
	case "update":
		story, err := updateStory(r.Context(), v.DB, req)
		if err != nil {
			return utils.PageFormError(err)
		}
		return fir.Page{Data: structs.Map(story)}

	case "delete":
		if err := deleteStory(r.Context(), v.DB, id); err != nil {
			return fir.PageError(err, "error deleting story")
		}
		http.Redirect(w, r, "/", http.StatusFound)
	default:
		return fir.PageError(err, "unknown request")

	}

	return fir.Page{}
}

func (v *View) OnEvent(event fir.Event) fir.Patchset {
	switch event.ID {
	case "story-update":
		return onStoryUpdate(v.DB, event)
	case "story-delete":
		return onStoryDelete(v.DB, event)
	default:
		log.Printf("unknown event: %s\n", event.ID)
		return nil
	}
}

type updateStoryReq struct {
	id          uuid.UUID
	Action      string `json:"action" schema:"action"`
	Title       string `json:"title" schema:"title"`
	Description string `json:"description" schema:"description"`
}

func updateStory(ctx context.Context, db *models.Client, req updateStoryReq) (*models.Story, error) {
	story, err := db.Story.
		UpdateOneID(req.id).
		SetTitle(req.Title).
		SetDescription(req.Description).
		Save(ctx)
	return story, err
}

func deleteStory(ctx context.Context, db *models.Client, id uuid.UUID) error {
	return db.Story.DeleteOneID(id).Exec(ctx)
}

func onStoryUpdate(db *models.Client, event fir.Event) fir.Patchset {
	var req updateStoryReq
	if err := event.DecodeFormParams(&req); err != nil {
		return fir.PatchError(err)
	}

	id, err := uuid.Parse(chi.URLParamFromCtx(event.RequestContext(), "storyID"))
	if err != nil {
		return fir.PatchError(err, "invalid story id")
	}
	req.id = id

	story, err := updateStory(event.RequestContext(), db, req)
	if err != nil {
		return utils.PatchFormError(err)
	}

	patchset := append(
		fir.UnsetFormErrors("title", "description"),
		fir.Morph{
			Selector: "story",
			Template: &fir.Template{
				Name: "story",
				Data: structs.Map(story),
			},
		})

	return patchset
}

func onStoryDelete(db *models.Client, event fir.Event) fir.Patchset {
	boardID, err := uuid.Parse(chi.URLParamFromCtx(event.RequestContext(), "boardID"))
	if err != nil {
		return fir.PatchError(err, "invalid board id")
	}

	id, err := uuid.Parse(chi.URLParamFromCtx(event.RequestContext(), "storyID"))
	if err != nil {
		return fir.PatchError(err, "invalid story id")
	}

	if err := deleteStory(event.RequestContext(), db, id); err != nil {
		return fir.PatchError(err)
	}
	return fir.Patchset{
		fir.Navigate{
			To: fmt.Sprintf("/%s/stories", boardID),
		},
	}
}
