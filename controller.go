package fir

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/schema"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/websocket"
	"github.com/lithammer/shortuuid/v4"
	"github.com/livefir/fir/internal/file"
	"github.com/livefir/fir/internal/logger"
	internalMarkdown "github.com/livefir/fir/internal/markdown"
	"github.com/livefir/fir/pubsub"
	servertiming "github.com/mitchellh/go-server-timing"
	"github.com/patrickmn/go-cache"
)

// Controller is an interface which encapsulates a group of views. It routes requests to the appropriate view.
// It routes events to the appropriate view. It also provides a way to register views.
type Controller interface {
	Route(route Route) http.HandlerFunc
	RouteFunc(options RouteFunc) http.HandlerFunc
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
	readFile              file.ReadFileFunc
	existFile             file.ExistFileFunc
	pubsub                pubsub.Adapter
	appName               string
	formDecoder           *schema.Decoder
	cookieName            string
	secureCookie          *securecookie.SecureCookie
	cache                 *cache.Cache
	funcMap               template.FuncMap
	dropDuplicateInterval time.Duration
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
		opt:    *o,
		name:   name,
		routes: make(map[string]*route),
	}
	if c.developmentMode {
		fmt.Println("controller starting in developer mode")
		c.debugLog = true
		c.enableWatch = true
		c.disableTemplateCache = true
	}

	if c.enableWatch {
		go watchTemplates(c)
	}

	if c.embedfs != nil {
		c.readFile = file.ReadFileFS(*c.embedfs)
		c.existFile = file.ExistFileFS(*c.embedfs)
	} else {
		c.readFile = file.ReadFileOS
		c.existFile = file.ExistFileOS
	}

	md := internalMarkdown.New(c.readFile, c.existFile)
	c.funcMap["markdown"] = md
	c.funcMap["md"] = md
	c.opt.channelFunc = c.defaultChannelFunc

	return c
}

type controller struct {
	name   string
	routes map[string]*route
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
	}
	return defaultRouteOpt
}

// Route returns an http.HandlerFunc that renders the route
func (c *controller) Route(route Route) http.HandlerFunc {
	defaultRouteOpt := c.defaults()
	for _, option := range route.Options() {
		option(defaultRouteOpt)
	}

	// create new route
	r, err := newRoute(c, defaultRouteOpt)
	if err != nil {
		// Return a handler that serves the error with a non-zero exit code
		return func(w http.ResponseWriter, req *http.Request) {
			logger.Errorf("route creation failed: %v", err)
			http.Error(w, fmt.Sprintf("route creation failed: %v", err), http.StatusInternalServerError)
		}
	}
	// register route in the controller
	c.routes[r.id] = r
	return servertiming.Middleware(r, nil).ServeHTTP
}

// RouteFunc returns an http.HandlerFunc that renders the route
func (c *controller) RouteFunc(opts RouteFunc) http.HandlerFunc {
	defaultRouteOpt := c.defaults()
	for _, option := range opts() {
		option(defaultRouteOpt)
	}
	// create new route
	r, err := newRoute(c, defaultRouteOpt)
	if err != nil {
		// Return a handler that serves the error with a non-zero exit code
		return func(w http.ResponseWriter, req *http.Request) {
			logger.Errorf("route creation failed: %v", err)
			http.Error(w, fmt.Sprintf("route creation failed: %v", err), http.StatusInternalServerError)
		}
	}
	// register route in the controller
	c.routes[r.id] = r

	return servertiming.Middleware(r, nil).ServeHTTP
}
