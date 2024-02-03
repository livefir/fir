package fir

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/livefir/fir/internal/logger"
)

func (cntrl *controller) defaultChannelFunc(r *http.Request, viewID string) *string {
	if viewID == "" {
		viewID = "root"
		if r.URL.Path != "/" {
			viewID = strings.Replace(r.URL.Path, "/", "_", -1)
		}
	}

	userID, _ := r.Context().Value(UserKey).(string)
	if userID == "" {
		cookie, err := r.Cookie(cntrl.opt.cookieName)
		if err != nil {
			logger.Errorf("decode session err: %v, can't join channel", err)
			return nil
		}
		sessionID, _, err := decodeSession(*cntrl.opt.secureCookie, cntrl.opt.cookieName, cookie.Value)
		if err != nil {
			logger.Errorf("decode session err: %v, can't join channel", err)
			return nil
		}
		userID = sessionID
	}
	channel := fmt.Sprintf("%s:%s", userID, viewID)
	return &channel
}
