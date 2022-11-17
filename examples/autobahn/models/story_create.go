// Code generated (@generated) by entc, DO NOT EDIT.

package models

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/adnaan/autobahn/models/board"
	"github.com/adnaan/autobahn/models/story"
	"github.com/google/uuid"
)

// StoryCreate is the builder for creating a Story entity.
type StoryCreate struct {
	config
	mutation *StoryMutation
	hooks    []Hook
}

// SetCreateTime sets the "create_time" field.
func (sc *StoryCreate) SetCreateTime(t time.Time) *StoryCreate {
	sc.mutation.SetCreateTime(t)
	return sc
}

// SetNillableCreateTime sets the "create_time" field if the given value is not nil.
func (sc *StoryCreate) SetNillableCreateTime(t *time.Time) *StoryCreate {
	if t != nil {
		sc.SetCreateTime(*t)
	}
	return sc
}

// SetUpdateTime sets the "update_time" field.
func (sc *StoryCreate) SetUpdateTime(t time.Time) *StoryCreate {
	sc.mutation.SetUpdateTime(t)
	return sc
}

// SetNillableUpdateTime sets the "update_time" field if the given value is not nil.
func (sc *StoryCreate) SetNillableUpdateTime(t *time.Time) *StoryCreate {
	if t != nil {
		sc.SetUpdateTime(*t)
	}
	return sc
}

// SetTitle sets the "title" field.
func (sc *StoryCreate) SetTitle(s string) *StoryCreate {
	sc.mutation.SetTitle(s)
	return sc
}

// SetDescription sets the "description" field.
func (sc *StoryCreate) SetDescription(s string) *StoryCreate {
	sc.mutation.SetDescription(s)
	return sc
}

// SetID sets the "id" field.
func (sc *StoryCreate) SetID(u uuid.UUID) *StoryCreate {
	sc.mutation.SetID(u)
	return sc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (sc *StoryCreate) SetNillableID(u *uuid.UUID) *StoryCreate {
	if u != nil {
		sc.SetID(*u)
	}
	return sc
}

// SetOwnerID sets the "owner" edge to the Board entity by ID.
func (sc *StoryCreate) SetOwnerID(id uuid.UUID) *StoryCreate {
	sc.mutation.SetOwnerID(id)
	return sc
}

// SetNillableOwnerID sets the "owner" edge to the Board entity by ID if the given value is not nil.
func (sc *StoryCreate) SetNillableOwnerID(id *uuid.UUID) *StoryCreate {
	if id != nil {
		sc = sc.SetOwnerID(*id)
	}
	return sc
}

// SetOwner sets the "owner" edge to the Board entity.
func (sc *StoryCreate) SetOwner(b *Board) *StoryCreate {
	return sc.SetOwnerID(b.ID)
}

// Mutation returns the StoryMutation object of the builder.
func (sc *StoryCreate) Mutation() *StoryMutation {
	return sc.mutation
}

// Save creates the Story in the database.
func (sc *StoryCreate) Save(ctx context.Context) (*Story, error) {
	var (
		err  error
		node *Story
	)
	sc.defaults()
	if len(sc.hooks) == 0 {
		if err = sc.check(); err != nil {
			return nil, err
		}
		node, err = sc.sqlSave(ctx)
	} else {
		var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
			mutation, ok := m.(*StoryMutation)
			if !ok {
				return nil, fmt.Errorf("unexpected mutation type %T", m)
			}
			if err = sc.check(); err != nil {
				return nil, err
			}
			sc.mutation = mutation
			if node, err = sc.sqlSave(ctx); err != nil {
				return nil, err
			}
			mutation.id = &node.ID
			mutation.done = true
			return node, err
		})
		for i := len(sc.hooks) - 1; i >= 0; i-- {
			if sc.hooks[i] == nil {
				return nil, fmt.Errorf("models: uninitialized hook (forgotten import models/runtime?)")
			}
			mut = sc.hooks[i](mut)
		}
		if _, err := mut.Mutate(ctx, sc.mutation); err != nil {
			return nil, err
		}
	}
	return node, err
}

