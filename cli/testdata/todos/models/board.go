// Code generated (@generated) by entc, DO NOT EDIT.

package models

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/adnaan/fir/cli/testdata/todos/models/board"
	"github.com/google/uuid"
)

// Board is the model entity for the Board schema.
type Board struct {
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
	// The values are being populated by the BoardQuery when eager-loading is set.
	Edges BoardEdges `json:"edges"`
}

// BoardEdges holds the relations/edges for other nodes in the graph.
type BoardEdges struct {
	// Todos holds the value of the todos edge.
	Todos []*Todo `json:"todos,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [1]bool
}

// TodosOrErr returns the Todos value or an error if the edge
// was not loaded in eager-loading.
func (e BoardEdges) TodosOrErr() ([]*Todo, error) {
	if e.loadedTypes[0] {
		return e.Todos, nil
	}
	return nil, &NotLoadedError{edge: "todos"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Board) scanValues(columns []string) ([]interface{}, error) {
	values := make([]interface{}, len(columns))
	for i := range columns {
		switch columns[i] {
		case board.FieldTitle, board.FieldDescription:
			values[i] = new(sql.NullString)
		case board.FieldCreateTime, board.FieldUpdateTime:
			values[i] = new(sql.NullTime)
		case board.FieldID:
			values[i] = new(uuid.UUID)
		default:
			return nil, fmt.Errorf("unexpected column %q for type Board", columns[i])
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Board fields.
func (b *Board) assignValues(columns []string, values []interface{}) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case board.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				b.ID = *value
			}
		case board.FieldCreateTime:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field create_time", values[i])
			} else if value.Valid {
				b.CreateTime = value.Time
			}
		case board.FieldUpdateTime:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field update_time", values[i])
			} else if value.Valid {
				b.UpdateTime = value.Time
			}
		case board.FieldTitle:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field title", values[i])
			} else if value.Valid {
				b.Title = value.String
			}
		case board.FieldDescription:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field description", values[i])
			} else if value.Valid {
				b.Description = value.String
			}
		}
	}
	return nil
}

// QueryTodos queries the "todos" edge of the Board entity.
func (b *Board) QueryTodos() *TodoQuery {
	return (&BoardClient{config: b.config}).QueryTodos(b)
}

// Update returns a builder for updating this Board.
// Note that you need to call Board.Unwrap() before calling this method if this Board
// was returned from a transaction, and the transaction was committed or rolled back.
func (b *Board) Update() *BoardUpdateOne {
	return (&BoardClient{config: b.config}).UpdateOne(b)
}

// Unwrap unwraps the Board entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (b *Board) Unwrap() *Board {
	tx, ok := b.config.driver.(*txDriver)
	if !ok {
		panic("models: Board is not a transactional entity")
	}
	b.config.driver = tx.drv
	return b
}

// String implements the fmt.Stringer.
func (b *Board) String() string {
	var builder strings.Builder
	builder.WriteString("Board(")
	builder.WriteString(fmt.Sprintf("id=%v", b.ID))
	builder.WriteString(", create_time=")
	builder.WriteString(b.CreateTime.Format(time.ANSIC))
	builder.WriteString(", update_time=")
	builder.WriteString(b.UpdateTime.Format(time.ANSIC))
	builder.WriteString(", title=")
	builder.WriteString(b.Title)
	builder.WriteString(", description=")
	builder.WriteString(b.Description)
	builder.WriteByte(')')
	return builder.String()
}

// Boards is a parsable slice of Board.
type Boards []*Board

func (b Boards) config(cfg config) {
	for _i := range b {
		b[_i].config = cfg
	}
}
