package fir

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewEvent(t *testing.T) {
	// Test NewEvent function
	event := NewEvent("test-event", map[string]any{"key": "value"})

	assert.Equal(t, "test-event", event.ID)
	assert.NotNil(t, event.Params)

	// Unmarshal params to verify content
	var params map[string]any
	err := json.Unmarshal(event.Params, &params)
	assert.NoError(t, err)
	assert.Equal(t, "value", params["key"])

	// Verify timestamp is NOT set by NewEvent (it's 0)
	assert.Equal(t, int64(0), event.Timestamp)

	// Test with nil data - this marshals to "null" JSON
	eventNil := NewEvent("nil-event", nil)
	assert.Equal(t, "nil-event", eventNil.ID)
	assert.Equal(t, json.RawMessage("null"), eventNil.Params) // nil marshals to "null"

	// Test with complex data
	complexData := map[string]any{
		"user": map[string]string{
			"name": "John",
			"role": "admin",
		},
		"count":  42,
		"active": true,
	}
	eventComplex := NewEvent("complex", complexData)
	assert.Equal(t, "complex", eventComplex.ID)

	var complexParams map[string]any
	err = json.Unmarshal(eventComplex.Params, &complexParams)
	assert.NoError(t, err)
	assert.Equal(t, 42.0, complexParams["count"]) // JSON unmarshals numbers as float64
	assert.Equal(t, true, complexParams["active"])

	// Test with unmarshalable data (covers error path)
	unmarshalable := make(chan int)
	eventUnmarshalable := NewEvent("unmarshalable", unmarshalable)
	assert.Equal(t, "unmarshalable", eventUnmarshalable.ID)
	assert.Nil(t, eventUnmarshalable.Params) // Should be nil when marshaling fails
}

func TestEvent_String(t *testing.T) {
	// Test with complete event
	params := map[string]any{
		"username": "testuser",
		"age":      25,
	}
	paramsJSON, _ := json.Marshal(params)

	event := Event{
		ID:         "user-action",
		Params:     paramsJSON,
		Target:     stringPtr("user-form"),
		IsForm:     true,
		SessionID:  stringPtr("session123"),
		ElementKey: stringPtr("element456"),
		Timestamp:  1234567890,
	}

	result := event.String()
	assert.Contains(t, result, "user-action")
	assert.Contains(t, result, "session123")
	assert.Contains(t, result, "element456")
	assert.Contains(t, result, "user-form")

	// Test with minimal event
	minimalEvent := Event{
		ID: "minimal",
	}

	result = minimalEvent.String()
	assert.Contains(t, result, "minimal")

	// Verify it's valid JSON by unmarshaling
	var eventMap map[string]any
	err := json.Unmarshal([]byte(result), &eventMap)
	assert.NoError(t, err)
	assert.Equal(t, "minimal", eventMap["event_id"])
}

func TestToUnixTime(t *testing.T) {
	// Test toUnixTime function - converts int64 milliseconds to time.Time
	timestamp := int64(1640995200000) // 2022-01-01 00:00:00 UTC in milliseconds
	result := toUnixTime(timestamp)

	expected := time.Unix(1640995200, 0) // Convert to seconds for comparison
	assert.Equal(t, expected, result)

	// Test with zero timestamp
	zeroResult := toUnixTime(0)
	expectedZero := time.Unix(0, 0)
	assert.Equal(t, expectedZero, zeroResult)

	// Test with timestamp that has milliseconds
	timestampWithMs := int64(1640995200123) // Has 123 milliseconds
	resultWithMs := toUnixTime(timestampWithMs)

	expectedWithMs := time.Unix(1640995200, 123*1000*1000) // 123 milliseconds in nanoseconds
	assert.Equal(t, expectedWithMs, resultWithMs)
}

