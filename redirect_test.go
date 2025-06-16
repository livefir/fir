package fir

import (
	"testing"

	"github.com/stretchr/testify/require"
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
			expectedHTML: `<button @fir:delete:ok="$fir.redirect(&#39;/&#39;)">Delete</button>`,
			wantErr:      false,
		},
		{
			name:         "x-fir-redirect with URL parameter",
			inputHTML:    `<button x-fir-redirect:home="delete:ok">Delete</button>`,
			expectedHTML: `<button @fir:delete:ok="$fir.redirect(&#39;/home&#39;)">Delete</button>`,
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

// TestRedirectActionHandler tests the RedirectActionHandler implementation
func TestRedirectActionHandler(t *testing.T) {
	handler := &RedirectActionHandler{}

	// Test basic properties
	require.Equal(t, "redirect", handler.Name())
	require.Equal(t, 90, handler.Precedence())

	// Test translation
	tests := []struct {
		name     string
		params   []string
		value    string
		expected string
		wantErr  bool
	}{
		{
			name:     "Default redirect (no params)",
			params:   []string{},
			value:    "delete",
			expected: `@fir:delete:ok="$fir.redirect('/')"`,
			wantErr:  false,
		},
		{
			name:     "Redirect with URL parameter",
			params:   []string{"home"},
			value:    "delete",
			expected: `@fir:delete:ok="$fir.redirect('/home')"`,
			wantErr:  false,
		},
		{
			name:     "Redirect with absolute path",
			params:   []string{"/dashboard"},
			value:    "submit",
			expected: `@fir:submit:ok="$fir.redirect('/dashboard')"`,
			wantErr:  false,
		},
		{
			name:     "Event with state",
			params:   []string{"success"},
			value:    "save:ok",
			expected: `@fir:save:ok="$fir.redirect('/success')"`,
			wantErr:  false,
		},
		{
			name:     "Event with modifier",
			params:   []string{"login"},
			value:    "auth.prevent",
			expected: `@fir:auth:ok.prevent="$fir.redirect('/login')"`,
			wantErr:  false,
		},
		{
			name:     "Multiple events",
			params:   []string{"complete"},
			value:    "save:ok,update:done",
			expected: `@fir:[save:ok,update:done]="$fir.redirect('/complete')"`,
			wantErr:  false,
		},
		{
			name:     "Empty parameter (fallback to default)",
			params:   []string{""},
			value:    "click",
			expected: `@fir:click:ok="$fir.redirect('/')"`,
			wantErr:  false,
		},
		{
			name:     "Whitespace parameter (fallback to default)",
			params:   []string{"  "},
			value:    "click",
			expected: `@fir:click:ok="$fir.redirect('/')"`,
			wantErr:  false,
		},
		{
			name:     "Event with target and action (ignored)",
			params:   []string{"profile"},
			value:    "update->form=>doUpdate",
			expected: `@fir:update:ok="$fir.redirect('/profile')"`,
			wantErr:  false,
		},
		{
			name:     "Empty value",
			params:   []string{"home"},
			value:    "",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := ActionInfo{
				AttrName:   "x-fir-redirect",
				ActionName: "redirect",
				Params:     tt.params,
				Value:      tt.value,
			}

			result, err := handler.Translate(info, nil)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}
