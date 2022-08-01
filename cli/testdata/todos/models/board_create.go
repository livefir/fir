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

// BoardCreate is the builder for creating a Board entity.
type BoardCreate struct {
	config
	mutation *BoardMutation
	hooks    []Hook
}

// SetCreateTime sets the "create_time" field.
func (bc *BoardCreate) SetCreateTime(t time.Time) *BoardCreate {
	bc.mutation.SetCreateTime(t)
	return bc
}

// SetNillableCreateTime sets the "create_time" field if the given value is not nil.
func (bc *BoardCreate) SetNillableCreateTime(t *time.Time) *BoardCreate {
	if t != nil {
		bc.SetCreateTime(*t)
	}
	return bc
}

// SetUpdateTime sets the "update_time" field.
func (bc *BoardCreate) SetUpdateTime(t time.Time) *BoardCreate {
	bc.mutation.SetUpdateTime(t)
	return bc
}

// SetNillableUpdateTime sets the "update_time" field if the given value is not nil.
func (bc *BoardCreate) SetNillableUpdateTime(t *time.Time) *BoardCreate {
	if t != nil {
		bc.SetUpdateTime(*t)
	}
	return bc
}

// SetTitle sets the "title" field.
func (bc *BoardCreate) SetTitle(s string) *BoardCreate {
	bc.mutation.SetTitle(s)
	return bc
}

// SetDescription sets the "description" field.
func (bc *BoardCreate) SetDescription(s string) *BoardCreate {
	bc.mutation.SetDescription(s)
	return bc
}

// SetID sets the "id" field.
func (bc *BoardCreate) SetID(u uuid.UUID) *BoardCreate {
	bc.mutation.SetID(u)
	return bc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (bc *BoardCreate) SetNillableID(u *uuid.UUID) *BoardCreate {
	if u != nil {
		bc.SetID(*u)
	}
	return bc
}

// AddTodoIDs adds the "todos" edge to the Todo entity by IDs.
func (bc *BoardCreate) AddTodoIDs(ids ...uuid.UUID) *BoardCreate {
	bc.mutation.AddTodoIDs(ids...)
	return bc
}

// AddTodos adds the "todos" edges to the Todo entity.
func (bc *BoardCreate) AddTodos(t ...*Todo) *BoardCreate {
	ids := make([]uuid.UUID, len(t))
	for i := range t {
		ids[i] = t[i].ID
	}
	return bc.AddTodoIDs(ids...)
}

// Mutation returns the BoardMutation object of the builder.
func (bc *BoardCreate) Mutation() *BoardMutation {
	return bc.mutation
}

// Save creates the Board in the database.
func (bc *BoardCreate) Save(ctx context.Context) (*Board, error) {
	var (
		err  error
		node *Board
	)
	bc.defaults()
	if len(bc.hooks) == 0 {
		if err = bc.check(); err != nil {
			return nil, err
		}
		node, err = bc.sqlSave(ctx)
	} else {
		var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
			mutation, ok := m.(*BoardMutation)
			if !ok {
				return nil, fmt.Errorf("unexpected mutation type %T", m)
			}
			if err = bc.check(); err != nil {
				return nil, err
			}
			bc.mutation = mutation
			if node, err = bc.sqlSave(ctx); err != nil {
				return nil, err
			}
			mutation.id = &node.ID
			mutation.done = true
			return node, err
		})
		for i := len(bc.hooks) - 1; i >= 0; i-- {
			if bc.hooks[i] == nil {
				return nil, fmt.Errorf("models: uninitialized hook (forgotten import models/runtime?)")
			}
			mut = bc.hooks[i](mut)
		}
		if _, err := mut.Mutate(ctx, bc.mutation); err != nil {
			return nil, err
		}
	}
	return node, err
}

