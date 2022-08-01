// Code generated (@generated) by entc, DO NOT EDIT.

package models

import (
	"context"
	"fmt"
	"log"

	"github.com/adnaan/fir/cli/testdata/todos/models/migrate"
	"github.com/google/uuid"

	"github.com/adnaan/fir/cli/testdata/todos/models/board"
	"github.com/adnaan/fir/cli/testdata/todos/models/todo"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
)

// Client is the client that holds all ent builders.
type Client struct {
	config
	// Schema is the client for creating, migrating and dropping schema.
	Schema *migrate.Schema
	// Board is the client for interacting with the Board builders.
	Board *BoardClient
	// Todo is the client for interacting with the Todo builders.
	Todo *TodoClient
}

// NewClient creates a new client configured with the given options.
func NewClient(opts ...Option) *Client {
	cfg := config{log: log.Println, hooks: &hooks{}}
	cfg.options(opts...)
	client := &Client{config: cfg}
	client.init()
	return client
}

func (c *Client) init() {
	c.Schema = migrate.NewSchema(c.driver)
	c.Board = NewBoardClient(c.config)
	c.Todo = NewTodoClient(c.config)
}

// Open opens a database/sql.DB specified by the driver name and
// the data source name, and returns a new client attached to it.
// Optional parameters can be added for configuring the client.
func Open(driverName, dataSourceName string, options ...Option) (*Client, error) {
	switch driverName {
	case dialect.MySQL, dialect.Postgres, dialect.SQLite:
		drv, err := sql.Open(driverName, dataSourceName)
		if err != nil {
			return nil, err
		}
		return NewClient(append(options, Driver(drv))...), nil
	default:
		return nil, fmt.Errorf("unsupported driver: %q", driverName)
	}
}

// Tx returns a new transactional client. The provided context
// is used until the transaction is committed or rolled back.
func (c *Client) Tx(ctx context.Context) (*Tx, error) {
	if _, ok := c.driver.(*txDriver); ok {
		return nil, fmt.Errorf("models: cannot start a transaction within a transaction")
	}
	tx, err := newTx(ctx, c.driver)
	if err != nil {
		return nil, fmt.Errorf("models: starting a transaction: %w", err)
	}
	cfg := c.config
	cfg.driver = tx
	return &Tx{
		ctx:    ctx,
		config: cfg,
		Board:  NewBoardClient(cfg),
		Todo:   NewTodoClient(cfg),
	}, nil
}

// BeginTx returns a transactional client with specified options.
func (c *Client) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	if _, ok := c.driver.(*txDriver); ok {
		return nil, fmt.Errorf("ent: cannot start a transaction within a transaction")
	}
	tx, err := c.driver.(interface {
		BeginTx(context.Context, *sql.TxOptions) (dialect.Tx, error)
	}).BeginTx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("ent: starting a transaction: %w", err)
	}
	cfg := c.config
	cfg.driver = &txDriver{tx: tx, drv: c.driver}
	return &Tx{
		ctx:    ctx,
		config: cfg,
		Board:  NewBoardClient(cfg),
		Todo:   NewTodoClient(cfg),
	}, nil
}

// Debug returns a new debug-client. It's used to get verbose logging on specific operations.
//
//	client.Debug().
//		Board.
//		Query().
//		Count(ctx)
//
func (c *Client) Debug() *Client {
	if c.debug {
		return c
	}
	cfg := c.config
	cfg.driver = dialect.Debug(c.driver, c.log)
	client := &Client{config: cfg}
	client.init()
	return client
}

// Close closes the database connection and prevents new queries from starting.
func (c *Client) Close() error {
	return c.driver.Close()
}

// Use adds the mutation hooks to all the entity clients.
// In order to add hooks to a specific client, call: `client.Node.Use(...)`.
func (c *Client) Use(hooks ...Hook) {
	c.Board.Use(hooks...)
	c.Todo.Use(hooks...)
}

// BoardClient is a client for the Board schema.
type BoardClient struct {
	config
}

