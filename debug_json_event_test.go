package fir

import (
	"net/http"
	"testing"

	"github.com/livefir/fir/internal/handlers"
	firHttp "github.com/livefir/fir/internal/http"
)

func TestDebugJSONEventSupportsRequest(t *testing.T) {
	// Create a request with the X-FIR-MODE header
	req := &firHttp.RequestModel{
		Method: http.MethodPost,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
			"X-Fir-Mode":   []string{"event"},
		},
	}

	// Check header access with different case variations
	firMode1 := req.Header.Get("X-FIR-MODE")
	firMode2 := req.Header.Get("X-Fir-Mode")
	firMode3 := req.Header.Get("x-fir-mode")
	t.Logf("X-FIR-MODE header value: '%s'", firMode1)
	t.Logf("X-Fir-Mode header value: '%s'", firMode2)
	t.Logf("x-fir-mode header value: '%s'", firMode3)

	// Check raw header map
	for k, v := range req.Header {
		t.Logf("Header '%s': %v", k, v)
	}

	// Create handler and test
	handler := handlers.NewJSONEventHandler(nil, nil, nil, nil)
	result := handler.SupportsRequest(req)
	t.Logf("SupportsRequest result: %v", result)

	if !result {
		t.Errorf("Expected handler to support the request, but it returned false")
	}
}
