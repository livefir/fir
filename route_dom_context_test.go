package fir

import (
	"net/http/httptest"
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func TestNewFirFuncMap(t *testing.T) {
	// Test newFirFuncMap function
	req := httptest.NewRequest("GET", "/test-path", nil)
	route := &route{
		routeOpt: routeOpt{
			opt: opt{
				appName:         "TestApp",
				developmentMode: true,
			},
		},
	}

	ctx := RouteContext{
		request: req,
		route:   route,
	}

	errs := map[string]any{
		"field1": "Error 1",
		"field2": "Error 2",
	}

	funcMap := newFirFuncMap(ctx, errs)
	assert.NotNil(t, funcMap)

	// Test that the "fir" function exists and returns a RouteDOMContext
	firFunc, exists := funcMap["fir"]
	assert.True(t, exists)
	assert.NotNil(t, firFunc)

	// Call the fir function
	firResult := firFunc.(func() *RouteDOMContext)()
	assert.NotNil(t, firResult)
	assert.Equal(t, "/test-path", firResult.URLPath)
	assert.Equal(t, "TestApp", firResult.Name)
	assert.True(t, firResult.Development)
	assert.Equal(t, errs, firResult.errors)
}

func TestNewRouteDOMContext(t *testing.T) {
	t.Run("with full context", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/home", nil)
		route := &route{
			routeOpt: routeOpt{
				opt: opt{
					appName:         "MyApp",
					developmentMode: false,
				},
			},
		}

		ctx := RouteContext{
			request: req,
			route:   route,
		}

		errs := map[string]any{"test": "error"}

		result := newRouteDOMContext(ctx, errs)
		assert.NotNil(t, result)
		assert.Equal(t, "/home", result.URLPath)
		assert.Equal(t, "MyApp", result.Name)
		assert.False(t, result.Development)
		assert.Equal(t, errs, result.errors)
	})

	t.Run("with nil request", func(t *testing.T) {
		route := &route{
			routeOpt: routeOpt{
				opt: opt{
					appName:         "TestApp",
					developmentMode: true,
				},
			},
		}

		ctx := RouteContext{
			request: nil,
			route:   route,
		}

		result := newRouteDOMContext(ctx, nil)
		assert.NotNil(t, result)
		assert.Equal(t, "", result.URLPath)
		assert.Equal(t, "TestApp", result.Name)
		assert.True(t, result.Development)
		assert.NotNil(t, result.errors)
		assert.Len(t, result.errors, 0)
	})

	t.Run("with nil route", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/test", nil)

		ctx := RouteContext{
			request: req,
			route:   nil,
		}

		result := newRouteDOMContext(ctx, nil)
		assert.NotNil(t, result)
		assert.Equal(t, "/api/test", result.URLPath)
		assert.Equal(t, "", result.Name)
		assert.False(t, result.Development)
		assert.NotNil(t, result.errors)
	})
}

