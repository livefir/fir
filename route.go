package fir

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
	"github.com/google/uuid"
	"github.com/livefir/fir/internal/dom"
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
type OnEventFunc func(ctx RouteContext) error

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
		opt.onEvents[strings.ToLower(name)] = onEventFunc
	}
}

type routeData map[string]any

func (r *routeData) Error() string {
	b, _ := json.Marshal(r)
	return string(b)
}

type routeRenderer func(data routeData) error
type patchRenderer func(patch ...dom.Patch) error
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
	cntrl            *controller
	template         *template.Template
	allTemplates     []string
	eventTemplateMap map[string]string
	routeOpt
}

func newRoute(cntrl *controller, routeOpt *routeOpt) *route {
	routeOpt.opt = cntrl.opt
	rt := &route{
		routeOpt:         *routeOpt,
		cntrl:            cntrl,
		eventTemplateMap: make(map[string]string),
	}
	rt.parseTemplate()
	return rt
}

func renderRoute(ctx RouteContext) routeRenderer {
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

func publishPatch(ctx context.Context, eventCtx RouteContext) patchRenderer {
	return func(patchset ...dom.Patch) error {
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

func renderPatch(ctx RouteContext) patchRenderer {
	return func(patchset ...dom.Patch) error {
		ctx.route.parseTemplate()
		channel := ctx.route.channelFunc(ctx.request, ctx.route.id)
		err := ctx.route.pubsub.Publish(ctx.request.Context(), *channel, patchset...)
		if err != nil {
			glog.Errorf("[onPatchEvent] error publishing patch: %v\n", err)
			return err
		}
		ctx.response.Write(dom.MarshalPatchset(ctx.route.template, patchset))
		return nil
	}
}

func (rt *route) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sessionID := rt.session.GetString(r.Context(), "id")
	if sessionID == "" {
		sessionID = uuid.New().String()
		rt.session.Put(r.Context(), "id", sessionID)
	}
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

		eventCtx := RouteContext{
			event:      event,
			request:    r,
			response:   w,
			route:      rt,
			domPatcher: dom.NewPatcher(),
		}

		onEventFunc, ok := rt.onEvents[strings.ToLower(event.ID)]
		if !ok {
			http.Error(w, "event id is not registered", http.StatusBadRequest)
			return
		}

		handleOnEventResult(onEventFunc(eventCtx), eventCtx, renderPatch(eventCtx))

	} else {
		// postForm
		if r.Method == http.MethodPost {
			formAction := ""
			values := r.URL.Query()
			if len(values) == 1 {
				event := values.Get("event")
				if event != "" {
					formAction = event
				}
			}
			if formAction == "" && len(rt.onEvents) > 1 {
				http.Error(w, "form action[?event=myaction] is missing and default onEvent can't be selected since there is more than 1", http.StatusBadRequest)
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
				FormID: &formAction,
			}

			eventCtx := RouteContext{
				event:      event,
				request:    r,
				response:   w,
				route:      rt,
				urlValues:  urlValues,
				domPatcher: dom.NewPatcher(),
				session:    rt.session,
			}

			onEventFunc, ok := rt.onEvents[event.ID]
			if !ok {
				http.Error(w, fmt.Sprintf("onEvent handler for %s not found", event.ID), http.StatusBadRequest)
				return
			}

			handlePostFormResult(onEventFunc(eventCtx), eventCtx)

		} else if r.Method == http.MethodGet {
			// onLoad
			event := Event{ID: rt.routeOpt.id}
			eventCtx := RouteContext{
				event:      event,
				request:    r,
				response:   w,
				route:      rt,
				isOnLoad:   true,
				domPatcher: dom.NewPatcher(),
				session:    rt.session,
			}
			handleOnLoadResult(rt.onLoad(eventCtx), nil, eventCtx)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func renderErrorBlock(ctx RouteContext, eventErrorID string, errs map[string]any) dom.Patchset {
	sourceID := ""
	if ctx.event.SourceID != nil {
		sourceID = *ctx.event.SourceID
	}
	return ctx.dom().DispatchEvent(
		eventErrorID,
		sourceID,
		ctx.renderBlock(ctx.route.eventTemplateMap[eventErrorID],
			map[string]any{"fir": newRouteDOMContext(ctx, map[string]any{ctx.event.ID: errs})})).Patchset()
}

func handleOnEventResult(err error, ctx RouteContext, render patchRenderer) {
	if err == nil {
		sourceID := ""
		if ctx.event.SourceID != nil {
			sourceID = *ctx.event.SourceID
		}
		patchsetData := ctx.dom().DispatchEvent(ctx.event.ID, sourceID, nil).Patchset()
		// check if error was previously set for this event
		eventErrorID := fmt.Sprintf("%s:error", ctx.event.ID)
		// if error was previously set, then remove it and dispatch event to unset error
		if ctx.session.GetInt(ctx.request.Context(), eventErrorID) == 1 {
			ctx.session.Remove(ctx.request.Context(), eventErrorID)
			patchsetData = renderErrorBlock(ctx, eventErrorID, nil)
		}

		render(patchsetData...)
		return
	}

	switch errVal := err.(type) {
	case *fieldErrors:
		fieldErrorsData := *errVal
		sourceID := ""
		if ctx.event.SourceID != nil {
			sourceID = *ctx.event.SourceID
		}
		patchsetData := ctx.dom().DispatchEvent(ctx.event.ID, sourceID, nil).Patchset()

		for field, err := range fieldErrorsData {
			eventErrorID := fmt.Sprintf("%s:%s:error", ctx.event.ID, field)
			// mark error as set in session
			ctx.session.Put(ctx.request.Context(), eventErrorID, 1)
			errs := map[string]any{ctx.event.ID: map[string]any{field: err.Error()}}
			patchsetData = renderErrorBlock(ctx, eventErrorID, errs)
		}

		// check if field errors were previously set for this event
		for errorEventID := range ctx.route.eventTemplateMap {
			// skip if not error event
			if !strings.HasSuffix(errorEventID, ":error") {
				continue
			}
			// skip if error is set for this field above
			_, ok := fieldErrorsData[errorEventID]
			if ok {
				continue
			}
			// remove error from session
			ctx.session.Remove(ctx.request.Context(), errorEventID)
			// unset error if previously set
			patchsetData = renderErrorBlock(ctx, errorEventID, nil)
		}

		render(patchsetData...)
		return
	case *routeData:
		data := *errVal
		eventSuccessID := fmt.Sprintf("%s:success", ctx.event.ID)

		block, ok := ctx.route.eventTemplateMap[eventSuccessID]
		if !ok {
			block, ok = ctx.route.eventTemplateMap[ctx.event.ID]
			if !ok {
				glog.Errorf("no block found to render for event %s", ctx.event.ID)
				return
			}

		}
		sourceID := ""
		if ctx.event.SourceID != nil {
			sourceID = *ctx.event.SourceID
		}
		patchsetData := ctx.dom().DispatchEvent(ctx.event.ID, sourceID,
			ctx.renderBlock(block, data)).Patchset()

		if ctx.event.FormID != nil && *ctx.event.FormID != "" {
			patchsetData = ctx.dom().ResetForm(fmt.Sprintf("#%s", *ctx.event.FormID)).Patchset()
		}

		render(patchsetData...)
		return
	default:
		var patchsetData dom.Patchset
		errs := map[string]any{ctx.event.ID: userError(ctx, err).Error()}
		eventErrorID := fmt.Sprintf("%s:error", ctx.event.ID)
		// mark error as set in session
		ctx.session.Put(ctx.request.Context(), eventErrorID, 1)
		patchsetData = renderErrorBlock(ctx, eventErrorID, errs)
		render(patchsetData...)
		return
	}
}

func handlePostFormResult(err error, ctx RouteContext) {
	if err == nil {
		http.Redirect(ctx.response, ctx.request, ctx.request.URL.Path, http.StatusFound)
		return
	}

	switch err.(type) {
	case *routeData:
		handleOnLoadResult(ctx.route.onLoad(ctx), nil, ctx)
		http.Redirect(ctx.response, ctx.request, ctx.request.URL.Path, http.StatusFound)
	case dom.Patcher:
		// ignore patchset since this is a full page render
		http.Redirect(ctx.response, ctx.request, ctx.request.URL.Path, http.StatusFound)
	default:
		handleOnLoadResult(ctx.route.onLoad(ctx), err, ctx)
	}
}

func handleOnLoadResult(err, onFormErr error, ctx RouteContext) {
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
					"route":      fmt.Sprintf("%v", fieldErrorsVal),
				}
			}
		}

		renderRoute(ctx)(routeData{"fir": newRouteDOMContext(ctx, errs)})
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
					"route":      fmt.Sprintf("%v", fieldErrorsVal),
				}
			}
		}
		onLoadData["fir"] = newRouteDOMContext(ctx, errs)
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
					"route":      fmt.Sprintf("%v", fieldErrorsVal),
				}
			}
		}

		renderRoute(ctx)(routeData{"fir": newRouteDOMContext(ctx, errs)})
	case dom.Patcher:
		glog.Errorf(`[warning] onLoad returned a dom.Patchset and was ignored for route: %+v,
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
					"route":      fmt.Sprintf("%v", fieldErrorsVal),
				}
			}
		}
		renderRoute(ctx)(routeData{"fir": newRouteDOMContext(ctx, errs)})
	default:
		var errs map[string]any
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
					"route":      fmt.Sprintf("%v", fieldErrorsVal),
				}
			}
		} else {
			errs = map[string]any{
				ctx.event.ID: err.Error(),
				"route":      err.Error()}
		}
		renderRoute(ctx)(routeData{"fir": newRouteDOMContext(ctx, errs)})
	}

}

func (rt *route) parseTemplate() {
	var err error
	if rt.template == nil || (rt.template != nil && rt.disableTemplateCache) {
		rt.template, err = parseTemplate(rt.routeOpt)
		if err != nil {
			panic(err)
		}
		rt.findAllTemplates()
		rt.buildEventRenderMapping()
	}
}

func (rt *route) findAllTemplates() {
	rt.allTemplates = []string{}
	for _, t := range rt.template.Templates() {
		tName := t.Name()
		rt.allTemplates = append(rt.allTemplates, tName)

	}
}

func (rt *route) buildEventRenderMapping() {
	opt := rt.routeOpt
	if opt.layout == "" && opt.content == "" {
		return
	}

	walkFile := func(page string) {
		pagePath := filepath.Join(opt.publicDir, page)
		// is layout html content or a file/directory
		if isFileOrString(pagePath, opt) {
			parseEventRenderMapping(rt.allTemplates, rt.eventTemplateMap, strings.NewReader(page))
		} else {
			// compile layout
			commonFiles := []string{pagePath}
			// global partials
			for _, partial := range opt.partials {
				commonFiles = append(commonFiles, find(opt, filepath.Join(opt.publicDir, partial), opt.extensions)...)
			}
			if opt.hasEmbedFS {
				for _, v := range commonFiles {
					r, err := opt.embedFS.Open(v)
					if err != nil {
						panic(err)
					}
					parseEventRenderMapping(rt.allTemplates, rt.eventTemplateMap, r)
				}
			} else {
				for _, v := range commonFiles {
					r, err := os.OpenFile(v, os.O_RDONLY, 0644)
					if err != nil {
						panic(err)
					}
					parseEventRenderMapping(rt.allTemplates, rt.eventTemplateMap, r)
				}

			}

		}
	}

	if opt.layout != "" {
		walkFile(opt.layout)
	}

	if opt.content != "" {
		walkFile(opt.content)
	}

}
