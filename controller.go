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
	Handler(view View) http.HandlerFunc
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
	errorView            View
	embedFS              embed.FS
	hasEmbedFS           bool
	pubsub               PubsubAdapter
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

// WithErrorView is an option to set a view to render error messages.
func WithErrorView(view View) ControllerOption {
	return func(o *opt) {
		o.errorView = view
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
		errorView:         &DefaultErrorView{},
		pubsub:            NewPubsubInmem(),
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
func (c *controller) Handler(view View) http.HandlerFunc {
	viewTemplate, err := parseTemplate(c.opt, view)
	if err != nil {
		panic(err)
	}

	errorViewTemplate, err := parseTemplate(c.opt, c.errorView)
	if err != nil {
		panic(err)
	}

	// non-blocking send even if there are no live connections(ws, sse) for this view in the current server instance.
	// this is to ensure that sending to the stream is non-blocking. since channel can only be constructed
	// within the scope of a live connection, publishing patch messages are only possible when there is a live connection.
	// TODO: explain this better
	streamCh := make(chan Patchset)
	go func() {
		for patch := range view.Publisher() {
			streamCh <- patch
		}
	}()

	mountData := make(map[string]any)
	return func(w http.ResponseWriter, r *http.Request) {
		v := &viewHandler{
			view:              view,
			errorView:         c.errorView,
			viewTemplate:      viewTemplate,
			errorViewTemplate: errorViewTemplate,
			mountData:         mountData,
			cntrl:             c,
			streamCh:          streamCh,
		}
		if r.Header.Get("X-FIR-MODE") == "event" && r.Method == "POST" {
			onPatchEvent(w, r, v)
		} else if r.Header.Get("Connection") == "Upgrade" &&
			r.Header.Get("Upgrade") == "websocket" {
			onWebsocket(w, r, v)
		} else {
			onRequest(w, r, v)
		}
	}
}
