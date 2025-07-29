package fir

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_hashID(t *testing.T) {
	// Test normal case
	result1 := hashID("test content")
	require.NotEmpty(t, result1)
	require.Len(t, result1, 16) // xxhash produces 8 bytes, hex encoded = 16 chars

	// Test with whitespace (should be removed)
	result2 := hashID("  test  content  ")
	require.Equal(t, result1, result2) // Should be same after space removal

	// Test empty string
	result3 := hashID("")
	require.NotEmpty(t, result3)
	require.Len(t, result3, 16)

	// Test consistency
	result4 := hashID("test content")
	require.Equal(t, result1, result4)
}
