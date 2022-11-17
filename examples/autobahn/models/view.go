// Code generated (@generated) by entc, DO NOT EDIT.

package models

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/adnaan/autobahn/models/view"
	"github.com/google/uuid"
)

// View is the model entity for the View schema.
type View struct {
	config `json:"-"`
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// CreateTime holds the value of the "create_time" field.
	CreateTime time.Time `json:"create_time,omitempty"`
	// UpdateTime holds the value of the "update_time" field.
	UpdateTime time.Time `json:"update_time,omitempty"`
	// Title holds the value of the "title" field.
	Title string `json:"title,omitempty"`
}

// scanValues returns the types for scanning values from sql.Rows.
func (*View) scanValues(columns []string) ([]interface{}, error) {
	values := make([]interface{}, len(columns))
	for i := range columns {
		switch columns[i] {
		case view.FieldTitle:
			values[i] = new(sql.NullString)
		case view.FieldCreateTime, view.FieldUpdateTime:
			values[i] = new(sql.NullTime)
		case view.FieldID:
			values[i] = new(uuid.UUID)
		default:
			return nil, fmt.Errorf("unexpected column %q for type View", columns[i])
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the View fields.
func (v *View) assignValues(columns []string, values []interface{}) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case view.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				v.ID = *value
			}
		case view.FieldCreateTime:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field create_time", values[i])
			} else if value.Valid {
				v.CreateTime = value.Time
			}
		case view.FieldUpdateTime:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field update_time", values[i])
			} else if value.Valid {
				v.UpdateTime = value.Time
			}
		case view.FieldTitle:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field title", values[i])
			} else if value.Valid {
				v.Title = value.String
			}
		}
	}
	return nil
}

// Update returns a builder for updating this View.
// Note that you need to call View.Unwrap() before calling this method if this View
// was returned from a transaction, and the transaction was committed or rolled back.
func (v *View) Update() *ViewUpdateOne {
	return (&ViewClient{config: v.config}).UpdateOne(v)
}

// Unwrap unwraps the View entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (v *View) Unwrap() *View {
	tx, ok := v.config.driver.(*txDriver)
	if !ok {
		panic("models: View is not a transactional entity")
	}
	v.config.driver = tx.drv
	return v
}

// String implements the fmt.Stringer.
func (v *View) String() string {
	var builder strings.Builder
	builder.WriteString("View(")
	builder.WriteString(fmt.Sprintf("id=%v", v.ID))
	builder.WriteString(", create_time=")
	builder.WriteString(v.CreateTime.Format(time.ANSIC))
	builder.WriteString(", update_time=")
	builder.WriteString(v.UpdateTime.Format(time.ANSIC))
	builder.WriteString(", title=")
	builder.WriteString(v.Title)
	builder.WriteByte(')')
	return builder.String()
}

// Views is a parsable slice of View.
type Views []*View

func (v Views) config(cfg config) {
	for _i := range v {
		v[_i].config = cfg
	}
}
