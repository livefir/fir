package testing

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// MockRedisClient provides a mock implementation of Redis functionality
type MockRedisClient struct {
	data     map[string]string
	channels map[string][]chan string
	mu       sync.RWMutex
}

// NewMockRedisClient creates a new mock Redis client
func NewMockRedisClient() *MockRedisClient {
	return &MockRedisClient{
		data:     make(map[string]string),
		channels: make(map[string][]chan string),
	}
}

// Set sets a key-value pair
func (m *MockRedisClient) Set(key, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	return nil
}

// Get retrieves a value by key
func (m *MockRedisClient) Get(key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, exists := m.data[key]
	if !exists {
		return "", fmt.Errorf("key not found: %s", key)
	}
	return value, nil
}

// Del deletes a key
func (m *MockRedisClient) Del(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	return nil
}

// Publish publishes a message to a channel
func (m *MockRedisClient) Publish(channel, message string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if subscribers, exists := m.channels[channel]; exists {
		for _, subscriber := range subscribers {
			select {
			case subscriber <- message:
			default:
				// Subscriber is blocked, skip
			}
		}
	}
	return nil
}

// Subscribe subscribes to a channel
func (m *MockRedisClient) Subscribe(channel string) <-chan string {
	m.mu.Lock()
	defer m.mu.Unlock()

	ch := make(chan string, 100) // Buffered channel
	if _, exists := m.channels[channel]; !exists {
		m.channels[channel] = make([]chan string, 0)
	}
	m.channels[channel] = append(m.channels[channel], ch)
	return ch
}

// Unsubscribe unsubscribes from a channel
func (m *MockRedisClient) Unsubscribe(channel string, subscriber <-chan string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if subscribers, exists := m.channels[channel]; exists {
		for i, sub := range subscribers {
			if sub == subscriber {
				// Remove this subscriber
				m.channels[channel] = append(subscribers[:i], subscribers[i+1:]...)
				close(sub)
				break
			}
		}
	}
}

// Clear clears all data and channels
func (m *MockRedisClient) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Close all channels
	for _, subscribers := range m.channels {
		for _, ch := range subscribers {
			close(ch)
		}
	}

	// Reset data
	m.data = make(map[string]string)
	m.channels = make(map[string][]chan string)
}

// MockWebSocketConnection provides a mock WebSocket connection
type MockWebSocketConnection struct {
	messages    []interface{}
	readIndex   int
	writeIndex  int
	closed      bool
	mu          sync.RWMutex
	readTimeout time.Duration
}

// NewMockWebSocketConnection creates a new mock WebSocket connection
func NewMockWebSocketConnection() *MockWebSocketConnection {
	return &MockWebSocketConnection{
		messages:    make([]interface{}, 0),
		readTimeout: 5 * time.Second,
	}
}

// WriteJSON simulates writing JSON to the WebSocket
func (m *MockWebSocketConnection) WriteJSON(v interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return fmt.Errorf("connection is closed")
	}

	m.messages = append(m.messages, v)
	m.writeIndex++
	return nil
}

// ReadJSON simulates reading JSON from the WebSocket
func (m *MockWebSocketConnection) ReadJSON(v interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return fmt.Errorf("connection is closed")
	}

	if m.readIndex >= len(m.messages) {
		return fmt.Errorf("no more messages to read")
	}

	// Simulate JSON marshaling/unmarshaling
	data, err := json.Marshal(m.messages[m.readIndex])
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, v)
	if err != nil {
		return err
	}

	m.readIndex++
	return nil
}

// Close simulates closing the WebSocket connection
func (m *MockWebSocketConnection) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.closed = true
	return nil
}

// SetReadDeadline simulates setting a read deadline
func (m *MockWebSocketConnection) SetReadDeadline(t time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if t.IsZero() {
		m.readTimeout = 0
	} else {
		m.readTimeout = time.Until(t)
	}
	return nil
}

// AddMessage adds a message to be read (for testing)
func (m *MockWebSocketConnection) AddMessage(message interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.messages = append(m.messages, message)
}

