package fir

import (
	"html/template"
	"net/http/httptest"
	"testing"

	"github.com/livefir/fir/internal/dom"
	"github.com/livefir/fir/internal/eventstate"
	"github.com/livefir/fir/pubsub"
	"github.com/stretchr/testify/require"
)

func Test_buildDOMEventFromTemplate_SpecialTemplateName(t *testing.T) {
	// Test the special case where templateName == "-"

	// Create a mock route context
	ctx := createMockRouteContext(t)

	eventID := "test-event"
	pubsubEvent := pubsub.Event{
		ID:         &eventID,
		State:      eventstate.OK,
		ElementKey: stringPtr("key1"),
		Target:     stringPtr("target1"),
		Detail: &dom.Detail{
			State: stringPtr("active"),
		},
	}

	eventIDWithState := "test-event:ok"
	templateName := "-"

	result := buildDOMEventFromTemplate(ctx, pubsubEvent, eventIDWithState, templateName)

	require.NotNil(t, result)
	require.Equal(t, eventID, result.ID)
	require.Equal(t, eventstate.OK, result.State)
	require.NotNil(t, result.Type)
	require.Equal(t, "fir:"+eventIDWithState, *result.Type) // fir function adds "fir:" prefix
	require.Equal(t, "key1", *result.Key)
	require.NotNil(t, result.Detail)
	require.Equal(t, stringPtr("active"), result.Detail.State)
}

func Test_buildDOMEventFromTemplate_SpecialTemplateName_WithNilDetail(t *testing.T) {
	// Test special case with nil detail
	ctx := createMockRouteContext(t)

	eventID := "test-event"
	pubsubEvent := pubsub.Event{
		ID:         &eventID,
		State:      eventstate.OK,
		ElementKey: stringPtr("key1"),
		Target:     stringPtr("target1"),
		Detail:     nil, // nil detail
	}

	eventIDWithState := "test-event:ok"
	templateName := "-"

	result := buildDOMEventFromTemplate(ctx, pubsubEvent, eventIDWithState, templateName)

	require.NotNil(t, result)
	require.NotNil(t, result.Detail)
	require.Nil(t, result.Detail.State) // Should be nil since pubsubEvent.Detail was nil
}

func Test_buildDOMEventFromTemplate_NormalTemplate_OKState(t *testing.T) {
	// Test normal template processing with OK state
	ctx := createMockRouteContextWithNamedTemplate(t, "test-template", `<div>{{.message}}</div>`)

	eventID := "test-event"
	testData := map[string]any{"message": "hello world"}
	pubsubEvent := pubsub.Event{
		ID:         &eventID,
		State:      eventstate.OK,
		ElementKey: stringPtr("key1"),
		Target:     stringPtr("target1"),
		Detail: &dom.Detail{
			Data:  testData,
			State: stringPtr("active"),
		},
	}

	eventIDWithState := "test-event:ok"
	templateName := "test-template"

	result := buildDOMEventFromTemplate(ctx, pubsubEvent, eventIDWithState, templateName)

	require.NotNil(t, result)
	require.Equal(t, eventIDWithState, result.ID)
	require.Equal(t, eventstate.OK, result.State)
	require.NotNil(t, result.Type)
	require.Equal(t, "fir:"+eventIDWithState+"::"+templateName, *result.Type)
	require.NotNil(t, result.Detail)
	require.Equal(t, stringPtr("active"), result.Detail.State)
	require.Contains(t, result.Detail.HTML, "hello world") // Should render template
}

func Test_buildDOMEventFromTemplate_ErrorState_ValidData(t *testing.T) {
	// Test with Error state and valid error data
	ctx := createMockRouteContextWithNamedTemplate(t, "error-template", `<span class="error">Error occurred</span>`)

	eventID := "test-event"
	errorData := map[string]any{
		"field1": "error message 1",
		"field2": "error message 2",
	}
	pubsubEvent := pubsub.Event{
		ID:         &eventID,
		State:      eventstate.Error,
		ElementKey: stringPtr("key1"),
		Target:     stringPtr("target1"),
		Detail: &dom.Detail{
			Data:  errorData,
			State: stringPtr("error"),
		},
	}

	eventIDWithState := "test-event:error"
	templateName := "error-template"

	result := buildDOMEventFromTemplate(ctx, pubsubEvent, eventIDWithState, templateName)

	require.NotNil(t, result)
	require.Equal(t, eventIDWithState, result.ID)
	require.Equal(t, eventstate.Error, result.State)
	require.NotNil(t, result.Detail)
	require.Equal(t, stringPtr("error"), result.Detail.State)
	require.Contains(t, result.Detail.HTML, "Error occurred") // Should render error template
}

func Test_buildDOMEventFromTemplate_ErrorState_InvalidData(t *testing.T) {
	// Test with Error state but invalid error data (not map[string]any)
	ctx := createMockRouteContext(t)

	eventID := "test-event"
	invalidData := "not a map" // Invalid data type
	pubsubEvent := pubsub.Event{
		ID:         &eventID,
		State:      eventstate.Error,
		ElementKey: stringPtr("key1"),
		Target:     stringPtr("target1"),
		Detail: &dom.Detail{
			Data:  invalidData,
			State: stringPtr("error"),
		},
	}

	eventIDWithState := "test-event:error"
	templateName := "error-template"

	result := buildDOMEventFromTemplate(ctx, pubsubEvent, eventIDWithState, templateName)

	// Should return nil due to invalid data type
	require.Nil(t, result)
}

