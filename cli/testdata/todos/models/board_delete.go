// Code generated (@generated) by entc, DO NOT EDIT.

package models

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/adnaan/fir/cli/testdata/todos/models/board"
	"github.com/adnaan/fir/cli/testdata/todos/models/predicate"
)

// BoardDelete is the builder for deleting a Board entity.
type BoardDelete struct {
	config
	hooks    []Hook
	mutation *BoardMutation
}

// Where appends a list predicates to the BoardDelete builder.
func (bd *BoardDelete) Where(ps ...predicate.Board) *BoardDelete {
	bd.mutation.Where(ps...)
	return bd
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (bd *BoardDelete) Exec(ctx context.Context) (int, error) {
	var (
		err      error
		affected int
	)
	if len(bd.hooks) == 0 {
		affected, err = bd.sqlExec(ctx)
	} else {
		var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
			mutation, ok := m.(*BoardMutation)
			if !ok {
				return nil, fmt.Errorf("unexpected mutation type %T", m)
			}
			bd.mutation = mutation
			affected, err = bd.sqlExec(ctx)
			mutation.done = true
			return affected, err
		})
		for i := len(bd.hooks) - 1; i >= 0; i-- {
			if bd.hooks[i] == nil {
				return 0, fmt.Errorf("models: uninitialized hook (forgotten import models/runtime?)")
			}
			mut = bd.hooks[i](mut)
		}
		if _, err := mut.Mutate(ctx, bd.mutation); err != nil {
			return 0, err
		}
	}
	return affected, err
}

// ExecX is like Exec, but panics if an error occurs.
func (bd *BoardDelete) ExecX(ctx context.Context) int {
	n, err := bd.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (bd *BoardDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := &sqlgraph.DeleteSpec{
		Node: &sqlgraph.NodeSpec{
			Table: board.Table,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeUUID,
				Column: board.FieldID,
			},
		},
	}
	if ps := bd.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return sqlgraph.DeleteNodes(ctx, bd.driver, _spec)
}

// BoardDeleteOne is the builder for deleting a single Board entity.
type BoardDeleteOne struct {
	bd *BoardDelete
}

// Exec executes the deletion query.
func (bdo *BoardDeleteOne) Exec(ctx context.Context) error {
	n, err := bdo.bd.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{board.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (bdo *BoardDeleteOne) ExecX(ctx context.Context) {
	bdo.bd.ExecX(ctx)
}
