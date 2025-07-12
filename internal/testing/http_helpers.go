package testing

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

// HTTPTestHelper provides common HTTP testing utilities for Fir framework tests
type HTTPTestHelper struct {
	t      *testing.T
	client *http.Client
	server *httptest.Server
}

// NewHTTPTestHelper creates a new HTTP test helper instance
func NewHTTPTestHelper(t *testing.T, server *httptest.Server) *HTTPTestHelper {
	return &HTTPTestHelper{
		t:      t,
		client: &http.Client{},
		server: server,
	}
}

// SessionManager handles session-related testing operations
type SessionManager struct {
	helper    *HTTPTestHelper
	sessionID string
}

// SessionID returns the session ID for this session manager
func (sm *SessionManager) SessionID() string {
	return sm.sessionID
}

// GetSession extracts session ID from response cookies
func (h *HTTPTestHelper) GetSession() *SessionManager {
	resp, err := h.client.Get(h.server.URL)
	require.NoError(h.t, err)
	defer resp.Body.Close()

	sessionID := h.extractSessionFromResponse(resp)
	return &SessionManager{
		helper:    h,
		sessionID: sessionID,
	}
}

// GetSessionWithID creates a session manager with a specific session ID
func (h *HTTPTestHelper) GetSessionWithID(sessionID string) *SessionManager {
	return &SessionManager{
		helper:    h,
		sessionID: sessionID,
	}
}

// extractSessionFromResponse extracts session ID from HTTP response
func (h *HTTPTestHelper) extractSessionFromResponse(resp *http.Response) string {
	for _, cookie := range resp.Cookies() {
		if strings.Contains(cookie.Name, "_fir_session_") {
			return cookie.Value
		}
	}
	return "test-session-" + fmt.Sprintf("%d", time.Now().UnixNano())
}

// SendEvent sends an event with the managed session
func (sm *SessionManager) SendEvent(eventName string, eventData interface{}) *http.Response {
	return sm.helper.SendEventWithSession(sm.sessionID, eventName, eventData)
}

// SendEventWithSession sends an event to the server using the correct JSON format
func (h *HTTPTestHelper) SendEventWithSession(sessionID, eventName string, eventData interface{}) *http.Response {
	// Prepare event data
	var params []byte
	var err error
	if eventData != nil {
		params, err = json.Marshal(eventData)
		require.NoError(h.t, err)
	}

	// Create event structure (matching Fir's Event type)
	event := map[string]interface{}{
		"id":         eventName,
		"is_form":    false,
		"params":     params,
		"session_id": &sessionID,
		"timestamp":  time.Now().UTC().UnixMilli(),
	}

	// Marshal to JSON
	eventJSON, err := json.Marshal(event)
	require.NoError(h.t, err)

	// Send POST request
	resp, err := h.client.Post(h.server.URL, "application/json", strings.NewReader(string(eventJSON)))
	require.NoError(h.t, err)

	return resp
}

// SendFormEvent sends a form-based event
func (h *HTTPTestHelper) SendFormEvent(sessionID, eventName string, formData map[string]string) *http.Response {
	// Create form data
	formParams, err := json.Marshal(formData)
	require.NoError(h.t, err)

	// Create form event
	event := map[string]interface{}{
		"id":         eventName,
		"is_form":    true,
		"params":     formParams,
		"session_id": &sessionID,
		"timestamp":  time.Now().UTC().UnixMilli(),
	}

	// Marshal to JSON
	eventJSON, err := json.Marshal(event)
	require.NoError(h.t, err)

	// Send POST request
	resp, err := h.client.Post(h.server.URL, "application/json", strings.NewReader(string(eventJSON)))
	require.NoError(h.t, err)

	return resp
}

// ResponseValidator provides utilities for validating HTTP responses
type ResponseValidator struct {
	t        *testing.T
	response *http.Response
	body     []byte
}

