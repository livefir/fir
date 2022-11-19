package fir

// CreateForm can be used to annotate an entgo schema field as input in a create form.
type CreateForm struct {
	Fields []string
}

func (CreateForm) Name() string {
	return "CreateForm"
}

// UpdateForm can be used to annotate an entgo schema field as input in an update form.
type UpdateForm struct {
	Fields []string
}

func (UpdateForm) Name() string {
	return "UpdateForm"
}

// ListItem can be used to annotate an entgo schema field as a member of a list item.
type ListItem struct {
	Fields []string
}

func (ListItem) Name() string {
	return "ListItem"
}
