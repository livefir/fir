package main

import (
	"net/http"

	pwc "github.com/adnaan/fir/controller"
)

type SimpleView struct {
	pwc.DefaultView
}

func (s *SimpleView) Content() string {
	return `{{define "content"}}<div>world</div>{{ end }}`
}

func (s *SimpleView) Layout() string {
	return `<div>Hello: {{template "content" .}}</div>`
}

func main() {
	glvc := pwc.Websocket("fir-simple", pwc.DevelopmentMode(true))
	http.Handle("/", glvc.Handler(&SimpleView{}))
	http.ListenAndServe(":9867", nil)
}
