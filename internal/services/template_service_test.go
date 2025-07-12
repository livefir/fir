package services

import (
	"html/template"
	"strings"
	"testing"
	"time"
)

func TestDefaultTemplateService_LoadTemplate(t *testing.T) {
	service := NewDefaultTemplateService(true)

	tests := []struct {
		name    string
		config  TemplateConfig
		wantErr bool
	}{
		{
			name: "inline content template",
			config: TemplateConfig{
				ContentPath: "<h1>{{.Title}}</h1>",
				RouteID:     "test-route",
			},
			wantErr: false,
		},
		{
			name: "template with function map",
			config: TemplateConfig{
				ContentPath: "<h1>{{upper .Title}}</h1>",
				RouteID:     "test-route-func",
				FuncMap: template.FuncMap{
					"upper": strings.ToUpper,
				},
			},
			wantErr: false,
		},
		{
			name: "empty content path",
			config: TemplateConfig{
				ContentPath: "",
				RouteID:     "empty-route",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := service.LoadTemplate(tt.config)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tmpl == nil {
				t.Errorf("expected template but got nil")
			}
		})
	}
}

func TestDefaultTemplateService_ParseTemplate(t *testing.T) {
	service := NewDefaultTemplateService(true)

	tests := []struct {
		name     string
		content  string
		layout   string
		partials []string
		funcMap  template.FuncMap
		wantErr  bool
	}{
		{
			name:    "simple inline template",
			content: "<h1>{{.Title}}</h1>",
			wantErr: false,
		},
		{
			name:    "template with function map",
			content: "<h1>{{upper .Title}}</h1>",
			funcMap: template.FuncMap{
				"upper": strings.ToUpper,
			},
			wantErr: false,
		},
		{
			name:    "empty content",
			content: "",
			wantErr: false,
		},
		{
			name:    "invalid template syntax",
			content: "<h1>{{.Title</h1>", // Missing closing }}
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := service.ParseTemplate(tt.content, tt.layout, tt.partials, tt.funcMap)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tmpl == nil {
				t.Errorf("expected template but got nil")
			}
		})
	}
}

func TestDefaultTemplateService_Caching(t *testing.T) {
	service := NewDefaultTemplateService(true)

	config := TemplateConfig{
		ContentPath: "<h1>{{.Title}}</h1>",
		RouteID:     "cache-test",
	}

	// Load template first time
	tmpl1, err := service.LoadTemplate(config)
	if err != nil {
		t.Fatalf("unexpected error loading template: %v", err)
	}

	// Load template second time (should come from cache)
	tmpl2, err := service.LoadTemplate(config)
	if err != nil {
		t.Fatalf("unexpected error loading cached template: %v", err)
	}

	// Both templates should be valid
	if tmpl1 == nil || tmpl2 == nil {
		t.Errorf("expected valid templates, got nil")
	}

	// Test cache clearing
	err = service.ClearCache()
	if err != nil {
		t.Errorf("unexpected error clearing cache: %v", err)
	}

	// Load template after cache clear
	tmpl3, err := service.LoadTemplate(config)
	if err != nil {
		t.Fatalf("unexpected error loading template after cache clear: %v", err)
	}

	if tmpl3 == nil {
		t.Errorf("expected valid template after cache clear, got nil")
	}
}

func TestDefaultTemplateService_SetCacheEnabled(t *testing.T) {
	service := NewDefaultTemplateService(true)

	// Test disabling cache
	service.SetCacheEnabled(false)

	// Test enabling cache
	service.SetCacheEnabled(true)

	// Should not error
}

func TestDefaultTemplateService_GetTemplate(t *testing.T) {
	service := NewDefaultTemplateService(true)

	// This should return an error as it's not implemented yet
	_, err := service.GetTemplate("test-route", StandardTemplate)
	if err == nil {
		t.Errorf("expected error for unimplemented GetTemplate, got none")
	}
}

