package fir

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	firErrors "github.com/livefir/fir/internal/errors"
	"github.com/livefir/fir/internal/eventstate"
	"github.com/livefir/fir/pubsub"
	servertiming "github.com/mitchellh/go-server-timing"
	"github.com/valyala/bytebufferpool"
	"k8s.io/klog/v2"
)

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

// ErrorLayout sets the layout for the route's template engine
func ErrorLayout(layout string) RouteOption {
	return func(opt *routeOpt) {
		opt.errorLayout = layout
	}
}

// ErrorContent sets the content for the route
func ErrorContent(content string) RouteOption {
	return func(opt *routeOpt) {
		opt.errorContent = content
	}
}

// ErrorLayoutContentName sets the name of the template which contains the content.
/*
 {{define "layout"}}
 {{ define "content" }}
 {{ end }}
 {{end}}

 Here "content" is the default layout content name
*/
func ErrorLayoutContentName(name string) RouteOption {
	return func(opt *routeOpt) {
		opt.errorLayoutContentName = name
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
		for k, v := range funcMap {
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

type routeRenderer func(data routeData) error
type eventPublisher func(event pubsub.Event) error

type routeTemplateConfig struct {
	Layout                 string   `json:"layout"`
	Content                string   `json:"content"`
	LayoutContentName      string   `json:"layoutContentName"`
	ErrorLayout            string   `json:"errorLayout"`
	ErrorContent           string   `json:"errorContent"`
	ErrorLayoutContentName string   `json:"errorLayoutContentName"`
	Extensions             []string `json:"extensions"`
	Partials               []string `json:"partials"`
	// file:///path/tofiles, embed:///path/tofiles, http://example.com/path/tofiles, s3://bucket/path/tofiles
	Dir string `json:"dir"`
}

type routeOpt struct {
	id                     string
	layout                 string
	errorLayout            string
	errorContent           string
	content                string
	layoutContentName      string
	errorLayoutContentName string
	partials               []string
	extensions             []string
	funcMap                template.FuncMap
	eventSender            chan Event
	onLoad                 OnEventFunc
	onEvents               map[string]OnEventFunc
	opt
}

type route struct {
	template       *template.Template
	errorTemplate  *template.Template
	eventTemplates eventTemplates

	cntrl *controller
	routeOpt
	sync.RWMutex
}

func newRoute(cntrl *controller, routeOpt *routeOpt) *route {
	routeOpt.opt = cntrl.opt
	rt := &route{
		routeOpt:       *routeOpt,
		cntrl:          cntrl,
		eventTemplates: make(eventTemplates),
	}
	rt.parseTemplates()
	return rt
}

func renderRoute(ctx RouteContext, errorRouteTemplate bool) routeRenderer {
	return func(data routeData) error {
		ctx.route.parseTemplates()
		buf := bytebufferpool.Get()
		defer bytebufferpool.Put(buf)

		tmpl := ctx.route.template
		if errorRouteTemplate {
			tmpl = ctx.route.errorTemplate
		}
		tmpl.Option("missingkey=zero")
		err := tmpl.Execute(buf, data)
		if err != nil {
			klog.Errorf("[renderRoute] error executing template: %v\n", err)
			return err
		}

		// encodedRouteID, err := ctx.route.cntrl.secureCookie.Encode(ctx.route.cookieName, ctx.route.id)
		// if err != nil {
		// 	klog.Errorf("[renderRoute] error encoding cookie: %v\n", err)
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

func publishEvents(ctx context.Context, eventCtx RouteContext) eventPublisher {
	return func(pubsubEvent pubsub.Event) error {
		channel := eventCtx.route.channelFunc(eventCtx.request, eventCtx.route.id)
		err := eventCtx.route.pubsub.Publish(ctx, *channel, pubsubEvent)
		if err != nil {
			klog.Errorf("[onWebsocket][getEventPatchset] error publishing patch: %v\n", err)
			return err
		}
		return nil
	}
}

func writeAndPublishEvents(ctx RouteContext) eventPublisher {
	return func(pubsubEvent pubsub.Event) error {
		channel := ctx.route.channelFunc(ctx.request, ctx.route.id)
		err := ctx.route.pubsub.Publish(ctx.request.Context(), *channel, pubsubEvent)
		if err != nil {
			klog.Warningf("[writeAndPublishEvents] error publishing patch: %v\n", err)
		}
		events := renderDOMEvents(ctx, pubsubEvent)

		eventsData, err := json.Marshal(events)
		if err != nil {
			klog.Errorf("[writeAndPublishEvents] error marshaling patch: %v\n", err)
			return err
		}
		ctx.response.Write(eventsData)
		return nil
	}
}

func (rt *route) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	timing := servertiming.FromContext(r.Context())
	defer timing.NewMetric("route").Start().Stop()
	if r.URL.Path == "/favicon.ico" {
		http.NotFound(w, r)
		return
	}

	if websocket.IsWebSocketUpgrade(r) {
		// onWebsocket: upgrade to websocket
		if rt.disableWebsocket {
			http.Error(w, "websocket is disabled", http.StatusForbidden)
			return
		}
	} else {
		if rt.pathParamsFunc != nil {
			r = r.WithContext(context.WithValue(r.Context(), PathParamsKey, rt.pathParamsFunc(r)))
		}
	}

	if websocket.IsWebSocketUpgrade(r) {
		onWebsocket(w, r, rt.cntrl)
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
			event:    event,
			request:  r,
			response: w,
			route:    rt,
		}

		onEventFunc, ok := rt.onEvents[strings.ToLower(event.ID)]
		if !ok {
			http.Error(w, "event id is not registered", http.StatusBadRequest)
			return
		}

		handleOnEventResult(onEventFunc(eventCtx), eventCtx, writeAndPublishEvents(eventCtx))

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
				IsForm: true,
			}

			eventCtx := RouteContext{
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

			handlePostFormResult(onEventFunc(eventCtx), eventCtx)

		} else if r.Method == http.MethodGet {
			// onLoad
			event := Event{ID: rt.routeOpt.id}
			eventCtx := RouteContext{
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

func handleOnEventResult(err error, ctx RouteContext, publish eventPublisher) {
	target := ""
	if ctx.event.Target != nil {
		target = *ctx.event.Target
	}
	if err == nil {
		publish(pubsub.Event{
			ID:         &ctx.event.ID,
			State:      eventstate.OK,
			Target:     &target,
			ElementKey: ctx.event.ElementKey,
			SessionID:  ctx.event.SessionID,
		})
		return
	}

	switch errVal := err.(type) {
	case *firErrors.Status:
		errs := map[string]any{
			ctx.event.ID: firErrors.User(errVal.Err).Error(),
			"onevent":    firErrors.User(errVal.Err).Error(),
		}
		publish(pubsub.Event{
			ID:         &ctx.event.ID,
			State:      eventstate.Error,
			Target:     &target,
			ElementKey: ctx.event.ElementKey,
			Detail:     errs,
			SessionID:  ctx.event.SessionID,
		})
		return
	case *firErrors.Fields:
		fieldErrorsData := *errVal
		fieldErrors := make(map[string]any)
		for field, err := range fieldErrorsData {
			fieldErrors[field] = err.Error()
		}
		errs := map[string]any{ctx.event.ID: fieldErrors}
		publish(pubsub.Event{
			ID:         &ctx.event.ID,
			State:      eventstate.Error,
			Target:     &target,
			ElementKey: ctx.event.ElementKey,
			Detail:     errs,
			SessionID:  ctx.event.SessionID,
		})
		return
	case *routeData:
		publish(pubsub.Event{
			ID:         &ctx.event.ID,
			State:      eventstate.OK,
			Target:     &target,
			ElementKey: ctx.event.ElementKey,
			Detail:     *errVal,
			SessionID:  ctx.event.SessionID,
		})
		return

	case *routeDataWithState:
		publish(pubsub.Event{
			ID:          &ctx.event.ID,
			State:       eventstate.OK,
			Target:      &target,
			ElementKey:  ctx.event.ElementKey,
			Detail:      *errVal.routeData,
			StateDetail: *errVal.stateData,
			SessionID:   ctx.event.SessionID,
		})
		return
	case *stateData:
		publish(pubsub.Event{
			ID:          &ctx.event.ID,
			State:       eventstate.OK,
			Target:      &target,
			ElementKey:  ctx.event.ElementKey,
			StateDetail: *errVal,
			SessionID:   ctx.event.SessionID,
		})
		return
	default:
		errs := map[string]any{
			ctx.event.ID: firErrors.User(err).Error(),
			"onevent":    firErrors.User(err).Error(),
		}
		publish(pubsub.Event{
			ID:         &ctx.event.ID,
			State:      eventstate.Error,
			Target:     &target,
			ElementKey: ctx.event.ElementKey,
			Detail:     errs,
			SessionID:  ctx.event.SessionID,
		})
		return
	}
}

func handlePostFormResult(err error, ctx RouteContext) {
	if err == nil {
		http.Redirect(ctx.response, ctx.request, ctx.request.URL.Path, http.StatusFound)
		return
	}

	switch err.(type) {
	case *routeData, *stateData, *routeDataWithState:
		http.Redirect(ctx.response, ctx.request, ctx.request.URL.Path, http.StatusFound)
	default:
		handleOnLoadResult(ctx.route.onLoad(ctx), err, ctx)
	}
}

func handleOnLoadResult(err, onFormErr error, ctx RouteContext) {
	if err == nil {
		errs := make(map[string]any)
		if onFormErr != nil {
			fieldErrorsVal, ok := onFormErr.(*firErrors.Fields)
			if !ok {
				errs = map[string]any{
					ctx.event.ID: onFormErr.Error(),
				}
			} else {
				errs = map[string]any{
					ctx.event.ID: fieldErrorsVal.Map(),
				}
			}
		}

		renderRoute(ctx, false)(routeData{"fir": newRouteDOMContext(ctx, errs)})
		return
	}

	switch errVal := err.(type) {
	case *routeData:
		onLoadData := *errVal
		errs := make(map[string]any)
		if onFormErr != nil {
			fieldErrorsVal, ok := onFormErr.(*firErrors.Fields)
			if !ok {
				errs = map[string]any{
					ctx.event.ID: onFormErr.Error(),
				}
			} else {
				errs = map[string]any{
					ctx.event.ID: fieldErrorsVal.Map(),
				}
			}
		}
		onLoadData["fir"] = newRouteDOMContext(ctx, errs)
		renderRoute(ctx, false)(onLoadData)

	case *routeDataWithState:
		onLoadData := *errVal.routeData
		errs := make(map[string]any)
		if onFormErr != nil {
			fieldErrorsVal, ok := onFormErr.(*firErrors.Fields)
			if !ok {
				errs = map[string]any{
					ctx.event.ID: onFormErr.Error(),
				}
			} else {
				errs = map[string]any{
					ctx.event.ID: fieldErrorsVal.Map(),
				}
			}
		}
		onLoadData["fir"] = newRouteDOMContext(ctx, errs)
		renderRoute(ctx, false)(onLoadData)

	case firErrors.Status:
		errs := make(map[string]any)
		if onFormErr != nil {
			fieldErrorsVal, ok := onFormErr.(*firErrors.Fields)
			if !ok {
				errs = map[string]any{
					ctx.event.ID: onFormErr.Error(),
					"onload":     fmt.Sprintf("%v", errVal.Error())}
			} else {
				errs = map[string]any{
					ctx.event.ID: fieldErrorsVal.Map(),
					"onload":     fmt.Sprintf("%v", errVal.Error()),
				}
			}
		}

		renderRoute(ctx, true)(routeData{"fir": newRouteDOMContext(ctx, errs)})
	case firErrors.Fields:
		errs := make(map[string]any)
		if onFormErr != nil {
			fieldErrorsVal, ok := onFormErr.(*firErrors.Fields)
			if !ok {
				errs = map[string]any{
					ctx.event.ID: onFormErr.Error(),
					"onload":     fmt.Sprintf("%v", errVal)}
			} else {
				errs = map[string]any{
					ctx.event.ID: fieldErrorsVal.Map(),
					"onload":     fmt.Sprintf("%v", errVal),
				}
			}
		}

		renderRoute(ctx, false)(routeData{"fir": newRouteDOMContext(ctx, errs)})
	default:
		var errs map[string]any
		if onFormErr != nil {
			fieldErrorsVal, ok := onFormErr.(*firErrors.Fields)
			if !ok {
				// err is not nil and not routeData and onFormErr is not nil and not fieldErrors
				// merge err and onFormErr

				errs = map[string]any{
					ctx.event.ID: onFormErr,
					"onload":     errVal,
				}
			} else {
				errs = map[string]any{
					ctx.event.ID: fieldErrorsVal.Map(),
					"onload":     fmt.Sprintf("%v", errVal),
				}
			}
		} else {
			errs = map[string]any{
				"onload": err.Error()}
		}
		renderRoute(ctx, false)(routeData{"fir": newRouteDOMContext(ctx, errs)})
	}

}

func (rt *route) parseTemplates() {
	var err error
	if rt.template == nil || (rt.template != nil && rt.disableTemplateCache) {
		var successEventTemplates eventTemplates
		rt.template, successEventTemplates, err = parseTemplate(rt.routeOpt)
		if err != nil {
			panic(err)
		}
		var errorEventTemplates eventTemplates
		rt.errorTemplate, errorEventTemplates, err = parseErrorTemplate(rt.routeOpt)
		if err != nil {
			panic(err)
		}

		rt.eventTemplates = deepMergeEventTemplates(errorEventTemplates, successEventTemplates)
		for eventID, templates := range rt.eventTemplates {
			var templatesStr string
			for k := range templates {
				if k == "-" {
					continue
				}
				templatesStr += k + " "
			}
			fmt.Println("eventID: ", eventID, " templates: ", templatesStr)
		}

	}
}
