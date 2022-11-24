package utils

import (
	"errors"
	"fmt"

	"github.com/adnaan/fir"
	"github.com/adnaan/fir/cli/testdata/todos/models"
)

func PageFormError(err error) fir.Pagedata {
	var validError *models.ValidationError
	if errors.As(err, &validError) {
		userError := validError.Unwrap()
		if errors.Unwrap(validError.Unwrap()) != nil {
			userError = errors.Unwrap(validError.Unwrap())
		}
		return fir.Pagedata{
			Data: map[string]any{
				fmt.Sprintf("%sError", validError.Name): userError.Error(),
			},
		}
	}
	return fir.PageError(err)
}

func PatchFormError(err error) fir.Patchset {
	var validError *models.ValidationError
	if errors.As(err, &validError) {
		userError := validError.Unwrap()
		if errors.Unwrap(validError.Unwrap()) != nil {
			userError = errors.Unwrap(validError.Unwrap())
		}

		return fir.Patchset{
			fir.Morph{
				Selector: fmt.Sprintf("#%s-error", validError.Name),
				HTML: &fir.Render{
					Template: fmt.Sprintf("%s-error", validError.Name),
					Data: map[string]any{
						fmt.Sprintf("%sError", validError.Name): userError.Error(),
					},
				},
			},
		}
	}
	return fir.PatchError(err)
}
