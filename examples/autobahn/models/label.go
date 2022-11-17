// Code generated (@generated) by entc, DO NOT EDIT.

package models

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/adnaan/autobahn/models/label"
	"github.com/google/uuid"
)

// Label is the model entity for the Label schema.
type Label struct {
	config `json:"-"`
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// CreateTime holds the value of the "create_time" field.
	CreateTime time.Time `json:"create_time,omitempty"`
	// UpdateTime holds the value of the "update_time" field.
	UpdateTime time.Time `json:"update_time,omitempty"`
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Label) scanValues(columns []string) ([]interface{}, error) {
	values := make([]interface{}, len(columns))
	for i := range columns {
		switch columns[i] {
		case label.FieldCreateTime, label.FieldUpdateTime:
			values[i] = new(sql.NullTime)
		case label.FieldID:
			values[i] = new(uuid.UUID)
		default:
			return nil, fmt.Errorf("unexpected column %q for type Label", columns[i])
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Label fields.
func (l *Label) assignValues(columns []string, values []interface{}) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case label.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				l.ID = *value
			}
		case label.FieldCreateTime:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field create_time", values[i])
			} else if value.Valid {
				l.CreateTime = value.Time
			}
		case label.FieldUpdateTime:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field update_time", values[i])
			} else if value.Valid {
				l.UpdateTime = value.Time
			}
		}
	}
	return nil
}

// Update returns a builder for updating this Label.
// Note that you need to call Label.Unwrap() before calling this method if this Label
// was returned from a transaction, and the transaction was committed or rolled back.
func (l *Label) Update() *LabelUpdateOne {
	return (&LabelClient{config: l.config}).UpdateOne(l)
}

// Unwrap unwraps the Label entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (l *Label) Unwrap() *Label {
	tx, ok := l.config.driver.(*txDriver)
	if !ok {
		panic("models: Label is not a transactional entity")
	}
	l.config.driver = tx.drv
	return l
}

// String implements the fmt.Stringer.
func (l *Label) String() string {
	var builder strings.Builder
	builder.WriteString("Label(")
	builder.WriteString(fmt.Sprintf("id=%v", l.ID))
	builder.WriteString(", create_time=")
	builder.WriteString(l.CreateTime.Format(time.ANSIC))
	builder.WriteString(", update_time=")
	builder.WriteString(l.UpdateTime.Format(time.ANSIC))
	builder.WriteByte(')')
	return builder.String()
}

// Labels is a parsable slice of Label.
type Labels []*Label

func (l Labels) config(cfg config) {
	for _i := range l {
		l[_i].config = cfg
	}
}
