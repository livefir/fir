package main

import (
	"log"
	"net/http"
	"time"

	"github.com/adnaan/fir"
	"github.com/timshannon/bolthold"
)

type Todo struct {
	ID        uint64 `json:"id" boltholdKey:"ID"`
	Text      string `json:"text"`
	Done      bool   `json:"done"`
	CreatedAt time.Time
}

func NewTodosView(db *bolthold.Store) *TodosView {
	return &TodosView{
		db: db,
	}
}

type TodosView struct {
	fir.DefaultView
	db *bolthold.Store
}

func (t *TodosView) Content() string {
	return "app.html"
}

func (t *TodosView) Partials() []string {
	return []string{"todos.html"}
}

func (t *TodosView) OnRequest(_ http.ResponseWriter, _ *http.Request) (fir.Status, fir.Data) {
	var todos []Todo
	if err := t.db.Find(&todos, &bolthold.Query{}); err != nil {
		return fir.Status{
			Code: 200,
		}, nil
	}
	return fir.Status{Code: 200}, fir.Data{"todos": todos}
}

func (t *TodosView) OnEvent(event fir.Event) fir.Patchset {
	var todo Todo
	if err := event.DecodeParams(&todo); err != nil {
		return nil
	}

	switch event.ID {

	case "todos/new":
		if len(todo.Text) < 4 {
			return fir.Patchset{
				fir.Store{
					Name: "formData",
					Data: map[string]any{
						"textError": "Min length is 4",
					},
				},
			}
		}
		if err := t.db.Insert(bolthold.NextSequence(), &todo); err != nil {
			return nil
		}
	case "todos/del":
		if err := t.db.Delete(todo.ID, &todo); err != nil {
			return nil
		}
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", event)
	}
	// list updated todos
	var todos []Todo
	if err := t.db.Find(&todos, &bolthold.Query{}); err != nil {
		return nil
	}

	return fir.Patchset{
		fir.Store{
			Name: "formData",
			Data: map[string]any{
				"textError": "",
			},
		},
		fir.Morph{
			Template: "todos",
			Selector: "#todos",
			Data:     fir.Data{"todos": todos},
		},
	}
}

func main() {
	db, err := bolthold.Open("todos.db", 0666, nil)
	if err != nil {
		panic(err)
	}
	c := fir.NewController("fir-todos", fir.DevelopmentMode(true))
	http.Handle("/", c.Handler(NewTodosView(db)))
	log.Println("listening on http://localhost:9867")
	http.ListenAndServe(":9867", nil)
}
