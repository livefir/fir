package patch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"

	"github.com/golang/glog"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
)

// OpType is the type of patch operation
type OpType string

const (
	morph       OpType = "morph"
	after       OpType = "after"
	before      OpType = "before"
	appendOp    OpType = "append"
	prepend     OpType = "prepend"
	remove      OpType = "remove"
	reload      OpType = "reload"
	updateStore OpType = "store"
	resetForm   OpType = "resetForm"
	navigate    OpType = "navigate"
)

// Op is an interface for all patch operations
type Op struct {
	// Type is the type of patch operation
	Type OpType `json:"op"`
	// Selector is the css selector for the element to patch
	Selector *string `json:"selector,omitempty"`
	// Value is the value for the patch operation
	Value any `json:"value,omitempty"`
}

type Set []Op

func (pl *Set) Error() string {
	b, _ := json.Marshal(pl)
	return string(b)
}

type TemplateRenderer interface {
	Name() string
	Data() any
}

type templateRenderer struct {
	name string
	data any
}

func (t *templateRenderer) Name() string {
	return t.name
}

func (t *templateRenderer) Data() any {
	return t.data
}

func Template(name string, data any) TemplateRenderer {
	return &templateRenderer{name: name, data: data}
}
func Block(name string, data any) TemplateRenderer {
	return Template(name, data)
}

func HTML(html string) TemplateRenderer {
	return Template("_fir_html", html)
}

func Morph(selector string, t TemplateRenderer) Op {
	return Op{
		Type:     morph,
		Selector: &selector,
		Value:    map[string]any{"name": t.Name(), "data": t.Data()},
	}
}

func After(selector string, t TemplateRenderer) Op {
	return Op{
		Type:     after,
		Selector: &selector,
		Value:    map[string]any{"name": t.Name(), "data": t.Data()},
	}
}

func Before(selector string, t TemplateRenderer) Op {
	return Op{
		Type:     before,
		Selector: &selector,
		Value:    map[string]any{"name": t.Name(), "data": t.Data()},
	}
}

func Append(selector string, t TemplateRenderer) Op {
	return Op{
		Type:     appendOp,
		Selector: &selector,
		Value:    map[string]any{"name": t.Name(), "data": t.Data()},
	}
}

func Prepend(selector string, t TemplateRenderer) Op {
	return Op{
		Type:     prepend,
		Selector: &selector,
		Value:    map[string]any{"name": t.Name(), "data": t.Data()},
	}
}

func Remove(selector string) Op {
	return Op{
		Type:     remove,
		Selector: &selector,
	}
}

func Reload() Op {
	return Op{
		Type: reload,
	}
}

func Store(name string, data any) Op {
	return Op{
		Type:     updateStore,
		Selector: &name,
		Value:    data,
	}
}

func ResetForm(selector string) Op {
	return Op{
		Type:     resetForm,
		Selector: &selector,
	}
}

func Navigate(url string) Op {
	return Op{
		Type:  navigate,
		Value: url,
	}
}

func RenderJSON(t *template.Template, patchset []Op) []byte {
	var renderedPatchset []Op
	firErrorPatchExists := false
	for _, p := range patchset {
		switch p.Type {
		case updateStore, navigate, resetForm, reload, remove:
			renderedPatchset = append(renderedPatchset, p)
		case morph, after, before, appendOp, prepend:
			tmpl, ok := p.Value.(map[string]any)
			if !ok {
				glog.Errorf("[buildPatchOperations] invalid patch template data: %v", p.Value)
				continue
			}

			if *p.Selector == "#fir-error" {
				firErrorPatchExists = true
			}

			var err error
			p.Value, err = buildTemplateValue(t, tmpl["name"].(string), tmpl["data"])
			if err != nil {
				glog.Errorf("[warning]buildPatchOperations error: %v,%+v \n", err, tmpl)
				continue
			}

			renderedPatchset = append(renderedPatchset, p)
		default:
			continue
		}
	}

	if !firErrorPatchExists {
		// unset error patch
		firError := "#fir-error"
		tmplVal, err := buildTemplateValue(t, "fir-error", nil)
		if err == nil {
			renderedPatchset = append([]Op{{
				Type:     morph,
				Selector: &firError,
				Value:    tmplVal,
			}}, renderedPatchset...)
		}
	}

	if len(renderedPatchset) == 0 {
		return nil
	}

	data, err := json.Marshal(renderedPatchset)
	if err != nil {
		glog.Errorf("buildPatchOperations marshal error: %+v, %v \n", renderedPatchset, err)
		return nil
	}
	return data
}

func buildTemplateValue(t *template.Template, name string, data any) (string, error) {
	var buf bytes.Buffer
	defer buf.Reset()
	if name == "_fir_html" {
		buf.WriteString(data.(string))
	} else {
		t.Option("missingkey=zero")
		err := t.ExecuteTemplate(&buf, name, data)
		if err != nil {
			return "", err
		}
	}

	m := minify.New()
	m.Add("text/html", &html.Minifier{})
	r := m.Reader("text/html", &buf)
	var buf1 bytes.Buffer
	defer buf1.Reset()
	_, err := io.Copy(&buf1, r)
	if err != nil {
		return "", err
	}
	value := buf1.String()
	return value, nil
}

func MorphError(name string) (func(err error) Op, func() Op) {
	selector := fmt.Sprintf("#%s", name)
	return func(err error) Op {
			return Morph(selector, Block(name, map[string]any{name: err}))
		}, func() Op {
			return Morph(selector, Block(name, map[string]any{name: ""}))
		}
}
