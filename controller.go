package fir

import (
	"embed"
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// Controller is an interface which encapsulates a group of views. It routes requests to the appropriate view.
// It routes events to the appropriate view. It also provides a way to register views.
type Controller interface {
	Route(opts ...RouteOption) http.HandlerFunc
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
	pubsub               PubsubAdapter
	appName              string
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
func WithPubsubAdapter(pubsub PubsubAdapter) ControllerOption {
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
			o.watchExts = extensions
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

	o := &opt{
		channelFunc:       DefaultChannelFunc,
		websocketUpgrader: websocket.Upgrader{EnableCompression: true},
		watchExts:         DefaultWatchExtensions,
		pubsub:            NewPubsubInmem(),
		appName:           name,
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

// Handler returns an http.HandlerFunc that handles the view.
func (c *controller) Route(opts ...RouteOption) http.HandlerFunc {
	defaultRouteOpt := &routeOpt{
		layoutContentName: "content",
		partials:          []string{"./templates/partials"},
		funcMap:           DefaultFuncMap(),
		extensions:        []string{".gohtml", ".gotmpl", ".html", ".tmpl"},
		eventSender:       make(chan Event),
		onLoad: func(event Event, render RouteRenderer) error {
			return nil
		},
	}
	for _, option := range opts {
		option(defaultRouteOpt)
	}

	rt := newRoute(c, defaultRouteOpt)
	return func(w http.ResponseWriter, r *http.Request) {
		rt.handle(w, r)
	}
}
