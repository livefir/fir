package fir

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var DefaultViewExtensions = []string{".gohtml", ".gotmpl", ".html", ".tmpl"}

type Page struct {
	Data    Data   `json:"data"`
	Code    int    `json:"statusCode"`
	Message string `json:"statusMessage"`
	Error   error  `json:"-"`
}

type AppContext struct {
	Name    string
	URLPath string
}

func (a *AppContext) ActiveRoute(path, class string) string {
	if a.URLPath == path {
		return class
	}
	return ""
}

func (a *AppContext) NotActiveRoute(path, class string) string {
	if a.URLPath != path {
		return class
	}
	return ""
}

type View interface {
	ID() string
	// Settings
	Content() string
	Layout() string
	LayoutContentName() string
	Partials() []string
	Extensions() []string
	FuncMap() template.FuncMap
	// Lifecyle
	OnGet(http.ResponseWriter, *http.Request) Page
	OnPost(http.ResponseWriter, *http.Request) Page
	OnEvent(event Event) Patchset
	Stream() <-chan Patch
}

type DefaultView struct{}

func (d DefaultView) ID() string {
	return ""
}

// Content returns either path to the content or a html string content
func (d DefaultView) Content() string {
	return ""
}

// Layout returns either path to the layout or a html string layout
// Layout represents the path to the base layout to be used.
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

