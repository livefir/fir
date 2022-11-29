package fir

import (
	"encoding/json"
	"log"
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
