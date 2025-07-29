package fir

import (
	"fmt"
	"net/http/httptest"
	"testing"

	firErrors "github.com/livefir/fir/internal/errors"
	"github.com/stretchr/testify/require"
)

func Test_handleOnLoadResult_NilErrors(t *testing.T) {
	// Test case: err == nil, onFormErr == nil
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	ctx := RouteContext{
		event:    Event{ID: "load"},
		request:  r,
		response: w,
		route:    nil, // minimal test with nil route
		isOnLoad: true,
	}

	// This should not panic even with a nil route
	require.NotPanics(t, func() {
		handleOnLoadResult(nil, nil, ctx)
	})
}

func Test_handleOnLoadResult_OnFormErrorOnly(t *testing.T) {
	// Test case: err == nil, onFormErr != nil
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	ctx := RouteContext{
		event:    Event{ID: "test"},
		request:  r,
		response: w,
		route:    nil,
		isOnLoad: true,
	}

	fieldErr := &firErrors.Fields{
		"name":  fmt.Errorf("required"),
		"email": fmt.Errorf("invalid format"),
	}

	require.NotPanics(t, func() {
		handleOnLoadResult(nil, fieldErr, ctx)
	})
}

func Test_handleOnLoadResult_RouteDataError(t *testing.T) {
	// Test case: err == *routeData
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	ctx := RouteContext{
		event:    Event{ID: "test"},
		request:  r,
		response: w,
		route:    nil,
		isOnLoad: true,
	}

	routeDataErr := &routeData{"message": "test data"}

	require.NotPanics(t, func() {
		handleOnLoadResult(routeDataErr, nil, ctx)
	})
}

func Test_handleOnLoadResult_StatusError(t *testing.T) {
	// Test case: err == firErrors.Status
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	ctx := RouteContext{
		event:    Event{ID: "test"},
		request:  r,
		response: w,
		route:    nil,
		isOnLoad: true,
	}

	statusErr := firErrors.Status{
		Code: 400,
		Err:  fmt.Errorf("bad request"),
	}

	require.NotPanics(t, func() {
		handleOnLoadResult(statusErr, nil, ctx)
	})
}

func Test_handleOnLoadResult_FieldsError(t *testing.T) {
	// Test case: err == firErrors.Fields
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	ctx := RouteContext{
		event:    Event{ID: "test"},
		request:  r,
		response: w,
		route:    nil,
		isOnLoad: true,
	}

	fieldsErr := firErrors.Fields{
		"name":  fmt.Errorf("required"),
		"email": fmt.Errorf("invalid"),
	}

	require.NotPanics(t, func() {
		handleOnLoadResult(fieldsErr, nil, ctx)
	})
}

func Test_handleOnLoadResult_DefaultError(t *testing.T) {
	// Test case: err == generic error
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	ctx := RouteContext{
		event:    Event{ID: "test"},
		request:  r,
		response: w,
		route:    nil,
		isOnLoad: true,
	}

	genericErr := fmt.Errorf("generic error")

	require.NotPanics(t, func() {
		handleOnLoadResult(genericErr, nil, ctx)
	})
}

func Test_handleOnLoadResult_CombinedErrors(t *testing.T) {
	// Test case: both err and onFormErr are set
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	ctx := RouteContext{
		event:    Event{ID: "test"},
		request:  r,
		response: w,
		route:    nil,
		isOnLoad: true,
	}

	routeDataErr := &routeData{"message": "test"}
	fieldErr := &firErrors.Fields{
		"field1": fmt.Errorf("validation error"),
	}

	require.NotPanics(t, func() {
		handleOnLoadResult(routeDataErr, fieldErr, ctx)
	})
}
