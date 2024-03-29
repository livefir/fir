// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/livefir/fir/examples/fira/ent/issue"
	"github.com/livefir/fir/examples/fira/ent/predicate"
	"github.com/livefir/fir/examples/fira/ent/project"

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
	TypeIssue   = "Issue"
	TypeProject = "Project"
)

// IssueMutation represents an operation that mutates the Issue nodes in the graph.
type IssueMutation struct {
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
	oldValue      func(context.Context) (*Issue, error)
	predicates    []predicate.Issue
}

var _ ent.Mutation = (*IssueMutation)(nil)

// issueOption allows management of the mutation configuration using functional options.
type issueOption func(*IssueMutation)

// newIssueMutation creates new mutation for the Issue entity.
func newIssueMutation(c config, op Op, opts ...issueOption) *IssueMutation {
	m := &IssueMutation{
		config:        c,
		op:            op,
		typ:           TypeIssue,
		clearedFields: make(map[string]struct{}),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// withIssueID sets the ID field of the mutation.
func withIssueID(id uuid.UUID) issueOption {
	return func(m *IssueMutation) {
		var (
			err   error
			once  sync.Once
			value *Issue
		)
		m.oldValue = func(ctx context.Context) (*Issue, error) {
			once.Do(func() {
				if m.done {
					err = errors.New("querying old values post mutation is not allowed")
				} else {
					value, err = m.Client().Issue.Get(ctx, id)
				}
			})
			return value, err
		}
		m.id = &id
	}
}

// withIssue sets the old Issue of the mutation.
func withIssue(node *Issue) issueOption {
	return func(m *IssueMutation) {
		m.oldValue = func(context.Context) (*Issue, error) {
			return node, nil
		}
		m.id = &node.ID
	}
}

// Client returns a new `ent.Client` from the mutation. If the mutation was
// executed in a transaction (ent.Tx), a transactional client is returned.
func (m IssueMutation) Client() *Client {
	client := &Client{config: m.config}
	client.init()
	return client
}

// Tx returns an `ent.Tx` for mutations that were executed in transactions;
// it returns an error otherwise.
func (m IssueMutation) Tx() (*Tx, error) {
	if _, ok := m.driver.(*txDriver); !ok {
		return nil, errors.New("ent: mutation is not running in a transaction")
	}
	tx := &Tx{config: m.config}
	tx.init()
	return tx, nil
}

// SetID sets the value of the id field. Note that this
// operation is only accepted on creation of Issue entities.
func (m *IssueMutation) SetID(id uuid.UUID) {
	m.id = &id
}

// ID returns the ID value in the mutation. Note that the ID is only available
// if it was provided to the builder or after it was returned from the database.
func (m *IssueMutation) ID() (id uuid.UUID, exists bool) {
	if m.id == nil {
		return
	}
	return *m.id, true
}

// IDs queries the database and returns the entity ids that match the mutation's predicate.
// That means, if the mutation is applied within a transaction with an isolation level such
// as sql.LevelSerializable, the returned ids match the ids of the rows that will be updated
// or updated by the mutation.
func (m *IssueMutation) IDs(ctx context.Context) ([]uuid.UUID, error) {
	switch {
	case m.op.Is(OpUpdateOne | OpDeleteOne):
		id, exists := m.ID()
		if exists {
			return []uuid.UUID{id}, nil
		}
		fallthrough
	case m.op.Is(OpUpdate | OpDelete):
		return m.Client().Issue.Query().Where(m.predicates...).IDs(ctx)
	default:
		return nil, fmt.Errorf("IDs is not allowed on %s operations", m.op)
	}
}

// SetCreateTime sets the "create_time" field.
func (m *IssueMutation) SetCreateTime(t time.Time) {
	m.create_time = &t
}

// CreateTime returns the value of the "create_time" field in the mutation.
func (m *IssueMutation) CreateTime() (r time.Time, exists bool) {
	v := m.create_time
	if v == nil {
		return
	}
	return *v, true
}

// OldCreateTime returns the old "create_time" field's value of the Issue entity.
// If the Issue object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *IssueMutation) OldCreateTime(ctx context.Context) (v time.Time, err error) {
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
func (m *IssueMutation) ResetCreateTime() {
	m.create_time = nil
}

// SetUpdateTime sets the "update_time" field.
func (m *IssueMutation) SetUpdateTime(t time.Time) {
	m.update_time = &t
}

// UpdateTime returns the value of the "update_time" field in the mutation.
func (m *IssueMutation) UpdateTime() (r time.Time, exists bool) {
	v := m.update_time
	if v == nil {
		return
	}
	return *v, true
}

// OldUpdateTime returns the old "update_time" field's value of the Issue entity.
// If the Issue object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *IssueMutation) OldUpdateTime(ctx context.Context) (v time.Time, err error) {
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
func (m *IssueMutation) ResetUpdateTime() {
	m.update_time = nil
}

// SetTitle sets the "title" field.
func (m *IssueMutation) SetTitle(s string) {
	m.title = &s
}

// Title returns the value of the "title" field in the mutation.
func (m *IssueMutation) Title() (r string, exists bool) {
	v := m.title
	if v == nil {
		return
	}
	return *v, true
}

// OldTitle returns the old "title" field's value of the Issue entity.
// If the Issue object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *IssueMutation) OldTitle(ctx context.Context) (v string, err error) {
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
func (m *IssueMutation) ResetTitle() {
	m.title = nil
}

// SetDescription sets the "description" field.
func (m *IssueMutation) SetDescription(s string) {
	m.description = &s
}

// Description returns the value of the "description" field in the mutation.
func (m *IssueMutation) Description() (r string, exists bool) {
	v := m.description
	if v == nil {
		return
	}
	return *v, true
}

// OldDescription returns the old "description" field's value of the Issue entity.
// If the Issue object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *IssueMutation) OldDescription(ctx context.Context) (v string, err error) {
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
func (m *IssueMutation) ResetDescription() {
	m.description = nil
}

// SetOwnerID sets the "owner" edge to the Project entity by id.
func (m *IssueMutation) SetOwnerID(id uuid.UUID) {
	m.owner = &id
}

// ClearOwner clears the "owner" edge to the Project entity.
func (m *IssueMutation) ClearOwner() {
	m.clearedowner = true
}

// OwnerCleared reports if the "owner" edge to the Project entity was cleared.
func (m *IssueMutation) OwnerCleared() bool {
	return m.clearedowner
}

// OwnerID returns the "owner" edge ID in the mutation.
func (m *IssueMutation) OwnerID() (id uuid.UUID, exists bool) {
	if m.owner != nil {
		return *m.owner, true
	}
	return
}

// OwnerIDs returns the "owner" edge IDs in the mutation.
// Note that IDs always returns len(IDs) <= 1 for unique edges, and you should use
// OwnerID instead. It exists only for internal usage by the builders.
func (m *IssueMutation) OwnerIDs() (ids []uuid.UUID) {
	if id := m.owner; id != nil {
		ids = append(ids, *id)
	}
	return
}

// ResetOwner resets all changes to the "owner" edge.
func (m *IssueMutation) ResetOwner() {
	m.owner = nil
	m.clearedowner = false
}

// Where appends a list predicates to the IssueMutation builder.
func (m *IssueMutation) Where(ps ...predicate.Issue) {
	m.predicates = append(m.predicates, ps...)
}

// Op returns the operation name.
func (m *IssueMutation) Op() Op {
	return m.op
}

// Type returns the node type of this mutation (Issue).
func (m *IssueMutation) Type() string {
	return m.typ
}

// Fields returns all fields that were changed during this mutation. Note that in
// order to get all numeric fields that were incremented/decremented, call
// AddedFields().
func (m *IssueMutation) Fields() []string {
	fields := make([]string, 0, 4)
	if m.create_time != nil {
		fields = append(fields, issue.FieldCreateTime)
	}
	if m.update_time != nil {
		fields = append(fields, issue.FieldUpdateTime)
	}
	if m.title != nil {
		fields = append(fields, issue.FieldTitle)
	}
	if m.description != nil {
		fields = append(fields, issue.FieldDescription)
	}
	return fields
}

// Field returns the value of a field with the given name. The second boolean
// return value indicates that this field was not set, or was not defined in the
// schema.
func (m *IssueMutation) Field(name string) (ent.Value, bool) {
	switch name {
	case issue.FieldCreateTime:
		return m.CreateTime()
	case issue.FieldUpdateTime:
		return m.UpdateTime()
	case issue.FieldTitle:
		return m.Title()
	case issue.FieldDescription:
		return m.Description()
	}
	return nil, false
}

// OldField returns the old value of the field from the database. An error is
// returned if the mutation operation is not UpdateOne, or the query to the
// database failed.
func (m *IssueMutation) OldField(ctx context.Context, name string) (ent.Value, error) {
	switch name {
	case issue.FieldCreateTime:
		return m.OldCreateTime(ctx)
	case issue.FieldUpdateTime:
		return m.OldUpdateTime(ctx)
	case issue.FieldTitle:
		return m.OldTitle(ctx)
	case issue.FieldDescription:
		return m.OldDescription(ctx)
	}
	return nil, fmt.Errorf("unknown Issue field %s", name)
}

// SetField sets the value of a field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *IssueMutation) SetField(name string, value ent.Value) error {
	switch name {
	case issue.FieldCreateTime:
		v, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetCreateTime(v)
		return nil
	case issue.FieldUpdateTime:
		v, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetUpdateTime(v)
		return nil
	case issue.FieldTitle:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetTitle(v)
		return nil
	case issue.FieldDescription:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetDescription(v)
		return nil
	}
	return fmt.Errorf("unknown Issue field %s", name)
}

// AddedFields returns all numeric fields that were incremented/decremented during
// this mutation.
func (m *IssueMutation) AddedFields() []string {
	return nil
}

// AddedField returns the numeric value that was incremented/decremented on a field
// with the given name. The second boolean return value indicates that this field
// was not set, or was not defined in the schema.
func (m *IssueMutation) AddedField(name string) (ent.Value, bool) {
	return nil, false
}

// AddField adds the value to the field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *IssueMutation) AddField(name string, value ent.Value) error {
	switch name {
	}
	return fmt.Errorf("unknown Issue numeric field %s", name)
}