func TestGoTemplateEngine_ParseContent(t *testing.T) {
	engine := NewGoTemplateEngine()

	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "valid template",
			content: "<h1>{{.Title}}</h1>",
			wantErr: false,
		},
		{
			name:    "invalid template",
			content: "<h1>{{.Title</h1>", // Missing closing }}
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handle, err := engine.ParseContent(tt.content)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if handle == nil {
				t.Errorf("expected template handle but got nil")
			}
		})
	}
}

func TestGoTemplateEngine_Execute(t *testing.T) {
	engine := NewGoTemplateEngine()

	handle, err := engine.ParseContent("<h1>{{.Title}}</h1>")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	data := map[string]interface{}{
		"Title": "Test Title",
	}

	result, err := engine.Execute(handle, data)
	if err != nil {
		t.Errorf("unexpected error executing template: %v", err)
		return
	}

	expected := "<h1>Test Title</h1>"
	if string(result) != expected {
		t.Errorf("expected %s, got %s", expected, string(result))
	}
}

func TestGoTemplateHandle_Execute(t *testing.T) {
	engine := NewGoTemplateEngine()

	handle, err := engine.ParseContent("<h1>{{.Title}}</h1>")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	data := map[string]interface{}{
		"Title": "Test Title",
	}

	result, err := handle.Execute(data)
	if err != nil {
		t.Errorf("unexpected error executing template: %v", err)
		return
	}

	expected := "<h1>Test Title</h1>"
	if string(result) != expected {
		t.Errorf("expected %s, got %s", expected, string(result))
	}
}

func TestGoTemplateHandle_Clone(t *testing.T) {
	engine := NewGoTemplateEngine()

	handle, err := engine.ParseContent("<h1>{{.Title}}</h1>")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	cloned, err := handle.Clone()
	if err != nil {
		t.Errorf("unexpected error cloning template: %v", err)
		return
	}

	if cloned == nil {
		t.Errorf("expected cloned template handle but got nil")
	}

	// Both should work independently
	data := map[string]interface{}{
		"Title": "Original",
	}

	result1, err := handle.Execute(data)
	if err != nil {
		t.Errorf("unexpected error executing original: %v", err)
	}

	data["Title"] = "Cloned"
	result2, err := cloned.Execute(data)
	if err != nil {
		t.Errorf("unexpected error executing cloned: %v", err)
	}

	if len(result1) == 0 || len(result2) == 0 {
		t.Errorf("expected non-empty results from both templates")
	}
}

func TestInMemoryTemplateCache(t *testing.T) {
	cache := NewInMemoryTemplateCache(5*time.Minute, 10*time.Minute)

	engine := NewGoTemplateEngine()
	handle, err := engine.ParseContent("<h1>{{.Title}}</h1>")
	if err != nil {
		t.Fatalf("failed to create test template: %v", err)
	}

	// Test cache miss
	_, found := cache.Get("test-key")
	if found {
		t.Errorf("expected cache miss but got hit")
	}

	// Test cache set and hit
	cache.Set("test-key", handle)
	retrieved, found := cache.Get("test-key")
	if !found {
		t.Errorf("expected cache hit but got miss")
	}
	if retrieved == nil {
		t.Errorf("expected retrieved template but got nil")
	}

	// Test cache delete
	cache.Delete("test-key")
	_, found = cache.Get("test-key")
	if found {
		t.Errorf("expected cache miss after delete but got hit")
	}

	// Test cache clear
	cache.Set("test-key-1", handle)
	cache.Set("test-key-2", handle)
	cache.Clear()

	_, found1 := cache.Get("test-key-1")
	_, found2 := cache.Get("test-key-2")
	if found1 || found2 {
		t.Errorf("expected cache misses after clear but got hits")
	}

	// Test stats
	stats := cache.Stats()
	if stats.Entries != 0 {
		t.Errorf("expected 0 entries after clear, got %d", stats.Entries)
	}
}
