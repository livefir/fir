package fir

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cespare/xxhash/v2"
	"github.com/icholy/replace"
	"github.com/sourcegraph/conc/pool"
	"github.com/teris-io/shortid"
	"golang.org/x/net/html"
)

type eventTemplate map[string]struct{}
type eventTemplates map[string]eventTemplate

func layoutEmptyContentSet(opt routeOpt, content, layoutContentName string) (*template.Template, eventTemplates, error) {
	// is content html content or a file/directory
	pageContentPath := filepath.Join(opt.publicDir, content)
	if !opt.existFile(pageContentPath) {
		return parseString(
			template.New(
				layoutContentName).
				Funcs(opt.getFuncMap()),
			content)
	}
	// content must be  a file or directory
	pageFiles := getPartials(opt, find(pageContentPath, opt.extensions, opt.embedfs))
	contentTemplate := template.New(filepath.Base(pageContentPath)).Funcs(opt.getFuncMap())

	return parseFiles(contentTemplate, opt.getFuncMap(), opt.readFile, pageFiles...)
}

func layoutSetContentEmpty(opt routeOpt, layout string) (*template.Template, eventTemplates, error) {
	pageLayoutPath := filepath.Join(opt.publicDir, layout)
	evt := make(eventTemplates)
	// is layout html content or a file/directory
	if !opt.existFile(pageLayoutPath) {
		return parseString(template.New("").Funcs(opt.getFuncMap()), layout)
	}

	// layout must be  a file
	if isDir(pageLayoutPath, opt.embedfs) {
		return nil, evt, fmt.Errorf("layout %s is a directory but must be a file", pageLayoutPath)
	}

	// compile layout
	commonFiles := getPartials(opt, []string{pageLayoutPath})
	layoutTemplate := template.New(filepath.Base(pageLayoutPath)).Funcs(opt.getFuncMap())

	return parseFiles(template.Must(layoutTemplate.Clone()), opt.getFuncMap(), opt.readFile, commonFiles...)
}

func layoutSetContentSet(opt routeOpt, content, layout, layoutContentName string) (*template.Template, eventTemplates, error) {
	layoutTemplate, evt, err := layoutSetContentEmpty(opt, layout)
	if err != nil {
		return nil, evt, err
	}

	//logger.Infof("compiled layoutTemplate...")
	//for _, v := range layoutTemplate.Templates() {
	//	fmt.Println("template => ", v.Name())
	//}

	// 2. add content to layout
	// check if content is a not a file or directory

	pageContentPath := filepath.Join(opt.publicDir, content)
	if !opt.existFile(pageContentPath) {
		pageTemplate, currEvt, err := parseString(layoutTemplate, content)
		if err != nil {
			panic(err)
		}
		evt = deepMergeEventTemplates(evt, currEvt)
		if err := checkPageContent(pageTemplate, layoutContentName); err != nil {
			return nil, nil, err
		}
		return pageTemplate, evt, nil
	} else {
		pageFiles := getPartials(opt, find(pageContentPath, opt.extensions, opt.embedfs))
		pageTemplate, currEvt, err := parseFiles(layoutTemplate.Funcs(opt.getFuncMap()), opt.getFuncMap(), opt.readFile, pageFiles...)
		if err != nil {
			panic(err)
		}
		evt = deepMergeEventTemplates(evt, currEvt)
		if err := checkPageContent(pageTemplate, layoutContentName); err != nil {
			return nil, nil, err
		}
		return pageTemplate, evt, nil
	}

}

func getPartials(opt routeOpt, files []string) []string {
	for _, partial := range opt.partials {
		files = append(files, find(filepath.Join(opt.publicDir, partial), opt.extensions, opt.embedfs)...)
	}
	return files
}

func checkPageContent(tmpl *template.Template, layoutContentName string) error {
	if ct := tmpl.Lookup(layoutContentName); ct == nil {
		return fmt.Errorf("err looking up layoutContent: the layout %s expects a template named %s",
			tmpl.Name(), layoutContentName)
	}
	return nil
}

