package fir

import (
	"encoding/json"
	"log"
)

type Op string

const (
	morph       Op = "morph"
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
