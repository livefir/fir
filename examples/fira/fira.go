package fira

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/livefir/fir"
	"github.com/livefir/fir/examples/fira/ent"
	projects "github.com/livefir/fir/examples/fira/routes/projects"
	_ "github.com/mattn/go-sqlite3"
)

func Index() fir.RouteOptions {
	// For e2e testing, use in-memory database
	db, err := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	if err := db.Schema.Create(ctx); err != nil {
		panic(err)
	}

	return projects.Index(db)()
}

func NewRoute() fir.RouteOptions {
	return Index()
}

func Run(port int) error {
	// For standalone running, use persistent database
	db, err := ent.Open("sqlite3", "file:fira.db?cache=shared&_fk=1")
	if err != nil {
		return fmt.Errorf("failed opening connection to sqlite: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.Schema.Create(ctx); err != nil {
		return fmt.Errorf("failed creating schema resources: %v", err)
	}

	pathParamsOpt := fir.WithPathParamsFunc(
		func(r *http.Request) fir.PathParams {
			return fir.PathParams{
				"id": chi.URLParam(r, "id"),
			}
		})

	controller := fir.NewController("fira", fir.DevelopmentMode(true), pathParamsOpt)
	r := chi.NewRouter()
	r.Handle("/", controller.RouteFunc(projects.Index(db)))
	r.Handle("/{id}/show", controller.RouteFunc(projects.Show(db)))
	log.Printf("Fira example listening on http://localhost:%d", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), r)
}
