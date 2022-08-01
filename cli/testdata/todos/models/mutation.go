// Code generated (@generated) by entc, DO NOT EDIT.

package models

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/adnaan/fir/cli/testdata/todos/models/board"
	"github.com/adnaan/fir/cli/testdata/todos/models/predicate"
	"github.com/adnaan/fir/cli/testdata/todos/models/todo"
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
	TypeBoard = "Board"
	TypeTodo  = "Todo"
)

// BoardMutation represents an operation that mutates the Board nodes in the graph.
type BoardMutation struct {
	config
	op            Op
	typ           string
	id            *uuid.UUID
	create_time   *time.Time
	update_time   *time.Time
	title         *string
	description   *string
	clearedFields map[string]struct{}
	todos         map[uuid.UUID]struct{}
	removedtodos  map[uuid.UUID]struct{}
	clearedtodos  bool
	done          bool
	oldValue      func(context.Context) (*Board, error)
	predicates    []predicate.Board
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

// AddTodoIDs adds the "todos" edge to the Todo entity by ids.
func (m *BoardMutation) AddTodoIDs(ids ...uuid.UUID) {
	if m.todos == nil {
		m.todos = make(map[uuid.UUID]struct{})
	}
	for i := range ids {
		m.todos[ids[i]] = struct{}{}
	}
}

// ClearTodos clears the "todos" edge to the Todo entity.
func (m *BoardMutation) ClearTodos() {
	m.clearedtodos = true
}

// TodosCleared reports if the "todos" edge to the Todo entity was cleared.
func (m *BoardMutation) TodosCleared() bool {
	return m.clearedtodos
}

// RemoveTodoIDs removes the "todos" edge to the Todo entity by IDs.
func (m *BoardMutation) RemoveTodoIDs(ids ...uuid.UUID) {
	if m.removedtodos == nil {
		m.removedtodos = make(map[uuid.UUID]struct{})
	}
	for i := range ids {
		delete(m.todos, ids[i])
		m.removedtodos[ids[i]] = struct{}{}
	}
}

// RemovedTodos returns the removed IDs of the "todos" edge to the Todo entity.
func (m *BoardMutation) RemovedTodosIDs() (ids []uuid.UUID) {
	for id := range m.removedtodos {
		ids = append(ids, id)
	}
	return
}

// TodosIDs returns the "todos" edge IDs in the mutation.
func (m *BoardMutation) TodosIDs() (ids []uuid.UUID) {
	for id := range m.todos {
		ids = append(ids, id)
	}
	return
}

// ResetTodos resets all changes to the "todos" edge.
func (m *BoardMutation) ResetTodos() {
	m.todos = nil
	m.clearedtodos = false
	m.removedtodos = nil
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
	if m.todos != nil {
		edges = append(edges, board.EdgeTodos)
	}
	return edges
}

// AddedIDs returns all IDs (to other nodes) that were added for the given edge
// name in this mutation.
func (m *BoardMutation) AddedIDs(name string) []ent.Value {
	switch name {
	case board.EdgeTodos:
		ids := make([]ent.Value, 0, len(m.todos))
		for id := range m.todos {
			ids = append(ids, id)
		}
		return ids
	}
	return nil
}

// RemovedEdges returns all edge names that were removed in this mutation.
func (m *BoardMutation) RemovedEdges() []string {
	edges := make([]string, 0, 1)
	if m.removedtodos != nil {
		edges = append(edges, board.EdgeTodos)
	}
	return edges
}

// RemovedIDs returns all IDs (to other nodes) that were removed for the edge with
// the given name in this mutation.
func (m *BoardMutation) RemovedIDs(name string) []ent.Value {
	switch name {
	case board.EdgeTodos:
		ids := make([]ent.Value, 0, len(m.removedtodos))
		for id := range m.removedtodos {
			ids = append(ids, id)
		}
		return ids
	}
	return nil
}

// ClearedEdges returns all edge names that were cleared in this mutation.
func (m *BoardMutation) ClearedEdges() []string {
	edges := make([]string, 0, 1)
	if m.clearedtodos {
		edges = append(edges, board.EdgeTodos)
	}
	return edges
}

// EdgeCleared returns a boolean which indicates if the edge with the given name
// was cleared in this mutation.
func (m *BoardMutation) EdgeCleared(name string) bool {
	switch name {
	case board.EdgeTodos:
		return m.clearedtodos
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
	case board.EdgeTodos:
		m.ResetTodos()
		return nil
	}
	return fmt.Errorf("unknown Board edge %s", name)
}

// TodoMutation represents an operation that mutates the Todo nodes in the graph.
type TodoMutation struct {
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
	oldValue      func(context.Context) (*Todo, error)
	predicates    []predicate.Todo
}

var _ ent.Mutation = (*TodoMutation)(nil)

// todoOption allows management of the mutation configuration using functional options.
type todoOption func(*TodoMutation)

// newTodoMutation creates new mutation for the Todo entity.
func newTodoMutation(c config, op Op, opts ...todoOption) *TodoMutation {
	m := &TodoMutation{
		config:        c,
		op:            op,
		typ:           TypeTodo,
		clearedFields: make(map[string]struct{}),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// withTodoID sets the ID field of the mutation.
func withTodoID(id uuid.UUID) todoOption {
	return func(m *TodoMutation) {
		var (
			err   error
			once  sync.Once
			value *Todo
		)
		m.oldValue = func(ctx context.Context) (*Todo, error) {
			once.Do(func() {
				if m.done {
					err = errors.New("querying old values post mutation is not allowed")
				} else {
					value, err = m.Client().Todo.Get(ctx, id)
				}
			})
			return value, err
		}
		m.id = &id
	}
}

// withTodo sets the old Todo of the mutation.
func withTodo(node *Todo) todoOption {
	return func(m *TodoMutation) {
		m.oldValue = func(context.Context) (*Todo, error) {
			return node, nil
		}
		m.id = &node.ID
	}
}

// Client returns a new `ent.Client` from the mutation. If the mutation was
// executed in a transaction (ent.Tx), a transactional client is returned.
func (m TodoMutation) Client() *Client {
	client := &Client{config: m.config}
	client.init()
	return client
}

// Tx returns an `ent.Tx` for mutations that were executed in transactions;
// it returns an error otherwise.
func (m TodoMutation) Tx() (*Tx, error) {
	if _, ok := m.driver.(*txDriver); !ok {
		return nil, errors.New("models: mutation is not running in a transaction")
	}
	tx := &Tx{config: m.config}
	tx.init()
	return tx, nil
}

// SetID sets the value of the id field. Note that this
// operation is only accepted on creation of Todo entities.
func (m *TodoMutation) SetID(id uuid.UUID) {
	m.id = &id
}

// ID returns the ID value in the mutation. Note that the ID is only available
// if it was provided to the builder or after it was returned from the database.
func (m *TodoMutation) ID() (id uuid.UUID, exists bool) {
	if m.id == nil {
		return
	}
	return *m.id, true
}

// IDs queries the database and returns the entity ids that match the mutation's predicate.
// That means, if the mutation is applied within a transaction with an isolation level such
// as sql.LevelSerializable, the returned ids match the ids of the rows that will be updated
// or updated by the mutation.
func (m *TodoMutation) IDs(ctx context.Context) ([]uuid.UUID, error) {
	switch {
	case m.op.Is(OpUpdateOne | OpDeleteOne):
		id, exists := m.ID()
		if exists {
			return []uuid.UUID{id}, nil
		}
		fallthrough
	case m.op.Is(OpUpdate | OpDelete):
		return m.Client().Todo.Query().Where(m.predicates...).IDs(ctx)
	default:
		return nil, fmt.Errorf("IDs is not allowed on %s operations", m.op)
	}
}

// SetCreateTime sets the "create_time" field.
func (m *TodoMutation) SetCreateTime(t time.Time) {
	m.create_time = &t
}

// CreateTime returns the value of the "create_time" field in the mutation.
func (m *TodoMutation) CreateTime() (r time.Time, exists bool) {
	v := m.create_time
	if v == nil {
		return
	}
	return *v, true
}

// OldCreateTime returns the old "create_time" field's value of the Todo entity.
// If the Todo object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *TodoMutation) OldCreateTime(ctx context.Context) (v time.Time, err error) {
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
func (m *TodoMutation) ResetCreateTime() {
	m.create_time = nil
}

// SetUpdateTime sets the "update_time" field.
func (m *TodoMutation) SetUpdateTime(t time.Time) {
	m.update_time = &t
}

// UpdateTime returns the value of the "update_time" field in the mutation.
func (m *TodoMutation) UpdateTime() (r time.Time, exists bool) {
	v := m.update_time
	if v == nil {
		return
	}
	return *v, true
}

// OldUpdateTime returns the old "update_time" field's value of the Todo entity.
// If the Todo object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *TodoMutation) OldUpdateTime(ctx context.Context) (v time.Time, err error) {
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
func (m *TodoMutation) ResetUpdateTime() {
	m.update_time = nil
}

// SetTitle sets the "title" field.
func (m *TodoMutation) SetTitle(s string) {
	m.title = &s
}

// Title returns the value of the "title" field in the mutation.
func (m *TodoMutation) Title() (r string, exists bool) {
	v := m.title
	if v == nil {
		return
	}
	return *v, true
}

// OldTitle returns the old "title" field's value of the Todo entity.
// If the Todo object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *TodoMutation) OldTitle(ctx context.Context) (v string, err error) {
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
func (m *TodoMutation) ResetTitle() {
	m.title = nil
}

// SetDescription sets the "description" field.
func (m *TodoMutation) SetDescription(s string) {
	m.description = &s
}

// Description returns the value of the "description" field in the mutation.
func (m *TodoMutation) Description() (r string, exists bool) {
	v := m.description
	if v == nil {
		return
	}
	return *v, true
}

// OldDescription returns the old "description" field's value of the Todo entity.
// If the Todo object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *TodoMutation) OldDescription(ctx context.Context) (v string, err error) {
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
func (m *TodoMutation) ResetDescription() {
	m.description = nil
}

// SetOwnerID sets the "owner" edge to the Board entity by id.
func (m *TodoMutation) SetOwnerID(id uuid.UUID) {
	m.owner = &id
}

// ClearOwner clears the "owner" edge to the Board entity.
func (m *TodoMutation) ClearOwner() {
	m.clearedowner = true
}

// OwnerCleared reports if the "owner" edge to the Board entity was cleared.
func (m *TodoMutation) OwnerCleared() bool {
	return m.clearedowner
}

// OwnerID returns the "owner" edge ID in the mutation.
func (m *TodoMutation) OwnerID() (id uuid.UUID, exists bool) {
	if m.owner != nil {
		return *m.owner, true
	}
	return
}

// OwnerIDs returns the "owner" edge IDs in the mutation.
// Note that IDs always returns len(IDs) <= 1 for unique edges, and you should use
// OwnerID instead. It exists only for internal usage by the builders.
func (m *TodoMutation) OwnerIDs() (ids []uuid.UUID) {
	if id := m.owner; id != nil {
		ids = append(ids, *id)
	}
	return
}

// ResetOwner resets all changes to the "owner" edge.
func (m *TodoMutation) ResetOwner() {
	m.owner = nil
	m.clearedowner = false
}

// Where appends a list predicates to the TodoMutation builder.
func (m *TodoMutation) Where(ps ...predicate.Todo) {
	m.predicates = append(m.predicates, ps...)
}

// Op returns the operation name.
func (m *TodoMutation) Op() Op {
	return m.op
}

// Type returns the node type of this mutation (Todo).
func (m *TodoMutation) Type() string {
	return m.typ
}

// Fields returns all fields that were changed during this mutation. Note that in
// order to get all numeric fields that were incremented/decremented, call
// AddedFields().
func (m *TodoMutation) Fields() []string {
	fields := make([]string, 0, 4)
	if m.create_time != nil {
		fields = append(fields, todo.FieldCreateTime)
	}
	if m.update_time != nil {
		fields = append(fields, todo.FieldUpdateTime)
	}
	if m.title != nil {
		fields = append(fields, todo.FieldTitle)
	}
	if m.description != nil {
		fields = append(fields, todo.FieldDescription)
	}
	return fields
}

// Field returns the value of a field with the given name. The second boolean
// return value indicates that this field was not set, or was not defined in the
// schema.
func (m *TodoMutation) Field(name string) (ent.Value, bool) {
	switch name {
	case todo.FieldCreateTime:
		return m.CreateTime()
	case todo.FieldUpdateTime:
		return m.UpdateTime()
	case todo.FieldTitle:
		return m.Title()
	case todo.FieldDescription:
		return m.Description()
	}
	return nil, false
}

// OldField returns the old value of the field from the database. An error is
// returned if the mutation operation is not UpdateOne, or the query to the
// database failed.
func (m *TodoMutation) OldField(ctx context.Context, name string) (ent.Value, error) {
	switch name {
	case todo.FieldCreateTime:
		return m.OldCreateTime(ctx)
	case todo.FieldUpdateTime:
		return m.OldUpdateTime(ctx)
	case todo.FieldTitle:
		return m.OldTitle(ctx)
	case todo.FieldDescription:
		return m.OldDescription(ctx)
	}
	return nil, fmt.Errorf("unknown Todo field %s", name)
}

// SetField sets the value of a field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *TodoMutation) SetField(name string, value ent.Value) error {
	switch name {
	case todo.FieldCreateTime:
		v, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetCreateTime(v)
		return nil
	case todo.FieldUpdateTime:
		v, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetUpdateTime(v)
		return nil
	case todo.FieldTitle:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetTitle(v)
		return nil
	case todo.FieldDescription:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetDescription(v)
		return nil
	}
	return fmt.Errorf("unknown Todo field %s", name)
}

// AddedFields returns all numeric fields that were incremented/decremented during
// this mutation.
func (m *TodoMutation) AddedFields() []string {
	return nil
}

// AddedField returns the numeric value that was incremented/decremented on a field
// with the given name. The second boolean return value indicates that this field
// was not set, or was not defined in the schema.
func (m *TodoMutation) AddedField(name string) (ent.Value, bool) {
	return nil, false
}

// AddField adds the value to the field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *TodoMutation) AddField(name string, value ent.Value) error {
	switch name {
	}
	return fmt.Errorf("unknown Todo numeric field %s", name)
}

// ClearedFields returns all nullable fields that were cleared during this
// mutation.
func (m *TodoMutation) ClearedFields() []string {
	return nil
}

// FieldCleared returns a boolean indicating if a field with the given name was
// cleared in this mutation.
func (m *TodoMutation) FieldCleared(name string) bool {
	_, ok := m.clearedFields[name]
	return ok
}

// ClearField clears the value of the field with the given name. It returns an
// error if the field is not defined in the schema.
func (m *TodoMutation) ClearField(name string) error {
	return fmt.Errorf("unknown Todo nullable field %s", name)
}

// ResetField resets all changes in the mutation for the field with the given name.
// It returns an error if the field is not defined in the schema.
func (m *TodoMutation) ResetField(name string) error {
	switch name {
	case todo.FieldCreateTime:
		m.ResetCreateTime()
		return nil
	case todo.FieldUpdateTime:
		m.ResetUpdateTime()
		return nil
	case todo.FieldTitle:
		m.ResetTitle()
		return nil
	case todo.FieldDescription:
		m.ResetDescription()
		return nil
	}
	return fmt.Errorf("unknown Todo field %s", name)
}

// AddedEdges returns all edge names that were set/added in this mutation.
func (m *TodoMutation) AddedEdges() []string {
	edges := make([]string, 0, 1)
	if m.owner != nil {
		edges = append(edges, todo.EdgeOwner)
	}
	return edges
}

// AddedIDs returns all IDs (to other nodes) that were added for the given edge
// name in this mutation.
func (m *TodoMutation) AddedIDs(name string) []ent.Value {
	switch name {
	case todo.EdgeOwner:
		if id := m.owner; id != nil {
			return []ent.Value{*id}
		}
	}
	return nil
}

// RemovedEdges returns all edge names that were removed in this mutation.
func (m *TodoMutation) RemovedEdges() []string {
	edges := make([]string, 0, 1)
	return edges
}

// RemovedIDs returns all IDs (to other nodes) that were removed for the edge with
// the given name in this mutation.
func (m *TodoMutation) RemovedIDs(name string) []ent.Value {
	switch name {
	}
	return nil
}

// ClearedEdges returns all edge names that were cleared in this mutation.
func (m *TodoMutation) ClearedEdges() []string {
	edges := make([]string, 0, 1)
	if m.clearedowner {
		edges = append(edges, todo.EdgeOwner)
	}
	return edges
}

// EdgeCleared returns a boolean which indicates if the edge with the given name
// was cleared in this mutation.
func (m *TodoMutation) EdgeCleared(name string) bool {
	switch name {
	case todo.EdgeOwner:
		return m.clearedowner
	}
	return false
}

// ClearEdge clears the value of the edge with the given name. It returns an error
// if that edge is not defined in the schema.
func (m *TodoMutation) ClearEdge(name string) error {
	switch name {
	case todo.EdgeOwner:
		m.ClearOwner()
		return nil
	}
	return fmt.Errorf("unknown Todo unique edge %s", name)
}

// ResetEdge resets all changes to the edge with the given name in this mutation.
// It returns an error if the edge is not defined in the schema.
func (m *TodoMutation) ResetEdge(name string) error {
	switch name {
	case todo.EdgeOwner:
		m.ResetOwner()
		return nil
	}
	return fmt.Errorf("unknown Todo edge %s", name)
}
