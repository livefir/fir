package fir

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/sourcegraph/conc/pool"
	"golang.org/x/exp/slices"
	"k8s.io/klog/v2"
)

type eventTemplate map[string]struct{}
type eventTemplates map[string]eventTemplate
type readFileFunc func(string) (string, []byte, error)

func layoutEmptyContentSet(opt routeOpt, content, layoutContentName string) (*template.Template, eventTemplates, error) {
	// is content html content or a file/directory
	pageContentPath := filepath.Join(opt.publicDir, content)
	if isFileOrString(pageContentPath, opt) {
		return parseString(
			template.New(
				layoutContentName).
				Funcs(opt.funcMap),
			content)
	}
	// content must be  a file or directory
	pageFiles := getPartials(opt, find(opt, pageContentPath, opt.extensions))
	contentTemplate := template.New(filepath.Base(pageContentPath)).Funcs(opt.funcMap)

	return parseFiles(contentTemplate, opt.readFile, pageFiles...)
}

func layoutSetContentEmpty(opt routeOpt, layout string) (*template.Template, eventTemplates, error) {
	pageLayoutPath := filepath.Join(opt.publicDir, layout)
	evt := make(eventTemplates)
	// is layout html content or a file/directory
	if isFileOrString(pageLayoutPath, opt) {
		return parseString(template.New("").Funcs(opt.funcMap), layout)
	}

	// layout must be  a file
	if isDir(pageLayoutPath, opt) {
		return nil, evt, fmt.Errorf("layout %s is a directory but must be a file", pageLayoutPath)
	}

	// compile layout
	commonFiles := getPartials(opt, []string{pageLayoutPath})
	layoutTemplate := template.New(filepath.Base(pageLayoutPath)).Funcs(opt.funcMap)

	return parseFiles(template.Must(layoutTemplate.Clone()), opt.readFile, commonFiles...)
}

