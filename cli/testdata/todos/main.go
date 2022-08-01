package main

import (
	"context"
	"log"
	"net/http"

	"github.com/adnaan/fir"
	"github.com/adnaan/fir/cli/testdata/todos/models"
	boardsIndex "github.com/adnaan/fir/cli/testdata/todos/views/boards/index"
	boardsShow "github.com/adnaan/fir/cli/testdata/todos/views/boards/show"
	todosIndex "github.com/adnaan/fir/cli/testdata/todos/views/todos/index"
	todosShow "github.com/adnaan/fir/cli/testdata/todos/views/todos/show"
	"github.com/go-chi/chi/v5"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	log.Println("starting server...")
	db, err := models.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	// Run the auto migration tool.
	if err := db.Schema.Create(ctx); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	c := fir.NewController("todos", fir.DevelopmentMode(true))
	r := chi.NewRouter()
	r.Handle("/", c.Handler(&boardsIndex.View{DB: db}))
	r.Handle("/{boardID}/show", c.Handler(&boardsShow.View{DB: db}))
	r.Handle("/{boardID}/todos", c.Handler(&todosIndex.View{DB: db}))
	r.Handle("/{boardID}/todos/{todoID}/show", c.Handler(&todosShow.View{DB: db}))

	log.Println("listening on http://localhost:9867")
	http.ListenAndServe(":9867", r)
}
