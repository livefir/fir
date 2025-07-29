package fir

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_formatValue(t *testing.T) {
	// Test with normal format (2 parts)
	result := formatValue("part1:part2")
	require.Equal(t, "part1:part2", result)

	// Test with single part (should return as-is)
	result = formatValue("singlepart")
	require.Equal(t, "singlepart", result)

	// Test with empty string
	result = formatValue("")
	require.Equal(t, "", result)

	// Test with more than 2 parts (should return as-is)
	result = formatValue("part1:part2:part3")
	require.Equal(t, "part1:part2:part3", result)

	// Test with only colon
	result = formatValue(":")
	require.Equal(t, ":", result)
}
