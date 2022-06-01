package controller

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/yosssi/gohtml"
)

type Op string

const (
	Morph       Op = "morph"
	Reload      Op = "reload"
	UpdateStore Op = "update-store"
)

type Operation struct {
	Op       Op          `json:"op"`
	Selector string      `json:"selector"`
	Value    interface{} `json:"value"`
}

func (m *Operation) Bytes() []byte {
	b, err := json.Marshal(m)
	if err != nil {
		log.Printf("error marshalling dom %v\n", err)
		return nil
	}
	return b
}

type Data map[string]any

type Socket interface {
	Event() Event
	Request() *http.Request
	ResponseWriter() http.ResponseWriter
	Morph(selector, template string, data Data)
	Store(...string) Storer
	Reload()
}

type Storer interface {
	Update(any)
	UpdateProp(string, any)
}

type store struct {
	names []string
	wc    *websocketController
	topic string
}

func (s *store) Update(v any) {

	data := map[string]any{
		"op":    UpdateStore,
		"value": v,
	}

	for _, name := range s.names {
		data["selector"] = name
		s.wc.writeJSON(s.topic, data)
	}

}

func (s *store) UpdateProp(k string, v any) {
	s.Update(map[string]any{k: v})
}

type Event struct {
	ID     string          `json:"id"`
	Params json.RawMessage `json:"params"`
}

func (e Event) String() string {
	data, _ := json.MarshalIndent(e, "", " ")
	return string(data)
}

type EventHandler func(s Socket) error

func (e Event) DecodeParams(v any) error {
	return json.NewDecoder(bytes.NewReader(e.Params)).Decode(v)
}

type socket struct {
	event        Event
	r            *http.Request
	w            http.ResponseWriter
	rootTemplate *template.Template
	topic        string
	wc           *websocketController
}

func (s socket) setError(userMessage string, errs ...error) {
	if len(errs) != 0 {
		var errstrs []string
		for _, err := range errs {
			if err == nil {
				continue
			}
			errstrs = append(errstrs, err.Error())
		}
		log.Printf("err: %v, errors: %v\n", userMessage, strings.Join(errstrs, ","))
	}

	s.Morph("#glv-error", "glv-error", Data{"error": userMessage})

}

func (s socket) unsetError() {
	s.Morph("#glv-error", "glv-error", nil)
}

func (s socket) Event() Event {
	return s.event
}

func (s socket) Request() *http.Request {
	return s.r
}

func (s socket) ResponseWriter() http.ResponseWriter {
	return s.w
}

func (s socket) Store(names ...string) Storer {
	if len(names) == 0 {
		names = append(names, "fir")
	}
	return &store{names: names, wc: s.wc, topic: s.topic}
}

func (s socket) Morph(selector, template string, data Data) {
	var buf bytes.Buffer
	err := s.rootTemplate.ExecuteTemplate(&buf, template, data)
	if err != nil {
		log.Printf("err %v with data => \n %+v\n", err, getJSON(data))
		return
	}
	if s.wc.debugLog {
		log.Printf("rendered template %+v, with data => \n %+v\n", template, getJSON(data))
	}
	html := buf.String()
	if s.wc.enableHTMLFormatting {
		html = gohtml.Format(html)
	}
	buf.Reset()

	m := &Operation{
		Op:       Morph,
		Selector: selector,
		Value:    html,
	}
	s.wc.message(s.topic, m.Bytes())
}

func (s socket) Reload() {
	m := &Operation{
		Op: Reload,
	}
	s.wc.message(s.topic, m.Bytes())
}

func getJSON(data Data) string {
	b, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return err.Error()
	}
	return string(b)
}
