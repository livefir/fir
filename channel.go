package fir

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

// UserIDKey is the key for the user id in the request context. It is used in the default channel function.
const UserIDKey = "key_user_id"

func DefaultChannelFunc(r *http.Request, viewID string) *string {
	if viewID == "" {
		viewID = "root"
		if r.URL.Path != "/" {
			viewID = strings.Replace(r.URL.Path, "/", "_", -1)
		}
	}

	userID, ok := r.Context().Value(UserIDKey).(string)
	if !ok || userID == "" {
		log.Printf("warning: no user id in request context. user is anonymous, viewID: %s \n", viewID)
		userID = "anonymous"
	}
	channel := fmt.Sprintf("%s:%s", userID, viewID)
	return &channel
}
