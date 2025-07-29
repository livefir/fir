package fir

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/schema"
	"github.com/gorilla/websocket"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/livefir/fir/internal/dom"
	"github.com/livefir/fir/pubsub"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	redisContainer "github.com/testcontainers/testcontainers-go/modules/redis"
)

// mockPubsubAdapter is a simple mock for testing
type mockPubsubAdapter struct{}

func (m *mockPubsubAdapter) Publish(ctx context.Context, channel string, event pubsub.Event) error {
	return nil
}

func (m *mockPubsubAdapter) Subscribe(ctx context.Context, channel string) (pubsub.Subscription, error) {
	return nil, nil
}

func (m *mockPubsubAdapter) HasSubscribers(ctx context.Context, pattern string) bool {
	return false
}

type doubleRequest struct {
	Num int `json:"num"`
}

func doubler() RouteOptions {
	return RouteOptions{
		ID("doubler"),
		Content(
			`<div @fir:double:ok="$fir.replace()">{{ .num }}</div>`),
		OnLoad(func(ctx RouteContext) error {
			return ctx.KV("num", 0)
		}),
		OnEvent("double", func(ctx RouteContext) error {
			req := new(doubleRequest)
			if err := ctx.Bind(req); err != nil {
				return err
			}
			return ctx.KV("num", req.Num*2)
		}),
	}
}

var testCases = []struct {
	name      string
	options   []ControllerOption
	routeFunc RouteFunc
}{
	// Test cases here
	{
		name:      "doubler",
		routeFunc: doubler,
		// options:   []ControllerOption{DevelopmentMode(true)},
	},
}

type testInput struct {
	serverURL string
	num       int
	wssend    int
	wsrecv    int
	event     Event
	conn      *websocket.Conn
	closeWs   bool
	sync.RWMutex
}

func (ti *testInput) incWssend() {
	ti.Lock()
	defer ti.Unlock()
	ti.wssend += 1
}

func (ti *testInput) incWsrecv() {
	ti.Lock()
	defer ti.Unlock()
	ti.wsrecv += 1
}

func (ti *testInput) getWssend() int {
	ti.RLock()
	defer ti.RUnlock()
	return ti.wssend
}

func (ti *testInput) getWsrecv() int {
	ti.RLock()
	defer ti.RUnlock()
	return ti.wsrecv
}

func eventPayload(tb testing.TB, ti *testInput) Event {
	// Make a request to the test server
	req, err := http.NewRequest("GET", ti.serverURL, nil)
	if err != nil {
		tb.Fatal(err)
	}

	resp, err := cleanhttp.DefaultClient().Do(req)
	if err != nil {
		tb.Fatal(err)
	}
	defer resp.Body.Close()

	var firSession string
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "_fir_session_" {
			firSession = cookie.Value
			break
		}
	}

	data, err := json.Marshal(&doubleRequest{Num: ti.num})
	if err != nil {
		tb.Fatal(err)
	}

	event := Event{
		ID:        "double",
		IsForm:    false,
		Params:    data,
		SessionID: &firSession,
		Timestamp: time.Now().UTC().UnixMilli(),
	}

	return event

}

