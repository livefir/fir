package fir

import "fmt"

// MinMax is a validator that checks if the value is between min and max. Can be used in an entgo schema.
func MinMax(minLen, maxLen int) func(s string) error {
	return func(s string) error {
		if len(s) < minLen {
			return fmt.Errorf("must be at least %v characters ", minLen)
		}

		if len(s) > maxLen {
			return fmt.Errorf("must be at most %v characters ", maxLen)
		}
		return nil
	}
}
