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
	"github.com/livefir/fir/internal/handlers"
	firHttp "github.com/livefir/fir/internal/http"
	"github.com/livefir/fir/internal/logger"
	routeFactory "github.com/livefir/fir/internal/route"
	"github.com/livefir/fir/internal/routeservices"
	"github.com/livefir/fir/internal/services"
	"github.com/livefir/fir/pubsub"
	servertiming "github.com/mitchellh/go-server-timing"
)

// Helper functions to access route components through RouteInterface

// getRenderer returns the renderer from the route interface
func getRenderer(ctx RouteContext) Renderer {
	return ctx.routeInterface.GetRenderer().(Renderer)
}

// getServices returns the services from the route interface
func getServices(ctx RouteContext) *routeservices.RouteServices {
	return ctx.routeInterface.Services()
}

// getRoute returns the underlying route from the route interface
func getRoute(ctx RouteContext) *route {
	return ctx.routeInterface.(*route)
}

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

	// Handler chain for processing requests
	handlerChain handlers.HandlerChain

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
	// Create route service factory for dependency injection
	factory := routeFactory.NewRouteServiceFactory(services)

	// Create handler chain using factory
	handlerChain := factory.CreateHandlerChain()

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
		handlerChain:         handlerChain,
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
		// Get route interface services and channel function
		services := ctx.routeInterface.Services()
		channel := services.ChannelFunc(ctx.request, ctx.routeInterface.ID())
		if channel == nil {
			logger.Errorf("error: channel is empty")
			http.Error(ctx.response, "channel is empty", http.StatusUnauthorized)
			return nil
		}
		err := services.PubSub.Publish(ctx.request.Context(), *channel, pubsubEvent)
		if err != nil {
			logger.Debugf("error publishing patch: %v", err)
		}

		return writeEventHTTP(ctx, pubsubEvent)
	}
}

func writeEventHTTP(ctx RouteContext, event pubsub.Event) error {
	// Get renderer from route interface
	renderer := ctx.routeInterface.GetRenderer().(Renderer)
	events := renderer.RenderDOMEvents(ctx, event)
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

func (rt *route) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	timing := servertiming.FromContext(r.Context())
	defer timing.NewMetric("route").Start().Stop()

	// Handle special requests using legacy code for now
	if !rt.handleSpecialRequests(w, r) {
		return
	}

	// Setup path parameters if needed
	r = rt.setupPathParameters(r)

	// Phase 1: Only use handler chain for POC route (/poc)
	// All other routes use legacy handling for now
	if r.Method == "GET" && r.URL.Path == "/poc" {
		err := rt.handleRequestWithChain(w, r)
		if err != nil {
			// Fallback to legacy handling if handler chain fails
			rt.handleRequestLegacy(w, r)
		}
	} else {
		// Use legacy handling for all non-POC routes
		rt.handleRequestLegacy(w, r)
	}
}

// handleRequestWithChain processes the request using the new handler chain
func (rt *route) handleRequestWithChain(w http.ResponseWriter, r *http.Request) error {
	// Create request/response pair using the HTTP adapter
	pair, err := firHttp.NewRequestResponsePair(w, r, rt.pathParamsFunc)
	if err != nil {
		return err
	}

	// Check if handler chain can handle this request type
	if !rt.canHandlerChainHandle(pair.Request) {
		logger.GetGlobalLogger().Debug("handler chain cannot handle request type, using legacy fallback",
			"method", r.Method,
			"path", r.URL.Path,
			"content_type", r.Header.Get("Content-Type"),
			"route_id", rt.id,
		)
		return fmt.Errorf("handler chain cannot handle request type: %s %s", r.Method, r.URL.Path)
	}

	// Process request through handler chain
	response, err := rt.handlerChain.Handle(r.Context(), pair.Request)
	if err != nil {
		// Add debugging information when handler chain fails
		logger.GetGlobalLogger().Debug("handler chain failed, falling back to legacy",
			"error", err.Error(),
			"method", r.Method,
			"path", r.URL.Path,
			"content_type", r.Header.Get("Content-Type"),
			"route_id", rt.id,
		)
		return err
	}

	// Write the response from handler chain to HTTP response writer
	if response != nil {
		err = pair.Response.WriteResponse(*response)
		if err != nil {
			logger.GetGlobalLogger().Debug("failed to write handler chain response",
				"error", err.Error(),
				"method", r.Method,
				"path", r.URL.Path,
				"route_id", rt.id,
			)
			return err
		}
	}

	return nil
}

// canHandlerChainHandle checks if the handler chain can process the given request
func (rt *route) canHandlerChainHandle(req *firHttp.RequestModel) bool {
	if rt.handlerChain == nil {
		return false
	}

	// Check if any handler in the chain supports this request
	chainHandlers := rt.handlerChain.GetHandlers()
	if len(chainHandlers) == 0 {
		logger.GetGlobalLogger().Debug("handler chain has no handlers configured",
			"route_id", rt.id,
		)
		return false
	}

	for _, handler := range chainHandlers {
		if handler.SupportsRequest(req) {
			logger.GetGlobalLogger().Debug("handler chain can handle request",
				"handler", handler.HandlerName(),
				"method", req.Method,
				"path", req.URL.Path,
				"route_id", rt.id,
			)
			return true
		}
	}

	logger.GetGlobalLogger().Debug("no handler in chain supports request",
		"method", req.Method,
		"path", req.URL.Path,
		"content_type", req.Header.Get("Content-Type"),
		"route_id", rt.id,
		"handler_count", len(chainHandlers),
	)
	return false
}

