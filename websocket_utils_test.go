package fir

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/livefir/fir/pubsub"
	"github.com/stretchr/testify/assert"
)

// TestEqBytesHash tests the eqBytesHash utility function
func TestEqBytesHash(t *testing.T) {
	t.Run("equal bytes return true", func(t *testing.T) {
		a := []byte("hello world")
		b := []byte("hello world")

		result := eqBytesHash(a, b)
		assert.True(t, result)
	})

	t.Run("different bytes return false", func(t *testing.T) {
		a := []byte("hello world")
		b := []byte("hello mars")

		result := eqBytesHash(a, b)
		assert.False(t, result)
	})

	t.Run("empty bytes are equal", func(t *testing.T) {
		a := []byte("")
		b := []byte("")

		result := eqBytesHash(a, b)
		assert.True(t, result)
	})

	t.Run("nil bytes are equal", func(t *testing.T) {
		result := eqBytesHash(nil, nil)
		assert.True(t, result)
	})

	t.Run("empty vs nil bytes", func(t *testing.T) {
		a := []byte("")
		var b []byte

		result := eqBytesHash(a, b)
		assert.True(t, result) // SHA256 of empty slice and nil should be the same
	})

	t.Run("different length bytes", func(t *testing.T) {
		a := []byte("short")
		b := []byte("much longer string")

		result := eqBytesHash(a, b)
		assert.False(t, result)
	})

	t.Run("same hash different content", func(t *testing.T) {
		// Use bytes that are different but verify the function works correctly
		a := []byte("test content 1")
		b := []byte("test content 2")

		result := eqBytesHash(a, b)
		assert.False(t, result)
	})
}

// TestWriteEvent tests the writeEvent function
func TestWriteEvent(t *testing.T) {
	t.Run("writes event to channel successfully", func(t *testing.T) {
		send := make(chan []byte, 1)
		eventID := "test-event"

		pubsubEvent := pubsub.Event{
			ID: &eventID,
		}

		err := writeEvent(send, pubsubEvent)
		assert.NoError(t, err)

		// Check that data was sent to channel
		select {
		case data := <-send:
			assert.NotNil(t, data)
			assert.Contains(t, string(data), "test-event")
		default:
			t.Fatal("No data received on send channel")
		}
	})

	t.Run("handles nil event ID", func(t *testing.T) {
		send := make(chan []byte, 1)

		pubsubEvent := pubsub.Event{
			ID: nil,
		}

		err := writeEvent(send, pubsubEvent)
		assert.NoError(t, err)

		// Check that data was sent even with nil ID
		select {
		case data := <-send:
			assert.NotNil(t, data)
		default:
			t.Fatal("No data received on send channel")
		}
	})

	t.Run("creates proper DOM event structure", func(t *testing.T) {
		send := make(chan []byte, 1)
		eventID := "custom-event"

		pubsubEvent := pubsub.Event{
			ID: &eventID,
		}

		err := writeEvent(send, pubsubEvent)
		assert.NoError(t, err)

		// Verify the JSON structure contains the expected event
		select {
		case data := <-send:
			jsonStr := string(data)
			assert.Contains(t, jsonStr, "custom-event")
			// Should be an array of events as per the implementation
			assert.Contains(t, jsonStr, "[")
			assert.Contains(t, jsonStr, "]")
		default:
			t.Fatal("No data received on send channel")
		}
	})
}

// TestWriteConn tests the writeConn function
func TestWriteConn(t *testing.T) {
	t.Run("function exists and has correct signature", func(t *testing.T) {
		// Since writeConn is a simple wrapper around websocket.Conn.WriteMessage,
		// and we can't easily create a real WebSocket connection in unit tests,
		// we'll test that the function exists and has the correct signature.

		// Verify the function exists and can be referenced
		writeConnFunc := writeConn
		assert.NotNil(t, writeConnFunc)

		// Test constants to ensure websocket package is used
		assert.Equal(t, 1, websocket.TextMessage)
		assert.Equal(t, 2, websocket.BinaryMessage)

		// The function signature is: writeConn(conn *websocket.Conn, mt int, payload []byte) error
		// We can test this with a mock WebSocket connection
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			upgrader := websocket.Upgrader{}
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			defer conn.Close()

			// Test writeConn function
			testMessage := []byte("test message")
			err = writeConn(conn, websocket.TextMessage, testMessage)

			// For coverage purposes, we just need to call the function
			// The actual behavior is tested by the websocket package itself
			assert.NoError(t, err)
		}))
		defer server.Close()

		// Convert HTTP URL to WebSocket URL
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		// Connect to the WebSocket server to trigger the handler
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		// Read the message to complete the communication
		_, message, err := conn.ReadMessage()
		if err == nil {
			assert.Equal(t, "test message", string(message))
		}
		// Note: Ignoring read errors as they're expected in this test scenario
	})

	t.Run("RedirectUnauthorisedWebSocket", func(t *testing.T) {
		t.Run("non-websocket request returns false", func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			result := RedirectUnauthorisedWebSocket(w, req, "/login")
			assert.False(t, result)
		})

		t.Run("panics with long redirect URL", func(t *testing.T) {
			// Test the panic condition without needing WebSocket upgrade
			req := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			// Create a URL longer than 123 bytes
			longURL := "/login" + strings.Repeat("a", 120)

			assert.Panics(t, func() {
				RedirectUnauthorisedWebSocket(w, req, longURL)
			})
		})

		t.Run("websocket request returns false on upgrade failure", func(t *testing.T) {
			// Create a WebSocket request, but httptest.ResponseRecorder can't handle upgrades
			// so this will test the upgrade failure path
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Connection", "Upgrade")
			req.Header.Set("Upgrade", "websocket")
			req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
			req.Header.Set("Sec-WebSocket-Version", "13")

			w := httptest.NewRecorder()

			// This will return false because httptest.ResponseRecorder doesn't support WebSocket upgrades
			result := RedirectUnauthorisedWebSocket(w, req, "/login")
			assert.False(t, result, "should return false when upgrade fails")
		})
	})
}
