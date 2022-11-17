// Code generated (@generated) by entc, DO NOT EDIT.

package models

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/adnaan/autobahn/models/predicate"
	"github.com/adnaan/autobahn/models/view"
)

// ViewUpdate is the builder for updating View entities.
type ViewUpdate struct {
	config
	hooks    []Hook
	mutation *ViewMutation
}

// Where appends a list predicates to the ViewUpdate builder.
func (vu *ViewUpdate) Where(ps ...predicate.View) *ViewUpdate {
	vu.mutation.Where(ps...)
	return vu
}

// SetUpdateTime sets the "update_time" field.
func (vu *ViewUpdate) SetUpdateTime(t time.Time) *ViewUpdate {
	vu.mutation.SetUpdateTime(t)
	return vu
}

// SetTitle sets the "title" field.
func (vu *ViewUpdate) SetTitle(s string) *ViewUpdate {
	vu.mutation.SetTitle(s)
	return vu
}

// Mutation returns the ViewMutation object of the builder.
func (vu *ViewUpdate) Mutation() *ViewMutation {
	return vu.mutation
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (vu *ViewUpdate) Save(ctx context.Context) (int, error) {
	var (
		err      error
		affected int
	)
	vu.defaults()
	if len(vu.hooks) == 0 {
		affected, err = vu.sqlSave(ctx)
	} else {
		var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
			mutation, ok := m.(*ViewMutation)
			if !ok {
				return nil, fmt.Errorf("unexpected mutation type %T", m)
			}
			vu.mutation = mutation
			affected, err = vu.sqlSave(ctx)
			mutation.done = true
			return affected, err
		})
		for i := len(vu.hooks) - 1; i >= 0; i-- {
			if vu.hooks[i] == nil {
				return 0, fmt.Errorf("models: uninitialized hook (forgotten import models/runtime?)")
			}
			mut = vu.hooks[i](mut)
		}
		if _, err := mut.Mutate(ctx, vu.mutation); err != nil {
			return 0, err
		}
	}
	return affected, err
}

// SaveX is like Save, but panics if an error occurs.
func (vu *ViewUpdate) SaveX(ctx context.Context) int {
	affected, err := vu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (vu *ViewUpdate) Exec(ctx context.Context) error {
	_, err := vu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (vu *ViewUpdate) ExecX(ctx context.Context) {
	if err := vu.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (vu *ViewUpdate) defaults() {
	if _, ok := vu.mutation.UpdateTime(); !ok {
		v := view.UpdateDefaultUpdateTime()
		vu.mutation.SetUpdateTime(v)
	}
}

func (vu *ViewUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := &sqlgraph.UpdateSpec{
		Node: &sqlgraph.NodeSpec{
			Table:   view.Table,
			Columns: view.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeUUID,
				Column: view.FieldID,
			},
		},
	}
	if ps := vu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := vu.mutation.UpdateTime(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeTime,
			Value:  value,
			Column: view.FieldUpdateTime,
		})
	}
	if value, ok := vu.mutation.Title(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: view.FieldTitle,
		})
	}
	if n, err = sqlgraph.UpdateNodes(ctx, vu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{view.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{err.Error(), err}
		}
		return 0, err
	}
	return n, nil
}

// ViewUpdateOne is the builder for updating a single View entity.
type ViewUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *ViewMutation
}

// SetUpdateTime sets the "update_time" field.
func (vuo *ViewUpdateOne) SetUpdateTime(t time.Time) *ViewUpdateOne {
	vuo.mutation.SetUpdateTime(t)
	return vuo
}

// SetTitle sets the "title" field.
func (vuo *ViewUpdateOne) SetTitle(s string) *ViewUpdateOne {
	vuo.mutation.SetTitle(s)
	return vuo
}

// Mutation returns the ViewMutation object of the builder.
func (vuo *ViewUpdateOne) Mutation() *ViewMutation {
	return vuo.mutation
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (vuo *ViewUpdateOne) Select(field string, fields ...string) *ViewUpdateOne {
	vuo.fields = append([]string{field}, fields...)
	return vuo
}

// Save executes the query and returns the updated View entity.
func (vuo *ViewUpdateOne) Save(ctx context.Context) (*View, error) {
	var (
		err  error
		node *View
	)
	vuo.defaults()
	if len(vuo.hooks) == 0 {
		node, err = vuo.sqlSave(ctx)
	} else {
		var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
			mutation, ok := m.(*ViewMutation)
			if !ok {
				return nil, fmt.Errorf("unexpected mutation type %T", m)
			}
			vuo.mutation = mutation
			node, err = vuo.sqlSave(ctx)
			mutation.done = true
			return node, err
		})
		for i := len(vuo.hooks) - 1; i >= 0; i-- {
			if vuo.hooks[i] == nil {
				return nil, fmt.Errorf("models: uninitialized hook (forgotten import models/runtime?)")
			}
			mut = vuo.hooks[i](mut)
		}
		if _, err := mut.Mutate(ctx, vuo.mutation); err != nil {
			return nil, err
		}
	}
	return node, err
}

// SaveX is like Save, but panics if an error occurs.
func (vuo *ViewUpdateOne) SaveX(ctx context.Context) *View {
	node, err := vuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (vuo *ViewUpdateOne) Exec(ctx context.Context) error {
	_, err := vuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (vuo *ViewUpdateOne) ExecX(ctx context.Context) {
	if err := vuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (vuo *ViewUpdateOne) defaults() {
	if _, ok := vuo.mutation.UpdateTime(); !ok {
		v := view.UpdateDefaultUpdateTime()
		vuo.mutation.SetUpdateTime(v)
	}
}

func (vuo *ViewUpdateOne) sqlSave(ctx context.Context) (_node *View, err error) {
	_spec := &sqlgraph.UpdateSpec{
		Node: &sqlgraph.NodeSpec{
			Table:   view.Table,
			Columns: view.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeUUID,
				Column: view.FieldID,
			},
		},
	}
	id, ok := vuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`models: missing "View.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := vuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, view.FieldID)
		for _, f := range fields {
			if !view.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("models: invalid field %q for query", f)}
			}
			if f != view.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := vuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := vuo.mutation.UpdateTime(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeTime,
			Value:  value,
			Column: view.FieldUpdateTime,
		})
	}
	if value, ok := vuo.mutation.Title(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: view.FieldTitle,
		})
	}
	_node = &View{config: vuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, vuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{view.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{err.Error(), err}
		}
		return nil, err
	}
	return _node, nil
}