// handleRequestLegacy provides fallback to the original request handling logic
func (rt *route) handleRequestLegacy(w http.ResponseWriter, r *http.Request) {
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

	switch errVal := err.(type) {
	case *routeData:
		// Render template with the data
		getRenderer(ctx).RenderRoute(ctx, *errVal, false)
	case *stateData:
		// For state data, render the template
		getRenderer(ctx).RenderRoute(ctx, routeData(*errVal), false)
	case *routeDataWithState:
		// Render template with route data
		getRenderer(ctx).RenderRoute(ctx, *errVal.routeData, false)
	default:
		// Get onLoad handler from EventRegistry
		handlerInterface, ok := getServices(ctx).EventRegistry.Get(ctx.routeInterface.ID(), "load")
		if ok {
			if onLoadFunc, ok := handlerInterface.(OnEventFunc); ok {
				// Create a new context for onLoad with initialized pointers
				onLoadCtx := RouteContext{
					event:            ctx.event,
					request:          ctx.request,
					response:         ctx.response,
					routeInterface:   ctx.routeInterface,
					urlValues:        ctx.urlValues,
					isOnLoad:         true,
					accumulatedData:  &map[string]any{},
					accumulatedState: &map[string]any{},
				}
				onLoadErr := onLoadFunc(onLoadCtx)
				accumulatedDataErr := onLoadCtx.GetAccumulatedData()

				// Use accumulated data if available, otherwise use handler error
				var onLoadResult error
				if accumulatedDataErr != nil {
					onLoadResult = accumulatedDataErr
				} else {
					onLoadResult = onLoadErr
				}
				handleOnLoadResult(onLoadResult, err, ctx)
			}
		} else {
			// Fallback to legacy onLoad
			handleOnLoadResult(getRoute(ctx).onLoad(ctx), err, ctx)
		}
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

		getRenderer(ctx).RenderRoute(ctx, routeData{"errors": errs}, false)
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
		getRenderer(ctx).RenderRoute(ctx, onLoadData, false)

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
		getRenderer(ctx).RenderRoute(ctx, onLoadData, false)

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

		getRenderer(ctx).RenderRoute(ctx, routeData{"errors": errs}, true)
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

		getRenderer(ctx).RenderRoute(ctx, routeData{"errors": errs}, false)
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
		getRenderer(ctx).RenderRoute(ctx, routeData{"errors": errs}, false)
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
		event:            event,
		request:          r,
		response:         w,
		routeInterface:   rt, // Use route as RouteInterface for both WebSocket and HTTP modes
		accumulatedData:  &map[string]any{},
		accumulatedState: &map[string]any{},
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
	handlerErr := onEventFunc(eventCtx)

	// Get accumulated data from the context
	accumulatedDataErr := eventCtx.GetAccumulatedData()

	// Use accumulated data if available, otherwise use handler error
	var result error
	if accumulatedDataErr != nil {
		result = accumulatedDataErr
	} else {
		result = handlerErr
	}

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

// handleJSONEventWithService handles JSON event requests using the new event service
//
//nolint:unused // Legacy method used by integration tests during migration phase
func (rt *route) handleJSONEventWithService(w http.ResponseWriter, r *http.Request) {
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

	withEventLogger := logger.GetGlobalLogger().WithFields(map[string]any{
		"route_id":    rt.id,
		"event_id":    event.ID,
		"transport":   "http",
		"remote_addr": r.RemoteAddr,
		"body_size":   bodySize,
	})

	if logger.GetGlobalLogger().IsDebugEnabled() {
		withEventLogger.Debug("received http event (service layer)",
			"params", event.Params,
			"method", r.Method,
			"user_agent", r.UserAgent(),
			"timestamp", startTime.Format(time.RFC3339),
		)
	}

	// Use the new event service if available
	if rt.services.EventService != nil {
		processor := NewRouteEventProcessor(rt.services.EventService, rt)

		handlerStartTime := time.Now()
		response, err := processor.ProcessEvent(r.Context(), event, r, w)
		handlerDuration := time.Since(handlerStartTime)

		if logger.GetGlobalLogger().IsDebugEnabled() {
			withEventLogger.Debug("event service processing completed",
				"handler_duration_ms", handlerDuration.Milliseconds(),
			)
		}

		if err != nil {
			withEventLogger.Error("event service processing failed", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Write the response
		rt.writeEventServiceResponse(w, response, withEventLogger, startTime, handlerDuration)
		return
	}

	// Fallback to legacy event handling
	rt.handleJSONEvent(w, r)
}

// writeEventServiceResponse writes the event service response to the HTTP response
//
//nolint:unused // Legacy method used by legacy event handling during migration phase
func (rt *route) writeEventServiceResponse(w http.ResponseWriter, response *services.EventResponse, logger *logger.Logger, startTime time.Time, handlerDuration time.Duration) {
	renderStartTime := time.Now()

	// Set status code
	w.WriteHeader(response.StatusCode)

	// Set headers
	for key, value := range response.Headers {
		w.Header().Set(key, value)
	}

	// Write body
	if len(response.Body) > 0 {
		w.Write(response.Body)
	}

	// Publish PubSub events
	for _, pubsubEvent := range response.PubSubEvents {
		channel := rt.services.ChannelFunc(nil, rt.id) // TODO: pass proper request
		channelStr := ""
		if channel != nil {
			channelStr = *channel
		}
		rt.services.PubSub.Publish(context.Background(), channelStr, pubsubEvent)
	}

	renderDuration := time.Since(renderStartTime)
	totalDuration := time.Since(startTime)

	if logger.IsDebugEnabled() {
		logger.Debug("http event processing complete (service layer)",
			"handler_duration_ms", handlerDuration.Milliseconds(),
			"render_duration_ms", renderDuration.Milliseconds(),
			"total_duration_ms", totalDuration.Milliseconds(),
		)
	}
}

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
		event:            event,
		request:          r,
		response:         w,
		routeInterface:   rt,
		urlValues:        r.PostForm,
		accumulatedData:  &map[string]any{},
		accumulatedState: &map[string]any{},
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

	// Call the event handler
	handlerErr := onEventFunc(eventCtx)

	// Get accumulated data from the context
	accumulatedDataErr := eventCtx.GetAccumulatedData()

	// Use accumulated data if available, otherwise use handler error
	var resultErr error
	if accumulatedDataErr != nil {
		resultErr = accumulatedDataErr
	} else {
		resultErr = handlerErr
	}

	handlePostFormResult(resultErr, eventCtx)
}

// determineFormAction extracts the form action from query parameters
func (rt *route) determineFormAction(r *http.Request) string {
	values := r.URL.Query()
	if event := values.Get("event"); event != "" {
		return event
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
		event:            event,
		request:          r,
		response:         w,
		routeInterface:   rt,
		isOnLoad:         true,
		accumulatedData:  &map[string]any{},
		accumulatedState: &map[string]any{},
	}

	// Get onLoad handler from EventRegistry
	handlerInterface, ok := rt.services.EventRegistry.Get(rt.id, "load")
	if ok {
		if onLoadFunc, ok := handlerInterface.(OnEventFunc); ok {
			// Call onLoad handler and get accumulated data
			onLoadErr := onLoadFunc(eventCtx)
			accumulatedDataErr := eventCtx.GetAccumulatedData()

			// Use accumulated data if available, otherwise use handler error
			var onLoadResult error
			if accumulatedDataErr != nil {
				onLoadResult = accumulatedDataErr
			} else {
				onLoadResult = onLoadErr
			}
			handleOnLoadResult(onLoadResult, nil, eventCtx)
		}
	} else {
		// Fallback to legacy onLoad
		handleOnLoadResult(rt.onLoad(eventCtx), nil, eventCtx)
	}
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

// GetCookieName returns the session cookie name
func (rt *route) GetCookieName() string {
	return rt.cookieName
}

// GetSecureCookie returns the secure cookie instance
func (rt *route) GetSecureCookie() interface{} {
	return rt.secureCookie
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

// parseTemplatesWithEngine uses a custom template engine if available,
// otherwise uses the standard template parsing
func (rt *route) parseTemplatesWithEngine() error {
	// If we have a template engine, use it
	if rt.templateEngine != nil {
		return rt.parseTemplatesUsingEngine()
	}

	// Use standard parsing using existing parse functions
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
		return fmt.Errorf("template engine does not implement required interface with LoadTemplate and LoadErrorTemplate methods")
	}

	// Build template config from route options
	config := rt.buildTemplateConfig()

	// Load main template using engine
	tmpl, err := engine.LoadTemplate(config)
	if err != nil {
		return fmt.Errorf("template engine failed to load template: %v", err)
	}

	// Convert and store the template
	goTmpl, ok := tmpl.(*template.Template)
	if !ok {
		return fmt.Errorf("template engine returned unsupported template type, expected *template.Template")
	}
	goTmpl.Option("missingkey=zero")
	rt.setTemplate(goTmpl)

	// Load error template using engine
	errorTmpl, err := engine.LoadErrorTemplate(config)
	if err != nil {
		return fmt.Errorf("template engine failed to load error template: %v", err)
	}

	// Convert and store the error template
	goErrorTmpl, ok := errorTmpl.(*template.Template)
	if !ok {
		return fmt.Errorf("template engine returned unsupported error template type, expected *template.Template")
	}
	goErrorTmpl.Option("missingkey=zero")
	rt.setErrorTemplate(goErrorTmpl)

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
