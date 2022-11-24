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
	return Morph("#fir-error", "fir-error", M{"error": err})
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
			fmt.Sprintf("%s-error", field),
			M{fmt.Sprintf("#%sError", field): ""},
		)
		patchset = append(patchset, m)
	}

	return patchset
}
