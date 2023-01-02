package dom

// NewTemplateRenderer creates a new template renderer
func NewTemplateRenderer(name string, data any) Renderer {
	return &TemplateRenderer{
		name: name,
		data: data,
	}
}

// NewJSONRenderer creates a new json renderer
func NewJSONRenderer(data any) Renderer {
	return &JSONRenderer{
		data: data,
	}
}

// Renderer is an interface for rendering partial templates
type Renderer interface {
	Name() string
	Data() any
}

type TemplateRenderer struct {
	name string
	data any
}

// Name of the template
func (t *TemplateRenderer) Name() string {
	return t.name
}

// Data to be hydrated into the template
func (t *TemplateRenderer) Data() any {
	return t.data
}

type JSONRenderer struct {
	name string
	data any
}

// Name of the template
func (t *JSONRenderer) Name() string {
	return t.name
}

// Data to be hydrated into the template
func (t *JSONRenderer) Data() any {
	return t.data
}
