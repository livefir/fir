package fir

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/securecookie"
)

var errInvalidSession = errors.New("invalid session")
var errEmptySession = errors.New("empty session")

func decodeSession(sc securecookie.SecureCookie, cookieName, cookieValue string) (string, string, error) {
	var session string
	if err := sc.Decode(cookieName, cookieValue, &session); err != nil {
		return "", "", err
	}
	if session == "" {
		return "", "", errEmptySession
	}

	parts := strings.Split(session, ":")
	if len(parts) != 2 {
		return "", "", errInvalidSession
	}
	return parts[0], parts[1], nil
}

func encodeSession(opt routeOpt, w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie(opt.cookieName)
	if err != nil {
		return err
	}
	sessionID, _, _ := decodeSession(*opt.secureCookie, opt.cookieName, cookie.Value)
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	session := sessionID + ":" + opt.id

	encodedSessionID, err := opt.secureCookie.Encode(opt.cookieName, session)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:   opt.cookieName,
		Value:  encodedSessionID,
		MaxAge: 0,
		Path:   "/",
	})
	return nil
}
