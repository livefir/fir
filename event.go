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
	Params    json.RawMessage `json:"params"`
	IsForm    bool            `json:"isForm"`
	request   *http.Request
	response  http.ResponseWriter
	urlValues url.Values
}

// String returns the string representation of the event
func (e Event) String() string {
	data, _ := json.MarshalIndent(e, "", " ")
	return string(data)
}

// DecodeParams decodes the event params into the given struct
func (e Event) DecodeParams(v any) error {
	if e.IsForm {
		if len(e.urlValues) == 0 {
			var urlValues url.Values
			if err := json.NewDecoder(bytes.NewReader(e.Params)).Decode(&urlValues); err != nil {
				return err
			}
			e.urlValues = urlValues
		}
		return decoder.Decode(v, e.urlValues)
	}
	return json.NewDecoder(bytes.NewReader(e.Params)).Decode(v)
}

// DecodeFormParams decodes the event params into the given struct
func (e Event) DecodeFormParams(v any) error {
	var urlValues url.Values
	if err := json.NewDecoder(bytes.NewReader(e.Params)).Decode(&urlValues); err != nil {
		log.Println("error decoding form params", err)
		return err
	}
	return decoder.Decode(v, urlValues)
}
func (e Event) Request() *http.Request {
	return e.request
}
func (e Event) Response() http.ResponseWriter {
	return e.response
}

func getJSON(data map[string]any) string {
	b, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return err.Error()
	}
	return string(b)
}
