package fir

import (
	"html/template"
	"log"
	"net/http"
)

// DefaultView is the default implementation of the View interface and can be used as a base for custom views.
// It can be embedded in a custom view to inherit the default behaviour. Typically, the Content and Layout methods
// are overridden to provide the content and layout for the view.
type DefaultView struct{}

func (d DefaultView) ID() string {
	return ""
}

// Content returns either path to the content or a html string content
func (d DefaultView) Content() string {
	return ""
}

// Layout returns either path to the layout or a html string layout
/*
		layout.html e.g.
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<title>{{.app_name}}</title>
			{{template "header" .}}
		</head>
		<body>
		{{template "navbar" .}}
		<div>
			{{template "content" .}}
		</div>
		{{template "footer" .}}
		</body>
		</html>
	 The {{template "content" .}} directive is replaced by the page in the path exposed by `Content`
*/
func (d DefaultView) Layout() string {
	return ""
}

// LayoutContentName is the defined template of the "content" to be replaced. Defaults to "content"
/*
e.g.
type SimpleView struct {
	fir.DefaultView
}

func (s *SimpleView) Content() string {
	return `{{define "content"}}<div>world</div>{{ end }}`
}

func (s *SimpleView) Layout() string {
	return `<div>Hello: {{template "content" .}}</div>`
}
*/
func (d DefaultView) LayoutContentName() string {
	return "content"
}

// Partials returns path to any partials used in the view. Defaults to "./templates/partials"
func (d DefaultView) Partials() []string {
	return []string{"./templates/partials"}
}

// Extensions returns the view extensions. Defaults to: html, tmpl, gohtml, gotmpl
func (d DefaultView) Extensions() []string {
	return DefaultViewExtensions
}

// FuncMap configures the html/template.FuncMap
func (d DefaultView) FuncMap() template.FuncMap {
	return DefaultFuncMap()
}

// OnGet is called when the page is first loaded for the http route.
func (d DefaultView) OnGet(w http.ResponseWriter, r *http.Request) Page {
	return Page{Code: 200, Message: "OK"}
}

// OnPost is called when a form is submitted for the http route.
func (d DefaultView) OnPost(w http.ResponseWriter, r *http.Request) Page {
	return Page{Code: 405, Message: "method not allowed"}
}

// OnEvent handles the events sent from the browser
func (d DefaultView) OnEvent(event Event) Patchset {
	switch event.ID {
	default:
		log.Printf("[defaultView] warning:handler not found for event => \n %+v\n", event)
	}
	return Patchset{}
}

func (d DefaultView) Publisher() <-chan Patchset {
	return nil
}

type DefaultErrorView struct{}

func (d DefaultErrorView) ID() string {
	return ""
}

func (d DefaultErrorView) Content() string {
	return `
{{ define "content"}}
    <div style="text-align:center"><h1>{{.statusCode}}</h1></div>
    <div style="text-align:center"><h1>{{.statusMessage}}</h1></div>
    <div style="text-align:center"><a href="javascript:history.back()">back</a></div>
    <div style="text-align:center"><a href="/">home</a></div>
{{ end }}`
}

func (d DefaultErrorView) Layout() string {
	return ""
}

func (d DefaultErrorView) LayoutContentName() string {
	return "content"
}

func (d DefaultErrorView) Partials() []string {
	return []string{}
}

func (d DefaultErrorView) Extensions() []string {
	return DefaultViewExtensions
}

func (d DefaultErrorView) FuncMap() template.FuncMap {
	return DefaultFuncMap()
}

func (d DefaultErrorView) OnGet(w http.ResponseWriter, r *http.Request) Page {
	return Page{Code: 500, Message: "Internal Server Error"}
}

func (d DefaultErrorView) OnPost(w http.ResponseWriter, r *http.Request) Page {
	return Page{Code: 405, Message: "method not allowed"}
}

func (d DefaultErrorView) OnEvent(event Event) Patchset {
	switch event.ID {
	default:
		log.Printf("[defaultView] warning:handler not found for event => \n %+v\n", event)
	}
	return Patchset{}
}

func (d DefaultErrorView) Publisher() <-chan Patchset {
	return nil
}
