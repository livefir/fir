// Code generated (@generated) by entc, DO NOT EDIT.

package models

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/adnaan/autobahn/models/board"
	"github.com/adnaan/autobahn/models/comment"
	"github.com/adnaan/autobahn/models/label"
	"github.com/adnaan/autobahn/models/predicate"
	"github.com/adnaan/autobahn/models/story"
	"github.com/adnaan/autobahn/models/view"
	"github.com/google/uuid"

	"entgo.io/ent"
)

const (
	// Operation types.
	OpCreate    = ent.OpCreate
	OpDelete    = ent.OpDelete
	OpDeleteOne = ent.OpDeleteOne
	OpUpdate    = ent.OpUpdate
	OpUpdateOne = ent.OpUpdateOne

	// Node types.
	TypeBoard   = "Board"
	TypeComment = "Comment"
	TypeLabel   = "Label"
	TypeStory   = "Story"
	TypeView    = "View"
)

// BoardMutation represents an operation that mutates the Board nodes in the graph.
type BoardMutation struct {
	config
	op             Op
	typ            string
	id             *uuid.UUID
	create_time    *time.Time
	update_time    *time.Time
	title          *string
	description    *string
	clearedFields  map[string]struct{}
	stories        map[uuid.UUID]struct{}
	removedstories map[uuid.UUID]struct{}
	clearedstories bool
	done           bool
	oldValue       func(context.Context) (*Board, error)
	predicates     []predicate.Board
}

var _ ent.Mutation = (*BoardMutation)(nil)

// boardOption allows management of the mutation configuration using functional options.
type boardOption func(*BoardMutation)

// newBoardMutation creates new mutation for the Board entity.
func newBoardMutation(c config, op Op, opts ...boardOption) *BoardMutation {
	m := &BoardMutation{
		config:        c,
		op:            op,
		typ:           TypeBoard,
		clearedFields: make(map[string]struct{}),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// withBoardID sets the ID field of the mutation.
func withBoardID(id uuid.UUID) boardOption {
	return func(m *BoardMutation) {
		var (
			err   error
			once  sync.Once
			value *Board
		)
		m.oldValue = func(ctx context.Context) (*Board, error) {
			once.Do(func() {
				if m.done {
					err = errors.New("querying old values post mutation is not allowed")
				} else {
					value, err = m.Client().Board.Get(ctx, id)
				}
			})
			return value, err
		}
		m.id = &id
	}
}

// withBoard sets the old Board of the mutation.
func withBoard(node *Board) boardOption {
	return func(m *BoardMutation) {
		m.oldValue = func(context.Context) (*Board, error) {
			return node, nil
		}
		m.id = &node.ID
	}
}

// Client returns a new `ent.Client` from the mutation. If the mutation was
// executed in a transaction (ent.Tx), a transactional client is returned.
func (m BoardMutation) Client() *Client {
	client := &Client{config: m.config}
	client.init()
	return client
}

// Tx returns an `ent.Tx` for mutations that were executed in transactions;
// it returns an error otherwise.
func (m BoardMutation) Tx() (*Tx, error) {
	if _, ok := m.driver.(*txDriver); !ok {
		return nil, errors.New("models: mutation is not running in a transaction")
	}
	tx := &Tx{config: m.config}
	tx.init()
	return tx, nil
}

// SetID sets the value of the id field. Note that this
// operation is only accepted on creation of Board entities.
func (m *BoardMutation) SetID(id uuid.UUID) {
	m.id = &id
}

// ID returns the ID value in the mutation. Note that the ID is only available
// if it was provided to the builder or after it was returned from the database.
func (m *BoardMutation) ID() (id uuid.UUID, exists bool) {
	if m.id == nil {
		return
	}
	return *m.id, true
}

// IDs queries the database and returns the entity ids that match the mutation's predicate.
// That means, if the mutation is applied within a transaction with an isolation level such
// as sql.LevelSerializable, the returned ids match the ids of the rows that will be updated
// or updated by the mutation.
func (m *BoardMutation) IDs(ctx context.Context) ([]uuid.UUID, error) {
	switch {
	case m.op.Is(OpUpdateOne | OpDeleteOne):
		id, exists := m.ID()
		if exists {
			return []uuid.UUID{id}, nil
		}
		fallthrough
	case m.op.Is(OpUpdate | OpDelete):
		return m.Client().Board.Query().Where(m.predicates...).IDs(ctx)
	default:
		return nil, fmt.Errorf("IDs is not allowed on %s operations", m.op)
	}
}

// SetCreateTime sets the "create_time" field.
func (m *BoardMutation) SetCreateTime(t time.Time) {
	m.create_time = &t
}

// CreateTime returns the value of the "create_time" field in the mutation.
func (m *BoardMutation) CreateTime() (r time.Time, exists bool) {
	v := m.create_time
	if v == nil {
		return
	}
	return *v, true
}

// OldCreateTime returns the old "create_time" field's value of the Board entity.
// If the Board object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *BoardMutation) OldCreateTime(ctx context.Context) (v time.Time, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldCreateTime is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldCreateTime requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldCreateTime: %w", err)
	}
	return oldValue.CreateTime, nil
}

// ResetCreateTime resets all changes to the "create_time" field.
func (m *BoardMutation) ResetCreateTime() {
	m.create_time = nil
}

// SetUpdateTime sets the "update_time" field.
func (m *BoardMutation) SetUpdateTime(t time.Time) {
	m.update_time = &t
}

// UpdateTime returns the value of the "update_time" field in the mutation.
func (m *BoardMutation) UpdateTime() (r time.Time, exists bool) {
	v := m.update_time
	if v == nil {
		return
	}
	return *v, true
}

// OldUpdateTime returns the old "update_time" field's value of the Board entity.
// If the Board object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *BoardMutation) OldUpdateTime(ctx context.Context) (v time.Time, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldUpdateTime is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldUpdateTime requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldUpdateTime: %w", err)
	}
	return oldValue.UpdateTime, nil
}

// ResetUpdateTime resets all changes to the "update_time" field.
func (m *BoardMutation) ResetUpdateTime() {
	m.update_time = nil
}

// SetTitle sets the "title" field.
func (m *BoardMutation) SetTitle(s string) {
	m.title = &s
}

// Title returns the value of the "title" field in the mutation.
func (m *BoardMutation) Title() (r string, exists bool) {
	v := m.title
	if v == nil {
		return
	}
	return *v, true
}

// OldTitle returns the old "title" field's value of the Board entity.
// If the Board object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *BoardMutation) OldTitle(ctx context.Context) (v string, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldTitle is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldTitle requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldTitle: %w", err)
	}
	return oldValue.Title, nil
}

// ResetTitle resets all changes to the "title" field.
func (m *BoardMutation) ResetTitle() {
	m.title = nil
}

// SetDescription sets the "description" field.
func (m *BoardMutation) SetDescription(s string) {
	m.description = &s
}

// Description returns the value of the "description" field in the mutation.
func (m *BoardMutation) Description() (r string, exists bool) {
	v := m.description
	if v == nil {
		return
	}
	return *v, true
}

// OldDescription returns the old "description" field's value of the Board entity.
// If the Board object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *BoardMutation) OldDescription(ctx context.Context) (v string, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldDescription is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldDescription requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldDescription: %w", err)
	}
	return oldValue.Description, nil
}

// ResetDescription resets all changes to the "description" field.
func (m *BoardMutation) ResetDescription() {
	m.description = nil
}

// AddStoryIDs adds the "stories" edge to the Story entity by ids.
func (m *BoardMutation) AddStoryIDs(ids ...uuid.UUID) {
	if m.stories == nil {
		m.stories = make(map[uuid.UUID]struct{})
	}
	for i := range ids {
		m.stories[ids[i]] = struct{}{}
	}
}