// ClearedFields returns all nullable fields that were cleared during this
// mutation.
func (m *IssueMutation) ClearedFields() []string {
	return nil
}

// FieldCleared returns a boolean indicating if a field with the given name was
// cleared in this mutation.
func (m *IssueMutation) FieldCleared(name string) bool {
	_, ok := m.clearedFields[name]
	return ok
}

// ClearField clears the value of the field with the given name. It returns an
// error if the field is not defined in the schema.
func (m *IssueMutation) ClearField(name string) error {
	return fmt.Errorf("unknown Issue nullable field %s", name)
}

// ResetField resets all changes in the mutation for the field with the given name.
// It returns an error if the field is not defined in the schema.
func (m *IssueMutation) ResetField(name string) error {
	switch name {
	case issue.FieldCreateTime:
		m.ResetCreateTime()
		return nil
	case issue.FieldUpdateTime:
		m.ResetUpdateTime()
		return nil
	case issue.FieldTitle:
		m.ResetTitle()
		return nil
	case issue.FieldDescription:
		m.ResetDescription()
		return nil
	}
	return fmt.Errorf("unknown Issue field %s", name)
}

// AddedEdges returns all edge names that were set/added in this mutation.
func (m *IssueMutation) AddedEdges() []string {
	edges := make([]string, 0, 1)
	if m.owner != nil {
		edges = append(edges, issue.EdgeOwner)
	}
	return edges
}

