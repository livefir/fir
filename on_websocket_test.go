package fir

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/websocket"
	"github.com/livefir/fir/pubsub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test onWebsocket function error paths to improve coverage

func Test_onWebsocket_NoCookie(t *testing.T) {
	// Test when no cookie is present - should trigger RedirectUnauthorisedWebSocket
	cntrl := createSimpleController()

	req := httptest.NewRequest("GET", "/ws", nil)
	resp := httptest.NewRecorder()

	require.NotPanics(t, func() {
		onWebsocket(resp, req, cntrl)
	})
}

func Test_onWebsocket_EmptyCookieValue(t *testing.T) {
	// Test when cookie exists but has empty value
	cntrl := createSimpleController()

	req := httptest.NewRequest("GET", "/ws", nil)
	req.AddCookie(&http.Cookie{
		Name:  cntrl.cookieName,
		Value: "",
	})
	resp := httptest.NewRecorder()

	require.NotPanics(t, func() {
		onWebsocket(resp, req, cntrl)
	})
}

func Test_onWebsocket_InvalidCookieData(t *testing.T) {
	// Test when cookie has invalid/malformed data - should trigger decode error
	cntrl := createSimpleController()

	req := httptest.NewRequest("GET", "/ws", nil)
	req.AddCookie(&http.Cookie{
		Name:  cntrl.cookieName,
		Value: "invalid-data",
	})
	resp := httptest.NewRecorder()

	require.NotPanics(t, func() {
		onWebsocket(resp, req, cntrl)
	})
}

func Test_onWebsocket_EmptySessionID(t *testing.T) {
	// Test when sessionID is empty after decode
	cntrl := createSimpleController()

	// Create a session with empty sessionID (format: "sessionID:routeID")
	sessionString := ":test-route" // Empty sessionID
	cookieValue, err := cntrl.secureCookie.Encode(cntrl.cookieName, sessionString)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/ws", nil)
	req.AddCookie(&http.Cookie{
		Name:  cntrl.cookieName,
		Value: cookieValue,
	})
	resp := httptest.NewRecorder()

	require.NotPanics(t, func() {
		onWebsocket(resp, req, cntrl)
	})
}

func Test_onWebsocket_EmptyRouteID(t *testing.T) {
	// Test when routeID is empty after decode
	cntrl := createSimpleController()

	// Create a session with empty routeID (format: "sessionID:routeID")
	sessionString := "test-session:" // Empty routeID
	cookieValue, err := cntrl.secureCookie.Encode(cntrl.cookieName, sessionString)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/ws", nil)
	req.AddCookie(&http.Cookie{
		Name:  cntrl.cookieName,
		Value: cookieValue,
	})
	resp := httptest.NewRecorder()

	require.NotPanics(t, func() {
		onWebsocket(resp, req, cntrl)
	})
}

func Test_onWebsocket_SocketConnectCallbackError(t *testing.T) {
	// Test when onSocketConnect callback returns an error
	cntrl := createSimpleController()

	// Set onSocketConnect to return an error - should cause early return
	cntrl.onSocketConnect = func(userOrSessionID string) error {
		return assert.AnError
	}

	// Create a valid session (format: "sessionID:routeID")
	sessionString := "test-session:test-route"
	cookieValue, err := cntrl.secureCookie.Encode(cntrl.cookieName, sessionString)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/ws", nil)
	req.AddCookie(&http.Cookie{
		Name:  cntrl.cookieName,
		Value: cookieValue,
	})
	resp := httptest.NewRecorder()

	require.NotPanics(t, func() {
		onWebsocket(resp, req, cntrl)
	})
}

// Helper function to create a minimal test controller
func createSimpleController() *controller {
	hashKey := securecookie.GenerateRandomKey(64)
	blockKey := securecookie.GenerateRandomKey(32)
	secureCookie := securecookie.New(hashKey, blockKey)

	return &controller{
		name:   "test",
		routes: make(map[string]*route),
		opt: opt{
			cookieName:        "test-session",
			secureCookie:      secureCookie,
			websocketUpgrader: websocket.Upgrader{},
			pubsub:            pubsub.NewInmem(),
			channelFunc: func(r *http.Request, viewID string) *string {
				ch := "test-channel"
				return &ch
			},
		},
	}
}
