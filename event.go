package fir

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"
)

type Data map[string]any

type Event struct {
	ID             string          `json:"id"`
	Params         json.RawMessage `json:"params"`
	requestContext context.Context
}

func (e Event) String() string {
	data, _ := json.MarshalIndent(e, "", " ")
	return string(data)
}

func (e Event) DecodeParams(v any) error {
	return json.NewDecoder(bytes.NewReader(e.Params)).Decode(v)
}

func (e Event) DecodeFormParams(v any) error {
	var urlValues url.Values
	if err := json.NewDecoder(bytes.NewReader(e.Params)).Decode(&urlValues); err != nil {
		return err
	}
	return decoder.Decode(v, urlValues)
}

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
