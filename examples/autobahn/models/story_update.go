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
	"github.com/adnaan/autobahn/models/board"
	"github.com/adnaan/autobahn/models/predicate"
	"github.com/adnaan/autobahn/models/story"
	"github.com/google/uuid"
)

// StoryUpdate is the builder for updating Story entities.
type StoryUpdate struct {
	config
	hooks    []Hook
	mutation *StoryMutation
}

// Where appends a list predicates to the StoryUpdate builder.
func (su *StoryUpdate) Where(ps ...predicate.Story) *StoryUpdate {
	su.mutation.Where(ps...)
	return su
}

// SetUpdateTime sets the "update_time" field.
func (su *StoryUpdate) SetUpdateTime(t time.Time) *StoryUpdate {
	su.mutation.SetUpdateTime(t)
	return su
}

// SetTitle sets the "title" field.
func (su *StoryUpdate) SetTitle(s string) *StoryUpdate {
	su.mutation.SetTitle(s)
	return su
}

// SetDescription sets the "description" field.
func (su *StoryUpdate) SetDescription(s string) *StoryUpdate {
	su.mutation.SetDescription(s)
	return su
}

// SetOwnerID sets the "owner" edge to the Board entity by ID.
func (su *StoryUpdate) SetOwnerID(id uuid.UUID) *StoryUpdate {
	su.mutation.SetOwnerID(id)
	return su
}

// SetNillableOwnerID sets the "owner" edge to the Board entity by ID if the given value is not nil.
func (su *StoryUpdate) SetNillableOwnerID(id *uuid.UUID) *StoryUpdate {
	if id != nil {
		su = su.SetOwnerID(*id)
	}
	return su
}

// SetOwner sets the "owner" edge to the Board entity.
func (su *StoryUpdate) SetOwner(b *Board) *StoryUpdate {
	return su.SetOwnerID(b.ID)
}

// Mutation returns the StoryMutation object of the builder.
func (su *StoryUpdate) Mutation() *StoryMutation {
	return su.mutation
}

// ClearOwner clears the "owner" edge to the Board entity.
func (su *StoryUpdate) ClearOwner() *StoryUpdate {
	su.mutation.ClearOwner()
	return su
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (su *StoryUpdate) Save(ctx context.Context) (int, error) {
	var (
		err      error
		affected int
	)
	su.defaults()
	if len(su.hooks) == 0 {
		if err = su.check(); err != nil {
			return 0, err
		}
		affected, err = su.sqlSave(ctx)
	} else {
		var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
			mutation, ok := m.(*StoryMutation)
			if !ok {
				return nil, fmt.Errorf("unexpected mutation type %T", m)
			}
			if err = su.check(); err != nil {
				return 0, err
			}
			su.mutation = mutation
			affected, err = su.sqlSave(ctx)
			mutation.done = true
			return affected, err
		})
		for i := len(su.hooks) - 1; i >= 0; i-- {
			if su.hooks[i] == nil {
				return 0, fmt.Errorf("models: uninitialized hook (forgotten import models/runtime?)")
			}
			mut = su.hooks[i](mut)
		}
		if _, err := mut.Mutate(ctx, su.mutation); err != nil {
			return 0, err
		}
	}
	return affected, err
}

// SaveX is like Save, but panics if an error occurs.
func (su *StoryUpdate) SaveX(ctx context.Context) int {
	affected, err := su.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (su *StoryUpdate) Exec(ctx context.Context) error {
	_, err := su.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (su *StoryUpdate) ExecX(ctx context.Context) {
	if err := su.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (su *StoryUpdate) defaults() {
	if _, ok := su.mutation.UpdateTime(); !ok {
		v := story.UpdateDefaultUpdateTime()
		su.mutation.SetUpdateTime(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (su *StoryUpdate) check() error {
	if v, ok := su.mutation.Title(); ok {
		if err := story.TitleValidator(v); err != nil {
			return &ValidationError{Name: "title", err: fmt.Errorf(`models: validator failed for field "Story.title": %w`, err)}
		}
	}
	if v, ok := su.mutation.Description(); ok {
		if err := story.DescriptionValidator(v); err != nil {
			return &ValidationError{Name: "description", err: fmt.Errorf(`models: validator failed for field "Story.description": %w`, err)}
		}
	}
	return nil
}

func (su *StoryUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := &sqlgraph.UpdateSpec{
		Node: &sqlgraph.NodeSpec{
			Table:   story.Table,
			Columns: story.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeUUID,
				Column: story.FieldID,
			},
		},
	}
	if ps := su.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := su.mutation.UpdateTime(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeTime,
			Value:  value,
			Column: story.FieldUpdateTime,
		})
	}
	if value, ok := su.mutation.Title(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: story.FieldTitle,
		})
	}
	if value, ok := su.mutation.Description(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: story.FieldDescription,
		})
	}
	if su.mutation.OwnerCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := su.mutation.OwnerIDs(); len(nodes) > 0 {
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, su.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{story.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{err.Error(), err}
		}
		return 0, err
	}
	return n, nil
}

// StoryUpdateOne is the builder for updating a single Story entity.
type StoryUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *StoryMutation
}