func Test_buildDOMEventFromTemplate_ErrorState_EmptyTemplateValue(t *testing.T) {
	// Test Error state with empty template value (should return nil)
	ctx := createMockRouteContextWithNamedTemplate(t, "empty-template", ``) // Empty template

	eventID := "test-event"
	errorData := map[string]any{"field": "error"}
	pubsubEvent := pubsub.Event{
		ID:         &eventID,
		State:      eventstate.Error,
		ElementKey: stringPtr("key1"),
		Target:     stringPtr("target1"),
		Detail: &dom.Detail{
			Data:  errorData,
			State: stringPtr("error"),
		},
	}

	eventIDWithState := "test-event:error"
	templateName := "empty-template"

	result := buildDOMEventFromTemplate(ctx, pubsubEvent, eventIDWithState, templateName)

	// Should return nil because Error state with empty value returns nil
	require.Nil(t, result)
}

func Test_buildDOMEventFromTemplate_OKState_NilTemplateData(t *testing.T) {
	// Test OK state with nil template data (should set value to empty string)
	ctx := createMockRouteContextWithNamedTemplate(t, "test-template", `<div>{{.message}}</div>`)

	eventID := "test-event"
	pubsubEvent := pubsub.Event{
		ID:         &eventID,
		State:      eventstate.OK,
		ElementKey: stringPtr("key1"),
		Target:     stringPtr("target1"),
		Detail:     nil, // Nil detail means nil template data
	}

	eventIDWithState := "test-event:ok"
	templateName := "test-template"

	result := buildDOMEventFromTemplate(ctx, pubsubEvent, eventIDWithState, templateName)

	require.NotNil(t, result)
	require.NotNil(t, result.Detail)
	require.Equal(t, "", result.Detail.HTML) // Should be empty string when templateData is nil
	require.Nil(t, result.Detail.State)      // Should be nil since pubsubEvent.Detail was nil
}

func Test_buildDOMEventFromTemplate_TemplateExecutionError(t *testing.T) {
	// Test template execution error with undefined template name
	ctx := createMockRouteContext(t) // Use base template without the specific named template

	eventID := "test-event"
	pubsubEvent := pubsub.Event{
		ID:         &eventID,
		State:      eventstate.OK,
		ElementKey: stringPtr("key1"),
		Target:     stringPtr("target1"),
		Detail: &dom.Detail{
			Data:  map[string]any{"message": "test"},
			State: stringPtr("active"),
		},
	}

	eventIDWithState := "test-event:ok"
	templateName := "undefined-template" // This template doesn't exist

	result := buildDOMEventFromTemplate(ctx, pubsubEvent, eventIDWithState, templateName)

	// Should return nil due to template not being defined
	require.Nil(t, result)
}

func Test_buildDOMEventFromTemplate_TargetHandling(t *testing.T) {
	// Test target handling with targetOrClassName function
	ctx := createMockRouteContext(t)

	testCases := []struct {
		name           string
		target         *string
		expectedTarget string
	}{
		{
			name:           "with non-empty target",
			target:         stringPtr("custom-target"),
			expectedTarget: "custom-target",
		},
		{
			name:           "with empty target",
			target:         stringPtr(""),
			expectedTarget: ".fir-test-event-ok", // Should use className with fir prefix
		},
		{
			name:           "with nil target",
			target:         nil,
			expectedTarget: ".fir-test-event-ok", // Should use className with fir prefix
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			eventID := "test-event"
			pubsubEvent := pubsub.Event{
				ID:         &eventID,
				State:      eventstate.OK,
				ElementKey: stringPtr("key1"),
				Target:     tc.target,
				Detail: &dom.Detail{
					State: stringPtr("active"),
				},
			}

			eventIDWithState := "test-event:ok"
			templateName := "-" // Use special template name for simple testing

			result := buildDOMEventFromTemplate(ctx, pubsubEvent, eventIDWithState, templateName)

			require.NotNil(t, result)
			require.NotNil(t, result.Target)
			require.Equal(t, tc.expectedTarget, *result.Target)
		})
	}
}

// Helper functions

func createMockRouteContext(t *testing.T) RouteContext {
	// Create a simple template
	tmpl := template.Must(template.New("test").Parse(`<div>{{.message}}</div>`))

	// Create a mock route
	route := &route{
		template: tmpl,
	}

	// Create mock request and response
	req := httptest.NewRequest("GET", "/test", nil)
	resp := httptest.NewRecorder()

	return RouteContext{
		request:  req,
		response: resp,
		route:    route,
		isOnLoad: false,
	}
}

func createMockRouteContextWithNamedTemplate(t *testing.T, templateName, templateContent string) RouteContext {
	// Create a template with a named sub-template
	tmpl := template.Must(template.New("test").Parse(`<div>root template</div>`))
	template.Must(tmpl.New(templateName).Parse(templateContent))

	route := &route{
		template: tmpl,
	}

	req := httptest.NewRequest("GET", "/test", nil)
	resp := httptest.NewRecorder()

	return RouteContext{
		request:  req,
		response: resp,
		route:    route,
		isOnLoad: false,
	}
}
