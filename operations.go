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

type Morph struct {
	Selector string
	Template string
	Data     map[string]any
}

func (m Morph) Op() Op {
	return morph
}

type After struct {
	Selector string
	Template string
	Data     map[string]any
}

func (a After) Op() Op {
	return after
}

type Before struct {
	Selector string
	Template string
	Data     map[string]any
}

func (b Before) Op() Op {
	return before
}

type Append struct {
	Selector string
	Template string
	Data     map[string]any
}

func (a Append) Op() Op {
	return appendOp
}

type Prepend struct {
	Selector string
	Template string
	Data     map[string]any
}

func (p Prepend) Op() Op {
	return prepend
}

type Remove struct {
	Selector string
	Template string
	Data     map[string]any
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

func Error(err error) Patchset {
	log.Printf("[controller] error: %s\n", err)
	return Patchset{Morph{
		Selector: "#fir-error",
		Template: "fir-error",
		Data:     Data{"error": UserError(err)},
	}}
}
