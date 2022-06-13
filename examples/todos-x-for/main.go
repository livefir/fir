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

func (t *TodosView) OnRequest(_ http.ResponseWriter, _ *http.Request) (fir.Status, fir.Data) {
	todos := make([]Todo, 0)
	if err := t.db.Find(&todos, &bolthold.Query{}); err != nil {
		return fir.Status{
			Code: 200,
		}, nil
	}
	b, _ := json.Marshal(todos)
	return fir.Status{Code: 200}, fir.Data{"todos": string(b)}
}

func (t *TodosView) OnEvent(s fir.Socket) error {
	var todo Todo
	if err := s.Event().DecodeParams(&todo); err != nil {
		return err
	}

	switch s.Event().ID {

	case "todos/new":
		if len(todo.Text) < 4 {
			s.Store("formData").UpdateProp("textError", "Min length is 4")
			return nil
		}
		s.Store("formData").UpdateProp("textError", "")
		if err := t.db.Insert(bolthold.NextSequence(), &todo); err != nil {
			return err
		}
	case "todos/del":
		if err := t.db.Delete(todo.ID, &todo); err != nil {
			return err
		}
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", s.Event())
	}
	// list updated todos
	todos := make([]Todo, 0) // important: initialise the array to return [] instead of null as a json response
	if err := t.db.Find(&todos, &bolthold.Query{}); err != nil {
		return err
	}
	s.Store("todos").Update(todos)
	return nil
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