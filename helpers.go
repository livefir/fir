package fir

import (
	"net/http"

	"github.com/gorilla/schema"
)

var decoder = schema.NewDecoder()

// DecodeForm decodes the form values from the request into the destination struct.
func DecodeForm(dst interface{}, r *http.Request) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	err = decoder.Decode(dst, r.PostForm)
	if err != nil {
		return err
	}

	return nil
}

// DecodeURLValues decodes the url values from the request into the destination struct.
func DecodeURLValues(dst interface{}, r *http.Request) error {
	err := decoder.Decode(dst, r.URL.Query())
	if err != nil {
		return err
	}

	return nil
}
