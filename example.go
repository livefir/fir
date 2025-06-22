package fir

import (
	"errors"
	"time"

	"github.com/timshannon/bolthold"
)

type Todo struct {
	ID        uint64    `json:"id" boltholdKey:"ID"`
	Text      string    `json:"text"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
}

func insertTodo(ctx RouteContext, db *bolthold.Store) (*Todo, error) {
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

func load(db *bolthold.Store) OnEventFunc {
	return func(ctx RouteContext) error {
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

func createTodo(db *bolthold.Store) OnEventFunc {
	return func(ctx RouteContext) error {
		todo, err := insertTodo(ctx, db)
		if err != nil {
			return err
		}
		return ctx.Data(todo)
	}
}

func toggleDone(db *bolthold.Store) OnEventFunc {
	type doneReq struct {
		TodoID uint64 `json:"todoID"`
	}
	return func(ctx RouteContext) error {
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

func deleteTodo(db *bolthold.Store) OnEventFunc {
	type deleteReq struct {
		TodoID uint64 `json:"todoID"`
	}
	return func(ctx RouteContext) error {
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

func todosRoute(db *bolthold.Store) RouteFunc {
	return func() RouteOptions {
		return RouteOptions{
			ID("todos"),
			Content("example.html"),
			OnLoad(load(db)),
			OnEvent("create", createTodo(db)),
			OnEvent("delete", deleteTodo(db)),
			OnEvent("toggle-done", toggleDone(db)),
		}
	}
}
