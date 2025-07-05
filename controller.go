package fir

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/schema"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/websocket"
	"github.com/lithammer/shortuuid/v4"
	"github.com/livefir/fir/internal/event"
	"github.com/livefir/fir/internal/logger"
	"github.com/livefir/fir/internal/routeservices"
	"github.com/livefir/fir/pubsub"
	servertiming "github.com/mitchellh/go-server-timing"
	"github.com/patrickmn/go-cache"
)

// Controller is an interface which encapsulates a group of views. It routes requests to the appropriate view.
// It routes events to the appropriate view. It also provides a way to register views.
type Controller interface {
	Route(route Route) http.HandlerFunc
	RouteFunc(options RouteFunc) http.HandlerFunc
	// GetEventRegistry returns the event registry for debug introspection
	GetEventRegistry() event.EventRegistry
}

type opt struct {
	onSocketConnect    func(userOrSessionID string) error
	onSocketDisconnect func(userOrSessionID string)
	channelFunc        func(r *http.Request, viewID string) *string
	pathParamsFunc     func(r *http.Request) PathParams
	websocketUpgrader  websocket.Upgrader

	disableTemplateCache  bool
	disableWebsocket      bool
	debugLog              bool
	enableWatch           bool
	watchExts             []string
	publicDir             string
	developmentMode       bool
	embedfs               *embed.FS
	readFile              readFileFunc
	existFile             existFileFunc
	pubsub                pubsub.Adapter
	appName               string
	formDecoder           *schema.Decoder
	cookieName            string
	secureCookie          *securecookie.SecureCookie
	cache                 *cache.Cache
	funcMap               template.FuncMap
	dropDuplicateInterval time.Duration
	renderer              Renderer
}

// ControllerOption is an option for the controller.
type ControllerOption func(*opt)

