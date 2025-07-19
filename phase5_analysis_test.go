package fir

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestPhase5HandlerChainAnalysis analyzes which requests use handler chain vs legacy fallback
func TestPhase5HandlerChainAnalysis(t *testing.T) {
	// Enable debug logging for this test
	testCases := []struct {
		name      string
		options   []ControllerOption
		routeFunc RouteFunc
	}{
		{
			name:      "doubler-debug",
			routeFunc: doubler,
			options:   []ControllerOption{DevelopmentMode(true)}, // Enable debug logging
		},
	}

	for _, tc := range testCases {
		controller := NewController(tc.name, tc.options...)
		server := httptest.NewServer(controller.RouteFunc(tc.routeFunc))
		defer server.Close()

		t.Run(tc.name, func(t *testing.T) {
			// Test 1: GET request (should test both handler chain and legacy)
			t.Run("GET_request", func(t *testing.T) {
				resp, err := http.Get(server.URL)
				if err != nil {
					t.Fatal(err)
				}
				defer resp.Body.Close()
				t.Logf("GET request status: %d", resp.StatusCode)
			})

			// Test 2: JSON event request
			t.Run("JSON_event_request", func(t *testing.T) {
				// Get session first
				getResp, err := http.Get(server.URL)
				if err != nil {
					t.Fatal(err)
				}
				defer getResp.Body.Close()

				var sessionID string
				for _, cookie := range getResp.Cookies() {
					if cookie.Name == "_fir_session_" {
						sessionID = cookie.Value
						break
					}
				}

				if sessionID == "" {
					t.Fatal("Could not get session cookie")
				}

				// Create event
				event := Event{
					ID:        "double",
					IsForm:    false,
					Params:    mustMarshalTest(t, map[string]int{"num": 5}),
					SessionID: &sessionID,
					Timestamp: time.Now().UTC().UnixMilli(),
				}

				payload := mustMarshalTest(t, event)
				req := httptest.NewRequest("POST", server.URL, bytes.NewReader(payload))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("X-FIR-MODE", "event")
				req.AddCookie(&http.Cookie{Name: "_fir_session_", Value: sessionID})

				resp := httptest.NewRecorder()
				server.Config.Handler.ServeHTTP(resp, req)

				t.Logf("JSON event request status: %d", resp.Code)
				if resp.Code != http.StatusOK {
					t.Logf("JSON event response body: %s", resp.Body.String())
				}
			})

			// Test 3: WebSocket upgrade request
			t.Run("WebSocket_upgrade_request", func(t *testing.T) {
				req := httptest.NewRequest("GET", server.URL, nil)
				req.Header.Set("Connection", "Upgrade")
				req.Header.Set("Upgrade", "websocket")
				req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
				req.Header.Set("Sec-WebSocket-Version", "13")

				resp := httptest.NewRecorder()
				server.Config.Handler.ServeHTTP(resp, req)

				t.Logf("WebSocket upgrade request status: %d", resp.Code)
			})

			// Test 4: Form POST request
			t.Run("Form_POST_request", func(t *testing.T) {
				formData := "test=value"
				req := httptest.NewRequest("POST", server.URL, bytes.NewReader([]byte(formData)))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

				resp := httptest.NewRecorder()
				server.Config.Handler.ServeHTTP(resp, req)

				t.Logf("Form POST request status: %d", resp.Code)
			})
		})
	}
}

func mustMarshalTest(t *testing.T, v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
