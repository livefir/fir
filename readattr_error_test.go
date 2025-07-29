package fir

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEventFormatError tests the eventFormatError function
func TestEventFormatError(t *testing.T) {
	t.Run("formats error message with event namespace", func(t *testing.T) {
		eventns := "invalid:format"

		result := eventFormatError(eventns)

		// Should contain the provided event namespace
		assert.Contains(t, result, eventns)

		// Should contain error description
		assert.Contains(t, result, "error: invalid event namespace")

		// Should contain format examples
		assert.Contains(t, result, "@fir:<event>:<state:ok|error|pending|done>::<block-name(optional)>")
		assert.Contains(t, result, "@fir:[event1:state,event2:state]::<block-name(optional)>")
	})

	t.Run("formats error message with empty event namespace", func(t *testing.T) {
		eventns := ""

		result := eventFormatError(eventns)

		// Should still contain basic error structure
		assert.Contains(t, result, "error: invalid event namespace")
		assert.Contains(t, result, "must be of either of the three formats")
	})

	t.Run("formats error message with complex event namespace", func(t *testing.T) {
		eventns := "@fir:complex:invalid:format:with:many:parts"

		result := eventFormatError(eventns)

		// Should contain the full complex namespace
		assert.Contains(t, result, eventns)

		// Should contain helpful format guidance
		assert.Contains(t, result, "1. @fir:")
		assert.Contains(t, result, "2. @fir:")
	})

	t.Run("error message contains newlines for readability", func(t *testing.T) {
		eventns := "test"

		result := eventFormatError(eventns)

		// Should contain newlines for formatting
		assert.True(t, strings.Contains(result, "\n"), "Error message should contain newlines for formatting")
	})

	t.Run("error message explains valid formats", func(t *testing.T) {
		eventns := "malformed"

		result := eventFormatError(eventns)

		// Should explain the valid state options
		assert.Contains(t, result, "ok|error|pending|done")

		// Should show optional block name syntax
		assert.Contains(t, result, "block-name(optional)")

		// Should show event grouping syntax
		assert.Contains(t, result, "[event1:state,event2:state]")
	})
}
