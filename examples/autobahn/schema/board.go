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

// Board holds the schema definition for the Board entity.
type Board struct {
	ent.Schema
}

func (Board) Annotations() []schema.Annotation {
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

func (Board) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

// Fields of the Board.
func (Board) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("title").Validate(fir.MinMax(3, 140)),
		field.Text("description").Validate(fir.MinMax(3, 280)),
	}
}

// Edges of the Board.
func (Board) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("stories", Story.Type),
	}
}
