package fir

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"strings"

	"github.com/goccy/go-json"

	"github.com/alecthomas/chroma/formatters/html"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"

	"github.com/davecgh/go-spew/spew"

	"github.com/Masterminds/sprig/v3"
)

func defaultFuncMap() template.FuncMap {
	allFuncs := make(template.FuncMap)
	for k, v := range sprig.FuncMap() {
		allFuncs[k] = v
	}
	allFuncs["bytesToMap"] = bytesToMap
	allFuncs["bytesToString"] = bytesToString
	allFuncs["dump"] = dump
	allFuncs["toJsonb64"] = toJsonb64
	allFuncs["textAreaRows"] = textAreaRows
	return allFuncs
}

func textAreaRows(s string) int {
	l := len(strings.Split(s, "\n")) + 1
	if l < 2 {
		return 2
	}
	if l > 15 {
		return 15
	}
	return l
}

func toJsonb64(data interface{}) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(jsonData), nil
}

func bytesToMap(data []byte) map[string]any {
	m := make(map[string]any)
	err := json.Unmarshal(data, &m)
	if err != nil {
		panic(err)
	}
	return m
}

func bytesToString(data []byte) string {
	return string(data)
}

func dump(val any) (template.HTML, error) {
	var buf bytes.Buffer
	defer buf.Reset()
	err := highlight(&buf, spew.Sdump(val), "dracula")
	if err != nil {
		return "", err
	}
	return template.HTML(fmt.Sprintf("<code>%v</code>", buf.String())), nil
}

func highlight(w io.Writer, source, style string) error {
	// Determine lexer.
	l := lexers.Get("go")
	l = chroma.Coalesce(l)

	// Determine formatter.
	f := html.New(html.WithClasses(false))

	// Determine style.
	s := styles.Get(style)
	if s == nil {
		s = styles.Fallback
	}

	it, err := l.Tokenise(nil, source)
	if err != nil {
		return err
	}
	return f.Format(w, s, it)
}
