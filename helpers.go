package fir

import (
	"net/http"

	"github.com/gorilla/schema"
)

var decoder = schema.NewDecoder()

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

func DecodeURLValues(dst interface{}, r *http.Request) error {
	err := decoder.Decode(dst, r.URL.Query())
	if err != nil {
		return err
	}

	return nil
}