// ClearStories clears the "stories" edge to the Story entity.
func (m *BoardMutation) ClearStories() {
	m.clearedstories = true
}

// StoriesCleared reports if the "stories" edge to the Story entity was cleared.
func (m *BoardMutation) StoriesCleared() bool {
	return m.clearedstories
}

// RemoveStoryIDs removes the "stories" edge to the Story entity by IDs.
func (m *BoardMutation) RemoveStoryIDs(ids ...uuid.UUID) {
	if m.removedstories == nil {
		m.removedstories = make(map[uuid.UUID]struct{})
	}
	for i := range ids {
		delete(m.stories, ids[i])
		m.removedstories[ids[i]] = struct{}{}
	}
}

// RemovedStories returns the removed IDs of the "stories" edge to the Story entity.
func (m *BoardMutation) RemovedStoriesIDs() (ids []uuid.UUID) {
	for id := range m.removedstories {
		ids = append(ids, id)
	}
	return
}

// StoriesIDs returns the "stories" edge IDs in the mutation.
func (m *BoardMutation) StoriesIDs() (ids []uuid.UUID) {
	for id := range m.stories {
		ids = append(ids, id)
	}
	return
}

// ResetStories resets all changes to the "stories" edge.
func (m *BoardMutation) ResetStories() {
	m.stories = nil
	m.clearedstories = false
	m.removedstories = nil
}

// Where appends a list predicates to the BoardMutation builder.
func (m *BoardMutation) Where(ps ...predicate.Board) {
	m.predicates = append(m.predicates, ps...)
}

// Op returns the operation name.
func (m *BoardMutation) Op() Op {
	return m.op
}

// Type returns the node type of this mutation (Board).
func (m *BoardMutation) Type() string {
	return m.typ
}

// Fields returns all fields that were changed during this mutation. Note that in
// order to get all numeric fields that were incremented/decremented, call
// AddedFields().
func (m *BoardMutation) Fields() []string {
	fields := make([]string, 0, 4)
	if m.create_time != nil {
		fields = append(fields, board.FieldCreateTime)
	}
	if m.update_time != nil {
		fields = append(fields, board.FieldUpdateTime)
	}
	if m.title != nil {
		fields = append(fields, board.FieldTitle)
	}
	if m.description != nil {
		fields = append(fields, board.FieldDescription)
	}
	return fields
}

// Field returns the value of a field with the given name. The second boolean
// return value indicates that this field was not set, or was not defined in the
// schema.
func (m *BoardMutation) Field(name string) (ent.Value, bool) {
	switch name {
	case board.FieldCreateTime:
		return m.CreateTime()
	case board.FieldUpdateTime:
		return m.UpdateTime()
	case board.FieldTitle:
		return m.Title()
	case board.FieldDescription:
		return m.Description()
	}
	return nil, false
}

// OldField returns the old value of the field from the database. An error is
// returned if the mutation operation is not UpdateOne, or the query to the
// database failed.
func (m *BoardMutation) OldField(ctx context.Context, name string) (ent.Value, error) {
	switch name {
	case board.FieldCreateTime:
		return m.OldCreateTime(ctx)
	case board.FieldUpdateTime:
		return m.OldUpdateTime(ctx)
	case board.FieldTitle:
		return m.OldTitle(ctx)
	case board.FieldDescription:
		return m.OldDescription(ctx)
	}
	return nil, fmt.Errorf("unknown Board field %s", name)
}

// SetField sets the value of a field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *BoardMutation) SetField(name string, value ent.Value) error {
	switch name {
	case board.FieldCreateTime:
		v, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetCreateTime(v)
		return nil
	case board.FieldUpdateTime:
		v, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetUpdateTime(v)
		return nil
	case board.FieldTitle:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetTitle(v)
		return nil
	case board.FieldDescription:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetDescription(v)
		return nil
	}
	return fmt.Errorf("unknown Board field %s", name)
}

// AddedFields returns all numeric fields that were incremented/decremented during
// this mutation.
func (m *BoardMutation) AddedFields() []string {
	return nil
}

// AddedField returns the numeric value that was incremented/decremented on a field
// with the given name. The second boolean return value indicates that this field
// was not set, or was not defined in the schema.
func (m *BoardMutation) AddedField(name string) (ent.Value, bool) {
	return nil, false
}

// AddField adds the value to the field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *BoardMutation) AddField(name string, value ent.Value) error {
	switch name {
	}
	return fmt.Errorf("unknown Board numeric field %s", name)
}

// ClearedFields returns all nullable fields that were cleared during this
// mutation.
func (m *BoardMutation) ClearedFields() []string {
	return nil
}

// FieldCleared returns a boolean indicating if a field with the given name was
// cleared in this mutation.
func (m *BoardMutation) FieldCleared(name string) bool {
	_, ok := m.clearedFields[name]
	return ok
}

// ClearField clears the value of the field with the given name. It returns an
// error if the field is not defined in the schema.
func (m *BoardMutation) ClearField(name string) error {
	return fmt.Errorf("unknown Board nullable field %s", name)
}

// ResetField resets all changes in the mutation for the field with the given name.
// It returns an error if the field is not defined in the schema.
func (m *BoardMutation) ResetField(name string) error {
	switch name {
	case board.FieldCreateTime:
		m.ResetCreateTime()
		return nil
	case board.FieldUpdateTime:
		m.ResetUpdateTime()
		return nil
	case board.FieldTitle:
		m.ResetTitle()
		return nil
	case board.FieldDescription:
		m.ResetDescription()
		return nil
	}
	return fmt.Errorf("unknown Board field %s", name)
}

// AddedEdges returns all edge names that were set/added in this mutation.
func (m *BoardMutation) AddedEdges() []string {
	edges := make([]string, 0, 1)
	if m.stories != nil {
		edges = append(edges, board.EdgeStories)
	}
	return edges
}

// AddedIDs returns all IDs (to other nodes) that were added for the given edge
// name in this mutation.
func (m *BoardMutation) AddedIDs(name string) []ent.Value {
	switch name {
	case board.EdgeStories:
		ids := make([]ent.Value, 0, len(m.stories))
		for id := range m.stories {
			ids = append(ids, id)
		}
		return ids
	}
	return nil
}

// RemovedEdges returns all edge names that were removed in this mutation.
func (m *BoardMutation) RemovedEdges() []string {
	edges := make([]string, 0, 1)
	if m.removedstories != nil {
		edges = append(edges, board.EdgeStories)
	}
	return edges
}

// RemovedIDs returns all IDs (to other nodes) that were removed for the edge with
// the given name in this mutation.
func (m *BoardMutation) RemovedIDs(name string) []ent.Value {
	switch name {
	case board.EdgeStories:
		ids := make([]ent.Value, 0, len(m.removedstories))
		for id := range m.removedstories {
			ids = append(ids, id)
		}
		return ids
	}
	return nil
}

// ClearedEdges returns all edge names that were cleared in this mutation.
func (m *BoardMutation) ClearedEdges() []string {
	edges := make([]string, 0, 1)
	if m.clearedstories {
		edges = append(edges, board.EdgeStories)
	}
	return edges
}

// EdgeCleared returns a boolean which indicates if the edge with the given name
// was cleared in this mutation.
func (m *BoardMutation) EdgeCleared(name string) bool {
	switch name {
	case board.EdgeStories:
		return m.clearedstories
	}
	return false
}

// ClearEdge clears the value of the edge with the given name. It returns an error
// if that edge is not defined in the schema.
func (m *BoardMutation) ClearEdge(name string) error {
	switch name {
	}
	return fmt.Errorf("unknown Board unique edge %s", name)
}

