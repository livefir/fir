package templateengine

import (
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"
)

// TestStressTemplateEngine_ConcurrentAccess tests template engine under heavy concurrent load
func TestStressTemplateEngine_ConcurrentAccess(t *testing.T) {
	engine := NewGoTemplateEngine()
	numWorkers := 100
	templatesPerWorker := 1000

	var wg sync.WaitGroup
	errors := make(chan error, numWorkers)

	start := time.Now()

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < templatesPerWorker; j++ {
				config := TemplateConfig{
					ContentPath:     fmt.Sprintf("stress-test-%d-%d", workerID, j),
					ContentTemplate: fmt.Sprintf("<h1>Worker %d</h1><p>Template %d</p>", workerID, j),
				}

				_, err := engine.LoadTemplate(config)
				if err != nil {
					select {
					case errors <- fmt.Errorf("worker %d template %d failed: %v", workerID, j, err):
					default:
					}
					return
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	duration := time.Since(start)
	totalTemplates := numWorkers * templatesPerWorker
	templatesPerSecond := float64(totalTemplates) / duration.Seconds()

	// Check for errors
	errorCount := 0
	for err := range errors {
		t.Logf("Error: %v", err)
		errorCount++
	}

	if errorCount > 0 {
		t.Errorf("Got %d errors during stress test", errorCount)
	}

	t.Logf("Stress test completed:")
	t.Logf("  Workers: %d", numWorkers)
	t.Logf("  Templates per worker: %d", templatesPerWorker)
	t.Logf("  Total templates: %d", totalTemplates)
	t.Logf("  Duration: %v", duration)
	t.Logf("  Templates per second: %.2f", templatesPerSecond)

	// Performance validation - should handle at least 1000 templates per second
	if templatesPerSecond < 1000 {
		t.Errorf("Performance too low: %.2f templates/sec, expected at least 1000", templatesPerSecond)
	}
}

// TestStressTemplateEngine_MemoryPressure tests memory usage under load
func TestStressTemplateEngine_MemoryPressure(t *testing.T) {
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	engine := NewGoTemplateEngine()
	numTemplates := 10000

	// Create many unique templates
	for i := 0; i < numTemplates; i++ {
		config := TemplateConfig{
			ContentPath:     "memory-pressure-" + strconv.Itoa(i),
			ContentTemplate: fmt.Sprintf("<h1>Template %d</h1><p>{{.content%d}}</p>", i, i),
		}

		_, err := engine.LoadTemplate(config)
		if err != nil {
			t.Fatalf("Failed to load template %d: %v", i, err)
		}

		// Force GC every 1000 templates
		if i%1000 == 0 {
			runtime.GC()
		}
	}

	runtime.GC()
	runtime.ReadMemStats(&m2)

	memUsed := m2.TotalAlloc - m1.TotalAlloc
	avgMemPerTemplate := memUsed / uint64(numTemplates)

	t.Logf("Memory pressure test completed:")
	t.Logf("  Templates created: %d", numTemplates)
	t.Logf("  Total memory used: %d bytes (%.2f MB)", memUsed, float64(memUsed)/1024/1024)
	t.Logf("  Average memory per template: %d bytes", avgMemPerTemplate)
	t.Logf("  Final heap size: %d bytes (%.2f MB)", m2.HeapInuse, float64(m2.HeapInuse)/1024/1024)

	// Memory usage validation - should be reasonable
	maxMemPerTemplate := uint64(100000) // 100KB per template seems reasonable
	if avgMemPerTemplate > maxMemPerTemplate {
		t.Errorf("Memory usage too high: %d bytes per template, expected less than %d", avgMemPerTemplate, maxMemPerTemplate)
	}

	// Total memory should be reasonable for 10K templates
	maxTotalMem := uint64(500 * 1024 * 1024) // 500MB total seems reasonable for 10K templates
	if memUsed > maxTotalMem {
		t.Errorf("Total memory usage too high: %d bytes, expected less than %d", memUsed, maxTotalMem)
	}
}

// TestStressTemplateEngine_CacheEfficiency tests cache hit rates
func TestStressTemplateEngine_CacheEfficiency(t *testing.T) {
	engine := NewGoTemplateEngine()

	// Create a limited set of templates that will be reused
	baseTemplates := 100
	totalRequests := 10000

	configs := make([]TemplateConfig, baseTemplates)
	for i := 0; i < baseTemplates; i++ {
		configs[i] = TemplateConfig{
			ContentPath:     "cache-efficiency-" + strconv.Itoa(i),
			ContentTemplate: fmt.Sprintf("<h1>Template %d</h1><p>{{.data}}</p>", i),
		}
	}

	start := time.Now()

	// Make many requests that should hit cache frequently
	for i := 0; i < totalRequests; i++ {
		config := configs[i%baseTemplates]
		_, err := engine.LoadTemplate(config)
		if err != nil {
			t.Fatalf("Failed to load template: %v", err)
		}
	}

	duration := time.Since(start)
	requestsPerSecond := float64(totalRequests) / duration.Seconds()

	t.Logf("Cache efficiency test completed:")
	t.Logf("  Base templates: %d", baseTemplates)
	t.Logf("  Total requests: %d", totalRequests)
	t.Logf("  Duration: %v", duration)
	t.Logf("  Requests per second: %.2f", requestsPerSecond)
	t.Logf("  Expected cache hit rate: %.1f%%", float64(totalRequests-baseTemplates)/float64(totalRequests)*100)

	// With caching, should be much faster than without
	// Should handle at least 10,000 requests per second with good caching
	if requestsPerSecond < 10000 {
		t.Errorf("Cache efficiency too low: %.2f requests/sec, expected at least 10,000", requestsPerSecond)
	}
}

// TestStressTemplateEngine_LongRunning tests engine stability over time
func TestStressTemplateEngine_LongRunning(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long-running stress test in short mode")
	}

	engine := NewGoTemplateEngine()
	duration := 30 * time.Second // Run for 30 seconds
	start := time.Now()

	requestCount := 0
	errorCount := 0

	for time.Since(start) < duration {
		config := TemplateConfig{
			ContentPath:     "long-running-" + strconv.Itoa(requestCount%1000),
			ContentTemplate: fmt.Sprintf("<h1>Request %d</h1><p>{{.timestamp}}</p>", requestCount),
		}

		_, err := engine.LoadTemplate(config)
		if err != nil {
			errorCount++
			if errorCount < 10 { // Only log first 10 errors
				t.Logf("Error in request %d: %v", requestCount, err)
			}
		}

		requestCount++

		// Periodically force GC to test stability
		if requestCount%1000 == 0 {
			runtime.GC()
		}
	}

	actualDuration := time.Since(start)
	requestsPerSecond := float64(requestCount) / actualDuration.Seconds()
	errorRate := float64(errorCount) / float64(requestCount) * 100

	t.Logf("Long-running stress test completed:")
	t.Logf("  Duration: %v", actualDuration)
	t.Logf("  Total requests: %d", requestCount)
	t.Logf("  Requests per second: %.2f", requestsPerSecond)
	t.Logf("  Error count: %d", errorCount)
	t.Logf("  Error rate: %.2f%%", errorRate)

	// Error rate should be very low
	if errorRate > 0.1 {
		t.Errorf("Error rate too high: %.2f%%, expected less than 0.1%%", errorRate)
	}

	// Should maintain good performance over time
	if requestsPerSecond < 500 {
		t.Errorf("Performance degraded over time: %.2f requests/sec, expected at least 500", requestsPerSecond)
	}
}

// BenchmarkStressTemplateEngine_HighConcurrency benchmarks very high concurrency
func BenchmarkStressTemplateEngine_HighConcurrency(b *testing.B) {
	engine := NewGoTemplateEngine()

	b.Run("1000Goroutines", func(b *testing.B) {
		b.SetParallelism(1000)
		b.RunParallel(func(pb *testing.PB) {
			templateID := 0
			for pb.Next() {
				config := TemplateConfig{
					ContentPath:     "high-concurrency-" + strconv.Itoa(templateID%100),
					ContentTemplate: fmt.Sprintf("<h1>Template %d</h1>", templateID),
				}

				_, err := engine.LoadTemplate(config)
				if err != nil {
					b.Errorf("Failed to load template: %v", err)
				}
				templateID++
			}
		})
	})
}

// TestStressTemplateEngine_ResourceCleanup tests that resources are properly cleaned up
func TestStressTemplateEngine_ResourceCleanup(t *testing.T) {
	var m1, m2, m3 runtime.MemStats

	// Measure initial memory
	runtime.GC()
	runtime.ReadMemStats(&m1)

	engine := NewGoTemplateEngine()

	// Create many templates
	for i := 0; i < 5000; i++ {
		config := TemplateConfig{
			ContentPath:     "cleanup-test-" + strconv.Itoa(i),
			ContentTemplate: fmt.Sprintf("<h1>Template %d</h1><p>{{.data}}</p>", i),
		}

		_, err := engine.LoadTemplate(config)
		if err != nil {
			t.Fatalf("Failed to load template %d: %v", i, err)
		}
	}

	// Measure memory after creating templates
	runtime.GC()
	runtime.ReadMemStats(&m2)

	// Clear cache and measure memory again
	engine.ClearCache()
	runtime.GC()
	runtime.ReadMemStats(&m3)

	initialMem := m1.HeapInuse
	afterTemplatesMem := m2.HeapInuse
	afterClearMem := m3.HeapInuse

	memoryGrowth := afterTemplatesMem - initialMem
	memoryFreed := afterTemplatesMem - afterClearMem
	cleanupEfficiency := float64(memoryFreed) / float64(memoryGrowth) * 100

	t.Logf("Resource cleanup test completed:")
	t.Logf("  Initial memory: %d bytes (%.2f MB)", initialMem, float64(initialMem)/1024/1024)
	t.Logf("  Memory after templates: %d bytes (%.2f MB)", afterTemplatesMem, float64(afterTemplatesMem)/1024/1024)
	t.Logf("  Memory after cleanup: %d bytes (%.2f MB)", afterClearMem, float64(afterClearMem)/1024/1024)
	t.Logf("  Memory growth: %d bytes (%.2f MB)", memoryGrowth, float64(memoryGrowth)/1024/1024)
	t.Logf("  Memory freed: %d bytes (%.2f MB)", memoryFreed, float64(memoryFreed)/1024/1024)
	t.Logf("  Cleanup efficiency: %.1f%%", cleanupEfficiency)

	// Cleanup should free a significant portion of memory
	if cleanupEfficiency < 50 {
		t.Errorf("Cleanup efficiency too low: %.1f%%, expected at least 50%%", cleanupEfficiency)
	}
}
