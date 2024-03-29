// Code generated by ent, DO NOT EDIT.

package ent

import (
	"time"

	"github.com/google/uuid"
	"github.com/livefir/fir/examples/fira/ent/issue"
	"github.com/livefir/fir/examples/fira/ent/project"
	"github.com/livefir/fir/examples/fira/ent/schema"
)

// The init function reads all schema descriptors with runtime code
// (default values, validators, hooks and policies) and stitches it
// to their package variables.
func init() {
	issueMixin := schema.Issue{}.Mixin()
	issueMixinFields0 := issueMixin[0].Fields()
	_ = issueMixinFields0
	issueFields := schema.Issue{}.Fields()
	_ = issueFields
	// issueDescCreateTime is the schema descriptor for create_time field.
	issueDescCreateTime := issueMixinFields0[0].Descriptor()
	// issue.DefaultCreateTime holds the default value on creation for the create_time field.
	issue.DefaultCreateTime = issueDescCreateTime.Default.(func() time.Time)
	// issueDescUpdateTime is the schema descriptor for update_time field.
	issueDescUpdateTime := issueMixinFields0[1].Descriptor()
	// issue.DefaultUpdateTime holds the default value on creation for the update_time field.
	issue.DefaultUpdateTime = issueDescUpdateTime.Default.(func() time.Time)
	// issue.UpdateDefaultUpdateTime holds the default value on update for the update_time field.
	issue.UpdateDefaultUpdateTime = issueDescUpdateTime.UpdateDefault.(func() time.Time)
	// issueDescTitle is the schema descriptor for title field.
	issueDescTitle := issueFields[1].Descriptor()
	// issue.TitleValidator is a validator for the "title" field. It is called by the builders before save.
	issue.TitleValidator = issueDescTitle.Validators[0].(func(string) error)
	// issueDescDescription is the schema descriptor for description field.
	issueDescDescription := issueFields[2].Descriptor()
	// issue.DescriptionValidator is a validator for the "description" field. It is called by the builders before save.
	issue.DescriptionValidator = issueDescDescription.Validators[0].(func(string) error)
	// issueDescID is the schema descriptor for id field.
	issueDescID := issueFields[0].Descriptor()
	// issue.DefaultID holds the default value on creation for the id field.
	issue.DefaultID = issueDescID.Default.(func() uuid.UUID)
	projectMixin := schema.Project{}.Mixin()
	projectMixinFields0 := projectMixin[0].Fields()
	_ = projectMixinFields0
	projectFields := schema.Project{}.Fields()
	_ = projectFields
	// projectDescCreateTime is the schema descriptor for create_time field.
	projectDescCreateTime := projectMixinFields0[0].Descriptor()
	// project.DefaultCreateTime holds the default value on creation for the create_time field.
	project.DefaultCreateTime = projectDescCreateTime.Default.(func() time.Time)
	// projectDescUpdateTime is the schema descriptor for update_time field.
	projectDescUpdateTime := projectMixinFields0[1].Descriptor()
	// project.DefaultUpdateTime holds the default value on creation for the update_time field.
	project.DefaultUpdateTime = projectDescUpdateTime.Default.(func() time.Time)
	// project.UpdateDefaultUpdateTime holds the default value on update for the update_time field.
	project.UpdateDefaultUpdateTime = projectDescUpdateTime.UpdateDefault.(func() time.Time)
	// projectDescTitle is the schema descriptor for title field.
	projectDescTitle := projectFields[1].Descriptor()
	// project.TitleValidator is a validator for the "title" field. It is called by the builders before save.
	project.TitleValidator = projectDescTitle.Validators[0].(func(string) error)
	// projectDescDescription is the schema descriptor for description field.
	projectDescDescription := projectFields[2].Descriptor()
	// project.DescriptionValidator is a validator for the "description" field. It is called by the builders before save.
	project.DescriptionValidator = projectDescDescription.Validators[0].(func(string) error)
	// projectDescID is the schema descriptor for id field.
	projectDescID := projectFields[0].Descriptor()
	// project.DefaultID holds the default value on creation for the id field.
	project.DefaultID = projectDescID.Default.(func() uuid.UUID)
}
