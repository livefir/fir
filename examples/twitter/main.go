package main

import (
	"net/http"

	"github.com/adnaan/fir"
	"github.com/adnaan/fir/examples/twitter/routes/app"
	"github.com/timshannon/bolthold"
)

func main() {

	db, err := bolthold.Open("todos.db", 0666, nil)
	if err != nil {
		panic(err)
	}

	controller := fir.NewController("app", fir.DevelopmentMode(true))
	http.Handle("/", controller.RouteFunc(app.Route(db)))
	http.ListenAndServe(":9867", nil)
}