// ResetEdge resets all changes to the edge with the given name in this mutation.
// It returns an error if the edge is not defined in the schema.
func (m *BoardMutation) ResetEdge(name string) error {
	switch name {
	case board.EdgeStories:
		m.ResetStories()
		return nil
	}
	return fmt.Errorf("unknown Board edge %s", name)
}

// CommentMutation represents an operation that mutates the Comment nodes in the graph.
type CommentMutation struct {
	config
	op            Op
	typ           string
	id            *uuid.UUID
	create_time   *time.Time
	update_time   *time.Time
	clearedFields map[string]struct{}
	done          bool
	oldValue      func(context.Context) (*Comment, error)
	predicates    []predicate.Comment
}

var _ ent.Mutation = (*CommentMutation)(nil)

// commentOption allows management of the mutation configuration using functional options.
type commentOption func(*CommentMutation)

// newCommentMutation creates new mutation for the Comment entity.
func newCommentMutation(c config, op Op, opts ...commentOption) *CommentMutation {
	m := &CommentMutation{
		config:        c,
		op:            op,
		typ:           TypeComment,
		clearedFields: make(map[string]struct{}),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// withCommentID sets the ID field of the mutation.
func withCommentID(id uuid.UUID) commentOption {
	return func(m *CommentMutation) {
		var (
			err   error
			once  sync.Once
			value *Comment
		)
		m.oldValue = func(ctx context.Context) (*Comment, error) {
			once.Do(func() {
				if m.done {
					err = errors.New("querying old values post mutation is not allowed")
				} else {
					value, err = m.Client().Comment.Get(ctx, id)
				}
			})
			return value, err
		}
		m.id = &id
	}
}

// withComment sets the old Comment of the mutation.
func withComment(node *Comment) commentOption {
	return func(m *CommentMutation) {
		m.oldValue = func(context.Context) (*Comment, error) {
			return node, nil
		}
		m.id = &node.ID
	}
}

// Client returns a new `ent.Client` from the mutation. If the mutation was
// executed in a transaction (ent.Tx), a transactional client is returned.
func (m CommentMutation) Client() *Client {
	client := &Client{config: m.config}
	client.init()
	return client
}

// Tx returns an `ent.Tx` for mutations that were executed in transactions;
// it returns an error otherwise.
func (m CommentMutation) Tx() (*Tx, error) {
	if _, ok := m.driver.(*txDriver); !ok {
		return nil, errors.New("models: mutation is not running in a transaction")
	}
	tx := &Tx{config: m.config}
	tx.init()
	return tx, nil
}

// SetID sets the value of the id field. Note that this
// operation is only accepted on creation of Comment entities.
func (m *CommentMutation) SetID(id uuid.UUID) {
	m.id = &id
}

// ID returns the ID value in the mutation. Note that the ID is only available
// if it was provided to the builder or after it was returned from the database.
func (m *CommentMutation) ID() (id uuid.UUID, exists bool) {
	if m.id == nil {
		return
	}
	return *m.id, true
}

// IDs queries the database and returns the entity ids that match the mutation's predicate.
// That means, if the mutation is applied within a transaction with an isolation level such
// as sql.LevelSerializable, the returned ids match the ids of the rows that will be updated
// or updated by the mutation.
func (m *CommentMutation) IDs(ctx context.Context) ([]uuid.UUID, error) {
	switch {
	case m.op.Is(OpUpdateOne | OpDeleteOne):
		id, exists := m.ID()
		if exists {
			return []uuid.UUID{id}, nil
		}
		fallthrough
	case m.op.Is(OpUpdate | OpDelete):
		return m.Client().Comment.Query().Where(m.predicates...).IDs(ctx)
	default:
		return nil, fmt.Errorf("IDs is not allowed on %s operations", m.op)
	}
}

// SetCreateTime sets the "create_time" field.
func (m *CommentMutation) SetCreateTime(t time.Time) {
	m.create_time = &t
}

// CreateTime returns the value of the "create_time" field in the mutation.
func (m *CommentMutation) CreateTime() (r time.Time, exists bool) {
	v := m.create_time
	if v == nil {
		return
	}
	return *v, true
}

// OldCreateTime returns the old "create_time" field's value of the Comment entity.
// If the Comment object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *CommentMutation) OldCreateTime(ctx context.Context) (v time.Time, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldCreateTime is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldCreateTime requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldCreateTime: %w", err)
	}
	return oldValue.CreateTime, nil
}

// ResetCreateTime resets all changes to the "create_time" field.
func (m *CommentMutation) ResetCreateTime() {
	m.create_time = nil
}

// SetUpdateTime sets the "update_time" field.
func (m *CommentMutation) SetUpdateTime(t time.Time) {
	m.update_time = &t
}

// UpdateTime returns the value of the "update_time" field in the mutation.
func (m *CommentMutation) UpdateTime() (r time.Time, exists bool) {
	v := m.update_time
	if v == nil {
		return
	}
	return *v, true
}

// OldUpdateTime returns the old "update_time" field's value of the Comment entity.
// If the Comment object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *CommentMutation) OldUpdateTime(ctx context.Context) (v time.Time, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldUpdateTime is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldUpdateTime requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldUpdateTime: %w", err)
	}
	return oldValue.UpdateTime, nil
}

// ResetUpdateTime resets all changes to the "update_time" field.
func (m *CommentMutation) ResetUpdateTime() {
	m.update_time = nil
}

// Where appends a list predicates to the CommentMutation builder.
func (m *CommentMutation) Where(ps ...predicate.Comment) {
	m.predicates = append(m.predicates, ps...)
}

// Op returns the operation name.
func (m *CommentMutation) Op() Op {
	return m.op
}

// Type returns the node type of this mutation (Comment).
func (m *CommentMutation) Type() string {
	return m.typ
}

// Fields returns all fields that were changed during this mutation. Note that in
// order to get all numeric fields that were incremented/decremented, call
// AddedFields().
func (m *CommentMutation) Fields() []string {
	fields := make([]string, 0, 2)
	if m.create_time != nil {
		fields = append(fields, comment.FieldCreateTime)
	}
	if m.update_time != nil {
		fields = append(fields, comment.FieldUpdateTime)
	}
	return fields
}

// Field returns the value of a field with the given name. The second boolean
// return value indicates that this field was not set, or was not defined in the
// schema.
func (m *CommentMutation) Field(name string) (ent.Value, bool) {
	switch name {
	case comment.FieldCreateTime:
		return m.CreateTime()
	case comment.FieldUpdateTime:
		return m.UpdateTime()
	}
	return nil, false
}

// OldField returns the old value of the field from the database. An error is
// returned if the mutation operation is not UpdateOne, or the query to the
// database failed.
func (m *CommentMutation) OldField(ctx context.Context, name string) (ent.Value, error) {
	switch name {
	case comment.FieldCreateTime:
		return m.OldCreateTime(ctx)
	case comment.FieldUpdateTime:
		return m.OldUpdateTime(ctx)
	}
	return nil, fmt.Errorf("unknown Comment field %s", name)
}

// SetField sets the value of a field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *CommentMutation) SetField(name string, value ent.Value) error {
	switch name {
	case comment.FieldCreateTime:
		v, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetCreateTime(v)
		return nil
	case comment.FieldUpdateTime:
		v, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetUpdateTime(v)
		return nil
	}
	return fmt.Errorf("unknown Comment field %s", name)
}

// AddedFields returns all numeric fields that were incremented/decremented during
// this mutation.
func (m *CommentMutation) AddedFields() []string {
	return nil
}