// creates a html/template for the route
func parseTemplate(opt routeOpt) (*template.Template, eventTemplates, error) {
	opt.addFunc("fir", newFirFuncMap(RouteContext{}, nil)["fir"])

	// if both layout and content is empty show a default page.
	if opt.layout == "" && opt.content == "" {
		return template.Must(template.New("").
			Parse(`<div style="text-align:center"> This is a default page. </div>`)), nil, nil
	}

	// if layout is set and content is empty
	if opt.layout != "" && opt.content == "" {
		return layoutSetContentEmpty(opt, opt.layout)
	}

	// if layout is empty and content is set
	if opt.layout == "" && opt.content != "" {
		return layoutEmptyContentSet(opt, opt.content, opt.layoutContentName)
	}

	// both layout and content are set
	return layoutSetContentSet(opt, opt.content, opt.layout, opt.layoutContentName)
}

// creates a html/template for the route errors
func parseErrorTemplate(opt routeOpt) (*template.Template, eventTemplates, error) {
	opt.addFunc("fir", newFirFuncMap(RouteContext{}, nil)["fir"])
	if opt.errorLayout == "" {
		opt.errorLayout = opt.layout
		opt.errorLayoutContentName = opt.layoutContentName
	}
	// if both layout and content is empty show a default page.
	if opt.errorLayout == "" && opt.errorContent == "" {
		return template.Must(template.New("").
			Parse(`<div style="text-align:center"> This is a default page. </div>`)), nil, nil
	}

	// if layout is set and content is empty
	if opt.errorLayout != "" && opt.errorContent == "" {
		return layoutSetContentEmpty(opt, opt.errorLayout)
	}

	// if layout is empty and content is set
	if opt.errorLayout == "" && opt.errorContent != "" {
		return layoutEmptyContentSet(opt, opt.errorContent, opt.errorLayoutContentName)
	}

	// both layout and content are set
	return layoutSetContentSet(opt, opt.errorContent, opt.errorLayout, opt.errorLayoutContentName)
}

type fileInfo struct {
	name           string
	content        []byte
	eventTemplates eventTemplates
	blocks         map[string]string
	err            error
}

var templateNameRegex = regexp.MustCompile(`^[ A-Za-z0-9\-:_]*$`)

func parseString(t *template.Template, content string) (*template.Template, eventTemplates, error) {
	fi := readAttributes(fileInfo{content: []byte(content)})
	t, err := t.Parse(string(fi.content))
	return t, fi.eventTemplates, err
}

func parseFiles(t *template.Template, funcs template.FuncMap, readFile func(string) (string, []byte, error), filenames ...string) (*template.Template, eventTemplates, error) {

	if len(filenames) == 0 {
		// Not really a problem, but be consistent.
		return t, nil, nil
	}
	resultPool := pool.NewWithResults[fileInfo]()
	for _, filename := range filenames {
		filename := filename
		resultPool.Go(func() fileInfo {
			name, b, err1 := readFile(filename)
			if err1 != nil {
				return fileInfo{name: name, err: err1}
			}
			b, blocks, err2 := extractTemplates(b)
			if err2 != nil {
				return fileInfo{name: name, err: err2}
			}

			return readAttributes(fileInfo{name: name, content: b, blocks: blocks})
		})
	}

	evt := make(eventTemplates)

	fileInfos := resultPool.Wait()
	for _, fi := range fileInfos {
		evt = deepMergeEventTemplates(evt, fi.eventTemplates)
		if fi.err != nil {
			return t, evt, fi.err
		}

		s := string(fi.content)
		// First template becomes return value if not already defined,
		// and we use that one for subsequent New calls to associate
		// all the templates together. Also, if this file has the same name
		// as t, this file becomes the contents of t, so
		//  t, err := New(name).Funcs(xxx).ParseFiles(name)
		// works. Otherwise we create a new template associated with t.
		var tmpl *template.Template
		if t == nil {
			t = template.New(fi.name)
		}
		if fi.name == t.Name() {
			tmpl = t
		} else {
			tmpl = t.New(fi.name)
		}

		_, err := tmpl.Parse(s)
		if err != nil {
			fmt.Println(s)
			return t, evt, fmt.Errorf("parsing %s: %v", fi.name, err)
		}
		for name, block := range fi.blocks {
			bt, err := template.New(name).Funcs(funcs).Parse(block)
			if err != nil {
				return t, evt, fmt.Errorf("file: %v, parsing  block template  %s: %v", fi.name, name, err)
			}
			tmpl, err = tmpl.AddParseTree(bt.Name(), bt.Tree)
			if err != nil {
				return t, evt, fmt.Errorf("file: %v, adding block template %s: %v", fi.name, name, err)
			}
		}
	}

	return t, evt, nil
}

