package fir

import (
	"fmt"
	"html/template"
	"net/http"

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

		// encodedRouteID, err := ctx.route.cntrl.secureCookie.Encode(ctx.route.cookieName, ctx.route.id)
		// if err != nil {
		// 	logger.Errorf("error encoding cookie: %v", err)
		// 	return err
		// }

		http.SetCookie(ctx.response, &http.Cookie{
			Name:   ctx.route.cookieName,
			Value:  ctx.route.id,
			MaxAge: 0,
			Path:   "/",
		})

		ctx.response.Write(addAttributes(buf.Bytes()))
		return nil
	}
}

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

// trackErrors is a function that processes pubsub events and returns a list of DOM events.
// It takes a RouteContext, a pubsub.Event, and a slice of dom.Event as input parameters.
// It tracks errors by comparing the previous errors stored in the cache with the new errors received in the events.
// The function updates the cache with the new errors and returns a list of new events.
// If there are no new events, it creates a new event based on the pubsubEvent and adds it to the list.
// The function returns the list of new events.
func trackErrors(ctx RouteContext, pubsubEvent pubsub.Event, events []dom.Event) []dom.Event {
	// get previously set errors from cache
	prevErrors := make(map[string]string)
	if pubsubEvent.SessionID != nil {
		v, ok := ctx.route.cache.Get(*pubsubEvent.SessionID)
		if ok {
			prevErrors, ok = v.(map[string]string)
			if !ok {
				panic("fir: cache value is not a map[string]string")
			}
		}
	}

	// set new errors & add events to newEvents
	newErrors := make(map[string]string)
	var newEvents []dom.Event

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