func TestEvent_fields(t *testing.T) {
	// Test individual field assignments
	event := Event{}

	// Test ID
	event.ID = "test-id"
	assert.Equal(t, "test-id", event.ID)

	// Test Params
	testParams := json.RawMessage(`{"test": true}`)
	event.Params = testParams
	assert.Equal(t, testParams, event.Params)

	// Test Target
	target := "target-element"
	event.Target = &target
	assert.Equal(t, "target-element", *event.Target)

	// Test IsForm
	event.IsForm = true
	assert.True(t, event.IsForm)

	// Test SessionID
	sessionID := "sess-123"
	event.SessionID = &sessionID
	assert.Equal(t, "sess-123", *event.SessionID)

	// Test ElementKey
	elementKey := "elem-456"
	event.ElementKey = &elementKey
	assert.Equal(t, "elem-456", *event.ElementKey)

	// Test Timestamp
	event.Timestamp = 1234567890
	assert.Equal(t, int64(1234567890), event.Timestamp)
}

func TestEvent_jsonSerialization(t *testing.T) {
	// Test that Event can be serialized to and from JSON
	originalEvent := Event{
		ID:         "serialization-test",
		Params:     json.RawMessage(`{"data":"test"}`), // No spaces to match standard JSON
		Target:     stringPtr("test-target"),
		IsForm:     true,
		SessionID:  stringPtr("test-session"),
		ElementKey: stringPtr("test-element"),
		Timestamp:  1234567890,
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(originalEvent)
	assert.NoError(t, err)

	// Deserialize from JSON
	var deserializedEvent Event
	err = json.Unmarshal(jsonData, &deserializedEvent)
	assert.NoError(t, err)

	// Verify all fields match
	assert.Equal(t, originalEvent.ID, deserializedEvent.ID)
	assert.Equal(t, *originalEvent.Target, *deserializedEvent.Target)
	assert.Equal(t, originalEvent.IsForm, deserializedEvent.IsForm)
	assert.Equal(t, *originalEvent.SessionID, *deserializedEvent.SessionID)
	assert.Equal(t, *originalEvent.ElementKey, *deserializedEvent.ElementKey)
	assert.Equal(t, originalEvent.Timestamp, deserializedEvent.Timestamp)

	// Compare params by content rather than exact bytes (JSON formatting may differ)
	var originalParams, deserializedParams map[string]any
	err = json.Unmarshal(originalEvent.Params, &originalParams)
	assert.NoError(t, err)
	err = json.Unmarshal(deserializedEvent.Params, &deserializedParams)
	assert.NoError(t, err)
	assert.Equal(t, originalParams, deserializedParams)
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

func TestFir(t *testing.T) {
	tests := []struct {
		name           string
		parts          []string
		expectedResult *string
		shouldPanic    bool
		panicMessage   string
	}{
		// No arguments should panic
		{
			name:         "No arguments should panic",
			parts:        []string{},
			shouldPanic:  true,
			panicMessage: "fir: fir() called with no arguments",
		},

		// Single argument cases
		{
			name:           "Single argument - simple event",
			parts:          []string{"click"},
			expectedResult: stringPtr("fir:click"),
		},
		{
			name:           "Single argument - complex event name",
			parts:          []string{"form-submit"},
			expectedResult: stringPtr("fir:form-submit"),
		},
		{
			name:           "Single argument - empty string",
			parts:          []string{""},
			expectedResult: stringPtr("fir:"),
		},
		{
			name:           "Single argument - special characters",
			parts:          []string{"user@action"},
			expectedResult: stringPtr("fir:user@action"),
		},

		// Two argument cases
		{
			name:           "Two arguments - event and template",
			parts:          []string{"click", "button-template"},
			expectedResult: stringPtr("fir:click::button-template"),
		},
		{
			name:           "Two arguments - complex names",
			parts:          []string{"form-submit", "user-form-template"},
			expectedResult: stringPtr("fir:form-submit::user-form-template"),
		},
		{
			name:           "Two arguments - empty event name",
			parts:          []string{"", "template"},
			expectedResult: stringPtr("fir:::template"),
		},
		{
			name:           "Two arguments - empty template name",
			parts:          []string{"event", ""},
			expectedResult: stringPtr("fir:event::"),
		},
		{
			name:           "Two arguments - both empty",
			parts:          []string{"", ""},
			expectedResult: stringPtr("fir:::"),
		},
		{
			name:           "Two arguments - with special characters",
			parts:          []string{"user@action", "template-name.html"},
			expectedResult: stringPtr("fir:user@action::template-name.html"),
		},

		// More than two arguments should panic
		{
			name:         "Three arguments should panic",
			parts:        []string{"event", "template", "extra"},
			shouldPanic:  true,
			panicMessage: "fir: fir() called with more than 2 arguments",
		},
		{
			name:         "Four arguments should panic",
			parts:        []string{"a", "b", "c", "d"},
			shouldPanic:  true,
			panicMessage: "fir: fir() called with more than 2 arguments",
		},
		{
			name:         "Many arguments should panic",
			parts:        []string{"a", "b", "c", "d", "e", "f", "g"},
			shouldPanic:  true,
			panicMessage: "fir: fir() called with more than 2 arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				defer func() {
					if r := recover(); r != nil {
						// Verify the panic message
						assert.Equal(t, tt.panicMessage, r)
					} else {
						t.Error("Expected panic but none occurred")
					}
				}()
				fir(tt.parts...)
			} else {
				result := fir(tt.parts...)

				// Verify the result is not nil
				assert.NotNil(t, result)

				// Verify the content matches expected
				assert.Equal(t, *tt.expectedResult, *result)

				// Additional verification: ensure it's a valid pointer
				assert.IsType(t, (*string)(nil), result)
			}
		})
	}
}

