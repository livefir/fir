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
	// TodosColumns holds the columns for the "todos" table.
	TodosColumns = []*schema.Column{
		{Name: "id", Type: field.TypeUUID},
		{Name: "create_time", Type: field.TypeTime},
		{Name: "update_time", Type: field.TypeTime},
		{Name: "title", Type: field.TypeString},
		{Name: "description", Type: field.TypeString, Size: 2147483647},
		{Name: "board_todos", Type: field.TypeUUID, Nullable: true},
	}
	// TodosTable holds the schema information for the "todos" table.
	TodosTable = &schema.Table{
		Name:       "todos",
		Columns:    TodosColumns,
		PrimaryKey: []*schema.Column{TodosColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "todos_boards_todos",
				Columns:    []*schema.Column{TodosColumns[5]},
				RefColumns: []*schema.Column{BoardsColumns[0]},
				OnDelete:   schema.SetNull,
			},
		},
	}
	// Tables holds all the tables in the schema.
	Tables = []*schema.Table{
		BoardsTable,
		TodosTable,
	}
)

func init() {
	TodosTable.ForeignKeys[0].RefTable = BoardsTable
}
