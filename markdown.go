package fir

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/livefir/fir/internal/file"

	embed "github.com/13rac1/goldmark-embed"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/livefir/fir/internal/logger"
	"github.com/valyala/bytebufferpool"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
)

type mdcache struct {
	values map[string]string
	sync.RWMutex
}

func (c *mdcache) get(key string) (string, bool) {
	c.RLock()
	defer c.RUnlock()
	value, ok := c.values[key]
	return value, ok
}

func (c *mdcache) set(key string, value string) {
	c.Lock()
	defer c.Unlock()
	c.values[key] = value
}

type filecache struct {
	files       map[string][]byte
	etags       map[string]string
	lastChecked map[string]time.Time
	sync.RWMutex
}

func (c *filecache) getFile(key string) ([]byte, bool) {
	c.RLock()
	defer c.RUnlock()
	value, ok := c.files[key]
	return value, ok
}

func (c *filecache) setFile(key string, value []byte) {
	c.Lock()
	defer c.Unlock()
	c.files[key] = value
}

func (c *filecache) getEtag(key string) string {
	c.RLock()
	defer c.RUnlock()
	value, ok := c.etags[key]
	if ok {
		return value
	}
	return ""
}

func (c *filecache) setEtag(key string, val string) {
	c.Lock()
	defer c.Unlock()
	c.etags[key] = val
}

func (c *filecache) getLastChecked(key string) time.Time {
	c.RLock()
	defer c.RUnlock()
	value, ok := c.lastChecked[key]
	if ok {
		return value
	}
	return time.Time{}
}

func (c *filecache) setLastChecked(key string, value time.Time) {
	c.Lock()
	defer c.Unlock()
	c.lastChecked[key] = value
}

func md5Key(in string, markers []string) string {
	hash := md5.Sum([]byte(in))
	checksum := hex.EncodeToString(hash[:])
	if len(markers) == 0 {
		return checksum
	}
	return checksum + "-" + strings.Join(markers, "-")
}

func fetchFileEtag(url string) string {
	// make a head request to the url and  get the etag from header
	// if etag is not present, return empty string
	// if etag is present, return etag

	// Create a new HEAD request
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		logger.Errorf("error creating request: %v", err)
		return ""
	}

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Errorf("error sending request: %v", err)
		return ""
	}
	defer resp.Body.Close()

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		logger.Errorf("request failed with status: %v", resp.Status)
		return ""
	}

	return resp.Header.Get("ETag")

}

func fetchFile(url string) ([]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		logger.Errorf("failed to fetch the file: %v\n", err)
		return nil, err
	}
	defer response.Body.Close()

	// Copy the content from the remote response to the local file
	data, err := io.ReadAll(response.Body)
	if err != nil {
		logger.Errorf("failed to save the file: %v\n", err)
		return nil, err
	}

	return data, nil
}

func isValidURL(input string) bool {
	_, err := url.ParseRequestURI(input)
	if err != nil {
		return false
	}

	u, err := url.Parse(input)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return u.Scheme == "http" || u.Scheme == "https"
}

func markdown(readFile file.ReadFileFunc, existFile file.ExistFileFunc) func(in string, markers ...string) string {
	cachemd := &mdcache{
		values: make(map[string]string),
	}

	cachefile := &filecache{
		files:       make(map[string][]byte),
		etags:       make(map[string]string),
		lastChecked: make(map[string]time.Time),
	}

	mdparser := markdownParser()

	return func(in string, markers ...string) string {
		var indata []byte
		var isFile bool

		if isValidURL(in) {
			fkey := md5Key(in, nil)
			var err error

			f, ok := cachefile.getFile(fkey)
			if !ok || time.Since(cachefile.getLastChecked(fkey)) > time.Minute*5 {
				etag := fetchFileEtag(in)
				cachefile.setLastChecked(fkey, time.Now())
				if etag != cachefile.getEtag(fkey) || (etag == "" && cachefile.getEtag(fkey) == "") {
					f, err = fetchFile(in)
					if err != nil {
						logger.Errorf("%v", err)
						return string("error fetching file")
					}
					cachefile.setEtag(fkey, etag)
					cachefile.setFile(fkey, f)
				}
			}
			// enclose the file in a code block
			ext := path.Ext(in)
			indata = []byte("```" + ext + "\n" + string(f) + "```")
			isFile = true

		} else {
			if existFile(in) {
				_, data, err := readFile(in)
				if err != nil {
					logger.Errorf("%v", err)
					return string("error reading file")
				}
				indata = data
				isFile = true
			} else {
				indata = []byte(in)
			}
		}
		// check if snippet is already in cache
		key := md5Key(in, markers)
		if value, ok := cachemd.get(key); ok {
			return string(value)
		}

		indata = snippets(indata, markers)

		buf := bytebufferpool.Get()
		defer bytebufferpool.Put(buf)
		if err := mdparser.Convert(indata, buf); err != nil {
			logger.Errorf("%v", err)
			return string("error converting to markdown")
		}
		result := buf.String()
		if isFile {
			cachemd.set(key, result)
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
		highlighting.WithGuessLanguage(true),
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
