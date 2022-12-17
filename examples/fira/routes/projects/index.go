package projects

import (
	"github.com/adnaan/fir"
	"github.com/adnaan/fir/examples/fira/ent"
	"github.com/adnaan/fir/examples/fira/ent/project"
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

func createProject(db *ent.Client) fir.OnEventFunc {
	return func(ctx fir.RouteContext) error {
		return ctx.Data(map[string]any{})
	}
}

func Index(db *ent.Client) fir.RouteFunc {
	return func() fir.RouteOptions {
		return fir.RouteOptions{
			fir.ID("projects"),
			fir.Content("routes/projects/index.html"),
			fir.Layout("routes/layout.html"),
			fir.Partials("routes/partials"),
			fir.OnLoad(loadProjects(db)),
			fir.OnEvent("createProject", createProject(db)),
		}
	}
}
