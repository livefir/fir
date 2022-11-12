package fir

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

func DefaultChannelFunc(r *http.Request, viewID string) *string {
	if viewID == "" {
		viewID = "root"
		if r.URL.Path != "/" {
			viewID = strings.Replace(r.URL.Path, "/", "_", -1)
		}
	}

	userID, ok := r.Context().Value(UserIDKey).(string)
	if !ok || userID == "" {
		log.Printf("[view] warning: no user id in request context. user is anonymous\n")
		userID = "anonymous"
	}
	channel := fmt.Sprintf("%s:%s", userID, viewID)

	log.Println("client subscribed to channel: ", channel)
	return &channel
}
