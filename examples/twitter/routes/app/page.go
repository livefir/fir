package app

import (
	"fmt"
	"time"

	"github.com/adnaan/fir"
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
	if err := ctx.DecodeParams(tweet); err != nil {
		return nil, err
	}
	if len(tweet.Body) < 3 {
		return nil, fmt.Errorf("Tweet is too short")
	}
	tweet.CreatedAt = time.Now()
	if err := db.Insert(bolthold.NextSequence(), tweet); err != nil {
		return nil, err
	}
	return tweet, nil
}

func createTweetForm(db *bolthold.Store) fir.OnEventFunc {
	return func(ctx fir.Context) error {
		_, err := insertTweet(ctx, db)
		if err != nil {
			return err
		}
		var tweets []Tweet
		if err := db.Find(&tweets, &bolthold.Query{}); err != nil {
			return err
		}
		return ctx.Data(fir.M{"tweets": tweets})
	}
}

func createTweetEvent(db *bolthold.Store) fir.OnEventFunc {
	return func(ctx fir.Context) error {
		tweet, err := insertTweet(ctx, db)
		if err != nil {
			return err
		}
		return ctx.Patch(fir.Append("#tweets", fir.Block("tweet", tweet)))
	}
}

func load(db *bolthold.Store) fir.OnEventFunc {
	return func(ctx fir.Context) error {
		var tweets []Tweet
		if err := db.Find(&tweets, &bolthold.Query{}); err != nil {
			return err
		}
		return ctx.Data(fir.M{"tweets": tweets})
	}
}

func Route(db *bolthold.Store) fir.RouteFunc {
	return func() fir.RouteOptions {
		return fir.RouteOptions{
			fir.ID("app"),
			fir.Content("routes/app/page.html"),
			fir.Layout("routes/layout.html"),
			fir.OnLoad(load(db)),
			fir.OnForm("createTweet", createTweetForm(db)),
			fir.OnEvent("createTweet", createTweetEvent(db)),
		}
	}
}
