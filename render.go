package fir

import (
	"fmt"
	"html/template"

	"github.com/livefir/fir/internal/dom"
	"github.com/livefir/fir/internal/eventstate"
	"github.com/livefir/fir/internal/logger"
	"github.com/livefir/fir/pubsub"
	"github.com/patrickmn/go-cache"
	"github.com/sourcegraph/conc/pool"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
	"github.com/valyala/bytebufferpool"
)

func renderRoute(ctx RouteContext, errorRouteTemplate bool) routeRenderer {
	return func(data routeData) error {
		ctx.route.parseTemplates()
		buf := bytebufferpool.Get()
		defer bytebufferpool.Put(buf)

		tmpl := ctx.route.template
		if errorRouteTemplate {
			tmpl = ctx.route.errorTemplate
		}
		var errs map[string]any
		errMap, ok := data["errors"]
		if ok {
			errs = errMap.(map[string]any)
		}

		tmpl = tmpl.Funcs(newFirFuncMap(ctx, errs))
		tmpl.Option("missingkey=zero")
		err := tmpl.Execute(buf, data)
		if err != nil {
			logger.Errorf("error executing template: %v", err)
			return err
		}

		err = encodeSession(ctx.route.routeOpt, ctx.response, ctx.request)
		if err != nil {
			logger.Errorf("error encoding session: %v", err)
			return err
		}

		_, err = ctx.response.Write(addAttributes(buf.Bytes()))
		if err != nil {
			logger.Errorf("error writing response: %v", err)
			return err
		}
		return nil
	}
}

// renderDOMEvents renders the DOM events generated from incoming pubsub event.
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
	result := resultPool.Wait()
	// filter out empty events
	var events []dom.Event
	for _, event := range result {
		if event.Type == nil {
			continue
		}
		events = append(events, event)
	}

	unsetErrorEvents := getUnsetErrorEvents(ctx.route.cache, pubsubEvent.SessionID, events)
	events = append(events, unsetErrorEvents...)

	if len(events) == 0 {
		// if no events are generated, create a default event with the pubsub event data
		// this is useful for events that don't have a template
		eventIDWithState := fmt.Sprintf("%s:%s", *pubsubEvent.ID, pubsubEvent.State)
		eventType := fir(eventIDWithState)
		events = append(events, dom.Event{
			ID:     *pubsubEvent.ID,
			State:  pubsubEvent.State,
			Type:   eventType,
			Key:    pubsubEvent.ElementKey,
			Target: targetOrClassName(pubsubEvent.Target, getClassName(*eventType)),
			Detail: pubsubEvent.Detail,
		})
	}

	return events
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
			Detail: pubsubEvent.StateDetail,
		}
	}
	eventType := fir(eventIDWithState, templateName)
	templateData := pubsubEvent.Detail
	routeTemplate := ctx.route.template.Funcs(newFirFuncMap(ctx, nil))
	if pubsubEvent.State == eventstate.Error && pubsubEvent.Detail != nil {
		errs, ok := pubsubEvent.Detail.(map[string]any)
		if !ok {
			logger.Errorf("error: %s", "pubsubEvent.Detail is not a map[string]any")
			return nil
		}
		templateData = nil
		routeTemplate = routeTemplate.Funcs(newFirFuncMap(ctx, errs))
	}
	value, err := buildTemplateValue(routeTemplate, templateName, templateData)
	if err != nil {
		logger.Errorf("error for eventType: %v, err: %v", *eventType, err)
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

func getUnsetErrorEvents(cch *cache.Cache, sessionID *string, events []dom.Event) []dom.Event {
	if sessionID == nil || cch == nil {
		return nil
	}

	// get previously set errors from cache
	prevErrors := make(map[string]string)
	if sessionID != nil {
		v, ok := cch.Get(*sessionID)
		if ok {
			prevErrors, ok = v.(map[string]string)
			if !ok {
				panic("fir: cache value is not a map[string]string")
			}
		}
	}

	// filter new errors
	currErrors := make(map[string]string)
	for _, event := range events {
		if event.Type == nil {
			continue
		}
		if event.State != eventstate.Error {
			continue
		}
		currErrors[*event.Type] = *event.Target
	}
	// set new errors in cache
	if sessionID != nil {
		cch.Set(*sessionID, currErrors, cache.DefaultExpiration)
	}

	// explicitly unset previously set errors that are not in new errors
	// this means generating an event with empty detail
	var newErrorEvents []dom.Event
	for k, v := range prevErrors {
		k := k
		v := v
		eventType := &k
		target := v
		// if the error is not in curr errors, generate an event with empty detail
		if _, ok := currErrors[*eventType]; ok {
			continue
		}
		newErrorEvents = append(newErrorEvents, dom.Event{
			Type:   eventType,
			Target: &target,
			Detail: "",
		})
	}

	return newErrorEvents
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
