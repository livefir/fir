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

// BoardUpdate is the builder for updating Board entities.
type BoardUpdate struct {
	config
	hooks    []Hook
	mutation *BoardMutation
}

// Where appends a list predicates to the BoardUpdate builder.
func (bu *BoardUpdate) Where(ps ...predicate.Board) *BoardUpdate {
	bu.mutation.Where(ps...)
	return bu
}

// SetUpdateTime sets the "update_time" field.
func (bu *BoardUpdate) SetUpdateTime(t time.Time) *BoardUpdate {
	bu.mutation.SetUpdateTime(t)
	return bu
}

// SetTitle sets the "title" field.
func (bu *BoardUpdate) SetTitle(s string) *BoardUpdate {
	bu.mutation.SetTitle(s)
	return bu
}

// SetDescription sets the "description" field.
func (bu *BoardUpdate) SetDescription(s string) *BoardUpdate {
	bu.mutation.SetDescription(s)
	return bu
}

// AddStoryIDs adds the "stories" edge to the Story entity by IDs.
func (bu *BoardUpdate) AddStoryIDs(ids ...uuid.UUID) *BoardUpdate {
	bu.mutation.AddStoryIDs(ids...)
	return bu
}

// AddStories adds the "stories" edges to the Story entity.
func (bu *BoardUpdate) AddStories(s ...*Story) *BoardUpdate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return bu.AddStoryIDs(ids...)
}

// Mutation returns the BoardMutation object of the builder.
func (bu *BoardUpdate) Mutation() *BoardMutation {
	return bu.mutation
}

// ClearStories clears all "stories" edges to the Story entity.
func (bu *BoardUpdate) ClearStories() *BoardUpdate {
	bu.mutation.ClearStories()
	return bu
}

// RemoveStoryIDs removes the "stories" edge to Story entities by IDs.
func (bu *BoardUpdate) RemoveStoryIDs(ids ...uuid.UUID) *BoardUpdate {
	bu.mutation.RemoveStoryIDs(ids...)
	return bu
}

// RemoveStories removes "stories" edges to Story entities.
func (bu *BoardUpdate) RemoveStories(s ...*Story) *BoardUpdate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return bu.RemoveStoryIDs(ids...)
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (bu *BoardUpdate) Save(ctx context.Context) (int, error) {
	var (
		err      error
		affected int
	)
	bu.defaults()
	if len(bu.hooks) == 0 {
		if err = bu.check(); err != nil {
			return 0, err
		}
		affected, err = bu.sqlSave(ctx)
	} else {
		var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
			mutation, ok := m.(*BoardMutation)
			if !ok {
				return nil, fmt.Errorf("unexpected mutation type %T", m)
			}
			if err = bu.check(); err != nil {
				return 0, err
			}
			bu.mutation = mutation
			affected, err = bu.sqlSave(ctx)
			mutation.done = true
			return affected, err
		})
		for i := len(bu.hooks) - 1; i >= 0; i-- {
			if bu.hooks[i] == nil {
				return 0, fmt.Errorf("models: uninitialized hook (forgotten import models/runtime?)")
			}
			mut = bu.hooks[i](mut)
		}
		if _, err := mut.Mutate(ctx, bu.mutation); err != nil {
			return 0, err
		}
	}
	return affected, err
}