func deepMergeEventTemplates(evt1, evt2 eventTemplates) eventTemplates {
	merged := make(eventTemplates)
	for eventID, templatesMap := range evt1 {
		merged[eventID] = templatesMap
	}
	for eventID, templatesMap := range evt2 {
		templatesMap1, ok := merged[eventID]
		if !ok {
			merged[eventID] = templatesMap
			continue
		}
		for templateName := range templatesMap {
			templatesMap1[templateName] = struct{}{}
		}
	}
	return merged

}

// extractTemplates extracts innerHTML content from fir event namespace string and updates the namespace string with a template
func extractTemplates(content []byte) ([]byte, map[string]string, error) {
	blocks := make(map[string]string)
	if len(content) == 0 {
		return content, blocks, nil
	}

	reader := replace.Chain(bytes.NewReader(content),
		// increment all numbers
		replace.RegexpStringFunc(regexp.MustCompile(`:error=`), func(match string) string {
			return fmt.Sprintf(":error::%s=", strings.ToLower(shortid.MustGenerate()))
		}),
		replace.RegexpStringFunc(regexp.MustCompile(`:error]=`), func(match string) string {
			return fmt.Sprintf(":error]::%s=", strings.ToLower(shortid.MustGenerate()))
		}),
		replace.RegexpStringFunc(regexp.MustCompile(`:ok=`), func(match string) string {
			return fmt.Sprintf(":ok::%s=", strings.ToLower(shortid.MustGenerate()))
		}),
		replace.RegexpStringFunc(regexp.MustCompile(`:ok]=`), func(match string) string {
			return fmt.Sprintf(":ok]::%s=", strings.ToLower(shortid.MustGenerate()))
		}),
	)

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(reader)
	if err != nil {
		return content, blocks, err
	}

	content = buf.Bytes()

	doc, err := html.Parse(bytes.NewReader(content))
	if err != nil {
		return content, blocks, err
	}

	var findFirNode func(*html.Node)
	var getHtml func(*html.Node) string
	findFirNode = func(node *html.Node) {
		if node.Type == html.ElementNode {
			for _, attr := range node.Attr {
				if strings.HasPrefix(attr.Key, "@fir") || strings.HasPrefix(attr.Key, "x-on:fir") {
					// check if fir event namespace string already contains a template
					if !strings.Contains(attr.Key, "::") {
						continue
					}

					block := getHtml(node)
					// check if innerHTML content is actually a html template
					if strings.Contains(block, "{{") && strings.Contains(block, "}}") {
						tempTemplateName := strings.Split(attr.Key, "::")[1]
						templateName := fmt.Sprintf("fir-%s", hashID(block))
						content = bytes.Replace(content, []byte(tempTemplateName), []byte(templateName), -1)
						blocks[templateName] = block
					}

					// update fir event namespace string with a template

					break
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			findFirNode(child)
		}
	}

	getHtml = func(node *html.Node) string {
		block := ""
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.TextNode {
				block += c.Data
			} else {
				//  append the attributes
				var attributes string
				for _, attr := range c.Attr {
					attrstr := fmt.Sprintf(` %s="%s"`, attr.Key, attr.Val)
					//fmt.Printf("attr key: %v, value: %v\n", attr.Key, attr.Val)
					attributes += attrstr
				}

				block += "<" + c.Data + attributes + ">"
				block += getHtml(c)
				block += "</" + c.Data + ">"
			}
		}
		return block
	}

	findFirNode(doc)

	return content, blocks, nil
}

func hashID(content string) string {
	xxhash := xxhash.New()
	_, err := xxhash.WriteString(content)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", xxhash.Sum(nil))
}
