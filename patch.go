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
	// GetRender returns the template for the patch operation
	GetRender() *Render
}

// Patchset is a set of patche operations to be applied to the DOM
type Patchset []Patch

type patch struct {
	OpVal    Op      `json:"op"`
	Selector string  `json:"selector"`
	Render   *Render `json:"render,omitempty"`
	Value    any     `json:"value,omitempty"`
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

func (p *patch) GetRender() *Render {
	return p.Render
}

// Render is a html/template to be rendered
type Render struct {
	// Template is the name of the template
	Template string `json:"template"`
	// Data is the data to be passed to the template
	Data any `json:"data"`
}

// Morph is a patch operation to morph a DOM element
type Morph struct {
	Selector string
	HTML     *Render
}

func (m Morph) Op() Op {
	return morph
}

func (m Morph) GetSelector() string {
	return m.Selector
}

func (m Morph) GetRender() *Render {
	return m.HTML
}

// After is a patch operation to insert a DOM element after a selector
type After struct {
	Selector string
	HTML     *Render
}

func (a After) Op() Op {
	return after
}

func (a After) GetSelector() string {
	return a.Selector
}

func (a After) GetRender() *Render {
	return a.HTML
}

// Before is a patch operation to insert a DOM element before a selector
type Before struct {
	Selector string
	HTML     *Render
}

func (b Before) GetSelector() string {
	return b.Selector
}

func (b Before) Op() Op {
	return before
}

func (b Before) GetRender() *Render {
	return b.HTML
}

// Append is a patch operation to append a DOM element to a selector
type Append struct {
	Selector string
	HTML     *Render
}

func (a Append) GetSelector() string {
	return a.Selector
}

func (a Append) Op() Op {
	return appendOp
}

func (a Append) GetRender() *Render {
	return a.HTML
}

// Prepend is a patch operation to prepend a DOM element to a selector
type Prepend struct {
	Selector string
	HTML     *Render
}

func (p Prepend) GetSelector() string {
	return p.Selector
}

func (p Prepend) Op() Op {
	return prepend
}

func (p Prepend) GetRender() *Render {
	return p.HTML
}

// Remove is a patch operation to remove a DOM element
type Remove struct {
	Selector string
	HTML     *Render
}

func (r Remove) GetSelector() string {
	return r.Selector
}

func (r Remove) Op() Op {
	return remove
}

func (r Remove) GetRender() *Render {
	return r.HTML
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

func (s Store) GetRender() *Render {
	return &Render{
		Template: s.Name,
		Data:     s.Data,
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

func (r Reload) GetRender() *Render {
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

func (r ResetForm) GetRender() *Render {
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

func (n Navigate) GetRender() *Render {
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
				Value:    p.GetRender().Data,
			})
		case navigate, reload, resetForm:
			patches = append(patches, patch{OpVal: p.Op(), Selector: p.GetSelector()})
		case morph, after, before, appendOp, prepend, remove:
			var buf bytes.Buffer
			err := t.ExecuteTemplate(&buf, p.GetRender().Template, p.GetRender().Data)
			if err != nil {
				log.Printf("buildPatchOperations error: %+v, %v \n", err, p.GetRender())
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
