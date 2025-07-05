package templateengine

import (
	"html/template"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// BenchmarkGoTemplateEngine_LoadTemplate benchmarks template loading performance
func BenchmarkGoTemplateEngine_LoadTemplate(b *testing.B) {
	engine := NewGoTemplateEngine()
	config := TemplateConfig{
		ContentPath:     "test-content",
		ContentTemplate: "<h1>Hello {{.Name}}</h1><p>{{.Content}}</p>",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.LoadTemplate(config)
		if err != nil {
			b.Fatalf("Failed to load template: %v", err)
		}
	}
}

// BenchmarkGoTemplateEngine_LoadTemplateWithCache benchmarks cached template loading
func BenchmarkGoTemplateEngine_LoadTemplateWithCache(b *testing.B) {
	engine := NewGoTemplateEngine()
	config := TemplateConfig{
		ContentPath:     "cached-test-content",
		ContentTemplate: "<h1>Hello {{.Name}}</h1><p>{{.Content}}</p>",
	}

	// Pre-load template to cache it
	_, err := engine.LoadTemplate(config)
	if err != nil {
		b.Fatalf("Failed to pre-load template: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.LoadTemplate(config)
		if err != nil {
			b.Fatalf("Failed to load cached template: %v", err)
		}
	}
}

// BenchmarkGoTemplateEngine_Render benchmarks template rendering performance
func BenchmarkGoTemplateEngine_Render(b *testing.B) {
	engine := NewGoTemplateEngine()
	config := TemplateConfig{
		ContentPath:     "render-test-content",
		ContentTemplate: "<h1>Hello {{.Name}}</h1><p>{{.Content}}</p><div>{{range .Items}}<span>{{.}}</span>{{end}}</div>",
	}

	tmpl, err := engine.LoadTemplate(config)
	if err != nil {
		b.Fatalf("Failed to load template: %v", err)
	}

	data := map[string]interface{}{
		"Name":    "Performance Test",
		"Content": "This is a benchmark test for template rendering performance.",
		"Items":   []string{"item1", "item2", "item3", "item4", "item5"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf strings.Builder
		err := engine.Render(tmpl, data, &buf)
		if err != nil {
			b.Fatalf("Failed to render template: %v", err)
		}
	}
}

// BenchmarkGoTemplateEngine_RenderWithContext benchmarks context-aware rendering
func BenchmarkGoTemplateEngine_RenderWithContext(b *testing.B) {
	engine := NewGoTemplateEngine()
	config := TemplateConfig{
		ContentPath:     "context-render-test",
		ContentTemplate: "<h1>{{customFunc}}</h1><p>{{.Content}}</p>",
		FuncMap: template.FuncMap{
			"customFunc": func() string { return "Custom Function Result" },
		},
	}

	ctx := TemplateContext{
		FuncMap: template.FuncMap{
			"customFunc": func() string { return "Custom Function Result" },
		},
	}

	tmpl, err := engine.LoadTemplateWithContext(config, ctx)
	if err != nil {
		b.Fatalf("Failed to load template with context: %v", err)
	}

	data := map[string]interface{}{
		"Content": "Context-aware rendering test",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf strings.Builder
		err := engine.RenderWithContext(tmpl, ctx, data, &buf)
		if err != nil {
			b.Fatalf("Failed to render template with context: %v", err)
		}
	}
}

// BenchmarkGoTemplateEngine_ConcurrentAccess benchmarks concurrent template access
func BenchmarkGoTemplateEngine_ConcurrentAccess(b *testing.B) {
	engine := NewGoTemplateEngine()
	configs := make([]TemplateConfig, 10)

	// Create multiple template configurations
	for i := 0; i < 10; i++ {
		configs[i] = TemplateConfig{
			ContentPath:     "concurrent-test-" + strconv.Itoa(i),
			ContentTemplate: "<h1>Template " + strconv.Itoa(i) + "</h1><p>{{.Content}}</p>",
		}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			config := configs[i%len(configs)]
			_, err := engine.LoadTemplate(config)
			if err != nil {
				b.Errorf("Failed to load template concurrently: %v", err)
			}
			i++
		}
	})
}

// BenchmarkGoTemplateEngine_EventTemplateExtraction benchmarks event template extraction
func BenchmarkGoTemplateEngine_EventTemplateExtraction(b *testing.B) {
	engine := NewGoTemplateEngine()
	config := TemplateConfig{
		ContentPath: "event-extraction-test",
		ContentTemplate: `
			<div>
				<button @fir:increment:ok>Increment</button>
				<button @fir:decrement:ok>Decrement</button>
				<template @fir:increment:ok>
					<p>Count: {{.count}}</p>
				</template>
				<template @fir:decrement:ok>
					<p>Count: {{.count}}</p>
				</template>
				<template @fir:reset:ok>
					<p>Reset to 0</p>
				</template>
			</div>
		`,
	}

	tmpl, err := engine.LoadTemplate(config)
	if err != nil {
		b.Fatalf("Failed to load template: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.ExtractEventTemplates(tmpl)
		if err != nil {
			b.Fatalf("Failed to extract event templates: %v", err)
		}
	}
}

// BenchmarkGoTemplateEngine_CacheOperations benchmarks cache operations
func BenchmarkGoTemplateEngine_CacheOperations(b *testing.B) {
	engine := NewGoTemplateEngine()

	// Create templates for caching
	templates := make([]Template, 100)
	for i := 0; i < 100; i++ {
		config := TemplateConfig{
			ContentPath:     "cache-test-" + strconv.Itoa(i),
			ContentTemplate: "<p>Template " + strconv.Itoa(i) + "</p>",
		}
		tmpl, err := engine.LoadTemplate(config)
		if err != nil {
			b.Fatalf("Failed to create template %d: %v", i, err)
		}
		templates[i] = tmpl
	}

	b.Run("CacheTemplate", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			engine.CacheTemplate("bench-"+strconv.Itoa(i%100), templates[i%100])
		}
	})

	b.Run("GetCachedTemplate", func(b *testing.B) {
		// Pre-cache some templates
		for i := 0; i < 50; i++ {
			engine.CacheTemplate("get-test-"+strconv.Itoa(i), templates[i])
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, found := engine.GetCachedTemplate("get-test-" + strconv.Itoa(i%50))
			if !found {
				b.Errorf("Expected to find cached template")
			}
		}
	})

	b.Run("ClearCache", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Add some templates to cache
			for j := 0; j < 10; j++ {
				engine.CacheTemplate("clear-test-"+strconv.Itoa(j), templates[j])
			}
			// Clear cache
			engine.ClearCache()
		}
	})
}

