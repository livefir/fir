package fir

import (
	"reflect"
	"testing"
)

type TestData struct {
	Name string
	Age  int
}

func TestBuildData(t *testing.T) {
	// Test case 1: No dataset provided
	err := buildData(false)
	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}

	// Test case 2: Only routeData provided
	data := map[string]any{"key": "value"}
	err = buildData(false, data)
	if err == nil {
		t.Errorf("Expected  error, got: %v", err)
	}
	r, ok := err.(*routeData)
	if !ok {
		t.Errorf("Expected error type *routeData, got: %v", reflect.TypeOf(err))
	}
	if !reflect.DeepEqual(*r, routeData{"key": "value"}) {
		t.Errorf("Expected error value %v, got: %v", data, *r)
	}

	// Test case 2: Only stateData provided
	err = buildData(true, data)
	if err == nil {
		t.Errorf("Expected  error, got: %v", err)
	}
	s, ok := err.(*stateData)
	if !ok {
		t.Errorf("Expected error type *stateData, got: %v", reflect.TypeOf(err))
	}
	if !reflect.DeepEqual(*s, stateData{"key": "value"}) {
		t.Errorf("Expected error value %v, got: %v", data, *s)
	}

	// Test case 3: Both routeData and stateData provided
	err = buildData(false, data, buildData(true, map[string]any{"key1": "value1"}))
	if err == nil {
		t.Errorf("Expected  error, got: %v", err)
	}
	rs, ok := err.(*routeDataWithState)
	if !ok {
		t.Errorf("Expected error type *routeDataWithState, got: %v", reflect.TypeOf(err))
	}
	expectedRouteData := routeData{"key": "value", "key1": "value1"}
	if !reflect.DeepEqual(*rs.routeData, expectedRouteData) {
		t.Errorf("Expected error value %v, got: %v", expectedRouteData, *rs.routeData)
	}
	if !reflect.DeepEqual(*rs.stateData, stateData{"key1": "value1"}) {
		t.Errorf("Expected error value %v, got: %v", data, *rs.stateData)
	}

}