// GetSentMessages returns all messages that were sent
func (m *MockWebSocketConnection) GetSentMessages() []interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return append([]interface{}{}, m.messages...)
}

// IsClosed returns whether the connection is closed
func (m *MockWebSocketConnection) IsClosed() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.closed
}

// MockHTTPHandler provides a mock HTTP handler for testing
type MockHTTPHandler struct {
	responses map[string]*http.Response
	requests  []*http.Request
	mu        sync.RWMutex
}

// NewMockHTTPHandler creates a new mock HTTP handler
func NewMockHTTPHandler() *MockHTTPHandler {
	return &MockHTTPHandler{
		responses: make(map[string]*http.Response),
		requests:  make([]*http.Request, 0),
	}
}

// ServeHTTP implements the http.Handler interface
func (m *MockHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	m.requests = append(m.requests, r)
	m.mu.Unlock()

	m.mu.RLock()
	key := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
	response, exists := m.responses[key]
	m.mu.RUnlock()

	if exists {
		// Copy headers
		for k, v := range response.Header {
			for _, val := range v {
				w.Header().Add(k, val)
			}
		}

		// Set status code
		w.WriteHeader(response.StatusCode)

		// Copy body if available
		if response.Body != nil {
			// Note: In a real implementation, you'd need to handle the body properly
			w.Write([]byte("mocked response"))
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	}
}

// SetResponse sets a mock response for a specific method and path
func (m *MockHTTPHandler) SetResponse(method, path string, response *http.Response) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s %s", method, path)
	m.responses[key] = response
}

// GetRequests returns all received requests
func (m *MockHTTPHandler) GetRequests() []*http.Request {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return append([]*http.Request{}, m.requests...)
}

// ClearRequests clears the request history
func (m *MockHTTPHandler) ClearRequests() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.requests = make([]*http.Request, 0)
}

// NetworkStub provides utilities for stubbing network calls
type NetworkStub struct {
	delays    map[string]time.Duration
	failures  map[string]error
	responses map[string]interface{}
	mu        sync.RWMutex
}

// NewNetworkStub creates a new network stub
func NewNetworkStub() *NetworkStub {
	return &NetworkStub{
		delays:    make(map[string]time.Duration),
		failures:  make(map[string]error),
		responses: make(map[string]interface{}),
	}
}

// SetDelay sets a delay for a specific endpoint
func (ns *NetworkStub) SetDelay(endpoint string, delay time.Duration) {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	ns.delays[endpoint] = delay
}

// SetFailure sets a failure for a specific endpoint
func (ns *NetworkStub) SetFailure(endpoint string, err error) {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	ns.failures[endpoint] = err
}

// SetResponse sets a response for a specific endpoint
func (ns *NetworkStub) SetResponse(endpoint string, response interface{}) {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	ns.responses[endpoint] = response
}

// SimulateCall simulates a network call
func (ns *NetworkStub) SimulateCall(endpoint string) (interface{}, error) {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	// Check for delay
	if delay, exists := ns.delays[endpoint]; exists {
		time.Sleep(delay)
	}

	// Check for failure
	if err, exists := ns.failures[endpoint]; exists {
		return nil, err
	}

	// Return response
	if response, exists := ns.responses[endpoint]; exists {
		return response, nil
	}

	return nil, fmt.Errorf("no mock response configured for endpoint: %s", endpoint)
}

// Clear clears all stubs
func (ns *NetworkStub) Clear() {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	ns.delays = make(map[string]time.Duration)
	ns.failures = make(map[string]error)
	ns.responses = make(map[string]interface{})
}

// MockSessionStore provides a mock session store for testing
type MockSessionStore struct {
	sessions map[string]map[string]interface{}
	mu       sync.RWMutex
}

// NewMockSessionStore creates a new mock session store
func NewMockSessionStore() *MockSessionStore {
	return &MockSessionStore{
		sessions: make(map[string]map[string]interface{}),
	}
}

// Set sets a value in a session
func (mss *MockSessionStore) Set(sessionID, key string, value interface{}) {
	mss.mu.Lock()
	defer mss.mu.Unlock()

	if _, exists := mss.sessions[sessionID]; !exists {
		mss.sessions[sessionID] = make(map[string]interface{})
	}
	mss.sessions[sessionID][key] = value
}

