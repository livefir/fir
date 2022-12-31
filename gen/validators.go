package gen

import (
	"fmt"

	firErrors "github.com/livefir/fir/internal/errors"
)

// MinMaxField is a validator that checks if the value is between min and max. Can be used in an entgo schema.
// It returns a FieldError which can be used in the template as {{.fir.Error "myevent.field"}}
func MinMaxLenField(field string, minLen, maxLen int) func(s string) error {
	return func(s string) error {
		if len(s) < minLen {
			return firErrors.Fields{field: fmt.Errorf("%s must be at least %v characters ", field, minLen)}
		}

		if len(s) > maxLen {
			return firErrors.Fields{field: fmt.Errorf("%s must be at most %v characters ", field, maxLen)}
		}
		return nil
	}
}
