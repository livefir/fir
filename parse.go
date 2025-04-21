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

const FirGenTemplatePrefix = "::fir-gen-templ-"
const FirDefaultTemplatePrefix = "fir-"

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
		if t == nil {
			t = template.New(fi.name)
		}

		tmpl := t
		if fi.name != t.Name() {
			tmpl = t.New(fi.name)
		}

		if _, err := tmpl.Parse(s); err != nil {
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

// processRenderAttributes parses x-fir-render attributes, collects associated x-fir-action-* attributes,
// translates them into the canonical @fir:... syntax, and replaces the original attributes on the node.
func processRenderAttributes(content []byte) ([]byte, error) {
	if len(content) == 0 {
		return content, nil
	}
	// Optimization: Check if the relevant attributes likely exist before parsing
	if !bytes.Contains(content, []byte("x-fir-render")) && !bytes.Contains(content, []byte("x-fir-action-")) {
		return content, nil // Return original content if attributes are not present
	}

	doc, err := html.Parse(bytes.NewReader(content))
	if err != nil {
		// If parsing fails, it might be an invalid fragment. Return original content for now.
		// Consider logging this error if it's unexpected.
		// logger.Debugf("HTML parsing failed for render attributes, returning original: %v", err)
		return content, nil
		// Or return the error if strict parsing is required:
		// return nil, fmt.Errorf("error parsing HTML content for render attributes: %w", err)
	}

	var traverseErr error
	var modified bool // Flag to track if any changes were made
	var traverse func(*html.Node)

	traverse = func(n *html.Node) {
		if traverseErr != nil {
			return // Stop traversal if an error occurred
		}

		if n.Type == html.ElementNode {
			actionsMap := make(map[string]string)
			attrsToRemove := make(map[string]struct{})
			var newAttrs []html.Attribute
			var renderAttr *html.Attribute // Store the attribute itself to check presence

			// First pass: identify attributes to process and remove
			for i := range n.Attr {
				attr := &n.Attr[i] // Use pointer
				if attr.Key == "x-fir-render" {
					renderAttr = attr // Found the attribute
					attrsToRemove[attr.Key] = struct{}{}
				} else if strings.HasPrefix(attr.Key, "x-fir-action-") {
					actionKey := strings.TrimPrefix(attr.Key, "x-fir-action-")
					if actionKey != "" {
						actionsMap[actionKey] = attr.Val
						attrsToRemove[attr.Key] = struct{}{} // Also remove action attributes
					}
				}
			}

			// Keep attributes that are not being removed
			for _, attr := range n.Attr {
				if _, found := attrsToRemove[attr.Key]; !found {
					newAttrs = append(newAttrs, attr)
				}
			}

			// If x-fir-render attribute was present (even if empty), translate and add new attributes
			if renderAttr != nil { // Check for presence, not non-empty value
				modified = true                                                          // Mark that we made changes
				translated, err := TranslateRenderExpression(renderAttr.Val, actionsMap) // Pass actionsMap
				if err != nil {
					// Store the error and stop further processing
					traverseErr = fmt.Errorf("error translating render expression for node %s, expr '%s': %w", n.Data, renderAttr.Val, err)
					return
				}

				// Split translated output into individual attributes (lines)
				translatedAttrs := strings.Split(translated, "\n")
				for _, translatedAttr := range translatedAttrs {
					if translatedAttr == "" {
						continue
					}
					parts := strings.SplitN(translatedAttr, "=", 2)
					if len(parts) == 2 {
						key := parts[0]
						// Remove surrounding quotes from the value
						val := strings.Trim(parts[1], `"`)
						newAttrs = append(newAttrs, html.Attribute{Key: key, Val: val})
					} else {
						// Log or handle malformed translation? For now, just skip.
						logger.Warnf("Skipping malformed translated attribute: %s", translatedAttr)
					}
				}
				// Replace the node's attributes
				n.Attr = newAttrs
			}
		}

		// Recursively traverse children
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
			if traverseErr != nil {
				return // Propagate error up
			}
		}
	}

	traverse(doc)
	if traverseErr != nil {
		return nil, traverseErr
	}

	// Only render if modifications were actually made
	if !modified {
		return content, nil // Return original content if no changes
	}

	var buf bytes.Buffer
	// Iterate through the direct children of the parsed document node.
	// html.Parse typically creates a structure like:
	// Document -> [comment] -> <html> -> <head> -> ...
	//                                  -> <body> -> [div] -> ...
	//                                           -> [comment] -> ...
	// We want to render the nodes outside <html> (like leading comments)
	// and the *children* of <head> and <body> inside <html>.
	for node := doc.FirstChild; node != nil; node = node.NextSibling {
		if node.Type == html.ElementNode && node.Data == "html" {
			// If this node is the <html> element, iterate through its children (<head>, <body>)
			for htmlChild := node.FirstChild; htmlChild != nil; htmlChild = htmlChild.NextSibling {
				// htmlChild is either <head> or <body>
				if htmlChild.Type == html.ElementNode && (htmlChild.Data == "head" || htmlChild.Data == "body") {
					// Render the *children* of <head> and <body>
					for contentChild := htmlChild.FirstChild; contentChild != nil; contentChild = contentChild.NextSibling {
						if err := html.Render(&buf, contentChild); err != nil {
							return nil, fmt.Errorf("error rendering modified HTML fragment node (%s child): %w", htmlChild.Data, err)
						}
					}
				}
				// Ignore other direct children of <html> if any (shouldn't be standard)
			}
		} else {
			// If the node is not the <html> element (e.g., a comment or doctype before <html>),
			// render it directly. This captures leading/trailing comments outside the main structure.
			if err := html.Render(&buf, node); err != nil {
				// Ignore doctype rendering errors if necessary, but generally render all top-level nodes.
				// Example check: if node.Type == html.DoctypeNode { continue }
				return nil, fmt.Errorf("error rendering modified HTML fragment node (non-html root child %v): %w", node.Type, err)
			}
		}
	}

	return buf.Bytes(), nil
}

