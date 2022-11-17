// Code generated (@generated) by entc, DO NOT EDIT.

package models

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/adnaan/autobahn/models/label"
	"github.com/google/uuid"
)

// LabelCreate is the builder for creating a Label entity.
type LabelCreate struct {
	config
	mutation *LabelMutation
	hooks    []Hook
}

// SetCreateTime sets the "create_time" field.
func (lc *LabelCreate) SetCreateTime(t time.Time) *LabelCreate {
	lc.mutation.SetCreateTime(t)
	return lc
}

// SetNillableCreateTime sets the "create_time" field if the given value is not nil.
func (lc *LabelCreate) SetNillableCreateTime(t *time.Time) *LabelCreate {
	if t != nil {
		lc.SetCreateTime(*t)
	}
	return lc
}

// SetUpdateTime sets the "update_time" field.
func (lc *LabelCreate) SetUpdateTime(t time.Time) *LabelCreate {
	lc.mutation.SetUpdateTime(t)
	return lc
}

// SetNillableUpdateTime sets the "update_time" field if the given value is not nil.
func (lc *LabelCreate) SetNillableUpdateTime(t *time.Time) *LabelCreate {
	if t != nil {
		lc.SetUpdateTime(*t)
	}
	return lc
}

// SetID sets the "id" field.
func (lc *LabelCreate) SetID(u uuid.UUID) *LabelCreate {
	lc.mutation.SetID(u)
	return lc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (lc *LabelCreate) SetNillableID(u *uuid.UUID) *LabelCreate {
	if u != nil {
		lc.SetID(*u)
	}
	return lc
}

// Mutation returns the LabelMutation object of the builder.
func (lc *LabelCreate) Mutation() *LabelMutation {
	return lc.mutation
}

// Save creates the Label in the database.
func (lc *LabelCreate) Save(ctx context.Context) (*Label, error) {
	var (
		err  error
		node *Label
	)
	lc.defaults()
	if len(lc.hooks) == 0 {
		if err = lc.check(); err != nil {
			return nil, err
		}
		node, err = lc.sqlSave(ctx)
	} else {
		var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
			mutation, ok := m.(*LabelMutation)
			if !ok {
				return nil, fmt.Errorf("unexpected mutation type %T", m)
			}
			if err = lc.check(); err != nil {
				return nil, err
			}
			lc.mutation = mutation
			if node, err = lc.sqlSave(ctx); err != nil {
				return nil, err
			}
			mutation.id = &node.ID
			mutation.done = true
			return node, err
		})
		for i := len(lc.hooks) - 1; i >= 0; i-- {
			if lc.hooks[i] == nil {
				return nil, fmt.Errorf("models: uninitialized hook (forgotten import models/runtime?)")
			}
			mut = lc.hooks[i](mut)
		}
		if _, err := mut.Mutate(ctx, lc.mutation); err != nil {
			return nil, err
		}
	}
	return node, err
}

// SaveX calls Save and panics if Save returns an error.
func (lc *LabelCreate) SaveX(ctx context.Context) *Label {
	v, err := lc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (lc *LabelCreate) Exec(ctx context.Context) error {
	_, err := lc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (lc *LabelCreate) ExecX(ctx context.Context) {
	if err := lc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (lc *LabelCreate) defaults() {
	if _, ok := lc.mutation.CreateTime(); !ok {
		v := label.DefaultCreateTime()
		lc.mutation.SetCreateTime(v)
	}
	if _, ok := lc.mutation.UpdateTime(); !ok {
		v := label.DefaultUpdateTime()
		lc.mutation.SetUpdateTime(v)
	}
	if _, ok := lc.mutation.ID(); !ok {
		v := label.DefaultID()
		lc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (lc *LabelCreate) check() error {
	if _, ok := lc.mutation.CreateTime(); !ok {
		return &ValidationError{Name: "create_time", err: errors.New(`models: missing required field "Label.create_time"`)}
	}
	if _, ok := lc.mutation.UpdateTime(); !ok {
		return &ValidationError{Name: "update_time", err: errors.New(`models: missing required field "Label.update_time"`)}
	}
	return nil
}

func (lc *LabelCreate) sqlSave(ctx context.Context) (*Label, error) {
	_node, _spec := lc.createSpec()
	if err := sqlgraph.CreateNode(ctx, lc.driver, _spec); err != nil {
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

func (lc *LabelCreate) createSpec() (*Label, *sqlgraph.CreateSpec) {
	var (
		_node = &Label{config: lc.config}
		_spec = &sqlgraph.CreateSpec{
			Table: label.Table,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeUUID,
				Column: label.FieldID,
			},
		}
	)
	if id, ok := lc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := lc.mutation.CreateTime(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeTime,
			Value:  value,
			Column: label.FieldCreateTime,
		})
		_node.CreateTime = value
	}
	if value, ok := lc.mutation.UpdateTime(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeTime,
			Value:  value,
			Column: label.FieldUpdateTime,
		})
		_node.UpdateTime = value
	}
	return _node, _spec
}

// LabelCreateBulk is the builder for creating many Label entities in bulk.
type LabelCreateBulk struct {
	config
	builders []*LabelCreate
}

// Save creates the Label entities in the database.
func (lcb *LabelCreateBulk) Save(ctx context.Context) ([]*Label, error) {
	specs := make([]*sqlgraph.CreateSpec, len(lcb.builders))
	nodes := make([]*Label, len(lcb.builders))
	mutators := make([]Mutator, len(lcb.builders))
	for i := range lcb.builders {
		func(i int, root context.Context) {
			builder := lcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*LabelMutation)
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
					_, err = mutators[i+1].Mutate(root, lcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, lcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, lcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (lcb *LabelCreateBulk) SaveX(ctx context.Context) []*Label {
	v, err := lcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (lcb *LabelCreateBulk) Exec(ctx context.Context) error {
	_, err := lcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (lcb *LabelCreateBulk) ExecX(ctx context.Context) {
	if err := lcb.Exec(ctx); err != nil {
		panic(err)
	}
}
