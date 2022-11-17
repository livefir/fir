// Code generated (@generated) by entc, DO NOT EDIT.

package models

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/adnaan/autobahn/models/board"
	"github.com/adnaan/autobahn/models/story"
	"github.com/google/uuid"
)

// Story is the model entity for the Story schema.
type Story struct {
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
	// The values are being populated by the StoryQuery when eager-loading is set.
	Edges         StoryEdges `json:"edges"`
	board_stories *uuid.UUID
}

// StoryEdges holds the relations/edges for other nodes in the graph.
type StoryEdges struct {
	// Owner holds the value of the owner edge.
	Owner *Board `json:"owner,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [1]bool
}

// OwnerOrErr returns the Owner value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e StoryEdges) OwnerOrErr() (*Board, error) {
	if e.loadedTypes[0] {
		if e.Owner == nil {
			// The edge owner was loaded in eager-loading,
			// but was not found.
			return nil, &NotFoundError{label: board.Label}
		}
		return e.Owner, nil
	}
	return nil, &NotLoadedError{edge: "owner"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Story) scanValues(columns []string) ([]interface{}, error) {
	values := make([]interface{}, len(columns))
	for i := range columns {
		switch columns[i] {
		case story.FieldTitle, story.FieldDescription:
			values[i] = new(sql.NullString)
		case story.FieldCreateTime, story.FieldUpdateTime:
			values[i] = new(sql.NullTime)
		case story.FieldID:
			values[i] = new(uuid.UUID)
		case story.ForeignKeys[0]: // board_stories
			values[i] = &sql.NullScanner{S: new(uuid.UUID)}
		default:
			return nil, fmt.Errorf("unexpected column %q for type Story", columns[i])
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Story fields.
func (s *Story) assignValues(columns []string, values []interface{}) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case story.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				s.ID = *value
			}
		case story.FieldCreateTime:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field create_time", values[i])
			} else if value.Valid {
				s.CreateTime = value.Time
			}
		case story.FieldUpdateTime:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field update_time", values[i])
			} else if value.Valid {
				s.UpdateTime = value.Time
			}
		case story.FieldTitle:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field title", values[i])
			} else if value.Valid {
				s.Title = value.String
			}
		case story.FieldDescription:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field description", values[i])
			} else if value.Valid {
				s.Description = value.String
			}
		case story.ForeignKeys[0]:
			if value, ok := values[i].(*sql.NullScanner); !ok {
				return fmt.Errorf("unexpected type %T for field board_stories", values[i])
			} else if value.Valid {
				s.board_stories = new(uuid.UUID)
				*s.board_stories = *value.S.(*uuid.UUID)
			}
		}
	}
	return nil
}

// QueryOwner queries the "owner" edge of the Story entity.
func (s *Story) QueryOwner() *BoardQuery {
	return (&StoryClient{config: s.config}).QueryOwner(s)
}

// Update returns a builder for updating this Story.
// Note that you need to call Story.Unwrap() before calling this method if this Story
// was returned from a transaction, and the transaction was committed or rolled back.
func (s *Story) Update() *StoryUpdateOne {
	return (&StoryClient{config: s.config}).UpdateOne(s)
}

// Unwrap unwraps the Story entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (s *Story) Unwrap() *Story {
	tx, ok := s.config.driver.(*txDriver)
	if !ok {
		panic("models: Story is not a transactional entity")
	}
	s.config.driver = tx.drv
	return s
}

// String implements the fmt.Stringer.
func (s *Story) String() string {
	var builder strings.Builder
	builder.WriteString("Story(")
	builder.WriteString(fmt.Sprintf("id=%v", s.ID))
	builder.WriteString(", create_time=")
	builder.WriteString(s.CreateTime.Format(time.ANSIC))
	builder.WriteString(", update_time=")
	builder.WriteString(s.UpdateTime.Format(time.ANSIC))
	builder.WriteString(", title=")
	builder.WriteString(s.Title)
	builder.WriteString(", description=")
	builder.WriteString(s.Description)
	builder.WriteByte(')')
	return builder.String()
}

// Stories is a parsable slice of Story.
type Stories []*Story

func (s Stories) config(cfg config) {
	for _i := range s {
		s[_i].config = cfg
	}
}
