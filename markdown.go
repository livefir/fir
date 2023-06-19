package fir

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"

	embed "github.com/13rac1/goldmark-embed"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/valyala/bytebufferpool"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"k8s.io/klog/v2"
)

type cachemd struct {
	values map[string]string
	sync.RWMutex
}

func (c *cachemd) get(key string) (string, bool) {
	c.RLock()
	defer c.RUnlock()
	value, ok := c.values[key]
	return value, ok
}

func (c *cachemd) set(key string, value string) {
	c.Lock()
	defer c.Unlock()
	c.values[key] = value
}

func hashKey(in string, markers []string) string {
	hash := md5.Sum([]byte(in))
	checksum := hex.EncodeToString(hash[:])
	if len(markers) == 0 {
		return checksum
	}
	return checksum + "-" + strings.Join(markers, "-")
}

func markdown(readFile readFileFunc, existFile existFileFunc) func(in string, markers ...string) string {
	cache := &cachemd{
		values: make(map[string]string),
	}

	mdparser := markdownParser()

	return func(in string, markers ...string) string {
		var indata []byte
		var isFile bool
		if existFile(in) {
			_, data, err := readFile(in)
			if err != nil {
				klog.Errorln(err)
				return string("error reading file")
			}
			indata = data
		} else {
			indata = []byte(in)
		}
		// check if snippet is already in cache
		key := hashKey(in, markers)
		if value, ok := cache.get(key); ok {
			return string(value)
		}

		indata = snippets(indata, markers)

		buf := bytebufferpool.Get()
		defer bytebufferpool.Put(buf)
		if err := mdparser.Convert(indata, buf); err != nil {
			klog.Errorln(err)
			return string("error converting to markdown")
		}
		result := buf.String()
		if isFile {
			cache.set(key, result)
		}

		return string(result)
	}
}

func snippets(in []byte, markers []string) []byte {
	if len(markers) == 0 {
		return in
	}
	var out []byte
	atleastOneMarker := false
	for _, marker := range markers {
		content, valid := snippet(in, marker)
		if len(content) == 0 && !valid {
			continue
		}
		if valid {
			atleastOneMarker = true
		}
		// Add a newline before the snippet if there is already content
		if len(out) > 0 {
			out = append(out, []byte("\n")...)
		}
		out = append(out, content...)
	}

	if len(out) == 0 && !atleastOneMarker {
		return in
	}

	return out
}

func snippet(in []byte, marker string) ([]byte, bool) {
	startMarker := fmt.Sprintf("<!-- start %s -->", marker)
	endMarker := fmt.Sprintf("<!-- end %s -->", marker)

	lines := bytes.Split(in, []byte("\n"))
	start := -1
	end := -1
	for i, line := range lines {
		if bytes.Equal(bytes.TrimSpace(line), []byte(startMarker)) {
			start = i + 1
		}
		if bytes.Equal(bytes.TrimSpace(line), []byte(endMarker)) {
			end = i
		}
	}

	if start == -1 && end == -1 {
		return []byte{}, false
	}

	if start > -1 && end == -1 {
		return bytes.Join(lines[start:], []byte("\n")), true
	}

	if start == -1 && end > -1 {
		return []byte{}, false
	}

	if end < start {
		return []byte{}, false
	}

	if start == end {
		return []byte{}, false
	}

	return bytes.Join(lines[start:end], []byte("\n")), true
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
	extensions = append(extensions, extension.Footnote)
	extensions = append(extensions, highlighting.NewHighlighting(
		highlighting.WithStyle("dracula"),
		highlighting.WithFormatOptions(
			chromahtml.WithLineNumbers(true),
		),
	))

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
