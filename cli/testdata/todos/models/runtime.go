// Code generated (@generated) by entc, DO NOT EDIT.

package models

import (
	"time"

	"github.com/adnaan/fir/cli/testdata/todos/models/board"
	"github.com/adnaan/fir/cli/testdata/todos/models/todo"
	"github.com/adnaan/fir/cli/testdata/todos/schema"
	"github.com/google/uuid"
)

// The init function reads all schema descriptors with runtime code
// (default values, validators, hooks and policies) and stitches it
// to their package variables.
func init() {
	boardMixin := schema.Board{}.Mixin()
	boardMixinFields0 := boardMixin[0].Fields()
	_ = boardMixinFields0
	boardFields := schema.Board{}.Fields()
	_ = boardFields
	// boardDescCreateTime is the schema descriptor for create_time field.
	boardDescCreateTime := boardMixinFields0[0].Descriptor()
	// board.DefaultCreateTime holds the default value on creation for the create_time field.
	board.DefaultCreateTime = boardDescCreateTime.Default.(func() time.Time)
	// boardDescUpdateTime is the schema descriptor for update_time field.
	boardDescUpdateTime := boardMixinFields0[1].Descriptor()
	// board.DefaultUpdateTime holds the default value on creation for the update_time field.
	board.DefaultUpdateTime = boardDescUpdateTime.Default.(func() time.Time)
	// board.UpdateDefaultUpdateTime holds the default value on update for the update_time field.
	board.UpdateDefaultUpdateTime = boardDescUpdateTime.UpdateDefault.(func() time.Time)
	// boardDescTitle is the schema descriptor for title field.
	boardDescTitle := boardFields[1].Descriptor()
	// board.TitleValidator is a validator for the "title" field. It is called by the builders before save.
	board.TitleValidator = boardDescTitle.Validators[0].(func(string) error)
	// boardDescDescription is the schema descriptor for description field.
	boardDescDescription := boardFields[2].Descriptor()
	// board.DescriptionValidator is a validator for the "description" field. It is called by the builders before save.
	board.DescriptionValidator = boardDescDescription.Validators[0].(func(string) error)
	// boardDescID is the schema descriptor for id field.
	boardDescID := boardFields[0].Descriptor()
	// board.DefaultID holds the default value on creation for the id field.
	board.DefaultID = boardDescID.Default.(func() uuid.UUID)
	todoMixin := schema.Todo{}.Mixin()
	todoMixinFields0 := todoMixin[0].Fields()
	_ = todoMixinFields0
	todoFields := schema.Todo{}.Fields()
	_ = todoFields
	// todoDescCreateTime is the schema descriptor for create_time field.
	todoDescCreateTime := todoMixinFields0[0].Descriptor()
	// todo.DefaultCreateTime holds the default value on creation for the create_time field.
	todo.DefaultCreateTime = todoDescCreateTime.Default.(func() time.Time)
	// todoDescUpdateTime is the schema descriptor for update_time field.
	todoDescUpdateTime := todoMixinFields0[1].Descriptor()
	// todo.DefaultUpdateTime holds the default value on creation for the update_time field.
	todo.DefaultUpdateTime = todoDescUpdateTime.Default.(func() time.Time)
	// todo.UpdateDefaultUpdateTime holds the default value on update for the update_time field.
	todo.UpdateDefaultUpdateTime = todoDescUpdateTime.UpdateDefault.(func() time.Time)
	// todoDescTitle is the schema descriptor for title field.
	todoDescTitle := todoFields[1].Descriptor()
	// todo.TitleValidator is a validator for the "title" field. It is called by the builders before save.
	todo.TitleValidator = todoDescTitle.Validators[0].(func(string) error)
	// todoDescDescription is the schema descriptor for description field.
	todoDescDescription := todoFields[2].Descriptor()
	// todo.DescriptionValidator is a validator for the "description" field. It is called by the builders before save.
	todo.DescriptionValidator = todoDescDescription.Validators[0].(func(string) error)
	// todoDescID is the schema descriptor for id field.
	todoDescID := todoFields[0].Descriptor()
	// todo.DefaultID holds the default value on creation for the id field.
	todo.DefaultID = todoDescID.Default.(func() uuid.UUID)
}
