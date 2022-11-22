package fir

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
)

// DefaultViewExtensions are the default view extensions configured in the DefaultView, DefaultErrorView
var DefaultViewExtensions = []string{".gohtml", ".gotmpl", ".html", ".tmpl"}

// View is an encapsulation of the template and the data to be rendered. It is used to render the page.
// The template is rendered using the data provided by the Page. The Page is populated by the OnGet and OnPost methods.
// The Page can be patched by the Patch operations sent to the browser in response to the events sent from the browser.
type View interface {
	/** Settings **/
	// ID returns the view id. Defaults to "". Used to construct the channel name to which the view is subscribed.
	ID() string
	// Content returns either path to the content or a html string content
	Content() string
	// Layout returns either path to the layout or a html string layout
	Layout() string
	// LayoutContentName is the defined template of the "content" to be replaced. Defaults to "content"
	LayoutContentName() string
	// Partials returns path to any partials used in the view. Defaults to "./templates/partials"
	Partials() []string
	// Extensions returns the view extensions. Defaults to: html, tmpl, gohtml, gotmpl
	Extensions() []string
	// FuncMap configures the html/template.FuncMap
	FuncMap() template.FuncMap
	/** Lifecyle **/
	// OnGet is called when the page is first loaded for the http route.
	OnGet(http.ResponseWriter, *http.Request) Page
	// OnPost is called when a form is submitted for the http route.
	OnPost(http.ResponseWriter, *http.Request) Page
	// OnEvent handles the events sent from the browser
	OnEvent(event Event) Patchset
	// Publisher returns a channel to which the view can publish patchsets
	Publisher() <-chan Patchset
}

type viewHandler struct {
	view              View
	errorView         View
	viewTemplate      *template.Template
	errorViewTemplate *template.Template
	mountData         map[string]any
	cntrl             *controller
	streamCh          chan Patchset
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

	layoutTemplate := template.New(filepath.Base(viewLayoutPath)).Funcs(view.FuncMap())
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

	contentTemplate := template.New(filepath.Base(viewContentPath)).Funcs(view.FuncMap())
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
