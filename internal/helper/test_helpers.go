package helper

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/valyala/bytebufferpool"
	"golang.org/x/net/html"
)

// AreNodesDeepEqual compares two HTML nodes recursively for equality
func AreNodesDeepEqual(node1, node2 *html.Node) error {
	if node1 == nil && node2 == nil {
		return fmt.Errorf("both nodes are nil")
	}

	if node1 == nil || node2 == nil {
		return fmt.Errorf("one of the nodes is nil")
	}

	if node1.Type != node2.Type {
		return fmt.Errorf("node types are not equal (%v != %v)", node1.Type, node2.Type)
	}

	if RemoveSpace(node1.Data) != RemoveSpace(node2.Data) {
		return fmt.Errorf("node data is not equal (%s != %s)", node1.Data, node2.Data)
	}

	if err := AreAttributesEqual(node1.Attr, node2.Attr); err != nil {
		return err
	}

	c1 := node1.FirstChild
	c2 := node2.FirstChild

	for c1 != nil && c2 != nil {
		if err := AreNodesDeepEqual(c1, c2); err != nil {
			return err
		}

		c1 = c1.NextSibling
		c2 = c2.NextSibling

	}

	if c1 != nil && c1.DataAtom.String() != "" {
		return fmt.Errorf("node1 has extra child: atom: %v, val: %v", c1.DataAtom, string(HtmlNodeToBytes(c1)))
	}
	if c2 != nil && c2.DataAtom.String() != "" {
		return fmt.Errorf("node2 has extra child: atom: %v, val: %v", c2.DataAtom, string(HtmlNodeToBytes(c2)))
	}

	return nil
}

// AreAttributesEqual compares two sets of HTML attributes for equality
func AreAttributesEqual(attr1, attr2 []html.Attribute) error {
	attr1Map := make(map[string]string)
	for _, a := range attr1 {
		attr1Map[a.Key] = a.Val
	}

	attr2Map := make(map[string]string)
	for _, a := range attr2 {
		attr2Map[a.Key] = a.Val
	}

	for k, v := range attr1Map {
		if k == "class" {
			if err := AreClassesEqual(v, attr2Map["class"]); err != nil {
				return err
			}
		} else {
			val, ok := attr2Map[k]
			if !ok {
				return fmt.Errorf("attr %v is not present in attr2Map %+v", k, attr2Map)
			}
			if val != v {
				return fmt.Errorf("attr %v has different values: %v != %v", k, val, v)
			}
		}

		delete(attr2Map, k)
	}

	if len(attr2Map) > 0 {
		return fmt.Errorf("attr2Map has extra attributes: %v", attr2Map)
	}

	return nil
}

// AreClassesEqual compares two space-separated class strings for equality
func AreClassesEqual(class1, class2 string) error {
	classSet1 := strings.Fields(class1)
	classSet2 := strings.Fields(class2)

	classMap := make(map[string]bool)
	for _, class := range classSet1 {
		classMap[class] = true
	}

	for _, class := range classSet2 {
		_, ok := classMap[class]
		if !ok {
			return fmt.Errorf("class %v is not present in classSet1", class)
		}
	}

	return nil
}

// RemoveSpace removes whitespace from a string using unicode.IsSpace
func RemoveSpace(s string) string {
	rr := make([]rune, 0, len(s))
	for _, r := range s {
		if !unicode.IsSpace(r) {
			rr = append(rr, r)
		}
	}
	return string(rr)
}

// HtmlNodeToBytes converts an HTML node to bytes
func HtmlNodeToBytes(node *html.Node) []byte {
	return []byte(htmlNodetoString(node))
}

// htmlNodetoString converts an HTML node to string
func htmlNodetoString(n *html.Node) string {
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)
	err := html.Render(buf, n)
	if err != nil {
		panic(fmt.Sprintf("failed to render HTML: %v", err))
	}
	return html.UnescapeString(buf.String())
}
