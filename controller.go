package fir

import (
	"flag"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/securecookie"

	"github.com/gorilla/sessions"

	"github.com/gorilla/websocket"
)

type Controller interface {
	Handler(view View) http.HandlerFunc
}

type controlOpt struct {
	subscribeTopicFunc func(r *http.Request) *string
	upgrader           websocket.Upgrader

	disableTemplateCache bool
	debugLog             bool
	enableWatch          bool
	watchExts            []string
	projectRoot          string
	developmentMode      bool
	errorView            View
	cookieStore          *sessions.CookieStore
}

type Option func(*controlOpt)

func WithSubscribeTopic(f func(r *http.Request) *string) Option {
	return func(o *controlOpt) {
		o.subscribeTopicFunc = f
	}
}

func WithUpgrader(upgrader websocket.Upgrader) Option {
	return func(o *controlOpt) {
		o.upgrader = upgrader
	}
}

func WithErrorView(view View) Option {
	return func(o *controlOpt) {
		o.errorView = view
	}
}

func WithCookieStore(cookieStore *sessions.CookieStore) Option {
	return func(o *controlOpt) {
		o.cookieStore = cookieStore
	}
}

func DisableTemplateCache() Option {
	return func(o *controlOpt) {
		o.disableTemplateCache = true
	}
}

func EnableDebugLog() Option {
	return func(o *controlOpt) {
		o.debugLog = true
	}
}

func EnableWatch(rootDir string, extensions ...string) Option {
	return func(o *controlOpt) {
		o.enableWatch = true
		if len(extensions) > 0 {
			o.projectRoot = rootDir
			o.watchExts = extensions
		}
	}
}

func DevelopmentMode(enable bool) Option {
	return func(o *controlOpt) {
		o.developmentMode = enable
	}
}

// ProjectRoot is for reloading template files on file change during development
func ProjectRoot(projectRoot string) Option {
	return func(o *controlOpt) {
		o.projectRoot = projectRoot
	}
}

func NewController(name string, options ...Option) Controller {
	if name == "" {
		panic("controller name is required")
	}

	o := &controlOpt{
		subscribeTopicFunc: func(r *http.Request) *string {
			topic := "root"
			if r.URL.Path != "/" {
				topic = strings.Replace(r.URL.Path, "/", "_", -1)
			}

			log.Println("client subscribed to topic: ", topic)
			return &topic
		},
		upgrader:    websocket.Upgrader{EnableCompression: true},
		watchExts:   DefaultWatchExtensions,
		errorView:   &DefaultErrorView{},
		cookieStore: sessions.NewCookieStore(securecookie.GenerateRandomKey(32)),
	}

	for _, option := range options {
		option(o)
	}

	if o.projectRoot == "" {
		var projectRoot string
		projectRootUsage := "project root directory that contains the template files."
		flag.StringVar(&projectRoot, "project", ".", projectRootUsage)
		flag.StringVar(&projectRoot, "p", ".", projectRootUsage+" (shortand)")
		flag.Parse()
		o.projectRoot = projectRoot
	}

	wc := &websocketController{
		cookieStore:      o.cookieStore,
		topicConnections: make(map[string]map[string]*websocket.Conn),
		controlOpt:       *o,
		name:             name,
	}
	if wc.developmentMode {
		log.Println("controller starting in developer mode ...", wc.developmentMode)
		wc.debugLog = true
		wc.enableWatch = true
		wc.disableTemplateCache = true
	}

	if wc.enableWatch {
		go watchTemplates(wc)
	}
	return wc
}

type userCount struct {
	n int
	sync.RWMutex
}

func (u *userCount) incr() int {
	u.Lock()
	defer u.Unlock()
	u.n = u.n + 1
	return u.n
}

type websocketController struct {
	name      string
	userCount userCount
	controlOpt
	cookieStore      *sessions.CookieStore
	topicConnections map[string]map[string]*websocket.Conn
	sync.RWMutex
}

func (wc *websocketController) addConnection(topic, connID string, sess *websocket.Conn) (created bool) {
	wc.Lock()
	defer wc.Unlock()
	_, ok := wc.topicConnections[topic]
	if !ok {
		// topic doesn't exit. create
		wc.topicConnections[topic] = make(map[string]*websocket.Conn)
		created = true
		log.Println("topic created", topic)
	}
	wc.topicConnections[topic][connID] = sess
	log.Println("addConnection", topic, connID, len(wc.topicConnections[topic]))
	return
}

