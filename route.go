package fir

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/adnaan/fir/patch"
	"github.com/golang/glog"
	"golang.org/x/exp/slices"
)

var firErrorPrefix = "fir-error-"

// RouteOption is a function that sets route options
type RouteOption func(*routeOpt)

// RouteOptions is a slice of RouteOption
type RouteOptions []RouteOption

// RouteFunc is a function that handles a route
type RouteFunc func() RouteOptions

// Route is an interface that represents a route
type Route interface{ Options() RouteOptions }

// OnEventFunc is a function that handles an http event request
type OnEventFunc func(ctx Context) error

// ID  sets the route unique identifier. This is used to identify the route in pubsub.
func ID(id string) RouteOption {
	return func(opt *routeOpt) {
		opt.id = id
	}
}

// Layout sets the layout for the route's template engine
func Layout(layout string) RouteOption {
	return func(opt *routeOpt) {
		opt.layout = layout
	}
}

// Content sets the content for the route
func Content(content string) RouteOption {
	return func(opt *routeOpt) {
		opt.content = content
	}
}

// LayoutContentName sets the name of the template which contains the content.
/*
 {{define "layout"}}
 {{ define "content" }}
 {{ end }}
 {{end}}

 Here "content" is the default layout content name
*/
func LayoutContentName(name string) RouteOption {
	return func(opt *routeOpt) {
		opt.layoutContentName = name
	}
}

// Partials sets the template partials for the route's template engine
func Partials(partials ...string) RouteOption {
	return func(opt *routeOpt) {
		opt.partials = partials
	}
}

// Extensions sets the template file extensions read for the route's template engine
func Extensions(extensions ...string) RouteOption {
	return func(opt *routeOpt) {
		opt.extensions = extensions
	}
}

// FuncMap appends to the default template function map for the route's template engine
func FuncMap(funcMap template.FuncMap) RouteOption {
	return func(opt *routeOpt) {
		mergedFuncMap := make(template.FuncMap)
		for k, v := range opt.funcMap {
			mergedFuncMap[k] = v
		}
		for k, v := range defaultFuncMap() {
			mergedFuncMap[k] = v
		}
		opt.funcMap = mergedFuncMap
	}
}

// EventSender sets the event sender for the route. It can be used to send events for the route
// without a corresponding user event. This is useful for sending events to the route event handler for use cases like:
// sending notifications, sending emails, etc.
func EventSender(eventSender chan Event) RouteOption {
	return func(opt *routeOpt) {
		opt.eventSender = eventSender
	}
}

// OnLoad sets the route's onload event handler
func OnLoad(f OnEventFunc) RouteOption {
	return func(opt *routeOpt) {
		opt.onLoad = f
	}
}

// OnEvent registers an event handler for the route per unique event name. It can be called multiple times
// to register multiple event handlers for the route.
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
type patchRenderer func(patch ...patch.Op) error
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
			glog.Errorf("[renderRoute] error executing template: %v\n", err)
			return err
		}

		ctx.response.Write(buf.Bytes())
		return nil
	}
}

func publishPatch(ctx context.Context, eventCtx Context) patchRenderer {
	return func(patchset ...patch.Op) error {
		channel := eventCtx.route.channelFunc(eventCtx.request, eventCtx.route.id)
		eventCtx.route.parseTemplate()
		err := eventCtx.route.pubsub.Publish(ctx, *channel, patchset...)
		if err != nil {
			glog.Errorf("[onWebsocket][getEventPatchset] error publishing patch: %v\n", err)
			return err
		}
		return nil
	}
}

func renderPatch(ctx Context) patchRenderer {
	return func(patchset ...patch.Op) error {
		ctx.route.parseTemplate()
		channel := ctx.route.channelFunc(ctx.request, ctx.route.id)
		err := ctx.route.pubsub.Publish(ctx.request.Context(), *channel, patchset...)
		if err != nil {
			glog.Errorf("[onPatchEvent] error publishing patch: %v\n", err)
			return err
		}
		ctx.response.Write(patch.RenderJSON(ctx.route.template, patchset))
		return nil
	}
}

