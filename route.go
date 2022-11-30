package fir

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

type M map[string]any
type OnEventFunc func(ctx Context) error
type RouteOptions []RouteOption
type RouteFunc func() RouteOptions
type Route interface{ Options() RouteOptions }
type RouteOption func(*routeOpt)

func ID(id string) RouteOption {
	return func(opt *routeOpt) {
		opt.id = id
	}
}

func Layout(layout string) RouteOption {
	return func(opt *routeOpt) {
		opt.layout = layout
	}
}

func Content(content string) RouteOption {
	return func(opt *routeOpt) {
		opt.content = content
	}
}

func LayoutContentName(name string) RouteOption {
	return func(opt *routeOpt) {
		opt.layoutContentName = name
	}
}

func Partials(partials ...string) RouteOption {
	return func(opt *routeOpt) {
		opt.partials = partials
	}
}

func Extensions(extensions ...string) RouteOption {
	return func(opt *routeOpt) {
		opt.extensions = extensions
	}
}

func FuncMap(funcMap template.FuncMap) RouteOption {
	return func(opt *routeOpt) {
		opt.funcMap = funcMap
	}
}

func EventSender(eventSender chan Event) RouteOption {
	return func(opt *routeOpt) {
		opt.eventSender = eventSender
	}
}

func OnLoad(f OnEventFunc) RouteOption {
	return func(opt *routeOpt) {
		opt.onLoad = f
	}
}

func OnEvent(name string, onEventFunc OnEventFunc) RouteOption {
	return func(opt *routeOpt) {
		if opt.onEvents == nil {
			opt.onEvents = make(map[string]OnEventFunc)
		}
		opt.onEvents[name] = onEventFunc
	}
}

type routeData map[string]any

func (r *routeData) Error() string {
	b, _ := json.Marshal(r)
	return string(b)
}

type routeRenderer func(data routeData) error
type patchRenderer func(patch ...Patch) error
type routeOpt struct {
	id                string
	layout            string
	content           string
	layoutContentName string
	partials          []string
	extensions        []string
	funcMap           template.FuncMap
	eventSender       chan Event
	onLoad            OnEventFunc
	onEvents          map[string]OnEventFunc
	opt
}

type route struct {
	cntrl             *controller
	template          *template.Template
	firErrorTemplates []string
	routeOpt
}

func newRoute(cntrl *controller, routeOpt *routeOpt) *route {
	routeOpt.opt = cntrl.opt
	rt := &route{
		routeOpt: *routeOpt,
		cntrl:    cntrl,
	}
	rt.parseTemplate()
	return rt
}

func renderRoute(ctx Context) routeRenderer {
	return func(data routeData) error {
		ctx.route.parseTemplate()
		ctx.route.template.Option("missingkey=zero")
		var buf bytes.Buffer
		err := ctx.route.template.Execute(&buf, data)
		if err != nil {
			log.Printf("[renderRoute] error executing template: %v\n", err)
			return err
		}
		if ctx.route.debugLog {
			// log.Printf("OnGet render view %+v, with data => \n %+v\n",
			// 	v.view.Content(), getJSON(route.Data))
		}

		ctx.response.Write(buf.Bytes())
		return nil
	}
}

func publishPatch(ctx context.Context, eventCtx Context) patchRenderer {
	return func(patchset ...Patch) error {
		channel := eventCtx.route.channelFunc(eventCtx.request, eventCtx.route.id)
		eventCtx.route.parseTemplate()
		err := eventCtx.route.pubsub.Publish(ctx, *channel, patchset...)
		if err != nil {
			log.Printf("[onWebsocket][getEventPatchset] error publishing patch: %v\n", err)
			return err
		}
		return nil
	}
}

func renderPatch(ctx Context) patchRenderer {
	return func(patchset ...Patch) error {
		ctx.route.parseTemplate()
		channel := ctx.route.channelFunc(ctx.request, ctx.route.id)
		err := ctx.route.pubsub.Publish(ctx.request.Context(), *channel, patchset...)
		if err != nil {
			log.Printf("[onPatchEvent] error publishing patch: %v\n", err)
			return err
		}
		ctx.response.Write(buildPatchOperations(ctx.route.template, patchset))
		return nil
	}
}

