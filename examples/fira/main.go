package main

import (
	"context"
	"log"
	"net/http"

	"github.com/livefir/fir"
	"github.com/livefir/fir/examples/fira/ent"
	projects "github.com/livefir/fir/examples/fira/routes/projects"
)

func main() {
	db, err := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	// db, err := ent.Open("sqlite3", "file:autobahn.db?cache=shared&_fk=1")
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	// Run the auto migration tool.
	if err := db.Schema.Create(ctx); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	controller := fir.NewController("app", fir.DevelopmentMode(true))
	http.Handle("/", controller.RouteFunc(projects.Index(db)))
	http.Handle("/{id}/show", controller.RouteFunc(projects.Show(db)))
	http.ListenAndServe(":9867", nil)
}
