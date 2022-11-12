package fir

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

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

type Option func(*opt)

func WithChannel(f func(r *http.Request, viewID string) *string) Option {
	return func(o *opt) {
		o.channelFunc = f
	}
}

func WithUpgrader(upgrader websocket.Upgrader) Option {
	return func(o *opt) {
		o.websocketUpgrader = upgrader
	}
}

func WithErrorView(view View) Option {
	return func(o *opt) {
		o.errorView = view
	}
}

func WithEmbedFS(fs embed.FS) Option {
	return func(o *opt) {
		o.embedFS = fs
		o.hasEmbedFS = true
	}
}

func DisableTemplateCache() Option {
	return func(o *opt) {
		o.disableTemplateCache = true
	}
}

func EnableDebugLog() Option {
	return func(o *opt) {
		o.debugLog = true
	}
}

func EnableWatch(rootDir string, extensions ...string) Option {
	return func(o *opt) {
		o.enableWatch = true
		if len(extensions) > 0 {
			o.publicDir = rootDir
			o.watchExts = extensions
		}
	}
}

func DevelopmentMode(enable bool) Option {
	return func(o *opt) {
		o.developmentMode = enable
	}
}

// PublicDir is the path to directory containing the public html template files.
func PublicDir(path string) Option {
	return func(o *opt) {
		o.publicDir = path
	}
}

func NewController(name string, options ...Option) Controller {
	if name == "" {
		panic("controller name is required")
	}

	o := &opt{
		channelFunc: func(r *http.Request, viewID string) *string {
			if viewID == "" {
				viewID = "root"
				if r.URL.Path != "/" {
					viewID = strings.Replace(r.URL.Path, "/", "_", -1)
				}
			}

			userID, ok := r.Context().Value(UserIDKey).(string)
			if !ok || userID == "" {
				log.Printf("[view] warning: no user id in request context\n")
				userID = "anonymous"
			}
			channel := fmt.Sprintf("%s:%s", userID, viewID)

			log.Println("client subscribed to channel: ", channel)
			return &channel
		},
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

func (c *controller) Handler(view View) http.HandlerFunc {
	viewTemplate, err := parseTemplate(c.opt, view)
	if err != nil {
		panic(err)
	}

	errorViewTemplate, err := parseTemplate(c.opt, c.errorView)
	if err != nil {
		panic(err)
	}

	mountData := make(Data)
	return func(w http.ResponseWriter, r *http.Request) {
		v := &viewHandler{
			view:              view,
			errorView:         c.errorView,
			viewTemplate:      viewTemplate,
			errorViewTemplate: errorViewTemplate,
			mountData:         mountData,
			cntrl:             c,
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
