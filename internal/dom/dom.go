package dom

import (
	"github.com/livefir/fir/internal/eventstate"
)

type Event struct {
	Type   *string `json:"type,omitempty"`
	Target *string `json:"target,omitempty"`
	Detail any     `json:"detail,omitempty"`
	Key    *string `json:"key,omitempty"`
	// Private fields
	ID    string          `json:"-"`
	State eventstate.Type `json:"-"`
}
