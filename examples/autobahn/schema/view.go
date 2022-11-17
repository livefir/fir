package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

// View holds the schema definition for the View entity.
type View struct {
	ent.Schema
}

func (View) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

// Fields of the View.
func (View) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("title"),
	}
}

// Edges of the View.
func (View) Edges() []ent.Edge {
	return nil
}