// AddedField returns the numeric value that was incremented/decremented on a field
// with the given name. The second boolean return value indicates that this field
// was not set, or was not defined in the schema.
func (m *CommentMutation) AddedField(name string) (ent.Value, bool) {
	return nil, false
}

// AddField adds the value to the field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *CommentMutation) AddField(name string, value ent.Value) error {
	switch name {
	}
	return fmt.Errorf("unknown Comment numeric field %s", name)
}

// ClearedFields returns all nullable fields that were cleared during this
// mutation.
func (m *CommentMutation) ClearedFields() []string {
	return nil
}

// FieldCleared returns a boolean indicating if a field with the given name was
// cleared in this mutation.
func (m *CommentMutation) FieldCleared(name string) bool {
	_, ok := m.clearedFields[name]
	return ok
}

// ClearField clears the value of the field with the given name. It returns an
// error if the field is not defined in the schema.
func (m *CommentMutation) ClearField(name string) error {
	return fmt.Errorf("unknown Comment nullable field %s", name)
}

// ResetField resets all changes in the mutation for the field with the given name.
// It returns an error if the field is not defined in the schema.
func (m *CommentMutation) ResetField(name string) error {
	switch name {
	case comment.FieldCreateTime:
		m.ResetCreateTime()
		return nil
	case comment.FieldUpdateTime:
		m.ResetUpdateTime()
		return nil
	}
	return fmt.Errorf("unknown Comment field %s", name)
}

// AddedEdges returns all edge names that were set/added in this mutation.
func (m *CommentMutation) AddedEdges() []string {
	edges := make([]string, 0, 0)
	return edges
}

// AddedIDs returns all IDs (to other nodes) that were added for the given edge
// name in this mutation.
func (m *CommentMutation) AddedIDs(name string) []ent.Value {
	return nil
}

// RemovedEdges returns all edge names that were removed in this mutation.
func (m *CommentMutation) RemovedEdges() []string {
	edges := make([]string, 0, 0)
	return edges
}

// RemovedIDs returns all IDs (to other nodes) that were removed for the edge with
// the given name in this mutation.
func (m *CommentMutation) RemovedIDs(name string) []ent.Value {
	return nil
}

// ClearedEdges returns all edge names that were cleared in this mutation.
func (m *CommentMutation) ClearedEdges() []string {
	edges := make([]string, 0, 0)
	return edges
}

// EdgeCleared returns a boolean which indicates if the edge with the given name
// was cleared in this mutation.
func (m *CommentMutation) EdgeCleared(name string) bool {
	return false
}

// ClearEdge clears the value of the edge with the given name. It returns an error
// if that edge is not defined in the schema.
func (m *CommentMutation) ClearEdge(name string) error {
	return fmt.Errorf("unknown Comment unique edge %s", name)
}

// ResetEdge resets all changes to the edge with the given name in this mutation.
// It returns an error if the edge is not defined in the schema.
func (m *CommentMutation) ResetEdge(name string) error {
	return fmt.Errorf("unknown Comment edge %s", name)
}

// LabelMutation represents an operation that mutates the Label nodes in the graph.
type LabelMutation struct {
	config
	op            Op
	typ           string
	id            *uuid.UUID
	create_time   *time.Time
	update_time   *time.Time
	clearedFields map[string]struct{}
	done          bool
	oldValue      func(context.Context) (*Label, error)
	predicates    []predicate.Label
}

var _ ent.Mutation = (*LabelMutation)(nil)

// labelOption allows management of the mutation configuration using functional options.
type labelOption func(*LabelMutation)

// newLabelMutation creates new mutation for the Label entity.
func newLabelMutation(c config, op Op, opts ...labelOption) *LabelMutation {
	m := &LabelMutation{
		config:        c,
		op:            op,
		typ:           TypeLabel,
		clearedFields: make(map[string]struct{}),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// withLabelID sets the ID field of the mutation.
func withLabelID(id uuid.UUID) labelOption {
	return func(m *LabelMutation) {
		var (
			err   error
			once  sync.Once
			value *Label
		)
		m.oldValue = func(ctx context.Context) (*Label, error) {
			once.Do(func() {
				if m.done {
					err = errors.New("querying old values post mutation is not allowed")
				} else {
					value, err = m.Client().Label.Get(ctx, id)
				}
			})
			return value, err
		}
		m.id = &id
	}
}

// withLabel sets the old Label of the mutation.
func withLabel(node *Label) labelOption {
	return func(m *LabelMutation) {
		m.oldValue = func(context.Context) (*Label, error) {
			return node, nil
		}
		m.id = &node.ID
	}
}

// Client returns a new `ent.Client` from the mutation. If the mutation was
// executed in a transaction (ent.Tx), a transactional client is returned.
func (m LabelMutation) Client() *Client {
	client := &Client{config: m.config}
	client.init()
	return client
}

// Tx returns an `ent.Tx` for mutations that were executed in transactions;
// it returns an error otherwise.
func (m LabelMutation) Tx() (*Tx, error) {
	if _, ok := m.driver.(*txDriver); !ok {
		return nil, errors.New("models: mutation is not running in a transaction")
	}
	tx := &Tx{config: m.config}
	tx.init()
	return tx, nil
}

// SetID sets the value of the id field. Note that this
// operation is only accepted on creation of Label entities.
func (m *LabelMutation) SetID(id uuid.UUID) {
	m.id = &id
}

// ID returns the ID value in the mutation. Note that the ID is only available
// if it was provided to the builder or after it was returned from the database.
func (m *LabelMutation) ID() (id uuid.UUID, exists bool) {
	if m.id == nil {
		return
	}
	return *m.id, true
}

// IDs queries the database and returns the entity ids that match the mutation's predicate.
// That means, if the mutation is applied within a transaction with an isolation level such
// as sql.LevelSerializable, the returned ids match the ids of the rows that will be updated
// or updated by the mutation.
func (m *LabelMutation) IDs(ctx context.Context) ([]uuid.UUID, error) {
	switch {
	case m.op.Is(OpUpdateOne | OpDeleteOne):
		id, exists := m.ID()
		if exists {
			return []uuid.UUID{id}, nil
		}
		fallthrough
	case m.op.Is(OpUpdate | OpDelete):
		return m.Client().Label.Query().Where(m.predicates...).IDs(ctx)
	default:
		return nil, fmt.Errorf("IDs is not allowed on %s operations", m.op)
	}
}

// SetCreateTime sets the "create_time" field.
func (m *LabelMutation) SetCreateTime(t time.Time) {
	m.create_time = &t
}

// CreateTime returns the value of the "create_time" field in the mutation.
func (m *LabelMutation) CreateTime() (r time.Time, exists bool) {
	v := m.create_time
	if v == nil {
		return
	}
	return *v, true
}

// OldCreateTime returns the old "create_time" field's value of the Label entity.
// If the Label object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *LabelMutation) OldCreateTime(ctx context.Context) (v time.Time, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldCreateTime is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldCreateTime requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldCreateTime: %w", err)
	}
	return oldValue.CreateTime, nil
}

// ResetCreateTime resets all changes to the "create_time" field.
func (m *LabelMutation) ResetCreateTime() {
	m.create_time = nil
}

// SetUpdateTime sets the "update_time" field.
func (m *LabelMutation) SetUpdateTime(t time.Time) {
	m.update_time = &t
}

// UpdateTime returns the value of the "update_time" field in the mutation.
func (m *LabelMutation) UpdateTime() (r time.Time, exists bool) {
	v := m.update_time
	if v == nil {
		return
	}
	return *v, true
}

