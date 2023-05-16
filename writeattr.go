package fir

import (
	"fmt"
	"strings"

	"github.com/valyala/bytebufferpool"
	"golang.org/x/exp/slices"
	"golang.org/x/net/html"
)

func htmlNodeToBytes(n *html.Node) []byte {
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)
	html.Render(buf, n)
	return buf.Bytes()
}

func setKeyAttr(n *html.Node, key string) {
	if n.Type == html.ElementNode {
		if !hasAlpineAttr(n) {
			return
		}

		n.Attr = append(n.Attr, html.Attribute{Key: "key", Val: key})
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		setKeyAttr(c, key)
	}
}

func removeAttr(n *html.Node, attr string) {
	if n.Type == html.ElementNode {
		for i, a := range n.Attr {
			if a.Key == attr {
				n.Attr = append(n.Attr[:i], n.Attr[i+1:]...)
				break
			}
		}
	}
}

func hasAlpineAttr(n *html.Node) bool {
	if n.Type == html.ElementNode {
		for _, a := range n.Attr {
			if strings.HasPrefix(a.Key, "@") || strings.HasPrefix(a.Key, "x-on") {
				return true
			}
		}
	}
	return false
}

func hasAttr(n *html.Node, attr string) bool {
	if n.Type == html.ElementNode {
		for _, a := range n.Attr {
			if a.Key == attr {
				return true
			}
		}
	}
	return false
}

func setAttr(n *html.Node, key, val string) {
	if n.Type == html.ElementNode {
		for i, a := range n.Attr {
			if a.Key == key {
				n.Attr[i].Val = val
				break
			}
		}
	}
}

func getAttr(n *html.Node, key string) string {
	if n.Type == html.ElementNode {
		for _, a := range n.Attr {
			if a.Key == key {
				return a.Val
			}
		}
	}
	return ""
}

func addClass(node *html.Node, class string) {
	if node.Type == html.ElementNode {
		for i, attr := range node.Attr {
			if attr.Key == "class" {
				classes := strings.Fields(attr.Val)
				if !slices.Contains(classes, class) {
					classes = append(classes, class)
					node.Attr[i].Val = strings.Join(classes, " ")
				}
				break
			}
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		addClass(child, class)
	}
}

func addAttributes(content []byte) []byte {
	doc, err := html.Parse(strings.NewReader(string(content)))
	if err != nil {
		return content
	}
	writeAttributes(doc)
	return htmlNodeToBytes(doc)
}

func writeAttributes(node *html.Node) {
	if node.Type == html.ElementNode {
		for _, attr := range node.Attr {
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
				removeAttr(node, attr.Key)
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
					setAttr(node, fmt.Sprintf("@fir:%s", eventnsWithModifiers), attr.Val)
				}

				// fir-myevent-ok--myblock
				key := getAttr(node, "key")
				classname := fmt.Sprintf("fir-%s", getClassNameWithKey(eventns, &key))
				addClass(node, classname)
				if key == "" {
					continue
				}
				setKeyAttr(node, key)
			}

		}
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		writeAttributes(c)
	}

}