func WithFuncMap(funcMap template.FuncMap) ControllerOption {
	return func(opt *opt) {
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

// WithSessionSecrets is an option to set the session secrets for the controller.
// used to sign and encrypt the session cookie.
func WithSessionSecrets(hashKey []byte, blockKey []byte) ControllerOption {
	return func(o *opt) {
		o.secureCookie = securecookie.New(hashKey, blockKey)
	}
}

// WithSessionName is an option to set the session name/cookie name for the controller.
func WithSessionName(name string) ControllerOption {
	return func(o *opt) {
		o.cookieName = name
	}
}

// WithChannelFunc is an option to set a function to construct the channel name for the controller's views.
func WithChannelFunc(f func(r *http.Request, viewID string) *string) ControllerOption {
	return func(o *opt) {
		o.channelFunc = f
	}
}

// WithPathParamsFunc is an option to set a function to construct the path params for the controller's views.
func WithPathParamsFunc(f func(r *http.Request) PathParams) ControllerOption {
	return func(o *opt) {
		o.pathParamsFunc = f
	}
}

// WithPubsubAdapter is an option to set a pubsub adapter for the controller's views.
func WithPubsubAdapter(pubsub pubsub.Adapter) ControllerOption {
	return func(o *opt) {
		o.pubsub = pubsub
	}
}

// WithWebsocketUpgrader is an option to set the websocket upgrader for the controller
func WithWebsocketUpgrader(upgrader websocket.Upgrader) ControllerOption {
	return func(o *opt) {
		o.websocketUpgrader = upgrader
	}
}

// WithEmbedFS is an option to set the embed.FS for the controller.
func WithEmbedFS(fs embed.FS) ControllerOption {
	return func(o *opt) {
		o.embedfs = &fs
	}
}

// WithPublicDir is the path to directory containing the public html template files.
func WithPublicDir(path string) ControllerOption {
	return func(o *opt) {
		o.publicDir = path
	}
}

// WithFormDecoder is an option to set the form decoder(gorilla/schema) for the controller.
func WithFormDecoder(decoder *schema.Decoder) ControllerOption {
	return func(o *opt) {
		o.formDecoder = decoder
	}
}

// WithDisableWebsocket is an option to disable websocket.
func WithDisableWebsocket() ControllerOption {
	return func(o *opt) {
		o.disableWebsocket = true
	}
}

// WithDropDuplicateInterval is an option to set the interval to drop duplicate events received by the websocket.
func WithDropDuplicateInterval(interval time.Duration) ControllerOption {
	return func(o *opt) {
		o.dropDuplicateInterval = interval
	}
}

// WithOnSocketConnect takes a function that is called when a new websocket connection is established.
// The function should return an error if the connection should be rejected.
// The user or fir's browser session id is passed to the function.
// user must be set in request.Context with the key UserKey by a developer supplied authentication mechanism.
// It can be used to track user connections and disconnections.
// It can be be used to reject connections based on user or session id.
// It can be used to refresh the page data when a user re-connects.
func WithOnSocketConnect(f func(userOrSessionID string) error) ControllerOption {
	return func(o *opt) {
		o.onSocketConnect = f
	}
}

// WithOnSocketDisconnect takes a function that is called when a websocket connection is disconnected.
func WithOnSocketDisconnect(f func(userOrSessionID string)) ControllerOption {
	return func(o *opt) {
		o.onSocketDisconnect = f
	}

}

// DisableTemplateCache is an option to disable template caching. This is useful for development.
func DisableTemplateCache() ControllerOption {
	return func(o *opt) {
		o.disableTemplateCache = true
	}
}

// EnableDebugLog is an option to enable debug logging.
func EnableDebugLog() ControllerOption {
	return func(o *opt) {
		o.debugLog = true
	}
}

// EnableWatch is an option to enable watching template files for changes.
func EnableWatch(rootDir string, extensions ...string) ControllerOption {
	return func(o *opt) {
		o.enableWatch = true
		if len(extensions) > 0 {
			o.publicDir = rootDir
			o.watchExts = append(o.watchExts, extensions...)
		}
	}
}

// DevelopmentMode is an option to enable development mode. It enables debug logging, template watching, and disables template caching.
func DevelopmentMode(enable bool) ControllerOption {
	return func(o *opt) {
		o.developmentMode = enable
	}
}

// WithRenderer is an option to set a custom renderer for the controller's routes.
func WithRenderer(renderer Renderer) ControllerOption {
	return func(o *opt) {
		o.renderer = renderer
	}
}

// WithDebug enables comprehensive debug mode with enhanced logging.
// This option configures the global logger for debug output and enables
// debug-specific instrumentation for the debug UI.
func WithDebug(enable bool) ControllerOption {
	return func(o *opt) {
		o.debugLog = enable
		if enable {
			// Configure global logger for debug mode
			config := logger.Config{
				Level:       logger.LevelDebug,
				Format:      "json",
				EnableDebug: true,
				Output:      os.Stdout,
			}
			debugLogger := logger.NewLogger(config)
			logger.SetGlobalLogger(debugLogger)
		}
	}
}

// NewController creates a new controller.
func NewController(name string, options ...ControllerOption) Controller {
	if name == "" {
		panic("controller name is required")
	}

	formDecoder := schema.NewDecoder()
	formDecoder.IgnoreUnknownKeys(true)
	formDecoder.SetAliasTag("json")

	validate := validator.New()
	// register function to get tag name from json tags.
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	o := &opt{
		websocketUpgrader: websocket.Upgrader{
			// disabled compression since its too noisy: https://github.com/gorilla/websocket/issues/859
			// EnableCompression: true,
			// ReadBufferSize:  4096,
			// WriteBufferSize: 4096,
			// WriteBufferPool: &sync.Pool{},
		},
		watchExts:   defaultWatchExtensions,
		pubsub:      pubsub.NewInmem(),
		appName:     name,
		formDecoder: formDecoder,
		cookieName:  "_fir_session_",
		secureCookie: securecookie.New(
			securecookie.GenerateRandomKey(64),
			securecookie.GenerateRandomKey(32),
		),
		cache:                 cache.New(5*time.Minute, 10*time.Minute),
		funcMap:               defaultFuncMap(),
		dropDuplicateInterval: 250 * time.Millisecond,
		publicDir:             ".",
	}

	for _, option := range options {
		option(o)
	}

	c := &controller{
		opt:           *o,
		name:          name,
		routes:        make(map[string]*route),
		eventRegistry: event.NewEventRegistry(),
	}

	if c.embedfs != nil {
		c.readFile = readFileFS(*c.embedfs)
		c.existFile = existFileFS(*c.embedfs)
	} else {
		c.readFile = readFileOS
		c.existFile = existFileOS
	}

	md := markdown(c.readFile, c.existFile)
	c.funcMap["markdown"] = md
	c.funcMap["md"] = md
	c.opt.channelFunc = c.defaultChannelFunc

	// Initialize RouteServices once for reuse (after channelFunc is set)
	c.routeServices = c.createRouteServices()

	if c.developmentMode {
		fmt.Println("controller starting in developer mode")
		c.debugLog = true
		c.enableWatch = true
		c.disableTemplateCache = true
	}

	if c.enableWatch {
		go watchTemplates(c)
	}

	return c
}

type controller struct {
	name          string
	routes        map[string]*route
	eventRegistry event.EventRegistry
	routeServices *routeservices.RouteServices // Cached RouteServices instance
	opt
}

func (c *controller) defaults() *routeOpt {
	defaultRouteOpt := &routeOpt{
		id:                shortuuid.New(),
		content:           "Hello Fir App!",
		layoutContentName: "content",
		partials:          []string{}, // Remove default routes/partials path
		funcMap:           c.opt.funcMap,
		extensions:        []string{".gohtml", ".gotmpl", ".html", ".tmpl"},
		eventSender:       make(chan Event),
		onLoad: func(ctx RouteContext) error {
			return nil
		},
		funcMapMutex: &sync.RWMutex{},
		opt:          c.opt, // Set the embedded opt struct
	}
	return defaultRouteOpt
}

// Route returns an http.HandlerFunc that renders the route
func (c *controller) Route(route Route) http.HandlerFunc {
	return c.createRouteHandler(route.Options())
}

// RouteFunc returns an http.HandlerFunc that renders the route
func (c *controller) RouteFunc(opts RouteFunc) http.HandlerFunc {
	return c.createRouteHandler(opts())
}

// GetEventRegistry returns the event registry for debug introspection
// This method is primarily intended for debug tools and static analysis
func (c *controller) GetEventRegistry() event.EventRegistry {
	return c.eventRegistry
}

// createRouteServices creates a RouteServices instance from the controller's configuration
func (c *controller) createRouteServices() *routeservices.RouteServices {
	options := &routeservices.Options{
		OnSocketConnect:       c.opt.onSocketConnect,
		OnSocketDisconnect:    c.opt.onSocketDisconnect,
		WebsocketUpgrader:     c.opt.websocketUpgrader,
		DisableTemplateCache:  c.opt.disableTemplateCache,
		DisableWebsocket:      c.opt.disableWebsocket,
		EnableWatch:           c.opt.enableWatch,
		WatchExts:             c.opt.watchExts,
		PublicDir:             c.opt.publicDir,
		DevelopmentMode:       c.opt.developmentMode,
		ReadFile:              c.opt.readFile,
		ExistFile:             c.opt.existFile,
		AppName:               c.opt.appName,
		FormDecoder:           c.opt.formDecoder,
		CookieName:            c.opt.cookieName,
		SecureCookie:          c.opt.secureCookie,
		Cache:                 c.opt.cache,
		FuncMap:               c.opt.funcMap,
		DropDuplicateInterval: c.opt.dropDuplicateInterval,
		DebugLog:              c.opt.debugLog,
	}

	// Ensure we have a renderer - use default if none specified
	renderer := c.opt.renderer
	if renderer == nil {
		renderer = NewTemplateRenderer()
	}

	// TODO: Create template engine factory for Milestone 5
	// For now, pass nil to maintain backward compatibility
	templateEngine := c.createTemplateEngineFactory()

	services := routeservices.NewRouteServicesWithTemplateEngine(c.eventRegistry, c.opt.pubsub, renderer, templateEngine, options)
	services.SetChannelFunc(c.opt.channelFunc)

	// Set WebSocketServices - controller implements WebSocketServices interface
	services.SetWebSocketServices(c)

	// Convert PathParams function signature
	if c.opt.pathParamsFunc != nil {
		services.SetPathParamsFunc(func(r *http.Request) map[string]string {
			pathParams := c.opt.pathParamsFunc(r)
			result := make(map[string]string)
			for k, v := range pathParams {
				result[k] = fmt.Sprintf("%v", v)
			}
			return result
		})
	}

	return services
}

// GetRouteServices returns the RouteServices instance for this controller
// This allows external code to access the services if needed for testing or debugging
func (c *controller) GetRouteServices() *routeservices.RouteServices {
	return c.routeServices
}

// UpdateRouteServices allows updating the RouteServices configuration
// This is useful for runtime configuration changes
func (c *controller) UpdateRouteServices() {
	c.routeServices = c.createRouteServices()
}

// RouteFactory encapsulates route creation logic and provides validation
type RouteFactory struct {
	controller *controller
}

// NewRouteFactory creates a new route factory for the controller
func (c *controller) NewRouteFactory() *RouteFactory {
	return &RouteFactory{controller: c}
}

// createRouteHandler is the main factory method that creates route handlers
// This method abstracts the route creation logic from the public Route/RouteFunc methods
func (c *controller) createRouteHandler(options RouteOptions) http.HandlerFunc {
	routeOpt, err := c.buildRouteOptions(options)
	if err != nil {
		return c.createErrorHandler("route option validation failed", err)
	}

	route, err := c.createAndValidateRoute(routeOpt)
	if err != nil {
		return c.createErrorHandler("route creation failed", err)
	}

	c.registerRoute(route)
	return c.wrapRouteHandler(route)
}

// buildRouteOptions creates and validates route options from the provided options
func (c *controller) buildRouteOptions(options RouteOptions) (*routeOpt, error) {
	defaultRouteOpt := c.defaults()

	// Apply all route options
	for _, option := range options {
		option(defaultRouteOpt)
	}

	// Validate the route options
	if err := c.validateRouteOptions(defaultRouteOpt); err != nil {
		return nil, fmt.Errorf("route validation failed: %v", err)
	}

	return defaultRouteOpt, nil
}

// createAndValidateRoute creates a new route and validates its creation
func (c *controller) createAndValidateRoute(routeOpt *routeOpt) (*route, error) {
	route, err := newRoute(c.routeServices, routeOpt)
	if err != nil {
		return nil, fmt.Errorf("failed to create route: %v", err)
	}

	// Additional post-creation validation if needed
	if err := c.validateCreatedRoute(route); err != nil {
		return nil, fmt.Errorf("route post-creation validation failed: %v", err)
	}

	return route, nil
}

// validateRouteOptions validates route configuration before creation
func (c *controller) validateRouteOptions(routeOpt *routeOpt) error {
	if routeOpt.id == "" {
		return fmt.Errorf("route ID cannot be empty")
	}

	// Check for duplicate route IDs
	if _, exists := c.routes[routeOpt.id]; exists {
		return fmt.Errorf("route with ID '%s' already exists", routeOpt.id)
	}

	// Validate content is provided
	if routeOpt.content == "" && routeOpt.layout == "" {
		return fmt.Errorf("route must have either content or layout specified")
	}

	// Validate template extensions
	if len(routeOpt.extensions) == 0 {
		logger.Warnf("route '%s' has no template extensions specified, using defaults", routeOpt.id)
	}

	return nil
}

// validateCreatedRoute performs post-creation validation on the route
func (c *controller) validateCreatedRoute(route *route) error {
	if route == nil {
		return fmt.Errorf("route is nil")
	}

	if route.services == nil {
		return fmt.Errorf("route services are not initialized")
	}

	if route.renderer == nil {
		return fmt.Errorf("route renderer is not initialized")
	}

	return nil
}

// registerRoute registers the route in the controller's route map
func (c *controller) registerRoute(route *route) {
	c.routes[route.id] = route
	logger.Debugf("registered route with ID: %s", route.id)
}

// wrapRouteHandler wraps the route with middleware and returns the final handler
func (c *controller) wrapRouteHandler(route *route) http.HandlerFunc {
	return servertiming.Middleware(route, nil).ServeHTTP
}

// createErrorHandler creates a handler that serves errors with proper HTTP status
func (c *controller) createErrorHandler(message string, err error) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		fullMessage := fmt.Sprintf("%s: %v", message, err)
		logger.Errorf("%s", fullMessage)
		http.Error(w, fullMessage, http.StatusInternalServerError)
	}
}

