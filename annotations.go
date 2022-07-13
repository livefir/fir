package fir

type CreateForm struct {
	Fields []string
}

func (CreateForm) Name() string {
	return "CreateForm"
}

type UpdateForm struct {
	Fields []string
}

func (UpdateForm) Name() string {
	return "UpdateForm"
}

type ListItem struct {
	Fields []string
}

func (ListItem) Name() string {
	return "ListItem"
}
