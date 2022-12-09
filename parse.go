package fir

import (
	"fmt"
	"html/template"
	"path/filepath"
)

func layoutEmptyContentSet(opt routeOpt) (*template.Template, error) {
	// is content html content or a file/directory
	pageContentPath := filepath.Join(opt.publicDir, opt.content)
	if isFileOrString(pageContentPath, opt) {
		return template.Must(
			template.New(
				opt.layoutContentName).
				Funcs(opt.funcMap).
				Parse(opt.content)), nil
	}
	// content must be  a file or directory

	var pageFiles []string
	// page and its partials
	pageFiles = append(pageFiles, find(opt, pageContentPath, opt.extensions)...)
	for _, partial := range opt.partials {
		pageFiles = append(pageFiles, find(opt, filepath.Join(opt.publicDir, partial), opt.extensions)...)
	}

	contentTemplate := template.New(filepath.Base(pageContentPath)).Funcs(opt.funcMap)
	if opt.hasEmbedFS {
		contentTemplate = template.Must(contentTemplate.Funcs(opt.funcMap).ParseFS(opt.embedFS, pageFiles...))
	} else {
		contentTemplate = template.Must(contentTemplate.Funcs(opt.funcMap).ParseFiles(pageFiles...))
	}

	return contentTemplate, nil
}

func layoutSetContentEmpty(opt routeOpt) (*template.Template, error) {
	pageLayoutPath := filepath.Join(opt.publicDir, opt.layout)
	// is layout html content or a file/directory
	if isFileOrString(pageLayoutPath, opt) {
		return template.Must(template.New("").Funcs(opt.funcMap).Parse(opt.layout)), nil
	}

	// layout must be  a file
	if isDir(pageLayoutPath, opt) {
		return nil, fmt.Errorf("layout %s is a directory but must be a file", pageLayoutPath)
	}
	// compile layout
	commonFiles := []string{pageLayoutPath}
	// global partials
	for _, partial := range opt.partials {
		commonFiles = append(commonFiles, find(opt, filepath.Join(opt.publicDir, partial), opt.extensions)...)
	}

	layoutTemplate := template.New(filepath.Base(pageLayoutPath)).Funcs(opt.funcMap)
	if opt.hasEmbedFS {
		layoutTemplate = template.Must(layoutTemplate.Funcs(opt.funcMap).ParseFS(opt.embedFS, commonFiles...))
	} else {
		layoutTemplate = template.Must(layoutTemplate.Funcs(opt.funcMap).ParseFiles(commonFiles...))
	}

	return template.Must(layoutTemplate.Clone()), nil
}

func layoutSetContentSet(opt routeOpt) (*template.Template, error) {
	layoutTemplate, err := layoutSetContentEmpty(opt)
	if err != nil {
		return nil, err
	}

	//log.Println("compiled layoutTemplate...")
	//for _, v := range layoutTemplate.Templates() {
	//	fmt.Println("template => ", v.Name())
	//}

	// 2. add content to layout
	// check if content is a not a file or directory
	var pageTemplate *template.Template
	pageContentPath := filepath.Join(opt.publicDir, opt.content)
	if isFileOrString(pageContentPath, opt) {
		pageTemplate = template.Must(layoutTemplate.Parse((opt.content)))
	} else {
		var pageFiles []string
		// page and its partials
		pageFiles = append(pageFiles, find(opt, filepath.Join(opt.publicDir, opt.content), opt.extensions)...)
		if opt.hasEmbedFS {
			pageTemplate = template.Must(layoutTemplate.Funcs(opt.funcMap).ParseFS(opt.embedFS, pageFiles...))
		} else {
			pageTemplate = template.Must(layoutTemplate.Funcs(opt.funcMap).ParseFiles(pageFiles...))
		}
	}

	// check if the final pageTemplate contains a content child template which is `content` by default.
	if ct := pageTemplate.Lookup(opt.layoutContentName); ct == nil {
		return nil,
			fmt.Errorf("err looking up layoutContent: the layout %s expects a template named %s",
				opt.layout, opt.layoutContentName)
	}

	return pageTemplate, nil
}

// creates a html/template from the Page type.
func parseTemplate(opt routeOpt) (*template.Template, error) {
	// if both layout and content is empty show a default page.
	if opt.layout == "" && opt.content == "" {
		return template.Must(template.New("").
			Parse(`<div style="text-align:center"> This is a default page. </div>`)), nil
	}

	// if layout is set and content is empty
	if opt.layout != "" && opt.content == "" {
		return layoutSetContentEmpty(opt)
	}

	// if layout is empty and content is set
	if opt.layout == "" && opt.content != "" {
		return layoutEmptyContentSet(opt)
	}

	// both layout and content are set
	return layoutSetContentSet(opt)
}
