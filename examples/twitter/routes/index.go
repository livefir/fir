package routes

import (
	"errors"
	"fmt"
	"time"

	"github.com/adnaan/fir"
	"github.com/adnaan/fir/patch"
	"github.com/timshannon/bolthold"
)

type Tweet struct {
	ID          uint64    `json:"id" boltholdKey:"ID"`
	Username    string    `json:"username"`
	Body        string    `json:"body"`
	LikesCount  int       `json:"likes_count"`
	RepostCount int       `json:"repost_count"`
	CreatedAt   time.Time `json:"created_at"`
}

func insertTweet(ctx fir.Context, db *bolthold.Store) (*Tweet, error) {
	tweet := new(Tweet)
	if err := ctx.Bind(tweet); err != nil {
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

func loadTweets(db *bolthold.Store) fir.OnEventFunc {
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
		return ctx.Append("#tweets", patch.Block("tweet", tweet))
	}
}

func deleteTweet(db *bolthold.Store) fir.OnEventFunc {
	type deleteReq struct {
		TweetID uint64 `json:"tweetID"`
	}
	return func(ctx fir.Context) error {
		req := new(deleteReq)
		if err := ctx.Bind(req); err != nil {
			return err
		}

		if err := db.Delete(req.TweetID, &Tweet{}); err != nil {
			return err
		}
		return ctx.Remove(fmt.Sprintf("#tweet-%d", req.TweetID))
	}
}

func Index(db *bolthold.Store) fir.RouteFunc {
	return func() fir.RouteOptions {
		return fir.RouteOptions{
			fir.ID("index"),
			fir.Content("routes/index.html"),
			fir.Layout("routes/layout.html"),
			fir.OnLoad(loadTweets(db)),
			fir.OnEvent("createTweet", createTweet(db)),
			fir.OnEvent("deleteTweet", deleteTweet(db)),
		}
	}
}
