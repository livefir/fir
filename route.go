package fir

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/goccy/go-json"

	"github.com/gorilla/schema"
	"github.com/gorilla/websocket"
	"github.com/livefir/fir/internal/dom"
	firErrors "github.com/livefir/fir/internal/errors"
	"github.com/livefir/fir/internal/eventstate"
	"github.com/livefir/fir/internal/logger"
	"github.com/livefir/fir/internal/routeservices"
	"github.com/livefir/fir/pubsub"
	servertiming "github.com/mitchellh/go-server-timing"
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

// Layout sets the layout for the route's template engine with automatic relative path resolution.
// This function detects the caller's file location and resolves relative paths accordingly,
// ensuring the final path is correctly resolved relative to the working directory.
func Layout(layout string) RouteOption {
	resolvedPath, _ := resolveTemplatePath(layout, 2)
	return func(opt *routeOpt) {
		opt.layout = resolvedPath
	}
}

// Content sets the content for the route with automatic relative path resolution.
// This function detects the caller's file location and resolves relative paths accordingly,
// ensuring the final path is correctly resolved relative to the working directory.
// It handles both relative and absolute paths, as well as inline HTML content.
func Content(content string) RouteOption {
	resolvedPath, _ := resolveTemplatePath(content, 2)

	// Check if content looks like a file path (has extension or path separators)
	// rather than inline HTML content
	looksLikeFilePath := strings.Contains(content, "/") || strings.Contains(content, "\\") ||
		(strings.Contains(content, ".") && !strings.Contains(content, " ") &&
			!strings.Contains(content, "<") && !strings.Contains(content, "{{"))

	// If it looks like a file path, always use the resolved path
	// If it doesn't look like a file path, treat as inline content
	if looksLikeFilePath {
		return func(opt *routeOpt) {
			opt.content = resolvedPath
		}
	} else {
		return func(opt *routeOpt) {
			opt.content = content
		}
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

// ErrorLayout sets the layout for the route's template engine with automatic relative path resolution.
// This function detects the caller's file location and resolves relative paths accordingly,
// ensuring the final path is correctly resolved relative to the working directory.
func ErrorLayout(layout string) RouteOption {
	resolvedPath, _ := resolveTemplatePath(layout, 2)
	return func(opt *routeOpt) {
		opt.errorLayout = resolvedPath
	}
}

// ErrorContent sets the content for the route with automatic relative path resolution.
// This function detects the caller's file location and resolves relative paths accordingly,
// ensuring the final path is correctly resolved relative to the working directory.
func ErrorContent(content string) RouteOption {
	resolvedPath, _ := resolveTemplatePath(content, 2)

	// Check if content looks like a file path (has extension or path separators)
	// rather than inline HTML content
	looksLikeFilePath := strings.Contains(content, "/") || strings.Contains(content, "\\") ||
		(strings.Contains(content, ".") && !strings.Contains(content, " ") &&
			!strings.Contains(content, "<") && !strings.Contains(content, "{{"))

	// If it looks like a file path, always use the resolved path
	// If it doesn't look like a file path, treat as inline content
	if looksLikeFilePath {
		return func(opt *routeOpt) {
			opt.errorContent = resolvedPath
		}
	} else {
		return func(opt *routeOpt) {
			opt.errorContent = content
		}
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

// Partials sets the template partials for the route's template engine with automatic relative path resolution.
// This function detects the caller's file location and resolves relative paths accordingly,
// ensuring the final paths are correctly resolved relative to the working directory.
func Partials(partials ...string) RouteOption {
	resolvedPartials := make([]string, len(partials))
	for i, partial := range partials {
		resolvedPartials[i], _ = resolveTemplatePath(partial, 2)
	}
	return func(opt *routeOpt) {
		opt.partials = append(opt.partials, resolvedPartials...)
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
		opt.mergeFuncMap(funcMap)
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

// TemplateEngine sets a custom template engine for the route.
// This allows routes to use specialized template engines while maintaining backward compatibility.
func TemplateEngine(engine interface{}) RouteOption {
	return func(opt *routeOpt) {
		opt.templateEngine = engine
	}
}

// DisableRouteTemplateCache disables template caching for this specific route.
// This can be useful for development or routes with dynamic templates.
func DisableRouteTemplateCache(disable bool) RouteOption {
	return func(opt *routeOpt) {
		opt.disableTemplateCache = disable
	}
}

type routeRenderer func(data routeData) error
type eventPublisher func(event pubsub.Event) error

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
	funcMapMutex           *sync.RWMutex
	eventSender            chan Event
	onLoad                 OnEventFunc
	// TODO: onEvents can be removed in a future version since events are now managed by EventRegistry
	// Keeping for backward compatibility during transition
	onEvents             map[string]OnEventFunc
	templateEngine       interface{} // Template engine for this route (optional)
	disableTemplateCache bool        // Whether to disable template caching for this route
	opt
}

// add func to funcMap
func (opt *routeOpt) addFunc(key string, f any) {
	opt.funcMapMutex.Lock()
	defer opt.funcMapMutex.Unlock()

	opt.funcMap[key] = f
}

// mergeFuncMap merges a value to the funcMap in a concurrency safe way.
func (opt *routeOpt) mergeFuncMap(funcMap template.FuncMap) {
	opt.funcMapMutex.Lock()
	defer opt.funcMapMutex.Unlock()
	for k, v := range funcMap {
		opt.funcMap[k] = v
	}
}

// getFuncMap lists the funcMap in a concurrency safe way.
func (opt *routeOpt) getFuncMap() template.FuncMap {
	opt.funcMapMutex.Lock()
	defer opt.funcMapMutex.Unlock()

	return opt.funcMap
}

type route struct {
	// Template handling - keeping old fields for backward compatibility during migration
	template       *template.Template
	errorTemplate  *template.Template
	eventTemplates eventTemplates

	// New template engine integration
	templateEngine interface{} // Will be TemplateEngine interface to avoid circular imports

	renderer Renderer

	services *routeservices.RouteServices

	// Commonly accessed configuration fields
	disableTemplateCache bool
	disableWebsocket     bool
	pathParamsFunc       func(r *http.Request) map[string]string

	routeOpt
	sync.RWMutex
}

func getTemplateEngine(services *routeservices.RouteServices, routeOpt *routeOpt) interface{} {
	if routeOpt.templateEngine != nil {
		return routeOpt.templateEngine
	}
	return services.TemplateEngine
}

// getTemplateCacheDisabled returns whether template caching should be disabled for a route.
// It prioritizes route-specific cache setting over the services default.
func getTemplateCacheDisabled(services *routeservices.RouteServices, routeOpt *routeOpt) bool {
	// If explicitly set on route options, use that setting
	if routeOpt.disableTemplateCache {
		return true
	}
	// Otherwise use the services default
	return services.Options.DisableTemplateCache
}

func newRoute(services *routeservices.RouteServices, routeOpt *routeOpt) (*route, error) {
	// Use the services' renderer if specified, otherwise use the default
	var renderer Renderer
	if services.Renderer != nil {
		var ok bool
		renderer, ok = services.Renderer.(Renderer)
		if !ok {
			return nil, fmt.Errorf("services.Renderer is not a valid Renderer type")
		}
	} else {
		renderer = NewTemplateRenderer()
	}

	rt := &route{
		routeOpt:             *routeOpt,
		services:             services,
		eventTemplates:       make(eventTemplates),
		renderer:             renderer,
		templateEngine:       getTemplateEngine(services, routeOpt),        // Use route-specific engine or fallback to services
		disableTemplateCache: getTemplateCacheDisabled(services, routeOpt), // Use route-specific cache setting or fallback to services
		disableWebsocket:     services.Options.DisableWebsocket,
		pathParamsFunc:       services.PathParamsFunc,
	}

	// Register events in the services' EventRegistry
	if routeOpt.onEvents != nil {
		for eventID, handler := range routeOpt.onEvents {
			err := services.EventRegistry.Register(routeOpt.id, eventID, handler)
			if err != nil {
				return nil, fmt.Errorf("failed to register event %s for route %s: %v", eventID, routeOpt.id, err)
			}
		}
	}

	// Register onLoad handler if present
	if routeOpt.onLoad != nil {
		err := services.EventRegistry.Register(routeOpt.id, "load", routeOpt.onLoad)
		if err != nil {
			return nil, fmt.Errorf("failed to register onLoad for route %s: %v", routeOpt.id, err)
		}
	}

	err := rt.parseTemplatesWithEngine()
	if err != nil {
		return nil, err
	}
	return rt, nil
}

func publishEvents(ctx context.Context, eventCtx RouteContext, channel string) eventPublisher {
	return func(pubsubEvent pubsub.Event) error {
		err := eventCtx.route.services.PubSub.Publish(ctx, channel, pubsubEvent)
		if err != nil {
			logger.Errorf("error publishing patch: %v", err)
			return err
		}
		return nil
	}
}

func publishEventsWithServices(ctx context.Context, pubsubAdapter pubsub.Adapter, channel string) eventPublisher {
	return func(pubsubEvent pubsub.Event) error {
		err := pubsubAdapter.Publish(ctx, channel, pubsubEvent)
		if err != nil {
			logger.Errorf("error publishing patch: %v", err)
			return err
		}
		return nil
	}
}

func writeAndPublishEvents(ctx RouteContext) eventPublisher {
	return func(pubsubEvent pubsub.Event) error {
		channel := ctx.route.services.ChannelFunc(ctx.request, ctx.route.id)
		if channel == nil {
			logger.Errorf("error: channel is empty")
			http.Error(ctx.response, "channel is empty", http.StatusUnauthorized)
			return nil
		}
		err := ctx.route.services.PubSub.Publish(ctx.request.Context(), *channel, pubsubEvent)
		if err != nil {
			logger.Debugf("error publishing patch: %v", err)
		}

		return writeEventHTTP(ctx, pubsubEvent)
	}
}

func writeEventHTTP(ctx RouteContext, event pubsub.Event) error {
	events := ctx.route.renderer.RenderDOMEvents(ctx, event)
	eventsData, err := json.Marshal(events)
	if err != nil {
		logger.Errorf("error marshaling patch: %v", err)
		return err
	}
	ctx.response.Write(eventsData)
	return nil
}

// set route template concurrency safe
func (rt *route) setTemplate(t *template.Template) {
	rt.template = t
}

// get route template concurrency safe

func (rt *route) getTemplate() *template.Template {
	return rt.template
}

// set route error template concurrency safe
func (rt *route) setErrorTemplate(t *template.Template) {
	rt.errorTemplate = t
}

// get route error template concurrency safe
func (rt *route) getErrorTemplate() *template.Template {
	return rt.errorTemplate
}

// set event templates concurrency safe
func (rt *route) setEventTemplates(templates eventTemplates) {
	rt.eventTemplates = templates
}

// get event templates concurrency safe
func (rt *route) getEventTemplates() eventTemplates {
	return rt.eventTemplates
}

func (rt *route) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	timing := servertiming.FromContext(r.Context())
	defer timing.NewMetric("route").Start().Stop()

	// Handle special requests
	if !rt.handleSpecialRequests(w, r) {
		return
	}

	// Setup path parameters if needed
	r = rt.setupPathParameters(r)

	// Route to appropriate handler based on request type
	if websocket.IsWebSocketUpgrade(r) {
		rt.handleWebSocketUpgrade(w, r)
	} else if rt.isJSONEventRequest(r) {
		rt.handleJSONEvent(w, r)
	} else if r.Method == http.MethodPost {
		rt.handleFormPost(w, r)
	} else if r.Method == http.MethodGet {
		rt.handleGetRequest(w, r)
	} else {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleOnEventResult(err error, ctx RouteContext, publish eventPublisher) *pubsub.Event {
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
		return nil
	}

	switch errVal := err.(type) {
	case *firErrors.Status:
		errs := map[string]any{
			ctx.event.ID: firErrors.User(errVal.Err).Error(),
			"onevent":    firErrors.User(errVal.Err).Error(),
		}
		return &pubsub.Event{
			ID:         &ctx.event.ID,
			State:      eventstate.Error,
			Target:     &target,
			ElementKey: ctx.event.ElementKey,
			Detail:     &dom.Detail{Data: errs},
			SessionID:  ctx.event.SessionID,
		}
	case *firErrors.Fields:
		fieldErrorsData := *errVal
		fieldErrors := make(map[string]any)
		for field, err := range fieldErrorsData {
			fieldErrors[field] = err.Error()
		}
		errs := map[string]any{ctx.event.ID: fieldErrors}
		return &pubsub.Event{
			ID:         &ctx.event.ID,
			State:      eventstate.Error,
			Target:     &target,
			ElementKey: ctx.event.ElementKey,
			Detail:     &dom.Detail{Data: errs},
			SessionID:  ctx.event.SessionID,
		}
	case *routeData:
		publish(pubsub.Event{
			ID:         &ctx.event.ID,
			State:      eventstate.OK,
			Target:     &target,
			ElementKey: ctx.event.ElementKey,
			Detail:     &dom.Detail{Data: *errVal},
			SessionID:  ctx.event.SessionID,
		})
		return nil

	case *routeDataWithState:
		publish(pubsub.Event{
			ID:         &ctx.event.ID,
			State:      eventstate.OK,
			Target:     &target,
			ElementKey: ctx.event.ElementKey,
			Detail:     &dom.Detail{Data: *errVal.routeData, State: *errVal.stateData},
			SessionID:  ctx.event.SessionID,
		})
		return nil
	case *stateData:
		publish(pubsub.Event{
			ID:         &ctx.event.ID,
			State:      eventstate.OK,
			Target:     &target,
			ElementKey: ctx.event.ElementKey,
			Detail:     &dom.Detail{State: *errVal},
			SessionID:  ctx.event.SessionID,
		})
		return nil
	default:
		errs := map[string]any{
			ctx.event.ID: firErrors.User(err).Error(),
			"onevent":    firErrors.User(err).Error(),
		}

		return &pubsub.Event{
			ID:         &ctx.event.ID,
			State:      eventstate.Error,
			Target:     &target,
			ElementKey: ctx.event.ElementKey,
			Detail:     &dom.Detail{Data: errs},
			SessionID:  ctx.event.SessionID,
		}
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

		ctx.route.renderer.RenderRoute(ctx, routeData{"errors": errs}, false)
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
		onLoadData["errors"] = errs
		ctx.route.renderer.RenderRoute(ctx, onLoadData, false)

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
		onLoadData["errors"] = errs
		ctx.route.renderer.RenderRoute(ctx, onLoadData, false)

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

		ctx.route.renderer.RenderRoute(ctx, routeData{"errors": errs}, true)
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

		ctx.route.renderer.RenderRoute(ctx, routeData{"errors": errs}, false)
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
		ctx.route.renderer.RenderRoute(ctx, routeData{"errors": errs}, false)
	}

}



// resolveTemplatePath resolves template paths with automatic relative path resolution.
// This function detects the caller's file location and resolves relative paths accordingly,
// ensuring the final path is correctly resolved relative to the working directory.
// It handles both relative and absolute paths, as well as inline HTML content.
// The callerDepth parameter specifies how many levels up the call stack to look for the caller.
func resolveTemplatePath(path string, callerDepth int) (string, bool) {
	trimmedPath := strings.TrimSpace(path)

	// First, check if this is a valid file path by trying to resolve it
	var resolvedPath string
	var isValidFilePath bool

	if filepath.IsAbs(trimmedPath) {
		// Absolute path - check if file exists
		if fileExists(trimmedPath) {
			resolvedPath = trimmedPath
			isValidFilePath = true
		}
	} else {
		// Relative path - resolve relative to caller's directory
		_, callerFile, _, ok := runtime.Caller(callerDepth)
		if ok {
			callerDir := filepath.Dir(callerFile)

			// Special handling for examples directory structure
			// If the caller is in an examples subdirectory, resolve paths relative to the example root
			var baseDir string
			if strings.Contains(callerFile, "/examples/") {
				// Find the example root directory (e.g., examples/fira, examples/routing)
				parts := strings.Split(callerFile, "/examples/")
				if len(parts) >= 2 {
					examplesPart := parts[1]
					exampleDirEnd := strings.Index(examplesPart, "/")
					if exampleDirEnd > 0 {
						exampleName := examplesPart[:exampleDirEnd]
						baseDir = filepath.Join(parts[0], "examples", exampleName)
					} else {
						// Direct file in examples root
						baseDir = filepath.Join(parts[0], "examples")
					}
				}
			}

			// If we couldn't determine an example root, use caller's directory
			if baseDir == "" {
				baseDir = callerDir
			}

			absolutePath := filepath.Join(baseDir, trimmedPath)
			absolutePath = filepath.Clean(absolutePath)

			if fileExists(absolutePath) {
				// File exists, convert back to relative path from working directory if possible
				wd, err := os.Getwd()
				if err == nil {
					relativePath, err := filepath.Rel(wd, absolutePath)
					if err == nil {
						resolvedPath = filepath.ToSlash(filepath.Clean(relativePath))
					} else {
						resolvedPath = absolutePath
					}
				} else {
					resolvedPath = absolutePath
				}
				isValidFilePath = true
			}
		}
	}

	// If it's not a valid file path, use the original path (might be inline content or will be handled later)
	if !isValidFilePath {
		resolvedPath = path
	}

	return resolvedPath, isValidFilePath
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// handleSpecialRequests handles favicon and HEAD requests
// Returns false if the request should be terminated early
func (rt *route) handleSpecialRequests(w http.ResponseWriter, r *http.Request) bool {
	if r.URL.Path == "/favicon.ico" {
		http.NotFound(w, r)
		return false
	}
	if r.Method == http.MethodHead {
		w.Header().Add("X-FIR-WEBSOCKET-ENABLED", strconv.FormatBool(!rt.disableWebsocket))
		w.WriteHeader(http.StatusNoContent)
		return false
	}
	return true
}

// setupPathParameters sets up path parameters in the request context if needed
func (rt *route) setupPathParameters(r *http.Request) *http.Request {
	if !websocket.IsWebSocketUpgrade(r) && rt.pathParamsFunc != nil {
		return r.WithContext(context.WithValue(r.Context(), PathParamsKey, rt.pathParamsFunc(r)))
	}
	return r
}

// isJSONEventRequest checks if the request is a JSON event request
func (rt *route) isJSONEventRequest(r *http.Request) bool {
	return r.Header.Get("X-FIR-MODE") == "event" && r.Method == http.MethodPost
}

// handleWebSocketUpgrade handles WebSocket upgrade requests
func (rt *route) handleWebSocketUpgrade(w http.ResponseWriter, r *http.Request) {
	if rt.disableWebsocket {
		http.Error(w, "websocket is disabled", http.StatusForbidden)
		return
	}

	// Use WebSocketServices directly from RouteServices
	if rt.services.HasWebSocketServices() {
		wsServices := rt.services.GetWebSocketServices()
		onWebsocket(w, r, wsServices)
	} else {
		logger.Errorf("ERROR: WebSocketServices not configured for route")
		http.Error(w, "WebSocket services not available", http.StatusInternalServerError)
	}
}

// handleJSONEvent handles JSON event requests
func (rt *route) handleJSONEvent(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	var bodySize int64

	// Read body size for metrics
	if r.ContentLength > 0 {
		bodySize = r.ContentLength
	}

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

	withEventLogger := logger.GetGlobalLogger().WithFields(map[string]any{
		"route_id":    rt.id,
		"event_id":    event.ID,
		"transport":   "http",
		"remote_addr": r.RemoteAddr,
		"body_size":   bodySize,
	})

	if logger.GetGlobalLogger().IsDebugEnabled() {
		withEventLogger.Debug("received http event",
			"params", event.Params,
			"method", r.Method,
			"user_agent", r.UserAgent(),
			"timestamp", startTime.Format(time.RFC3339),
		)
	}

	handlerInterface, ok := rt.services.EventRegistry.Get(rt.id, strings.ToLower(event.ID))
	if !ok {
		http.Error(w, "event id is not registered", http.StatusBadRequest)
		return
	}

	onEventFunc, ok := handlerInterface.(OnEventFunc)
	if !ok {
		http.Error(w, "invalid event handler type", http.StatusInternalServerError)
		return
	}

	// Time the event handler execution
	handlerStartTime := time.Now()
	result := onEventFunc(eventCtx)
	handlerDuration := time.Since(handlerStartTime)

	if logger.GetGlobalLogger().IsDebugEnabled() {
		withEventLogger.Debug("event handler completed",
			"handler_duration_ms", handlerDuration.Milliseconds(),
		)
	}

	renderStartTime := time.Now()
	// error event is not published
	errorEvent := handleOnEventResult(result, eventCtx, writeAndPublishEvents(eventCtx))
	renderDuration := time.Since(renderStartTime)
	totalDuration := time.Since(startTime)

	if logger.GetGlobalLogger().IsDebugEnabled() {
		withEventLogger.Debug("http event processing complete",
			"handler_duration_ms", handlerDuration.Milliseconds(),
			"render_duration_ms", renderDuration.Milliseconds(),
			"total_duration_ms", totalDuration.Milliseconds(),
		)
	}

	if errorEvent != nil {
		writeEventHTTP(eventCtx, *errorEvent)
	}
}

// handleFormPost handles form POST requests
func (rt *route) handleFormPost(w http.ResponseWriter, r *http.Request) {
	formAction := rt.determineFormAction(r)
	if formAction == "" {
		if len(rt.onEvents) > 1 {
			http.Error(w, "form action[?event=myaction] is missing and default onEvent can't be selected since there is more than 1", http.StatusBadRequest)
			return
		}
		// Use the single event handler if only one exists
		for k := range rt.onEvents {
			formAction = k
		}
	}

	event, err := rt.parseFormEvent(r, formAction)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	eventCtx := RouteContext{
		event:     event,
		request:   r,
		response:  w,
		route:     rt,
		urlValues: r.PostForm,
	}

	handlerInterface, ok := rt.services.EventRegistry.Get(rt.id, event.ID)
	if !ok {
		http.Error(w, fmt.Sprintf("onEvent handler for %s not found", event.ID), http.StatusBadRequest)
		return
	}

	onEventFunc, ok := handlerInterface.(OnEventFunc)
	if !ok {
		http.Error(w, "invalid event handler type", http.StatusInternalServerError)
		return
	}

	handlePostFormResult(onEventFunc(eventCtx), eventCtx)
}

// determineFormAction extracts the form action from query parameters
func (rt *route) determineFormAction(r *http.Request) string {
	values := r.URL.Query()
	if len(values) == 1 {
		if event := values.Get("event"); event != "" {
			return event
		}
	}
	return ""
}

// parseFormEvent parses form data into an Event struct
func (rt *route) parseFormEvent(r *http.Request, formAction string) (Event, error) {
	err := r.ParseForm()
	if err != nil {
		return Event{}, err
	}

	params, err := json.Marshal(r.PostForm)
	if err != nil {
		return Event{}, err
	}

	return Event{
		ID:     formAction,
		Params: params,
		IsForm: true,
	}, nil
}

// handleGetRequest handles GET requests (onLoad)
func (rt *route) handleGetRequest(w http.ResponseWriter, r *http.Request) {
	event := Event{ID: rt.routeOpt.id}
	eventCtx := RouteContext{
		event:    event,
		request:  r,
		response: w,
		route:    rt,
		isOnLoad: true,
	}
	handleOnLoadResult(rt.onLoad(eventCtx), nil, eventCtx)
}

// RouteInterface implementation methods
// These methods enable the route to be used as a RouteInterface in WebSocketServices

// ID returns the route ID
func (rt *route) ID() string {
	return rt.id
}

// Options returns the route options
func (rt *route) Options() interface{} {
	// Return the route options - we can return the routeOpt embedded struct
	return rt.routeOpt
}

// ChannelFunc returns the channel function
func (rt *route) ChannelFunc() func(r *http.Request, viewID string) *string {
	return rt.channelFunc
}

// PubSub returns the pubsub adapter
func (rt *route) PubSub() pubsub.Adapter {
	return rt.pubsub
}

// DevelopmentMode returns whether development mode is enabled
func (rt *route) DevelopmentMode() bool {
	return rt.developmentMode
}

// EventSender returns the event sender channel
func (rt *route) EventSender() interface{} {
	return rt.eventSender
}

// Services returns the route services
func (rt *route) Services() *routeservices.RouteServices {
	return rt.services
}

// FormDecoder returns the form decoder
func (rt *route) FormDecoder() *schema.Decoder {
	return rt.formDecoder
}

// GetRenderer returns the renderer
func (rt *route) GetRenderer() interface{} {
	return rt.renderer
}

// GetEventTemplates returns the event templates
func (rt *route) GetEventTemplates() interface{} {
	return rt.eventTemplates
}

// GetTemplate returns the route template
func (rt *route) GetTemplate() interface{} {
	return rt.getTemplate()
}

// GetAppName returns the app name
func (rt *route) GetAppName() string {
	return rt.appName
}

// buildTemplateConfig creates a template configuration from route options
func (rt *route) buildTemplateConfig() interface{} {
	// Create a simple config struct that can be used by template engines
	// This is a basic implementation - could be enhanced to use the full TemplateConfig interface
	return map[string]interface{}{
		"layout":                 rt.layout,
		"content":                rt.content,
		"errorLayout":            rt.errorLayout,
		"errorContent":           rt.errorContent,
		"partials":               rt.partials,
		"extensions":             rt.extensions,
		"layoutContentName":      rt.layoutContentName,
		"errorLayoutContentName": rt.errorLayoutContentName,
		"funcMap":                rt.getFuncMap(),
		"disableCache":           rt.disableTemplateCache,
	}
}

// parseTemplatesWithEngine attempts to use the new template engine if available,
// parseTemplatesWithEngine attempts to use the new template engine if available,
// otherwise uses the standard template parsing
func (rt *route) parseTemplatesWithEngine() error {
	// If we have a template engine, use it
	if rt.templateEngine != nil {
		return rt.parseTemplatesUsingEngine()
	}

	// Fall back to standard parsing using existing parse functions
	return rt.parseTemplatesStandard()
}

// parseTemplatesStandard uses the existing parse functions directly
func (rt *route) parseTemplatesStandard() error {
	rt.Lock()
	defer rt.Unlock()

	// Skip if template is already loaded and caching is enabled
	if rt.getTemplate() != nil && !rt.disableTemplateCache {
		return nil
	}

	var err error
	var successEventTemplates eventTemplates
	var rtTemplate *template.Template
	rtTemplate, successEventTemplates, err = parseTemplate(rt.routeOpt)
	if err != nil {
		logger.Errorf("error parsing template: %v", err)
		return err
	}
	rtTemplate.Option("missingkey=zero")
	rt.setTemplate(rtTemplate)

	// Store success event templates temporarily
	rt.setEventTemplates(successEventTemplates)

	var errorEventTemplates eventTemplates
	var rtErrorTemplate *template.Template
	rtErrorTemplate, errorEventTemplates, err = parseErrorTemplate(rt.routeOpt)
	if err != nil {
		return err
	}
	rtErrorTemplate.Option("missingkey=zero")
	rt.setErrorTemplate(rtErrorTemplate)

	rtEventTemplates := deepMergeEventTemplates(errorEventTemplates, successEventTemplates)
	rt.setEventTemplates(rtEventTemplates)

	return nil
}

// parseTemplatesUsingEngine uses the new template engine to parse templates
func (rt *route) parseTemplatesUsingEngine() error {
	rt.Lock()
	defer rt.Unlock()

	// Skip if template is already loaded and caching is enabled
	if rt.getTemplate() != nil && !rt.disableTemplateCache {
		return nil
	}

	// Cast template engine to the proper interface
	engine, ok := rt.templateEngine.(interface {
		LoadTemplate(config interface{}) (interface{}, error)
		LoadErrorTemplate(config interface{}) (interface{}, error)
	})
	if !ok {
		// Template engine doesn't support our interface, fall back to standard parsing
		return rt.parseTemplatesStandard()
	}

	// Build template config from route options
	config := rt.buildTemplateConfig()

	// Load main template using engine
	tmpl, err := engine.LoadTemplate(config)
	if err != nil {
		return fmt.Errorf("template engine failed to load template: %v", err)
	}

	// Convert and store the template
	if goTmpl, ok := tmpl.(*template.Template); ok {
		goTmpl.Option("missingkey=zero")
		rt.setTemplate(goTmpl)
	} else {
		// Template engine returned non-Go template, fall back to standard parsing
		return rt.parseTemplatesStandard()
	}

	// Load error template using engine
	errorTmpl, err := engine.LoadErrorTemplate(config)
	if err != nil {
		return fmt.Errorf("template engine failed to load error template: %v", err)
	}

	// Convert and store the error template
	if goErrorTmpl, ok := errorTmpl.(*template.Template); ok {
		goErrorTmpl.Option("missingkey=zero")
		rt.setErrorTemplate(goErrorTmpl)
	} else {
		// Error template engine returned non-Go template, fall back to standard parsing
		return rt.parseTemplatesStandard()
	}

	// TODO: Handle event templates through template engine
	// For now, we'll extract them from the main template using standard logic
	if mainTmpl := rt.getTemplate(); mainTmpl != nil {
		eventTemplates := make(eventTemplates)
		// Parse event templates from the main template content
		// This is a simplified extraction - in a full implementation we'd use the template engine
		rt.setEventTemplates(eventTemplates)
	}

	return nil
}
