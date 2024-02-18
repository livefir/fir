package dom

import (
	"github.com/livefir/fir/internal/eventstate"
)

type Detail struct {
	HTML  string `json:"html,omitempty"`
	State any    `json:"state,omitempty"`
	Data  any    `json:"data,omitempty"`
}

type Event struct {
	Type   *string `json:"type,omitempty"`
	Target *string `json:"target,omitempty"`
	Detail *Detail `json:"detail,omitempty"`
	Key    *string `json:"key,omitempty"`
	// Private fields
	ID    string          `json:"-"`
	State eventstate.Type `json:"-"`
}
