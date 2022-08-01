// Code generated (@generated) by entc, DO NOT EDIT.

package models

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/adnaan/fir/cli/testdata/todos/models/board"
	"github.com/adnaan/fir/cli/testdata/todos/models/todo"
	"github.com/google/uuid"
)

// TodoCreate is the builder for creating a Todo entity.
type TodoCreate struct {
	config
	mutation *TodoMutation
	hooks    []Hook
}

// SetCreateTime sets the "create_time" field.
func (tc *TodoCreate) SetCreateTime(t time.Time) *TodoCreate {
	tc.mutation.SetCreateTime(t)
	return tc
}

// SetNillableCreateTime sets the "create_time" field if the given value is not nil.
func (tc *TodoCreate) SetNillableCreateTime(t *time.Time) *TodoCreate {
	if t != nil {
		tc.SetCreateTime(*t)
	}
	return tc
}

// SetUpdateTime sets the "update_time" field.
func (tc *TodoCreate) SetUpdateTime(t time.Time) *TodoCreate {
	tc.mutation.SetUpdateTime(t)
	return tc
}

// SetNillableUpdateTime sets the "update_time" field if the given value is not nil.
func (tc *TodoCreate) SetNillableUpdateTime(t *time.Time) *TodoCreate {
	if t != nil {
		tc.SetUpdateTime(*t)
	}
	return tc
}

// SetTitle sets the "title" field.
func (tc *TodoCreate) SetTitle(s string) *TodoCreate {
	tc.mutation.SetTitle(s)
	return tc
}

// SetDescription sets the "description" field.
func (tc *TodoCreate) SetDescription(s string) *TodoCreate {
	tc.mutation.SetDescription(s)
	return tc
}

// SetID sets the "id" field.
func (tc *TodoCreate) SetID(u uuid.UUID) *TodoCreate {
	tc.mutation.SetID(u)
	return tc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (tc *TodoCreate) SetNillableID(u *uuid.UUID) *TodoCreate {
	if u != nil {
		tc.SetID(*u)
	}
	return tc
}

// SetOwnerID sets the "owner" edge to the Board entity by ID.
func (tc *TodoCreate) SetOwnerID(id uuid.UUID) *TodoCreate {
	tc.mutation.SetOwnerID(id)
	return tc
}

// SetNillableOwnerID sets the "owner" edge to the Board entity by ID if the given value is not nil.
func (tc *TodoCreate) SetNillableOwnerID(id *uuid.UUID) *TodoCreate {
	if id != nil {
		tc = tc.SetOwnerID(*id)
	}
	return tc
}

// SetOwner sets the "owner" edge to the Board entity.
func (tc *TodoCreate) SetOwner(b *Board) *TodoCreate {
	return tc.SetOwnerID(b.ID)
}

// Mutation returns the TodoMutation object of the builder.
func (tc *TodoCreate) Mutation() *TodoMutation {
	return tc.mutation
}

// Save creates the Todo in the database.
func (tc *TodoCreate) Save(ctx context.Context) (*Todo, error) {
	var (
		err  error
		node *Todo
	)
	tc.defaults()
	if len(tc.hooks) == 0 {
		if err = tc.check(); err != nil {
			return nil, err
		}
		node, err = tc.sqlSave(ctx)
	} else {
		var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
			mutation, ok := m.(*TodoMutation)
			if !ok {
				return nil, fmt.Errorf("unexpected mutation type %T", m)
			}
			if err = tc.check(); err != nil {
				return nil, err
			}
			tc.mutation = mutation
			if node, err = tc.sqlSave(ctx); err != nil {
				return nil, err
			}
			mutation.id = &node.ID
			mutation.done = true
			return node, err
		})
		for i := len(tc.hooks) - 1; i >= 0; i-- {
			if tc.hooks[i] == nil {
				return nil, fmt.Errorf("models: uninitialized hook (forgotten import models/runtime?)")
			}
			mut = tc.hooks[i](mut)
		}
		if _, err := mut.Mutate(ctx, tc.mutation); err != nil {
			return nil, err
		}
	}
	return node, err
}

