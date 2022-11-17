package main

import (
	"context"
	"embed"
	"log"
	"net/http"
	"os"

	"github.com/adnaan/autobahn/models"
	boardsIndex "github.com/adnaan/autobahn/views/boards/index"
	boardsShow "github.com/adnaan/autobahn/views/boards/show"
	storiesIndex "github.com/adnaan/autobahn/views/stories/index"
	storiesShow "github.com/adnaan/autobahn/views/stories/show"

	"github.com/adnaan/fir"
	"github.com/go-chi/chi/v5"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed public
var publicFS embed.FS

func main() {

	// db, err := models.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	db, err := models.Open("sqlite3", "file:autobahn.db?cache=shared&_fk=1")
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	// Run the auto migration tool.
	if err := db.Schema.Create(ctx); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	publicDir := "."
	var opts []fir.Option
	if os.Getenv("ENV") == "production" {
		publicDir = "public"
		opts = append(opts, fir.WithEmbedFS(publicFS))
	} else {
		opts = append(opts, fir.DevelopmentMode(true))
	}

	opts = append(opts, fir.PublicDir(publicDir))

	c := fir.NewController("autobahn", opts...)
	r := chi.NewRouter()
	r.Handle("/", c.Handler(&boardsIndex.View{DB: db}))
	r.Handle("/{boardID}/show", c.Handler(&boardsShow.View{DB: db}))
	r.Handle("/{boardID}/stories", c.Handler(&storiesIndex.View{DB: db}))
	r.Handle("/{boardID}/stories/{storyID}/show", c.Handler(&storiesShow.View{DB: db}))

	log.Println("listening on http://localhost:9867")
	http.ListenAndServe(":9867", r)
}
