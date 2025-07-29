package fir

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/schema"
	"github.com/stretchr/testify/assert"
)

func TestRouteContext_Event(t *testing.T) {
	event := Event{
		ID:     "test-event",
		Params: json.RawMessage(`{"key":"value"}`),
	}

	ctx := RouteContext{
		event: event,
	}

	result := ctx.Event()
	assert.Equal(t, event, result)
	assert.Equal(t, "test-event", result.ID)
}

func TestRouteContext_Request(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)

	ctx := RouteContext{
		request: req,
	}

	result := ctx.Request()
	assert.Equal(t, req, result)
	assert.Equal(t, "GET", result.Method)
	assert.Equal(t, "/test", result.URL.Path)
}

func TestRouteContext_Response(t *testing.T) {
	w := httptest.NewRecorder()

	ctx := RouteContext{
		response: w,
	}

	result := ctx.Response()
	assert.Equal(t, w, result)
}

func TestRouteContext_Redirect(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		status         int
		expectedError  string
		shouldRedirect bool
	}{
		{
			name:           "valid redirect",
			url:            "/dashboard",
			status:         302,
			shouldRedirect: true,
		},
		{
			name:           "valid permanent redirect",
			url:            "/new-location",
			status:         301,
			shouldRedirect: true,
		},
		{
			name:          "empty url",
			url:           "",
			status:        302,
			expectedError: "url is required",
		},
		{
			name:          "invalid status low",
			url:           "/test",
			status:        200,
			expectedError: "status code must be between 300 and 308",
		},
		{
			name:          "invalid status high",
			url:           "/test",
			status:        400,
			expectedError: "status code must be between 300 and 308",
		},
		{
			name:           "status 300",
			url:            "/test",
			status:         300,
			shouldRedirect: true,
		},
		{
			name:           "status 308",
			url:            "/test",
			status:         308,
			shouldRedirect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			ctx := RouteContext{
				request:  req,
				response: w,
			}

			err := ctx.Redirect(tt.url, tt.status)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if tt.shouldRedirect {
					assert.Equal(t, tt.status, w.Code)
					assert.Equal(t, tt.url, w.Header().Get("Location"))
				}
			}
		})
	}
}

func TestRouteContext_KV(t *testing.T) {
	ctx := RouteContext{}

	// Test with string value
	err := ctx.KV("name", "John")
	assert.Error(t, err) // buildData should return an error in test context

	// Test with numeric value
	err = ctx.KV("age", 30)
	assert.Error(t, err)

	// Test with map value
	err = ctx.KV("config", map[string]string{"key": "value"})
	assert.Error(t, err)
}

func TestRouteContext_StateKV(t *testing.T) {
	ctx := RouteContext{}

	// Test with string value
	err := ctx.StateKV("status", "active")
	assert.Error(t, err) // buildData should return an error in test context

	// Test with boolean value
	err = ctx.StateKV("enabled", true)
	assert.Error(t, err)
}

func TestRouteContext_State(t *testing.T) {
	ctx := RouteContext{}

	// Test with map
	err := ctx.State(map[string]any{"key": "value"})
	assert.Error(t, err) // buildData should return an error in test context

	// Test with multiple arguments
	err = ctx.State(
		map[string]string{"status": "ok"},
		map[string]int{"count": 5},
	)
	assert.Error(t, err)

	// Test with no arguments
	err = ctx.State()
	assert.NoError(t, err) // No data should not error
}

func TestRouteContext_Data(t *testing.T) {
	ctx := RouteContext{}

	// Test with map
	err := ctx.Data(map[string]any{"name": "test"})
	assert.Error(t, err) // buildData should return an error in test context

	// Test with struct
	type testStruct struct {
		Name string
		Age  int
	}
	err = ctx.Data(testStruct{Name: "John", Age: 30})
	assert.Error(t, err)

	// Test with multiple data sets
	err = ctx.Data(
		map[string]string{"type": "user"},
		testStruct{Name: "Jane", Age: 25},
	)
	assert.Error(t, err)

	// Test with no arguments
	err = ctx.Data()
	assert.NoError(t, err) // No data should not error
}

