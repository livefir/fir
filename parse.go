package fir

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/cespare/xxhash/v2"
	"github.com/icholy/replace"
	"github.com/livefir/fir/internal/logger"
	"github.com/sourcegraph/conc/pool"
	"github.com/teris-io/shortid"
	"golang.org/x/net/html"
)

type FirEventState string
type FirEventModifier string

const (
	FirAtPrefix  = "@fir"
	FirXonPrefix = "x-on:fir"

	StateOK        FirEventState    = "ok"
	StateError     FirEventState    = "error"
	StatePending   FirEventState    = "pending"
	StateDone      FirEventState    = "done"
	ModifierNoHTML FirEventModifier = ".nohtml"
)

// IsValid checks if the FirEventState is valid.
func (s FirEventState) IsValid() bool {
	switch s {
	case StateOK, StateError, StatePending, StateDone:
		return true
	default:
		return false
	}
}

// IsValid checks if the FirEventModifier is valid.
func (m FirEventModifier) IsValid() bool {
	switch m {
	case ModifierNoHTML:
		return true
	default:
		return false
	}
}

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
			opt.getFuncMap(),
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
		return parseString(template.New("").Funcs(opt.getFuncMap()), opt.getFuncMap(), layout)
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
		pageTemplate, currEvt, err := parseString(layoutTemplate, opt.getFuncMap(), content)
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

var templateNameRegex = regexp.MustCompile(`^[ A-Za-z0-9\-:_.]*$`)

