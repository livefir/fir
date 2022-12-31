package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

// Issue holds the schema definition for the Issue entity.
type Issue struct {
	ent.Schema
}

func (Issue) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

// Fields of the Issue.
func (Issue) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("title").MinLen(3).MaxLen(140),
		field.Text("description").MinLen(3).MaxLen(280),
	}
}

// Edges of the Issue.
func (Issue) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", Project.Type).Ref("issues").Unique(),
	}
}
