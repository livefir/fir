package dom

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

// NewPatcher creates a new dom patcher
func NewPatcher() Patcher {
	return &patcher{
		patchset: Patchset{},
	}
}

type Patcher interface {
	// ReplaceEl patches the dom element at the given selector with the rendered template
	ReplaceEl(selector string, t TemplateRenderer) Patcher
	// ReplaceContent patches the dom element content at the given selector with the rendered template
	ReplaceContent(selector string, t TemplateRenderer) Patcher
	// After patches the dom element after the given selector with the rendered template
	AfterEl(selector string, t TemplateRenderer) Patcher
	// Before patches the dom element before the given selector with the rendered template
	BeforeEl(selector string, t TemplateRenderer) Patcher
	// Append patches the dom element  after the given selector with the rendered template
	AppendEl(selector string, t TemplateRenderer) Patcher
	// Prepend patches the dom element before the given selector with the rendered template
	PrependEl(selector string, t TemplateRenderer) Patcher
	// Remove patches the dom element to remove the given selector
	RemoveEl(selector string) Patcher
	// Reload patches the dom  to reload the page
	Reload() Patcher
	// Store patches the dom  to update the alpinejs store
	Store(name string, data any) Patcher
	// ResetForm patches the dom to reset the form
	ResetForm(selector string) Patcher
	// Navigate patches the dom to navigate to the given url
	Navigate(url string) Patcher
	// DispatchEvent patches the dom to dispatch the given event
	DispatchEvent(selector, eventSourceID string, t TemplateRenderer) Patcher
	// Patchset returns the patchset
	Patchset() Patchset
	// Error satisfies the error interface so that Context can return a Patcher
	Error() string
}

var _ error = (Patcher)(nil)

// PatchType is the type of patch operation
type PatchType string

const (
	ReplaceElement PatchType = "replaceElement"
	ReplaceContent PatchType = "replaceContent"
	After          PatchType = "after"
	Before         PatchType = "before"
	Append         PatchType = "append"
	Prepend        PatchType = "prepend"
	Remove         PatchType = "remove"
	Reload         PatchType = "reload"
	Store          PatchType = "store"
	ResetForm      PatchType = "resetForm"
	Navigate       PatchType = "navigate"
	DispatchEvent  PatchType = "dispatchEvent"
)

// Patch is an interface for all patch operations
type Patch struct {
	// Type is the type of patch operation
	Type PatchType `json:"op"`
	// Selector is the css selector for the element to patch
	Selector *string `json:"selector,omitempty"`
	// Value is the value for the patch operation
	Value any `json:"value,omitempty"`
	// EventSourceID is the id of the element that triggered the event
	EventSourceID *string `json:"eid,omitempty"`
}

// Patchset is a collection of patch operations
type Patchset []Patch

type patcher struct {
	patchset Patchset
}

func (p *patcher) ReplaceEl(selector string, t TemplateRenderer) Patcher {
	templateName := t.Name()
	templateData := t.Data()
	p.patchset = append(p.patchset, Patch{
		Type:     ReplaceElement,
		Selector: &selector,
		Value:    map[string]any{"name": templateName, "data": templateData},
	})
	return p
}

func (p *patcher) ReplaceContent(selector string, t TemplateRenderer) Patcher {
	templateName := t.Name()
	templateData := t.Data()
	p.patchset = append(p.patchset, Patch{
		Type:     ReplaceContent,
		Selector: &selector,
		Value:    map[string]any{"name": templateName, "data": templateData},
	})
	return p
}

func (p *patcher) AfterEl(selector string, t TemplateRenderer) Patcher {
	templateName := t.Name()
	templateData := t.Data()
	p.patchset = append(p.patchset, Patch{
		Type:     After,
		Selector: &selector,
		Value:    map[string]any{"name": templateName, "data": templateData},
	})
	return p
}

func (p *patcher) BeforeEl(selector string, t TemplateRenderer) Patcher {
	templateName := t.Name()
	templateData := t.Data()
	p.patchset = append(p.patchset, Patch{
		Type:     Before,
		Selector: &selector,
		Value:    map[string]any{"name": templateName, "data": templateData},
	})
	return p
}

func (p *patcher) AppendEl(selector string, t TemplateRenderer) Patcher {
	templateName := t.Name()
	templateData := t.Data()
	p.patchset = append(p.patchset, Patch{
		Type:     Append,
		Selector: &selector,
		Value:    map[string]any{"name": templateName, "data": templateData},
	})
	return p
}

func (p *patcher) PrependEl(selector string, t TemplateRenderer) Patcher {
	templateName := t.Name()
	templateData := t.Data()
	p.patchset = append(p.patchset, Patch{
		Type:     Prepend,
		Selector: &selector,
		Value:    map[string]any{"name": templateName, "data": templateData},
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

func (p *patcher) DispatchEvent(selector, eventSourceID string, t TemplateRenderer) Patcher {
	eventID := fmt.Sprintf("fir:%s", selector)
	patch := Patch{
		Type:          DispatchEvent,
		Selector:      &eventID,
		Value:         nil,
		EventSourceID: &eventSourceID,
	}

	if t != nil {
		templateName := t.Name()
		templateData := t.Data()
		selector = fmt.Sprintf("fir:%s:%s", selector, templateName)
		patch.Selector = &selector
		patch.Value = map[string]any{"name": templateName, "data": templateData}
	}
	p.patchset = append(p.patchset, patch)
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
	for _, p := range patchset {
		switch p.Type {
		case Store, Navigate, ResetForm, Reload, Remove:
			renderedPatchset = append(renderedPatchset, p)
		case ReplaceElement, After, Before, Append, Prepend, DispatchEvent:
			if p.Value == nil {
				renderedPatchset = append(renderedPatchset, p)
				continue
			}
			tmpl, ok := p.Value.(map[string]any)
			if !ok {
				glog.Errorf("[buildPatchOperations] invalid patch template data: %v", p.Value)
				continue
			}

			var err error
			p.Value, err = buildTemplateValue(t, tmpl["name"].(string), tmpl["data"])
			if err != nil {
				glog.Errorf("[warning]buildPatchOperations error: %v,%+v \n", err, tmpl)
				continue
			}

			renderedPatchset = append(renderedPatchset, p)
		default:
			renderedPatchset = append(renderedPatchset, p)
			continue
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
