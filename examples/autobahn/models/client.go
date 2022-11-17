// Code generated (@generated) by entc, DO NOT EDIT.

package models

import (
	"context"
	"fmt"
	"log"

	"github.com/adnaan/autobahn/models/migrate"
	"github.com/google/uuid"

	"github.com/adnaan/autobahn/models/board"
	"github.com/adnaan/autobahn/models/comment"
	"github.com/adnaan/autobahn/models/label"
	"github.com/adnaan/autobahn/models/story"
	"github.com/adnaan/autobahn/models/view"

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
	// Comment is the client for interacting with the Comment builders.
	Comment *CommentClient
	// Label is the client for interacting with the Label builders.
	Label *LabelClient
	// Story is the client for interacting with the Story builders.
	Story *StoryClient
	// View is the client for interacting with the View builders.
	View *ViewClient
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
	c.Comment = NewCommentClient(c.config)
	c.Label = NewLabelClient(c.config)
	c.Story = NewStoryClient(c.config)
	c.View = NewViewClient(c.config)
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
		ctx:     ctx,
		config:  cfg,
		Board:   NewBoardClient(cfg),
		Comment: NewCommentClient(cfg),
		Label:   NewLabelClient(cfg),
		Story:   NewStoryClient(cfg),
		View:    NewViewClient(cfg),
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
		ctx:     ctx,
		config:  cfg,
		Board:   NewBoardClient(cfg),
		Comment: NewCommentClient(cfg),
		Label:   NewLabelClient(cfg),
		Story:   NewStoryClient(cfg),
		View:    NewViewClient(cfg),
	}, nil
}

// Debug returns a new debug-client. It's used to get verbose logging on specific operations.
//
//	client.Debug().
//		Board.
//		Query().
//		Count(ctx)
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
	c.Comment.Use(hooks...)
	c.Label.Use(hooks...)
	c.Story.Use(hooks...)
	c.View.Use(hooks...)
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