// AddedIDs returns all IDs (to other nodes) that were added for the given edge
// name in this mutation.
func (m *IssueMutation) AddedIDs(name string) []ent.Value {
	switch name {
	case issue.EdgeOwner:
		if id := m.owner; id != nil {
			return []ent.Value{*id}
		}
	}
	return nil
}

// RemovedEdges returns all edge names that were removed in this mutation.
func (m *IssueMutation) RemovedEdges() []string {
	edges := make([]string, 0, 1)
	return edges
}

// RemovedIDs returns all IDs (to other nodes) that were removed for the edge with
// the given name in this mutation.
func (m *IssueMutation) RemovedIDs(name string) []ent.Value {
	return nil
}

// ClearedEdges returns all edge names that were cleared in this mutation.
func (m *IssueMutation) ClearedEdges() []string {
	edges := make([]string, 0, 1)
	if m.clearedowner {
		edges = append(edges, issue.EdgeOwner)
	}
	return edges
}

// EdgeCleared returns a boolean which indicates if the edge with the given name
// was cleared in this mutation.
func (m *IssueMutation) EdgeCleared(name string) bool {
	switch name {
	case issue.EdgeOwner:
		return m.clearedowner
	}
	return false
}

// ClearEdge clears the value of the edge with the given name. It returns an error
// if that edge is not defined in the schema.
func (m *IssueMutation) ClearEdge(name string) error {
	switch name {
	case issue.EdgeOwner:
		m.ClearOwner()
		return nil
	}
	return fmt.Errorf("unknown Issue unique edge %s", name)
}

