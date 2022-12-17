package main

import (
	"net/http"

	"github.com/livefir/fir"
	"github.com/livefir/fir/examples/twitter/routes"
	"github.com/timshannon/bolthold"
)

func main() {

	db, err := bolthold.Open("todos.db", 0666, nil)
	if err != nil {
		panic(err)
	}

	controller := fir.NewController("app", fir.DevelopmentMode(true))
	http.Handle("/", controller.RouteFunc(routes.Index(db)))
	http.ListenAndServe(":9867", nil)
}