// SaveX calls Save and panics if Save returns an error.
func (tc *TodoCreate) SaveX(ctx context.Context) *Todo {
	v, err := tc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (tc *TodoCreate) Exec(ctx context.Context) error {
	_, err := tc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (tc *TodoCreate) ExecX(ctx context.Context) {
	if err := tc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (tc *TodoCreate) defaults() {
	if _, ok := tc.mutation.CreateTime(); !ok {
		v := todo.DefaultCreateTime()
		tc.mutation.SetCreateTime(v)
	}
	if _, ok := tc.mutation.UpdateTime(); !ok {
		v := todo.DefaultUpdateTime()
		tc.mutation.SetUpdateTime(v)
	}
	if _, ok := tc.mutation.ID(); !ok {
		v := todo.DefaultID()
		tc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (tc *TodoCreate) check() error {
	if _, ok := tc.mutation.CreateTime(); !ok {
		return &ValidationError{Name: "create_time", err: errors.New(`models: missing required field "Todo.create_time"`)}
	}
	if _, ok := tc.mutation.UpdateTime(); !ok {
		return &ValidationError{Name: "update_time", err: errors.New(`models: missing required field "Todo.update_time"`)}
	}
	if _, ok := tc.mutation.Title(); !ok {
		return &ValidationError{Name: "title", err: errors.New(`models: missing required field "Todo.title"`)}
	}
	if v, ok := tc.mutation.Title(); ok {
		if err := todo.TitleValidator(v); err != nil {
			return &ValidationError{Name: "title", err: fmt.Errorf(`models: validator failed for field "Todo.title": %w`, err)}
		}
	}
	if _, ok := tc.mutation.Description(); !ok {
		return &ValidationError{Name: "description", err: errors.New(`models: missing required field "Todo.description"`)}
	}
	if v, ok := tc.mutation.Description(); ok {
		if err := todo.DescriptionValidator(v); err != nil {
			return &ValidationError{Name: "description", err: fmt.Errorf(`models: validator failed for field "Todo.description": %w`, err)}
		}
	}
	return nil
}

func (tc *TodoCreate) sqlSave(ctx context.Context) (*Todo, error) {
	_node, _spec := tc.createSpec()
	if err := sqlgraph.CreateNode(ctx, tc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{err.Error(), err}
		}
		return nil, err
	}
	if _spec.ID.Value != nil {
		if id, ok := _spec.ID.Value.(*uuid.UUID); ok {
			_node.ID = *id
		} else if err := _node.ID.Scan(_spec.ID.Value); err != nil {
			return nil, err
		}
	}
	return _node, nil
}

func (tc *TodoCreate) createSpec() (*Todo, *sqlgraph.CreateSpec) {
	var (
		_node = &Todo{config: tc.config}
		_spec = &sqlgraph.CreateSpec{
			Table: todo.Table,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeUUID,
				Column: todo.FieldID,
			},
		}
	)
	if id, ok := tc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := tc.mutation.CreateTime(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeTime,
			Value:  value,
			Column: todo.FieldCreateTime,
		})
		_node.CreateTime = value
	}
	if value, ok := tc.mutation.UpdateTime(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeTime,
			Value:  value,
			Column: todo.FieldUpdateTime,
		})
		_node.UpdateTime = value
	}
	if value, ok := tc.mutation.Title(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: todo.FieldTitle,
		})
		_node.Title = value
	}
	if value, ok := tc.mutation.Description(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: todo.FieldDescription,
		})
		_node.Description = value
	}
	if nodes := tc.mutation.OwnerIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   todo.OwnerTable,
			Columns: []string{todo.OwnerColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: &sqlgraph.FieldSpec{
					Type:   field.TypeUUID,
					Column: board.FieldID,
				},
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.board_todos = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// TodoCreateBulk is the builder for creating many Todo entities in bulk.
type TodoCreateBulk struct {
	config
	builders []*TodoCreate
}

// Save creates the Todo entities in the database.
func (tcb *TodoCreateBulk) Save(ctx context.Context) ([]*Todo, error) {
	specs := make([]*sqlgraph.CreateSpec, len(tcb.builders))
	nodes := make([]*Todo, len(tcb.builders))
	mutators := make([]Mutator, len(tcb.builders))
	for i := range tcb.builders {
		func(i int, root context.Context) {
			builder := tcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*TodoMutation)
				if !ok {
					return nil, fmt.Errorf("unexpected mutation type %T", m)
				}
				if err := builder.check(); err != nil {
					return nil, err
				}
				builder.mutation = mutation
				nodes[i], specs[i] = builder.createSpec()
				var err error
				if i < len(mutators)-1 {
					_, err = mutators[i+1].Mutate(root, tcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, tcb.driver, spec); err != nil {
						if sqlgraph.IsConstraintError(err) {
							err = &ConstraintError{err.Error(), err}
						}
					}
				}
				if err != nil {
					return nil, err
				}
				mutation.id = &nodes[i].ID
				mutation.done = true
				return nodes[i], nil
			})
			for i := len(builder.hooks) - 1; i >= 0; i-- {
				mut = builder.hooks[i](mut)
			}
			mutators[i] = mut
		}(i, ctx)
	}
	if len(mutators) > 0 {
		if _, err := mutators[0].Mutate(ctx, tcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (tcb *TodoCreateBulk) SaveX(ctx context.Context) []*Todo {
	v, err := tcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (tcb *TodoCreateBulk) Exec(ctx context.Context) error {
	_, err := tcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (tcb *TodoCreateBulk) ExecX(ctx context.Context) {
	if err := tcb.Exec(ctx); err != nil {
		panic(err)
	}
}