// ResetEdge resets all changes to the edge with the given name in this mutation.
// It returns an error if the edge is not defined in the schema.
func (m *IssueMutation) ResetEdge(name string) error {
	switch name {
	case issue.EdgeOwner:
		m.ResetOwner()
		return nil
	}
	return fmt.Errorf("unknown Issue edge %s", name)
}

// ProjectMutation represents an operation that mutates the Project nodes in the graph.
type ProjectMutation struct {
	config
	op            Op
	typ           string
	id            *uuid.UUID
	create_time   *time.Time
	update_time   *time.Time
	title         *string
	description   *string
	clearedFields map[string]struct{}
	issues        map[uuid.UUID]struct{}
	removedissues map[uuid.UUID]struct{}
	clearedissues bool
	done          bool
	oldValue      func(context.Context) (*Project, error)
	predicates    []predicate.Project
}

var _ ent.Mutation = (*ProjectMutation)(nil)

// projectOption allows management of the mutation configuration using functional options.
type projectOption func(*ProjectMutation)

// newProjectMutation creates new mutation for the Project entity.
func newProjectMutation(c config, op Op, opts ...projectOption) *ProjectMutation {
	m := &ProjectMutation{
		config:        c,
		op:            op,
		typ:           TypeProject,
		clearedFields: make(map[string]struct{}),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// withProjectID sets the ID field of the mutation.
func withProjectID(id uuid.UUID) projectOption {
	return func(m *ProjectMutation) {
		var (
			err   error
			once  sync.Once
			value *Project
		)
		m.oldValue = func(ctx context.Context) (*Project, error) {
			once.Do(func() {
				if m.done {
					err = errors.New("querying old values post mutation is not allowed")
				} else {
					value, err = m.Client().Project.Get(ctx, id)
				}
			})
			return value, err
		}
		m.id = &id
	}
}

// withProject sets the old Project of the mutation.
func withProject(node *Project) projectOption {
	return func(m *ProjectMutation) {
		m.oldValue = func(context.Context) (*Project, error) {
			return node, nil
		}
		m.id = &node.ID
	}
}

// Client returns a new `ent.Client` from the mutation. If the mutation was
// executed in a transaction (ent.Tx), a transactional client is returned.
func (m ProjectMutation) Client() *Client {
	client := &Client{config: m.config}
	client.init()
	return client
}

// Tx returns an `ent.Tx` for mutations that were executed in transactions;
// it returns an error otherwise.
func (m ProjectMutation) Tx() (*Tx, error) {
	if _, ok := m.driver.(*txDriver); !ok {
		return nil, errors.New("ent: mutation is not running in a transaction")
	}
	tx := &Tx{config: m.config}
	tx.init()
	return tx, nil
}

// SetID sets the value of the id field. Note that this
// operation is only accepted on creation of Project entities.
func (m *ProjectMutation) SetID(id uuid.UUID) {
	m.id = &id
}

// ID returns the ID value in the mutation. Note that the ID is only available
// if it was provided to the builder or after it was returned from the database.
func (m *ProjectMutation) ID() (id uuid.UUID, exists bool) {
	if m.id == nil {
		return
	}
	return *m.id, true
}

// IDs queries the database and returns the entity ids that match the mutation's predicate.
// That means, if the mutation is applied within a transaction with an isolation level such
// as sql.LevelSerializable, the returned ids match the ids of the rows that will be updated
// or updated by the mutation.
func (m *ProjectMutation) IDs(ctx context.Context) ([]uuid.UUID, error) {
	switch {
	case m.op.Is(OpUpdateOne | OpDeleteOne):
		id, exists := m.ID()
		if exists {
			return []uuid.UUID{id}, nil
		}
		fallthrough
	case m.op.Is(OpUpdate | OpDelete):
		return m.Client().Project.Query().Where(m.predicates...).IDs(ctx)
	default:
		return nil, fmt.Errorf("IDs is not allowed on %s operations", m.op)
	}
}

// SetCreateTime sets the "create_time" field.
func (m *ProjectMutation) SetCreateTime(t time.Time) {
	m.create_time = &t
}

// CreateTime returns the value of the "create_time" field in the mutation.
func (m *ProjectMutation) CreateTime() (r time.Time, exists bool) {
	v := m.create_time
	if v == nil {
		return
	}
	return *v, true
}

// OldCreateTime returns the old "create_time" field's value of the Project entity.
// If the Project object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *ProjectMutation) OldCreateTime(ctx context.Context) (v time.Time, err error) {
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
func (m *ProjectMutation) ResetCreateTime() {
	m.create_time = nil
}

// SetUpdateTime sets the "update_time" field.
func (m *ProjectMutation) SetUpdateTime(t time.Time) {
	m.update_time = &t
}

// UpdateTime returns the value of the "update_time" field in the mutation.
func (m *ProjectMutation) UpdateTime() (r time.Time, exists bool) {
	v := m.update_time
	if v == nil {
		return
	}
	return *v, true
}

// OldUpdateTime returns the old "update_time" field's value of the Project entity.
// If the Project object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *ProjectMutation) OldUpdateTime(ctx context.Context) (v time.Time, err error) {
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
func (m *ProjectMutation) ResetUpdateTime() {
	m.update_time = nil
}

// SetTitle sets the "title" field.
func (m *ProjectMutation) SetTitle(s string) {
	m.title = &s
}

// Title returns the value of the "title" field in the mutation.
func (m *ProjectMutation) Title() (r string, exists bool) {
	v := m.title
	if v == nil {
		return
	}
	return *v, true
}

// OldTitle returns the old "title" field's value of the Project entity.
// If the Project object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *ProjectMutation) OldTitle(ctx context.Context) (v string, err error) {
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
func (m *ProjectMutation) ResetTitle() {
	m.title = nil
}

// SetDescription sets the "description" field.
func (m *ProjectMutation) SetDescription(s string) {
	m.description = &s
}

// Description returns the value of the "description" field in the mutation.
func (m *ProjectMutation) Description() (r string, exists bool) {
	v := m.description
	if v == nil {
		return
	}
	return *v, true
}

// OldDescription returns the old "description" field's value of the Project entity.
// If the Project object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *ProjectMutation) OldDescription(ctx context.Context) (v string, err error) {
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
func (m *ProjectMutation) ResetDescription() {
	m.description = nil
}

// AddIssueIDs adds the "issues" edge to the Issue entity by ids.
func (m *ProjectMutation) AddIssueIDs(ids ...uuid.UUID) {
	if m.issues == nil {
		m.issues = make(map[uuid.UUID]struct{})
	}
	for i := range ids {
		m.issues[ids[i]] = struct{}{}
	}
}

// ClearIssues clears the "issues" edge to the Issue entity.
func (m *ProjectMutation) ClearIssues() {
	m.clearedissues = true
}

// IssuesCleared reports if the "issues" edge to the Issue entity was cleared.
func (m *ProjectMutation) IssuesCleared() bool {
	return m.clearedissues
}

// RemoveIssueIDs removes the "issues" edge to the Issue entity by IDs.
func (m *ProjectMutation) RemoveIssueIDs(ids ...uuid.UUID) {
	if m.removedissues == nil {
		m.removedissues = make(map[uuid.UUID]struct{})
	}
	for i := range ids {
		delete(m.issues, ids[i])
		m.removedissues[ids[i]] = struct{}{}
	}
}

// RemovedIssues returns the removed IDs of the "issues" edge to the Issue entity.
func (m *ProjectMutation) RemovedIssuesIDs() (ids []uuid.UUID) {
	for id := range m.removedissues {
		ids = append(ids, id)
	}
	return
}

// IssuesIDs returns the "issues" edge IDs in the mutation.
func (m *ProjectMutation) IssuesIDs() (ids []uuid.UUID) {
	for id := range m.issues {
		ids = append(ids, id)
	}
	return
}

// ResetIssues resets all changes to the "issues" edge.
func (m *ProjectMutation) ResetIssues() {
	m.issues = nil
	m.clearedissues = false
	m.removedissues = nil
}

// Where appends a list predicates to the ProjectMutation builder.
func (m *ProjectMutation) Where(ps ...predicate.Project) {
	m.predicates = append(m.predicates, ps...)
}

// Op returns the operation name.
func (m *ProjectMutation) Op() Op {
	return m.op
}

// Type returns the node type of this mutation (Project).
func (m *ProjectMutation) Type() string {
	return m.typ
}

// Fields returns all fields that were changed during this mutation. Note that in
// order to get all numeric fields that were incremented/decremented, call
// AddedFields().
func (m *ProjectMutation) Fields() []string {
	fields := make([]string, 0, 4)
	if m.create_time != nil {
		fields = append(fields, project.FieldCreateTime)
	}
	if m.update_time != nil {
		fields = append(fields, project.FieldUpdateTime)
	}
	if m.title != nil {
		fields = append(fields, project.FieldTitle)
	}
	if m.description != nil {
		fields = append(fields, project.FieldDescription)
	}
	return fields
}

// Field returns the value of a field with the given name. The second boolean
// return value indicates that this field was not set, or was not defined in the
// schema.
func (m *ProjectMutation) Field(name string) (ent.Value, bool) {
	switch name {
	case project.FieldCreateTime:
		return m.CreateTime()
	case project.FieldUpdateTime:
		return m.UpdateTime()
	case project.FieldTitle:
		return m.Title()
	case project.FieldDescription:
		return m.Description()
	}
	return nil, false
}

// OldField returns the old value of the field from the database. An error is
// returned if the mutation operation is not UpdateOne, or the query to the
// database failed.
func (m *ProjectMutation) OldField(ctx context.Context, name string) (ent.Value, error) {
	switch name {
	case project.FieldCreateTime:
		return m.OldCreateTime(ctx)
	case project.FieldUpdateTime:
		return m.OldUpdateTime(ctx)
	case project.FieldTitle:
		return m.OldTitle(ctx)
	case project.FieldDescription:
		return m.OldDescription(ctx)
	}
	return nil, fmt.Errorf("unknown Project field %s", name)
}

// SetField sets the value of a field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *ProjectMutation) SetField(name string, value ent.Value) error {
	switch name {
	case project.FieldCreateTime:
		v, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetCreateTime(v)
		return nil
	case project.FieldUpdateTime:
		v, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetUpdateTime(v)
		return nil
	case project.FieldTitle:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetTitle(v)
		return nil
	case project.FieldDescription:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetDescription(v)
		return nil
	}
	return fmt.Errorf("unknown Project field %s", name)
}

// AddedFields returns all numeric fields that were incremented/decremented during
// this mutation.
func (m *ProjectMutation) AddedFields() []string {
	return nil
}

// AddedField returns the numeric value that was incremented/decremented on a field
// with the given name. The second boolean return value indicates that this field
// was not set, or was not defined in the schema.
func (m *ProjectMutation) AddedField(name string) (ent.Value, bool) {
	return nil, false
}

// AddField adds the value to the field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *ProjectMutation) AddField(name string, value ent.Value) error {
	switch name {
	}
	return fmt.Errorf("unknown Project numeric field %s", name)
}

