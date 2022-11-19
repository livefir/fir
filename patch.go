package fir

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
)

type Op string

const (
	morph       Op = "morph"
	after       Op = "after"
	before      Op = "before"
	appendOp    Op = "append"
	prepend     Op = "prepend"
	remove      Op = "remove"
	reload      Op = "reload"
	updateStore Op = "store"
	resetForm   Op = "resetForm"
	navigate    Op = "navigate"
)

// Patch is an interface for all patch operations
type Patch interface {
	Op() Op
	GetSelector() string
	GetTemplate() *Template
}

// Patchset is a set of patches to be applied to the DOM
type Patchset []Patch

type patch struct {
	OpVal    Op        `json:"op"`
	Selector string    `json:"selector"`
	Template *Template `json:"template,omitempty"`
	Value    any       `json:"value,omitempty"`
}

func (p *patch) Op() Op {
	return p.OpVal
}

func (p *patch) toPatch() Patch {
	return p
}

func (p *patch) GetSelector() string {
	return p.Selector
}

func (p *patch) GetTemplate() *Template {
	return p.Template
}

// Template is a html/template to be rendered
type Template struct {
	Name string `json:"name"`
	Data any    `json:"data"`
}

type Morph struct {
	Selector string
	Template *Template
}

func (m Morph) Op() Op {
	return morph
}

func (m Morph) GetSelector() string {
	return m.Selector
}

func (m Morph) GetTemplate() *Template {
	return m.Template
}

type After struct {
	Selector string
	Template *Template
}

func (a After) Op() Op {
	return after
}

func (a After) GetSelector() string {
	return a.Selector
}

func (a After) GetTemplate() *Template {
	return a.Template
}

type Before struct {
	Selector string
	Template *Template
}

func (b Before) GetSelector() string {
	return b.Selector
}

func (b Before) Op() Op {
	return before
}

func (b Before) GetTemplate() *Template {
	return b.Template
}

type Append struct {
	Selector string
	Template *Template
}

func (a Append) GetSelector() string {
	return a.Selector
}

func (a Append) Op() Op {
	return appendOp
}

func (a Append) GetTemplate() *Template {
	return a.Template
}

type Prepend struct {
	Selector string
	Template *Template
}

func (p Prepend) GetSelector() string {
	return p.Selector
}

func (p Prepend) Op() Op {
	return prepend
}

func (p Prepend) GetTemplate() *Template {
	return p.Template
}

type Remove struct {
	Selector string
	Template *Template
}

func (r Remove) GetSelector() string {
	return r.Selector
}

func (r Remove) Op() Op {
	return remove
}

func (r Remove) GetTemplate() *Template {
	return r.Template
}

type Store struct {
	Name string
	Data any
}

func (s Store) GetSelector() string {
	return s.Name
}

func (s Store) Op() Op {
	return updateStore
}

func (s Store) GetTemplate() *Template {
	return &Template{
		Name: s.Name,
		Data: s.Data,
	}
}

type Reload struct{}

func (r Reload) GetSelector() string {
	return ""
}

func (r Reload) Op() Op {
	return reload
}

func (r Reload) GetTemplate() *Template {
	return nil
}

type ResetForm struct {
	Selector string
}

func (r ResetForm) GetSelector() string {
	return r.Selector
}

func (r ResetForm) Op() Op {
	return resetForm
}

func (r ResetForm) GetTemplate() *Template {
	return nil
}

type Navigate struct {
	To string
}

func (n Navigate) GetSelector() string {
	return n.To
}

func (n Navigate) Op() Op {
	return navigate
}

func (n Navigate) GetTemplate() *Template {
	return nil
}

func buildPatchOperations(t *template.Template, patchset Patchset) []byte {
	var patches []patch
	for _, p := range patchset {
		switch p.Op() {
		case updateStore:
			patches = append(patches, patch{
				OpVal:    updateStore,
				Selector: p.GetSelector(),
				Value:    p.GetTemplate().Data,
			})
		case reload:
			patches = append(patches, patch{OpVal: reload})
		case resetForm:
			patches = append(patches, patch{OpVal: resetForm, Selector: p.GetSelector()})
		case navigate:
			patches = append(patches, patch{OpVal: navigate, Value: p.GetSelector()})
		case morph, after, before, appendOp, prepend, remove:
			var buf bytes.Buffer
			err := t.ExecuteTemplate(&buf, p.GetTemplate().Name, p.GetTemplate().Data)
			if err != nil {
				log.Printf("buildPatchOperations error: %+v, %v \n", err, p.GetTemplate())
				continue
			}

			html := buf.String()
			buf.Reset()
			patches = append(patches, patch{OpVal: p.Op(), Selector: p.GetSelector(), Value: html})

		default:
			continue
		}
	}

	data, err := json.Marshal(patches)
	if err != nil {
		log.Printf("buildPatchOperations marshal error: %+v, %v \n", patches, err)
		return nil
	}
	return data
}