// OldUpdateTime returns the old "update_time" field's value of the Label entity.
// If the Label object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *LabelMutation) OldUpdateTime(ctx context.Context) (v time.Time, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldUpdateTime is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldUpdateTime requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldUpdateTime: %w", err)
	}
	return oldValue.UpdateTime, nil
}

// ResetUpdateTime resets all changes to the "update_time" field.
func (m *LabelMutation) ResetUpdateTime() {
	m.update_time = nil
}

// Where appends a list predicates to the LabelMutation builder.
func (m *LabelMutation) Where(ps ...predicate.Label) {
	m.predicates = append(m.predicates, ps...)
}

// Op returns the operation name.
func (m *LabelMutation) Op() Op {
	return m.op
}

// Type returns the node type of this mutation (Label).
func (m *LabelMutation) Type() string {
	return m.typ
}

// Fields returns all fields that were changed during this mutation. Note that in
// order to get all numeric fields that were incremented/decremented, call
// AddedFields().
func (m *LabelMutation) Fields() []string {
	fields := make([]string, 0, 2)
	if m.create_time != nil {
		fields = append(fields, label.FieldCreateTime)
	}
	if m.update_time != nil {
		fields = append(fields, label.FieldUpdateTime)
	}
	return fields
}

// Field returns the value of a field with the given name. The second boolean
// return value indicates that this field was not set, or was not defined in the
// schema.
func (m *LabelMutation) Field(name string) (ent.Value, bool) {
	switch name {
	case label.FieldCreateTime:
		return m.CreateTime()
	case label.FieldUpdateTime:
		return m.UpdateTime()
	}
	return nil, false
}

// OldField returns the old value of the field from the database. An error is
// returned if the mutation operation is not UpdateOne, or the query to the
// database failed.
func (m *LabelMutation) OldField(ctx context.Context, name string) (ent.Value, error) {
	switch name {
	case label.FieldCreateTime:
		return m.OldCreateTime(ctx)
	case label.FieldUpdateTime:
		return m.OldUpdateTime(ctx)
	}
	return nil, fmt.Errorf("unknown Label field %s", name)
}

// SetField sets the value of a field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *LabelMutation) SetField(name string, value ent.Value) error {
	switch name {
	case label.FieldCreateTime:
		v, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetCreateTime(v)
		return nil
	case label.FieldUpdateTime:
		v, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetUpdateTime(v)
		return nil
	}
	return fmt.Errorf("unknown Label field %s", name)
}

// AddedFields returns all numeric fields that were incremented/decremented during
// this mutation.
func (m *LabelMutation) AddedFields() []string {
	return nil
}

// AddedField returns the numeric value that was incremented/decremented on a field
// with the given name. The second boolean return value indicates that this field
// was not set, or was not defined in the schema.
func (m *LabelMutation) AddedField(name string) (ent.Value, bool) {
	return nil, false
}

// AddField adds the value to the field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *LabelMutation) AddField(name string, value ent.Value) error {
	switch name {
	}
	return fmt.Errorf("unknown Label numeric field %s", name)
}

// ClearedFields returns all nullable fields that were cleared during this
// mutation.
func (m *LabelMutation) ClearedFields() []string {
	return nil
}

// FieldCleared returns a boolean indicating if a field with the given name was
// cleared in this mutation.
func (m *LabelMutation) FieldCleared(name string) bool {
	_, ok := m.clearedFields[name]
	return ok
}

// ClearField clears the value of the field with the given name. It returns an
// error if the field is not defined in the schema.
func (m *LabelMutation) ClearField(name string) error {
	return fmt.Errorf("unknown Label nullable field %s", name)
}

// ResetField resets all changes in the mutation for the field with the given name.
// It returns an error if the field is not defined in the schema.
func (m *LabelMutation) ResetField(name string) error {
	switch name {
	case label.FieldCreateTime:
		m.ResetCreateTime()
		return nil
	case label.FieldUpdateTime:
		m.ResetUpdateTime()
		return nil
	}
	return fmt.Errorf("unknown Label field %s", name)
}

// AddedEdges returns all edge names that were set/added in this mutation.
func (m *LabelMutation) AddedEdges() []string {
	edges := make([]string, 0, 0)
	return edges
}

// AddedIDs returns all IDs (to other nodes) that were added for the given edge
// name in this mutation.
func (m *LabelMutation) AddedIDs(name string) []ent.Value {
	return nil
}

// RemovedEdges returns all edge names that were removed in this mutation.
func (m *LabelMutation) RemovedEdges() []string {
	edges := make([]string, 0, 0)
	return edges
}

// RemovedIDs returns all IDs (to other nodes) that were removed for the edge with
// the given name in this mutation.
func (m *LabelMutation) RemovedIDs(name string) []ent.Value {
	return nil
}

// ClearedEdges returns all edge names that were cleared in this mutation.
func (m *LabelMutation) ClearedEdges() []string {
	edges := make([]string, 0, 0)
	return edges
}

// EdgeCleared returns a boolean which indicates if the edge with the given name
// was cleared in this mutation.
func (m *LabelMutation) EdgeCleared(name string) bool {
	return false
}

// ClearEdge clears the value of the edge with the given name. It returns an error
// if that edge is not defined in the schema.
func (m *LabelMutation) ClearEdge(name string) error {
	return fmt.Errorf("unknown Label unique edge %s", name)
}

// ResetEdge resets all changes to the edge with the given name in this mutation.
// It returns an error if the edge is not defined in the schema.
func (m *LabelMutation) ResetEdge(name string) error {
	return fmt.Errorf("unknown Label edge %s", name)
}

// StoryMutation represents an operation that mutates the Story nodes in the graph.
type StoryMutation struct {
	config
	op            Op
	typ           string
	id            *uuid.UUID
	create_time   *time.Time
	update_time   *time.Time
	title         *string
	description   *string
	clearedFields map[string]struct{}
	owner         *uuid.UUID
	clearedowner  bool
	done          bool
	oldValue      func(context.Context) (*Story, error)
	predicates    []predicate.Story
}

var _ ent.Mutation = (*StoryMutation)(nil)

// storyOption allows management of the mutation configuration using functional options.
type storyOption func(*StoryMutation)