func parseString(t *template.Template, funcs template.FuncMap, content string) (*template.Template, eventTemplates, error) {
	b, blocks, err := extractTemplates([]byte(content))
	if err != nil {
		return nil, nil, err
	}

	fi := readAttributes(fileInfo{content: b, blocks: blocks})
	t, err = t.Parse(string(fi.content))
	if err != nil {
		return t, fi.eventTemplates, fmt.Errorf("parsing %s: %v", fi.name, err)
	}
	for name, block := range fi.blocks {
		bt, err := template.New(name).Funcs(funcs).Parse(block)
		if err != nil {
			logger.Warnf("file: %v, error parsing auto extracted template  %s: %v", fi.name, name, err)
			bt = template.Must(template.New(name).Funcs(funcs).Parse("<!-- error parsing auto extracted template -->"))
		}
		t, err = t.AddParseTree(bt.Name(), bt.Tree)
		if err != nil {
			return t, fi.eventTemplates, fmt.Errorf("file: %v, error adding block template %s: %v", fi.name, name, err)
		}
	}
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
			return t, evt, fmt.Errorf("parsing %s: %v", fi.name, err)
		}
		for name, block := range fi.blocks {
			bt, err := template.New(name).Funcs(funcs).Parse(block)
			if err != nil {
				logger.Warnf("file: %v, error parsing auto extracted template  %s: %v", fi.name, name, err)
				bt = template.Must(template.New(name).Funcs(funcs).Parse("<!-- error parsing auto extracted template -->"))
			}
			tmpl, err = tmpl.AddParseTree(bt.Name(), bt.Tree)
			if err != nil {
				return t, evt, fmt.Errorf("file: %v, error adding block template %s: %v", fi.name, name, err)
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

	// Replace placeholders in the content
	content, err := replacePlaceholders(content)
	if err != nil {
		return content, blocks, err
	}

	// Parse the content into an HTML document
	doc, err := html.Parse(bytes.NewReader(content))
	if err != nil {
		return content, blocks, err
	}

	// Replace and extract templates
	content, err = processTemplates(doc, content, blocks)
	if err != nil {
		return content, blocks, err
	}

	return content, blocks, nil
}

var (
	stateErrorRegex1 = regexp.MustCompile(fmt.Sprintf(`:%s=`, StateError))
	stateErrorRegex2 = regexp.MustCompile(fmt.Sprintf(`:%s]=`, StateError))
	stateOKRegex1    = regexp.MustCompile(fmt.Sprintf(`:%s=`, StateOK))
	stateOKRegex2    = regexp.MustCompile(fmt.Sprintf(`:%s]=`, StateOK))
)

// replacePlaceholders replaces specific placeholders in the given content with generated template names.
// The placeholders are identified using regular expressions, and their replacements are generated
// dynamically based on the placeholder type.
//
// Parameters:
//   - content: A byte slice containing the input content where placeholders need to be replaced.
//
// Returns:
//   - A byte slice with the placeholders replaced by their corresponding template names.
//   - An error if there is an issue during the replacement process.
//
// If the input content is empty, the function returns the content as is without any modifications.
func replacePlaceholders(content []byte) ([]byte, error) {
	if len(content) == 0 {
		return content, nil
	}

	reader := replace.Chain(bytes.NewReader(content),
		replace.RegexpStringFunc(stateErrorRegex1, generateTemplateName(string(StateError))),
		replace.RegexpStringFunc(stateErrorRegex2, generateTemplateName(string(StateError)+"]")),
		replace.RegexpStringFunc(stateOKRegex1, generateTemplateName(string(StateOK))),
		replace.RegexpStringFunc(stateOKRegex2, generateTemplateName(string(StateOK)+"]")),
	)

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(reader)
	if err != nil {
		return content, err
	}

	return buf.Bytes(), nil
}

// generateTemplateName returns a function that generates a unique template name
// based on the provided prefix. The returned function takes a string argument
// (match) and produces a formatted string containing the prefix and a randomly
// generated short ID in lowercase.
//
// Parameters:
//   - prefix: A string that serves as the prefix for the generated template name.
//
// Returns:
//   - A function that takes a string (match) and returns a formatted template name.
func generateTemplateName(prefix string) func(string) string {
	return func(match string) string {
		return fmt.Sprintf(":%s::fir-gen-templ-%s=", prefix, strings.ToLower(shortid.MustGenerate()))
	}
}

// processTemplates processes an HTML document to extract and replace template blocks
// based on specific attributes and conditions.
//
// Parameters:
//   - doc: The root HTML node of the document to process.
//   - content: The byte slice representing the content of the document.
//   - blocks: A map to store extracted template blocks, where the key is the template name
//     and the value is the corresponding HTML block.
//
// Returns:
// - A modified byte slice of the content with replaced templates.
// - An error if parsing the HTML content fails.
//
// The function performs the following operations:
//  1. Replaces template names in the content with their corresponding HTML blocks
//     if the block is identified as an HTML template.
//  2. Removes template namespaces from the content if the block is not an HTML template.
//  3. Extracts template blocks from the document and stores them in the provided `blocks` map.
//
// The function uses three nested helper functions:
// - replacer: Recursively traverses the HTML nodes to replace template names or remove namespaces.
// - extractor: Recursively traverses the HTML nodes to extract template blocks and store them in the map.
// - getHtml: Recursively generates the HTML string representation of a node and its children.
//
// Note: The function assumes the presence of utility functions such as `isFirEvent`,
// `extractTemplateName`, `isHtmlTemplate`, `replaceTemplateName`, `removeTemplateNamespace`,
// and `formatHtmlNode` to perform specific operations.
func processTemplates(doc *html.Node, content []byte, blocks map[string]string) ([]byte, error) {
	var replacer func(*html.Node)
	var extractor func(*html.Node)
	var getHtml func(*html.Node) string

	replacer = func(node *html.Node) {
		if node.Type == html.ElementNode {
			for _, attr := range node.Attr {
				if isFirEvent(attr.Key) && strings.Contains(attr.Key, "::fir-gen-templ-") {
					block := getHtml(node)
					tempTemplateName := extractTemplateName(attr.Key)
					hasNoHtmlModifier := strings.Contains(tempTemplateName, ".nohtml")

					if isHtmlTemplate(block) {
						content = replaceTemplateName(content, tempTemplateName, block, hasNoHtmlModifier)
					} else {
						content = removeTemplateNamespace(content, tempTemplateName)
					}
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			replacer(child)
		}
	}

	extractor = func(node *html.Node) {
		if node.Type == html.ElementNode {
			for _, attr := range node.Attr {
				if isFirEvent(attr.Key) && strings.Contains(attr.Key, "::fir-") {
					block := getHtml(node)
					templateName := extractTemplateName(attr.Key)
					if isHtmlTemplate(block) {
						blocks[templateName] = block
					}
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			extractor(child)
		}
	}

	getHtml = func(node *html.Node) string {
		var block strings.Builder
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.TextNode {
				block.WriteString(c.Data)
			} else {
				block.WriteString(formatHtmlNode(c, getHtml))
			}
		}
		return block.String()
	}

	replacer(doc)

	doc, err := html.Parse(bytes.NewReader(content))
	if err != nil {
		return content, err
	}

	extractor(doc)

	return content, nil
}

func isFirEvent(key string) bool {
	return strings.HasPrefix(key, FirAtPrefix) || strings.HasPrefix(key, FirXonPrefix)
}

func extractTemplateName(key string) string {
	return strings.Split(key, "::")[1]
}

func isHtmlTemplate(block string) bool {
	return strings.Contains(block, "{{") && strings.Contains(block, "}}")
}

func replaceTemplateName(content []byte, tempTemplateName, block string, hasNoHtmlModifier bool) []byte {
	if !bytes.Contains(content, []byte(tempTemplateName)) {
		return content
	}

	templateName := fmt.Sprintf("fir-%s", hashID(block))
	if hasNoHtmlModifier {
		templateName = fmt.Sprintf("%s.nohtml", templateName)
	}

	return bytes.Replace(content, []byte(tempTemplateName), []byte(templateName), -1)
}

func removeTemplateNamespace(content []byte, tempTemplateName string) []byte {
	return bytes.Replace(content, []byte(fmt.Sprintf("::%s", tempTemplateName)), []byte(""), -1)
}

func formatHtmlNode(node *html.Node, getHtml func(*html.Node) string) string {
	var attributes strings.Builder
	for _, attr := range node.Attr {
		attributes.WriteString(fmt.Sprintf(` %s="%s"`, attr.Key, attr.Val))
	}
	return fmt.Sprintf("<%s%s>%s</%s>", node.Data, attributes.String(), getHtml(node), node.Data)
}

func removeSpace(s string) string {
	rr := make([]rune, 0, len(s))
	for _, r := range s {
		if !unicode.IsSpace(r) {
			rr = append(rr, r)
		}
	}
	return string(rr)
}

func hashID(content string) string {
	content = removeSpace(content)
	xxhash := xxhash.New()
	_, err := xxhash.WriteString(content)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", xxhash.Sum(nil))
}
