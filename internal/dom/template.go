package dom

// NewTemplateRenderer creates a new template renderer
func NewTemplateRenderer(name string, data any) TemplateRenderer {
	return &templateRenderer{
		name: name,
		data: data,
	}
}

// TemplateRenderer is an interface for rendering partial templates
type TemplateRenderer interface {
	Name() string
	Data() any
}

type templateRenderer struct {
	name string
	data any
}

// Name of the template
func (t *templateRenderer) Name() string {
	return t.name
}

// Data to be hydrated into the template
func (t *templateRenderer) Data() any {
	return t.data
}
