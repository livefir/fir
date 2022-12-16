package dom

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io"

	"github.com/golang/glog"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tidwall/match"
)

// NewPatcher creates a new dom patcher
func NewPatcher() Patcher {
	return &patcher{
		patchset: Patchset{},
	}
}

type Patcher interface {
	// ReplaceEl patches the dom at the given selector with the rendered template
	ReplaceEl(selector string, t TemplateRenderer) Patcher
	// After patches the dom after the given selector with the rendered template
	AfterEl(selector string, t TemplateRenderer) Patcher
	// Before patches the dom before the given selector with the rendered template
	BeforeEl(selector string, t TemplateRenderer) Patcher
	// Append patches the dom after the given selector with the rendered template
	AppendEl(selector string, t TemplateRenderer) Patcher
	// Prepend patches the dom before the given selector with the rendered template
	PrependEl(selector string, t TemplateRenderer) Patcher
	// Remove patches the dom to remove the given selector
	RemoveEl(selector string) Patcher
	// Reload patches the dom to reload the page
	Reload() Patcher
	// Store patches the dom to update the alpinejs store
	Store(name string, data any) Patcher
	// ResetForm patches the dom to reset the form
	ResetForm(selector string) Patcher
	// Navigate patches the dom to navigate to the given url
	Navigate(url string) Patcher
	// Patchset returns the patchset
	Patchset() Patchset
	// Error satisfies the error interface so that Context can return a Patcher
	Error() string
}

var _ error = (Patcher)(nil)

// PatchType is the type of patch operation
type PatchType string

const (
	Replace   PatchType = "replace"
	After     PatchType = "after"
	Before    PatchType = "before"
	Append    PatchType = "append"
	Prepend   PatchType = "prepend"
	Remove    PatchType = "remove"
	Reload    PatchType = "reload"
	Store     PatchType = "store"
	ResetForm PatchType = "resetForm"
	Navigate  PatchType = "navigate"
)

// Patch is an interface for all patch operations
type Patch struct {
	// Type is the type of patch operation
	Type PatchType `json:"op"`
	// Selector is the css selector for the element to patch
	Selector *string `json:"selector,omitempty"`
	// Value is the value for the patch operation
	Value any `json:"value,omitempty"`
}

// Patchset is a collection of patch operations
type Patchset []Patch

type patcher struct {
	patchset Patchset
}

func (p *patcher) ReplaceEl(selector string, t TemplateRenderer) Patcher {
	p.patchset = append(p.patchset, Patch{
		Type:     Replace,
		Selector: &selector,
		Value:    map[string]any{"name": t.Name(), "data": t.Data()},
	})
	return p
}

func (p *patcher) AfterEl(selector string, t TemplateRenderer) Patcher {
	p.patchset = append(p.patchset, Patch{
		Type:     After,
		Selector: &selector,
		Value:    map[string]any{"name": t.Name(), "data": t.Data()},
	})
	return p
}

func (p *patcher) BeforeEl(selector string, t TemplateRenderer) Patcher {
	p.patchset = append(p.patchset, Patch{
		Type:     Before,
		Selector: &selector,
		Value:    map[string]any{"name": t.Name(), "data": t.Data()},
	})
	return p
}

func (p *patcher) AppendEl(selector string, t TemplateRenderer) Patcher {
	p.patchset = append(p.patchset, Patch{
		Type:     Append,
		Selector: &selector,
		Value:    map[string]any{"name": t.Name(), "data": t.Data()},
	})
	return p
}

func (p *patcher) PrependEl(selector string, t TemplateRenderer) Patcher {
	p.patchset = append(p.patchset, Patch{
		Type:     Prepend,
		Selector: &selector,
		Value:    map[string]any{"name": t.Name(), "data": t.Data()},
	})
	return p
}

func (p *patcher) RemoveEl(selector string) Patcher {
	p.patchset = append(p.patchset, Patch{
		Type:     Remove,
		Selector: &selector,
	})
	return p
}

func (p *patcher) Reload() Patcher {
	p.patchset = append(p.patchset, Patch{
		Type: Reload,
	})
	return p
}

func (p *patcher) Store(name string, data any) Patcher {
	p.patchset = append(p.patchset, Patch{
		Type:     Store,
		Selector: &name,
		Value:    data,
	})
	return p
}

func (p *patcher) ResetForm(selector string) Patcher {
	p.patchset = append(p.patchset, Patch{
		Type:     ResetForm,
		Selector: &selector,
	})
	return p
}

func (p *patcher) Navigate(url string) Patcher {
	p.patchset = append(p.patchset, Patch{
		Type:  Navigate,
		Value: url,
	})
	return p
}

func (p *patcher) Patchset() Patchset {
	return p.patchset
}

func (p *patcher) Error() string {
	b, _ := json.Marshal(p.patchset)
	return string(b)
}

// MarshalPatchset renders the patch operations to a json string
func MarshalPatchset(t *template.Template, patchset []Patch) []byte {
	var renderedPatchset []Patch
	firErrorPatchExists := false
	for _, p := range patchset {
		switch p.Type {
		case Store, Navigate, ResetForm, Reload, Remove:
			renderedPatchset = append(renderedPatchset, p)
		case Replace, After, Before, Append, Prepend:
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
			renderedPatchset = append([]Patch{{
				Type:     Replace,
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
		for _, tmpl := range t.Templates() {
			if !match.IsPattern(tmpl.Name()) {
				continue
			}
			if ok, stopped := match.MatchLimit(name, tmpl.Name(), 10); ok || stopped {
				if stopped {
					glog.Errorf("template match stopped: %s, %s", name, tmpl.Name())
					break
				}
				name = tmpl.Name()
				break
			}
		}
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
