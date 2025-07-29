package session

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/securecookie"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeSession(t *testing.T) {
	// Create a secure cookie for testing
	hashKey := []byte("very-secret-hash-key-32-bytes!!!")
	blockKey := []byte("very-secret-block-key-32-bytes!!")
	sc := securecookie.New(hashKey, blockKey)

	t.Run("valid session", func(t *testing.T) {
		cookieName := "test-session"
		sessionData := "session-123:route-456"

		// Encode the session
		encodedValue, err := sc.Encode(cookieName, sessionData)
		require.NoError(t, err)

		// Decode the session
		sessionID, routeID, err := DecodeSession(*sc, cookieName, encodedValue)

		assert.NoError(t, err)
		assert.Equal(t, "session-123", sessionID)
		assert.Equal(t, "route-456", routeID)
	})

	t.Run("invalid cookie format", func(t *testing.T) {
		cookieName := "test-session"
		invalidCookie := "invalid-cookie-value"

		sessionID, routeID, err := DecodeSession(*sc, cookieName, invalidCookie)

		assert.Error(t, err)
		assert.Empty(t, sessionID)
		assert.Empty(t, routeID)
	})

	t.Run("empty session data", func(t *testing.T) {
		cookieName := "test-session"

		// Encode empty session
		encodedValue, err := sc.Encode(cookieName, "")
		require.NoError(t, err)

		sessionID, routeID, err := DecodeSession(*sc, cookieName, encodedValue)

		assert.Equal(t, ErrEmptySession, err)
		assert.Empty(t, sessionID)
		assert.Empty(t, routeID)
	})

	t.Run("invalid session format - missing colon", func(t *testing.T) {
		cookieName := "test-session"
		sessionData := "session-without-colon"

		// Encode invalid session format
		encodedValue, err := sc.Encode(cookieName, sessionData)
		require.NoError(t, err)

		sessionID, routeID, err := DecodeSession(*sc, cookieName, encodedValue)

		assert.Equal(t, ErrInvalidSession, err)
		assert.Empty(t, sessionID)
		assert.Empty(t, routeID)
	})

	t.Run("invalid session format - too many parts", func(t *testing.T) {
		cookieName := "test-session"
		sessionData := "session:route:extra"

		// Encode invalid session format
		encodedValue, err := sc.Encode(cookieName, sessionData)
		require.NoError(t, err)

		sessionID, routeID, err := DecodeSession(*sc, cookieName, encodedValue)

		assert.Equal(t, ErrInvalidSession, err)
		assert.Empty(t, sessionID)
		assert.Empty(t, routeID)
	})

	t.Run("tampered cookie", func(t *testing.T) {
		cookieName := "test-session"

		// Create a properly encoded cookie and then tamper with it
		sessionData := "session-123:route-456"
		encodedValue, err := sc.Encode(cookieName, sessionData)
		require.NoError(t, err)

		// Tamper with the cookie by changing the last character
		tamperedCookie := encodedValue[:len(encodedValue)-1] + "X"

		sessionID, routeID, err := DecodeSession(*sc, cookieName, tamperedCookie)

		assert.Error(t, err)
		assert.Empty(t, sessionID)
		assert.Empty(t, routeID)
	})
}

