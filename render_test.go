package fir

import (
	"html/template"
	"testing"
)

func TestBuildTemplateValue(t *testing.T) {
	// Test case 1: nil template
	result, err := buildTemplateValue(nil, "templateName", "data")
	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}
	if result != "" {
		t.Errorf("Expected empty result, got: %s", result)
	}

	// Test case 2: empty template name
	result, err = buildTemplateValue(&template.Template{}, "", "data")
	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}
	if result != "" {
		t.Errorf("Expected empty result, got: %s", result)
	}

	// Test case 3: templateName == "_fir_html"
	result, err = buildTemplateValue(&template.Template{}, "_fir_html", "data")
	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}
	if result != "data" {
		t.Errorf("Expected result 'data', got: %s", result)
	}

	// Test case 4: normal template execution
	tmpl := template.Must(template.New("test").Parse("Hello, {{.}}!"))
	result, err = buildTemplateValue(tmpl, "test", "World")
	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}
	if result != "Hello, World!" {
		t.Errorf("Expected result 'Hello, World!', got: %s", result)
	}

	// Test case 5: template execution error
	tmpl = template.Must(template.New("test").Parse("Hello, {{.World}}!"))
	result, err = buildTemplateValue(tmpl, "test", TestData{Name: "World"})
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	expectedError := "template: test:1:9: executing \"test\" at <.World>: can't evaluate field World in type fir.TestData"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got: %v", expectedError, err)
	}
	if result != "" {
		t.Errorf("Expected empty result, got: %s", result)
	}
}
func TestTargetOrClassName(t *testing.T) {
	// Test case 1: target is not nil and not empty
	target := new(string)
	*target = "target"
	className := "class"
	result := targetOrClassName(target, className)
	expected := "target"
	if *result != expected {
		t.Errorf("Expected result '%s', got: %s", expected, *result)
	}

	// Test case 2: target is nil
	target = nil
	result = targetOrClassName(target, className)
	expected = ".class"
	if *result != expected {
		t.Errorf("Expected result '%s', got: %s", expected, *result)
	}

	// Test case 3: target is empty
	target = new(string)
	result = targetOrClassName(target, className)
	expected = ".class"
	if *result != expected {
		t.Errorf("Expected result '%s', got: %s", expected, *result)
	}
}