func (rt *route) handle(w http.ResponseWriter, r *http.Request) {
	rt.parseTemplate()
	if r.Header.Get("Connection") == "Upgrade" &&
		r.Header.Get("Upgrade") == "websocket" {
		// onWebsocket: upgrade to websocket
		onWebsocket(w, r, rt)
	} else if r.Header.Get("X-FIR-MODE") == "event" && r.Method == http.MethodPost {
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
		if r.Method == http.MethodPost {
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

		} else if r.Method == http.MethodGet {
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
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func handleOnEventResult(err error, ctx Context, render patchRenderer) {
	unsetErrors := map[string]any{}
	for _, v := range ctx.route.firErrorTemplates {
		unsetErrors[v] = struct{}{}
	}

	if err == nil {
		var patchsetData []patch.Op
		for k := range unsetErrors {
			errs := map[string]any{ctx.event.ID: nil}
			patchsetData = append(patchsetData,
				patch.Morph(fmt.Sprintf("#%s", k),
					patch.Block(k, map[string]any{"fir": newRouteContext(ctx, errs)})))
		}

		render(patchsetData...)
		return
	}

	switch errVal := err.(type) {
	case *routeData:
		render(patch.Store("fir", *errVal))
		return
	case *patch.Set:
		patchsetData := *errVal
		if ctx.event.IsForm {
			patchsetData = append(patchsetData,
				patch.ResetForm(fmt.Sprintf("#%s", ctx.event.ID)))
		}
		for k := range unsetErrors {
			errs := map[string]any{ctx.event.ID: nil}
			patchsetData = append(patchsetData,
				patch.Morph(fmt.Sprintf("#%s", k),
					patch.Block(k, map[string]any{"fir": newRouteContext(ctx, errs)})))
		}
		render(patchsetData...)
		return
	case *fieldErrors:
		fieldErrorsData := *errVal
		var patchsetData []patch.Op

		for k, v := range fieldErrorsData {
			fieldErrorName := fmt.Sprintf("%s%s-%s", firErrorPrefix, ctx.event.ID, k)
			// eror is set, don't unset it
			delete(unsetErrors, fieldErrorName)
			errs := map[string]any{
				ctx.event.ID: map[string]any{
					k: v.Error()},
				"route": v.Error(),
			}
			patchsetData = append(patchsetData,
				patch.Morph(fmt.Sprintf("#%s", fieldErrorName),
					patch.Block(fieldErrorName, map[string]any{"fir": newRouteContext(ctx, errs)})))
		}
		// unset errors that are not set
		for k := range unsetErrors {
			errs := map[string]any{ctx.event.ID: nil}
			patchsetData = append(patchsetData,
				patch.Morph(fmt.Sprintf("#%s", k),
					patch.Block(k, map[string]any{"fir": newRouteContext(ctx, errs)})))
		}

		render(patchsetData...)
		return
	default:
		var patchsetData []patch.Op
		userErr := userError(ctx, err)
		errs := map[string]any{
			ctx.event.ID: userErr.Error(),
			"route":      userErr.Error()}

		eventIdName := fmt.Sprintf("%s%s", firErrorPrefix, ctx.event.ID)
		eventNameSelector := fmt.Sprintf("#%s", eventIdName)
		if slices.Contains(ctx.route.firErrorTemplates, eventIdName) {
			patchsetData = append(patchsetData,
				patch.Morph(eventNameSelector,
					patch.Block(eventIdName, map[string]any{"fir": newRouteContext(ctx, errs)})))
		}

		routeName := fmt.Sprintf("%s-route", firErrorPrefix)
		routeNameSelector := fmt.Sprintf("#%s", routeName)
		if slices.Contains(ctx.route.firErrorTemplates, routeName) {
			patchsetData = append(patchsetData,
				patch.Morph(routeNameSelector,
					patch.Block(routeName, map[string]any{"fir": newRouteContext(ctx, errs)})))
		}

		render(patchsetData...)
		return
	}
}

func handleOnFormResult(err error, ctx Context) {
	if err == nil {
		http.Redirect(ctx.response, ctx.request, ctx.request.URL.Path, http.StatusFound)
		return
	}

	switch errVal := err.(type) {
	case *routeData:
		onFormData := *errVal
		onFormData["fir"] = newRouteContext(ctx, map[string]any{})
		renderRoute(ctx)(onFormData)
	case *patch.Set:
		// ignore patchset
		http.Redirect(ctx.response, ctx.request, ctx.request.URL.Path, http.StatusFound)
	default:
		handleOnLoadResult(ctx.route.onLoad(ctx), err, ctx)
	}
}

func handleOnLoadResult(err, onFormErr error, ctx Context) {
	if err == nil {
		errs := make(map[string]any)
		if onFormErr != nil {
			fieldErrorsVal, ok := onFormErr.(*fieldErrors)
			if !ok {
				errs = map[string]any{
					ctx.event.ID: onFormErr.Error(),
					"route":      onFormErr.Error()}
			} else {
				errs = map[string]any{
					ctx.event.ID: fieldErrorsVal.toMap(),
					//"route":      fmt.Sprintf("%v", fieldErrorsVal),
				}
			}
		}

		renderRoute(ctx)(routeData{"fir": newRouteContext(ctx, errs)})
		return
	}

	switch errVal := err.(type) {
	case *routeData:
		onLoadData := *errVal
		errs := make(map[string]any)
		if onFormErr != nil {
			fieldErrorsVal, ok := onFormErr.(*fieldErrors)
			if !ok {
				errs = map[string]any{
					ctx.event.ID: onFormErr.Error(),
					"route":      onFormErr.Error()}
			} else {
				errs = map[string]any{
					ctx.event.ID: fieldErrorsVal.toMap(),
					//"route":      fmt.Sprintf("%v", fieldErrorsVal),
				}
			}
		}
		onLoadData["fir"] = newRouteContext(ctx, errs)
		renderRoute(ctx)(onLoadData)
	case fieldErrors:
		errs := make(map[string]any)
		if onFormErr != nil {
			fieldErrorsVal, ok := onFormErr.(*fieldErrors)
			if !ok {
				errs = map[string]any{
					ctx.event.ID: onFormErr.Error(),
					"route":      onFormErr.Error()}
			} else {
				errs = map[string]any{
					ctx.event.ID: fieldErrorsVal.toMap(),
					//"route":      fmt.Sprintf("%v", fieldErrorsVal),
				}
			}
		}

		renderRoute(ctx)(routeData{"fir": newRouteContext(ctx, errs)})
	case *patch.Set:
		glog.Errorf(`[warning] onLoad returned a []patch.Patch and was ignored for route: %+v,
		 onLoad must return either an error or call ctx.Data, ctx.KV \n`, ctx.route)
		errs := make(map[string]any)
		if onFormErr != nil {
			fieldErrorsVal, ok := onFormErr.(*fieldErrors)
			if !ok {
				errs = map[string]any{
					ctx.event.ID: onFormErr.Error(),
					"route":      onFormErr.Error()}
			} else {
				errs = map[string]any{
					ctx.event.ID: fieldErrorsVal.toMap(),
					//"route":      fmt.Sprintf("%v", fieldErrorsVal),
				}
			}
		}
		renderRoute(ctx)(routeData{"fir": newRouteContext(ctx, errs)})
	default:
		errs := make(map[string]any)
		if onFormErr != nil {
			fieldErrorsVal, ok := onFormErr.(*fieldErrors)
			if !ok {
				// err is not nil and not routeData and onFormErr is not nil and not fieldErrors
				// merge err and onFormErr
				mergedErr := fmt.Errorf("%v %v", err, onFormErr)
				errs = map[string]any{
					ctx.event.ID: mergedErr,
					"route":      mergedErr,
				}
			} else {
				errs = map[string]any{
					ctx.event.ID: fieldErrorsVal.toMap(),
					//"route":      fmt.Sprintf("%v", fieldErrorsVal),
				}
			}
		} else {
			errs = map[string]any{
				ctx.event.ID: err.Error(),
				"route":      err.Error()}
		}
		renderRoute(ctx)(routeData{"fir": newRouteContext(ctx, errs)})
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
				if strings.Contains(t1.Name(), firErrorPrefix) {
					rt.firErrorTemplates = append(rt.firErrorTemplates, t1.Name())
				}
			}
		}

	}
}
