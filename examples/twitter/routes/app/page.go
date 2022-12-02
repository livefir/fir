package app

import (
	"errors"
	"fmt"
	"time"

	"github.com/adnaan/fir"
	"github.com/timshannon/bolthold"
)

type Tweet struct {
	ID          uint64    `json:"id" boltholdKey:"ID"`
	Username    string    `json:"username"`
	Body        string    `json:"body" validate:"min=3"`
	LikesCount  int       `json:"likes_count"`
	RepostCount int       `json:"repost_count"`
	CreatedAt   time.Time `json:"created_at"`
}

func insertTweet(ctx fir.Context, db *bolthold.Store) (*Tweet, error) {
	tweet := new(Tweet)
	if err := ctx.DecodeParams(tweet); err != nil {
		return nil, err
	}
	if len(tweet.Body) < 3 {
		return nil, ctx.FieldError("body", errors.New("tweet is too short"))
	}
	tweet.CreatedAt = time.Now()
	if err := db.Insert(bolthold.NextSequence(), tweet); err != nil {
		return nil, err
	}
	return tweet, nil
}

func load(db *bolthold.Store) fir.OnEventFunc {
	return func(ctx fir.Context) error {
		var tweets []Tweet
		if err := db.Find(&tweets, &bolthold.Query{}); err != nil {
			return err
		}
		return ctx.KV("tweets", tweets)
	}
}

func createTweet(db *bolthold.Store) fir.OnEventFunc {
	return func(ctx fir.Context) error {
		tweet, err := insertTweet(ctx, db)
		if err != nil {
			return err
		}
		return ctx.Append("#tweets", fir.Block("tweet", tweet))
	}
}

func deleteTweet(db *bolthold.Store) fir.OnEventFunc {
	type deleteReq struct {
		TweetID uint64 `json:"tweetID"`
	}
	return func(ctx fir.Context) error {
		req := new(deleteReq)
		if err := ctx.DecodeParams(req); err != nil {
			return err
		}

		if err := db.Delete(req.TweetID, &Tweet{}); err != nil {
			return err
		}
		return ctx.Remove(fmt.Sprintf("#tweet-%d", req.TweetID))
	}
}

func Route(db *bolthold.Store) fir.RouteFunc {
	return func() fir.RouteOptions {
		return fir.RouteOptions{
			fir.ID("app"),
			fir.Content("routes/app/page.html"),
			fir.Layout("routes/layout.html"),
			fir.OnLoad(load(db)),
			fir.OnEvent("createTweet", createTweet(db)),
			fir.OnEvent("deleteTweet", deleteTweet(db)),
		}
	}
}
