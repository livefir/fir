package fir

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

func getUserMessage(status int, userMessage []string) string {
	msg := http.StatusText(status)
	if len(userMessage) > 0 {
		msg = strings.Join(userMessage, " ")
	}
	return msg
}

func morphError(err string) Patch {
	return Morph("#fir-error", Template("fir-error", M{"error": err}))
}

// PatchError returns a patchset that sets an error for selector #fir-error.
func PatchError(err error, userMessage ...string) []Patch {
	msg := "internal error"
	if err != nil && len(userMessage) == 0 {
		msg = err.Error()
		log.Printf("[controller] patch error: %s\n", err)
	}
	if len(userMessage) > 0 {
		msg = strings.Join(userMessage, " ")
		log.Printf("[controller] patch error: %s, message: %s\n", err, msg)
	}
	log.Printf("[controller] patch error: %s, %s\n", err, msg)
	return []Patch{morphError(msg)}
}

// PageError returns a Page with an error.
func PageError(err error, userMessage ...string) M {
	msg := "internal error"
	if err != nil && len(userMessage) == 0 {
		msg = err.Error()
		log.Printf("[controller] page error: %s\n", err)
	}
	if len(userMessage) > 0 {
		msg = strings.Join(userMessage, " ")
		log.Printf("[controller] page error: %s, message: %s\n", err, msg)
	}

	data := map[string]any{"error": msg}
	if msg == "" {
		data = nil
	}
	return data
}

// UnsetPatchFormErrors returns a patchset that unsets the error for a form.
func UnsetPatchFormErrors(fields ...string) []Patch {
	var patchset []Patch
	for _, field := range fields {
		m := Morph(
			fmt.Sprintf("#%s-error", field),
			Template(fmt.Sprintf("%s-error", field), M{fmt.Sprintf("#%sError", field): ""}),
		)
		patchset = append(patchset, m)
	}

	return patchset
}

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
