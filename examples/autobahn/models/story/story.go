// Code generated (@generated) by entc, DO NOT EDIT.

package story

import (
	"time"

	"github.com/google/uuid"
)

const (
	// Label holds the string label denoting the story type in the database.
	Label = "story"
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
	// EdgeOwner holds the string denoting the owner edge name in mutations.
	EdgeOwner = "owner"
	// Table holds the table name of the story in the database.
	Table = "stories"
	// OwnerTable is the table that holds the owner relation/edge.
	OwnerTable = "stories"
	// OwnerInverseTable is the table name for the Board entity.
	// It exists in this package in order to avoid circular dependency with the "board" package.
	OwnerInverseTable = "boards"
	// OwnerColumn is the table column denoting the owner relation/edge.
	OwnerColumn = "board_stories"
)

// Columns holds all SQL columns for story fields.
var Columns = []string{
	FieldID,
	FieldCreateTime,
	FieldUpdateTime,
	FieldTitle,
	FieldDescription,
}

// ForeignKeys holds the SQL foreign-keys that are owned by the "stories"
// table and are not defined as standalone fields in the schema.
var ForeignKeys = []string{
	"board_stories",
}

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	for i := range ForeignKeys {
		if column == ForeignKeys[i] {
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
