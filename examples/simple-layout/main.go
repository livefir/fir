package main

import (
	"net/http"

	"github.com/adnaan/fir"
)

type LayoutView struct {
	fir.DefaultView
}

func (l *LayoutView) Content() string {
	return `{{define "content"}}<div>world</div>{{ end }}`
}

func (l *LayoutView) Layout() string {
	return `<div>Hello: {{template "content" .}}</div>`
}

func main() {
	c := fir.NewController("fir-simple", fir.DevelopmentMode(true))
	http.Handle("/", c.Handler(&LayoutView{}))
	http.ListenAndServe(":9867", nil)
}