func (d DefaultView) Stream() <-chan Patch {
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

// OnEvent handles the events sent from the browser
func (d DefaultErrorView) OnEvent(event Event) Patchset {
	switch event.ID {
	default:
		log.Printf("[defaultView] warning:handler not found for event => \n %+v\n", event)
	}
	return Patchset{}
}

func (d DefaultErrorView) Stream() <-chan Patch {
	return nil
}

type viewHandler struct {
	view              View
	errorView         View
	viewTemplate      *template.Template
	errorViewTemplate *template.Template
	mountData         Data
	cntrl             *controller
}

func (v *viewHandler) reloadTemplates() {
	var err error
	if v.cntrl.disableTemplateCache {

		v.viewTemplate, err = parseTemplate(v.cntrl.opt, v.view)
		if err != nil {
			panic(err)
		}

		v.errorViewTemplate, err = parseTemplate(v.cntrl.opt, v.errorView)
		if err != nil {
			panic(err)
		}
	}
}

func buildDOMPatch(t *template.Template, patch Patch) (Operation, error) {
	var op Op
	var patchTemplate *Template
	var selector string

	switch v := patch.(type) {
	case Morph:
		op = v.Op()
		patchTemplate = v.Template
		selector = v.Selector
	case After:
		op = v.Op()
		patchTemplate = v.Template
		selector = v.Selector
	case Before:
		op = v.Op()
		patchTemplate = v.Template
		selector = v.Selector
	case Append:
		op = v.Op()
		patchTemplate = v.Template
		selector = v.Selector
	case Prepend:
		op = v.Op()
		patchTemplate = v.Template
		selector = v.Selector
	case Remove:
		op = v.Op()
		patchTemplate = v.Template
		selector = v.Selector
	}

	if patchTemplate == nil {
		return Operation{}, fmt.Errorf("error: patch %v template is nil", patch.Op())
	}
	if selector == "" {
		return Operation{}, fmt.Errorf("error: patch %v selector is empty", patch.Op())
	}

	var buf bytes.Buffer
	err := t.ExecuteTemplate(&buf, patchTemplate.Name, patchTemplate.Data)
	if err != nil {
		// if s.wc.debugLog {
		// 	log.Printf("[controller][error] %v with data => \n %+v\n", err, getJSON(data))
		// }
		return Operation{}, err
	}
	// if s.wc.debugLog {
	// 	log.Printf("[controller]rendered template %+v, with data => \n %+v\n", tmpl, getJSON(data))
	// }
	html := buf.String()
	buf.Reset()
	return Operation{
		Op:       op,
		Selector: selector,
		Value:    html,
	}, nil
}

func buildStorePatch(patch Patch) (Operation, error) {
	storePatch := patch.(Store)
	if storePatch.Name == "" {
		return Operation{}, fmt.Errorf("error: patch %v name is empty", patch.Op())
	}
	if storePatch.Data == nil {
		return Operation{}, fmt.Errorf("error: patch %v data is nil", patch.Op())
	}

	return Operation{
		Op:       updateStore,
		Selector: storePatch.Name,
		Value:    storePatch.Data,
	}, nil
}

func buildOperation(t *template.Template, patch Patch) (Operation, error) {
	switch patch.Op() {
	case morph, after, before, appendOp, prepend, remove:
		operation, err := buildDOMPatch(t, patch)
		if err != nil {
			return Operation{}, err
		}
		return operation, nil
	case reload:
		return Operation{Op: reload}, nil
	case resetForm:
		p := patch.(ResetForm)
		return Operation{Op: resetForm, Selector: p.Selector}, nil
	case navigate:
		p := patch.(Navigate)
		return Operation{Op: navigate, Value: p.To}, nil
	case updateStore:
		return buildStorePatch(patch)
	default:
		return Operation{}, fmt.Errorf("operation unknown")
	}

}

func layoutSetContentEmpty(opt opt, view View) (*template.Template, error) {
	viewLayoutPath := filepath.Join(opt.publicDir, view.Layout())
	// is layout html content or a file/directory
	if isFileHTML(viewLayoutPath, opt) {
		return template.Must(template.New("").Funcs(view.FuncMap()).Parse(view.Layout())), nil
	}

	// layout must be  a file or directory
	if !isDir(viewLayoutPath, opt) {
		return nil, fmt.Errorf("layout %s is not a file or directory", viewLayoutPath)
	}
	// compile layout
	commonFiles := []string{viewLayoutPath}
	// global partials
	for _, p := range view.Partials() {
		commonFiles = append(commonFiles, find(opt, filepath.Join(opt.publicDir, p), view.Extensions())...)
	}

	layoutTemplate := template.New(viewLayoutPath).Funcs(view.FuncMap())
	if opt.hasEmbedFS {
		layoutTemplate = template.Must(layoutTemplate.ParseFS(opt.embedFS, commonFiles...))
	} else {
		layoutTemplate = template.Must(layoutTemplate.ParseFiles(commonFiles...))
	}

	return template.Must(layoutTemplate.Clone()), nil
}

func layoutEmptyContentSet(opt opt, view View) (*template.Template, error) {
	// is content html content or a file/directory
	viewContentPath := filepath.Join(opt.publicDir, view.Content())
	if isFileHTML(viewContentPath, opt) {
		return template.Must(
			template.New(
				view.LayoutContentName()).
				Funcs(view.FuncMap()).
				Parse(view.Content()),
		), nil
	}
	// content must be  a file or directory

	var pageFiles []string
	// view and its partials
	pageFiles = append(pageFiles, find(opt, viewContentPath, view.Extensions())...)
	for _, p := range view.Partials() {
		pageFiles = append(pageFiles, find(opt, filepath.Join(opt.publicDir, p), view.Extensions())...)
	}

	contentTemplate := template.New(viewContentPath).Funcs(view.FuncMap())
	if opt.hasEmbedFS {
		contentTemplate = template.Must(contentTemplate.ParseFS(opt.embedFS, pageFiles...))
	} else {
		contentTemplate = template.Must(contentTemplate.ParseFiles(pageFiles...))
	}

	return contentTemplate, nil
}

func layoutSetContentSet(opt opt, view View) (*template.Template, error) {
	// 1. build layout template
	viewLayoutPath := filepath.Join(opt.publicDir, view.Layout())
	var layoutTemplate *template.Template
	// is layout,  html content or a file/directory
	if isFileHTML(viewLayoutPath, opt) {
		layoutTemplate = template.Must(template.New("base").Funcs(view.FuncMap()).Parse(view.Layout()))
	} else {
		// layout must be  a file or directory
		if isDir(viewLayoutPath, opt) {
			return nil, fmt.Errorf("layout %s is a directory but must be a file", viewLayoutPath)
		}

		// compile layout
		commonFiles := []string{viewLayoutPath}
		// global partials
		for _, p := range view.Partials() {
			commonFiles = append(commonFiles, find(opt, filepath.Join(opt.publicDir, p), view.Extensions())...)
		}

		layoutTemplate = template.New(filepath.Base(viewLayoutPath)).Funcs(view.FuncMap())
		if opt.hasEmbedFS {
			layoutTemplate = template.Must(layoutTemplate.ParseFS(opt.embedFS, commonFiles...))
		} else {
			layoutTemplate = template.Must(layoutTemplate.ParseFiles(commonFiles...))
		}
	}

	//log.Println("compiled layoutTemplate...")
	//for _, v := range layoutTemplate.Templates() {
	//	fmt.Println("template => ", v.Name())
	//}

	// 2. add content to layout
	// check if content is a not a file or directory
	var viewTemplate *template.Template
	viewContentPath := filepath.Join(opt.publicDir, view.Content())
	if isFileHTML(viewContentPath, opt) {
		viewTemplate = template.Must(layoutTemplate.Parse(view.Content()))
	} else {
		var pageFiles []string
		// view and its partials
		pageFiles = append(pageFiles, find(opt, filepath.Join(opt.publicDir, view.Content()), view.Extensions())...)
		if opt.hasEmbedFS {
			viewTemplate = template.Must(layoutTemplate.ParseFS(opt.embedFS, pageFiles...))
		} else {
			viewTemplate = template.Must(layoutTemplate.ParseFiles(pageFiles...))
		}
	}

	// check if the final viewTemplate contains a content child template which is `content` by default.
	if ct := viewTemplate.Lookup(view.LayoutContentName()); ct == nil {
		return nil,
			fmt.Errorf("err looking up layoutContent: the layout %s expects a template named %s",
				view.Layout(), view.LayoutContentName())
	}

	return viewTemplate, nil
}

// creates a html/template from the View type.
func parseTemplate(opt opt, view View) (*template.Template, error) {
	// if both layout and content is empty show a default view.
	if view.Layout() == "" && view.Content() == "" {
		return template.Must(template.New("").
			Parse(`<div style="text-align:center"> This is a default view. </div>`)), nil
	}

	// if layout is set and content is empty
	if view.Layout() != "" && view.Content() == "" {
		return layoutSetContentEmpty(opt, view)
	}

	// if layout is empty and content is set
	if view.Layout() == "" && view.Content() != "" {
		return layoutEmptyContentSet(opt, view)
	}

	// both layout and content are set
	return layoutSetContentSet(opt, view)
}

var DefaultUserErrorMessage = "internal error"

func UserError(err error) string {
	userMessage := DefaultUserErrorMessage
	if userError := errors.Unwrap(err); userError != nil {
		userMessage = userError.Error()
	}
	return userMessage
}

func find(opt opt, p string, extensions []string) []string {
	var files []string
	var fi fs.FileInfo
	var err error

	if opt.hasEmbedFS {
		fi, err = fs.Stat(opt.embedFS, p)
		if err != nil {
			return files
		}
	} else {
		fi, err = os.Stat(p)
		if err != nil {
			return files
		}
	}

	if !fi.IsDir() {
		if !contains(extensions, filepath.Ext(p)) {
			return files
		}
		files = append(files, p)
		return files
	}

	if opt.hasEmbedFS {
		err = fs.WalkDir(opt.embedFS, p, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if contains(extensions, filepath.Ext(d.Name())) {
				files = append(files, path)
			}
			return nil
		})

		if err != nil {
			panic(err)
		}

	} else {

		err = filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if contains(extensions, filepath.Ext(d.Name())) {
				files = append(files, path)
			}
			return nil
		})

		if err != nil {
			panic(err)
		}

	}

	return files
}

func contains(arr []string, s string) bool {
	for _, a := range arr {
		if a == s {
			return true
		}
	}
	return false
}

func isDir(path string, opt opt) bool {
	if opt.hasEmbedFS {
		fileInfo, err := fs.Stat(opt.embedFS, path)
		if err != nil {
			fmt.Println("[warning]isDir warn: ", err)
			return false
		}
		return fileInfo.IsDir()
	}
	fileInfo, err := os.Stat(path)
	if err != nil {
		fmt.Println("[warning]isDir error: ", err)
		return false
	}

	return fileInfo.IsDir()
}

func isFileHTML(path string, opt opt) bool {
	if opt.hasEmbedFS {
		if _, err := fs.Stat(opt.embedFS, path); err != nil {
			return true
		}
		return false
	}
	if _, err := os.Stat(path); err != nil {
		return true
	}
	return false
}