func layoutSetContentSet(opt routeOpt, content, layout, layoutContentName string) (*template.Template, eventTemplates, error) {
	layoutTemplate, evt, err := layoutSetContentEmpty(opt, layout)
	if err != nil {
		return nil, evt, err
	}

	//log.Println("compiled layoutTemplate...")
	//for _, v := range layoutTemplate.Templates() {
	//	fmt.Println("template => ", v.Name())
	//}

	// 2. add content to layout
	// check if content is a not a file or directory

	pageContentPath := filepath.Join(opt.publicDir, content)
	if isFileOrString(pageContentPath, opt) {
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
		pageFiles := getPartials(opt, []string{pageContentPath})
		pageTemplate, currEvt, err := parseFiles(layoutTemplate.Funcs(opt.funcMap), opt.readFile, pageFiles...)
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
		files = append(files, find(opt, filepath.Join(opt.publicDir, partial), opt.extensions)...)
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
	err            error
}

var templateNameRegex = regexp.MustCompile(`^[ A-Za-z0-9\-:]*$`)

func parseString(t *template.Template, content string) (*template.Template, eventTemplates, error) {
	fi := query(fileInfo{content: []byte(content)})
	t, err := t.Parse(string(fi.content))
	return t, fi.eventTemplates, err
}

func parseFiles(t *template.Template, readFile func(string) (string, []byte, error), filenames ...string) (*template.Template, eventTemplates, error) {

	if len(filenames) == 0 {
		// Not really a problem, but be consistent.
		return t, nil, nil
	}
	resultPool := pool.NewWithResults[fileInfo]()
	for _, filename := range filenames {
		filename := filename
		resultPool.Go(func() fileInfo {
			name, b, err := readFile(filename)
			return query(fileInfo{name: name, content: b, err: err})
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
			return t, evt, err
		}
	}

	return t, evt, nil
}

func eventFormatError(eventns string) string {
	return fmt.Sprintf(`
	error: invalid event namespace: %s. must be of either of the two formats =>
	1. @fir:<event>:<ok|error>::<block-name|optional>
	2. @fir:<event>:<pending|done>`, eventns)
}

func transform(content []byte) []byte {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(content))
	if err != nil {
		panic(err)
	}

	doc.Find("*").Each(func(_ int, node *goquery.Selection) {
		for _, attr := range node.Get(0).Attr {
			if !strings.HasPrefix(attr.Key, "@fir:") && !strings.HasPrefix(attr.Key, "x-on:fir:") {
				continue
			}

			eventns := strings.TrimPrefix(attr.Key, "@fir:")
			eventns = strings.TrimPrefix(eventns, "x-on:fir:")
			// eventns might have modifiers like .prevent, .stop, .self, .once, .window, .document etc. remove them
			eventnsParts := strings.SplitN(eventns, ".", -1)
			var modifiers string
			if len(eventnsParts) > 0 {
				eventns = eventnsParts[0]
			}
			if len(eventnsParts) > 1 {
				modifiers = strings.Join(eventnsParts[1:], ".")
			}

			// eventns might have a filter:[e1:ok,e2:ok] containing multiple event:state separated by comma
			eventnsList, filterExists := getEventNsList(eventns)
			// if filter exists remove the current attribute from the node
			if filterExists {
				node.RemoveAttr(attr.Key)
			}

			for _, eventns := range eventnsList {
				eventns = strings.TrimSpace(eventns)
				// set @fir|x-on:fir:eventns attribute to the node
				eventnsWithModifiers := fmt.Sprintf("%s.%s", eventns, modifiers)
				if len(modifiers) == 0 {
					eventnsWithModifiers = eventns
				}
				_, atFirOk := node.Attr(fmt.Sprintf("@fir:%s", eventnsWithModifiers))
				_, xOnFirOk := node.Attr(fmt.Sprintf("x-on:fir:%s", eventnsWithModifiers))
				// if the node already has @fir:x attribute, then skip
				if !atFirOk && !xOnFirOk {
					node.SetAttr(fmt.Sprintf("@fir:%s", eventnsWithModifiers), attr.Val)
				}

				// fir-myevent-ok--myblock
				key, _ := node.Attr("key")
				classname := fmt.Sprintf("fir-%s", getClassName(eventns, &key))
				if !node.HasClass(classname) {
					node.AddClass(classname)
				}

			}

		}
	})
	html, err := doc.Html()
	if err != nil {
		panic(err)
	}
	return []byte(html)
}

func getClassName(eventns string, key *string) string {
	cls := strings.ReplaceAll(eventns, ":", "-")
	if key != nil && *key != "" {
		cls = cls + "--" + strings.ReplaceAll(*key, " ", "-")
	}
	return cls
}

func query(fi fileInfo) fileInfo {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(fi.content))
	if err != nil {
		panic(err)
	}
	evt := make(eventTemplates)
	doc.Find("*").Each(func(_ int, node *goquery.Selection) {
		for _, attr := range node.Get(0).Attr {
			if !strings.HasPrefix(attr.Key, "@fir:") && !strings.HasPrefix(attr.Key, "x-on:fir:") {
				continue
			}

			eventns := strings.TrimPrefix(attr.Key, "@fir:")
			eventns = strings.TrimPrefix(eventns, "x-on:fir:")
			// eventns might have modifiers like .prevent, .stop, .self, .once, .window, .document etc. remove them
			eventnsParts := strings.SplitN(eventns, ".", -1)

			if len(eventnsParts) > 0 {
				eventns = eventnsParts[0]
			}

			// eventns might have a filter:[e1:ok,e2:ok] containing multiple event:state separated by comma
			eventnsList, _ := getEventNsList(eventns)

			for _, eventns := range eventnsList {
				eventns = strings.TrimSpace(eventns)
				// set @fir|x-on:fir:eventns attribute to the node

				// myevent:ok::myblock
				eventnsParts = strings.SplitN(eventns, "::", -1)
				if len(eventnsParts) == 0 {
					continue
				}

				// [myevent:ok, myblock]
				if len(eventnsParts) > 2 {
					klog.Errorf(eventFormatError(eventns))
					continue
				}

				// myevent:ok
				eventID := eventnsParts[0]
				// [myevent, ok]
				eventIDParts := strings.SplitN(eventID, ":", -1)
				if len(eventIDParts) != 2 {
					klog.Errorf(eventFormatError(eventns))
					continue
				}
				// event name can only be followed by ok, error, pending, done
				if !slices.Contains([]string{"ok", "error", "pending", "done"}, eventIDParts[1]) {
					klog.Errorf(eventFormatError(eventns))
					continue
				}
				// assert myevent:ok::myblock or myevent:error::myblock
				if len(eventnsParts) == 2 && !slices.Contains([]string{"ok", "error"}, eventIDParts[1]) {
					klog.Errorf(eventFormatError(eventns))
					continue

				}
				// template name is declared for event state i.e. myevent:ok::myblock
				templateName := "-"
				if len(eventnsParts) == 2 {
					templateName = eventnsParts[1]
				}

				templates, ok := evt[eventID]
				if !ok {
					templates = make(eventTemplate)
				}

				if !templateNameRegex.MatchString(templateName) {
					klog.Errorf("error: invalid template name in event binding: only hyphen(-) and colon(:) are allowed: %v\n", templateName)
					continue
				}

				templates[templateName] = struct{}{}
				// fmt.Printf("eventID: %s, templateName: %s\n", eventID, templateName)

				evt[eventID] = templates

			}

		}
	})

	return fileInfo{
		name:           fi.name,
		content:        fi.content,
		err:            fi.err,
		eventTemplates: evt,
	}
}

