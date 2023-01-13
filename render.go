package fir

import (
	"bytes"
	"fmt"
	"html/template"
	"io"

	"github.com/golang/glog"
	"github.com/livefir/fir/internal/dom"
	"github.com/livefir/fir/internal/eventstate"
	"github.com/livefir/fir/pubsub"
	"github.com/patrickmn/go-cache"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
)

func renderDOMEvents(ctx RouteContext, pubsubEvent pubsub.Event) []dom.Event {
	eventIDWithState := fmt.Sprintf("%s:%s", *pubsubEvent.ID, pubsubEvent.State)
	templateNames := ctx.route.bindings.GetTemplate(eventIDWithState)
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
		if pubsubEvent.State == eventstate.Error && pubsubEvent.Detail != nil {
			errs, ok := pubsubEvent.Detail.(map[string]any)
			if !ok {
				glog.Errorf("Bindings.Events error: %s", "pubsubEvent.Detail is not a map[string]any")
				continue
			}
			pubsubEvent.Detail = map[string]any{"fir": newRouteDOMContext(ctx, errs)}

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

	v, ok := ctx.route.cache.Get(*pubsubEvent.SessionID)
	if !ok {
		return events
	}
	prevErrors, ok := v.(map[string]string)
	if !ok {
		return events
	}
	// unset previously set errors
	for k, v := range prevErrors {
		k := k
		v := v
		eventType := &k
		target := v
		events = append(events, dom.Event{
			Type:   eventType,
			Target: &target,
		})
	}
	newErrors := make(map[string]string)
	// track new errors
	for _, event := range events {
		if event.State == eventstate.OK {
			continue
		}
		newErrors[*event.Type] = *event.Target
	}

	ctx.route.cache.Set(*pubsubEvent.SessionID, newErrors, cache.DefaultExpiration)

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
