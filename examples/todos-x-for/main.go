package main

import (
	"encoding/json"
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

func (t *TodosView) OnGet(_ http.ResponseWriter, _ *http.Request) fir.Page {
	todos := make([]Todo, 0)
	if err := t.db.Find(&todos, &bolthold.Query{}); err != nil {
		return fir.Page{}
	}
	b, _ := json.Marshal(todos)
	return fir.Page{Data: fir.Data{"todos": string(b)}}
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
	todos := make([]Todo, 0) // important: initialise the array to return [] instead of null as a json response
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
		fir.Store{
			Name: "todos",
			Data: todos,
		},
	}
}

func main() {
	db, err := bolthold.Open("todos.db", 0666, nil)
	if err != nil {
		panic(err)
	}
	c := fir.NewController("fir-todos-x-for", fir.DevelopmentMode(true))
	http.Handle("/", c.Handler(NewTodosView(db)))
	log.Println("listening on http://localhost:9867")
	http.ListenAndServe(":9867", nil)
}
