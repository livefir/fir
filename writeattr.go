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

// Recursive function to set the "fir-key" attribute to all nested children
func setKeyToChildren(node *html.Node, key string) {
	if node == nil || node.Type != html.ElementNode {
		return
	}

	if key == "" {
		for _, attr := range node.Attr {
			if attr.Key == "fir-key" {
				key = attr.Val
				break
			}
		}
	} else {
		for _, attr := range node.Attr {
			if attr.Key == "fir-key" {
				if key != attr.Val {
					setKeyToChildren(node, attr.Val)
				}
				break
			}
		}
	}

	if key == "" {
		return
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		setKeyToChildren(child, key)

		if child.Type == html.ElementNode {
			hasPrefixAttribute := false
			for _, attr := range child.Attr {
				if strings.HasPrefix(attr.Key, "@") || strings.HasPrefix(attr.Key, "x-on") {
					hasPrefixAttribute = true
					break
				}
			}

			if !hasPrefixAttribute {
				continue
			}

			hasKeyAttribute := false
			for _, attr := range child.Attr {
				if attr.Key == "fir-key" {
					hasKeyAttribute = true
					break
				}
			}

			if !hasKeyAttribute {
				child.Attr = append(child.Attr, html.Attribute{Key: "fir-key", Val: key})
			}
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
		setKeyToChildren(node, "")

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
				key := getAttr(node, "fir-key")
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
