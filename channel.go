package fir

import (
	"fmt"
	"net/http"
	"strings"

	"k8s.io/klog/v2"
)

func defaultChannelFunc(r *http.Request, viewID string) *string {
	if viewID == "" {
		viewID = "root"
		if r.URL.Path != "/" {
			viewID = strings.Replace(r.URL.Path, "/", "_", -1)
		}
	}

	userID, ok := r.Context().Value(UserKey).(string)
	if !ok || userID == "" {
		klog.Warningf("warning: no user id in request context. user is anonymous, viewID: %s \n", viewID)
		userID = "anonymous"
	}
	channel := fmt.Sprintf("%s:%s", userID, viewID)
	return &channel
}