func TestRouteDOMContext_ActiveRoute(t *testing.T) {
	domCtx := &RouteDOMContext{
		URLPath: "/current/path",
	}

	testCases := []struct {
		name     string
		path     string
		class    string
		expected string
	}{
		{
			name:     "matching path",
			path:     "/current/path",
			class:    "active",
			expected: "active",
		},
		{
			name:     "non-matching path",
			path:     "/other/path",
			class:    "active",
			expected: "",
		},
		{
			name:     "empty path",
			path:     "",
			class:    "selected",
			expected: "",
		},
		{
			name:     "root path matching",
			path:     "/current/path",
			class:    "highlighted",
			expected: "highlighted",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := domCtx.ActiveRoute(tc.path, tc.class)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestRouteDOMContext_NotActiveRoute(t *testing.T) {
	domCtx := &RouteDOMContext{
		URLPath: "/current/path",
	}

	testCases := []struct {
		name     string
		path     string
		class    string
		expected string
	}{
		{
			name:     "matching path",
			path:     "/current/path",
			class:    "inactive",
			expected: "",
		},
		{
			name:     "non-matching path",
			path:     "/other/path",
			class:    "inactive",
			expected: "inactive",
		},
		{
			name:     "empty path",
			path:     "",
			class:    "hidden",
			expected: "hidden",
		},
		{
			name:     "different path",
			path:     "/admin",
			class:    "disabled",
			expected: "disabled",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := domCtx.NotActiveRoute(tc.path, tc.class)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestRouteDOMContext_Error(t *testing.T) {
	t.Run("with errors", func(t *testing.T) {
		errors := map[string]any{
			"field1":       "Simple error",
			"nested":       map[string]any{"field": "Nested error"},
			"form":         map[string]any{"email": "Invalid email", "password": "Too short"},
			"array_field":  []string{"Error 1", "Error 2"},
			"number_field": 42,
		}

		domCtx := &RouteDOMContext{
			errors: errors,
		}

		testCases := []struct {
			name     string
			paths    []string
			expected any
		}{
			{
				name:     "simple field",
				paths:    []string{"field1"},
				expected: "Simple error",
			},
			{
				name:     "nested field with single path",
				paths:    []string{"nested.field"},
				expected: "Nested error",
			},
			{
				name:     "nested field with multiple paths",
				paths:    []string{"form", "email"},
				expected: "Invalid email",
			},
			{
				name:     "number field",
				paths:    []string{"number_field"},
				expected: float64(42), // JSON unmarshaling converts to float64
			},
			{
				name:     "non-existent field",
				paths:    []string{"missing"},
				expected: nil,
			},
			{
				name:     "empty paths",
				paths:    []string{},
				expected: nil,
			},
			{
				name:     "path to object (should return nil)",
				paths:    []string{"form"},
				expected: nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := domCtx.Error(tc.paths...)
				assert.Equal(t, tc.expected, result)
			})
		}
	})

	t.Run("with empty errors", func(t *testing.T) {
		domCtx := &RouteDOMContext{
			errors: map[string]any{},
		}

		result := domCtx.Error("any", "field")
		assert.Nil(t, result)
	})

	t.Run("with nil errors", func(t *testing.T) {
		domCtx := &RouteDOMContext{
			errors: nil,
		}

		result := domCtx.Error("test")
		assert.Nil(t, result)
	})
}

func TestGetErrorLookupPath(t *testing.T) {
	testCases := []struct {
		name     string
		paths    []string
		expected string
	}{
		{
			name:     "empty paths",
			paths:    []string{},
			expected: "default",
		},
		{
			name:     "single path",
			paths:    []string{"field"},
			expected: "field",
		},
		{
			name:     "multiple paths",
			paths:    []string{"form", "email"},
			expected: "form.email",
		},
		{
			name:     "paths with dots",
			paths:    []string{".field.", ".nested."},
			expected: "field.nested",
		},
		{
			name:     "complex path",
			paths:    []string{"user", "profile", "address", "street"},
			expected: "user.profile.address.street",
		},
		{
			name:     "path with leading/trailing dots",
			paths:    []string{"...test...", "...field..."},
			expected: "test.field",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := getErrorLookupPath(tc.paths...)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestRouteDOMContextIntegration(t *testing.T) {
	// Test integration with template system
	req := httptest.NewRequest("GET", "/profile", nil)
	route := &route{
		routeOpt: routeOpt{
			opt: opt{
				appName:         "ProfileApp",
				developmentMode: true,
			},
		},
	}

	ctx := RouteContext{
		request: req,
		route:   route,
	}

	errors := map[string]any{
		"username": "Username is required",
		"profile": map[string]any{
			"bio": "Bio too long",
		},
	}

	// Create function map
	funcMap := newFirFuncMap(ctx, errors)

	// Test template execution
	tmplText := `
{{- $fir := fir -}}
Name: {{$fir.Name}}
Development: {{$fir.Development}}
URL: {{$fir.URLPath}}
{{- if $fir.Error "username"}}
Username Error: {{$fir.Error "username"}}
{{- end}}
{{- if $fir.Error "profile" "bio"}}
Bio Error: {{$fir.Error "profile" "bio"}}
{{- end}}
Active Profile: {{$fir.ActiveRoute "/profile" "active"}}
Active Home: {{$fir.ActiveRoute "/home" "active"}}
Not Active Profile: {{$fir.NotActiveRoute "/profile" "inactive"}}
Not Active Home: {{$fir.NotActiveRoute "/home" "inactive"}}
`

	tmpl, err := template.New("test").Funcs(funcMap).Parse(tmplText)
	assert.NoError(t, err)

	var buf strings.Builder
	err = tmpl.Execute(&buf, nil)
	assert.NoError(t, err)

	result := buf.String()
	assert.Contains(t, result, "Name: ProfileApp")
	assert.Contains(t, result, "Development: true")
	assert.Contains(t, result, "URL: /profile")
	assert.Contains(t, result, "Username Error: Username is required")
	assert.Contains(t, result, "Bio Error: Bio too long")
	assert.Contains(t, result, "Active Profile: active")
	assert.Contains(t, result, "Active Home:")
	assert.Contains(t, result, "Not Active Profile:")
	assert.Contains(t, result, "Not Active Home: inactive")
}
