package fir

import (
	"testing"
)

func TestRedirectAction(t *testing.T) {
	tests := []struct {
		name         string
		inputHTML    string
		expectedHTML string
		wantErr      bool
	}{
		{
			name:         "Basic x-fir-redirect",
			inputHTML:    `<button x-fir-redirect="delete:ok">Delete</button>`,
			expectedHTML: `<button @fir:delete:ok.nohtml="$fir.redirect(&#39;/&#39;)">Delete</button>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-redirect with URL parameter",
			inputHTML:    `<button x-fir-redirect:home="delete:ok">Delete</button>`,
			expectedHTML: `<button @fir:delete:ok.nohtml="$fir.redirect(&#39;/home&#39;)">Delete</button>`,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processRenderAttributes([]byte(tt.inputHTML))
			if (err != nil) != tt.wantErr {
				t.Errorf("processRenderAttributes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && string(result) != tt.expectedHTML {
				t.Errorf("processRenderAttributes() = %v, want %v", string(result), tt.expectedHTML)
			}
		})
	}
}
