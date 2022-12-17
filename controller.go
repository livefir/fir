package fir

import (
	"embed"
	"flag"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/schema"
	"github.com/gorilla/websocket"
	"github.com/livefir/fir/pubsub"
)

// Controller is an interface which encapsulates a group of views. It routes requests to the appropriate view.
// It routes events to the appropriate view. It also provides a way to register views.
type Controller interface {
	Route(route Route) http.HandlerFunc
	RouteFunc(options RouteFunc) http.HandlerFunc
}

type opt struct {
	channelFunc       func(r *http.Request, viewID string) *string
	websocketUpgrader websocket.Upgrader

	disableTemplateCache bool
	debugLog             bool
	enableWatch          bool
	watchExts            []string
	publicDir            string
	developmentMode      bool
	embedFS              embed.FS
	hasEmbedFS           bool
	pubsub               pubsub.Adapter
	appName              string
	formDecoder          *schema.Decoder
	validator            *validator.Validate
	session              *scs.SessionManager
}

// ControllerOption is an option for the controller.
type ControllerOption func(*opt)

// WithChannelFunc is an option to set a function to construct the channel name for the controller's views.
func WithChannel(f func(r *http.Request, viewID string) *string) ControllerOption {
	return func(o *opt) {
		o.channelFunc = f
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
		o.embedFS = fs
		o.hasEmbedFS = true
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

// WithValidator is an option to set the validator(go-playground/validator) for the controller.
func WithValidator(validator *validator.Validate) ControllerOption {
	return func(o *opt) {
		o.validator = validator
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

	sessionManager := scs.New()
	sessionManager.Cookie.Name = "_fir_session_"

	o := &opt{
		channelFunc:       defaultChannelFunc,
		websocketUpgrader: websocket.Upgrader{EnableCompression: true},
		watchExts:         defaultWatchExtensions,
		pubsub:            pubsub.NewInmem(),
		appName:           name,
		formDecoder:       formDecoder,
		validator:         validate,
		session:           sessionManager,
	}

	for _, option := range options {
		option(o)
	}

	if o.publicDir == "" {
		var publicDir string
		publicDirUsage := "public directory that contains the html template files."
		flag.StringVar(&publicDir, "public", ".", publicDirUsage)
		flag.StringVar(&publicDir, "p", ".", publicDirUsage+" (shortand)")
		flag.Parse()
		o.publicDir = publicDir
	}

	c := &controller{
		opt:  *o,
		name: name,
	}
	if c.developmentMode {
		log.Println("controller starting in developer mode ...", c.developmentMode)
		c.debugLog = true
		c.enableWatch = true
		c.disableTemplateCache = true
	}

	if c.enableWatch {
		go watchTemplates(c)
	}

	if c.hasEmbedFS {
		log.Println("read template files embedded in the binary")
	} else {
		log.Println("read template files from disk")
	}
	return c
}

type controller struct {
	name string
	opt
}

var defaultRouteOpt = &routeOpt{
	content:           "Hello Fir App!",
	layoutContentName: "content",
	partials:          []string{"./routes/partials"},
	funcMap:           defaultFuncMap(),
	extensions:        []string{".gohtml", ".gotmpl", ".html", ".tmpl"},
	eventSender:       make(chan Event),
	onLoad: func(ctx RouteContext) error {
		return nil
	},
}

// RouteFunc returns an http.HandlerFunc that renders the route
func (c *controller) Route(route Route) http.HandlerFunc {
	for _, option := range route.Options() {
		option(defaultRouteOpt)
	}

	return c.sessionHandlerFunc(newRoute(c, defaultRouteOpt))
}

// RouteFunc returns an http.HandlerFunc that renders the route
func (c *controller) RouteFunc(opts RouteFunc) http.HandlerFunc {
	for _, option := range opts() {
		option(defaultRouteOpt)
	}

	return c.sessionHandlerFunc(newRoute(c, defaultRouteOpt))
}

func (c *controller) sessionHandlerFunc(route *route) http.HandlerFunc {
	return c.session.LoadAndSave(route).ServeHTTP
}