// checks if the event string is of the format [event1:ok,event2:ok]:tmpl1 and returns the unbundled list of event strings
// event1:ok:tmpl1,event2:ok:tmpl1. if not, returns original event string
func getEventNsList(input string) ([]string, bool) {
	ef := getEventFilter(input)
	if ef == nil {
		return []string{input}, false
	}
	if len(ef.Values) == 0 {
		return []string{input}, false
	}
	var eventnsList []string
	for _, v := range ef.Values {
		eventnsList = append(eventnsList, ef.BeforeBracket+v+ef.AfterBracket)
	}
	return eventnsList, true
}

type eventFilter struct {
	BeforeBracket string
	Values        []string
	AfterBracket  string
}

func getEventFilter(input string) *eventFilter {
	// Extract the part of the string before the open square bracket
	beforeRe := regexp.MustCompile(`^(.*?)\[`)
	beforeMatch := beforeRe.FindStringSubmatch(input)

	if len(beforeMatch) < 2 {
		return nil
	}
	beforeBracket := beforeMatch[1]

	// Extract the part of the string after the closed square bracket
	afterRe := regexp.MustCompile(`\](.*)$`)
	afterMatch := afterRe.FindStringSubmatch(input)

	if len(afterMatch) < 2 {
		return nil
	}
	afterBracket := afterMatch[1]

	// Extract the contents of a closed square bracket
	re := regexp.MustCompile(`\[(.*?)\]`)
	matches := re.FindStringSubmatch(input)

	if len(matches) < 2 {
		return nil
	}

	// Remove whitespace and split the contents by comma
	contents := strings.ReplaceAll(matches[1], " ", "")
	values := strings.Split(contents, ",")

	// Validate and format each value
	validValues := make([]string, 0)
	for _, value := range values {
		if !isValidValue(value) {
			return nil
		}
		validValues = append(validValues, formatValue(value))
	}

	extractedValues := &eventFilter{
		BeforeBracket: beforeBracket,
		Values:        validValues,
		AfterBracket:  afterBracket,
	}

	return extractedValues
}

func isValidValue(value string) bool {
	re := regexp.MustCompile(`^[^-\s]+:(ok|pending|error|done)$`)
	return re.MatchString(value)
}

func formatValue(value string) string {
	parts := strings.Split(value, ":")
	return fmt.Sprintf("%s:%s", parts[0], parts[1])
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

func readFileOS(file string) (name string, b []byte, err error) {
	name = filepath.Base(file)
	b, err = os.ReadFile(file)
	return
}

func readFileFS(fsys fs.FS) func(string) (string, []byte, error) {
	return func(file string) (name string, b []byte, err error) {
		name = path.Base(file)
		b, err = fs.ReadFile(fsys, file)
		return
	}
}