// SaveX calls Save and panics if Save returns an error.
func (sc *StoryCreate) SaveX(ctx context.Context) *Story {
	v, err := sc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (sc *StoryCreate) Exec(ctx context.Context) error {
	_, err := sc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (sc *StoryCreate) ExecX(ctx context.Context) {
	if err := sc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (sc *StoryCreate) defaults() {
	if _, ok := sc.mutation.CreateTime(); !ok {
		v := story.DefaultCreateTime()
		sc.mutation.SetCreateTime(v)
	}
	if _, ok := sc.mutation.UpdateTime(); !ok {
		v := story.DefaultUpdateTime()
		sc.mutation.SetUpdateTime(v)
	}
	if _, ok := sc.mutation.ID(); !ok {
		v := story.DefaultID()
		sc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (sc *StoryCreate) check() error {
	if _, ok := sc.mutation.CreateTime(); !ok {
		return &ValidationError{Name: "create_time", err: errors.New(`models: missing required field "Story.create_time"`)}
	}
	if _, ok := sc.mutation.UpdateTime(); !ok {
		return &ValidationError{Name: "update_time", err: errors.New(`models: missing required field "Story.update_time"`)}
	}
	if _, ok := sc.mutation.Title(); !ok {
		return &ValidationError{Name: "title", err: errors.New(`models: missing required field "Story.title"`)}
	}
	if v, ok := sc.mutation.Title(); ok {
		if err := story.TitleValidator(v); err != nil {
			return &ValidationError{Name: "title", err: fmt.Errorf(`models: validator failed for field "Story.title": %w`, err)}
		}
	}
	if _, ok := sc.mutation.Description(); !ok {
		return &ValidationError{Name: "description", err: errors.New(`models: missing required field "Story.description"`)}
	}
	if v, ok := sc.mutation.Description(); ok {
		if err := story.DescriptionValidator(v); err != nil {
			return &ValidationError{Name: "description", err: fmt.Errorf(`models: validator failed for field "Story.description": %w`, err)}
		}
	}
	return nil
}

func (sc *StoryCreate) sqlSave(ctx context.Context) (*Story, error) {
	_node, _spec := sc.createSpec()
	if err := sqlgraph.CreateNode(ctx, sc.driver, _spec); err != nil {
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

func (sc *StoryCreate) createSpec() (*Story, *sqlgraph.CreateSpec) {
	var (
		_node = &Story{config: sc.config}
		_spec = &sqlgraph.CreateSpec{
			Table: story.Table,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeUUID,
				Column: story.FieldID,
			},
		}
	)
	if id, ok := sc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := sc.mutation.CreateTime(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeTime,
			Value:  value,
			Column: story.FieldCreateTime,
		})
		_node.CreateTime = value
	}
	if value, ok := sc.mutation.UpdateTime(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeTime,
			Value:  value,
			Column: story.FieldUpdateTime,
		})
		_node.UpdateTime = value
	}
	if value, ok := sc.mutation.Title(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: story.FieldTitle,
		})
		_node.Title = value
	}
	if value, ok := sc.mutation.Description(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: story.FieldDescription,
		})
		_node.Description = value
	}
	if nodes := sc.mutation.OwnerIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   story.OwnerTable,
			Columns: []string{story.OwnerColumn},
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
		_node.board_stories = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// StoryCreateBulk is the builder for creating many Story entities in bulk.
type StoryCreateBulk struct {
	config
	builders []*StoryCreate
}

// Save creates the Story entities in the database.
func (scb *StoryCreateBulk) Save(ctx context.Context) ([]*Story, error) {
	specs := make([]*sqlgraph.CreateSpec, len(scb.builders))
	nodes := make([]*Story, len(scb.builders))
	mutators := make([]Mutator, len(scb.builders))
	for i := range scb.builders {
		func(i int, root context.Context) {
			builder := scb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*StoryMutation)
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
					_, err = mutators[i+1].Mutate(root, scb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, scb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, scb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (scb *StoryCreateBulk) SaveX(ctx context.Context) []*Story {
	v, err := scb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (scb *StoryCreateBulk) Exec(ctx context.Context) error {
	_, err := scb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (scb *StoryCreateBulk) ExecX(ctx context.Context) {
	if err := scb.Exec(ctx); err != nil {
		panic(err)
	}
}
