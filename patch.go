package fir

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
)

// Op is the type of patch operation
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
	// Op returns the patch operation
	Op() Op
	// GetSelector returns the selector for the patch operation
	GetSelector() string
	// GetTemplate returns the template for the patch operation
	GetTemplate() *Template
}

// Patchset is a set of patche operations to be applied to the DOM
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
	// Name is the name of the template
	Name string `json:"name"`
	// Data is the data to be passed to the template
	Data any `json:"data"`
}

type Block = Template

// Morph is a patch operation to morph a DOM element
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

// After is a patch operation to insert a DOM element after a selector
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

// Before is a patch operation to insert a DOM element before a selector
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

// Append is a patch operation to append a DOM element to a selector
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

// Prepend is a patch operation to prepend a DOM element to a selector
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

// Remove is a patch operation to remove a DOM element
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

// Store is a patch operation to update alpine.js store in the browser
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

// Reload is a patch operation to reload the page in development mode
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

// ResetForm is a patch operation to reset a form
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

// Navigate is a patch operation to navigate to a new page
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
		case navigate, reload, resetForm:
			patches = append(patches, patch{OpVal: p.Op(), Selector: p.GetSelector()})
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
