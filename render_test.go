package fir

import (
	"html/template"
	"reflect"
	"testing"

	"github.com/livefir/fir/internal/dom"
	"github.com/livefir/fir/internal/eventstate"
	"github.com/patrickmn/go-cache"
)

func ptr(s string) *string {
	return &s
}

func TestBuildTemplateValue(t *testing.T) {
	// Test case 1: nil template
	result, err := buildTemplateValue(nil, "templateName", "data")
	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}
	if result != "" {
		t.Errorf("Expected empty result, got: %s", result)
	}

	// Test case 2: empty template name
	result, err = buildTemplateValue(&template.Template{}, "", "data")
	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}
	if result != "" {
		t.Errorf("Expected empty result, got: %s", result)
	}

	// Test case 3: templateName == "_fir_html"
	result, err = buildTemplateValue(&template.Template{}, "_fir_html", "data")
	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}
	if result != "data" {
		t.Errorf("Expected result 'data', got: %s", result)
	}

	// Test case 4: normal template execution
	tmpl := template.Must(template.New("test").Parse("Hello, {{.}}!"))
	result, err = buildTemplateValue(tmpl, "test", "World")
	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}
	if result != "Hello, World!" {
		t.Errorf("Expected result 'Hello, World!', got: %s", result)
	}

	// Test case 5: template execution error
	tmpl = template.Must(template.New("test").Parse("Hello, {{.World}}!"))
	result, err = buildTemplateValue(tmpl, "test", TestData{Name: "World"})
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	expectedError := "template: test:1:9: executing \"test\" at <.World>: can't evaluate field World in type fir.TestData"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got: %v", expectedError, err)
	}
	if result != "" {
		t.Errorf("Expected empty result, got: %s", result)
	}
}
func TestTargetOrClassName(t *testing.T) {
	// Test case 1: target is not nil and not empty
	target := new(string)
	*target = "target"
	className := "class"
	result := targetOrClassName(target, className)
	expected := "target"
	if *result != expected {
		t.Errorf("Expected result '%s', got: %s", expected, *result)
	}

	// Test case 2: target is nil
	target = nil
	result = targetOrClassName(target, className)
	expected = ".class"
	if *result != expected {
		t.Errorf("Expected result '%s', got: %s", expected, *result)
	}

	// Test case 3: target is empty
	target = new(string)
	result = targetOrClassName(target, className)
	expected = ".class"
	if *result != expected {
		t.Errorf("Expected result '%s', got: %s", expected, *result)
	}
}
func TestGetUnsetErrorEvents(t *testing.T) {
	// Test case 1: sessionID is nil
	cch := cache.New(cache.NoExpiration, cache.NoExpiration)
	sessionID := (*string)(nil)
	events := []dom.Event{
		{Type: ptr("error"), Target: ptr("target1"), State: eventstate.Error},
		{Type: ptr("error"), Target: ptr("target2"), State: eventstate.Error},
	}
	result := getUnsetErrorEvents(cch, sessionID, events)
	if result != nil {
		t.Errorf("Expected nil result, got: %v", result)
	}

	// Test case 2: cch is nil
	cch = nil
	sessionID = ptr("session1")
	events = []dom.Event{
		{Type: ptr("error"), Target: ptr("target1"), State: eventstate.Error},
		{Type: ptr("error"), Target: ptr("target2"), State: eventstate.Error},
	}
	result = getUnsetErrorEvents(cch, sessionID, events)
	if result != nil {
		t.Errorf("Expected nil result, got: %v", result)
	}

	// Test case 3: sessionID and cch are not nil
	// Test case 3.1: no new errors, previous errors to unset
	cch = cache.New(cache.NoExpiration, cache.NoExpiration)
	sessionID = ptr("session1")
	events = []dom.Event{}
	cch.Set(*sessionID,
		map[string]string{
			"error": "target2",
		}, cache.NoExpiration)

	result = getUnsetErrorEvents(cch, sessionID, events)
	expected := []dom.Event{
		{Type: ptr("error"), Target: ptr("target2")},
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected result %v, got: %v", expected, result)
	}
	// Test case 3.2: one new error, one previous error to unset
	cch = cache.New(cache.NoExpiration, cache.NoExpiration)
	sessionID = ptr("session1")
	events = []dom.Event{
		{Type: ptr("error"), Target: ptr("target1"), State: eventstate.Error},
	}
	cch.Set(*sessionID,
		map[string]string{
			"error2": "target2",
		}, cache.NoExpiration)

	result = getUnsetErrorEvents(cch, sessionID, events)

	expected = []dom.Event{
		{Type: ptr("error2"), Target: ptr("target2")},
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected result %v, got: %v", expected, result)
	}

	// Test case 3.3: two new errors, nothing to unset
	cch = cache.New(cache.NoExpiration, cache.NoExpiration)
	sessionID = ptr("session1")
	events = []dom.Event{
		{Type: ptr("error"), Target: ptr("target1"), State: eventstate.Error},
		{Type: ptr("error2"), Target: ptr("targe2"), State: eventstate.Error},
	}
	cch.Set(*sessionID,
		map[string]string{
			"error":  "target1",
			"error2": "target2",
		}, cache.NoExpiration)

	result = getUnsetErrorEvents(cch, sessionID, events)

	if len(result) != 0 {
		t.Errorf("Expected empty result, got: %v", result)
	}

}