// SaveX is like Save, but panics if an error occurs.
func (bu *BoardUpdate) SaveX(ctx context.Context) int {
	affected, err := bu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (bu *BoardUpdate) Exec(ctx context.Context) error {
	_, err := bu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (bu *BoardUpdate) ExecX(ctx context.Context) {
	if err := bu.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (bu *BoardUpdate) defaults() {
	if _, ok := bu.mutation.UpdateTime(); !ok {
		v := board.UpdateDefaultUpdateTime()
		bu.mutation.SetUpdateTime(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (bu *BoardUpdate) check() error {
	if v, ok := bu.mutation.Title(); ok {
		if err := board.TitleValidator(v); err != nil {
			return &ValidationError{Name: "title", err: fmt.Errorf(`models: validator failed for field "Board.title": %w`, err)}
		}
	}
	if v, ok := bu.mutation.Description(); ok {
		if err := board.DescriptionValidator(v); err != nil {
			return &ValidationError{Name: "description", err: fmt.Errorf(`models: validator failed for field "Board.description": %w`, err)}
		}
	}
	return nil
}

func (bu *BoardUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := &sqlgraph.UpdateSpec{
		Node: &sqlgraph.NodeSpec{
			Table:   board.Table,
			Columns: board.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeUUID,
				Column: board.FieldID,
			},
		},
	}
	if ps := bu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := bu.mutation.UpdateTime(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeTime,
			Value:  value,
			Column: board.FieldUpdateTime,
		})
	}
	if value, ok := bu.mutation.Title(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: board.FieldTitle,
		})
	}
	if value, ok := bu.mutation.Description(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: board.FieldDescription,
		})
	}
	if bu.mutation.StoriesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   board.StoriesTable,
			Columns: []string{board.StoriesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: &sqlgraph.FieldSpec{
					Type:   field.TypeUUID,
					Column: story.FieldID,
				},
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := bu.mutation.RemovedStoriesIDs(); len(nodes) > 0 && !bu.mutation.StoriesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   board.StoriesTable,
			Columns: []string{board.StoriesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: &sqlgraph.FieldSpec{
					Type:   field.TypeUUID,
					Column: story.FieldID,
				},
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := bu.mutation.StoriesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   board.StoriesTable,
			Columns: []string{board.StoriesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: &sqlgraph.FieldSpec{
					Type:   field.TypeUUID,
					Column: story.FieldID,
				},
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, bu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{board.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{err.Error(), err}
		}
		return 0, err
	}
	return n, nil
}

// BoardUpdateOne is the builder for updating a single Board entity.
type BoardUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *BoardMutation
}

// SetUpdateTime sets the "update_time" field.
func (buo *BoardUpdateOne) SetUpdateTime(t time.Time) *BoardUpdateOne {
	buo.mutation.SetUpdateTime(t)
	return buo
}

// SetTitle sets the "title" field.
func (buo *BoardUpdateOne) SetTitle(s string) *BoardUpdateOne {
	buo.mutation.SetTitle(s)
	return buo
}

// SetDescription sets the "description" field.
func (buo *BoardUpdateOne) SetDescription(s string) *BoardUpdateOne {
	buo.mutation.SetDescription(s)
	return buo
}

// AddStoryIDs adds the "stories" edge to the Story entity by IDs.
func (buo *BoardUpdateOne) AddStoryIDs(ids ...uuid.UUID) *BoardUpdateOne {
	buo.mutation.AddStoryIDs(ids...)
	return buo
}

// AddStories adds the "stories" edges to the Story entity.
func (buo *BoardUpdateOne) AddStories(s ...*Story) *BoardUpdateOne {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return buo.AddStoryIDs(ids...)
}

// Mutation returns the BoardMutation object of the builder.
func (buo *BoardUpdateOne) Mutation() *BoardMutation {
	return buo.mutation
}

// ClearStories clears all "stories" edges to the Story entity.
func (buo *BoardUpdateOne) ClearStories() *BoardUpdateOne {
	buo.mutation.ClearStories()
	return buo
}

// RemoveStoryIDs removes the "stories" edge to Story entities by IDs.
func (buo *BoardUpdateOne) RemoveStoryIDs(ids ...uuid.UUID) *BoardUpdateOne {
	buo.mutation.RemoveStoryIDs(ids...)
	return buo
}

