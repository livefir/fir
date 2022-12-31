package projects

import (
	"errors"

	"github.com/livefir/fir"
	"github.com/livefir/fir/examples/fira/ent"
	"github.com/livefir/fir/examples/fira/ent/project"
)

var defaultPageSize = 5

type queryReq struct {
	Order           string `json:"order"`
	Search          string `json:"search"`
	Offset          int    `json:"offset"`
	Limit           int    `json:"limit"`
	pageSize        int
	defaultPageSize int
}

func projectQuery(db *ent.Client, req queryReq) *ent.ProjectQuery {
	if req.Limit == 0 {
		req.Limit = defaultPageSize
	}
	q := db.Project.Query().Offset(req.Offset).Limit(req.Limit)

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
	prev := req.Offset - defaultPageSize
	hasPrevious := true
	if prev < 0 || req.Offset == 0 {
		hasPrevious = false
	}
	next := defaultPageSize + req.Offset
	hasNext := true
	if req.pageSize < defaultPageSize {
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

func loadProjects(db *ent.Client) fir.OnEventFunc {
	return func(ctx fir.RouteContext) error {
		var q queryReq
		if err := ctx.Bind(&q); err != nil {
			return err
		}

		projects, err := projectQuery(db, q).All(ctx.Request().Context())
		if err != nil {
			return err
		}

		q.pageSize = len(projects)
		q.defaultPageSize = defaultPageSize

		data := map[string]any{"projects": projects}
		return ctx.Data(data, paginationData(q))
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
			fir.Partials("routes/partials", "routes/projects/partials"),
			fir.OnLoad(loadProjects(db)),
			fir.OnEvent("create-project", createProject(db)),
		}
	}
}