// ... rest of parse.go ...

// extractTemplates extracts innerHTML content from fir event namespace string and updates the namespace string with a template
func extractTemplates(content []byte) ([]byte, map[string]string, error) {
	var err error
	blocks := make(map[string]string)
	if len(content) == 0 {
		return content, blocks, nil
	}

	// Step 0: Process render attributes
	content, err = processRenderAttributes(content)
	if err != nil {
		return content, blocks, fmt.Errorf("error processing render attributes: %w", err)
	}

	// Step 1: Generate default template names for event bindings
	updatedContent, err := generateDefaultTemplateNames(content)
	if err != nil {
		return content, blocks, fmt.Errorf("error generating default template names: %w", err)
	}

	// Step 2: Parse the content into an HTML document
	doc, err := html.Parse(bytes.NewReader(updatedContent))
	if err != nil {
		return content, blocks, fmt.Errorf("error parsing HTML content: %w", err)
	}

	// Step 3: Validate and update generated template names
	updatedContent, err = validateGeneratedTemplateNames(doc, updatedContent)
	if err != nil {
		return content, blocks, fmt.Errorf("error validating generated template names: %w", err)
	}

	// Step 4: Parse the updated content into a new HTML document
	updatedDoc, err := html.Parse(bytes.NewReader(updatedContent))
	if err != nil {
		return content, blocks, fmt.Errorf("error reparsing updated HTML content: %w", err)
	}

	// Step 5: Extract template blocks
	extractTemplateBlocks(updatedDoc, blocks)

	return updatedContent, blocks, nil
}

