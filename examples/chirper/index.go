package main

import (
	"errors"
	"time"

	"github.com/livefir/fir"
	"github.com/timshannon/bolthold"
)

type Chirp struct {
	ID          uint64    `json:"id" boltholdKey:"ID"`
	Username    string    `json:"username"`
	Body        string    `json:"body"`
	LikesCount  int       `json:"likes_count"`
	RepostCount int       `json:"repost_count"`
	CreatedAt   time.Time `json:"created_at"`
}

func insertChirp(ctx fir.RouteContext, db *bolthold.Store) (*Chirp, error) {
	chirp := new(Chirp)
	if err := ctx.Bind(chirp); err != nil {
		return nil, err
	}
	if len(chirp.Body) < 3 {
		return nil, ctx.FieldError("body", errors.New("chirp is too short"))
	}
	chirp.CreatedAt = time.Now()
	if err := db.Insert(bolthold.NextSequence(), chirp); err != nil {
		return nil, err
	}
	return chirp, nil
}

type queryReq struct {
	Order  string `json:"order"`
	Search string `json:"search"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
}

func loadChirps(db *bolthold.Store) fir.OnEventFunc {
	return func(ctx fir.RouteContext) error {
		var req queryReq
		if err := ctx.Bind(&req); err != nil {
			return err
		}
		var chirps []Chirp
		q := &bolthold.Query{}
		q = q.SortBy("CreatedAt").Reverse()
		if err := db.Find(&chirps, q); err != nil {
			return err
		}
		return ctx.Data(map[string]any{"chirps": chirps})
	}
}

func createChirp(db *bolthold.Store) fir.OnEventFunc {
	return func(ctx fir.RouteContext) error {
		chirp, err := insertChirp(ctx, db)
		if err != nil {
			return err
		}
		// simulate a delay
		time.Sleep(500 * time.Millisecond)
		return ctx.Data(chirp)
	}
}

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

func Index(db *bolthold.Store) fir.RouteFunc {
	return func() fir.RouteOptions {
		return fir.RouteOptions{
			fir.ID("index"),
			fir.Content("index.html"),
			fir.OnLoad(loadChirps(db)),
			fir.OnEvent("create-chirp", createChirp(db)),
			fir.OnEvent("delete-chirp", deleteChirp(db)),
			fir.OnEvent("like-chirp", likeChirp(db)),
		}
	}
}