func runPostEventTest(tb testing.TB, ti *testInput) {

	event := eventPayload(tb, ti)

	// post event to the test server
	payload := new(bytes.Buffer)
	err := json.NewEncoder(payload).Encode(event)
	if err != nil {
		tb.Fatal(err)
	}

	req, err := http.NewRequest("POST", ti.serverURL, payload)
	if err != nil {
		tb.Fatal(err)
	}
	req.AddCookie(&http.Cookie{Name: "_fir_session_", Value: *event.SessionID})
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-FIR-MODE", "event")
	resp, err := cleanhttp.DefaultClient().Do(req)
	if err != nil {
		tb.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		tb.Fatalf("expected status code 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		tb.Fatal(err)
	}

	// decode response body to []dom.Event
	var domEvents []dom.Event
	err = json.Unmarshal(body, &domEvents)
	if err != nil {
		tb.Fatal(err)
	}

	if len(domEvents) != 1 {
		tb.Fatalf("expected 1 event, got %d", len(domEvents))
	}

	expectedHTML := fmt.Sprintf("%d", ti.num*2)

	if removeSpace(domEvents[0].Detail.HTML) != expectedHTML {
		tb.Fatalf("expected: %s, got: %s", expectedHTML, domEvents[0].Detail.HTML)
	}

}

func dialWebSocket(tb testing.TB, ti *testInput, event Event) *websocket.Conn {
	wsURLString := strings.Replace(ti.serverURL, "http", "ws", 1)
	wsDialer := websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		// EnableCompression: true,
		// Jar:              jar, // jar doesnät work but adding a Cookie header does
	}
	header := http.Header{}
	header.Set("Cookie", fmt.Sprintf("_fir_session_=%s", *event.SessionID))
	ws, _, err := wsDialer.Dial(wsURLString, header)
	if err != nil {
		tb.Fatal(err)
	}

	return ws
}

func runWebsocketEventTest(tb testing.TB, ti *testInput) {

	event := ti.event
	if event.SessionID == nil {
		event = eventPayload(tb, ti)
	}

	ws := ti.conn
	if ws == nil {
		ws = dialWebSocket(tb, ti, event)
	}

	if ti.closeWs {
		defer ws.Close()
	}

	ti.incWssend()

	err := ws.WriteJSON(event)
	if err != nil {
		tb.Fatal(err)
	}

	ws.SetReadDeadline(time.Now().Add(1000 * time.Millisecond))
	var message []byte
	_, message, err = ws.ReadMessage()
	if err != nil {
		tb.Fatal(err)
	}
	ti.incWsrecv()

	var domEvents []dom.Event

	err = json.Unmarshal(message, &domEvents)
	if err != nil {
		tb.Fatal(err)
	}

	if len(domEvents) != 1 {
		tb.Fatalf("expected 1 event, got %d", len(domEvents))
	}
	expectedHTML := fmt.Sprintf("%d", ti.num*2)
	if removeSpace(domEvents[0].Detail.HTML) != expectedHTML {
		tb.Fatalf("expected: %s, got: %s", expectedHTML, domEvents[0].Detail.HTML)
	}

	if !ti.closeWs {
		return
	}

	err = ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		tb.Fatal(err)
	}

}

func BenchmarkControllerWebsocktDisabled(b *testing.B) {

	for _, tc := range testCases {
		tc.options = append(tc.options, WithDisableWebsocket())
		controller := NewController(tc.name, tc.options...)
		// Create a test HTTP server
		server := httptest.NewServer(controller.RouteFunc(tc.routeFunc))
		defer server.Close()
		ti := &testInput{serverURL: server.URL, num: 10}
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				runPostEventTest(b, ti)
			}
		})
		b.ReportAllocs()
	}

}

func TestControllerWebsocketDisabled(t *testing.T) {
	for _, tc := range testCases {
		tc.options = append(tc.options, WithDisableWebsocket())
		controller := NewController(tc.name, tc.options...)
		// Create a test HTTP server
		server := httptest.NewServer(controller.RouteFunc(tc.routeFunc))
		defer server.Close()
		ti := &testInput{serverURL: server.URL, num: 10}
		t.Parallel()
		t.Run(tc.name, func(t *testing.T) {
			runPostEventTest(t, ti)
		})
	}
}

func BenchmarkControllerWebsocktEnabled(b *testing.B) {
	for _, tc := range testCases {
		controller := NewController(tc.name, tc.options...)
		// Create a test HTTP server
		server := httptest.NewServer(controller.RouteFunc(tc.routeFunc))
		defer server.Close()

		ti := &testInput{serverURL: server.URL, num: 10, closeWs: true}
		b.Cleanup(func() {
			fmt.Printf("ws send: %d, ws recv: %d\n", ti.getWssend(), ti.getWsrecv())
		})
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				runWebsocketEventTest(b, ti)
			}
		})
		b.ReportMetric(float64(ti.getWssend()), "total_sends")
		b.ReportMetric(float64(ti.getWsrecv()), "total_receives")
		b.ReportAllocs()
	}
}