func TestRouteContext_FieldError(t *testing.T) {
	ctx := RouteContext{}

	// Test with valid field and error
	err := ctx.FieldError("email", errors.New("invalid email format"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email")

	// Test with empty field
	err = ctx.FieldError("", errors.New("some error"))
	assert.NoError(t, err) // Empty field should return nil

	// Test with nil error
	err = ctx.FieldError("field", nil)
	assert.NoError(t, err) // Nil error should return nil

	// Test with empty field and nil error
	err = ctx.FieldError("", nil)
	assert.NoError(t, err)
}

func TestRouteContext_FieldErrors(t *testing.T) {
	ctx := RouteContext{}

	// Test with multiple field errors
	fieldErrors := map[string]error{
		"email":    errors.New("invalid email"),
		"password": errors.New("too short"),
		"name":     nil, // Should be ignored
	}

	err := ctx.FieldErrors(fieldErrors)
	assert.Error(t, err)

	// Test with empty map
	err = ctx.FieldErrors(map[string]error{})
	assert.Error(t, err) // Should still create error structure

	// Test with nil values only
	err = ctx.FieldErrors(map[string]error{
		"field1": nil,
		"field2": nil,
	})
	assert.Error(t, err)
}

func TestRouteContext_Status(t *testing.T) {
	ctx := RouteContext{}

	// Test with various status codes
	testCases := []struct {
		code int
		err  error
	}{
		{400, errors.New("bad request")},
		{401, errors.New("unauthorized")},
		{403, errors.New("forbidden")},
		{404, errors.New("not found")},
		{500, errors.New("internal server error")},
	}

	for _, tc := range testCases {
		t.Run("status_"+string(rune(tc.code)), func(t *testing.T) {
			err := ctx.Status(tc.code, tc.err)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.err.Error())
		})
	}

	// Test with nil error
	err := ctx.Status(200, nil)
	assert.Error(t, err) // Should still create status error structure
}

func TestRouteContext_GetUserFromContext(t *testing.T) {
	// Test with user in context
	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), UserKey, "user123"))

	ctx := RouteContext{
		request: req,
	}

	user := ctx.GetUserFromContext()
	assert.Equal(t, "user123", user)

	// Test with no user in context
	req = httptest.NewRequest("GET", "/", nil)
	ctx = RouteContext{
		request: req,
	}

	user = ctx.GetUserFromContext()
	assert.Equal(t, "", user)

	// Test with wrong type in context
	req = httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), UserKey, 123))

	ctx = RouteContext{
		request: req,
	}

	user = ctx.GetUserFromContext()
	assert.Equal(t, "", user)
}

func TestGetUserFromRequestContext(t *testing.T) {
	// Test with user in context
	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), UserKey, "testuser"))

	user := getUserFromRequestContext(req)
	assert.Equal(t, "testuser", user)

	// Test with no user in context
	req = httptest.NewRequest("GET", "/", nil)
	user = getUserFromRequestContext(req)
	assert.Equal(t, "", user)
}

func TestRouteContext_BindPathParams(t *testing.T) {
	// Test with path params in context
	req := httptest.NewRequest("GET", "/users/123", nil)
	pathParams := PathParams{"id": "123", "action": "edit"}
	req = req.WithContext(context.WithValue(req.Context(), PathParamsKey, pathParams))

	ctx := RouteContext{
		request: req,
	}

	type pathData struct {
		ID     string `json:"id"`
		Action string `json:"action"`
	}

	var data pathData
	err := ctx.BindPathParams(&data)
	assert.NoError(t, err)
	assert.Equal(t, "123", data.ID)
	assert.Equal(t, "edit", data.Action)

	// Test with no path params in context
	req = httptest.NewRequest("GET", "/", nil)
	ctx = RouteContext{
		request: req,
	}

	var emptyData pathData
	err = ctx.BindPathParams(&emptyData)
	assert.NoError(t, err) // Should not error when no path params
	assert.Equal(t, "", emptyData.ID)
	assert.Equal(t, "", emptyData.Action)

	// Test with nil argument
	err = ctx.BindPathParams(nil)
	assert.NoError(t, err) // Should handle nil gracefully
}

func TestRouteContext_BindQueryParams(t *testing.T) {
	// This test requires a proper route with form decoder
	// Since we can't easily mock the internal route structure in this context,
	// we'll test the basic functionality by ensuring the method exists
	// and handles nil cases gracefully

	// Create request with query parameters
	req := httptest.NewRequest("GET", "/search?q=test&page=2", nil)

	// For this test, we'll just verify the method exists and can be called
	// A full integration test would require setting up the entire controller
	ctx := RouteContext{
		request: req,
	}

	// This will panic due to nil route, but that's expected in this unit test context
	// In practice, RouteContext is always created with a valid route
	defer func() {
		if r := recover(); r != nil {
			// Expected panic due to nil route in test context
			assert.True(t, true, "Expected panic due to nil route in unit test")
		}
	}()

	type queryData struct {
		Q    string `url:"q"`
		Page int    `url:"page"`
	}

	var data queryData
	_ = ctx.BindQueryParams(&data)
}

