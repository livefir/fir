package chirper

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/livefir/fir"
	"github.com/livefir/fir/internal/dev"
	"github.com/timshannon/bolthold"
)

// Chirp represents a single chirp message.
type Chirp struct {
	ID          uint64    `json:"id" boltholdKey:"ID"`
	Username    string    `json:"username"`
	Body        string    `json:"body"`
	LikesCount  int       `json:"likes_count"`
	RepostCount int       `json:"repost_count"`
	CreatedAt   time.Time `json:"created_at"`
}

// loadChirps loads the chirps from the database and returns an OnEventFunc
func loadChirps(db *bolthold.Store) fir.OnEventFunc {
	return func(ctx fir.RouteContext) error {
		var chirps []Chirp
		q := &bolthold.Query{}
		q = q.SortBy("CreatedAt").Reverse()
		if err := db.Find(&chirps, q); err != nil {
			return err
		}
		return ctx.Data(map[string]any{"chirps": chirps})
	}
}

// createChirp is a function that returns an OnEventFunc for creating a new chirp.
func createChirp(db *bolthold.Store) fir.OnEventFunc {
	return func(ctx fir.RouteContext) error {
		chirp := new(Chirp)
		if err := ctx.Bind(chirp); err != nil {
			return err
		}
		if len(chirp.Body) < 3 {
			return ctx.FieldError("body", errors.New("chirp is too short"))
		}
		chirp.CreatedAt = time.Now()
		if err := db.Insert(bolthold.NextSequence(), chirp); err != nil {
			return err
		}
		// simulate a delay
		time.Sleep(500 * time.Millisecond)
		return ctx.Data(chirp)
	}
}

// likeChirp increments the like count of a chirp and updates it in the database.
func likeChirp(db *bolthold.Store) fir.OnEventFunc {
	type likeReq struct {
		ChirpID uint64 `json:"chirpID"`
	}
	return func(ctx fir.RouteContext) error {
		req := new(likeReq)
		if err := ctx.Bind(req); err != nil {
			return err
		}
		var chirp Chirp
		if err := db.Get(req.ChirpID, &chirp); err != nil {
			return err
		}
		chirp.LikesCount++
		if err := db.Update(req.ChirpID, &chirp); err != nil {
			return err
		}
		return ctx.Data(chirp)
	}
}

// deleteChirp is a function that returns an OnEventFunc for deleting a chirp from the database.
func deleteChirp(db *bolthold.Store) fir.OnEventFunc {
	type deleteReq struct {
		ChirpID uint64 `json:"chirpID"`
	}
	return func(ctx fir.RouteContext) error {
		req := new(deleteReq)
		if err := ctx.Bind(req); err != nil {
			return err
		}

		if err := db.Delete(req.ChirpID, &Chirp{}); err != nil {
			return err
		}
		return nil
	}
}

// Index returns route options for the main chirper interface
func Index() fir.RouteOptions {
	db, err := bolthold.Open("chirper.db", 0666, nil)
	if err != nil {
		panic(err)
	}
	return fir.RouteOptions{
		fir.ID("index"),
		fir.Content("index.html"),
		fir.OnLoad(loadChirps(db)),
		fir.OnEvent("create-chirp", createChirp(db)),
		fir.OnEvent("delete-chirp", deleteChirp(db)),
		fir.OnEvent("like-chirp", likeChirp(db)),
	}
}

// NewRoute returns route options for e2e testing
func NewRoute() fir.RouteOptions {
	return Index()
}

// NoJSIndex returns route options for the no-JS version
func NoJSIndex() fir.RouteOptions {
	opts := Index()
	opts = append(opts,
		fir.ID("index-no-js"),
		fir.Content("index_no_js.html"))
	return opts
}

func Run(port int) error {
	dev.SetupAlpinePluginServer()
	db, err := bolthold.Open("chirper.db", 0666, nil)
	if err != nil {
		return err
	}

	controller := fir.NewController("app", fir.DevelopmentMode(true))
	http.Handle("/", controller.RouteFunc(func() fir.RouteOptions {
		return fir.RouteOptions{
			fir.ID("index"),
			fir.Content("index.html"),
			fir.OnLoad(loadChirps(db)),
			fir.OnEvent("create-chirp", createChirp(db)),
			fir.OnEvent("delete-chirp", deleteChirp(db)),
			fir.OnEvent("like-chirp", likeChirp(db)),
		}
	}))
	http.Handle("/nojs", controller.RouteFunc(func() fir.RouteOptions {
		opts := fir.RouteOptions{
			fir.ID("index"),
			fir.Content("index.html"),
			fir.OnLoad(loadChirps(db)),
			fir.OnEvent("create-chirp", createChirp(db)),
			fir.OnEvent("delete-chirp", deleteChirp(db)),
			fir.OnEvent("like-chirp", likeChirp(db)),
		}
		opts = append(opts,
			fir.ID("index-no-js"),
			fir.Content("index_no_js.html"))
		return opts
	}))
	log.Printf("Chirper example listening on http://localhost:%d", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
