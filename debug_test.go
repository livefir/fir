package fir

import (
	"net/http/httptest"
	"testing"
)

func TestSimpleRoute(t *testing.T) {
	controller := NewController("test", DevelopmentMode(true))

	simpleRoute := func() RouteOptions {
		return RouteOptions{
			ID("simple"),
			Content("<html><body><h1>Hello World</h1></body></html>"),
		}
	}

	server := httptest.NewServer(controller.RouteFunc(simpleRoute))
	defer server.Close()

	resp := getPageWithSession(t, server.URL)

	t.Logf("Simple route - Status: %d", resp.statusCode)
	t.Logf("Simple route - Body: %s", resp.body)
	t.Logf("Simple route - Cookies: %d", len(resp.cookies))

	if len(resp.body) == 0 {
		t.Fatal("Expected response body, got empty")
	}
}
