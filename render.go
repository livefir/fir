package fir

import (
	"bytes"
	"fmt"
	"html/template"
	"io"

	"github.com/golang/glog"
	"github.com/livefir/fir/internal/dom"
	"github.com/livefir/fir/pubsub"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
)

func renderDOMEvents(ctx RouteContext, pubsubEvent pubsub.Event, templateNames []string) []dom.Event {
	eventIDWithState := fmt.Sprintf("%s:%s", *pubsubEvent.ID, pubsubEvent.State)
	var events []dom.Event
	for _, templateName := range templateNames {
		if templateName == "-" {
			events = append(events, dom.Event{
				ID:     *pubsubEvent.ID,
				State:  pubsubEvent.State,
				Type:   fir(eventIDWithState),
				Target: pubsubEvent.Target,
				Detail: pubsubEvent.Detail,
			})
			continue
		}
		value, err := buildTemplateValue(ctx.route.template, templateName, pubsubEvent.Detail)
		if err != nil {
			glog.Errorf("Bindings.Events buildTemplateValue error: %s", err)
			continue
		}
		events = append(events, dom.Event{
			ID:     eventIDWithState,
			State:  pubsubEvent.State,
			Type:   fir(eventIDWithState, templateName),
			Target: pubsubEvent.Target,
			Detail: value,
		})
	}

	return events
}

func buildTemplateValue(t *template.Template, name string, data any) (string, error) {
	if t == nil {
		return "", nil
	}
	if name == "" {
		return "", nil
	}
	var buf bytes.Buffer
	defer buf.Reset()
	if name == "_fir_html" {
		buf.WriteString(data.(string))
	} else {
		t.Option("missingkey=zero")
		err := t.ExecuteTemplate(&buf, name, data)
		if err != nil {
			return "", err
		}
	}

	m := minify.New()
	m.Add("text/html", &html.Minifier{})
	r := m.Reader("text/html", &buf)
	var buf1 bytes.Buffer
	defer buf1.Reset()
	_, err := io.Copy(&buf1, r)
	if err != nil {
		return "", err
	}
	value := buf1.String()
	return value, nil
}
