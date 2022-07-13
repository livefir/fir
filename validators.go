package fir

import "fmt"

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