// ValidateResponse creates a response validator
func (h *HTTPTestHelper) ValidateResponse(resp *http.Response) *ResponseValidator {
	body, err := io.ReadAll(resp.Body)
	require.NoError(h.t, err)
	resp.Body.Close()

	return &ResponseValidator{
		t:        h.t,
		response: resp,
		body:     body,
	}
}

// StatusCode validates the HTTP status code
func (rv *ResponseValidator) StatusCode(expected int) *ResponseValidator {
	if rv.response.StatusCode != expected {
		rv.t.Errorf("Expected status code %d, got %d", expected, rv.response.StatusCode)
	}
	return rv
}

// ContainsHTML validates that the response body contains specific HTML
func (rv *ResponseValidator) ContainsHTML(expected string) *ResponseValidator {
	bodyStr := string(rv.body)
	if !strings.Contains(bodyStr, expected) {
		rv.t.Errorf("Response body does not contain expected HTML: %s\nActual body: %s", expected, bodyStr)
	}
	return rv
}

// NotContainsHTML validates that the response body does not contain specific HTML
func (rv *ResponseValidator) NotContainsHTML(unexpected string) *ResponseValidator {
	bodyStr := string(rv.body)
	if strings.Contains(bodyStr, unexpected) {
		rv.t.Errorf("Response body contains unexpected HTML: %s", unexpected)
	}
	return rv
}

// JSONContains validates that the response contains specific JSON data
func (rv *ResponseValidator) JSONContains(expected map[string]interface{}) *ResponseValidator {
	var actual map[string]interface{}
	err := json.Unmarshal(rv.body, &actual)
	require.NoError(rv.t, err)

	for key, expectedValue := range expected {
		actualValue, exists := actual[key]
		if !exists {
			rv.t.Errorf("Expected JSON key %s not found in response", key)
			continue
		}
		if actualValue != expectedValue {
			rv.t.Errorf("Expected JSON value %v for key %s, got %v", expectedValue, key, actualValue)
		}
	}
	return rv
}

// GetBody returns the response body as string
func (rv *ResponseValidator) GetBody() string {
	return string(rv.body)
}

// GetBodyBytes returns the response body as bytes
func (rv *ResponseValidator) GetBodyBytes() []byte {
	return rv.body
}

// WebSocketHelper provides utilities for WebSocket testing
type WebSocketHelper struct {
	t         *testing.T
	serverURL string
}

// NewWebSocketHelper creates a new WebSocket helper
func (h *HTTPTestHelper) NewWebSocketHelper() *WebSocketHelper {
	wsURL := "ws" + strings.TrimPrefix(h.server.URL, "http")
	return &WebSocketHelper{
		t:         h.t,
		serverURL: wsURL,
	}
}

// Connect establishes a WebSocket connection
func (wsh *WebSocketHelper) Connect() *websocket.Conn {
	dialer := websocket.Dialer{}
	conn, resp, err := dialer.Dial(wsh.serverURL, nil)
	require.NoError(wsh.t, err)
	if resp != nil {
		resp.Body.Close()
	}
	return conn
}

// ConnectMultiple establishes multiple WebSocket connections
func (wsh *WebSocketHelper) ConnectMultiple(count int) []*websocket.Conn {
	connections := make([]*websocket.Conn, count)
	for i := 0; i < count; i++ {
		connections[i] = wsh.Connect()
	}
	return connections
}

// SendEvent sends an event through WebSocket
func (wsh *WebSocketHelper) SendEvent(conn *websocket.Conn, eventName string, data map[string]interface{}) {
	if data == nil {
		data = make(map[string]interface{})
	}

	eventData := map[string]interface{}{
		"event_id":  eventName,
		"params":    []byte("{}"),
		"timestamp": time.Now().UnixMilli(),
	}

	if len(data) > 0 {
		params, err := json.Marshal(data)
		require.NoError(wsh.t, err)
		eventData["params"] = params
	}

	err := conn.WriteJSON(eventData)
	require.NoError(wsh.t, err)
}

