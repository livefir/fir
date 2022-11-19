package fir

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/url"
)

// Data is a map of data to be passed to the template
type Data map[string]any

// Event is a struct that holds the data for an event
type Event struct {
	// Name is the name of the event
	ID string `json:"id"`
	// Params is the json rawmessage to be passed to the event
	Params         json.RawMessage `json:"params"`
	requestContext context.Context
}

// String returns the string representation of the event
func (e Event) String() string {
	data, _ := json.MarshalIndent(e, "", " ")
	return string(data)
}

// DecodeParams decodes the event params into the given struct
func (e Event) DecodeParams(v any) error {
	return json.NewDecoder(bytes.NewReader(e.Params)).Decode(v)
}

// DecodeFormParams decodes the event params into the given struct
func (e Event) DecodeFormParams(v any) error {
	var urlValues url.Values
	if err := json.NewDecoder(bytes.NewReader(e.Params)).Decode(&urlValues); err != nil {
		return err
	}
	return decoder.Decode(v, urlValues)
}

// RequestContext returns the request context
func (e Event) RequestContext() context.Context {
	return e.requestContext
}

func getJSON(data Data) string {
	b, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return err.Error()
	}
	return string(b)
}

func getEventPatchset(event Event, view View) Patchset {
	patchset := view.OnEvent(event)
	if patchset == nil {
		log.Printf("[view] warning: no patchset returned for event: %v\n", event)
		patchset = Patchset{}
	}

	firErrorPatchExists := false

	for _, patch := range patchset {
		if patch.GetSelector() == "#fir-error" {
			firErrorPatchExists = true
		}
	}

	if !firErrorPatchExists {
		// unset error patch
		patchset = append([]Patch{morphError("")}, patchset...)
	}

	return patchset

}
