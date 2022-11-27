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
	Body        string    `json:"body" schema:"body"`
	LikesCount  int       `json:"likes_count"`
	RepostCount int       `json:"repost_count"`
	CreatedAt   time.Time `json:"created_at"`
}

func listTweets(r fir.RouteRenderer, db *bolthold.Store) error {
	var tweets []Tweet
	if err := db.Find(&tweets, &bolthold.Query{}); err != nil {
		return fmt.Errorf("error loading tweets: %v", err)
	}
	return r(fir.M{"tweets": tweets})
}

func insertTweet(e fir.Event, db *bolthold.Store) (*Tweet, error) {
	tweet := new(Tweet)
	if err := e.DecodeParams(tweet); err != nil {
		return nil, err
	}

	tweet.CreatedAt = time.Now()
	if err := db.Insert(bolthold.NextSequence(), tweet); err != nil {
		return nil, err
	}
	return tweet, nil
}

func load(db *bolthold.Store) fir.OnLoadFunc {
	return func(e fir.Event, r fir.RouteRenderer) error {
		return listTweets(r, db)
	}
}

func createTweetForm(db *bolthold.Store) fir.OnFormFunc {
	return func(e fir.Event, r fir.RouteRenderer) error {
		_, err := insertTweet(e, db)
		if err != nil {
			return err
		}
		return listTweets(r, db)
	}
}

func createTweetEvent(db *bolthold.Store) fir.OnEventFunc {
	return func(e fir.Event, r fir.PatchRenderer) error {
		tweet, err := insertTweet(e, db)
		if err != nil {
			return err
		}
		return r(fir.Append("#tweets", fir.Block("tweet", tweet)))
	}
}

func Route(db *bolthold.Store) fir.RouteFunc {
	return func() []fir.RouteOption {
		return []fir.RouteOption{
			fir.ID("app"),
			fir.Content("routes/app/page.html"),
			fir.Layout("routes/layout.html"),
			fir.OnLoad(load(db)),
			fir.OnForm("default", createTweetForm(db)),
			fir.OnEvent("createTweet", createTweetEvent(db)),
		}
	}
}