// newStoryMutation creates new mutation for the Story entity.
func newStoryMutation(c config, op Op, opts ...storyOption) *StoryMutation {
	m := &StoryMutation{
		config:        c,
		op:            op,
		typ:           TypeStory,
		clearedFields: make(map[string]struct{}),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// withStoryID sets the ID field of the mutation.
func withStoryID(id uuid.UUID) storyOption {
	return func(m *StoryMutation) {
		var (
			err   error
			once  sync.Once
			value *Story
		)
		m.oldValue = func(ctx context.Context) (*Story, error) {
			once.Do(func() {
				if m.done {
					err = errors.New("querying old values post mutation is not allowed")
				} else {
					value, err = m.Client().Story.Get(ctx, id)
				}
			})
			return value, err
		}
		m.id = &id
	}
}

// withStory sets the old Story of the mutation.
func withStory(node *Story) storyOption {
	return func(m *StoryMutation) {
		m.oldValue = func(context.Context) (*Story, error) {
			return node, nil
		}
		m.id = &node.ID
	}
}

// Client returns a new `ent.Client` from the mutation. If the mutation was
// executed in a transaction (ent.Tx), a transactional client is returned.
func (m StoryMutation) Client() *Client {
	client := &Client{config: m.config}
	client.init()
	return client
}

// Tx returns an `ent.Tx` for mutations that were executed in transactions;
// it returns an error otherwise.
func (m StoryMutation) Tx() (*Tx, error) {
	if _, ok := m.driver.(*txDriver); !ok {
		return nil, errors.New("models: mutation is not running in a transaction")
	}
	tx := &Tx{config: m.config}
	tx.init()
	return tx, nil
}

// SetID sets the value of the id field. Note that this
// operation is only accepted on creation of Story entities.
func (m *StoryMutation) SetID(id uuid.UUID) {
	m.id = &id
}

// ID returns the ID value in the mutation. Note that the ID is only available
// if it was provided to the builder or after it was returned from the database.
func (m *StoryMutation) ID() (id uuid.UUID, exists bool) {
	if m.id == nil {
		return
	}
	return *m.id, true
}

// IDs queries the database and returns the entity ids that match the mutation's predicate.
// That means, if the mutation is applied within a transaction with an isolation level such
// as sql.LevelSerializable, the returned ids match the ids of the rows that will be updated
// or updated by the mutation.
func (m *StoryMutation) IDs(ctx context.Context) ([]uuid.UUID, error) {
	switch {
	case m.op.Is(OpUpdateOne | OpDeleteOne):
		id, exists := m.ID()
		if exists {
			return []uuid.UUID{id}, nil
		}
		fallthrough
	case m.op.Is(OpUpdate | OpDelete):
		return m.Client().Story.Query().Where(m.predicates...).IDs(ctx)
	default:
		return nil, fmt.Errorf("IDs is not allowed on %s operations", m.op)
	}
}

// SetCreateTime sets the "create_time" field.
func (m *StoryMutation) SetCreateTime(t time.Time) {
	m.create_time = &t
}

// CreateTime returns the value of the "create_time" field in the mutation.
func (m *StoryMutation) CreateTime() (r time.Time, exists bool) {
	v := m.create_time
	if v == nil {
		return
	}
	return *v, true
}

// OldCreateTime returns the old "create_time" field's value of the Story entity.
// If the Story object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *StoryMutation) OldCreateTime(ctx context.Context) (v time.Time, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldCreateTime is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldCreateTime requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldCreateTime: %w", err)
	}
	return oldValue.CreateTime, nil
}

// ResetCreateTime resets all changes to the "create_time" field.
func (m *StoryMutation) ResetCreateTime() {
	m.create_time = nil
}

// SetUpdateTime sets the "update_time" field.
func (m *StoryMutation) SetUpdateTime(t time.Time) {
	m.update_time = &t
}

// UpdateTime returns the value of the "update_time" field in the mutation.
func (m *StoryMutation) UpdateTime() (r time.Time, exists bool) {
	v := m.update_time
	if v == nil {
		return
	}
	return *v, true
}

// OldUpdateTime returns the old "update_time" field's value of the Story entity.
// If the Story object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *StoryMutation) OldUpdateTime(ctx context.Context) (v time.Time, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldUpdateTime is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldUpdateTime requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldUpdateTime: %w", err)
	}
	return oldValue.UpdateTime, nil
}

// ResetUpdateTime resets all changes to the "update_time" field.
func (m *StoryMutation) ResetUpdateTime() {
	m.update_time = nil
}

// SetTitle sets the "title" field.
func (m *StoryMutation) SetTitle(s string) {
	m.title = &s
}

// Title returns the value of the "title" field in the mutation.
func (m *StoryMutation) Title() (r string, exists bool) {
	v := m.title
	if v == nil {
		return
	}
	return *v, true
}

// OldTitle returns the old "title" field's value of the Story entity.
// If the Story object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *StoryMutation) OldTitle(ctx context.Context) (v string, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldTitle is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldTitle requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldTitle: %w", err)
	}
	return oldValue.Title, nil
}

// ResetTitle resets all changes to the "title" field.
func (m *StoryMutation) ResetTitle() {
	m.title = nil
}

// SetDescription sets the "description" field.
func (m *StoryMutation) SetDescription(s string) {
	m.description = &s
}

// Description returns the value of the "description" field in the mutation.
func (m *StoryMutation) Description() (r string, exists bool) {
	v := m.description
	if v == nil {
		return
	}
	return *v, true
}

// OldDescription returns the old "description" field's value of the Story entity.
// If the Story object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *StoryMutation) OldDescription(ctx context.Context) (v string, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldDescription is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldDescription requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldDescription: %w", err)
	}
	return oldValue.Description, nil
}

// ResetDescription resets all changes to the "description" field.
func (m *StoryMutation) ResetDescription() {
	m.description = nil
}

// SetOwnerID sets the "owner" edge to the Board entity by id.
func (m *StoryMutation) SetOwnerID(id uuid.UUID) {
	m.owner = &id
}

// ClearOwner clears the "owner" edge to the Board entity.
func (m *StoryMutation) ClearOwner() {
	m.clearedowner = true
}

// OwnerCleared reports if the "owner" edge to the Board entity was cleared.
func (m *StoryMutation) OwnerCleared() bool {
	return m.clearedowner
}

// OwnerID returns the "owner" edge ID in the mutation.
func (m *StoryMutation) OwnerID() (id uuid.UUID, exists bool) {
	if m.owner != nil {
		return *m.owner, true
	}
	return
}

// OwnerIDs returns the "owner" edge IDs in the mutation.
// Note that IDs always returns len(IDs) <= 1 for unique edges, and you should use
// OwnerID instead. It exists only for internal usage by the builders.
func (m *StoryMutation) OwnerIDs() (ids []uuid.UUID) {
	if id := m.owner; id != nil {
		ids = append(ids, *id)
	}
	return
}

// ResetOwner resets all changes to the "owner" edge.
func (m *StoryMutation) ResetOwner() {
	m.owner = nil
	m.clearedowner = false
}

// Where appends a list predicates to the StoryMutation builder.
func (m *StoryMutation) Where(ps ...predicate.Story) {
	m.predicates = append(m.predicates, ps...)
}

// Op returns the operation name.
func (m *StoryMutation) Op() Op {
	return m.op
}

// Type returns the node type of this mutation (Story).
func (m *StoryMutation) Type() string {
	return m.typ
}

// Fields returns all fields that were changed during this mutation. Note that in
// order to get all numeric fields that were incremented/decremented, call
// AddedFields().
func (m *StoryMutation) Fields() []string {
	fields := make([]string, 0, 4)
	if m.create_time != nil {
		fields = append(fields, story.FieldCreateTime)
	}
	if m.update_time != nil {
		fields = append(fields, story.FieldUpdateTime)
	}
	if m.title != nil {
		fields = append(fields, story.FieldTitle)
	}
	if m.description != nil {
		fields = append(fields, story.FieldDescription)
	}
	return fields
}

// Field returns the value of a field with the given name. The second boolean
// return value indicates that this field was not set, or was not defined in the
// schema.
func (m *StoryMutation) Field(name string) (ent.Value, bool) {
	switch name {
	case story.FieldCreateTime:
		return m.CreateTime()
	case story.FieldUpdateTime:
		return m.UpdateTime()
	case story.FieldTitle:
		return m.Title()
	case story.FieldDescription:
		return m.Description()
	}
	return nil, false
}

// OldField returns the old value of the field from the database. An error is
// returned if the mutation operation is not UpdateOne, or the query to the
// database failed.
func (m *StoryMutation) OldField(ctx context.Context, name string) (ent.Value, error) {
	switch name {
	case story.FieldCreateTime:
		return m.OldCreateTime(ctx)
	case story.FieldUpdateTime:
		return m.OldUpdateTime(ctx)
	case story.FieldTitle:
		return m.OldTitle(ctx)
	case story.FieldDescription:
		return m.OldDescription(ctx)
	}
	return nil, fmt.Errorf("unknown Story field %s", name)
}

