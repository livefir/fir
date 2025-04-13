package fir

import "testing"

func TestTranslateRenderExpression(t *testing.T) {
	parser, err := getRenderExpressionParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	input := "create:ok->todo,delete:error=>fir.replace;update:pending->done=>archive"
	expected := `@fir:create:ok::todo
@fir:delete:error="fir.replace"
@fir:update:pending::done="archive"`

	output, err := TranslateRenderExpression(parser, input)
	if err != nil {
		t.Fatalf("Translation failed: %v", err)
	}

	if output != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, output)
	}
}