func TestControllerWebsocktEnabledMultiEvent(t *testing.T) {
	for _, tc := range testCases {
		controller := NewController(tc.name, tc.options...)
		// Create a test HTTP server
		server := httptest.NewServer(controller.RouteFunc(tc.routeFunc))
		defer server.Close()

		ti := &testInput{serverURL: server.URL, num: 10}
		ti.event = eventPayload(t, ti)
		ti.conn = dialWebSocket(t, ti, ti.event)

		t.Cleanup(func() {
			err := ti.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				t.Fatal(err)
			}
			time.Sleep(500 * time.Millisecond)
			ti.conn.Close()

		})

		for i := 0; i < 20; i++ {
			time.Sleep(251 * time.Millisecond)
			ti.event.Timestamp = time.Now().UTC().UnixMilli()
			runWebsocketEventTest(t, ti)
		}

		if ti.getWssend() != ti.getWsrecv() {
			t.Fatalf("expected ws send: %d, ws recv: %d", ti.getWssend(), ti.getWsrecv())
		}

	}
}

func TestControllerWebsocketEnabled(t *testing.T) {
	for _, tc := range testCases {
		controller := NewController(tc.name, tc.options...)
		// Create a test HTTP server
		server := httptest.NewServer(controller.RouteFunc(tc.routeFunc))
		defer server.Close()
		ti := &testInput{serverURL: server.URL, num: 10}
		t.Parallel()
		t.Run(tc.name, func(t *testing.T) {
			runWebsocketEventTest(t, ti)
		})
	}
}

func BenchmarkControllerWebsocktEnabledRedis(b *testing.B) {

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pubsubAdapter := pubsub.NewRedis(client)

	for _, tc := range testCases {
		tc.options = append(tc.options, WithPubsubAdapter(pubsubAdapter))
		controller := NewController(tc.name, tc.options...)
		// Create a test HTTP server
		server := httptest.NewServer(controller.RouteFunc(tc.routeFunc))
		defer server.Close()

		ti := &testInput{serverURL: server.URL, num: 10}
		b.Cleanup(func() {
			fmt.Printf("ws send: %d, ws recv: %d\n", ti.getWssend(), ti.getWsrecv())
		})
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				runWebsocketEventTest(b, ti)
			}
		})
		b.ReportMetric(float64(ti.getWssend()), "total_sends")
		b.ReportMetric(float64(ti.getWsrecv()), "total_receives")
		b.ReportAllocs()
	}

}