// Get gets a value from a session
func (mss *MockSessionStore) Get(sessionID, key string) (interface{}, bool) {
	mss.mu.RLock()
	defer mss.mu.RUnlock()

	session, exists := mss.sessions[sessionID]
	if !exists {
		return nil, false
	}

	value, exists := session[key]
	return value, exists
}

// Delete deletes a key from a session
func (mss *MockSessionStore) Delete(sessionID, key string) {
	mss.mu.Lock()
	defer mss.mu.Unlock()

	if session, exists := mss.sessions[sessionID]; exists {
		delete(session, key)
	}
}

// Clear clears a session
func (mss *MockSessionStore) Clear(sessionID string) {
	mss.mu.Lock()
	defer mss.mu.Unlock()

	delete(mss.sessions, sessionID)
}

// GetAllSessions returns all sessions (for testing)
func (mss *MockSessionStore) GetAllSessions() map[string]map[string]interface{} {
	mss.mu.RLock()
	defer mss.mu.RUnlock()

	result := make(map[string]map[string]interface{})
	for sessionID, session := range mss.sessions {
		sessionCopy := make(map[string]interface{})
		for k, v := range session {
			sessionCopy[k] = v
		}
		result[sessionID] = sessionCopy
	}
	return result
}

// TestFixture provides a complete testing environment
type TestFixture struct {
	RedisClient    *MockRedisClient
	SessionStore   *MockSessionStore
	NetworkStub    *NetworkStub
	HTTPHandler    *MockHTTPHandler
	WebSocketConns []*MockWebSocketConnection
}

// NewTestFixture creates a new test fixture with all mocks
func NewTestFixture() *TestFixture {
	return &TestFixture{
		RedisClient:    NewMockRedisClient(),
		SessionStore:   NewMockSessionStore(),
		NetworkStub:    NewNetworkStub(),
		HTTPHandler:    NewMockHTTPHandler(),
		WebSocketConns: make([]*MockWebSocketConnection, 0),
	}
}

// AddWebSocketConnection adds a mock WebSocket connection
func (tf *TestFixture) AddWebSocketConnection() *MockWebSocketConnection {
	conn := NewMockWebSocketConnection()
	tf.WebSocketConns = append(tf.WebSocketConns, conn)
	return conn
}

// Cleanup cleans up all mocks and resets state
func (tf *TestFixture) Cleanup() {
	tf.RedisClient.Clear()
	tf.NetworkStub.Clear()
	tf.HTTPHandler.ClearRequests()

	// Close all WebSocket connections
	for _, conn := range tf.WebSocketConns {
		conn.Close()
	}
	tf.WebSocketConns = make([]*MockWebSocketConnection, 0)
}

// SetupRedisScenario sets up a typical Redis testing scenario
func (tf *TestFixture) SetupRedisScenario() {
	// Set some initial data
	tf.RedisClient.Set("test:counter", "0")
	tf.RedisClient.Set("test:user:1", `{"id":1,"name":"Test User"}`)
}

// SetupWebSocketScenario sets up a typical WebSocket testing scenario
func (tf *TestFixture) SetupWebSocketScenario(numConnections int) []*MockWebSocketConnection {
	connections := make([]*MockWebSocketConnection, numConnections)
	for i := 0; i < numConnections; i++ {
		connections[i] = tf.AddWebSocketConnection()
	}
	return connections
}

// SimulateNetworkLatency simulates network latency for all endpoints
func (tf *TestFixture) SimulateNetworkLatency(delay time.Duration) {
	endpoints := []string{"/api/events", "/api/websocket", "/api/session"}
	for _, endpoint := range endpoints {
		tf.NetworkStub.SetDelay(endpoint, delay)
	}
}

// SimulateNetworkFailures simulates network failures
func (tf *TestFixture) SimulateNetworkFailures() {
	tf.NetworkStub.SetFailure("/api/events", fmt.Errorf("network timeout"))
	tf.NetworkStub.SetFailure("/api/websocket", fmt.Errorf("connection refused"))
}
