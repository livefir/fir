package fir

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/valyala/bytebufferpool"
	"golang.org/x/exp/slices"
	"golang.org/x/net/html"
)

func htmlNodeToBytes(n *html.Node) []byte {
	return []byte(htmlNodetoString(n))
}

func htmlNodetoString(n *html.Node) string {
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)
	err := html.Render(buf, n)
	if err != nil {
		panic(fmt.Sprintf("failed to render HTML: %v", err))
	}
	return html.UnescapeString(buf.String())
}

// Recursive function to set the "key" attribute to all nested children
func setKeyToChildren(node *html.Node, key string) {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		// Only modify element nodes
		if child.Type == html.ElementNode {
			// Check if the child already has a "key" attribute
			hasKeyAttribute := false
			for _, attr := range child.Attr {
				if attr.Key == "key" {
					hasKeyAttribute = true
					break
				}
			}

			// Check if the child doesn't have a "key" attribute and has an attribute with prefix "@" or "x-on"
			if !hasKeyAttribute {
				hasPrefixAttribute := false
				for _, attr := range child.Attr {
					if strings.HasPrefix(attr.Key, "@") || strings.HasPrefix(attr.Key, "x-on") {
						hasPrefixAttribute = true
						break
					}
				}

				// Set the "key" attribute if the child has a matching attribute
				if hasPrefixAttribute {
					child.Attr = append(child.Attr, html.Attribute{Key: "key", Val: key})
				}
			}

			// Recurse through the child nodes
			setKeyToChildren(child, key)
		}
	}
}

func removeAttr(n *html.Node, attr string) {
	for i, a := range n.Attr {
		if a.Key == attr {
			n.Attr = append(n.Attr[:i], n.Attr[i+1:]...)
			break
		}
	}

}

func hasAttr(n *html.Node, attr string) bool {
	for _, a := range n.Attr {
		if a.Key == attr {
			return true
		}
	}

	return false
}

func setAttr(n *html.Node, key, val string) {
	n.Attr = append(n.Attr, html.Attribute{Key: key, Val: val})
}

func getAttr(n *html.Node, key string) string {

	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}

	return ""
}

func addAttributes(content []byte) []byte {
	doc, err := html.Parse(bytes.NewReader(content))
	if err != nil {
		panic(err)
	}
	writeAttributes(doc)
	return htmlNodeToBytes(doc)
}

func writeAttributes(node *html.Node) {
	if node.Type == html.ElementNode {
		if hasAttr(node, "key") {
			setKeyToChildren(node, getAttr(node, "key"))
		}

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
				modifiers = strings.Join(eventnsParts[1:], ".")
			}

			// eventns might have a filter:[e1:ok,e2:ok] containing multiple event:state separated by comma
			eventnsList, filterExists := getEventNsList(eventns)
			// if filter exists remove the current attribute from the node
			if filterExists {
				removeAttr(node, attrKey)
			}

			for _, eventns := range eventnsList {
				eventns = strings.TrimSpace(eventns)
				// set @fir|x-on:fir:eventns attribute to the node
				eventnsWithModifiers := fmt.Sprintf("%s.%s", eventns, modifiers)
				if len(modifiers) == 0 {
					eventnsWithModifiers = eventns
				}
				atFirOk := hasAttr(node, fmt.Sprintf("@fir:%s", eventnsWithModifiers))
				xOnFirOk := hasAttr(node, fmt.Sprintf("x-on:fir:%s", eventnsWithModifiers))
				// if the node already has @fir:x attribute, then skip
				if !atFirOk && !xOnFirOk {
					setAttr(node, fmt.Sprintf("@fir:%s", eventnsWithModifiers), attrVal)
				}

				// fir-myevent-ok--myblock
				key := getAttr(node, "key")
				firKey := getAttr(node, "fir-key")
				if len(firKey) != 0 {
					key = firKey
				}
				targetClass := fmt.Sprintf("fir-%s", getClassNameWithKey(eventns, &key))
				classes := strings.Fields(getAttr(node, "class"))
				if !slices.Contains(classes, targetClass) {
					classes = append(classes, targetClass)
					removeAttr(node, "class")
					node.Attr = append(node.Attr, html.Attribute{Key: "class", Val: strings.Join(classes, " ")})
				}
			}

		}
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		writeAttributes(c)
	}

}
