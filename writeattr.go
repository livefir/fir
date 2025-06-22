package fir

import (
	"bytes"
	"fmt"
	"strings"

	"slices"

	"github.com/livefir/fir/internal/firattr"
	"golang.org/x/net/html"
)

func addAttributes(content []byte) []byte {
	// First, process x-fir-* attributes and convert them to @fir: attributes
	// Since this is called during rendering, Go templates have been executed
	// so we can safely process x-fir-* attributes without corrupting template syntax
	processedContent, err := processRenderAttributes(content)
	if err != nil {
		// If processing fails, continue with original content
		processedContent = content
	}

	doc, err := html.Parse(bytes.NewReader(processedContent))
	if err != nil {
		panic(err)
	}
	writeAttributes(doc)
	return firattr.HTMLNodeToBytes(doc)
}

func writeAttributes(node *html.Node) {
	if node.Type == html.ElementNode {
		firattr.SetKeyToChildren(node, "")

		attrMap := make(map[string]string)
		for _, attr := range node.Attr {
			attrMap[attr.Key] = attr.Val
		}

		for attrKey, attrVal := range attrMap {
			if !strings.HasPrefix(attrKey, "@fir:") && !strings.HasPrefix(attrKey, "x-on:fir:") {
				continue
			}

			eventns := strings.TrimPrefix(attrKey, "@fir:")
			eventns = strings.TrimPrefix(eventns, "x-on:fir:")
			// eventns might have modifiers like .prevent, .stop, .self, .once, .window, .document etc. remove them
			eventnsParts := strings.SplitN(eventns, ".", -1)
			var modifiers string
			if len(eventnsParts) > 0 {
				eventns = eventnsParts[0]
			}
			if len(eventnsParts) > 1 {
				modifierParts := eventnsParts[1:]
				// All modifiers are preserved (no special handling needed)
				modifiers = strings.Join(modifierParts, ".")
			}

			// eventns might have a filter:[e1:ok,e2:ok] containing multiple event:state separated by comma
			eventnsList, filterExists := firattr.GetEventNsList(eventns)
			// if filter exists remove the current attribute from the node
			if filterExists {
				firattr.RemoveAttr(node, attrKey)
			}

			for _, eventns := range eventnsList {
				if strings.Contains(eventns, ":pending") || strings.Contains(eventns, ":done") {
					// remove template from the eventns if it exists
					parts := strings.Split(eventns, "::")
					if len(parts) == 2 {
						eventns = parts[0]
					}
				}
				eventns = strings.TrimSpace(eventns)

				// set @fir|x-on:fir:eventns attribute to the node
				eventnsWithModifiers := fmt.Sprintf("%s.%s", eventns, modifiers)
				if len(modifiers) == 0 {
					eventnsWithModifiers = eventns
				}
				atFirOk := firattr.HasAttr(node, fmt.Sprintf("@fir:%s", eventnsWithModifiers))
				xOnFirOk := firattr.HasAttr(node, fmt.Sprintf("x-on:fir:%s", eventnsWithModifiers))
				// if the node already has @fir:x attribute, then skip
				if !atFirOk && !xOnFirOk {
					firattr.SetAttr(node, fmt.Sprintf("@fir:%s", eventnsWithModifiers), attrVal)
				}

				// set class fir-myevent-ok--myblock
				key := firattr.GetAttr(node, "fir-key")
				targetClass := fmt.Sprintf("fir-%s", firattr.GetClassNameWithKey(eventns, &key))
				classes := strings.Fields(firattr.GetAttr(node, "class"))
				if !slices.Contains(classes, targetClass) {
					classes = append(classes, targetClass)
					firattr.RemoveAttr(node, "class")
					node.Attr = append(node.Attr, html.Attribute{Key: "class", Val: strings.Join(classes, " ")})
				}
			}

		}
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		writeAttributes(c)
	}

}