func (rt *route) handle(w http.ResponseWriter, r *http.Request) {
	rt.parseTemplate()
	if r.Header.Get("Connection") == "Upgrade" &&
		r.Header.Get("Upgrade") == "websocket" {
		// onWebsocket: upgrade to websocket
		onWebsocket(w, r, rt)
	} else if r.Header.Get("X-FIR-MODE") == "event" && r.Method == "POST" {
		// onEvents
		var event Event
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		err := decoder.Decode(&event)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if decoder.More() {
			http.Error(w, "unknown fields in request body", http.StatusBadRequest)
			return
		}
		if event.ID == "" {
			http.Error(w, "event id is missing", http.StatusBadRequest)
			return
		}

		eventCtx := Context{
			event:    event,
			request:  r,
			response: w,
			route:    rt,
		}

		onEventFunc, ok := rt.onEvents[event.ID]
		if !ok {
			http.Error(w, "event id is not registered", http.StatusBadRequest)
			return
		}

		handleOnEventResult(onEventFunc(eventCtx), eventCtx, renderPatch(eventCtx))

	} else {
		// onForms
		if r.Method == "POST" {
			formAction := ""
			values := r.URL.Query()
			if len(values) == 1 {
				id := values.Get("id")
				if id != "" {
					formAction = id
				}
			}
			if formAction == "" && len(rt.onEvents) > 1 {
				http.Error(w, "form action[?id=myaction] is missing and default onEvent can't be selected since there is more than 1", http.StatusBadRequest)
				return
			} else if formAction == "" && len(rt.onEvents) == 1 {
				for k := range rt.onEvents {
					formAction = k
				}
			}

			err := r.ParseForm()
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			urlValues := r.PostForm
			params, err := json.Marshal(urlValues)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			event := Event{
				ID:     formAction,
				Params: params,
				IsForm: true,
			}

			eventCtx := Context{
				event:     event,
				request:   r,
				response:  w,
				route:     rt,
				urlValues: urlValues,
			}

			onEventFunc, ok := rt.onEvents[event.ID]
			if !ok {
				http.Error(w, fmt.Sprintf("onEvent handler for %s not found", event.ID), http.StatusBadRequest)
				return
			}

			handleOnFormResult(onEventFunc(eventCtx), eventCtx)

		} else {
			// onLoad
			event := Event{ID: rt.routeOpt.id}
			eventCtx := Context{
				event:    event,
				request:  r,
				response: w,
				route:    rt,
				isOnLoad: true,
			}
			handleOnLoadResult(rt.onLoad(eventCtx), nil, eventCtx)
		}
	}
}

func getFirData(ctx Context) routeData {
	errors := M{}
	for k := range ctx.route.onEvents {
		errors[k] = M{}
	}
	return routeData{"app_name": ctx.route.appName, "errors": errors}
}

func handleOnEventResult(err error, ctx Context, render patchRenderer) {
	firData := getFirData(ctx)
	unsetErrors := M{}
	for _, v := range ctx.route.firErrorTemplates {
		unsetErrors[v] = struct{}{}
	}
	setError, unsetError := morphFirErrors(ctx.event.ID)
	if err == nil {
		var patchlistData []Patch
		patchlistData = append(patchlistData, unsetError())
		for k := range unsetErrors {
			firData["errors"] = M{ctx.event.ID: M{}}
			patchlistData = append(patchlistData,
				Morph(fmt.Sprintf("#%s", k),
					Block(k, M{"fir": firData})))
		}

		render(patchlistData...)
		return
	}

	switch errVal := err.(type) {
	case *routeData:
		render(Store("fir", *errVal))
		return
	case *patchlist:
		patchlistData := *errVal
		patchlistData = append(patchlistData, unsetError())
		for k := range unsetErrors {
			firData["errors"] = M{ctx.event.ID: M{}}
			patchlistData = append(patchlistData,
				Morph(fmt.Sprintf("#%s", k),
					Block(k, M{"fir": firData})))
		}
		render(patchlistData...)
		return
	case fieldErrors:
		fieldErrorsData := errVal
		var patchlistData []Patch

		for k, v := range fieldErrorsData {
			fieldErrorName := fmt.Sprintf("fir-errors-%s-%s", ctx.event.ID, k)
			// eror is set, don't unset it
			delete(unsetErrors, fieldErrorName)
			firData["errors"] = M{ctx.event.ID: M{k: v.Error()}}
			patchlistData = append(patchlistData,
				Morph(fmt.Sprintf("#%s", fieldErrorName),
					Block(fieldErrorName, M{"fir": firData})))
		}
		// unset errors that are not set
		for k := range unsetErrors {
			firData["errors"] = M{ctx.event.ID: M{}}
			patchlistData = append(patchlistData,
				Morph(fmt.Sprintf("#%s", k),
					Block(k, M{"fir": firData})))
		}

		render(patchlistData...)
		return
	default:
		render(setError(err))
		return
	}
}

