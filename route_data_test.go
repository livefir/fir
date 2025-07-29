package fir

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestRouteData_Error(t *testing.T) {
	testCases := []struct {
		name     string
		data     routeData
		expected string
	}{
		{
			name:     "empty route data",
			data:     routeData{},
			expected: "{}",
		},
		{
			name: "simple route data",
			data: routeData{
				"message": "Hello World",
				"count":   42,
			},
			expected: `{"count":42,"message":"Hello World"}`,
		},
		{
			name: "route data with nil values",
			data: routeData{
				"nullable": nil,
				"value":    "not null",
			},
			expected: `{"nullable":null,"value":"not null"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.data.Error()
			assert.NotEmpty(t, result)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestStateData_Error(t *testing.T) {
	testCases := []struct {
		name     string
		data     stateData
		expected string
	}{
		{
			name:     "empty state data",
			data:     stateData{},
			expected: "{}",
		},
		{
			name: "simple state data",
			data: stateData{
				"status":  "active",
				"counter": 10,
			},
			expected: `{"counter":10,"status":"active"}`,
		},
		{
			name: "state data with mixed types",
			data: stateData{
				"string":  "value",
				"number":  123,
				"boolean": false,
			},
			expected: `{"boolean":false,"number":123,"string":"value"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.data.Error()
			assert.NotEmpty(t, result)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestRouteDataWithState_ErrorMethods(t *testing.T) {
	// Test the routeDataWithState struct creation and usage
	routeData := &routeData{
		"page":  "home",
		"title": "Welcome",
	}

	stateData := &stateData{
		"loading": true,
		"user_id": 123,
	}

	combined := routeDataWithState{
		routeData: routeData,
		stateData: stateData,
	}

	// Test that both data structures are accessible
	assert.NotNil(t, combined.routeData)
	assert.NotNil(t, combined.stateData)

	// Test that Error methods work on both
	routeError := combined.routeData.Error()
	stateError := combined.stateData.Error()

	assert.Contains(t, routeError, "page")
	assert.Contains(t, routeError, "title")
	assert.Contains(t, stateError, "loading")
	assert.Contains(t, stateError, "user_id")

	// Test the combined Error method
	combinedError := combined.Error()
	assert.Contains(t, combinedError, "routeData:")
	assert.Contains(t, combinedError, "stateData:")
	assert.Contains(t, combinedError, "page")
	assert.Contains(t, combinedError, "loading")
}

func TestRouteDataWithState_Error(t *testing.T) {
	routeData := &routeData{
		"status": "success",
		"count":  42,
	}

	stateData := &stateData{
		"active": true,
		"mode":   "test",
	}

	combined := routeDataWithState{
		routeData: routeData,
		stateData: stateData,
	}

	result := combined.Error()

	// Should contain the formatted string with both parts
	assert.Contains(t, result, "routeData:")
	assert.Contains(t, result, "stateData:")

	// Should contain data from both routeData and stateData
	assert.Contains(t, result, "status")
	assert.Contains(t, result, "success")
	assert.Contains(t, result, "count")
	assert.Contains(t, result, "42")
	assert.Contains(t, result, "active")
	assert.Contains(t, result, "true")
	assert.Contains(t, result, "mode")
	assert.Contains(t, result, "test")

	// Should have the specific format with newline separator
	assert.Contains(t, result, "\n")
}
