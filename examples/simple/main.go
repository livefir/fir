package main

import (
	"net/http"

	"github.com/adnaan/fir"
)

type SimpleView struct {
	fir.DefaultView
}

func (s *SimpleView) Content() string {
	return `{{define "content"}}<div>world</div>{{ end }}`
}

func (s *SimpleView) Layout() string {
	return `<div>Hello: {{template "content" .}}</div>`
}

func main() {
	c := fir.NewController("fir-simple", fir.DevelopmentMode(true))
	http.Handle("/", c.Handler(&SimpleView{}))
	http.ListenAndServe(":9867", nil)
}
