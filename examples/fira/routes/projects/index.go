package projects

import (
	"errors"
	"fmt"

	"github.com/livefir/fir"
	"github.com/livefir/fir/examples/fira/ent"
	"github.com/livefir/fir/examples/fira/ent/project"
)

var pageSize = 5

type queryReq struct {
	Order      string `json:"order"`
	Search     string `json:"search"`
	Page       int    `json:"page"`
	resultSize int
}

func projectQuery(db *ent.Client, req queryReq) *ent.ProjectQuery {
	offset := (req.Page - 1) * pageSize
	q := db.Project.Query().Offset(offset).Limit(pageSize + 1)

	if req.Search != "" {
		q = q.Where(project.TitleContains(req.Search))
	}

	if req.Order == "oldest" {
		q = q.Order(ent.Desc("create_time"))
	} else {
		q = q.Order(ent.Asc("create_time"))
	}

	return q
}

func paginationData(req queryReq) map[string]any {
	var hasNext bool
	if req.resultSize > pageSize {
		hasNext = true
	}

	var hasPrev bool
	if req.Page > 1 {
		hasPrev = true
	}

	return map[string]any{
		"prev":        req.Page - 1,
		"next":        req.Page + 1,
		"hasPrevious": hasPrev,
		"hasNext":     hasNext,
		"search":      req.Search,
		"order":       req.Order,
	}
}

func loadProjects(db *ent.Client) fir.OnEventFunc {
	return func(ctx fir.RouteContext) error {
		fmt.Println("loadProjects")
		var q queryReq
		if err := ctx.Bind(&q); err != nil {
			return err
		}

		if q.Page == 0 {
			q.Page = 1
		}

		projects, err := projectQuery(db, q).All(ctx.Request().Context())
		if err != nil {
			return err
		}

		q.resultSize = len(projects)

		data := map[string]any{"projects": projects}
		pageData := paginationData(q)
		return ctx.Data(data, pageData)
	}
}

type createReq struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func toFieldError(ctx fir.RouteContext, err error) error {
	var validError *ent.ValidationError
	if errors.As(err, &validError) {
		return ctx.FieldError(validError.Name, validError.Unwrap())
	}
	return err
}

func createProject(db *ent.Client) fir.OnEventFunc {
	return func(ctx fir.RouteContext) error {
		var req createReq
		if err := ctx.Bind(&req); err != nil {
			return err
		}
		project, err := db.Project.
			Create().
			SetTitle(req.Title).
			SetDescription(req.Description).
			Save(ctx.Request().Context())
		if err != nil {
			return toFieldError(ctx, err)
		}
		return ctx.Data(project)
	}
}

func Index(db *ent.Client) fir.RouteFunc {
	return func() fir.RouteOptions {
		return fir.RouteOptions{
			fir.ID("projects"),
			fir.Content("routes/projects/index.html"),
			fir.Layout("routes/layout.html"),
			fir.Partials("routes/partials"),
			fir.Partials("routes/projects/partials"),
			fir.OnLoad(loadProjects(db)),
			fir.OnEvent("create", createProject(db)),
			fir.OnEvent("query", loadProjects(db)),
		}
	}
}