// RouteCreationOptions provides options for advanced route creation
type RouteCreationOptions struct {
	ValidateBeforeCreation bool
	SkipDuplicateCheck     bool
	EnableDebugLogging     bool
}

// CreateRouteWithOptions creates a route with advanced options (for future extensibility)
func (c *controller) CreateRouteWithOptions(options RouteOptions, creationOpts RouteCreationOptions) (http.HandlerFunc, error) {
	if creationOpts.EnableDebugLogging {
		logger.Debugf("creating route with advanced options: %+v", creationOpts)
	}

	routeOpt, err := c.buildRouteOptions(options)
	if err != nil {
		return nil, err
	}

	// Skip duplicate check if requested
	if creationOpts.SkipDuplicateCheck {
		if _, exists := c.routes[routeOpt.id]; exists {
			logger.Warnf("duplicate route ID '%s' detected but skipping check as requested", routeOpt.id)
		}
	}

	route, err := c.createAndValidateRoute(routeOpt)
	if err != nil {
		return nil, err
	}

	c.registerRoute(route)
	return c.wrapRouteHandler(route), nil
}

// WebSocketServices interface implementation
// These methods enable the controller to act as a WebSocketServices provider

// GetWebSocketUpgrader returns the WebSocket upgrader configuration
func (c *controller) GetWebSocketUpgrader() *websocket.Upgrader {
	return &c.opt.websocketUpgrader
}