func handleOnFormResult(err error, ctx Context) {
	if err == nil {
		handleOnLoadResult(ctx.route.onLoad(ctx), nil, ctx)
		return
	}

	switch errVal := err.(type) {
	case *routeData:
		onFormData := *errVal
		onFormData["fir"] = getFirData(ctx)
		renderRoute(ctx)(onFormData)
	case *patchlist:
		// ignore patchlist
		handleOnLoadResult(ctx.route.onLoad(ctx), nil, ctx)
	default:
		handleOnLoadResult(ctx.route.onLoad(ctx), err, ctx)
	}

}

func handleOnLoadResult(err, onFormErr error, ctx Context) {
	firData := getFirData(ctx)
	if err == nil {
		if onFormErr != nil {
			fieldErrorsVal, ok := onFormErr.(*fieldErrors)
			if !ok {
				firData["errors"] = M{ctx.event.ID: onFormErr.Error()}
			} else {
				firData["errors"] = M{ctx.event.ID: *fieldErrorsVal}
			}
		}

		renderRoute(ctx)(routeData{"fir": firData})
		return
	}

	switch errVal := err.(type) {

	case *routeData:
		onLoadData := *errVal
		if onFormErr != nil {
			fieldErrorsVal, ok := onFormErr.(*fieldErrors)
			if !ok {
				firData["errors"] = M{ctx.event.ID: onFormErr.Error()}
			} else {
				firData["errors"] = M{ctx.event.ID: *fieldErrorsVal}
			}
		}
		onLoadData["fir"] = firData
		renderRoute(ctx)(onLoadData)
	case fieldErrors:
		if onFormErr != nil {
			fieldErrorsVal, ok := onFormErr.(*fieldErrors)
			if !ok {
				errVal["onForm"] = onFormErr
			} else {
				for k, v := range *fieldErrorsVal {
					errVal[k] = v
				}
			}
		}

		firData["errors"] = M{ctx.event.ID: errVal}
		renderRoute(ctx)(routeData{"fir": firData})
	case *patchlist:
		log.Printf("[warning] onLoad returned a []Patch and was ignored for route: %+v, onLoad must return either an error or call ctx.Data, ctx.KV \n", ctx.route)
		if onFormErr != nil {
			fieldErrorsVal, ok := onFormErr.(*fieldErrors)
			if !ok {
				firData["errors"] = M{ctx.event.ID: onFormErr.Error()}
			} else {
				firData["errors"] = M{ctx.event.ID: *fieldErrorsVal}
			}
		}
		renderRoute(ctx)(routeData{"fir": firData})
	default:
		if onFormErr != nil {
			fieldErrorsVal, ok := onFormErr.(*fieldErrors)
			if !ok {
				// err is not nil and not routeData and onFormErr is not nil and not fieldErrors
				// merge err and onFormErr
				firData["errors"] = M{ctx.event.ID: fmt.Errorf("%v %v", err, onFormErr)}
			} else {
				firData["errors"] = M{ctx.event.ID: *fieldErrorsVal}
			}
		} else {
			firData["errors"] = M{ctx.event.ID: err.Error()}
		}
		renderRoute(ctx)(routeData{"fir": firData})

	}

}

func (rt *route) parseTemplate() {
	var err error
	if rt.template == nil || (rt.template != nil && rt.disableTemplateCache) {
		rt.template, err = parseTemplate(rt.routeOpt)
		if err != nil {
			panic(err)
		}
		rt.findFirErrorTemplates()
	}
}

func (rt *route) findFirErrorTemplates() {
	for _, t := range rt.template.Templates() {
		if t.Name() == rt.layoutContentName {
			for _, t1 := range t.Templates() {
				if strings.Contains(t1.Name(), "fir-errors-") {
					rt.firErrorTemplates = append(rt.firErrorTemplates, t1.Name())
				}
			}
		}

	}

}
