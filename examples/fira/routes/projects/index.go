package projects

import (
	"errors"

	"github.com/livefir/fir"
	"github.com/livefir/fir/examples/fira/ent"
	"github.com/livefir/fir/examples/fira/ent/project"
)

var defaultPageSize = 5

type queryReq struct {
	Order  string `json:"order"`
	Search string `json:"search"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
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

func paginationData(req queryReq, boardLen int) map[string]any {
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

		data := map[string]any{"projects": projects}
		for k, v := range paginationData(q, len(projects)) {
			data[k] = v
		}
		return ctx.Data(data)
	}
}

type createReq struct {
	Title       string `json:"title" schema:"title,required"`
	Description string `json:"description" schema:"description,required"`
}

func createProject(db *ent.Client) fir.OnEventFunc {
	return func(ctx fir.RouteContext) error {
		var req createReq
		if err := ctx.Bind(&req); err != nil {
			return err
		}
		if len(req.Title) < 3 {
			return ctx.FieldError("title", errors.New("title is too short"))
		}
		if len(req.Description) < 3 {
			return ctx.FieldError("description", errors.New("description is too short"))
		}
		project, err := db.Project.
			Create().
			SetTitle(req.Title).
			SetDescription(req.Description).
			Save(ctx.Request().Context())
		if err != nil {
			return err
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