// GetRoutes returns the routes map as RouteInterface map
func (c *controller) GetRoutes() map[string]routeservices.RouteInterface {
	routes := make(map[string]routeservices.RouteInterface)
	for id, route := range c.routes {
		routes[id] = route
	}
	return routes
}

// DecodeSession decodes a session ID and returns user/session ID and route ID
func (c *controller) DecodeSession(sessionID string) (userOrSessionID, routeID string, err error) {
	return decodeSession(*c.opt.secureCookie, c.opt.cookieName, sessionID)
}

// GetCookieName returns the session cookie name
func (c *controller) GetCookieName() string {
	return c.opt.cookieName
}

// GetDropDuplicateInterval returns the event deduplication interval
func (c *controller) GetDropDuplicateInterval() time.Duration {
	return c.opt.dropDuplicateInterval
}

// IsWebSocketDisabled returns whether WebSocket is disabled
func (c *controller) IsWebSocketDisabled() bool {
	return c.opt.disableWebsocket
}

// OnSocketConnect handles socket connection events
func (c *controller) OnSocketConnect(userOrSessionID string) error {
	if c.opt.onSocketConnect != nil {
		return c.opt.onSocketConnect(userOrSessionID)
	}
	return nil
}

// OnSocketDisconnect handles socket disconnection events
func (c *controller) OnSocketDisconnect(userOrSessionID string) {
	if c.opt.onSocketDisconnect != nil {
		c.opt.onSocketDisconnect(userOrSessionID)
	}
}

// createTemplateEngineFactory creates a template engine factory with default configuration
func (c *controller) createTemplateEngineFactory() interface{} {
	// Since we can't import templateengine here directly due to circular imports,
	// we'll create the template engine using a factory pattern.
	// This approach allows us to avoid circular imports while enabling template engine usage.

	// For now, we'll continue returning nil to maintain backward compatibility.
	// Template engines can be explicitly provided via route options.
	// In a future version, we'll enable a default template engine here.
	return nil
}
