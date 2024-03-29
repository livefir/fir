// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/livefir/fir/examples/fira/ent/issue"
	"github.com/livefir/fir/examples/fira/ent/project"
)

// Issue is the model entity for the Issue schema.
type Issue struct {
	config `json:"-"`
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// CreateTime holds the value of the "create_time" field.
	CreateTime time.Time `json:"create_time,omitempty"`
	// UpdateTime holds the value of the "update_time" field.
	UpdateTime time.Time `json:"update_time,omitempty"`
	// Title holds the value of the "title" field.
	Title string `json:"title,omitempty"`
	// Description holds the value of the "description" field.
	Description string `json:"description,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the IssueQuery when eager-loading is set.
	Edges          IssueEdges `json:"edges"`
	project_issues *uuid.UUID
}

// IssueEdges holds the relations/edges for other nodes in the graph.
type IssueEdges struct {
	// Owner holds the value of the owner edge.
	Owner *Project `json:"owner,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [1]bool
}

// OwnerOrErr returns the Owner value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e IssueEdges) OwnerOrErr() (*Project, error) {
	if e.loadedTypes[0] {
		if e.Owner == nil {
			// Edge was loaded but was not found.
			return nil, &NotFoundError{label: project.Label}
		}
		return e.Owner, nil
	}
	return nil, &NotLoadedError{edge: "owner"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Issue) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case issue.FieldTitle, issue.FieldDescription:
			values[i] = new(sql.NullString)
		case issue.FieldCreateTime, issue.FieldUpdateTime:
			values[i] = new(sql.NullTime)
		case issue.FieldID:
			values[i] = new(uuid.UUID)
		case issue.ForeignKeys[0]: // project_issues
			values[i] = &sql.NullScanner{S: new(uuid.UUID)}
		default:
			return nil, fmt.Errorf("unexpected column %q for type Issue", columns[i])
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Issue fields.
func (i *Issue) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for j := range columns {
		switch columns[j] {
		case issue.FieldID:
			if value, ok := values[j].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[j])
			} else if value != nil {
				i.ID = *value
			}
		case issue.FieldCreateTime:
			if value, ok := values[j].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field create_time", values[j])
			} else if value.Valid {
				i.CreateTime = value.Time
			}
		case issue.FieldUpdateTime:
			if value, ok := values[j].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field update_time", values[j])
			} else if value.Valid {
				i.UpdateTime = value.Time
			}
		case issue.FieldTitle:
			if value, ok := values[j].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field title", values[j])
			} else if value.Valid {
				i.Title = value.String
			}
		case issue.FieldDescription:
			if value, ok := values[j].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field description", values[j])
			} else if value.Valid {
				i.Description = value.String
			}
		case issue.ForeignKeys[0]:
			if value, ok := values[j].(*sql.NullScanner); !ok {
				return fmt.Errorf("unexpected type %T for field project_issues", values[j])
			} else if value.Valid {
				i.project_issues = new(uuid.UUID)
				*i.project_issues = *value.S.(*uuid.UUID)
			}
		}
	}
	return nil
}

// QueryOwner queries the "owner" edge of the Issue entity.
func (i *Issue) QueryOwner() *ProjectQuery {
	return (&IssueClient{config: i.config}).QueryOwner(i)
}

// Update returns a builder for updating this Issue.
// Note that you need to call Issue.Unwrap() before calling this method if this Issue
// was returned from a transaction, and the transaction was committed or rolled back.
func (i *Issue) Update() *IssueUpdateOne {
	return (&IssueClient{config: i.config}).UpdateOne(i)
}

// Unwrap unwraps the Issue entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (i *Issue) Unwrap() *Issue {
	_tx, ok := i.config.driver.(*txDriver)
	if !ok {
		panic("ent: Issue is not a transactional entity")
	}
	i.config.driver = _tx.drv
	return i
}

// String implements the fmt.Stringer.
func (i *Issue) String() string {
	var builder strings.Builder
	builder.WriteString("Issue(")
	builder.WriteString(fmt.Sprintf("id=%v, ", i.ID))
	builder.WriteString("create_time=")
	builder.WriteString(i.CreateTime.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("update_time=")
	builder.WriteString(i.UpdateTime.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("title=")
	builder.WriteString(i.Title)
	builder.WriteString(", ")
	builder.WriteString("description=")
	builder.WriteString(i.Description)
	builder.WriteByte(')')
	return builder.String()
}

// Issues is a parsable slice of Issue.
type Issues []*Issue

func (i Issues) config(cfg config) {
	for _i := range i {
		i[_i].config = cfg
	}
}
