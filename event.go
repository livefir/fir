package fir

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
)

func NewEvent(id string, params any) Event {
	data, err := json.Marshal(params)
	if err != nil {
		log.Printf("error marshaling event params: %v, %v, %v \n,", id, params, err)
		return Event{
			ID: id,
		}
	}
	return Event{
		ID:     id,
		Params: data,
	}
}

// Event is a struct that holds the data for an event
type Event struct {
	// Name is the name of the event
	ID string `json:"id"`
	// Params is the json rawmessage to be passed to the event
	Params json.RawMessage `json:"params"`
	IsForm bool            `json:"isForm"`
}

// String returns the string representation of the event
func (e Event) String() string {
	data, _ := json.MarshalIndent(e, "", " ")
	return string(data)
}

type Context struct {
	event     Event
	request   *http.Request
	response  http.ResponseWriter
	urlValues url.Values
	route     *route
}

// DecodeParams decodes the event params into the given struct
func (c Context) DecodeParams(v any) error {
	if c.event.IsForm {
		if len(c.urlValues) == 0 {
			var urlValues url.Values
			if err := json.NewDecoder(bytes.NewReader(c.event.Params)).Decode(&urlValues); err != nil {
				return err
			}
			c.urlValues = urlValues
		}
		return c.route.formDecoder.Decode(v, c.urlValues)
	}
	return json.NewDecoder(bytes.NewReader(c.event.Params)).Decode(v)
}

func (c Context) Request() *http.Request {
	return c.request
}
func (c Context) Response() http.ResponseWriter {
	return c.response
}

func (c Context) Data(data map[string]any) error {
	m := routeData{}
	for k, v := range data {
		m[k] = v
	}
	return &m
}
func (c *Context) Patch(patch ...Patch) error {
	var pl patchlist
	for _, p := range patch {
		pl = append(pl, p)
	}
	return &pl
}

func getJSON(data map[string]any) string {
	b, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return err.Error()
	}
	return string(b)
}