// SetUpdateTime sets the "update_time" field.
func (suo *StoryUpdateOne) SetUpdateTime(t time.Time) *StoryUpdateOne {
	suo.mutation.SetUpdateTime(t)
	return suo
}

// SetTitle sets the "title" field.
func (suo *StoryUpdateOne) SetTitle(s string) *StoryUpdateOne {
	suo.mutation.SetTitle(s)
	return suo
}

// SetDescription sets the "description" field.
func (suo *StoryUpdateOne) SetDescription(s string) *StoryUpdateOne {
	suo.mutation.SetDescription(s)
	return suo
}

// SetOwnerID sets the "owner" edge to the Board entity by ID.
func (suo *StoryUpdateOne) SetOwnerID(id uuid.UUID) *StoryUpdateOne {
	suo.mutation.SetOwnerID(id)
	return suo
}

// SetNillableOwnerID sets the "owner" edge to the Board entity by ID if the given value is not nil.
func (suo *StoryUpdateOne) SetNillableOwnerID(id *uuid.UUID) *StoryUpdateOne {
	if id != nil {
		suo = suo.SetOwnerID(*id)
	}
	return suo
}

// SetOwner sets the "owner" edge to the Board entity.
func (suo *StoryUpdateOne) SetOwner(b *Board) *StoryUpdateOne {
	return suo.SetOwnerID(b.ID)
}

// Mutation returns the StoryMutation object of the builder.
func (suo *StoryUpdateOne) Mutation() *StoryMutation {
	return suo.mutation
}

// ClearOwner clears the "owner" edge to the Board entity.
func (suo *StoryUpdateOne) ClearOwner() *StoryUpdateOne {
	suo.mutation.ClearOwner()
	return suo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (suo *StoryUpdateOne) Select(field string, fields ...string) *StoryUpdateOne {
	suo.fields = append([]string{field}, fields...)
	return suo
}

// Save executes the query and returns the updated Story entity.
func (suo *StoryUpdateOne) Save(ctx context.Context) (*Story, error) {
	var (
		err  error
		node *Story
	)
	suo.defaults()
	if len(suo.hooks) == 0 {
		if err = suo.check(); err != nil {
			return nil, err
		}
		node, err = suo.sqlSave(ctx)
	} else {
		var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
			mutation, ok := m.(*StoryMutation)
			if !ok {
				return nil, fmt.Errorf("unexpected mutation type %T", m)
			}
			if err = suo.check(); err != nil {
				return nil, err
			}
			suo.mutation = mutation
			node, err = suo.sqlSave(ctx)
			mutation.done = true
			return node, err
		})
		for i := len(suo.hooks) - 1; i >= 0; i-- {
			if suo.hooks[i] == nil {
				return nil, fmt.Errorf("models: uninitialized hook (forgotten import models/runtime?)")
			}
			mut = suo.hooks[i](mut)
		}
		if _, err := mut.Mutate(ctx, suo.mutation); err != nil {
			return nil, err
		}
	}
	return node, err
}

// SaveX is like Save, but panics if an error occurs.
func (suo *StoryUpdateOne) SaveX(ctx context.Context) *Story {
	node, err := suo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (suo *StoryUpdateOne) Exec(ctx context.Context) error {
	_, err := suo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (suo *StoryUpdateOne) ExecX(ctx context.Context) {
	if err := suo.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (suo *StoryUpdateOne) defaults() {
	if _, ok := suo.mutation.UpdateTime(); !ok {
		v := story.UpdateDefaultUpdateTime()
		suo.mutation.SetUpdateTime(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (suo *StoryUpdateOne) check() error {
	if v, ok := suo.mutation.Title(); ok {
		if err := story.TitleValidator(v); err != nil {
			return &ValidationError{Name: "title", err: fmt.Errorf(`models: validator failed for field "Story.title": %w`, err)}
		}
	}
	if v, ok := suo.mutation.Description(); ok {
		if err := story.DescriptionValidator(v); err != nil {
			return &ValidationError{Name: "description", err: fmt.Errorf(`models: validator failed for field "Story.description": %w`, err)}
		}
	}
	return nil
}

func (suo *StoryUpdateOne) sqlSave(ctx context.Context) (_node *Story, err error) {
	_spec := &sqlgraph.UpdateSpec{
		Node: &sqlgraph.NodeSpec{
			Table:   story.Table,
			Columns: story.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeUUID,
				Column: story.FieldID,
			},
		},
	}
	id, ok := suo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`models: missing "Story.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := suo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, story.FieldID)
		for _, f := range fields {
			if !story.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("models: invalid field %q for query", f)}
			}
			if f != story.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := suo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := suo.mutation.UpdateTime(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeTime,
			Value:  value,
			Column: story.FieldUpdateTime,
		})
	}
	if value, ok := suo.mutation.Title(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: story.FieldTitle,
		})
	}
	if value, ok := suo.mutation.Description(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: story.FieldDescription,
		})
	}
	if suo.mutation.OwnerCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := suo.mutation.OwnerIDs(); len(nodes) > 0 {
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_node = &Story{config: suo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, suo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{story.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{err.Error(), err}
		}
		return nil, err
	}
	return _node, nil
}
