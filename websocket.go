package fir

import (
	"bytes"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/livefir/fir/internal/logger"
	"github.com/minio/sha256-simd"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 20 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 55 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 1024

	EventSocketConnected    = "fir_socket_connected"
	EventSocketDisconnected = "fir_socket_disconnected"
)

type SocketStatus struct {
	Connected bool
	User      string
}

// RedirectUnauthorisedWebSocket sends a 4001 close message to the client
// It sends the redirect url in the close message payload
// If the request is not a websocket request or has error upgrading and writing the close message, it returns false
// redirect url must be less than 123 bytes
func RedirectUnauthorisedWebSocket(w http.ResponseWriter, r *http.Request, redirect string) bool {
	if len(redirect) > 123 {
		panic("redirect url is too long: max size 123 bytes")
	}
	if !websocket.IsWebSocketUpgrade(r) {
		return false
	}

	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Errorf("upgrade err: %v", err)
		return false
	}

	if logger.GetGlobalLogger().IsDebugEnabled() {
		logger.GetGlobalLogger().Debug("websocket connection redirect",
			"remote_addr", r.RemoteAddr,
			"redirect_url", redirect,
		)
	}

	err = conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(4001, redirect), time.Now().Add(writeWait))
	if err != nil {
		logger.Errorf("write control err: %v", err)
		return false
	}
	defer conn.Close()

	return true
}

func onWebsocket(w http.ResponseWriter, r *http.Request, cntrl *controller) {
	startTime := time.Now()

	if logger.GetGlobalLogger().IsDebugEnabled() {
		logger.GetGlobalLogger().Debug("websocket connection attempt",
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
			"timestamp", startTime.Format(time.RFC3339),
		)
	}

	// Create new connection
	conn, err := NewConnection(w, r, cntrl)
	if err != nil {
		return
	}
	defer conn.Close()

	// Upgrade to WebSocket
	if err := conn.Upgrade(); err != nil {
		return
	}

	if logger.GetGlobalLogger().IsDebugEnabled() {
		logger.GetGlobalLogger().Debug("websocket connection established",
			"remote_addr", r.RemoteAddr,
			"session_id", conn.sessionID,
			"upgrade_duration_ms", time.Since(startTime).Milliseconds(),
		)
	}

	// Start pubsub listeners
	if err := conn.StartPubSubListeners(); err != nil {
		return
	}

	// Send connected event
	conn.SendConnectedEvent()

	logger.GetGlobalLogger().Info("websocket connected",
		"remote_addr", r.RemoteAddr,
		"session_id", conn.sessionID,
	)

	// Start write pump
	conn.StartWritePump()

	// Start read loop (blocks until connection closes)
	conn.ReadLoop()

	if logger.GetGlobalLogger().IsDebugEnabled() {
		logger.GetGlobalLogger().Debug("websocket connection closed",
			"remote_addr", r.RemoteAddr,
			"session_id", conn.sessionID,
			"total_duration_ms", time.Since(startTime).Milliseconds(),
		)
	}
}

func eqBytesHash(a, b []byte) bool {
	w := sha256.New()
	w.Write(a)
	aHash := w.Sum(nil)
	w.Reset()
	w.Write(b)
	bHash := w.Sum(nil)
	return bytes.Equal(aHash, bHash)
}
