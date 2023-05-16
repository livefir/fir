package fir

import (
	"fmt"
	"html/template"

	"github.com/livefir/fir/internal/dom"
	"github.com/livefir/fir/internal/eventstate"
	"github.com/livefir/fir/pubsub"
	"github.com/patrickmn/go-cache"
	"github.com/sourcegraph/conc/pool"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
	"github.com/valyala/bytebufferpool"
	"k8s.io/klog/v2"
)

// renderDOMEvents renders the DOM events for the given pubsub event.
// the associated templates for the event are rendered and the dom events are returned.
func renderDOMEvents(ctx RouteContext, pubsubEvent pubsub.Event) []dom.Event {
	eventIDWithState := fmt.Sprintf("%s:%s", *pubsubEvent.ID, pubsubEvent.State)
	var templateNames []string
	for k := range ctx.route.eventTemplates[eventIDWithState] {
		templateNames = append(templateNames, k)
	}

	resultPool := pool.NewWithResults[dom.Event]()
	for _, templateName := range templateNames {
		templateName := templateName
		resultPool.Go(func() dom.Event {
			ev := buildDOMEventFromTemplate(ctx, pubsubEvent, eventIDWithState, templateName)
			if ev == nil {
				return dom.Event{}
			}
			return *ev
		})
	}
	events := resultPool.Wait()

	return trackErrors(ctx, pubsubEvent, events)
}

func targetOrClassName(target *string, className string) *string {
	if target != nil && *target != "" {
		return target
	}
	cls := fmt.Sprintf(".%s", className)
	return &cls
}

func buildDOMEventFromTemplate(ctx RouteContext, pubsubEvent pubsub.Event, eventIDWithState, templateName string) *dom.Event {
	if templateName == "-" {
		eventType := fir(eventIDWithState)
		return &dom.Event{
			ID:     *pubsubEvent.ID,
			State:  pubsubEvent.State,
			Type:   eventType,
			Key:    pubsubEvent.ElementKey,
			Target: targetOrClassName(pubsubEvent.Target, getClassName(*eventType)),
			Detail: pubsubEvent.Detail,
		}
	}
	eventType := fir(eventIDWithState, templateName)
	templateData := pubsubEvent.Detail
	if pubsubEvent.State == eventstate.Error && pubsubEvent.Detail != nil {
		errs, ok := pubsubEvent.Detail.(map[string]any)
		if !ok {
			klog.Errorf("Bindings.Events error: %s", "pubsubEvent.Detail is not a map[string]any")
			return nil
		}
		templateData = map[string]any{"fir": newRouteDOMContext(ctx, errs)}
	}
	value, err := buildTemplateValue(ctx.route.template, templateName, templateData)
	if err != nil {
		klog.Errorf("Bindings.Events buildTemplateValue error for eventType: %v, err: %v", *eventType, err)
		return nil
	}
	if pubsubEvent.State == eventstate.Error && value == "" {
		return nil
	}

	return &dom.Event{
		ID:     eventIDWithState,
		State:  pubsubEvent.State,
		Type:   eventType,
		Key:    pubsubEvent.ElementKey,
		Target: targetOrClassName(pubsubEvent.Target, getClassName(*eventType)),
		Detail: value,
	}

}

func trackErrors(ctx RouteContext, pubsubEvent pubsub.Event, events []dom.Event) []dom.Event {
	var prevErrors map[string]string
	if pubsubEvent.SessionID != nil {
		v, ok := ctx.route.cache.Get(*pubsubEvent.SessionID)
		if ok {
			prevErrors, ok = v.(map[string]string)
			if !ok {
				panic("fir: cache value is not a map[string]string")
			}
		} else {
			prevErrors = make(map[string]string)
		}
	} else {
		prevErrors = make(map[string]string)
	}

	newErrors := make(map[string]string)
	var newEvents []dom.Event
	// set new errors & add events to newEvents
	for _, event := range events {
		if event.Type == nil {
			continue
		}
		if event.State == eventstate.OK {
			newEvents = append(newEvents, event)
			continue
		}
		newErrors[*event.Type] = *event.Target
		newEvents = append(newEvents, event)
	}
	// set new errors
	if pubsubEvent.SessionID != nil {
		ctx.route.cache.Set(*pubsubEvent.SessionID, newErrors, cache.DefaultExpiration)
	}
	// unset previously set errors
	for k, v := range prevErrors {
		k := k
		v := v
		eventType := &k
		target := v
		if _, ok := newErrors[*eventType]; ok {
			continue
		}
		newEvents = append(newEvents, dom.Event{
			Type:   eventType,
			Target: &target,
			Detail: "",
		})
	}

	if len(newEvents) == 0 {
		eventIDWithState := fmt.Sprintf("%s:%s", *pubsubEvent.ID, pubsubEvent.State)
		eventType := fir(eventIDWithState)
		newEvents = append(newEvents, dom.Event{
			ID:     *pubsubEvent.ID,
			State:  pubsubEvent.State,
			Type:   eventType,
			Key:    pubsubEvent.ElementKey,
			Target: targetOrClassName(pubsubEvent.Target, getClassName(*eventType)),
			Detail: pubsubEvent.Detail,
		})
	}
	return newEvents
}

func buildTemplateValue(t *template.Template, templateName string, data any) (string, error) {
	if t == nil {
		return "", nil
	}
	if templateName == "" {
		return "", nil
	}
	dataBuf := bytebufferpool.Get()
	defer bytebufferpool.Put(dataBuf)
	if templateName == "_fir_html" {
		dataBuf.WriteString(data.(string))
	} else {
		t.Option("missingkey=zero")
		err := t.ExecuteTemplate(dataBuf, templateName, data)
		if err != nil {
			return "", err
		}
	}

	m := minify.New()
	m.Add("text/html", &html.Minifier{
		KeepDefaultAttrVals: true,
	})
	rd, err := m.Bytes("text/html", addAttributes(dataBuf.Bytes()))
	if err != nil {
		panic(err)
	}

	return string(rd), nil
}
