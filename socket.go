package fir

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strings"
)

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
		"op":    updateStore,
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

func (s socket) setEventError(userMessage string, errs ...error) {
	if len(errs) != 0 {
		var errstrs []string
		for _, err := range errs {
			if err == nil {
				continue
			}
			errstrs = append(errstrs, err.Error())
		}
		log.Printf("[controller][error]  %v, errors: %v\n", userMessage, strings.Join(errstrs, ","))
	}

	s.Morph("#fir-event-error", "fir-event-error", Data{"eventError": userMessage})

}

func (s socket) unsetEventError() {
	s.Morph("#fir-event-error", "fir-event-error", Data{"eventError": nil})
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

func (s socket) Morph(selector, tmpl string, data Data) {
	var buf bytes.Buffer
	err := s.rootTemplate.ExecuteTemplate(&buf, tmpl, data)
	if err != nil {
		if s.wc.debugLog {
			log.Printf("[controller][error] %v with data => \n %+v\n", err, getJSON(data))
		}
		return
	}
	if s.wc.debugLog {
		log.Printf("[controller]rendered template %+v, with data => \n %+v\n", tmpl, getJSON(data))
	}
	html := buf.String()
	buf.Reset()

	m := &Operation{
		Op:       morph,
		Selector: selector,
		Value:    html,
	}
	s.wc.message(s.topic, m.Bytes())
}

func (s socket) Reload() {
	m := &Operation{
		Op: reload,
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