func (wc *websocketController) removeConnection(topic, connID string) (destroyed bool) {
	wc.Lock()
	defer wc.Unlock()
	connMap, ok := wc.topicConnections[topic]
	if !ok {
		return
	}
	// delete connection from topic
	conn, ok := connMap[connID]
	if ok {
		delete(connMap, connID)
		conn.Close()
	}
	// no connections for the topic, remove it
	if len(connMap) == 0 {
		delete(wc.topicConnections, topic)
		destroyed = true
		log.Println("topic destroyed", topic)
	}

	log.Println("removeConnection", topic, connID, len(wc.topicConnections[topic]))
	return
}

func (wc *websocketController) message(topic string, message []byte) {
	wc.Lock()
	defer wc.Unlock()
	preparedMessage, err := websocket.NewPreparedMessage(websocket.TextMessage, message)
	if err != nil {
		log.Printf("err preparing message %v\n", err)
		return
	}

	conns, ok := wc.topicConnections[topic]
	if !ok {
		log.Printf("warn: topic %v doesn't exist\n", topic)
		return
	}

	for connID, conn := range conns {
		err := conn.WritePreparedMessage(preparedMessage)
		if err != nil {
			log.Printf("error: writing message for topic:%v, closing conn %s with err %v", topic, connID, err)
			conn.Close()
			continue
		}
	}
}

func (wc *websocketController) writeJSON(topic string, v any) {
	wc.Lock()
	defer wc.Unlock()

	conns, ok := wc.topicConnections[topic]
	if !ok {
		log.Printf("warn: topic %v doesn't exist\n", topic)
		return
	}

	for connID, conn := range conns {
		err := conn.WriteJSON(v)
		if err != nil {
			log.Printf("error: writing message for topic:%v, closing conn %s with err %v", topic, connID, err)
			conn.Close()
			continue
		}
	}
}

func (wc *websocketController) writeJSONAll(v any) {
	wc.Lock()
	defer wc.Unlock()

	for _, cm := range wc.topicConnections {
		for connID, conn := range cm {
			err := conn.WriteJSON(v)
			if err != nil {
				log.Printf("error: writing json message. closing conn %s with err %v", connID, err)
				conn.Close()
				continue
			}
		}
	}
}

func (wc *websocketController) messageAll(message []byte) {
	wc.Lock()
	defer wc.Unlock()
	preparedMessage, err := websocket.NewPreparedMessage(websocket.TextMessage, message)
	if err != nil {
		log.Printf("err preparing message %v\n", err)
		return
	}

	for _, cm := range wc.topicConnections {
		for connID, conn := range cm {
			err := conn.WritePreparedMessage(preparedMessage)
			if err != nil {
				log.Printf("error: writing message %v, closing conn %s with err %v", message, connID, err)
				conn.Close()
				continue
			}
		}
	}
}

func (wc *websocketController) getUser(w http.ResponseWriter, r *http.Request) (int, error) {
	wc.cookieStore.MaxAge(0)
	cookieSession, _ := wc.cookieStore.Get(r, "_fir_session_")

	user := cookieSession.Values["user"]
	if user == nil {
		c := wc.userCount.incr()
		cookieSession.Values["user"] = c
		user = c
	}
	err := cookieSession.Save(r, w)
	if err != nil {
		log.Printf("getUser err %v\n", err)
		return -1, err
	}

	return user.(int), nil
}

func (wc *websocketController) Handler(view View) http.HandlerFunc {
	viewTemplate, err := parseTemplate(wc.projectRoot, view)
	if err != nil {
		panic(err)
	}

	errorViewTemplate, err := parseTemplate(wc.projectRoot, wc.errorView)
	if err != nil {
		panic(err)
	}

	mountData := make(Data)
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := wc.getUser(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		v := &viewHandler{
			view:              view,
			errorView:         wc.errorView,
			viewTemplate:      viewTemplate,
			errorViewTemplate: errorViewTemplate,
			mountData:         mountData,
			wc:                wc,
			user:              user,
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
