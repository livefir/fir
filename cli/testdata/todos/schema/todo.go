package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/adnaan/fir"
	"github.com/google/uuid"
)

// Todo holds the schema definition for the Todo entity.
type Todo struct {
	ent.Schema
}

func (Todo) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

func (Todo) Annotations() []schema.Annotation {
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

// Fields of the Todo.
func (Todo) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("title").Validate(fir.MinMax(3, 140)),
		field.Text("description").Validate(fir.MinMax(3, 280)),
	}
}

// Edges of the Todo.
func (Todo) Edges() []ent.Edge {
	return nil
}
