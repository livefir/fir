package utils

import (
	"errors"
	"fmt"

	"github.com/adnaan/autobahn/models"
	"github.com/adnaan/fir"
)

func PageFormError(err error) fir.Page {
	var validError *models.ValidationError
	if errors.As(err, &validError) {
		userError := validError.Unwrap()
		if errors.Unwrap(validError.Unwrap()) != nil {
			userError = errors.Unwrap(validError.Unwrap())
		}
		return fir.Page{
			Data: fir.Data{
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
				Template: &fir.Template{
					Name: fmt.Sprintf("%s-error", validError.Name),
					Data: fir.Data{
						fmt.Sprintf("%sError", validError.Name): userError.Error(),
					},
				},
			},
		}
	}
	return fir.PatchError(err)
}
