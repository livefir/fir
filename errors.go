package fir

import (
	"errors"
	"fmt"
	"log"
	"strings"
)

var DefaultUserError = errors.New("internal error")

func MorphError(name string) (func(err error) Patch, func() Patch) {
	selector := fmt.Sprintf("#%s", name)
	return func(err error) Patch {
			return Morph(selector, Block(name, M{name: err}))
		}, func() Patch {
			return Morph(selector, Block(name, M{name: ""}))
		}
}

func morphFirErrors(eventID string) (func(err error) Patch, func() Patch) {
	id := fmt.Sprintf("fir-errors-%s", eventID)
	selector := fmt.Sprintf("#%s", id)
	return func(err error) Patch {
			return Morph(selector, Block(id, M{"fir": M{"errors": M{eventID: err.Error()}}}))
		}, func() Patch {
			return Morph(selector, Block(id, M{"fir": M{"errors": M{eventID: ""}}}))
		}
}

type fieldErrors map[string]error

func (f fieldErrors) Error() string {
	var errs []string
	for field, err := range f {
		errs = append(errs, fmt.Sprintf("%s: %s", field, err.Error()))
	}
	return strings.Join(errs, ", ")
}

func UserError(ctx Context, err error) error {
	log.Printf("error: %v", err)
	userError := DefaultUserError
	if wrappedUserError := errors.Unwrap(err); wrappedUserError != nil {
		userError = wrappedUserError
	}
	return userError
}