// ReadEvent reads an event from WebSocket with timeout
func (wsh *WebSocketHelper) ReadEvent(conn *websocket.Conn, timeout time.Duration) (map[string]interface{}, error) {
	conn.SetReadDeadline(time.Now().Add(timeout))

	var event map[string]interface{}
	err := conn.ReadJSON(&event)
	return event, err
}

// CloseConnections closes multiple WebSocket connections
func (wsh *WebSocketHelper) CloseConnections(connections []*websocket.Conn) {
	for _, conn := range connections {
		if conn != nil {
			conn.Close()
		}
	}
}

// WaitForEvent waits for a specific event type with timeout
func (wsh *WebSocketHelper) WaitForEvent(conn *websocket.Conn, expectedEventType string, timeout time.Duration) map[string]interface{} {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		event, err := wsh.ReadEvent(conn, time.Until(deadline))
		if err != nil {
			continue
		}

		if eventType, ok := event["event_id"].(string); ok && eventType == expectedEventType {
			return event
		}
	}

	wsh.t.Fatalf("Expected event %s not received within timeout %v", expectedEventType, timeout)
	return nil
}

// PerformanceHelper provides utilities for performance testing
type PerformanceHelper struct {
	t      *testing.T
	helper *HTTPTestHelper
}

// NewPerformanceHelper creates a new performance testing helper
func (h *HTTPTestHelper) NewPerformanceHelper() *PerformanceHelper {
	return &PerformanceHelper{
		t:      h.t,
		helper: h,
	}
}

// MeasureEventThroughput measures events per second for a given operation
func (ph *PerformanceHelper) MeasureEventThroughput(sessionID string, eventName string, count int, concurrent bool) (float64, time.Duration) {
	start := time.Now()

	if concurrent {
		// Concurrent execution
		ch := make(chan struct{}, count)
		for i := 0; i < 10; i++ { // 10 workers
			go func() {
				for range ch {
					ph.helper.SendEventWithSession(sessionID, eventName, nil)
				}
			}()
		}

		for i := 0; i < count; i++ {
			ch <- struct{}{}
		}
		close(ch)
	} else {
		// Sequential execution
		for i := 0; i < count; i++ {
			ph.helper.SendEventWithSession(sessionID, eventName, nil)
		}
	}

	duration := time.Since(start)
	throughput := float64(count) / duration.Seconds()

	return throughput, duration
}

// MeasureMemoryUsage measures memory usage before and after an operation
func (ph *PerformanceHelper) MeasureMemoryUsage(operation func()) (uint64, uint64) {
	// Force GC and measure baseline
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Execute operation
	operation()

	// Force GC and measure final
	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	return m1.TotalAlloc, m2.TotalAlloc
}

// ConcurrencyHelper provides utilities for concurrent testing
type ConcurrencyHelper struct {
	t      *testing.T
	helper *HTTPTestHelper
}

// NewConcurrencyHelper creates a new concurrency testing helper
func (h *HTTPTestHelper) NewConcurrencyHelper() *ConcurrencyHelper {
	return &ConcurrencyHelper{
		t:      h.t,
		helper: h,
	}
}

// RunConcurrentEvents runs multiple events concurrently
func (ch *ConcurrencyHelper) RunConcurrentEvents(sessionIDs []string, eventName string, data interface{}) {
	var wg sync.WaitGroup

	for _, sessionID := range sessionIDs {
		wg.Add(1)
		go func(sid string) {
			defer wg.Done()
			ch.helper.SendEventWithSession(sid, eventName, data)
		}(sessionID)
	}

	wg.Wait()
}

// RunConcurrentOperations runs multiple operations concurrently
func (ch *ConcurrencyHelper) RunConcurrentOperations(operations []func()) {
	var wg sync.WaitGroup

	for _, op := range operations {
		wg.Add(1)
		go func(operation func()) {
			defer wg.Done()
			operation()
		}(op)
	}

	wg.Wait()
}
