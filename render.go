package fir

import (
	"fmt"
	"html/template"

	"github.com/livefir/fir/internal/dom"
	"github.com/livefir/fir/internal/eventstate"
	"github.com/livefir/fir/internal/firattr"
	"github.com/livefir/fir/internal/logger"
	"github.com/livefir/fir/internal/renderer"
	"github.com/livefir/fir/pubsub"
	"github.com/patrickmn/go-cache"
	"github.com/sourcegraph/conc/pool"
	"github.com/valyala/bytebufferpool"
)

func renderRoute(ctx RouteContext, errorRouteTemplate bool) routeRenderer {
	return func(data routeData) error {
		// Get route interface and parse templates
		route := ctx.routeInterface.(*route)
		route.parseTemplatesWithEngine()
		buf := bytebufferpool.Get()
		defer bytebufferpool.Put(buf)

		var tmpl *template.Template
		if errorRouteTemplate {
			tmpl = route.getErrorTemplate()
		} else {
			tmpl = route.getTemplate()
		}
		var errs map[string]any
		errMap, ok := data["errors"]
		if ok {
			errs = errMap.(map[string]any)
		}

		tmpl = tmpl.Funcs(newFirFuncMap(ctx, errs))
		err := tmpl.Execute(buf, data)
		if err != nil {
			logger.Errorf("error executing template: %v", err)
			return err
		}

		err = encodeSession(route.routeOpt, ctx.response, ctx.request)
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
func renderDOMEvents(ctx RouteContext, pubsubEvent pubsub.Event) []dom.Event {
	eventIDWithState := fmt.Sprintf("%s:%s", *pubsubEvent.ID, pubsubEvent.State)
	var templateNames []string

	// Get event templates from RouteInterface
	eventTemplatesIface := ctx.routeInterface.GetEventTemplates()
	if eventTemplatesIface == nil {
		return []dom.Event{}
	}

	// Type assert to the named type eventTemplates
	if eventTemplatesMap, ok := eventTemplatesIface.(eventTemplates); ok {
		for k := range eventTemplatesMap[eventIDWithState] {
			templateNames = append(templateNames, k)
		}
	} else {
		logger.Errorf("failed to type assert event templates, got type: %T", eventTemplatesIface)
		return []dom.Event{}
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
		if !isEmptyEvent(event) {
			events = append(events, event)
		}
	}
	return events
}

// targetOrClassName returns the target or className based on the parameters
func targetOrClassName(target *string, className string) *string {
	return renderer.TargetOrClassName(target, className)
}

func buildDOMEventFromTemplate(ctx RouteContext, pubsubEvent pubsub.Event, eventIDWithState, templateName string) *dom.Event {
	if templateName == "-" {
		eventType := fir(eventIDWithState)
		detail := &dom.Detail{}
		if pubsubEvent.Detail != nil {
			detail.State = pubsubEvent.Detail.State
		}
		return &dom.Event{
			ID:     *pubsubEvent.ID,
			State:  pubsubEvent.State,
			Type:   eventType,
			Key:    pubsubEvent.ElementKey,
			Target: targetOrClassName(pubsubEvent.Target, firattr.GetClassName(*eventType)),
			Detail: detail,
		}
	}
	eventType := fir(eventIDWithState, templateName)
	var templateData any
	if pubsubEvent.Detail != nil {
		templateData = pubsubEvent.Detail.Data
	}

	// Get template from RouteInterface
	templateIface := ctx.routeInterface.GetTemplate()
	if templateIface == nil {
		logger.Errorf("template not found for route: %s", ctx.routeInterface.ID())
		return nil
	}

	// Type assert to *template.Template
	routeTemplate, ok := templateIface.(*template.Template)
	if !ok {
		logger.Errorf("template is not of type *template.Template for route: %s", ctx.routeInterface.ID())
		return nil
	}

	routeTemplate = routeTemplate.Funcs(newFirFuncMap(ctx, nil))
	if pubsubEvent.State == eventstate.Error && pubsubEvent.Detail != nil {
		errs, ok := pubsubEvent.Detail.Data.(map[string]any)
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

	if pubsubEvent.State == eventstate.OK && templateData == nil {
		value = ""
	}

	detail := &dom.Detail{}
	if pubsubEvent.Detail != nil {
		detail.State = pubsubEvent.Detail.State
		detail.Data = pubsubEvent.Detail.Data
	}
	detail.HTML = value
	return &dom.Event{
		ID:     *pubsubEvent.ID,
		State:  pubsubEvent.State,
		Type:   eventType,
		Key:    pubsubEvent.ElementKey,
		Target: targetOrClassName(pubsubEvent.Target, firattr.GetClassName(*eventType)),
		Detail: detail,
	}
}

func getUnsetErrorEvents(cch *cache.Cache, sessionID *string, events []dom.Event) []dom.Event {
	return renderer.GetUnsetErrorEvents(cch, sessionID, events)
}

func buildTemplateValue(t *template.Template, templateName string, data any) (string, error) {
	return renderer.BuildTemplateValue(t, templateName, data, addAttributes)
}

func isEmptyEvent(event dom.Event) bool {
	return renderer.IsEmptyEvent(event)
}
