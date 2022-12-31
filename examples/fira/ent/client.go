// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/livefir/fir/examples/fira/ent/migrate"

	"github.com/livefir/fir/examples/fira/ent/issue"
	"github.com/livefir/fir/examples/fira/ent/project"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
)

// Client is the client that holds all ent builders.
type Client struct {
	config
	// Schema is the client for creating, migrating and dropping schema.
	Schema *migrate.Schema
	// Issue is the client for interacting with the Issue builders.
	Issue *IssueClient
	// Project is the client for interacting with the Project builders.
	Project *ProjectClient
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
	c.Issue = NewIssueClient(c.config)
	c.Project = NewProjectClient(c.config)
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
		return nil, errors.New("ent: cannot start a transaction within a transaction")
	}
	tx, err := newTx(ctx, c.driver)
	if err != nil {
		return nil, fmt.Errorf("ent: starting a transaction: %w", err)
	}
	cfg := c.config
	cfg.driver = tx
	return &Tx{
		ctx:     ctx,
		config:  cfg,
		Issue:   NewIssueClient(cfg),
		Project: NewProjectClient(cfg),
	}, nil
}

// BeginTx returns a transactional client with specified options.
func (c *Client) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	if _, ok := c.driver.(*txDriver); ok {
		return nil, errors.New("ent: cannot start a transaction within a transaction")
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
		Issue:   NewIssueClient(cfg),
		Project: NewProjectClient(cfg),
	}, nil
}

// Debug returns a new debug-client. It's used to get verbose logging on specific operations.
//
//	client.Debug().
//		Issue.
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
	c.Issue.Use(hooks...)
	c.Project.Use(hooks...)
}

// IssueClient is a client for the Issue schema.
type IssueClient struct {
	config
}

// NewIssueClient returns a client for the Issue from the given config.
func NewIssueClient(c config) *IssueClient {
	return &IssueClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `issue.Hooks(f(g(h())))`.
func (c *IssueClient) Use(hooks ...Hook) {
	c.hooks.Issue = append(c.hooks.Issue, hooks...)
}

// Create returns a builder for creating a Issue entity.
func (c *IssueClient) Create() *IssueCreate {
	mutation := newIssueMutation(c.config, OpCreate)
	return &IssueCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of Issue entities.
func (c *IssueClient) CreateBulk(builders ...*IssueCreate) *IssueCreateBulk {
	return &IssueCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for Issue.
func (c *IssueClient) Update() *IssueUpdate {
	mutation := newIssueMutation(c.config, OpUpdate)
	return &IssueUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *IssueClient) UpdateOne(i *Issue) *IssueUpdateOne {
	mutation := newIssueMutation(c.config, OpUpdateOne, withIssue(i))
	return &IssueUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *IssueClient) UpdateOneID(id uuid.UUID) *IssueUpdateOne {
	mutation := newIssueMutation(c.config, OpUpdateOne, withIssueID(id))
	return &IssueUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for Issue.
func (c *IssueClient) Delete() *IssueDelete {
	mutation := newIssueMutation(c.config, OpDelete)
	return &IssueDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a builder for deleting the given entity.
func (c *IssueClient) DeleteOne(i *Issue) *IssueDeleteOne {
	return c.DeleteOneID(i.ID)
}

// DeleteOneID returns a builder for deleting the given entity by its id.
func (c *IssueClient) DeleteOneID(id uuid.UUID) *IssueDeleteOne {
	builder := c.Delete().Where(issue.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &IssueDeleteOne{builder}
}

// Query returns a query builder for Issue.
func (c *IssueClient) Query() *IssueQuery {
	return &IssueQuery{
		config: c.config,
	}
}

// Get returns a Issue entity by its id.
func (c *IssueClient) Get(ctx context.Context, id uuid.UUID) (*Issue, error) {
	return c.Query().Where(issue.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *IssueClient) GetX(ctx context.Context, id uuid.UUID) *Issue {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// QueryOwner queries the owner edge of a Issue.
func (c *IssueClient) QueryOwner(i *Issue) *ProjectQuery {
	query := &ProjectQuery{config: c.config}
	query.path = func(context.Context) (fromV *sql.Selector, _ error) {
		id := i.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(issue.Table, issue.FieldID, id),
			sqlgraph.To(project.Table, project.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, issue.OwnerTable, issue.OwnerColumn),
		)
		fromV = sqlgraph.Neighbors(i.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// Hooks returns the client hooks.
func (c *IssueClient) Hooks() []Hook {
	return c.hooks.Issue
}

// ProjectClient is a client for the Project schema.
type ProjectClient struct {
	config
}

// NewProjectClient returns a client for the Project from the given config.
func NewProjectClient(c config) *ProjectClient {
	return &ProjectClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `project.Hooks(f(g(h())))`.
func (c *ProjectClient) Use(hooks ...Hook) {
	c.hooks.Project = append(c.hooks.Project, hooks...)
}

// Create returns a builder for creating a Project entity.
func (c *ProjectClient) Create() *ProjectCreate {
	mutation := newProjectMutation(c.config, OpCreate)
	return &ProjectCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of Project entities.
func (c *ProjectClient) CreateBulk(builders ...*ProjectCreate) *ProjectCreateBulk {
	return &ProjectCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for Project.
func (c *ProjectClient) Update() *ProjectUpdate {
	mutation := newProjectMutation(c.config, OpUpdate)
	return &ProjectUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *ProjectClient) UpdateOne(pr *Project) *ProjectUpdateOne {
	mutation := newProjectMutation(c.config, OpUpdateOne, withProject(pr))
	return &ProjectUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *ProjectClient) UpdateOneID(id uuid.UUID) *ProjectUpdateOne {
	mutation := newProjectMutation(c.config, OpUpdateOne, withProjectID(id))
	return &ProjectUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for Project.
func (c *ProjectClient) Delete() *ProjectDelete {
	mutation := newProjectMutation(c.config, OpDelete)
	return &ProjectDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a builder for deleting the given entity.
func (c *ProjectClient) DeleteOne(pr *Project) *ProjectDeleteOne {
	return c.DeleteOneID(pr.ID)
}

// DeleteOneID returns a builder for deleting the given entity by its id.
func (c *ProjectClient) DeleteOneID(id uuid.UUID) *ProjectDeleteOne {
	builder := c.Delete().Where(project.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &ProjectDeleteOne{builder}
}

// Query returns a query builder for Project.
func (c *ProjectClient) Query() *ProjectQuery {
	return &ProjectQuery{
		config: c.config,
	}
}

// Get returns a Project entity by its id.
func (c *ProjectClient) Get(ctx context.Context, id uuid.UUID) (*Project, error) {
	return c.Query().Where(project.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *ProjectClient) GetX(ctx context.Context, id uuid.UUID) *Project {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// QueryIssues queries the issues edge of a Project.
func (c *ProjectClient) QueryIssues(pr *Project) *IssueQuery {
	query := &IssueQuery{config: c.config}
	query.path = func(context.Context) (fromV *sql.Selector, _ error) {
		id := pr.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(project.Table, project.FieldID, id),
			sqlgraph.To(issue.Table, issue.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, project.IssuesTable, project.IssuesColumn),
		)
		fromV = sqlgraph.Neighbors(pr.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// Hooks returns the client hooks.
func (c *ProjectClient) Hooks() []Hook {
	return c.hooks.Project
}
