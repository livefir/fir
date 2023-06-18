package fir

import (
	"bytes"
	"html/template"

	embed "github.com/13rac1/goldmark-embed"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"k8s.io/klog/v2"
)

var mdparser = markdownParser()

func markdown(readFile readFileFunc, existFile existFileFunc) func(in string, linenum ...int) template.HTML {
	return func(in string, linenum ...int) template.HTML {
		var indata []byte
		if existFile(in) {
			_, data, err := readFile(in)
			if err != nil {
				klog.Errorln(err)
				return ""
			}
			indata = data
		} else {
			indata = []byte(in)
		}

		if len(linenum) > 0 {
			min := linenum[0]
			var max int
			if len(linenum) > 1 {
				max = linenum[1]
			}

			parts := bytes.SplitN(indata, []byte("\n"), -1)
			if max > len(parts) || max == 0 {
				max = len(parts)
			}

			chunk := parts[min:max]
			indata = bytes.Join(chunk, []byte("\n"))
		}

		var buf bytes.Buffer
		if err := mdparser.Convert(indata, &buf); err != nil {
			klog.Errorln(err)
			return ""
		}
		return template.HTML(buf.String())
	}
}

func markdownParser() goldmark.Markdown {
	var (
		extensions      []goldmark.Extender
		parserOptions   []parser.Option
		rendererOptions []renderer.Option
	)

	rendererOptions = append(rendererOptions, html.WithHardWraps())
	rendererOptions = append(rendererOptions, html.WithXHTML())
	rendererOptions = append(rendererOptions, html.WithUnsafe())

	extensions = append(extensions, extension.Table)
	extensions = append(extensions, extension.Strikethrough)
	extensions = append(extensions, extension.Linkify)
	extensions = append(extensions, extension.TaskList)
	extensions = append(extensions, extension.Typographer)
	extensions = append(extensions, extension.DefinitionList)
	extensions = append(extensions, extension.Footnote)
	extensions = append(extensions, extension.GFM)
	extensions = append(extensions, extension.CJK)
	extensions = append(extensions, embed.New())

	parserOptions = append(parserOptions, parser.WithAutoHeadingID())
	parserOptions = append(parserOptions, parser.WithAttribute())

	md := goldmark.New(
		goldmark.WithExtensions(
			extensions...,
		),
		goldmark.WithParserOptions(
			parserOptions...,
		),
		goldmark.WithRendererOptions(
			rendererOptions...,
		),
	)

	return md
}