// ClearedFields returns all nullable fields that were cleared during this
// mutation.
func (m *ProjectMutation) ClearedFields() []string {
	return nil
}

// FieldCleared returns a boolean indicating if a field with the given name was
// cleared in this mutation.
func (m *ProjectMutation) FieldCleared(name string) bool {
	_, ok := m.clearedFields[name]
	return ok
}

// ClearField clears the value of the field with the given name. It returns an
// error if the field is not defined in the schema.
func (m *ProjectMutation) ClearField(name string) error {
	return fmt.Errorf("unknown Project nullable field %s", name)
}

// ResetField resets all changes in the mutation for the field with the given name.
// It returns an error if the field is not defined in the schema.
func (m *ProjectMutation) ResetField(name string) error {
	switch name {
	case project.FieldCreateTime:
		m.ResetCreateTime()
		return nil
	case project.FieldUpdateTime:
		m.ResetUpdateTime()
		return nil
	case project.FieldTitle:
		m.ResetTitle()
		return nil
	case project.FieldDescription:
		m.ResetDescription()
		return nil
	}
	return fmt.Errorf("unknown Project field %s", name)
}

// AddedEdges returns all edge names that were set/added in this mutation.
func (m *ProjectMutation) AddedEdges() []string {
	edges := make([]string, 0, 1)
	if m.issues != nil {
		edges = append(edges, project.EdgeIssues)
	}
	return edges
}

