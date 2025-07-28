package session

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/securecookie"
)

var ErrInvalidSession = errors.New("invalid session")
var ErrEmptySession = errors.New("empty session")

// RouteOpt represents the configuration needed for session management
type RouteOpt struct {
	SecureCookie *securecookie.SecureCookie
	CookieName   string
	ID           string
}

func DecodeSession(sc securecookie.SecureCookie, cookieName, cookieValue string) (string, string, error) {
	var session string

	if err := sc.Decode(cookieName, cookieValue, &session); err != nil {
		return "", "", err
	}

	if session == "" {
		return "", "", ErrEmptySession
	}

	parts := strings.Split(session, ":")
	if len(parts) != 2 {
		return "", "", ErrInvalidSession
	}
	return parts[0], parts[1], nil
}

func EncodeSession(opt RouteOpt, w http.ResponseWriter, r *http.Request) error {
	var sessionID string
	cookie, err := r.Cookie(opt.CookieName)
	if err == nil && cookie != nil {
		sessionID, _, _ = DecodeSession(*opt.SecureCookie, opt.CookieName, cookie.Value)
		if sessionID == "" {
			sessionID = uuid.New().String()
		}
	}
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	session := sessionID + ":" + opt.ID

	encodedSessionID, err := opt.SecureCookie.Encode(opt.CookieName, session)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:   opt.CookieName,
		Value:  encodedSessionID,
		MaxAge: 0,
		Path:   "/",
	})
	return nil
}
