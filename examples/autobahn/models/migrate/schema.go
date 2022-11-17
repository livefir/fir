// Code generated (@generated) by entc, DO NOT EDIT.

package migrate

import (
	"entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/schema/field"
)

var (
	// BoardsColumns holds the columns for the "boards" table.
	BoardsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeUUID},
		{Name: "create_time", Type: field.TypeTime},
		{Name: "update_time", Type: field.TypeTime},
		{Name: "title", Type: field.TypeString},
		{Name: "description", Type: field.TypeString, Size: 2147483647},
	}
	// BoardsTable holds the schema information for the "boards" table.
	BoardsTable = &schema.Table{
		Name:       "boards",
		Columns:    BoardsColumns,
		PrimaryKey: []*schema.Column{BoardsColumns[0]},
	}
	// CommentsColumns holds the columns for the "comments" table.
	CommentsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeUUID},
		{Name: "create_time", Type: field.TypeTime},
		{Name: "update_time", Type: field.TypeTime},
	}
	// CommentsTable holds the schema information for the "comments" table.
	CommentsTable = &schema.Table{
		Name:       "comments",
		Columns:    CommentsColumns,
		PrimaryKey: []*schema.Column{CommentsColumns[0]},
	}
	// LabelsColumns holds the columns for the "labels" table.
	LabelsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeUUID},
		{Name: "create_time", Type: field.TypeTime},
		{Name: "update_time", Type: field.TypeTime},
	}
	// LabelsTable holds the schema information for the "labels" table.
	LabelsTable = &schema.Table{
		Name:       "labels",
		Columns:    LabelsColumns,
		PrimaryKey: []*schema.Column{LabelsColumns[0]},
	}
	// StoriesColumns holds the columns for the "stories" table.
	StoriesColumns = []*schema.Column{
		{Name: "id", Type: field.TypeUUID},
		{Name: "create_time", Type: field.TypeTime},
		{Name: "update_time", Type: field.TypeTime},
		{Name: "title", Type: field.TypeString, Size: 255},
		{Name: "description", Type: field.TypeString, Size: 2550},
		{Name: "board_stories", Type: field.TypeUUID, Nullable: true},
	}
	// StoriesTable holds the schema information for the "stories" table.
	StoriesTable = &schema.Table{
		Name:       "stories",
		Columns:    StoriesColumns,
		PrimaryKey: []*schema.Column{StoriesColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "stories_boards_stories",
				Columns:    []*schema.Column{StoriesColumns[5]},
				RefColumns: []*schema.Column{BoardsColumns[0]},
				OnDelete:   schema.SetNull,
			},
		},
	}
	// ViewsColumns holds the columns for the "views" table.
	ViewsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeUUID},
		{Name: "create_time", Type: field.TypeTime},
		{Name: "update_time", Type: field.TypeTime},
		{Name: "title", Type: field.TypeString},
	}
	// ViewsTable holds the schema information for the "views" table.
	ViewsTable = &schema.Table{
		Name:       "views",
		Columns:    ViewsColumns,
		PrimaryKey: []*schema.Column{ViewsColumns[0]},
	}
	// Tables holds all the tables in the schema.
	Tables = []*schema.Table{
		BoardsTable,
		CommentsTable,
		LabelsTable,
		StoriesTable,
		ViewsTable,
	}
)

func init() {
	StoriesTable.ForeignKeys[0].RefTable = BoardsTable
}
