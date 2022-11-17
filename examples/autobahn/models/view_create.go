// Code generated (@generated) by entc, DO NOT EDIT.

package models

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/adnaan/autobahn/models/view"
	"github.com/google/uuid"
)

// ViewCreate is the builder for creating a View entity.
type ViewCreate struct {
	config
	mutation *ViewMutation
	hooks    []Hook
}

// SetCreateTime sets the "create_time" field.
func (vc *ViewCreate) SetCreateTime(t time.Time) *ViewCreate {
	vc.mutation.SetCreateTime(t)
	return vc
}

// SetNillableCreateTime sets the "create_time" field if the given value is not nil.
func (vc *ViewCreate) SetNillableCreateTime(t *time.Time) *ViewCreate {
	if t != nil {
		vc.SetCreateTime(*t)
	}
	return vc
}

// SetUpdateTime sets the "update_time" field.
func (vc *ViewCreate) SetUpdateTime(t time.Time) *ViewCreate {
	vc.mutation.SetUpdateTime(t)
	return vc
}

// SetNillableUpdateTime sets the "update_time" field if the given value is not nil.
func (vc *ViewCreate) SetNillableUpdateTime(t *time.Time) *ViewCreate {
	if t != nil {
		vc.SetUpdateTime(*t)
	}
	return vc
}

// SetTitle sets the "title" field.
func (vc *ViewCreate) SetTitle(s string) *ViewCreate {
	vc.mutation.SetTitle(s)
	return vc
}

// SetID sets the "id" field.
func (vc *ViewCreate) SetID(u uuid.UUID) *ViewCreate {
	vc.mutation.SetID(u)
	return vc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (vc *ViewCreate) SetNillableID(u *uuid.UUID) *ViewCreate {
	if u != nil {
		vc.SetID(*u)
	}
	return vc
}

// Mutation returns the ViewMutation object of the builder.
func (vc *ViewCreate) Mutation() *ViewMutation {
	return vc.mutation
}

// Save creates the View in the database.
func (vc *ViewCreate) Save(ctx context.Context) (*View, error) {
	var (
		err  error
		node *View
	)
	vc.defaults()
	if len(vc.hooks) == 0 {
		if err = vc.check(); err != nil {
			return nil, err
		}
		node, err = vc.sqlSave(ctx)
	} else {
		var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
			mutation, ok := m.(*ViewMutation)
			if !ok {
				return nil, fmt.Errorf("unexpected mutation type %T", m)
			}
			if err = vc.check(); err != nil {
				return nil, err
			}
			vc.mutation = mutation
			if node, err = vc.sqlSave(ctx); err != nil {
				return nil, err
			}
			mutation.id = &node.ID
			mutation.done = true
			return node, err
		})
		for i := len(vc.hooks) - 1; i >= 0; i-- {
			if vc.hooks[i] == nil {
				return nil, fmt.Errorf("models: uninitialized hook (forgotten import models/runtime?)")
			}
			mut = vc.hooks[i](mut)
		}
		if _, err := mut.Mutate(ctx, vc.mutation); err != nil {
			return nil, err
		}
	}
	return node, err
}

// SaveX calls Save and panics if Save returns an error.
func (vc *ViewCreate) SaveX(ctx context.Context) *View {
	v, err := vc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (vc *ViewCreate) Exec(ctx context.Context) error {
	_, err := vc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (vc *ViewCreate) ExecX(ctx context.Context) {
	if err := vc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (vc *ViewCreate) defaults() {
	if _, ok := vc.mutation.CreateTime(); !ok {
		v := view.DefaultCreateTime()
		vc.mutation.SetCreateTime(v)
	}
	if _, ok := vc.mutation.UpdateTime(); !ok {
		v := view.DefaultUpdateTime()
		vc.mutation.SetUpdateTime(v)
	}
	if _, ok := vc.mutation.ID(); !ok {
		v := view.DefaultID()
		vc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (vc *ViewCreate) check() error {
	if _, ok := vc.mutation.CreateTime(); !ok {
		return &ValidationError{Name: "create_time", err: errors.New(`models: missing required field "View.create_time"`)}
	}
	if _, ok := vc.mutation.UpdateTime(); !ok {
		return &ValidationError{Name: "update_time", err: errors.New(`models: missing required field "View.update_time"`)}
	}
	if _, ok := vc.mutation.Title(); !ok {
		return &ValidationError{Name: "title", err: errors.New(`models: missing required field "View.title"`)}
	}
	return nil
}

func (vc *ViewCreate) sqlSave(ctx context.Context) (*View, error) {
	_node, _spec := vc.createSpec()
	if err := sqlgraph.CreateNode(ctx, vc.driver, _spec); err != nil {
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

func (vc *ViewCreate) createSpec() (*View, *sqlgraph.CreateSpec) {
	var (
		_node = &View{config: vc.config}
		_spec = &sqlgraph.CreateSpec{
			Table: view.Table,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeUUID,
				Column: view.FieldID,
			},
		}
	)
	if id, ok := vc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := vc.mutation.CreateTime(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeTime,
			Value:  value,
			Column: view.FieldCreateTime,
		})
		_node.CreateTime = value
	}
	if value, ok := vc.mutation.UpdateTime(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeTime,
			Value:  value,
			Column: view.FieldUpdateTime,
		})
		_node.UpdateTime = value
	}
	if value, ok := vc.mutation.Title(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: view.FieldTitle,
		})
		_node.Title = value
	}
	return _node, _spec
}

// ViewCreateBulk is the builder for creating many View entities in bulk.
type ViewCreateBulk struct {
	config
	builders []*ViewCreate
}

// Save creates the View entities in the database.
func (vcb *ViewCreateBulk) Save(ctx context.Context) ([]*View, error) {
	specs := make([]*sqlgraph.CreateSpec, len(vcb.builders))
	nodes := make([]*View, len(vcb.builders))
	mutators := make([]Mutator, len(vcb.builders))
	for i := range vcb.builders {
		func(i int, root context.Context) {
			builder := vcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*ViewMutation)
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
					_, err = mutators[i+1].Mutate(root, vcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, vcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, vcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (vcb *ViewCreateBulk) SaveX(ctx context.Context) []*View {
	v, err := vcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (vcb *ViewCreateBulk) Exec(ctx context.Context) error {
	_, err := vcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (vcb *ViewCreateBulk) ExecX(ctx context.Context) {
	if err := vcb.Exec(ctx); err != nil {
		panic(err)
	}
}
