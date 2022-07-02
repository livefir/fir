package fir

import (
	"encoding/json"
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

type Operation struct {
	Op       Op          `json:"op"`
	Selector string      `json:"selector"`
	Value    interface{} `json:"value"`
}

func (m *Operation) Bytes() []byte {
	b, err := json.Marshal(m)
	if err != nil {
		log.Printf("error marshalling dom %v\n", err)
		return nil
	}
	return b
}

type Patch interface {
	Op() Op
}

type Patchset []Patch

type Template struct {
	Name string `json:"name"`
	Data Data   `json:"data"`
}

type Morph struct {
	Selector string
	Template *Template
}

func (m Morph) Op() Op {
	return morph
}

type After struct {
	Selector string
	Template *Template
}

func (a After) Op() Op {
	return after
}

type Before struct {
	Selector string
	Template *Template
}

func (b Before) Op() Op {
	return before
}

type Append struct {
	Selector string
	Template *Template
}

func (a Append) Op() Op {
	return appendOp
}

type Prepend struct {
	Selector string
	Template *Template
}

func (p Prepend) Op() Op {
	return prepend
}

type Remove struct {
	Selector string
	Template *Template
}

func (r Remove) Op() Op {
	return remove
}

type Store struct {
	Name string
	Data any
}

func (s Store) Op() Op {
	return updateStore
}

type Reload struct{}

func (r Reload) Op() Op {
	return reload
}

type ResetForm struct {
	Selector string
}

func (r ResetForm) Op() Op {
	return resetForm
}

type Navigate struct {
	To string
}

func (n Navigate) Op() Op {
	return navigate
}

func morphError(err string) Patch {
	return Morph{
		Selector: "#fir-error",
		Template: &Template{
			Name: "fir-error",
			Data: Data{"error": err}},
	}
}

func Error(err error) Patchset {
	log.Printf("[controller] error: %s\n", err)
	return Patchset{morphError(UserError(err))}
}
