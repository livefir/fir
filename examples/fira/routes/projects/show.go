package projects

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/livefir/fir"
	"github.com/livefir/fir/examples/fira/ent"
)

func load(db *ent.Client) fir.OnEventFunc {
	type query struct {
		ID string `json:"id"`
	}
	return func(ctx fir.RouteContext) error {
		var q query
		if err := ctx.Bind(&q); err != nil {
			return err
		}
		uid, err := uuid.Parse(q.ID)
		if err != nil {
			return err
		}
		project, err := db.Project.Get(ctx.Request().Context(), uid)
		if err != nil {
			return err
		}
		return ctx.Data(project)
	}
}

func updateProject(db *ent.Client) fir.OnEventFunc {
	type updateReq struct {
		ID          string `json:"projectID"`
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	return func(ctx fir.RouteContext) error {
		var req updateReq
		if err := ctx.Bind(&req); err != nil {
			return err
		}
		uid, err := uuid.Parse(req.ID)
		if err != nil {
			return err
		}

		project, err := db.Project.UpdateOneID(uid).
			SetTitle(req.Title).
			SetDescription(req.Description).
			Save(ctx.Request().Context())
		if err != nil {
			fmt.Println("save error", err)
			return toFieldError(ctx, err)
		}
		return ctx.Data(project)
	}
}

func deleteProject(db *ent.Client) fir.OnEventFunc {
	type deleteReq struct {
		ID string `json:"id"`
	}
	return func(ctx fir.RouteContext) error {
		var req deleteReq
		if err := ctx.Bind(&req); err != nil {
			return err
		}
		uid, err := uuid.Parse(req.ID)
		if err != nil {
			return err
		}
		err = db.Project.DeleteOneID(uid).Exec(ctx.Request().Context())
		if err != nil {
			return err
		}
		return ctx.Data(nil)
	}
}

func Show(db *ent.Client) fir.RouteFunc {
	return func() fir.RouteOptions {
		return fir.RouteOptions{
			fir.ID("project"),
			fir.Content("routes/projects/show.html"),
			fir.Layout("routes/layout.html"),
			fir.Partials("routes/partials"),
			fir.OnLoad(load(db)),
			fir.OnEvent("update", updateProject(db)),
			fir.OnEvent("delete", deleteProject(db)),
		}
	}
}