// SaveX calls Save and panics if Save returns an error.
func (bc *BoardCreate) SaveX(ctx context.Context) *Board {
	v, err := bc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (bc *BoardCreate) Exec(ctx context.Context) error {
	_, err := bc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (bc *BoardCreate) ExecX(ctx context.Context) {
	if err := bc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (bc *BoardCreate) defaults() {
	if _, ok := bc.mutation.CreateTime(); !ok {
		v := board.DefaultCreateTime()
		bc.mutation.SetCreateTime(v)
	}
	if _, ok := bc.mutation.UpdateTime(); !ok {
		v := board.DefaultUpdateTime()
		bc.mutation.SetUpdateTime(v)
	}
	if _, ok := bc.mutation.ID(); !ok {
		v := board.DefaultID()
		bc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (bc *BoardCreate) check() error {
	if _, ok := bc.mutation.CreateTime(); !ok {
		return &ValidationError{Name: "create_time", err: errors.New(`models: missing required field "Board.create_time"`)}
	}
	if _, ok := bc.mutation.UpdateTime(); !ok {
		return &ValidationError{Name: "update_time", err: errors.New(`models: missing required field "Board.update_time"`)}
	}
	if _, ok := bc.mutation.Title(); !ok {
		return &ValidationError{Name: "title", err: errors.New(`models: missing required field "Board.title"`)}
	}
	if v, ok := bc.mutation.Title(); ok {
		if err := board.TitleValidator(v); err != nil {
			return &ValidationError{Name: "title", err: fmt.Errorf(`models: validator failed for field "Board.title": %w`, err)}
		}
	}
	if _, ok := bc.mutation.Description(); !ok {
		return &ValidationError{Name: "description", err: errors.New(`models: missing required field "Board.description"`)}
	}
	if v, ok := bc.mutation.Description(); ok {
		if err := board.DescriptionValidator(v); err != nil {
			return &ValidationError{Name: "description", err: fmt.Errorf(`models: validator failed for field "Board.description": %w`, err)}
		}
	}
	return nil
}

func (bc *BoardCreate) sqlSave(ctx context.Context) (*Board, error) {
	_node, _spec := bc.createSpec()
	if err := sqlgraph.CreateNode(ctx, bc.driver, _spec); err != nil {
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

func (bc *BoardCreate) createSpec() (*Board, *sqlgraph.CreateSpec) {
	var (
		_node = &Board{config: bc.config}
		_spec = &sqlgraph.CreateSpec{
			Table: board.Table,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeUUID,
				Column: board.FieldID,
			},
		}
	)
	if id, ok := bc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := bc.mutation.CreateTime(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeTime,
			Value:  value,
			Column: board.FieldCreateTime,
		})
		_node.CreateTime = value
	}
	if value, ok := bc.mutation.UpdateTime(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeTime,
			Value:  value,
			Column: board.FieldUpdateTime,
		})
		_node.UpdateTime = value
	}
	if value, ok := bc.mutation.Title(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: board.FieldTitle,
		})
		_node.Title = value
	}
	if value, ok := bc.mutation.Description(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: board.FieldDescription,
		})
		_node.Description = value
	}
	if nodes := bc.mutation.TodosIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   board.TodosTable,
			Columns: []string{board.TodosColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: &sqlgraph.FieldSpec{
					Type:   field.TypeUUID,
					Column: todo.FieldID,
				},
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// BoardCreateBulk is the builder for creating many Board entities in bulk.
type BoardCreateBulk struct {
	config
	builders []*BoardCreate
}

// Save creates the Board entities in the database.
func (bcb *BoardCreateBulk) Save(ctx context.Context) ([]*Board, error) {
	specs := make([]*sqlgraph.CreateSpec, len(bcb.builders))
	nodes := make([]*Board, len(bcb.builders))
	mutators := make([]Mutator, len(bcb.builders))
	for i := range bcb.builders {
		func(i int, root context.Context) {
			builder := bcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*BoardMutation)
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
					_, err = mutators[i+1].Mutate(root, bcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, bcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, bcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (bcb *BoardCreateBulk) SaveX(ctx context.Context) []*Board {
	v, err := bcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (bcb *BoardCreateBulk) Exec(ctx context.Context) error {
	_, err := bcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (bcb *BoardCreateBulk) ExecX(ctx context.Context) {
	if err := bcb.Exec(ctx); err != nil {
		panic(err)
	}
}