// NewBoardClient returns a client for the Board from the given config.
func NewBoardClient(c config) *BoardClient {
	return &BoardClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `board.Hooks(f(g(h())))`.
func (c *BoardClient) Use(hooks ...Hook) {
	c.hooks.Board = append(c.hooks.Board, hooks...)
}

// Create returns a create builder for Board.
func (c *BoardClient) Create() *BoardCreate {
	mutation := newBoardMutation(c.config, OpCreate)
	return &BoardCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of Board entities.
func (c *BoardClient) CreateBulk(builders ...*BoardCreate) *BoardCreateBulk {
	return &BoardCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for Board.
func (c *BoardClient) Update() *BoardUpdate {
	mutation := newBoardMutation(c.config, OpUpdate)
	return &BoardUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *BoardClient) UpdateOne(b *Board) *BoardUpdateOne {
	mutation := newBoardMutation(c.config, OpUpdateOne, withBoard(b))
	return &BoardUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *BoardClient) UpdateOneID(id uuid.UUID) *BoardUpdateOne {
	mutation := newBoardMutation(c.config, OpUpdateOne, withBoardID(id))
	return &BoardUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for Board.
func (c *BoardClient) Delete() *BoardDelete {
	mutation := newBoardMutation(c.config, OpDelete)
	return &BoardDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a delete builder for the given entity.
func (c *BoardClient) DeleteOne(b *Board) *BoardDeleteOne {
	return c.DeleteOneID(b.ID)
}

// DeleteOneID returns a delete builder for the given id.
func (c *BoardClient) DeleteOneID(id uuid.UUID) *BoardDeleteOne {
	builder := c.Delete().Where(board.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &BoardDeleteOne{builder}
}

// Query returns a query builder for Board.
func (c *BoardClient) Query() *BoardQuery {
	return &BoardQuery{
		config: c.config,
	}
}

// Get returns a Board entity by its id.
func (c *BoardClient) Get(ctx context.Context, id uuid.UUID) (*Board, error) {
	return c.Query().Where(board.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *BoardClient) GetX(ctx context.Context, id uuid.UUID) *Board {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// QueryTodos queries the todos edge of a Board.
func (c *BoardClient) QueryTodos(b *Board) *TodoQuery {
	query := &TodoQuery{config: c.config}
	query.path = func(ctx context.Context) (fromV *sql.Selector, _ error) {
		id := b.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(board.Table, board.FieldID, id),
			sqlgraph.To(todo.Table, todo.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, board.TodosTable, board.TodosColumn),
		)
		fromV = sqlgraph.Neighbors(b.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// Hooks returns the client hooks.
func (c *BoardClient) Hooks() []Hook {
	return c.hooks.Board
}

// TodoClient is a client for the Todo schema.
type TodoClient struct {
	config
}

// NewTodoClient returns a client for the Todo from the given config.
func NewTodoClient(c config) *TodoClient {
	return &TodoClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `todo.Hooks(f(g(h())))`.
func (c *TodoClient) Use(hooks ...Hook) {
	c.hooks.Todo = append(c.hooks.Todo, hooks...)
}

// Create returns a create builder for Todo.
func (c *TodoClient) Create() *TodoCreate {
	mutation := newTodoMutation(c.config, OpCreate)
	return &TodoCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of Todo entities.
func (c *TodoClient) CreateBulk(builders ...*TodoCreate) *TodoCreateBulk {
	return &TodoCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for Todo.
func (c *TodoClient) Update() *TodoUpdate {
	mutation := newTodoMutation(c.config, OpUpdate)
	return &TodoUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *TodoClient) UpdateOne(t *Todo) *TodoUpdateOne {
	mutation := newTodoMutation(c.config, OpUpdateOne, withTodo(t))
	return &TodoUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *TodoClient) UpdateOneID(id uuid.UUID) *TodoUpdateOne {
	mutation := newTodoMutation(c.config, OpUpdateOne, withTodoID(id))
	return &TodoUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for Todo.
func (c *TodoClient) Delete() *TodoDelete {
	mutation := newTodoMutation(c.config, OpDelete)
	return &TodoDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a delete builder for the given entity.
func (c *TodoClient) DeleteOne(t *Todo) *TodoDeleteOne {
	return c.DeleteOneID(t.ID)
}

// DeleteOneID returns a delete builder for the given id.
func (c *TodoClient) DeleteOneID(id uuid.UUID) *TodoDeleteOne {
	builder := c.Delete().Where(todo.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &TodoDeleteOne{builder}
}

// Query returns a query builder for Todo.
func (c *TodoClient) Query() *TodoQuery {
	return &TodoQuery{
		config: c.config,
	}
}

// Get returns a Todo entity by its id.
func (c *TodoClient) Get(ctx context.Context, id uuid.UUID) (*Todo, error) {
	return c.Query().Where(todo.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *TodoClient) GetX(ctx context.Context, id uuid.UUID) *Todo {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// QueryOwner queries the owner edge of a Todo.
func (c *TodoClient) QueryOwner(t *Todo) *BoardQuery {
	query := &BoardQuery{config: c.config}
	query.path = func(ctx context.Context) (fromV *sql.Selector, _ error) {
		id := t.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(todo.Table, todo.FieldID, id),
			sqlgraph.To(board.Table, board.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, todo.OwnerTable, todo.OwnerColumn),
		)
		fromV = sqlgraph.Neighbors(t.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// Hooks returns the client hooks.
func (c *TodoClient) Hooks() []Hook {
	return c.hooks.Todo
}