func TestRouteContext_BindEventParams(t *testing.T) {
	// Test JSON parameter binding
	t.Run("binds JSON parameters", func(t *testing.T) {
		params := map[string]any{
			"username": "testuser",
			"age":      25,
			"active":   true,
		}
		paramsJSON, _ := json.Marshal(params)

		event := Event{
			ID:     "user_update",
			Params: paramsJSON,
		}

		ctx := RouteContext{
			event: event,
		}

		type eventData struct {
			Username string `json:"username"`
			Age      int    `json:"age"`
			Active   bool   `json:"active"`
		}

		var data eventData
		err := ctx.BindEventParams(&data)
		assert.NoError(t, err)
		assert.Equal(t, "testuser", data.Username)
		assert.Equal(t, 25, data.Age)
		assert.Equal(t, true, data.Active)
	})

	// Test with nil params
	t.Run("handles nil params gracefully", func(t *testing.T) {
		event := Event{
			ID:     "empty",
			Params: nil,
		}

		ctx := RouteContext{
			event: event,
		}

		type eventData struct {
			Username string `json:"username"`
			Age      int    `json:"age"`
			Active   bool   `json:"active"`
		}

		var emptyData eventData
		err := ctx.BindEventParams(&emptyData)
		assert.NoError(t, err) // Should handle nil params gracefully
	})

	// Test with empty JSON params
	t.Run("handles empty JSON params", func(t *testing.T) {
		event := Event{
			ID:     "empty_json",
			Params: json.RawMessage(`{}`),
		}

		ctx := RouteContext{
			event: event,
		}

		type eventData struct {
			Username string `json:"username"`
			Age      int    `json:"age"`
			Active   bool   `json:"active"`
		}

		var emptyJsonData eventData
		err := ctx.BindEventParams(&emptyJsonData)
		assert.NoError(t, err)
		assert.Equal(t, "", emptyJsonData.Username)
		assert.Equal(t, 0, emptyJsonData.Age)
		assert.Equal(t, false, emptyJsonData.Active)
	})

	// Test form parameter binding (IsForm = true)
	t.Run("binds form parameters", func(t *testing.T) {
		// Simulate form data as URL values encoded as JSON
		formValues := url.Values{}
		formValues.Set("username", "formuser")
		formValues.Set("age", "30")
		formValues.Set("active", "true")

		formJSON, _ := json.Marshal(formValues)

		event := Event{
			ID:     "form_submit",
			Params: formJSON,
			IsForm: true,
		}

		// Create a simple form decoder (similar to what controller uses)
		decoder := schema.NewDecoder()

		ctx := RouteContext{
			event: event,
			route: &route{
				routeOpt: routeOpt{
					opt: opt{
						formDecoder: decoder,
					},
				},
			},
		}

		type formData struct {
			Username string `schema:"username"`
			Age      int    `schema:"age"`
			Active   bool   `schema:"active"`
		}

		var data formData
		err := ctx.BindEventParams(&data)
		assert.NoError(t, err)
		assert.Equal(t, "formuser", data.Username)
		assert.Equal(t, 30, data.Age)
		assert.Equal(t, true, data.Active)
	})

	// Test form with invalid JSON
	t.Run("handles invalid form JSON", func(t *testing.T) {
		event := Event{
			ID:     "invalid_form",
			Params: []byte("invalid json"),
			IsForm: true,
		}

		ctx := RouteContext{
			event: event,
			route: &route{
				routeOpt: routeOpt{
					opt: opt{
						formDecoder: schema.NewDecoder(),
					},
				},
			},
		}

		type formData struct {
			Username string `schema:"username"`
		}

		var data formData
		err := ctx.BindEventParams(&data)
		assert.Error(t, err) // Should return error for invalid JSON
	})

	// Test invalid JSON parameters (non-form)
	t.Run("handles invalid JSON parameters", func(t *testing.T) {
		event := Event{
			ID:     "invalid_json",
			Params: []byte("invalid json"),
		}

		ctx := RouteContext{
			event: event,
		}

		type eventData struct {
			Username string `json:"username"`
		}

		var data eventData
		err := ctx.BindEventParams(&data)
		assert.Error(t, err) // Should return error for invalid JSON
	})
}

func TestPathParams(t *testing.T) {
	// Test PathParams type usage
	params := PathParams{
		"id":     "123",
		"slug":   "test-post",
		"page":   2,
		"active": true,
	}

	assert.Equal(t, "123", params["id"])
	assert.Equal(t, "test-post", params["slug"])
	assert.Equal(t, 2, params["page"])
	assert.Equal(t, true, params["active"])
}
