package main

import (
	"log"
	"net/http"
	"time"

	pwc "github.com/adnaan/pineview/controller"
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
	pwc.DefaultView
	db *bolthold.Store
}

func (t *TodosView) Content() string {
	return "app.html"
}

func (t *TodosView) Partials() []string {
	return []string{"todos.html"}
}

func (t *TodosView) OnMount(_ http.ResponseWriter, _ *http.Request) (pwc.Status, pwc.M) {
	var todos []Todo
	if err := t.db.Find(&todos, &bolthold.Query{}); err != nil {
		return pwc.Status{
			Code: 200,
		}, nil
	}
	return pwc.Status{Code: 200}, pwc.M{"todos": todos}
}

func (t *TodosView) OnLiveEvent(ctx pwc.Context) error {
	var todo Todo
	if err := ctx.Event().DecodeParams(&todo); err != nil {
		return err
	}

	switch ctx.Event().ID {

	case "todos/new":
		if len(todo.Text) < 4 {
			ctx.Store("formData").UpdateProp("textError", "Min length is 4")
			return nil
		}
		ctx.Store("formData").UpdateProp("textError", "")
		if err := t.db.Insert(bolthold.NextSequence(), &todo); err != nil {
			return err
		}
	case "todos/del":
		if err := t.db.Delete(todo.ID, &todo); err != nil {
			return err
		}
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", ctx.Event())
	}
	// list updated todos
	var todos []Todo
	if err := t.db.Find(&todos, &bolthold.Query{}); err != nil {
		return err
	}
	ctx.Morph("#todos", "todos", pwc.M{"todos": todos})
	return nil
}

func main() {
	db, err := bolthold.Open("todos.db", 0666, nil)
	if err != nil {
		panic(err)
	}
	glvc := pwc.Websocket("pineview-todos", pwc.DevelopmentMode(true))
	http.Handle("/", glvc.Handler(NewTodosView(db)))
	log.Println("listening on http://localhost:9867")
	http.ListenAndServe(":9867", nil)
}