func TestEncodeSession(t *testing.T) {
	// Create a secure cookie for testing
	hashKey := []byte("very-secret-hash-key-32-bytes!!!")
	blockKey := []byte("very-secret-block-key-32-bytes!!")
	sc := securecookie.New(hashKey, blockKey)

	t.Run("new session creation", func(t *testing.T) {
		// Create a new HTTP request without any cookies
		req := httptest.NewRequest("GET", "/", nil)
		recorder := httptest.NewRecorder()

		opt := RouteOpt{
			SecureCookie: sc,
			CookieName:   "test-session",
			ID:           "route-123",
		}

		err := EncodeSession(opt, recorder, req)
		assert.NoError(t, err)

		// Check that a cookie was set
		cookies := recorder.Result().Cookies()
		require.Len(t, cookies, 1)

		cookie := cookies[0]
		assert.Equal(t, "test-session", cookie.Name)
		assert.Equal(t, "/", cookie.Path)
		assert.Equal(t, 0, cookie.MaxAge)
		assert.NotEmpty(t, cookie.Value)

		// Verify we can decode the cookie
		sessionID, routeID, err := DecodeSession(*sc, opt.CookieName, cookie.Value)
		assert.NoError(t, err)
		assert.NotEmpty(t, sessionID)
		assert.Equal(t, "route-123", routeID)
	})

	t.Run("existing valid session", func(t *testing.T) {
		// Create an existing session
		existingSessionID := "existing-session-id"
		sessionData := existingSessionID + ":old-route"
		encodedValue, err := sc.Encode("test-session", sessionData)
		require.NoError(t, err)

		// Create request with existing cookie
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{
			Name:  "test-session",
			Value: encodedValue,
		})
		recorder := httptest.NewRecorder()

		opt := RouteOpt{
			SecureCookie: sc,
			CookieName:   "test-session",
			ID:           "new-route-456",
		}

		err = EncodeSession(opt, recorder, req)
		assert.NoError(t, err)

		// Check that a cookie was set
		cookies := recorder.Result().Cookies()
		require.Len(t, cookies, 1)

		cookie := cookies[0]
		assert.Equal(t, "test-session", cookie.Name)

		// Verify the session ID is preserved but route ID is updated
		sessionID, routeID, err := DecodeSession(*sc, opt.CookieName, cookie.Value)
		assert.NoError(t, err)
		assert.Equal(t, existingSessionID, sessionID)
		assert.Equal(t, "new-route-456", routeID)
	})

	t.Run("existing invalid session", func(t *testing.T) {
		// Create request with invalid cookie
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{
			Name:  "test-session",
			Value: "invalid-cookie-value",
		})
		recorder := httptest.NewRecorder()

		opt := RouteOpt{
			SecureCookie: sc,
			CookieName:   "test-session",
			ID:           "route-789",
		}

		err := EncodeSession(opt, recorder, req)
		assert.NoError(t, err)

		// Check that a new cookie was set
		cookies := recorder.Result().Cookies()
		require.Len(t, cookies, 1)

		cookie := cookies[0]
		assert.Equal(t, "test-session", cookie.Name)

		// Verify a new session was created
		sessionID, routeID, err := DecodeSession(*sc, opt.CookieName, cookie.Value)
		assert.NoError(t, err)
		assert.NotEmpty(t, sessionID)
		assert.Equal(t, "route-789", routeID)
	})

	t.Run("missing secure cookie", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		recorder := httptest.NewRecorder()

		opt := RouteOpt{
			SecureCookie: nil, // Missing secure cookie
			CookieName:   "test-session",
			ID:           "route-123",
		}

		// The function should panic with nil SecureCookie
		assert.Panics(t, func() {
			EncodeSession(opt, recorder, req)
		})
	})

	t.Run("empty route ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		recorder := httptest.NewRecorder()

		opt := RouteOpt{
			SecureCookie: sc,
			CookieName:   "test-session",
			ID:           "", // Empty route ID
		}

		err := EncodeSession(opt, recorder, req)
		assert.NoError(t, err)

		cookies := recorder.Result().Cookies()
		require.Len(t, cookies, 1)

		// Verify the session was created with empty route ID
		sessionID, routeID, err := DecodeSession(*sc, opt.CookieName, cookies[0].Value)
		assert.NoError(t, err)
		assert.NotEmpty(t, sessionID)
		assert.Equal(t, "", routeID)
	})
}

func TestSessionErrors(t *testing.T) {
	t.Run("error constants", func(t *testing.T) {
		assert.Equal(t, "invalid session", ErrInvalidSession.Error())
		assert.Equal(t, "empty session", ErrEmptySession.Error())
	})
}

func TestRouteOpt(t *testing.T) {
	t.Run("route opt creation", func(t *testing.T) {
		hashKey := []byte("very-secret-hash-key-32-bytes!!!")
		blockKey := []byte("very-secret-block-key-32-bytes!!")
		sc := securecookie.New(hashKey, blockKey)

		opt := RouteOpt{
			SecureCookie: sc,
			CookieName:   "my-session",
			ID:           "my-route",
		}

		assert.NotNil(t, opt.SecureCookie)
		assert.Equal(t, "my-session", opt.CookieName)
		assert.Equal(t, "my-route", opt.ID)
	})
}