// event handler format for a single event is firPrefix:event:state::templateName.modifier="any string"
// event handler format for multiple events is firPrefix:[event1:state,eventN:state]::templateName.modifier="any string"
// where firPrefix is @fir or x-on:fir
// where event is a an alphanumeric string
// where state is a valid FirEventState. if state is not specified, it defaults to ok
// where templateName is a valid go html template name
// where modifier is a valid FirEventModifier, if modifier is not specified, it defaults to replace
// where "any string" is a valid go html template string
var (
	stateErrorRegex1 = regexp.MustCompile(fmt.Sprintf(`:%s=`, StateError))
	stateErrorRegex2 = regexp.MustCompile(fmt.Sprintf(`:%s]=`, StateError))
	stateOKRegex1    = regexp.MustCompile(fmt.Sprintf(`:%s=`, StateOK))
	stateOKRegex2    = regexp.MustCompile(fmt.Sprintf(`:%s]=`, StateOK))
)

// generateDefaultTemplateNames adds default template names to the event bindings
// default template is added if the event binding does not have a template
// e.g. @fir:create="any string" will be converted to @fir:create::fir-gen-templ-<random string>="any string"
func generateDefaultTemplateNames(content []byte) ([]byte, error) {
	if len(content) == 0 {
		return content, nil
	}
	// check if template is missing and add default template
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
		return fmt.Sprintf(":%s%s%s=", prefix, FirGenTemplatePrefix, strings.ToLower(shortid.MustGenerate()))
	}
}

// validateGeneratedTemplateNames traverses the HTML document and replaces generated template names in the content
// with a new template name(fir-<random-id>) if the content is a valid HTML template.
func validateGeneratedTemplateNames(doc *html.Node, content []byte) ([]byte, error) {
	var replacer func(*html.Node)
	replacer = func(node *html.Node) {
		if node.Type == html.ElementNode {
			for _, attr := range node.Attr {
				if isFirEvent(attr.Key) && strings.Contains(attr.Key, FirGenTemplatePrefix) {
					block := getHtmlContent(node)
					generatedTemplateName := extractTemplateName(attr.Key)
					hasNoHtmlModifier := strings.Contains(generatedTemplateName, string(ModifierNoHTML))
					// if the content is a valid HTML template, replace the template name with fir-<random string>
					// if the content is not a valid HTML template, remove the generated template name
					if isHtmlTemplate(block) {
						content = replaceGeneratedTemplateName(content, generatedTemplateName, block, hasNoHtmlModifier)
					} else {
						content = removeTemplateNamespace(content, generatedTemplateName)
					}
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			replacer(child)
		}
	}

	replacer(doc)
	return content, nil
}

// extractTemplateBlocks traverses the HTML document and extracts template blocks into the provided map.
func extractTemplateBlocks(doc *html.Node, blocks map[string]string) {
	var extractor func(*html.Node)

	extractor = func(node *html.Node) {
		if node.Type == html.ElementNode {
			for _, attr := range node.Attr {
				if isFirEvent(attr.Key) && strings.Contains(attr.Key, "::"+FirDefaultTemplatePrefix) {
					if block := getHtmlContent(node); isHtmlTemplate(block) {
						blocks[extractTemplateName(attr.Key)] = block
					}
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			extractor(child)
		}
	}

	extractor(doc)
}

// Helper function to extract inner HTML content of a node.
func getHtmlContent(node *html.Node) string {
	var block strings.Builder
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			block.WriteString(c.Data)
		} else {
			block.WriteString(formatHtmlNode(c, getHtmlContent))
		}
	}
	return block.String()
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

// replaceGeneratedTemplateName replaces the template name in the content with a new template name
// and adds a ".nohtml" modifier if specified.
func replaceGeneratedTemplateName(content []byte, generatedTemplateName, block string, hasNoHtmlModifier bool) []byte {
	if !bytes.Contains(content, []byte(generatedTemplateName)) {
		return content
	}

	templateName := fmt.Sprintf("%s%s", FirDefaultTemplatePrefix, hashID(block))
	if hasNoHtmlModifier {
		templateName = fmt.Sprintf("%s.%s", templateName, ModifierNoHTML)
	}

	return bytes.Replace(content, []byte(generatedTemplateName), []byte(templateName), -1)
}

func removeTemplateNamespace(content []byte, generatedTemplateName string) []byte {
	return bytes.Replace(content, []byte(fmt.Sprintf("::%s", generatedTemplateName)), []byte(""), -1)
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
