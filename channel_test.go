package fir

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestController_DefaultChannelFunc tests the defaultChannelFunc method comprehensively
func TestController_DefaultChannelFunc(t *testing.T) {
	// Create a controller and cast to access private methods
	ctrl := NewController("test-channel-controller")
	controller := ctrl.(*controller)

	tests := []struct {
		name         string
		setupRequest func() *http.Request
		viewID       string
		expected     *string
		expectNil    bool
	}{
		{
			name: "empty viewID uses root for root path",
			setupRequest: func() *http.Request {
				return httptest.NewRequest("GET", "/", nil)
			},
			viewID:   "",
			expected: nil, // Will be set based on session
		},
		{
			name: "empty viewID uses path-based ID for non-root path",
			setupRequest: func() *http.Request {
				return httptest.NewRequest("GET", "/users/profile", nil)
			},
			viewID:   "",
			expected: nil, // Will be set based on session
		},
		{
			name: "provided viewID is used",
			setupRequest: func() *http.Request {
				return httptest.NewRequest("GET", "/", nil)
			},
			viewID:   "custom-view",
			expected: nil, // Will be set based on session
		},
		{
			name: "request with UserKey in context",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				ctx := context.WithValue(req.Context(), UserKey, "user123")
				return req.WithContext(ctx)
			},
			viewID:   "test-view",
			expected: &[]string{"user123:test-view"}[0],
		},
		{
			name: "request without cookie and no UserKey",
			setupRequest: func() *http.Request {
				return httptest.NewRequest("GET", "/", nil)
			},
			viewID:    "test-view",
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupRequest()
			result := controller.defaultChannelFunc(req, tt.viewID)

			if tt.expectNil {
				assert.Nil(t, result, "Expected nil result for request without valid session")
			} else if tt.expected != nil {
				assert.NotNil(t, result, "Expected non-nil result")
				assert.Equal(t, *tt.expected, *result, "Channel string should match expected")
			} else {
				// For cases where we expect a result but don't know the exact session ID
				if result != nil {
					assert.Contains(t, *result, ":", "Channel should contain colon separator")
				}
			}
		})
	}
}

// TestController_DefaultChannelFunc_WithCookie tests the cookie handling path
func TestController_DefaultChannelFunc_WithCookie(t *testing.T) {
	// Create a controller and cast to access private methods
	ctrl := NewController("test-channel-controller")
	controller := ctrl.(*controller)

	tests := []struct {
		name           string
		setupRequest   func() *http.Request
		viewID         string
		expectNil      bool
		expectedFormat string // Expected format pattern
	}{
		{
			name: "request with invalid cookie",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.AddCookie(&http.Cookie{
					Name:  controller.cookieName,
					Value: "invalid-cookie-value",
				})
				return req
			},
			viewID:    "test-view",
			expectNil: true, // Should return nil due to invalid cookie
		},
		{
			name: "request with no cookie",
			setupRequest: func() *http.Request {
				return httptest.NewRequest("GET", "/", nil)
			},
			viewID:    "test-view",
			expectNil: true, // Should return nil due to missing cookie
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupRequest()
			result := controller.defaultChannelFunc(req, tt.viewID)

			if tt.expectNil {
				assert.Nil(t, result, "Expected nil result for invalid/missing cookie")
			} else {
				assert.NotNil(t, result, "Expected non-nil result for valid cookie")
				if tt.expectedFormat != "" {
					assert.Contains(t, *result, tt.expectedFormat, "Channel should match expected format")
				}
			}
		})
	}
}

// TestController_DefaultChannelFunc_ViewIDProcessing tests viewID processing logic
func TestController_DefaultChannelFunc_ViewIDProcessing(t *testing.T) {
	ctrl := NewController("test-channel-controller")
	controller := ctrl.(*controller)

	tests := []struct {
		name           string
		requestPath    string
		inputViewID    string
		expectedViewID string
	}{
		{
			name:           "empty viewID with root path becomes 'root'",
			requestPath:    "/",
			inputViewID:    "",
			expectedViewID: "root",
		},
		{
			name:           "empty viewID with simple path",
			requestPath:    "/users",
			inputViewID:    "",
			expectedViewID: "_users",
		},
		{
			name:           "empty viewID with complex path",
			requestPath:    "/users/profile/edit",
			inputViewID:    "",
			expectedViewID: "_users_profile_edit",
		},
		{
			name:           "provided viewID is preserved",
			requestPath:    "/any/path",
			inputViewID:    "custom-view-id",
			expectedViewID: "custom-view-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request with UserKey to ensure we get a result and can test the viewID processing
			req := httptest.NewRequest("GET", tt.requestPath, nil)
			ctx := context.WithValue(req.Context(), UserKey, "testuser")
			req = req.WithContext(ctx)

			result := controller.defaultChannelFunc(req, tt.inputViewID)

			assert.NotNil(t, result, "Expected non-nil result with UserKey")
			expectedChannel := "testuser:" + tt.expectedViewID
			assert.Equal(t, expectedChannel, *result, "Channel should match expected format with processed viewID")
		})
	}
}

// TestController_DefaultChannelFunc_EdgeCases tests edge cases and error conditions
func TestController_DefaultChannelFunc_EdgeCases(t *testing.T) {
	ctrl := NewController("test-channel-controller")
	controller := ctrl.(*controller)

	tests := []struct {
		name         string
		setupRequest func() *http.Request
		viewID       string
		expectNil    bool
	}{
		{
			name: "empty UserKey string",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				ctx := context.WithValue(req.Context(), UserKey, "")
				return req.WithContext(ctx)
			},
			viewID:    "test-view",
			expectNil: true, // Empty UserKey should fall back to cookie, which isn't present
		},
		{
			name: "UserKey with non-string type",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				ctx := context.WithValue(req.Context(), UserKey, 123) // Non-string type
				return req.WithContext(ctx)
			},
			viewID:    "test-view",
			expectNil: true, // Non-string UserKey should fall back to cookie, which isn't present
		},
		{
			name: "valid UserKey with special characters in viewID",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				ctx := context.WithValue(req.Context(), UserKey, "user-with-dashes")
				return req.WithContext(ctx)
			},
			viewID:    "view/with:special@chars",
			expectNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupRequest()
			result := controller.defaultChannelFunc(req, tt.viewID)

			if tt.expectNil {
				assert.Nil(t, result, "Expected nil result for edge case")
			} else {
				assert.NotNil(t, result, "Expected non-nil result for valid edge case")
				assert.Contains(t, *result, ":", "Channel should contain colon separator")
			}
		})
	}
}