// AddedIDs returns all IDs (to other nodes) that were added for the given edge
// name in this mutation.
func (m *ProjectMutation) AddedIDs(name string) []ent.Value {
	switch name {
	case project.EdgeIssues:
		ids := make([]ent.Value, 0, len(m.issues))
		for id := range m.issues {
			ids = append(ids, id)
		}
		return ids
	}
	return nil
}

// RemovedEdges returns all edge names that were removed in this mutation.
func (m *ProjectMutation) RemovedEdges() []string {
	edges := make([]string, 0, 1)
	if m.removedissues != nil {
		edges = append(edges, project.EdgeIssues)
	}
	return edges
}

// RemovedIDs returns all IDs (to other nodes) that were removed for the edge with
// the given name in this mutation.
func (m *ProjectMutation) RemovedIDs(name string) []ent.Value {
	switch name {
	case project.EdgeIssues:
		ids := make([]ent.Value, 0, len(m.removedissues))
		for id := range m.removedissues {
			ids = append(ids, id)
		}
		return ids
	}
	return nil
}

// ClearedEdges returns all edge names that were cleared in this mutation.
func (m *ProjectMutation) ClearedEdges() []string {
	edges := make([]string, 0, 1)
	if m.clearedissues {
		edges = append(edges, project.EdgeIssues)
	}
	return edges
}

// EdgeCleared returns a boolean which indicates if the edge with the given name
// was cleared in this mutation.
func (m *ProjectMutation) EdgeCleared(name string) bool {
	switch name {
	case project.EdgeIssues:
		return m.clearedissues
	}
	return false
}

// ClearEdge clears the value of the edge with the given name. It returns an error
// if that edge is not defined in the schema.
func (m *ProjectMutation) ClearEdge(name string) error {
	switch name {
	}
	return fmt.Errorf("unknown Project unique edge %s", name)
}

// ResetEdge resets all changes to the edge with the given name in this mutation.
// It returns an error if the edge is not defined in the schema.
func (m *ProjectMutation) ResetEdge(name string) error {
	switch name {
	case project.EdgeIssues:
		m.ResetIssues()
		return nil
	}
	return fmt.Errorf("unknown Project edge %s", name)
}