// SetField sets the value of a field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *StoryMutation) SetField(name string, value ent.Value) error {
	switch name {
	case story.FieldCreateTime:
		v, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetCreateTime(v)
		return nil
	case story.FieldUpdateTime:
		v, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetUpdateTime(v)
		return nil
	case story.FieldTitle:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetTitle(v)
		return nil
	case story.FieldDescription:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetDescription(v)
		return nil
	}
	return fmt.Errorf("unknown Story field %s", name)
}

// AddedFields returns all numeric fields that were incremented/decremented during
// this mutation.
func (m *StoryMutation) AddedFields() []string {
	return nil
}

// AddedField returns the numeric value that was incremented/decremented on a field
// with the given name. The second boolean return value indicates that this field
// was not set, or was not defined in the schema.
func (m *StoryMutation) AddedField(name string) (ent.Value, bool) {
	return nil, false
}

// AddField adds the value to the field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *StoryMutation) AddField(name string, value ent.Value) error {
	switch name {
	}
	return fmt.Errorf("unknown Story numeric field %s", name)
}

// ClearedFields returns all nullable fields that were cleared during this
// mutation.
func (m *StoryMutation) ClearedFields() []string {
	return nil
}

// FieldCleared returns a boolean indicating if a field with the given name was
// cleared in this mutation.
func (m *StoryMutation) FieldCleared(name string) bool {
	_, ok := m.clearedFields[name]
	return ok
}

// ClearField clears the value of the field with the given name. It returns an
// error if the field is not defined in the schema.
func (m *StoryMutation) ClearField(name string) error {
	return fmt.Errorf("unknown Story nullable field %s", name)
}

// ResetField resets all changes in the mutation for the field with the given name.
// It returns an error if the field is not defined in the schema.
func (m *StoryMutation) ResetField(name string) error {
	switch name {
	case story.FieldCreateTime:
		m.ResetCreateTime()
		return nil
	case story.FieldUpdateTime:
		m.ResetUpdateTime()
		return nil
	case story.FieldTitle:
		m.ResetTitle()
		return nil
	case story.FieldDescription:
		m.ResetDescription()
		return nil
	}
	return fmt.Errorf("unknown Story field %s", name)
}

// AddedEdges returns all edge names that were set/added in this mutation.
func (m *StoryMutation) AddedEdges() []string {
	edges := make([]string, 0, 1)
	if m.owner != nil {
		edges = append(edges, story.EdgeOwner)
	}
	return edges
}

// AddedIDs returns all IDs (to other nodes) that were added for the given edge
// name in this mutation.
func (m *StoryMutation) AddedIDs(name string) []ent.Value {
	switch name {
	case story.EdgeOwner:
		if id := m.owner; id != nil {
			return []ent.Value{*id}
		}
	}
	return nil
}

// RemovedEdges returns all edge names that were removed in this mutation.
func (m *StoryMutation) RemovedEdges() []string {
	edges := make([]string, 0, 1)
	return edges
}

// RemovedIDs returns all IDs (to other nodes) that were removed for the edge with
// the given name in this mutation.
func (m *StoryMutation) RemovedIDs(name string) []ent.Value {
	switch name {
	}
	return nil
}

// ClearedEdges returns all edge names that were cleared in this mutation.
func (m *StoryMutation) ClearedEdges() []string {
	edges := make([]string, 0, 1)
	if m.clearedowner {
		edges = append(edges, story.EdgeOwner)
	}
	return edges
}

// EdgeCleared returns a boolean which indicates if the edge with the given name
// was cleared in this mutation.
func (m *StoryMutation) EdgeCleared(name string) bool {
	switch name {
	case story.EdgeOwner:
		return m.clearedowner
	}
	return false
}

// ClearEdge clears the value of the edge with the given name. It returns an error
// if that edge is not defined in the schema.
func (m *StoryMutation) ClearEdge(name string) error {
	switch name {
	case story.EdgeOwner:
		m.ClearOwner()
		return nil
	}
	return fmt.Errorf("unknown Story unique edge %s", name)
}

// ResetEdge resets all changes to the edge with the given name in this mutation.
// It returns an error if the edge is not defined in the schema.
func (m *StoryMutation) ResetEdge(name string) error {
	switch name {
	case story.EdgeOwner:
		m.ResetOwner()
		return nil
	}
	return fmt.Errorf("unknown Story edge %s", name)
}

// ViewMutation represents an operation that mutates the View nodes in the graph.
type ViewMutation struct {
	config
	op            Op
	typ           string
	id            *uuid.UUID
	create_time   *time.Time
	update_time   *time.Time
	title         *string
	clearedFields map[string]struct{}
	done          bool
	oldValue      func(context.Context) (*View, error)
	predicates    []predicate.View
}

var _ ent.Mutation = (*ViewMutation)(nil)

// viewOption allows management of the mutation configuration using functional options.
type viewOption func(*ViewMutation)

// newViewMutation creates new mutation for the View entity.
func newViewMutation(c config, op Op, opts ...viewOption) *ViewMutation {
	m := &ViewMutation{
		config:        c,
		op:            op,
		typ:           TypeView,
		clearedFields: make(map[string]struct{}),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// withViewID sets the ID field of the mutation.
func withViewID(id uuid.UUID) viewOption {
	return func(m *ViewMutation) {
		var (
			err   error
			once  sync.Once
			value *View
		)
		m.oldValue = func(ctx context.Context) (*View, error) {
			once.Do(func() {
				if m.done {
					err = errors.New("querying old values post mutation is not allowed")
				} else {
					value, err = m.Client().View.Get(ctx, id)
				}
			})
			return value, err
		}
		m.id = &id
	}
}

// withView sets the old View of the mutation.
func withView(node *View) viewOption {
	return func(m *ViewMutation) {
		m.oldValue = func(context.Context) (*View, error) {
			return node, nil
		}
		m.id = &node.ID
	}
}

// Client returns a new `ent.Client` from the mutation. If the mutation was
// executed in a transaction (ent.Tx), a transactional client is returned.
func (m ViewMutation) Client() *Client {
	client := &Client{config: m.config}
	client.init()
	return client
}

// Tx returns an `ent.Tx` for mutations that were executed in transactions;
// it returns an error otherwise.
func (m ViewMutation) Tx() (*Tx, error) {
	if _, ok := m.driver.(*txDriver); !ok {
		return nil, errors.New("models: mutation is not running in a transaction")
	}
	tx := &Tx{config: m.config}
	tx.init()
	return tx, nil
}

// SetID sets the value of the id field. Note that this
// operation is only accepted on creation of View entities.
func (m *ViewMutation) SetID(id uuid.UUID) {
	m.id = &id
}

// ID returns the ID value in the mutation. Note that the ID is only available
// if it was provided to the builder or after it was returned from the database.
func (m *ViewMutation) ID() (id uuid.UUID, exists bool) {
	if m.id == nil {
		return
	}
	return *m.id, true
}

// IDs queries the database and returns the entity ids that match the mutation's predicate.
// That means, if the mutation is applied within a transaction with an isolation level such
// as sql.LevelSerializable, the returned ids match the ids of the rows that will be updated
// or updated by the mutation.
func (m *ViewMutation) IDs(ctx context.Context) ([]uuid.UUID, error) {
	switch {
	case m.op.Is(OpUpdateOne | OpDeleteOne):
		id, exists := m.ID()
		if exists {
			return []uuid.UUID{id}, nil
		}
		fallthrough
	case m.op.Is(OpUpdate | OpDelete):
		return m.Client().View.Query().Where(m.predicates...).IDs(ctx)
	default:
		return nil, fmt.Errorf("IDs is not allowed on %s operations", m.op)
	}
}

// SetCreateTime sets the "create_time" field.
func (m *ViewMutation) SetCreateTime(t time.Time) {
	m.create_time = &t
}

// CreateTime returns the value of the "create_time" field in the mutation.
func (m *ViewMutation) CreateTime() (r time.Time, exists bool) {
	v := m.create_time
	if v == nil {
		return
	}
	return *v, true
}

// OldCreateTime returns the old "create_time" field's value of the View entity.
// If the View object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *ViewMutation) OldCreateTime(ctx context.Context) (v time.Time, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldCreateTime is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldCreateTime requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldCreateTime: %w", err)
	}
	return oldValue.CreateTime, nil
}

