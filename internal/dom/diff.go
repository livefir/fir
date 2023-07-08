package dom

import (
	"fmt"
	"html/template"
	"strings"
	"text/template/parse"

	"golang.org/x/exp/slices"
)

type Diff struct {
	Dynamic          []string
	Static           []string
	DefinedTemplates map[string]parse.Node
}

func CalcDiff(t *template.Template) *Diff {
	return calcDiff(t.Tree.Root, &Diff{
		DefinedTemplates: definedTemplateNodes(t),
	})
}

var parseableNodes = []parse.NodeType{
	parse.NodeText,
	parse.NodeTemplate,
	parse.NodeList,
	parse.NodeRange,
}

func calcDiff(node parse.Node, diff *Diff) *Diff {
	fmt.Println(node.String())
	if !slices.Contains(parseableNodes, node.Type()) {
		diff.Dynamic = append(diff.Dynamic, node.String())
	}

	if node.Type() == parse.NodeText {
		diff.Static = append(diff.Static, node.String())
	}

	if node.Type() == parse.NodeTemplate {
		n := node.(*parse.TemplateNode)
		if diff.DefinedTemplates != nil && diff.DefinedTemplates[n.Name] != nil {
			diff = calcDiff(diff.DefinedTemplates[n.Name], diff)
		}
	}

	if node.Type() == parse.NodeRange {
		rn := node.(*parse.RangeNode)
		for _, n := range rn.List.Nodes {
			diff = calcDiff(n, diff)
		}
	}

	if ln, ok := node.(*parse.ListNode); ok {
		for _, n := range ln.Nodes {
			diff = calcDiff(n, diff)
		}
	}
	return diff
}

func definedTemplateNodes(t *template.Template) map[string]parse.Node {
	if t.DefinedTemplates() == "" {
		return nil
	}
	s := strings.TrimPrefix(t.DefinedTemplates(), "; defined templates are: ")
	if s == "" {
		return nil
	}
	s = strings.ReplaceAll(s, `"`, "")
	s = strings.ReplaceAll(s, " ", "")
	fmt.Println("defined templates string ", s)
	tmplNames := strings.FieldsFunc(s, func(r rune) bool {
		return r == ','
	})

	fmt.Printf("defined templates: %q \n ", tmplNames)

	nodes := make(map[string]parse.Node)
	for _, tmplName := range tmplNames {
		dt := t.Lookup(tmplName)
		if dt == nil {
			continue
		}
		nodes[tmplName] = dt.Tree.Root
	}

	return nodes
}
