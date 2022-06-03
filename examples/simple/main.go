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
	glvc := fir.Websocket("fir-simple", fir.DevelopmentMode(true))
	http.Handle("/", glvc.Handler(&SimpleView{}))
	http.ListenAndServe(":9867", nil)
}
