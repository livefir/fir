package fir

import (
	"errors"
	"testing"

	firErrors "github.com/livefir/fir/internal/errors"
	"github.com/livefir/fir/internal/eventstate"
	"github.com/livefir/fir/pubsub"
	"github.com/stretchr/testify/assert"
)

func TestHandleOnEventResult(t *testing.T) {
	// Helper function to create a basic RouteContext
	createRouteContext := func(eventID, target string) RouteContext {
		return RouteContext{
			event: Event{
				ID:         eventID,
				Target:     ptr(target),
				ElementKey: ptr("test-key"),
				SessionID:  ptr("test-session"),
			},
		}
	}

	// Helper function to create a mock event publisher
	createMockPublisher := func(expectedEvent *pubsub.Event) eventPublisher {
		return func(event pubsub.Event) error {
			if expectedEvent != nil {
				assert.Equal(t, *expectedEvent.ID, *event.ID)
				assert.Equal(t, expectedEvent.State, event.State)
				assert.Equal(t, *expectedEvent.Target, *event.Target)
				assert.Equal(t, expectedEvent.ElementKey, event.ElementKey)
				assert.Equal(t, expectedEvent.SessionID, event.SessionID)
			}
			return nil
		}
	}

	t.Run("NoError_Success", func(t *testing.T) {
		ctx := createRouteContext("event1", "#button1")

		expectedEvent := &pubsub.Event{
			ID:         &ctx.event.ID,
			State:      eventstate.OK,
			Target:     ctx.event.Target,
			ElementKey: ctx.event.ElementKey,
			SessionID:  ctx.event.SessionID,
		}

		publish := createMockPublisher(expectedEvent)

		result := handleOnEventResult(nil, ctx, publish)

		assert.Nil(t, result, "Should return nil for successful event")
	})

	t.Run("NoError_NilTarget", func(t *testing.T) {
		ctx := RouteContext{
			event: Event{
				ID:         "event1",
				Target:     nil, // nil target
				ElementKey: ptr("test-key"),
				SessionID:  ptr("test-session"),
			},
		}

		emptyTarget := ""
		expectedEvent := &pubsub.Event{
			ID:         &ctx.event.ID,
			State:      eventstate.OK,
			Target:     &emptyTarget,
			ElementKey: ctx.event.ElementKey,
			SessionID:  ctx.event.SessionID,
		}

		publish := createMockPublisher(expectedEvent)

		result := handleOnEventResult(nil, ctx, publish)

		assert.Nil(t, result, "Should return nil for successful event with nil target")
	})

	t.Run("StatusError", func(t *testing.T) {
		ctx := createRouteContext("event2", "#form1")

		statusErr := &firErrors.Status{
			Err: errors.New("validation failed"),
		}

		publish := createMockPublisher(nil) // Don't expect publication for error

		result := handleOnEventResult(statusErr, ctx, publish)

		assert.NotNil(t, result, "Should return error event")
		assert.Equal(t, ctx.event.ID, *result.ID)
		assert.Equal(t, eventstate.Error, result.State)
		assert.Equal(t, *ctx.event.Target, *result.Target)
		assert.Equal(t, ctx.event.ElementKey, result.ElementKey)
		assert.Equal(t, ctx.event.SessionID, result.SessionID)

		// Check error data structure
		assert.NotNil(t, result.Detail)
		assert.NotNil(t, result.Detail.Data)

		data := result.Detail.Data.(map[string]any)
		assert.Contains(t, data, ctx.event.ID)
		assert.Contains(t, data, "onevent")
		assert.Equal(t, "validation failed", data[ctx.event.ID])
		assert.Equal(t, "validation failed", data["onevent"])
	})

	t.Run("FieldsError", func(t *testing.T) {
		ctx := createRouteContext("event3", "#form2")

		fieldsErr := &firErrors.Fields{
			"username": errors.New("username is required"),
			"email":    errors.New("email is invalid"),
		}

		publish := createMockPublisher(nil) // Don't expect publication for error

		result := handleOnEventResult(fieldsErr, ctx, publish)

		assert.NotNil(t, result, "Should return error event")
		assert.Equal(t, ctx.event.ID, *result.ID)
		assert.Equal(t, eventstate.Error, result.State)
		assert.Equal(t, *ctx.event.Target, *result.Target)

		// Check field errors structure
		assert.NotNil(t, result.Detail)
		assert.NotNil(t, result.Detail.Data)

		data := result.Detail.Data.(map[string]any)
		assert.Contains(t, data, ctx.event.ID)

		fieldErrors := data[ctx.event.ID].(map[string]any)
		assert.Equal(t, "username is required", fieldErrors["username"])
		assert.Equal(t, "email is invalid", fieldErrors["email"])
	})

	t.Run("RouteData", func(t *testing.T) {
		ctx := createRouteContext("event4", "#content")

		routeDataValue := routeData{
			"message": "Hello World",
			"count":   42,
		}

		expectedEvent := &pubsub.Event{
			ID:         &ctx.event.ID,
			State:      eventstate.OK,
			Target:     ctx.event.Target,
			ElementKey: ctx.event.ElementKey,
			SessionID:  ctx.event.SessionID,
		}

		publish := createMockPublisher(expectedEvent)

		result := handleOnEventResult(&routeDataValue, ctx, publish)

		assert.Nil(t, result, "Should return nil for route data")
	})

	t.Run("RouteDataWithState", func(t *testing.T) {
		ctx := createRouteContext("event5", "#state-content")

		routeDataWithStateValue := routeDataWithState{
			routeData: &routeData{
				"data": "some data",
			},
			stateData: &stateData{
				"state": "active",
			},
		}

		expectedEvent := &pubsub.Event{
			ID:         &ctx.event.ID,
			State:      eventstate.OK,
			Target:     ctx.event.Target,
			ElementKey: ctx.event.ElementKey,
			SessionID:  ctx.event.SessionID,
		}

		publish := createMockPublisher(expectedEvent)

		result := handleOnEventResult(&routeDataWithStateValue, ctx, publish)

		assert.Nil(t, result, "Should return nil for route data with state")
	})

	t.Run("StateData", func(t *testing.T) {
		ctx := createRouteContext("event6", "#state-only")

		stateDataValue := stateData{
			"status":   "loading",
			"progress": 50,
		}

		expectedEvent := &pubsub.Event{
			ID:         &ctx.event.ID,
			State:      eventstate.OK,
			Target:     ctx.event.Target,
			ElementKey: ctx.event.ElementKey,
			SessionID:  ctx.event.SessionID,
		}

		publish := createMockPublisher(expectedEvent)

		result := handleOnEventResult(&stateDataValue, ctx, publish)

		assert.Nil(t, result, "Should return nil for state data")
	})

	t.Run("DefaultError", func(t *testing.T) {
		ctx := createRouteContext("event7", "#generic")

		genericErr := errors.New("something went wrong")

		publish := createMockPublisher(nil) // Don't expect publication for error

		result := handleOnEventResult(genericErr, ctx, publish)

		assert.NotNil(t, result, "Should return error event")
		assert.Equal(t, ctx.event.ID, *result.ID)
		assert.Equal(t, eventstate.Error, result.State)
		assert.Equal(t, *ctx.event.Target, *result.Target)

		// Check default error structure
		assert.NotNil(t, result.Detail)
		assert.NotNil(t, result.Detail.Data)

		data := result.Detail.Data.(map[string]any)
		assert.Contains(t, data, ctx.event.ID)
		assert.Contains(t, data, "onevent")
		assert.Equal(t, "something went wrong", data[ctx.event.ID])
		assert.Equal(t, "something went wrong", data["onevent"])
	})

	t.Run("PublishError_DoesNotAffectResult", func(t *testing.T) {
		ctx := createRouteContext("event8", "#publish-error")

		// Publisher that returns an error
		failingPublisher := func(event pubsub.Event) error {
			return errors.New("publish failed")
		}

		result := handleOnEventResult(nil, ctx, failingPublisher)

		// Even if publish fails, the function should still return nil for success case
		assert.Nil(t, result, "Should return nil despite publish error")
	})

	t.Run("ComplexRouteData", func(t *testing.T) {
		ctx := createRouteContext("event9", "#complex")

		complexData := routeData{
			"user": map[string]any{
				"name":  "John Doe",
				"email": "john@example.com",
			},
			"items": []string{"item1", "item2", "item3"},
			"nested": map[string]any{
				"level2": map[string]any{
					"value": 123,
				},
			},
		}

		publish := func(event pubsub.Event) error {
			// Verify the complex data is preserved
			assert.NotNil(t, event.Detail)
			assert.NotNil(t, event.Detail.Data)

			data := event.Detail.Data.(routeData)
			user := data["user"].(map[string]any)
			assert.Equal(t, "John Doe", user["name"])
			assert.Equal(t, "john@example.com", user["email"])

			items := data["items"].([]string)
			assert.Len(t, items, 3)
			assert.Equal(t, "item1", items[0])

			return nil
		}

		result := handleOnEventResult(&complexData, ctx, publish)

		assert.Nil(t, result, "Should return nil for complex route data")
	})

	t.Run("EmptyFieldsError", func(t *testing.T) {
		ctx := createRouteContext("event10", "#empty-fields")

		// Empty fields error
		fieldsErr := &firErrors.Fields{}

		publish := createMockPublisher(nil)

		result := handleOnEventResult(fieldsErr, ctx, publish)

		assert.NotNil(t, result, "Should return error event")
		assert.Equal(t, eventstate.Error, result.State)

		// Check that empty fields are handled correctly
		data := result.Detail.Data.(map[string]any)
		fieldErrors := data[ctx.event.ID].(map[string]any)
		assert.Empty(t, fieldErrors, "Empty fields error should result in empty field errors map")
	})

	t.Run("NilEventPublisher", func(t *testing.T) {
		ctx := createRouteContext("event11", "#nil-publisher")

		// This test ensures the function doesn't panic with nil publisher
		// Note: In real code, this should probably not happen, but testing defensive programming
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Function should not panic with nil publisher: %v", r)
			}
		}()

		result := handleOnEventResult(errors.New("test error"), ctx, nil)

		assert.NotNil(t, result, "Should return error event even with nil publisher")
	})

	t.Run("EventIDsConsistency", func(t *testing.T) {
		// Test that event IDs are consistently used across different scenarios
		eventID := "consistency-test"
		ctx := RouteContext{
			event: Event{
				ID:         eventID,
				Target:     ptr("#consistency"),
				ElementKey: ptr("consistency-key"),
				SessionID:  ptr("consistency-session"),
			},
		}

		testCases := []struct {
			name string
			err  error
		}{
			{"Status Error", &firErrors.Status{Err: errors.New("status")}},
			{"Fields Error", &firErrors.Fields{"field": errors.New("field error")}},
			{"Generic Error", errors.New("generic")},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := handleOnEventResult(tc.err, ctx, func(event pubsub.Event) error { return nil })

				assert.NotNil(t, result)
				assert.Equal(t, eventID, *result.ID, "Event ID should be consistent")
				assert.Equal(t, ctx.event.ElementKey, result.ElementKey, "Element key should be consistent")
				assert.Equal(t, ctx.event.SessionID, result.SessionID, "Session ID should be consistent")
			})
		}
	})
}
