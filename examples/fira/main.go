package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/livefir/fir"
	"github.com/livefir/fir/examples/fira/ent"
	projects "github.com/livefir/fir/examples/fira/routes/projects"
	_ "github.com/mattn/go-sqlite3"
)

func main() {

	// db, err := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	db, err := ent.Open("sqlite3", "file:fira.db?cache=shared&_fk=1")
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	// Run the auto migration tool.
	if err := db.Schema.Create(ctx); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
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
	http.ListenAndServe(":9867", r)
}