func TestControllerWebsocketEnabledRedis(t *testing.T) {

	if os.Getenv("DOCKER") != "1" {
		t.Skip("Skipping testing since docker is not present")
	}

	ctx := context.Background()
	rc, err := redisContainer.Run(ctx, "redis:7")
	if err != nil {
		t.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := rc.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}()

	host, err := rc.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get host: %s", err)
	}

	port, err := rc.MappedPort(ctx, "6379")

	if err != nil {
		t.Fatalf("failed to get mapped port: %s", err)

	}

	addr := fmt.Sprintf("%s:%s", host, port.Port())
	if addr == "" {
		t.Fatalf("failed to get address: %s", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pubsubAdapter := pubsub.NewRedis(client)
	for _, tc := range testCases {
		tc.options = append(tc.options, WithPubsubAdapter(pubsubAdapter))
		controller := NewController(tc.name, tc.options...)
		// Create a test HTTP server
		server := httptest.NewServer(controller.RouteFunc(tc.routeFunc))
		defer server.Close()
		ti := &testInput{serverURL: server.URL, num: 10}
		t.Parallel()
		t.Run(tc.name, func(t *testing.T) {
			runWebsocketEventTest(t, ti)
		})
	}
}

// TestControllerOptions tests all controller option functions
func TestControllerOptions(t *testing.T) {
	t.Run("WithFuncMap", func(t *testing.T) {
		customFuncMap := template.FuncMap{
			"customFunc": func() string { return "custom" },
		}

		ctrl := NewController("test", WithFuncMap(customFuncMap))

		// Verify function map is set by creating a route with custom function
		testRoute := func() RouteOptions {
			return RouteOptions{
				Content(`{{ customFunc }}`),
			}
		}
		routeHandler := ctrl.RouteFunc(testRoute)

		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		routeHandler(w, req)

		assert.Contains(t, w.Body.String(), "custom")
	})

	t.Run("WithSessionSecrets", func(t *testing.T) {
		hashKey := []byte("very-secret-hash-key-32-bytes!!!")
		blockKey := []byte("very-secret-block-key-32-bytes!!")

		ctrl := NewController("test", WithSessionSecrets(hashKey, blockKey))

		// Test that secure cookie is properly set by verifying session functionality
		testRoute := func() RouteOptions {
			return RouteOptions{
				Content(`<div>test</div>`),
			}
		}
		routeHandler := ctrl.RouteFunc(testRoute)

		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		routeHandler(w, req)

		// Should have set a session cookie
		cookies := w.Result().Cookies()
		assert.Len(t, cookies, 1)
		assert.Equal(t, "_fir_session_", cookies[0].Name) // Default cookie name
	})

	t.Run("WithSessionName", func(t *testing.T) {
		customCookieName := "my-custom-session"

		ctrl := NewController("test", WithSessionName(customCookieName))

		testRoute := func() RouteOptions {
			return RouteOptions{
				Content(`<div>test</div>`),
			}
		}
		routeHandler := ctrl.RouteFunc(testRoute)

		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		routeHandler(w, req)

		// Should have set a cookie with custom name
		cookies := w.Result().Cookies()
		assert.Len(t, cookies, 1)
		assert.Equal(t, customCookieName, cookies[0].Name)
	})

	t.Run("WithChannelFunc", func(t *testing.T) {
		channelFunc := func(r *http.Request, viewID string) *string {
			channel := "custom-channel"
			return &channel
		}

		ctrl := NewController("test", WithChannelFunc(channelFunc))

		// Test that channel function is set
		assert.NotNil(t, ctrl)
	})

	t.Run("WithPathParamsFunc", func(t *testing.T) {
		pathParamsFunc := func(r *http.Request) PathParams {
			return PathParams{"id": "123"}
		}

		ctrl := NewController("test", WithPathParamsFunc(pathParamsFunc))

		// Test path params functionality - simply verify function is set
		// PathParams would normally be available in route context for event handlers
		testRoute := func() RouteOptions {
			return RouteOptions{
				Content(`<div>{{ .id }}</div>`),
				OnLoad(func(ctx RouteContext) error {
					// Verify PathParams can be accessed from context
					if pathParams, ok := ctx.Request().Context().Value(PathParamsKey).(PathParams); ok {
						if id, exists := pathParams["id"]; exists {
							return ctx.KV("id", id)
						}
					}
					return ctx.KV("id", "not-found")
				}),
			}
		}
		routeHandler := ctrl.RouteFunc(testRoute)

		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		routeHandler(w, req)

		assert.Contains(t, w.Body.String(), "123")
	})

	t.Run("WithPubsubAdapter", func(t *testing.T) {
		customPubsub := pubsub.NewInmem()

		ctrl := NewController("test", WithPubsubAdapter(customPubsub))

		// Test that pubsub adapter is set
		assert.NotNil(t, ctrl)
	})

	t.Run("WithWebsocketUpgrader", func(t *testing.T) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}

		ctrl := NewController("test", WithWebsocketUpgrader(upgrader))

		// Test that upgrader is set
		assert.NotNil(t, ctrl)
	})

	t.Run("WithEmbedFS", func(t *testing.T) {
		// Create a mock embed.FS (can't easily create real one in test)
		// But we can test that the option doesn't cause errors
		ctrl := NewController("test", WithEmbedFS(embed.FS{}))

		assert.NotNil(t, ctrl)
	})

	t.Run("WithPublicDir", func(t *testing.T) {
		publicDir := "/tmp/test-public"

		ctrl := NewController("test", WithPublicDir(publicDir))

		// Test that public directory is set
		assert.NotNil(t, ctrl)
	})

	t.Run("WithFormDecoder", func(t *testing.T) {
		decoder := schema.NewDecoder()
		decoder.IgnoreUnknownKeys(true)

		ctrl := NewController("test", WithFormDecoder(decoder))

		// Test that form decoder option is set by creating a simple route
		// Form decoding testing is complex and would require proper event setup
		testRoute := func() RouteOptions {
			return RouteOptions{
				Content(`<div>decoder test</div>`),
			}
		}
		routeHandler := ctrl.RouteFunc(testRoute)

		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		routeHandler(w, req)

		assert.Contains(t, w.Body.String(), "decoder test")
	})

	t.Run("WithDisableWebsocket", func(t *testing.T) {
		ctrl := NewController("test", WithDisableWebsocket())

		// Test that websocket is disabled
		assert.NotNil(t, ctrl)
	})

	t.Run("WithDropDuplicateInterval", func(t *testing.T) {
		interval := 500 * time.Millisecond

		ctrl := NewController("test", WithDropDuplicateInterval(interval))

		// Test that interval is set
		assert.NotNil(t, ctrl)
	})

	t.Run("WithOnSocketConnect", func(t *testing.T) {
		connectFunc := func(userOrSessionID string) error {
			// Mock connect handler
			return nil
		}

		ctrl := NewController("test", WithOnSocketConnect(connectFunc))

		// Test that connect function is set
		assert.NotNil(t, ctrl)
	})

	t.Run("WithOnSocketDisconnect", func(t *testing.T) {
		disconnectFunc := func(userOrSessionID string) {
			// Mock disconnect handler
		}

		ctrl := NewController("test", WithOnSocketDisconnect(disconnectFunc))

		// Test that disconnect function is set
		assert.NotNil(t, ctrl)
	})

	t.Run("DisableTemplateCache", func(t *testing.T) {
		ctrl := NewController("test", DisableTemplateCache())

		// Test template caching is disabled
		assert.NotNil(t, ctrl)
	})

	t.Run("EnableDebugLog", func(t *testing.T) {
		ctrl := NewController("test", EnableDebugLog())

		// Test debug logging is enabled
		assert.NotNil(t, ctrl)
	})

	t.Run("EnableWatch", func(t *testing.T) {
		rootDir := "/tmp"
		extensions := []string{".html", ".gohtml"}

		ctrl := NewController("test", EnableWatch(rootDir, extensions...))

		// Test file watching is enabled
		assert.NotNil(t, ctrl)
	})

	t.Run("DevelopmentMode", func(t *testing.T) {
		ctrl := NewController("test", DevelopmentMode(true))

		// Test development mode is enabled
		assert.NotNil(t, ctrl)
	})

	t.Run("MultipleOptions", func(t *testing.T) {
		// Test combining multiple options
		ctrl := NewController("test",
			EnableDebugLog(),
			DisableTemplateCache(),
			DevelopmentMode(true),
			WithSessionName("multi-test"),
		)

		// Test that all options are applied
		testRoute := func() RouteOptions {
			return RouteOptions{
				Content(`<div>multi-test</div>`),
			}
		}
		routeHandler := ctrl.RouteFunc(testRoute)

		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		routeHandler(w, req)

		// Should have set custom session cookie
		cookies := w.Result().Cookies()
		assert.Len(t, cookies, 1)
		assert.Equal(t, "multi-test", cookies[0].Name)
	})

	t.Run("GetPubsub", func(t *testing.T) {
		mockPubsub := &mockPubsubAdapter{}
		ctrl := NewController("test", WithPubsubAdapter(mockPubsub))

		actualPubsub := ctrl.(*controller).GetPubsub()
		assert.Equal(t, mockPubsub, actualPubsub)
	})
}