// ResetCreateTime resets all changes to the "create_time" field.
func (m *ViewMutation) ResetCreateTime() {
	m.create_time = nil
}

// SetUpdateTime sets the "update_time" field.
func (m *ViewMutation) SetUpdateTime(t time.Time) {
	m.update_time = &t
}

// UpdateTime returns the value of the "update_time" field in the mutation.
func (m *ViewMutation) UpdateTime() (r time.Time, exists bool) {
	v := m.update_time
	if v == nil {
		return
	}
	return *v, true
}

// OldUpdateTime returns the old "update_time" field's value of the View entity.
// If the View object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *ViewMutation) OldUpdateTime(ctx context.Context) (v time.Time, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldUpdateTime is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldUpdateTime requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldUpdateTime: %w", err)
	}
	return oldValue.UpdateTime, nil
}

// ResetUpdateTime resets all changes to the "update_time" field.
func (m *ViewMutation) ResetUpdateTime() {
	m.update_time = nil
}

// SetTitle sets the "title" field.
func (m *ViewMutation) SetTitle(s string) {
	m.title = &s
}

// Title returns the value of the "title" field in the mutation.
func (m *ViewMutation) Title() (r string, exists bool) {
	v := m.title
	if v == nil {
		return
	}
	return *v, true
}

// OldTitle returns the old "title" field's value of the View entity.
// If the View object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *ViewMutation) OldTitle(ctx context.Context) (v string, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldTitle is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldTitle requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldTitle: %w", err)
	}
	return oldValue.Title, nil
}

// ResetTitle resets all changes to the "title" field.
func (m *ViewMutation) ResetTitle() {
	m.title = nil
}

// Where appends a list predicates to the ViewMutation builder.
func (m *ViewMutation) Where(ps ...predicate.View) {
	m.predicates = append(m.predicates, ps...)
}

// Op returns the operation name.
func (m *ViewMutation) Op() Op {
	return m.op
}

// Type returns the node type of this mutation (View).
func (m *ViewMutation) Type() string {
	return m.typ
}

// Fields returns all fields that were changed during this mutation. Note that in
// order to get all numeric fields that were incremented/decremented, call
// AddedFields().
func (m *ViewMutation) Fields() []string {
	fields := make([]string, 0, 3)
	if m.create_time != nil {
		fields = append(fields, view.FieldCreateTime)
	}
	if m.update_time != nil {
		fields = append(fields, view.FieldUpdateTime)
	}
	if m.title != nil {
		fields = append(fields, view.FieldTitle)
	}
	return fields
}

// Field returns the value of a field with the given name. The second boolean
// return value indicates that this field was not set, or was not defined in the
// schema.
func (m *ViewMutation) Field(name string) (ent.Value, bool) {
	switch name {
	case view.FieldCreateTime:
		return m.CreateTime()
	case view.FieldUpdateTime:
		return m.UpdateTime()
	case view.FieldTitle:
		return m.Title()
	}
	return nil, false
}

// OldField returns the old value of the field from the database. An error is
// returned if the mutation operation is not UpdateOne, or the query to the
// database failed.
func (m *ViewMutation) OldField(ctx context.Context, name string) (ent.Value, error) {
	switch name {
	case view.FieldCreateTime:
		return m.OldCreateTime(ctx)
	case view.FieldUpdateTime:
		return m.OldUpdateTime(ctx)
	case view.FieldTitle:
		return m.OldTitle(ctx)
	}
	return nil, fmt.Errorf("unknown View field %s", name)
}

// SetField sets the value of a field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *ViewMutation) SetField(name string, value ent.Value) error {
	switch name {
	case view.FieldCreateTime:
		v, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetCreateTime(v)
		return nil
	case view.FieldUpdateTime:
		v, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetUpdateTime(v)
		return nil
	case view.FieldTitle:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetTitle(v)
		return nil
	}
	return fmt.Errorf("unknown View field %s", name)
}

// AddedFields returns all numeric fields that were incremented/decremented during
// this mutation.
func (m *ViewMutation) AddedFields() []string {
	return nil
}

// AddedField returns the numeric value that was incremented/decremented on a field
// with the given name. The second boolean return value indicates that this field
// was not set, or was not defined in the schema.
func (m *ViewMutation) AddedField(name string) (ent.Value, bool) {
	return nil, false
}

// AddField adds the value to the field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *ViewMutation) AddField(name string, value ent.Value) error {
	switch name {
	}
	return fmt.Errorf("unknown View numeric field %s", name)
}

// ClearedFields returns all nullable fields that were cleared during this
// mutation.
func (m *ViewMutation) ClearedFields() []string {
	return nil
}

// FieldCleared returns a boolean indicating if a field with the given name was
// cleared in this mutation.
func (m *ViewMutation) FieldCleared(name string) bool {
	_, ok := m.clearedFields[name]
	return ok
}

// ClearField clears the value of the field with the given name. It returns an
// error if the field is not defined in the schema.
func (m *ViewMutation) ClearField(name string) error {
	return fmt.Errorf("unknown View nullable field %s", name)
}

// ResetField resets all changes in the mutation for the field with the given name.
// It returns an error if the field is not defined in the schema.
func (m *ViewMutation) ResetField(name string) error {
	switch name {
	case view.FieldCreateTime:
		m.ResetCreateTime()
		return nil
	case view.FieldUpdateTime:
		m.ResetUpdateTime()
		return nil
	case view.FieldTitle:
		m.ResetTitle()
		return nil
	}
	return fmt.Errorf("unknown View field %s", name)
}

// AddedEdges returns all edge names that were set/added in this mutation.
func (m *ViewMutation) AddedEdges() []string {
	edges := make([]string, 0, 0)
	return edges
}

// AddedIDs returns all IDs (to other nodes) that were added for the given edge
// name in this mutation.
func (m *ViewMutation) AddedIDs(name string) []ent.Value {
	return nil
}

// RemovedEdges returns all edge names that were removed in this mutation.
func (m *ViewMutation) RemovedEdges() []string {
	edges := make([]string, 0, 0)
	return edges
}

// RemovedIDs returns all IDs (to other nodes) that were removed for the edge with
// the given name in this mutation.
func (m *ViewMutation) RemovedIDs(name string) []ent.Value {
	return nil
}

// ClearedEdges returns all edge names that were cleared in this mutation.
func (m *ViewMutation) ClearedEdges() []string {
	edges := make([]string, 0, 0)
	return edges
}

// EdgeCleared returns a boolean which indicates if the edge with the given name
// was cleared in this mutation.
func (m *ViewMutation) EdgeCleared(name string) bool {
	return false
}

// ClearEdge clears the value of the edge with the given name. It returns an error
// if that edge is not defined in the schema.
func (m *ViewMutation) ClearEdge(name string) error {
	return fmt.Errorf("unknown View unique edge %s", name)
}

// ResetEdge resets all changes to the edge with the given name in this mutation.
// It returns an error if the edge is not defined in the schema.
func (m *ViewMutation) ResetEdge(name string) error {
	return fmt.Errorf("unknown View edge %s", name)
}
