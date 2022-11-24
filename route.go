package fir

import (
	"bytes"
	"context"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type M map[string]any

type OnLoadFunc func(event Event, render RouteRenderer) error
type OnEventFunc func(event Event, render PatchRenderer) error
type RouteRenderer func(data any) error
type PatchRenderer func(patch ...Patch) error

type routeOpt struct {
	id                string
	layout            string
	content           string
	layoutContentName string
	partials          []string
	extensions        []string
	funcMap           template.FuncMap
	eventSender       chan Event
	onLoad            OnLoadFunc
	onForms           map[string]OnLoadFunc
	onEvents          map[string]OnEventFunc
	opt
}

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

func OnLoad(onLoadFunc OnLoadFunc) RouteOption {
	return func(opt *routeOpt) {
		opt.onLoad = onLoadFunc
	}
}
func OnForm(name string, onForm OnLoadFunc) RouteOption {
	return func(opt *routeOpt) {
		if opt.onForms == nil {
			opt.onForms = make(map[string]OnLoadFunc)
		}
		opt.onForms[name] = onForm
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

type route struct {
	cntrl    *controller
	template *template.Template
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

func routeRenderer(w http.ResponseWriter, r *http.Request, route *route) RouteRenderer {
	return func(data any) error {
		route.parseTemplate()
		route.template.Option("missingkey=zero")
		var buf bytes.Buffer
		err := route.template.Execute(&buf, data)
		if err != nil {
			return err
		}
		if route.debugLog {
			// log.Printf("OnGet render view %+v, with data => \n %+v\n",
			// 	v.view.Content(), getJSON(route.Data))
		}

		w.Write(buf.Bytes())
		return nil
	}
}

func patchSocketRenderer(ctx context.Context, conn *websocket.Conn, channel string, route *route) PatchRenderer {
	return func(patchset ...Patch) error {
		route.parseTemplate()
		err := route.pubsub.Publish(ctx, channel, patchset...)
		if err != nil {
			log.Printf("[onWebsocket][getEventPatchset] error publishing patch: %v\n", err)
			return err
		}
		return nil
	}
}

func patchRenderer(w http.ResponseWriter, r *http.Request, route *route) PatchRenderer {
	return func(patchset ...Patch) error {
		route.parseTemplate()
		channel := *route.channelFunc(r, route.id)
		err := route.pubsub.Publish(r.Context(), channel, patchset...)
		if err != nil {
			log.Printf("[onPatchEvent] error publishing patch: %v\n", err)
			return err
		}
		w.Write(buildPatchOperations(route.template, patchset))
		return nil
	}
}

func (rt *route) handle(w http.ResponseWriter, r *http.Request) {
	rt.parseTemplate()
	if r.Header.Get("Connection") == "Upgrade" &&
		r.Header.Get("Upgrade") == "websocket" {
		// onWebsocket
		onWebsocket(w, r, rt)
	} else if r.Header.Get("X-FIR-MODE") == "event" && r.Method == "POST" {
		// onPatchEvent
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

		event.request = r
		event.response = w

		err = rt.routeOpt.onEvents[event.ID](event, patchRenderer(w, r, rt))
		if err != nil {
			log.Printf("error in OnEvent: %v,  %v", event.ID, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// onRequest
		if r.Method == "POST" {
			formAction := "default"
			values := r.URL.Query()
			if len(values) == 1 {
				for k := range values {
					formAction = k
				}
			}
			body, err := ioutil.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			event := Event{
				ID:       formAction,
				Params:   body,
				request:  r,
				response: w,
			}

			err = rt.routeOpt.onForms[event.ID](event, routeRenderer(w, r, rt))
			if err != nil {
				log.Printf("error in OnForm: %v, %v,  %v", formAction, event, err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			event := Event{ID: rt.routeOpt.id, request: r, response: w}
			err := rt.routeOpt.onLoad(event, routeRenderer(w, r, rt))
			if err != nil {
				log.Printf("error in OnGet: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}
}

func (rt *route) parseTemplate() {
	var err error
	if rt.template == nil || (rt.template != nil && rt.disableTemplateCache) {
		rt.template, err = parseTemplate(rt.routeOpt)
		if err != nil {
			panic(err)
		}
	}
}
