package fir

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/lithammer/shortuuid"
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
	OnTopicCreated(string)
	OnTopicDestroyed(string)
	Stream() <-chan Patch
}

type DefaultView struct{}

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

// OnTopicCreated is called when the first client subscribes for websocket connection.
func (d DefaultView) OnTopicCreated(topic string) {
	log.Println("topic created", topic)
}

// OnTopicDestroyed is called when the last client unsubscribes from websocket connection.
func (d DefaultView) OnTopicDestroyed(topic string) {
	log.Println("topic destroyed", topic)
}

func (d DefaultView) Stream() <-chan Patch {
	return nil
}

type DefaultErrorView struct{}

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

// OnTopicCreated is called when the first client subscribes for websocket connection.
func (d DefaultErrorView) OnTopicCreated(topic string) {
	log.Println("topic created", topic)
}

// OnTopicDestroyed is called when the last client unsubscribes from websocket connection.
func (d DefaultErrorView) OnTopicDestroyed(topic string) {
	log.Println("topic destroyed", topic)
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
	user              int
	wc                *websocketController
}

func (v *viewHandler) reloadTemplates() {
	var err error
	if v.wc.disableTemplateCache {

		v.viewTemplate, err = parseTemplate(v.wc.projectRoot, v.view)
		if err != nil {
			panic(err)
		}

		v.errorViewTemplate, err = parseTemplate(v.wc.projectRoot, v.errorView)
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

func onPatchEvent(w http.ResponseWriter, r *http.Request, v *viewHandler) {
	v.reloadTemplates()
	var event Event
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&event)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if decoder.More() {
		http.Error(w, "unknown fields in request body", http.StatusBadRequest)
		return
	}
	event.requestContext = r.Context()
	patchset := v.view.OnEvent(event)
	if patchset == nil {
		log.Printf("[view] warning: no patchset returned for event: %v\n", event)
		patchset = Patchset{}
	}

	firErrorPatchExists := false

	for _, patch := range patchset {
		if patch.GetSelector() == "#fir-error" {
			firErrorPatchExists = true
		}
	}

	if !firErrorPatchExists {
		// unset error patch
		patchset = append([]Patch{morphError("")}, patchset...)
	}

	operations := make([]Operation, 0)
	for _, patch := range patchset {
		operation, err := buildOperation(v.viewTemplate, patch)
		if err != nil {
			if strings.ContainsAny("fir-error", err.Error()) {
				continue
			}
			log.Printf("[view] buildOperation error: %v\n", err)
			continue
		}

		operations = append(operations, operation)
	}
	json.NewEncoder(w).Encode(operations)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func onRequest(w http.ResponseWriter, r *http.Request, v *viewHandler) {
	v.reloadTemplates()

	var err error
	var page Page

	if r.Method == "POST" {
		page = v.view.OnPost(w, r)
	} else {
		page = v.view.OnGet(w, r)
	}
	v.mountData = page.Data
	if v.mountData == nil {
		v.mountData = make(Data)
	}

	v.mountData["app_name"] = v.wc.name
	v.mountData["fir"] = &AppContext{
		Name:    v.wc.name,
		URLPath: r.URL.Path,
	}

	page.Data = v.mountData

	if page.Code == 0 {
		page.Code = http.StatusOK
	}
	if page.Message == "" {
		page.Message = http.StatusText(page.Code)
	}

	if page.Code > 299 {
		log.Printf("page error: %v\n", page.Error)
		onRequestError(w, r, v, &page)
		return
	}

	v.viewTemplate.Option("missingkey=zero")
	var buf bytes.Buffer
	err = v.viewTemplate.Execute(&buf, page.Data)
	if err != nil {
		log.Printf("OnGet viewTemplate.Execute error:  %v", err)
		onRequestError(w, r, v, nil)
	}
	if v.wc.debugLog {
		log.Printf("OnGet render view %+v, with data => \n %+v\n",
			v.view.Content(), getJSON(page.Data))
	}

	w.WriteHeader(page.Code)
	w.Write(buf.Bytes())

}

func onRequestError(w http.ResponseWriter, r *http.Request, v *viewHandler, page *Page) {
	errorPage := v.errorView.OnGet(w, r)
	if page == nil {
		page = &errorPage
	}
	v.mountData = page.Data
	if v.mountData == nil {
		v.mountData = make(Data)
	}
	v.mountData["statusCode"] = page.Code
	v.mountData["statusMessage"] = page.Message

	page.Data = v.mountData

	v.viewTemplate.Option("missingkey=zero")
	var buf bytes.Buffer
	err := v.errorViewTemplate.Execute(&buf, page.Data)
	if err != nil {
		log.Printf("err rendering error template: %v\n", err)
		_, errWrite := w.Write([]byte("Something went wrong"))
		if errWrite != nil {
			panic(errWrite)
		}
	}

	if v.wc.debugLog {
		log.Printf("OnGet render error view %+v, with data => \n %+v\n",
			v.view.Content(), getJSON(page.Data))
	}

	w.WriteHeader(page.Code)
	w.Write(buf.Bytes())

}

func onWebsocket(w http.ResponseWriter, r *http.Request, v *viewHandler) {
	var topic *string
	if v.wc.subscribeTopicFunc != nil {
		topic = v.wc.subscribeTopicFunc(r)
	}

	c, err := v.wc.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()

	connID := shortuuid.New()
	if topic != nil {
		if created := v.wc.addConnection(*topic, connID, c); created {
			v.view.OnTopicCreated(*topic)
		}
	}

	topicVal := ""
	if topic != nil {
		topicVal = *topic
	}

	done := make(chan struct{})

	if v.view.Stream() != nil {
		go func() {
			for {
				select {
				case patch := <-v.view.Stream():
					operation, err := buildOperation(v.viewTemplate, patch)
					if err != nil {
						continue
					}
					v.wc.writeJSON(topicVal, []Operation{operation})
				case <-done:
					return
				}
			}

		}()
	}

loop:
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("c.readMessage error: ", err)
			break loop
		}

		event := new(Event)
		err = json.NewDecoder(bytes.NewReader(message)).Decode(event)
		if err != nil {
			log.Printf("err: parsing event, msg %s \n", string(message))
			continue
		}

		if event.ID == "" {
			log.Printf("err: event %v, field event.id is required\n", event)
			continue
		}

		v.reloadTemplates()
	}
	if v.view.Stream() != nil {
		done <- struct{}{}
	}
	if topic != nil {
		if destroyed := v.wc.removeConnection(*topic, connID); destroyed {
			v.view.OnTopicDestroyed(*topic)
		}
	}
}

