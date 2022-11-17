// Code generated (@generated) by entc, DO NOT EDIT.

package models

import (
	"time"

	"github.com/adnaan/autobahn/models/board"
	"github.com/adnaan/autobahn/models/comment"
	"github.com/adnaan/autobahn/models/label"
	"github.com/adnaan/autobahn/models/story"
	"github.com/adnaan/autobahn/models/view"
	"github.com/adnaan/autobahn/schema"
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
	commentMixin := schema.Comment{}.Mixin()
	commentMixinFields0 := commentMixin[0].Fields()
	_ = commentMixinFields0
	commentFields := schema.Comment{}.Fields()
	_ = commentFields
	// commentDescCreateTime is the schema descriptor for create_time field.
	commentDescCreateTime := commentMixinFields0[0].Descriptor()
	// comment.DefaultCreateTime holds the default value on creation for the create_time field.
	comment.DefaultCreateTime = commentDescCreateTime.Default.(func() time.Time)
	// commentDescUpdateTime is the schema descriptor for update_time field.
	commentDescUpdateTime := commentMixinFields0[1].Descriptor()
	// comment.DefaultUpdateTime holds the default value on creation for the update_time field.
	comment.DefaultUpdateTime = commentDescUpdateTime.Default.(func() time.Time)
	// comment.UpdateDefaultUpdateTime holds the default value on update for the update_time field.
	comment.UpdateDefaultUpdateTime = commentDescUpdateTime.UpdateDefault.(func() time.Time)
	// commentDescID is the schema descriptor for id field.
	commentDescID := commentFields[0].Descriptor()
	// comment.DefaultID holds the default value on creation for the id field.
	comment.DefaultID = commentDescID.Default.(func() uuid.UUID)
	labelMixin := schema.Label{}.Mixin()
	labelMixinFields0 := labelMixin[0].Fields()
	_ = labelMixinFields0
	labelFields := schema.Label{}.Fields()
	_ = labelFields
	// labelDescCreateTime is the schema descriptor for create_time field.
	labelDescCreateTime := labelMixinFields0[0].Descriptor()
	// label.DefaultCreateTime holds the default value on creation for the create_time field.
	label.DefaultCreateTime = labelDescCreateTime.Default.(func() time.Time)
	// labelDescUpdateTime is the schema descriptor for update_time field.
	labelDescUpdateTime := labelMixinFields0[1].Descriptor()
	// label.DefaultUpdateTime holds the default value on creation for the update_time field.
	label.DefaultUpdateTime = labelDescUpdateTime.Default.(func() time.Time)
	// label.UpdateDefaultUpdateTime holds the default value on update for the update_time field.
	label.UpdateDefaultUpdateTime = labelDescUpdateTime.UpdateDefault.(func() time.Time)
	// labelDescID is the schema descriptor for id field.
	labelDescID := labelFields[0].Descriptor()
	// label.DefaultID holds the default value on creation for the id field.
	label.DefaultID = labelDescID.Default.(func() uuid.UUID)
	storyMixin := schema.Story{}.Mixin()
	storyMixinFields0 := storyMixin[0].Fields()
	_ = storyMixinFields0
	storyFields := schema.Story{}.Fields()
	_ = storyFields
	// storyDescCreateTime is the schema descriptor for create_time field.
	storyDescCreateTime := storyMixinFields0[0].Descriptor()
	// story.DefaultCreateTime holds the default value on creation for the create_time field.
	story.DefaultCreateTime = storyDescCreateTime.Default.(func() time.Time)
	// storyDescUpdateTime is the schema descriptor for update_time field.
	storyDescUpdateTime := storyMixinFields0[1].Descriptor()
	// story.DefaultUpdateTime holds the default value on creation for the update_time field.
	story.DefaultUpdateTime = storyDescUpdateTime.Default.(func() time.Time)
	// story.UpdateDefaultUpdateTime holds the default value on update for the update_time field.
	story.UpdateDefaultUpdateTime = storyDescUpdateTime.UpdateDefault.(func() time.Time)
	// storyDescTitle is the schema descriptor for title field.
	storyDescTitle := storyFields[1].Descriptor()
	// story.TitleValidator is a validator for the "title" field. It is called by the builders before save.
	story.TitleValidator = func() func(string) error {
		validators := storyDescTitle.Validators
		fns := [...]func(string) error{
			validators[0].(func(string) error),
			validators[1].(func(string) error),
			validators[2].(func(string) error),
		}
		return func(title string) error {
			for _, fn := range fns {
				if err := fn(title); err != nil {
					return err
				}
			}
			return nil
		}
	}()
	// storyDescDescription is the schema descriptor for description field.
	storyDescDescription := storyFields[2].Descriptor()
	// story.DescriptionValidator is a validator for the "description" field. It is called by the builders before save.
	story.DescriptionValidator = func() func(string) error {
		validators := storyDescDescription.Validators
		fns := [...]func(string) error{
			validators[0].(func(string) error),
			validators[1].(func(string) error),
			validators[2].(func(string) error),
		}
		return func(description string) error {
			for _, fn := range fns {
				if err := fn(description); err != nil {
					return err
				}
			}
			return nil
		}
	}()
	// storyDescID is the schema descriptor for id field.
	storyDescID := storyFields[0].Descriptor()
	// story.DefaultID holds the default value on creation for the id field.
	story.DefaultID = storyDescID.Default.(func() uuid.UUID)
	viewMixin := schema.View{}.Mixin()
	viewMixinFields0 := viewMixin[0].Fields()
	_ = viewMixinFields0
	viewFields := schema.View{}.Fields()
	_ = viewFields
	// viewDescCreateTime is the schema descriptor for create_time field.
	viewDescCreateTime := viewMixinFields0[0].Descriptor()
	// view.DefaultCreateTime holds the default value on creation for the create_time field.
	view.DefaultCreateTime = viewDescCreateTime.Default.(func() time.Time)
	// viewDescUpdateTime is the schema descriptor for update_time field.
	viewDescUpdateTime := viewMixinFields0[1].Descriptor()
	// view.DefaultUpdateTime holds the default value on creation for the update_time field.
	view.DefaultUpdateTime = viewDescUpdateTime.Default.(func() time.Time)
	// view.UpdateDefaultUpdateTime holds the default value on update for the update_time field.
	view.UpdateDefaultUpdateTime = viewDescUpdateTime.UpdateDefault.(func() time.Time)
	// viewDescID is the schema descriptor for id field.
	viewDescID := viewFields[0].Descriptor()
	// view.DefaultID holds the default value on creation for the id field.
	view.DefaultID = viewDescID.Default.(func() uuid.UUID)
}
