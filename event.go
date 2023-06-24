package fir

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"k8s.io/klog/v2"
)

// NewEvent creates a new event
func NewEvent(id string, params any) Event {
	data, err := json.Marshal(params)
	if err != nil {
		klog.Errorf("error marshaling event params: %v, %v, %v \n,", id, params, err)
		return Event{
			ID: id,
		}
	}
	return Event{
		ID:     id,
		Params: data,
	}
}

type JSTime struct {
	time.Time
}

func (j *JSTime) UnmarshalJSON(b []byte) error {
	fmt.Println(string(b))
	i, err := strconv.ParseInt(string(b[1:len(b)-1]), 10, 64)
	if err != nil {
		return err
	}
	fmt.Println(i)
	*j = JSTime{Time: time.Unix(i/1000, (i%1000)*1000*1000)}
	return nil

}

func toUnixTime(ts int64) time.Time {
	return time.Unix(ts/1000, (ts%1000)*1000*1000)
}

// Event is a struct that holds the data for an incoming user event
type Event struct {
	// ID is the event id
	ID string `json:"event_id"`
	// Params is the data to be passed to the event
	Params json.RawMessage `json:"params"`
	// Target is the dom element id which emitted the event. The DOM events generated by the event will be targeted to this element.
	Target *string `json:"target,omitempty"`
	// IsForm is a boolean that indicates whether the event was triggered by a form submission
	IsForm bool `json:"is_form,omitempty"`
	// SessionID is the id of the session that the event was triggered for
	SessionID  *string `json:"session_id,omitempty"`
	ElementKey *string `json:"element_key,omitempty"`
	Timestamp  int64   `json:"ts,omitempty"`
}

// String returns the string representation of the event
func (e Event) String() string {
	data, _ := json.MarshalIndent(e, "", " ")
	return string(data)
}

func fir(parts ...string) *string {
	if len(parts) == 0 {
		panic("fir: fir() called with no arguments")
	}
	if len(parts) == 1 {
		s := fmt.Sprintf("fir:%s", parts[0])
		return &s
	}
	if len(parts) == 2 {
		s := fmt.Sprintf("fir:%s::%s", parts[0], parts[1])
		return &s
	}

	if len(parts) > 2 {
		panic("fir: fir() called with more than 2 arguments")
	}

	return nil
}
