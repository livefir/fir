// Code generated (@generated) by entc, DO NOT EDIT.

package models

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/adnaan/autobahn/models/label"
	"github.com/adnaan/autobahn/models/predicate"
)

// LabelDelete is the builder for deleting a Label entity.
type LabelDelete struct {
	config
	hooks    []Hook
	mutation *LabelMutation
}

// Where appends a list predicates to the LabelDelete builder.
func (ld *LabelDelete) Where(ps ...predicate.Label) *LabelDelete {
	ld.mutation.Where(ps...)
	return ld
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (ld *LabelDelete) Exec(ctx context.Context) (int, error) {
	var (
		err      error
		affected int
	)
	if len(ld.hooks) == 0 {
		affected, err = ld.sqlExec(ctx)
	} else {
		var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
			mutation, ok := m.(*LabelMutation)
			if !ok {
				return nil, fmt.Errorf("unexpected mutation type %T", m)
			}
			ld.mutation = mutation
			affected, err = ld.sqlExec(ctx)
			mutation.done = true
			return affected, err
		})
		for i := len(ld.hooks) - 1; i >= 0; i-- {
			if ld.hooks[i] == nil {
				return 0, fmt.Errorf("models: uninitialized hook (forgotten import models/runtime?)")
			}
			mut = ld.hooks[i](mut)
		}
		if _, err := mut.Mutate(ctx, ld.mutation); err != nil {
			return 0, err
		}
	}
	return affected, err
}

// ExecX is like Exec, but panics if an error occurs.
func (ld *LabelDelete) ExecX(ctx context.Context) int {
	n, err := ld.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (ld *LabelDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := &sqlgraph.DeleteSpec{
		Node: &sqlgraph.NodeSpec{
			Table: label.Table,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeUUID,
				Column: label.FieldID,
			},
		},
	}
	if ps := ld.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return sqlgraph.DeleteNodes(ctx, ld.driver, _spec)
}

// LabelDeleteOne is the builder for deleting a single Label entity.
type LabelDeleteOne struct {
	ld *LabelDelete
}

// Exec executes the deletion query.
func (ldo *LabelDeleteOne) Exec(ctx context.Context) error {
	n, err := ldo.ld.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{label.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (ldo *LabelDeleteOne) ExecX(ctx context.Context) {
	ldo.ld.ExecX(ctx)
}