// QueryStories queries the stories edge of a Board.
func (c *BoardClient) QueryStories(b *Board) *StoryQuery {
	query := &StoryQuery{config: c.config}
	query.path = func(ctx context.Context) (fromV *sql.Selector, _ error) {
		id := b.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(board.Table, board.FieldID, id),
			sqlgraph.To(story.Table, story.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, board.StoriesTable, board.StoriesColumn),
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

// CommentClient is a client for the Comment schema.
type CommentClient struct {
	config
}

// NewCommentClient returns a client for the Comment from the given config.
func NewCommentClient(c config) *CommentClient {
	return &CommentClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `comment.Hooks(f(g(h())))`.
func (c *CommentClient) Use(hooks ...Hook) {
	c.hooks.Comment = append(c.hooks.Comment, hooks...)
}

// Create returns a create builder for Comment.
func (c *CommentClient) Create() *CommentCreate {
	mutation := newCommentMutation(c.config, OpCreate)
	return &CommentCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of Comment entities.
func (c *CommentClient) CreateBulk(builders ...*CommentCreate) *CommentCreateBulk {
	return &CommentCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for Comment.
func (c *CommentClient) Update() *CommentUpdate {
	mutation := newCommentMutation(c.config, OpUpdate)
	return &CommentUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *CommentClient) UpdateOne(co *Comment) *CommentUpdateOne {
	mutation := newCommentMutation(c.config, OpUpdateOne, withComment(co))
	return &CommentUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *CommentClient) UpdateOneID(id uuid.UUID) *CommentUpdateOne {
	mutation := newCommentMutation(c.config, OpUpdateOne, withCommentID(id))
	return &CommentUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for Comment.
func (c *CommentClient) Delete() *CommentDelete {
	mutation := newCommentMutation(c.config, OpDelete)
	return &CommentDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a delete builder for the given entity.
func (c *CommentClient) DeleteOne(co *Comment) *CommentDeleteOne {
	return c.DeleteOneID(co.ID)
}

// DeleteOneID returns a delete builder for the given id.
func (c *CommentClient) DeleteOneID(id uuid.UUID) *CommentDeleteOne {
	builder := c.Delete().Where(comment.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &CommentDeleteOne{builder}
}

// Query returns a query builder for Comment.
func (c *CommentClient) Query() *CommentQuery {
	return &CommentQuery{
		config: c.config,
	}
}

// Get returns a Comment entity by its id.
func (c *CommentClient) Get(ctx context.Context, id uuid.UUID) (*Comment, error) {
	return c.Query().Where(comment.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *CommentClient) GetX(ctx context.Context, id uuid.UUID) *Comment {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// Hooks returns the client hooks.
func (c *CommentClient) Hooks() []Hook {
	return c.hooks.Comment
}

// LabelClient is a client for the Label schema.
type LabelClient struct {
	config
}

// NewLabelClient returns a client for the Label from the given config.
func NewLabelClient(c config) *LabelClient {
	return &LabelClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `label.Hooks(f(g(h())))`.
func (c *LabelClient) Use(hooks ...Hook) {
	c.hooks.Label = append(c.hooks.Label, hooks...)
}

// Create returns a create builder for Label.
func (c *LabelClient) Create() *LabelCreate {
	mutation := newLabelMutation(c.config, OpCreate)
	return &LabelCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of Label entities.
func (c *LabelClient) CreateBulk(builders ...*LabelCreate) *LabelCreateBulk {
	return &LabelCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for Label.
func (c *LabelClient) Update() *LabelUpdate {
	mutation := newLabelMutation(c.config, OpUpdate)
	return &LabelUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *LabelClient) UpdateOne(l *Label) *LabelUpdateOne {
	mutation := newLabelMutation(c.config, OpUpdateOne, withLabel(l))
	return &LabelUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *LabelClient) UpdateOneID(id uuid.UUID) *LabelUpdateOne {
	mutation := newLabelMutation(c.config, OpUpdateOne, withLabelID(id))
	return &LabelUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for Label.
func (c *LabelClient) Delete() *LabelDelete {
	mutation := newLabelMutation(c.config, OpDelete)
	return &LabelDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a delete builder for the given entity.
func (c *LabelClient) DeleteOne(l *Label) *LabelDeleteOne {
	return c.DeleteOneID(l.ID)
}

// DeleteOneID returns a delete builder for the given id.
func (c *LabelClient) DeleteOneID(id uuid.UUID) *LabelDeleteOne {
	builder := c.Delete().Where(label.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &LabelDeleteOne{builder}
}

// Query returns a query builder for Label.
func (c *LabelClient) Query() *LabelQuery {
	return &LabelQuery{
		config: c.config,
	}
}

// Get returns a Label entity by its id.
func (c *LabelClient) Get(ctx context.Context, id uuid.UUID) (*Label, error) {
	return c.Query().Where(label.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *LabelClient) GetX(ctx context.Context, id uuid.UUID) *Label {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// Hooks returns the client hooks.
func (c *LabelClient) Hooks() []Hook {
	return c.hooks.Label
}

// StoryClient is a client for the Story schema.
type StoryClient struct {
	config
}

// NewStoryClient returns a client for the Story from the given config.
func NewStoryClient(c config) *StoryClient {
	return &StoryClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `story.Hooks(f(g(h())))`.
func (c *StoryClient) Use(hooks ...Hook) {
	c.hooks.Story = append(c.hooks.Story, hooks...)
}

// Create returns a create builder for Story.
func (c *StoryClient) Create() *StoryCreate {
	mutation := newStoryMutation(c.config, OpCreate)
	return &StoryCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of Story entities.
func (c *StoryClient) CreateBulk(builders ...*StoryCreate) *StoryCreateBulk {
	return &StoryCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for Story.
func (c *StoryClient) Update() *StoryUpdate {
	mutation := newStoryMutation(c.config, OpUpdate)
	return &StoryUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *StoryClient) UpdateOne(s *Story) *StoryUpdateOne {
	mutation := newStoryMutation(c.config, OpUpdateOne, withStory(s))
	return &StoryUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *StoryClient) UpdateOneID(id uuid.UUID) *StoryUpdateOne {
	mutation := newStoryMutation(c.config, OpUpdateOne, withStoryID(id))
	return &StoryUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for Story.
func (c *StoryClient) Delete() *StoryDelete {
	mutation := newStoryMutation(c.config, OpDelete)
	return &StoryDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a delete builder for the given entity.
func (c *StoryClient) DeleteOne(s *Story) *StoryDeleteOne {
	return c.DeleteOneID(s.ID)
}

// DeleteOneID returns a delete builder for the given id.
func (c *StoryClient) DeleteOneID(id uuid.UUID) *StoryDeleteOne {
	builder := c.Delete().Where(story.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &StoryDeleteOne{builder}
}

// Query returns a query builder for Story.
func (c *StoryClient) Query() *StoryQuery {
	return &StoryQuery{
		config: c.config,
	}
}

// Get returns a Story entity by its id.
func (c *StoryClient) Get(ctx context.Context, id uuid.UUID) (*Story, error) {
	return c.Query().Where(story.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *StoryClient) GetX(ctx context.Context, id uuid.UUID) *Story {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// QueryOwner queries the owner edge of a Story.
func (c *StoryClient) QueryOwner(s *Story) *BoardQuery {
	query := &BoardQuery{config: c.config}
	query.path = func(ctx context.Context) (fromV *sql.Selector, _ error) {
		id := s.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(story.Table, story.FieldID, id),
			sqlgraph.To(board.Table, board.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, story.OwnerTable, story.OwnerColumn),
		)
		fromV = sqlgraph.Neighbors(s.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// Hooks returns the client hooks.
func (c *StoryClient) Hooks() []Hook {
	return c.hooks.Story
}

// ViewClient is a client for the View schema.
type ViewClient struct {
	config
}

// NewViewClient returns a client for the View from the given config.
func NewViewClient(c config) *ViewClient {
	return &ViewClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `view.Hooks(f(g(h())))`.
func (c *ViewClient) Use(hooks ...Hook) {
	c.hooks.View = append(c.hooks.View, hooks...)
}

// Create returns a create builder for View.
func (c *ViewClient) Create() *ViewCreate {
	mutation := newViewMutation(c.config, OpCreate)
	return &ViewCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of View entities.
func (c *ViewClient) CreateBulk(builders ...*ViewCreate) *ViewCreateBulk {
	return &ViewCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for View.
func (c *ViewClient) Update() *ViewUpdate {
	mutation := newViewMutation(c.config, OpUpdate)
	return &ViewUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *ViewClient) UpdateOne(v *View) *ViewUpdateOne {
	mutation := newViewMutation(c.config, OpUpdateOne, withView(v))
	return &ViewUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *ViewClient) UpdateOneID(id uuid.UUID) *ViewUpdateOne {
	mutation := newViewMutation(c.config, OpUpdateOne, withViewID(id))
	return &ViewUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for View.
func (c *ViewClient) Delete() *ViewDelete {
	mutation := newViewMutation(c.config, OpDelete)
	return &ViewDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a delete builder for the given entity.
func (c *ViewClient) DeleteOne(v *View) *ViewDeleteOne {
	return c.DeleteOneID(v.ID)
}

// DeleteOneID returns a delete builder for the given id.
func (c *ViewClient) DeleteOneID(id uuid.UUID) *ViewDeleteOne {
	builder := c.Delete().Where(view.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &ViewDeleteOne{builder}
}

// Query returns a query builder for View.
func (c *ViewClient) Query() *ViewQuery {
	return &ViewQuery{
		config: c.config,
	}
}

// Get returns a View entity by its id.
func (c *ViewClient) Get(ctx context.Context, id uuid.UUID) (*View, error) {
	return c.Query().Where(view.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *ViewClient) GetX(ctx context.Context, id uuid.UUID) *View {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// Hooks returns the client hooks.
func (c *ViewClient) Hooks() []Hook {
	return c.hooks.View
}