func TestUniques(t *testing.T) {
	t.Run("EmptySlice", func(t *testing.T) {
		events := []dom.Event{}
		result := uniques(events)

		if len(result) != 0 {
			t.Errorf("Expected empty result, got %d events", len(result))
		}
	})

	t.Run("SingleEvent", func(t *testing.T) {
		events := []dom.Event{
			{
				Type:   ptr("click"),
				Target: ptr("#button1"),
				Key:    ptr("btn1"),
			},
		}
		result := uniques(events)

		if len(result) != 1 {
			t.Errorf("Expected 1 event, got %d", len(result))
		}
		if !reflect.DeepEqual(result[0], events[0]) {
			t.Errorf("Expected %+v, got %+v", events[0], result[0])
		}
	})

	t.Run("NoDuplicates", func(t *testing.T) {
		events := []dom.Event{
			{
				Type:   ptr("click"),
				Target: ptr("#button1"),
				Key:    ptr("btn1"),
			},
			{
				Type:   ptr("click"),
				Target: ptr("#button2"),
				Key:    ptr("btn2"),
			},
			{
				Type:   ptr("keydown"),
				Target: ptr("#input1"),
				Key:    ptr("inp1"),
			},
		}
		result := uniques(events)

		if len(result) != 3 {
			t.Errorf("Expected 3 events, got %d", len(result))
		}
		for i, expected := range events {
			if !reflect.DeepEqual(result[i], expected) {
				t.Errorf("Event %d: expected %+v, got %+v", i, expected, result[i])
			}
		}
	})

	t.Run("ExactDuplicates", func(t *testing.T) {
		event1 := dom.Event{
			Type:   ptr("click"),
			Target: ptr("#button1"),
			Key:    ptr("btn1"),
		}
		event2 := dom.Event{
			Type:   ptr("click"),
			Target: ptr("#button1"),
			Key:    ptr("btn1"),
		}

		events := []dom.Event{event1, event2}
		result := uniques(events)

		if len(result) != 1 {
			t.Errorf("Expected 1 event after deduplication, got %d", len(result))
		}
		// Should keep the second (newer) event
		if !reflect.DeepEqual(result[0], event2) {
			t.Errorf("Expected %+v, got %+v", event2, result[0])
		}
	})

	t.Run("MultipleDuplicates", func(t *testing.T) {
		event1 := dom.Event{
			Type:   ptr("click"),
			Target: ptr("#button1"),
			Key:    ptr("btn1"),
		}
		event2 := dom.Event{
			Type:   ptr("click"),
			Target: ptr("#button1"),
			Key:    ptr("btn1"),
		}
		event3 := dom.Event{
			Type:   ptr("click"),
			Target: ptr("#button1"),
			Key:    ptr("btn1"),
		}

		events := []dom.Event{event1, event2, event3}
		result := uniques(events)

		if len(result) != 1 {
			t.Errorf("Expected 1 event after deduplication, got %d", len(result))
		}
		// Should keep the last (newest) event
		if !reflect.DeepEqual(result[0], event3) {
			t.Errorf("Expected %+v, got %+v", event3, result[0])
		}
	})

	t.Run("PartialDuplicates", func(t *testing.T) {
		events := []dom.Event{
			{
				Type:   ptr("click"),
				Target: ptr("#button1"),
				Key:    ptr("btn1"),
			},
			{
				Type:   ptr("click"),
				Target: ptr("#button2"),
				Key:    ptr("btn2"),
			},
			{
				Type:   ptr("click"),
				Target: ptr("#button1"),
				Key:    ptr("btn1"),
			},
			{
				Type:   ptr("keydown"),
				Target: ptr("#input1"),
				Key:    ptr("inp1"),
			},
		}
		result := uniques(events)

		if len(result) != 3 {
			t.Errorf("Expected 3 events after deduplication, got %d", len(result))
		}

		// Check that we have the right number and they are unique
		// The function should replace duplicates with the later version
		found := make(map[string]bool)
		for _, event := range result {
			key := ""
			if event.Type != nil {
				key += *event.Type + "|"
			} else {
				key += "nil|"
			}
			if event.Target != nil {
				key += *event.Target + "|"
			} else {
				key += "nil|"
			}
			if event.Key != nil {
				key += *event.Key
			} else {
				key += "nil"
			}

			if found[key] {
				t.Errorf("Duplicate event found: %s", key)
			}
			found[key] = true
		}
	})

	t.Run("NilFields", func(t *testing.T) {
		events := []dom.Event{
			{
				Type:   ptr("click"),
				Target: nil,
				Key:    ptr("btn1"),
			},
			{
				Type:   nil,
				Target: ptr("#button1"),
				Key:    ptr("btn2"),
			},
			{
				Type:   ptr("click"),
				Target: ptr("#button1"),
				Key:    nil,
			},
		}
		result := uniques(events)

		if len(result) != 3 {
			t.Errorf("Expected 3 events (all different due to nil fields), got %d", len(result))
		}
	})

	t.Run("NilFieldDuplicates", func(t *testing.T) {
		event1 := dom.Event{
			Type:   nil,
			Target: nil,
			Key:    nil,
		}
		event2 := dom.Event{
			Type:   nil,
			Target: nil,
			Key:    nil,
		}

		events := []dom.Event{event1, event2}
		result := uniques(events)

		if len(result) != 1 {
			t.Errorf("Expected 1 event after deduplication (nil fields), got %d", len(result))
		}
		// Should keep the second event
		if !reflect.DeepEqual(result[0], event2) {
			t.Errorf("Expected %+v, got %+v", event2, result[0])
		}
	})

	t.Run("DifferentKeysOnly", func(t *testing.T) {
		events := []dom.Event{
			{
				Type:   ptr("click"),
				Target: ptr("#button1"),
				Key:    ptr("key1"),
			},
			{
				Type:   ptr("click"),
				Target: ptr("#button1"),
				Key:    ptr("key2"),
			},
		}
		result := uniques(events)

		if len(result) != 2 {
			t.Errorf("Expected 2 events (different keys), got %d", len(result))
		}
	})

	t.Run("DifferentTargetsOnly", func(t *testing.T) {
		events := []dom.Event{
			{
				Type:   ptr("click"),
				Target: ptr("#button1"),
				Key:    ptr("btn"),
			},
			{
				Type:   ptr("click"),
				Target: ptr("#button2"),
				Key:    ptr("btn"),
			},
		}
		result := uniques(events)

		if len(result) != 2 {
			t.Errorf("Expected 2 events (different targets), got %d", len(result))
		}
	})

	t.Run("DifferentTypesOnly", func(t *testing.T) {
		events := []dom.Event{
			{
				Type:   ptr("click"),
				Target: ptr("#button1"),
				Key:    ptr("btn"),
			},
			{
				Type:   ptr("keydown"),
				Target: ptr("#button1"),
				Key:    ptr("btn"),
			},
		}
		result := uniques(events)

		if len(result) != 2 {
			t.Errorf("Expected 2 events (different types), got %d", len(result))
		}
	})

	t.Run("LargeDataset", func(t *testing.T) {
		// Test with a larger dataset to ensure performance is reasonable
		events := make([]dom.Event, 1000)
		for i := 0; i < 1000; i++ {
			events[i] = dom.Event{
				Type:   ptr("click"),
				Target: ptr("#button"),
				Key:    ptr("btn"),
			}
		}

		result := uniques(events)

		if len(result) != 1 {
			t.Errorf("Expected 1 event after deduplication of 1000 identical events, got %d", len(result))
		}
	})

	t.Run("MixedScenario", func(t *testing.T) {
		// Complex scenario with various combinations
		events := []dom.Event{
			{Type: ptr("click"), Target: ptr("#btn1"), Key: ptr("k1")},    // unique
			{Type: ptr("click"), Target: ptr("#btn2"), Key: ptr("k2")},    // unique
			{Type: ptr("click"), Target: ptr("#btn1"), Key: ptr("k1")},    // duplicate of first
			{Type: ptr("keydown"), Target: ptr("#input"), Key: ptr("k3")}, // unique
			{Type: nil, Target: ptr("#btn3"), Key: ptr("k4")},             // unique (nil type)
			{Type: ptr("click"), Target: nil, Key: ptr("k5")},             // unique (nil target)
			{Type: ptr("click"), Target: ptr("#btn2"), Key: ptr("k2")},    // duplicate of second
			{Type: nil, Target: ptr("#btn3"), Key: ptr("k4")},             // duplicate of nil type
		}

		result := uniques(events)

		if len(result) != 5 {
			t.Errorf("Expected 5 unique events, got %d", len(result))
		}

		// Verify uniqueness by creating keys and ensuring no duplicates
		seen := make(map[string]bool)
		for _, event := range result {
			key := ""
			if event.Type != nil {
				key += *event.Type + "|"
			} else {
				key += "nil|"
			}
			if event.Target != nil {
				key += *event.Target + "|"
			} else {
				key += "nil|"
			}
			if event.Key != nil {
				key += *event.Key
			} else {
				key += "nil"
			}

			if seen[key] {
				t.Errorf("Found duplicate event with key: %s", key)
			}
			seen[key] = true
		}

		// Also verify specific expected unique combinations exist
		expectedKeys := []string{
			"click|#btn2|k2",
			"click|#btn1|k1",
			"keydown|#input|k3",
			"click|nil|k5",
			"nil|#btn3|k4",
		}

		for _, expectedKey := range expectedKeys {
			if !seen[expectedKey] {
				t.Errorf("Expected event key not found: %s", expectedKey)
			}
		}
	})
}
