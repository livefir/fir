package fir

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/goccy/go-json"

	"github.com/fatih/structs"
)

type routeData map[string]any

func (r *routeData) Error() string {
	b, _ := json.Marshal(r)
	return string(b)
}

type stateData map[string]any

func (r *stateData) Error() string {
	b, _ := json.Marshal(r)
	return string(b)
}

type routeDataWithState struct {
	routeData *routeData
	stateData *stateData
}

func (r *routeDataWithState) Error() string {
	b1, _ := json.Marshal(r.routeData)
	b2, _ := json.Marshal(r.stateData)
	return fmt.Sprintf("routeData: %s\n stateData: %s", string(b1), string(b2))
}

func buildData(stateOnly bool, dataset ...any) error {
	if len(dataset) == 0 {
		return nil
	}

	m := make(map[string]any)
	hasState := false
	state := make(stateData)

	for _, data := range dataset {
		if data == nil {
			continue
		}
		if sv, ok := data.(*stateData); ok {
			hasState = true
			for k, v := range *sv {
				state[k] = v
				m[k] = v
			}
		}
		val := reflect.ValueOf(data)

		if val.Kind() == reflect.Ptr {
			el := val.Elem() // dereference the pointer
			if el.Kind() == reflect.Struct {
				for k, v := range structs.Map(data) {
					m[k] = v
				}
			}
		} else if val.Kind() == reflect.Struct {
			for k, v := range structs.Map(data) {
				m[k] = v
			}
		} else if val.Kind() == reflect.Map {
			ms, ok := data.(map[string]any)
			if !ok {
				return errors.New("data must be a map[string]any , struct or pointer to a struct")
			}

			for k, v := range ms {
				m[k] = v
			}
		} else {
			return errors.New("data must be a map[string]any , struct or pointer to a struct")
		}
	}

	if stateOnly {
		t := stateData(m)
		return &t
	}

	if hasState {
		r := routeData(m)
		t := routeDataWithState{
			routeData: &r,
			stateData: &state,
		}
		return &t
	}

	r := routeData(m)
	return &r
}
