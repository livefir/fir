package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/adnaan/fir"
	"github.com/google/uuid"
)

// Story holds the schema definition for the Story entity.
type Story struct {
	ent.Schema
}

func (Story) Annotations() []schema.Annotation {
	return []schema.Annotation{
		fir.CreateForm{
			Fields: []string{"title", "description"},
		},
		fir.UpdateForm{
			Fields: []string{"title", "description"},
		},
		fir.ListItem{
			Fields: []string{"title"},
		},
	}
}

func (Story) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

// Fields of the Story.
func (Story) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("title").MinLen(3).MaxLen(255).NotEmpty(),
		field.Text("description").MinLen(3).MaxLen(2550).NotEmpty(),
	}
}

// Edges of the Story.
func (Story) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", Board.Type).Ref("stories").Unique(),
	}
}
