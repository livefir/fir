// Code generated (@generated) by entc, DO NOT EDIT.

package board

import (
	"time"

	"github.com/google/uuid"
)

const (
	// Label holds the string label denoting the board type in the database.
	Label = "board"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldCreateTime holds the string denoting the create_time field in the database.
	FieldCreateTime = "create_time"
	// FieldUpdateTime holds the string denoting the update_time field in the database.
	FieldUpdateTime = "update_time"
	// FieldTitle holds the string denoting the title field in the database.
	FieldTitle = "title"
	// FieldDescription holds the string denoting the description field in the database.
	FieldDescription = "description"
	// EdgeStories holds the string denoting the stories edge name in mutations.
	EdgeStories = "stories"
	// Table holds the table name of the board in the database.
	Table = "boards"
	// StoriesTable is the table that holds the stories relation/edge.
	StoriesTable = "stories"
	// StoriesInverseTable is the table name for the Story entity.
	// It exists in this package in order to avoid circular dependency with the "story" package.
	StoriesInverseTable = "stories"
	// StoriesColumn is the table column denoting the stories relation/edge.
	StoriesColumn = "board_stories"
)

// Columns holds all SQL columns for board fields.
var Columns = []string{
	FieldID,
	FieldCreateTime,
	FieldUpdateTime,
	FieldTitle,
	FieldDescription,
}

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	return false
}

var (
	// DefaultCreateTime holds the default value on creation for the "create_time" field.
	DefaultCreateTime func() time.Time
	// DefaultUpdateTime holds the default value on creation for the "update_time" field.
	DefaultUpdateTime func() time.Time
	// UpdateDefaultUpdateTime holds the default value on update for the "update_time" field.
	UpdateDefaultUpdateTime func() time.Time
	// TitleValidator is a validator for the "title" field. It is called by the builders before save.
	TitleValidator func(string) error
	// DescriptionValidator is a validator for the "description" field. It is called by the builders before save.
	DescriptionValidator func(string) error
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() uuid.UUID
)