// creates a html/template from the View type.
func parseTemplate(projectRoot string, view View) (*template.Template, error) {
	// if both layout and content is empty show a default view.
	if view.Layout() == "" && view.Content() == "" {
		return template.Must(template.New("").
			Parse(`<div style="text-align:center"> This is a default view. </div>`)), nil
	}

	// if layout is set and content is empty
	if view.Layout() != "" && view.Content() == "" {
		var layoutTemplate *template.Template
		// check if layout is not a file or directory
		if _, err := os.Stat(filepath.Join(projectRoot, view.Layout())); err != nil {
			// is not a file but html content
			layoutTemplate = template.Must(template.New("").Funcs(view.FuncMap()).Parse(view.Layout()))
		} else {
			// layout must be a file
			viewLayoutPath := filepath.Join(projectRoot, view.Layout())
			ok, err := isDirectory(viewLayoutPath)
			if err == nil && ok {
				return nil, fmt.Errorf("layout is a directory but it must be a file")
			}

			if err != nil {
				return nil, err
			}
			// compile layout
			commonFiles := []string{viewLayoutPath}
			// global partials
			for _, p := range view.Partials() {
				commonFiles = append(commonFiles, find(filepath.Join(projectRoot, p), view.Extensions())...)
			}
			layoutTemplate = template.Must(template.New(viewLayoutPath).
				Funcs(view.FuncMap()).
				ParseFiles(commonFiles...))
		}
		return template.Must(layoutTemplate.Clone()), nil
	}

	// if layout is empty and content is set
	if view.Layout() == "" && view.Content() != "" {
		// check if content is a not a file or directory
		if _, err := os.Stat(filepath.Join(projectRoot, view.Content())); err != nil {
			return template.Must(
				template.New(
					view.LayoutContentName()).
					Funcs(view.FuncMap()).
					Parse(view.Content()),
			), nil
		} else {
			// is a file or directory
			viewContentPath := filepath.Join(projectRoot, view.Content())
			var pageFiles []string
			// view and its partials
			pageFiles = append(pageFiles, find(viewContentPath, view.Extensions())...)
			for _, p := range view.Partials() {
				pageFiles = append(pageFiles, find(filepath.Join(projectRoot, p), view.Extensions())...)
			}
			return template.Must(template.New(filepath.Base(viewContentPath)).
				Funcs(view.FuncMap()).
				ParseFiles(pageFiles...)), nil
		}
	}

	// if both layout and content are set
	var viewTemplate *template.Template
	// 1. build layout
	var layoutTemplate *template.Template
	// check if layout is not a file or directory
	if _, err := os.Stat(filepath.Join(projectRoot, view.Layout())); err != nil {
		// is not a file but html content
		layoutTemplate = template.Must(template.New("base").Funcs(view.FuncMap()).Parse(view.Layout()))
	} else {
		// layout must be a file
		viewLayoutPath := filepath.Join(projectRoot, view.Layout())
		ok, err := isDirectory(viewLayoutPath)
		if err == nil && ok {
			return nil, fmt.Errorf("layout is a directory but it must be a file")
		}

		if err != nil {
			return nil, err
		}
		// compile layout
		commonFiles := []string{viewLayoutPath}
		// global partials
		for _, p := range view.Partials() {
			commonFiles = append(commonFiles, find(filepath.Join(projectRoot, p), view.Extensions())...)
		}
		layoutTemplate = template.Must(
			template.New(filepath.Base(viewLayoutPath)).
				Funcs(view.FuncMap()).
				ParseFiles(commonFiles...))

		//log.Println("compiled layoutTemplate...")
		//for _, v := range layoutTemplate.Templates() {
		//	fmt.Println("template => ", v.Name())
		//}
	}

	// 2. add content
	// check if content is a not a file or directory
	if _, err := os.Stat(filepath.Join(projectRoot, view.Content())); err != nil {
		// content is not a file or directory but html content
		viewTemplate = template.Must(layoutTemplate.Parse(view.Content()))
	} else {
		// content is a file or directory
		var pageFiles []string
		// view and its partials
		pageFiles = append(pageFiles, find(filepath.Join(projectRoot, view.Content()), view.Extensions())...)

		viewTemplate = template.Must(layoutTemplate.ParseFiles(pageFiles...))
	}

	// check if the final viewTemplate contains a content child template which is `content` by default.
	if ct := viewTemplate.Lookup(view.LayoutContentName()); ct == nil {
		return nil,
			fmt.Errorf("err looking up layoutContent: the layout %s expects a template named %s",
				view.Layout(), view.LayoutContentName())
	}

	return viewTemplate, nil
}

var DefaultUserErrorMessage = "internal error"

func UserError(err error) string {
	userMessage := DefaultUserErrorMessage
	if userError := errors.Unwrap(err); userError != nil {
		userMessage = userError.Error()
	}
	return userMessage
}

func find(p string, extensions []string) []string {
	var files []string

	fi, err := os.Stat(p)
	if err != nil {
		return files
	}
	if !fi.IsDir() {
		if !contains(extensions, filepath.Ext(p)) {
			return files
		}
		files = append(files, p)
		return files
	}
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

func isDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return fileInfo.IsDir(), err
}
