package fir

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_fir_Panics(t *testing.T) {
	// Test panic with no arguments
	require.Panics(t, func() {
		fir()
	})

	// Test panic with more than 2 arguments
	require.Panics(t, func() {
		fir("a", "b", "c")
	})
}

func Test_fir_EdgeCases(t *testing.T) {
	// Test with empty strings
	result := fir("")
	require.NotNil(t, result)
	require.Equal(t, "fir:", *result)

	result = fir("", "")
	require.NotNil(t, result)
	require.Equal(t, "fir:::", *result)

	// Test normal cases to ensure full coverage
	result = fir("test")
	require.NotNil(t, result)
	require.Equal(t, "fir:test", *result)

	result = fir("test", "template")
	require.NotNil(t, result)
	require.Equal(t, "fir:test::template", *result)
}
