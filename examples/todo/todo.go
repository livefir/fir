package todo

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/livefir/fir"
	"github.com/timshannon/bolthold"
)

type Todo struct {
	ID        uint64    `json:"id" boltholdKey:"ID"`
	Text      string    `json:"text"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
}

func insertTodo(ctx fir.RouteContext, db *bolthold.Store) (*Todo, error) {
	todo := new(Todo)
	if err := ctx.Bind(todo); err != nil {
		return nil, err
	}
	if len(todo.Text) < 3 {
		return nil, ctx.FieldError("text", errors.New("todo is too short"))
	}
	todo.CreatedAt = time.Now()
	if err := db.Insert(bolthold.NextSequence(), todo); err != nil {
		return nil, err
	}
	return todo, nil
}

type queryReq struct {
	Order  string `json:"order" schema:"order"`
	Search string `json:"search" schema:"search"`
	Offset int    `json:"offset" schema:"offset"`
	Limit  int    `json:"limit" schema:"limit"`
}

func load(db *bolthold.Store) fir.OnEventFunc {
	return func(ctx fir.RouteContext) error {
		var req queryReq
		if err := ctx.Bind(&req); err != nil {
			return err
		}
		var todos []Todo
		if err := db.Find(&todos, &bolthold.Query{}); err != nil {
			return err
		}
		return ctx.Data(map[string]any{"todos": todos})
	}
}

func createTodo(db *bolthold.Store) fir.OnEventFunc {
	return func(ctx fir.RouteContext) error {
		todo, err := insertTodo(ctx, db)
		if err != nil {
			return err
		}
		return ctx.Data(todo)
	}
}

func updateTodo(db *bolthold.Store) fir.OnEventFunc {
	type updateReq struct {
		TodoID uint64 `json:"todoID"`
		Text   string `json:"text"`
	}
	return func(ctx fir.RouteContext) error {
		req := new(updateReq)
		if err := ctx.Bind(req); err != nil {
			return err
		}
		var todo Todo
		if err := db.Get(req.TodoID, &todo); err != nil {
			return err
		}
		todo.Text = req.Text
		if err := db.Update(req.TodoID, &todo); err != nil {
			return err
		}
		return ctx.Data(todo)
	}
}

func toggleDone(db *bolthold.Store) fir.OnEventFunc {
	type doneReq struct {
		TodoID uint64 `json:"todoID"`
	}
	return func(ctx fir.RouteContext) error {
		req := new(doneReq)
		if err := ctx.Bind(req); err != nil {
			return err
		}
		var todo Todo
		if err := db.Get(req.TodoID, &todo); err != nil {
			return err
		}
		todo.Done = !todo.Done
		if err := db.Update(req.TodoID, &todo); err != nil {
			return err
		}
		return ctx.Data(todo)
	}
}

func deleteTodo(db *bolthold.Store) fir.OnEventFunc {
	type deleteReq struct {
		TodoID uint64 `json:"todoID"`
	}
	return func(ctx fir.RouteContext) error {
		req := new(deleteReq)
		if err := ctx.Bind(req); err != nil {
			return err
		}

		if err := db.Delete(req.TodoID, &Todo{}); err != nil {
			return err
		}
		return nil
	}
}

func Index(db *bolthold.Store) fir.RouteFunc {
	return func() fir.RouteOptions {
		return fir.RouteOptions{
			fir.ID("todos"),
			fir.Content("todo.html"),
			fir.OnLoad(load(db)),
			fir.OnEvent("create", createTodo(db)),
			fir.OnEvent("delete", deleteTodo(db)),
			fir.OnEvent("toggle-done", toggleDone(db)),
		}
	}
}

func db() *bolthold.Store {
	dbfile, err := os.CreateTemp("", "todos")
	if err != nil {
		panic(err)
	}
	db, err := bolthold.Open(dbfile.Name(), 0666, nil)
	if err != nil {
		panic(fmt.Errorf("failed to open database: %w", err))
	}
	return db
}

func NewRoute() fir.RouteOptions {
	return Index(db())()
}

func Run(port int) error {

	c := fir.NewController("fir-todo", fir.DevelopmentMode(true))
	http.Handle("/", c.RouteFunc(Index(db())))
	log.Printf("Todo example listening on http://localhost:%d", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
