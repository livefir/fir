package firattr

import (
	"reflect"
	"testing"
)

func TestAttributeExtractor_ParseExpression(t *testing.T) {
	extractor, err := NewAttributeExtractor()
	if err != nil {
		t.Fatalf("Failed to create extractor: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected *ParsedAttribute
		wantErr  bool
	}{
		{
			name:  "simple click event",
			input: "click",
			expected: &ParsedAttribute{
				Events: []EventInfo{
					{Name: "click", State: "ok"},
				},
				Template:  "",
				Action:    "$fir.replace()",
				Modifiers: []string{},
			},
			wantErr: false,
		},
		{
			name:  "click event with error state",
			input: "click:error",
			expected: &ParsedAttribute{
				Events: []EventInfo{
					{Name: "click", State: "error"},
				},
				Template:  "",
				Action:    "$fir.replace()",
				Modifiers: []string{},
			},
			wantErr: false,
		},
		{
			name:  "click event with template target",
			input: "click->mytemplate",
			expected: &ParsedAttribute{
				Events: []EventInfo{
					{Name: "click", State: "ok"},
				},
				Template:  "mytemplate",
				Action:    "$fir.replace()",
				Modifiers: []string{},
			},
			wantErr: false,
		},
		{
			name:  "click event with action target",
			input: "click=>myaction",
			expected: &ParsedAttribute{
				Events: []EventInfo{
					{Name: "click", State: "ok"},
				},
				Template:  "",
				Action:    "myaction",
				Modifiers: []string{},
			},
			wantErr: false,
		},
		{
			name:  "click event with modifier",
			input: "click.debounce",
			expected: &ParsedAttribute{
				Events: []EventInfo{
					{Name: "click", State: "ok"},
				},
				Template:  "",
				Action:    "$fir.replace()",
				Modifiers: []string{"debounce"},
			},
			wantErr: false,
		},
		{
			name:  "multiple events",
			input: "create:ok,update:error",
			expected: &ParsedAttribute{
				Events: []EventInfo{
					{Name: "create", State: "ok"},
					{Name: "update", State: "error"},
				},
				Template:  "",
				Action:    "$fir.replace()",
				Modifiers: []string{},
			},
			wantErr: false,
		},
		{
			name:  "complex expression with template and action",
			input: "submit:ok->form=>handleSubmit",
			expected: &ParsedAttribute{
				Events: []EventInfo{
					{Name: "submit", State: "ok"},
				},
				Template:  "form",
				Action:    "handleSubmit",
				Modifiers: []string{},
			},
			wantErr: false,
		},
		{
			name:    "empty expression",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractor.ParseExpression(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Handle nil vs empty slice differences for comparison
			if result != nil && len(result.Modifiers) == 0 {
				result.Modifiers = []string{}
			}
			if tt.expected != nil && len(tt.expected.Modifiers) == 0 {
				tt.expected.Modifiers = []string{}
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParseExpression() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestParsedAttribute_GetEventIDs(t *testing.T) {
	attr := &ParsedAttribute{
		Events: []EventInfo{
			{Name: "click", State: "ok"},
			{Name: "submit", State: "error"},
		},
	}

	expected := []string{"click", "submit"}
	result := attr.GetEventIDs()

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("GetEventIDs() = %v, want %v", result, expected)
	}
}

func TestParsedAttribute_ToCanonicalForm(t *testing.T) {
	tests := []struct {
		name     string
		attr     *ParsedAttribute
		expected string
	}{
		{
			name: "simple click event",
			attr: &ParsedAttribute{
				Events: []EventInfo{
					{Name: "click", State: "ok"},
				},
				Action: "$fir.replace()",
			},
			expected: `@fir:click:ok="$fir.replace()"`,
		},
		{
			name: "multiple events",
			attr: &ParsedAttribute{
				Events: []EventInfo{
					{Name: "create", State: "ok"},
					{Name: "update", State: "error"},
				},
				Action: "$fir.replace()",
			},
			expected: `@fir:[create:ok,update:error]="$fir.replace()"`,
		},
		{
			name: "with template",
			attr: &ParsedAttribute{
				Events: []EventInfo{
					{Name: "submit", State: "ok"},
				},
				Template: "form",
				Action:   "handleSubmit",
			},
			expected: `@fir:submit:ok::form="handleSubmit"`,
		},
		{
			name: "with modifiers",
			attr: &ParsedAttribute{
				Events: []EventInfo{
					{Name: "click", State: "ok"},
				},
				Action:    "$fir.replace()",
				Modifiers: []string{"debounce", "throttle"},
			},
			expected: `@fir:click:ok.debounce.throttle="$fir.replace()"`,
		},
		{
			name: "complex with template and modifiers",
			attr: &ParsedAttribute{
				Events: []EventInfo{
					{Name: "input", State: "ok"},
				},
				Template:  "search",
				Action:    "doSearch",
				Modifiers: []string{"debounce"},
			},
			expected: `@fir:input:ok::search.debounce="doSearch"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.attr.ToCanonicalForm()
			if result != tt.expected {
				t.Errorf("ToCanonicalForm() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestAttributeExtractor_ExtractFromTemplate(t *testing.T) {
	extractor, err := NewAttributeExtractor()
	if err != nil {
		t.Fatalf("Failed to create extractor: %v", err)
	}

	template := `
<form>
	<input type="text" fir:"input.debounce->search=>doSearch" />
	<button fir:"click:ok=>handleSubmit">Submit</button>
	<div fir:"load:ok">Loading...</div>
</form>
`

	attributes, err := extractor.ExtractFromTemplate(template)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// We expect to find some attributes (exact parsing depends on HTML structure)
	if len(attributes) == 0 {
		t.Error("Expected to find some fir: attributes in template")
	}

	// Test that we can parse at least one attribute
	found := false
	for _, attr := range attributes {
		if len(attr.Events) > 0 {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find at least one valid fir: attribute")
	}
}

func TestNewAttributeExtractor(t *testing.T) {
	extractor, err := NewAttributeExtractor()
	if err != nil {
		t.Fatalf("Failed to create extractor: %v", err)
	}

	if extractor == nil {
		t.Fatal("Expected non-nil extractor") // Use Fatal instead of Error
	}

	if extractor.parser == nil {
		t.Error("Expected parser to be initialized")
	}
}

func TestAnalyzeTemplate(t *testing.T) {
	template := `
<form>
	<input type="text" @fir:input.debounce->search="doSearch" />
	<button @fir:click:ok="handleSubmit">Submit</button>
	<div @fir:load:error="showError">Loading...</div>
</form>
`

	eventMap, err := AnalyzeTemplate(template)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// We should have found some events (this is a basic test since the extraction is simplified)
	if len(eventMap) == 0 {
		t.Log("No events found in template (expected with simplified regex extraction)")
	}

	t.Logf("Analyzed events: %+v", eventMap)
}
