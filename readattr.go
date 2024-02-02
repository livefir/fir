package fir

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"slices"

	"github.com/livefir/fir/internal/logger"
	"github.com/sourcegraph/conc/pool"
	"golang.org/x/net/html"
)

func readAttributes(fi fileInfo) fileInfo {
	doc, err := html.Parse(bytes.NewReader(fi.content))
	if err != nil {
		panic(err)
	}

	attributes := firAttributes(doc)
	resultPool := pool.NewWithResults[eventTemplates]()
	for _, attr := range attributes {
		attr := attr
		resultPool.Go(func() eventTemplates {
			return eventTemplatesFromAttr(attr)
		})
	}
	evtArr := resultPool.Wait()
	evt := make(eventTemplates)
	for _, evtMap := range evtArr {
		evt = deepMergeEventTemplates(evt, evtMap)
	}

	return fileInfo{
		name:           fi.name,
		content:        fi.content,
		err:            fi.err,
		eventTemplates: evt,
	}
}

func eventTemplatesFromAttr(attr html.Attribute) eventTemplates {
	evt := make(eventTemplates)
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
			logger.Errorf(eventFormatError(eventns))
			continue
		}

		// myevent:ok
		eventID := eventnsParts[0]
		// [myevent, ok]
		eventIDParts := strings.SplitN(eventID, ":", -1)
		if len(eventIDParts) != 2 {
			logger.Errorf(eventFormatError(eventns))
			continue
		}
		// event name can only be followed by ok, error, pending, done
		if !slices.Contains([]string{"ok", "error", "pending", "done"}, eventIDParts[1]) {
			logger.Errorf(eventFormatError(eventns))
			continue
		}
		// assert myevent:ok::myblock or myevent:error::myblock
		if len(eventnsParts) == 2 && !slices.Contains([]string{"ok", "error"}, eventIDParts[1]) {
			logger.Errorf(eventFormatError(eventns))
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
			logger.Errorf("error: invalid template name in event binding: only hyphen(-) and colon(:) are allowed: %v", templateName)
			continue
		}

		templates[templateName] = struct{}{}
		// fmt.Printf("eventID: %s, templateName: %s", eventID, templateName)

		evt[eventID] = templates
	}

	return evt

}

// checks if the event string is of the format [event1:ok,event2:ok]:tmpl1 and returns the unbundled list of event strings
// event1:ok:tmpl1,event2:ok:tmpl1. if not, returns original event string
func getEventNsList(input string) ([]string, bool) {
	ef, err := getEventFilter(input)
	if err != nil {
		logger.Errorf("error parsing event filter: %v", err)
		return []string{input}, false
	}
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

var ErrorEventFilterFormat = fmt.Errorf("error parsing event filter. must match ^[a-zA-Z0-9-]+:(ok|pending|error|done)$")

type eventFilter struct {
	BeforeBracket string
	Values        []string
	AfterBracket  string
}

func getEventFilter(input string) (*eventFilter, error) {
	// Extract the part of the string before the open square bracket
	beforeRe := regexp.MustCompile(`^(.*?)\[`)
	beforeMatch := beforeRe.FindStringSubmatch(input)

	var beforeBracket string
	if len(beforeMatch) == 2 {
		beforeBracket = beforeMatch[1]
	}

	// Extract the part of the string after the closed square bracket
	afterRe := regexp.MustCompile(`\](.*)$`)
	afterMatch := afterRe.FindStringSubmatch(input)
	var afterBracket string
	if len(afterMatch) == 2 {
		afterBracket = afterMatch[1]
	}

	// Extract the contents of a closed square bracket
	re := regexp.MustCompile(`\[(.*?)\]`)
	matches := re.FindStringSubmatch(input)
	if len(matches) < 2 {
		return nil, nil
	}

	// Remove whitespace and split the contents by comma
	contents := strings.ReplaceAll(matches[1], " ", "")
	values := strings.Split(contents, ",")

	// Validate and format each value
	validValues := make([]string, 0)
	for _, value := range values {
		if !isValidValue(value) {
			return nil, ErrorEventFilterFormat
		}
		validValues = append(validValues, formatValue(value))
	}

	extractedValues := &eventFilter{
		BeforeBracket: beforeBracket,
		Values:        validValues,
		AfterBracket:  afterBracket,
	}

	return extractedValues, nil
}

func isValidValue(value string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9-]+:(ok|pending|error|done)$`)
	return re.MatchString(value)
}

func formatValue(value string) string {
	parts := strings.Split(value, ":")
	return fmt.Sprintf("%s:%s", parts[0], parts[1])
}

func getClassNameWithKey(eventns string, key *string) string {
	cls := getClassName(eventns)
	if key != nil && *key != "" {
		cls = cls + "--" + strings.ReplaceAll(*key, " ", "-")
	}
	return cls
}

func getClassName(eventns string) string {
	return strings.ReplaceAll(eventns, ":", "-")
}

func firAttributes(n *html.Node) []html.Attribute {
	var attributes []html.Attribute
	if n.Type == html.ElementNode {
		for _, attr := range n.Attr {
			if !strings.HasPrefix(attr.Key, "@fir:") && !strings.HasPrefix(attr.Key, "x-on:fir:") {
				continue
			}
			attributes = append(attributes, attr)
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		attributes = append(firAttributes(c), attributes...)
	}
	return attributes
}

func eventFormatError(eventns string) string {
	return fmt.Sprintf(`
	error: invalid event namespace: %s. must be of either of the three formats =>
	1. @fir:<event>:<state:ok|error|pending|done>::<block-name(optional)>
	2. @fir:[event1:state,event2:state]::<block-name(optional)>
	`, eventns)
}