// TestGoTemplateEngine_MemoryUsage tests memory efficiency
func TestGoTemplateEngine_MemoryUsage(t *testing.T) {
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	engine := NewGoTemplateEngine()

	// Load many templates to test memory usage
	for i := 0; i < 1000; i++ {
		config := TemplateConfig{
			ContentPath:     "memory-test-" + strconv.Itoa(i%100),
			ContentTemplate: "<h1>Template " + strconv.Itoa(i%100) + "</h1><p>{{.content}}</p>",
		}
		_, err := engine.LoadTemplate(config)
		if err != nil {
			t.Fatalf("Failed to load template %d: %v", i, err)
		}
	}

	runtime.GC()
	runtime.ReadMemStats(&m2)

	// Calculate memory usage
	memUsed := m2.TotalAlloc - m1.TotalAlloc
	avgMemPerTemplate := memUsed / 1000

	t.Logf("Total memory used: %d bytes", memUsed)
	t.Logf("Average memory per template: %d bytes", avgMemPerTemplate)

	// Basic memory usage validation (templates should be reasonably efficient)
	if avgMemPerTemplate > 50000 { // 50KB per template seems excessive
		t.Errorf("Memory usage per template is too high: %d bytes", avgMemPerTemplate)
	}
}

// BenchmarkTemplateEngine_vs_Legacy simulates legacy vs new template engine comparison
func BenchmarkTemplateEngine_vs_Legacy(b *testing.B) {
	// Simulate legacy template parsing (simple approach)
	b.Run("Legacy", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tmpl := template.New("legacy-test")
			_, err := tmpl.Parse("<h1>Hello {{.Name}}</h1><p>{{.Content}}</p>")
			if err != nil {
				b.Fatalf("Failed to parse legacy template: %v", err)
			}

			data := map[string]interface{}{
				"Name":    "Test",
				"Content": "Legacy template content",
			}

			var buf strings.Builder
			err = tmpl.Execute(&buf, data)
			if err != nil {
				b.Fatalf("Failed to execute legacy template: %v", err)
			}
		}
	})

	// New template engine approach
	b.Run("TemplateEngine", func(b *testing.B) {
		engine := NewGoTemplateEngine()
		config := TemplateConfig{
			ContentPath:     "new-engine-test",
			ContentTemplate: "<h1>Hello {{.Name}}</h1><p>{{.Content}}</p>",
		}

		for i := 0; i < b.N; i++ {
			tmpl, err := engine.LoadTemplate(config)
			if err != nil {
				b.Fatalf("Failed to load template: %v", err)
			}

			data := map[string]interface{}{
				"Name":    "Test",
				"Content": "New template engine content",
			}

			var buf strings.Builder
			err = engine.Render(tmpl, data, &buf)
			if err != nil {
				b.Fatalf("Failed to render template: %v", err)
			}
		}
	})

	// New template engine with caching
	b.Run("TemplateEngineWithCache", func(b *testing.B) {
		engine := NewGoTemplateEngine()
		config := TemplateConfig{
			ContentPath:     "cached-engine-test",
			ContentTemplate: "<h1>Hello {{.Name}}</h1><p>{{.Content}}</p>",
		}

		// Pre-load template for caching
		_, err := engine.LoadTemplate(config)
		if err != nil {
			b.Fatalf("Failed to pre-load template: %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tmpl, err := engine.LoadTemplate(config)
			if err != nil {
				b.Fatalf("Failed to load cached template: %v", err)
			}

			data := map[string]interface{}{
				"Name":    "Test",
				"Content": "Cached template engine content",
			}

			var buf strings.Builder
			err = engine.Render(tmpl, data, &buf)
			if err != nil {
				b.Fatalf("Failed to render cached template: %v", err)
			}
		}
	})
}

// BenchmarkConcurrentTemplateLoading tests concurrent template loading under high load
func BenchmarkConcurrentTemplateLoading(b *testing.B) {
	engine := NewGoTemplateEngine()
	numWorkers := 100

	b.Run("HighConcurrency", func(b *testing.B) {
		b.SetParallelism(numWorkers)
		b.RunParallel(func(pb *testing.PB) {
			workerID := 0
			for pb.Next() {
				config := TemplateConfig{
					ContentPath:     "concurrent-worker-" + strconv.Itoa(workerID%50),
					ContentTemplate: "<h1>Worker {{.ID}}</h1><p>{{.Content}}</p>",
				}

				_, err := engine.LoadTemplate(config)
				if err != nil {
					b.Errorf("Worker %d failed to load template: %v", workerID, err)
				}
				workerID++
			}
		})
	})
}