// RemoveStories removes "stories" edges to Story entities.
func (buo *BoardUpdateOne) RemoveStories(s ...*Story) *BoardUpdateOne {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return buo.RemoveStoryIDs(ids...)
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (buo *BoardUpdateOne) Select(field string, fields ...string) *BoardUpdateOne {
	buo.fields = append([]string{field}, fields...)
	return buo
}

// Save executes the query and returns the updated Board entity.
func (buo *BoardUpdateOne) Save(ctx context.Context) (*Board, error) {
	var (
		err  error
		node *Board
	)
	buo.defaults()
	if len(buo.hooks) == 0 {
		if err = buo.check(); err != nil {
			return nil, err
		}
		node, err = buo.sqlSave(ctx)
	} else {
		var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
			mutation, ok := m.(*BoardMutation)
			if !ok {
				return nil, fmt.Errorf("unexpected mutation type %T", m)
			}
			if err = buo.check(); err != nil {
				return nil, err
			}
			buo.mutation = mutation
			node, err = buo.sqlSave(ctx)
			mutation.done = true
			return node, err
		})
		for i := len(buo.hooks) - 1; i >= 0; i-- {
			if buo.hooks[i] == nil {
				return nil, fmt.Errorf("models: uninitialized hook (forgotten import models/runtime?)")
			}
			mut = buo.hooks[i](mut)
		}
		if _, err := mut.Mutate(ctx, buo.mutation); err != nil {
			return nil, err
		}
	}
	return node, err
}

// SaveX is like Save, but panics if an error occurs.
func (buo *BoardUpdateOne) SaveX(ctx context.Context) *Board {
	node, err := buo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (buo *BoardUpdateOne) Exec(ctx context.Context) error {
	_, err := buo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (buo *BoardUpdateOne) ExecX(ctx context.Context) {
	if err := buo.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (buo *BoardUpdateOne) defaults() {
	if _, ok := buo.mutation.UpdateTime(); !ok {
		v := board.UpdateDefaultUpdateTime()
		buo.mutation.SetUpdateTime(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (buo *BoardUpdateOne) check() error {
	if v, ok := buo.mutation.Title(); ok {
		if err := board.TitleValidator(v); err != nil {
			return &ValidationError{Name: "title", err: fmt.Errorf(`models: validator failed for field "Board.title": %w`, err)}
		}
	}
	if v, ok := buo.mutation.Description(); ok {
		if err := board.DescriptionValidator(v); err != nil {
			return &ValidationError{Name: "description", err: fmt.Errorf(`models: validator failed for field "Board.description": %w`, err)}
		}
	}
	return nil
}

func (buo *BoardUpdateOne) sqlSave(ctx context.Context) (_node *Board, err error) {
	_spec := &sqlgraph.UpdateSpec{
		Node: &sqlgraph.NodeSpec{
			Table:   board.Table,
			Columns: board.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeUUID,
				Column: board.FieldID,
			},
		},
	}
	id, ok := buo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`models: missing "Board.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := buo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, board.FieldID)
		for _, f := range fields {
			if !board.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("models: invalid field %q for query", f)}
			}
			if f != board.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := buo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := buo.mutation.UpdateTime(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeTime,
			Value:  value,
			Column: board.FieldUpdateTime,
		})
	}
	if value, ok := buo.mutation.Title(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: board.FieldTitle,
		})
	}
	if value, ok := buo.mutation.Description(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: board.FieldDescription,
		})
	}
	if buo.mutation.StoriesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   board.StoriesTable,
			Columns: []string{board.StoriesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: &sqlgraph.FieldSpec{
					Type:   field.TypeUUID,
					Column: story.FieldID,
				},
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := buo.mutation.RemovedStoriesIDs(); len(nodes) > 0 && !buo.mutation.StoriesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   board.StoriesTable,
			Columns: []string{board.StoriesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: &sqlgraph.FieldSpec{
					Type:   field.TypeUUID,
					Column: story.FieldID,
				},
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := buo.mutation.StoriesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   board.StoriesTable,
			Columns: []string{board.StoriesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: &sqlgraph.FieldSpec{
					Type:   field.TypeUUID,
					Column: story.FieldID,
				},
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_node = &Board{config: buo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, buo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{board.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{err.Error(), err}
		}
		return nil, err
	}
	return _node, nil
}
