package fir

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFirEventState_IsValid tests the IsValid method of FirEventState
func TestFirEventState_IsValid(t *testing.T) {
	t.Run("valid states return true", func(t *testing.T) {
		validStates := []FirEventState{
			StateOK,
			StateError,
			StatePending,
			StateDone,
		}

		for _, state := range validStates {
			result := state.IsValid()
			assert.True(t, result, "State %s should be valid", state)
		}
	})

	t.Run("invalid states return false", func(t *testing.T) {
		invalidStates := []FirEventState{
			"invalid",
			"unknown",
			"completed",
			"failed",
			"success",
			"",
			"OK", // case sensitive
			"ERROR",
			"PENDING",
			"DONE",
		}

		for _, state := range invalidStates {
			result := state.IsValid()
			assert.False(t, result, "State %s should be invalid", state)
		}
	})

	t.Run("state constants have expected values", func(t *testing.T) {
		assert.Equal(t, FirEventState("ok"), StateOK)
		assert.Equal(t, FirEventState("error"), StateError)
		assert.Equal(t, FirEventState("pending"), StatePending)
		assert.Equal(t, FirEventState("done"), StateDone)
	})

	t.Run("all defined states are valid", func(t *testing.T) {
		// Ensure all the predefined constants are valid
		assert.True(t, StateOK.IsValid())
		assert.True(t, StateError.IsValid())
		assert.True(t, StatePending.IsValid())
		assert.True(t, StateDone.IsValid())
	})

	t.Run("custom states are invalid", func(t *testing.T) {
		customStates := []FirEventState{
			"custom",
			"user-defined",
			"maybe",
			"working",
		}

		for _, state := range customStates {
			result := state.IsValid()
			assert.False(t, result, "Custom state %s should be invalid", state)
		}
	})
}