func TestFir_realWorldUsage(t *testing.T) {
	// Test real-world usage patterns that might appear in templates

	// Common event names
	result1 := fir("load")
	assert.Equal(t, "fir:load", *result1)

	result2 := fir("submit")
	assert.Equal(t, "fir:submit", *result2)

	result3 := fir("click")
	assert.Equal(t, "fir:click", *result3)

	// Event with template specification
	result4 := fir("user-create", "user-list")
	assert.Equal(t, "fir:user-create::user-list", *result4)

	result5 := fir("form-validation", "error-display")
	assert.Equal(t, "fir:form-validation::error-display", *result5)

	// Edge cases that might occur in real usage
	result6 := fir("multi-word-event")
	assert.Equal(t, "fir:multi-word-event", *result6)

	result7 := fir("event123", "template456")
	assert.Equal(t, "fir:event123::template456", *result7)
}

func TestFir_consistentResults(t *testing.T) {
	// Test that calling the same function multiple times produces consistent results

	// Single argument consistency
	result1a := fir("test")
	result1b := fir("test")
	assert.Equal(t, *result1a, *result1b)
	assert.Equal(t, "fir:test", *result1a)

	// Two argument consistency
	result2a := fir("event", "template")
	result2b := fir("event", "template")
	assert.Equal(t, *result2a, *result2b)
	assert.Equal(t, "fir:event::template", *result2a)

	// Verify that different calls produce different results
	resultDiff1 := fir("different")
	resultDiff2 := fir("event")
	assert.NotEqual(t, *resultDiff1, *resultDiff2)
}

func TestFir_stringPointerBehavior(t *testing.T) {
	// Test that the function returns proper string pointers

	result := fir("test")

	// Verify it's a valid pointer
	assert.NotNil(t, result)

	// Verify we can dereference it
	value := *result
	assert.Equal(t, "fir:test", value)

	// Verify the pointer points to the correct memory
	// (modify through pointer and verify change)
	originalValue := *result
	newValue := "modified"
	*result = newValue
	assert.Equal(t, newValue, *result)
	assert.NotEqual(t, originalValue, *result)
}
