package services

import (
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)

// DefaultTemplateService is the default implementation of TemplateService
type DefaultTemplateService struct {
	cache         *cache.Cache
	cacheEnabled  bool
	defaultFuncMap template.FuncMap
	mutex         sync.RWMutex
}

// NewDefaultTemplateService creates a new default template service
func NewDefaultTemplateService(cacheEnabled bool) *DefaultTemplateService {
	return &DefaultTemplateService{
		cache:        cache.New(5*time.Minute, 10*time.Minute),
		cacheEnabled: cacheEnabled,
		defaultFuncMap: make(template.FuncMap),
		mutex:        sync.RWMutex{},
	}
}

// LoadTemplate loads a template with the given configuration
func (s *DefaultTemplateService) LoadTemplate(config TemplateConfig) (*template.Template, error) {
	s.mutex.RLock()
	cacheKey := s.buildCacheKey(config)
	
	// Check cache first if enabled
	if s.cacheEnabled && !config.CacheDisabled {
		if cached, found := s.cache.Get(cacheKey); found {
			s.mutex.RUnlock()
			if tmpl, ok := cached.(*template.Template); ok {
				return tmpl.Clone()
			}
		}
	}
	s.mutex.RUnlock()

	// Load and parse template
	tmpl, err := s.parseTemplate(config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	// Cache the template if caching is enabled
	if s.cacheEnabled && !config.CacheDisabled {
		s.mutex.Lock()
		s.cache.Set(cacheKey, tmpl, cache.DefaultExpiration)
		s.mutex.Unlock()
	}

	return tmpl.Clone()
}

// ParseTemplate parses template content with partials and layout
func (s *DefaultTemplateService) ParseTemplate(content, layout string, partials []string, funcMap template.FuncMap) (*template.Template, error) {
	// Merge function maps
	mergedFuncMap := s.mergeFuncMaps(s.defaultFuncMap, funcMap)
	
	// Create new template
	tmpl := template.New("content").Funcs(mergedFuncMap)
	
	// Parse layout first if provided
	if layout != "" {
		var err error
		tmpl, err = tmpl.ParseFiles(layout)
		if err != nil {
			return nil, fmt.Errorf("failed to parse layout %s: %w", layout, err)
		}
	}
	
	// Parse partials
	if len(partials) > 0 {
		var err error
		tmpl, err = tmpl.ParseFiles(partials...)
		if err != nil {
			return nil, fmt.Errorf("failed to parse partials: %w", err)
		}
	}
	
	// Parse main content
	if content != "" {
		// Determine if content is a file path or inline content
		if s.isFilePath(content) {
			var err error
			tmpl, err = tmpl.ParseFiles(content)
			if err != nil {
				return nil, fmt.Errorf("failed to parse content file %s: %w", content, err)
			}
		} else {
			// Inline content
			var err error
			tmpl, err = tmpl.Parse(content)
			if err != nil {
				return nil, fmt.Errorf("failed to parse inline content: %w", err)
			}
		}
	}
	
	return tmpl, nil
}

// GetTemplate retrieves a cached template or loads it if not cached
func (s *DefaultTemplateService) GetTemplate(routeID string, templateType TemplateType) (*template.Template, error) {
	// This would typically be implemented with a template registry
	// For now, return an error indicating the template needs to be loaded via LoadTemplate
	return nil, fmt.Errorf("template lookup by routeID not implemented yet - use LoadTemplate instead")
}

// ClearCache clears the template cache
func (s *DefaultTemplateService) ClearCache() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.cache.Flush()
	return nil
}

// SetCacheEnabled enables or disables template caching
func (s *DefaultTemplateService) SetCacheEnabled(enabled bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.cacheEnabled = enabled
	if !enabled {
		s.cache.Flush()
	}
}

// SetDefaultFuncMap sets the default function map for templates
func (s *DefaultTemplateService) SetDefaultFuncMap(funcMap template.FuncMap) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.defaultFuncMap = funcMap
}

// parseTemplate parses a template based on the configuration
func (s *DefaultTemplateService) parseTemplate(config TemplateConfig) (*template.Template, error) {
	// Merge function maps
	mergedFuncMap := s.mergeFuncMaps(s.defaultFuncMap, config.FuncMap)
	
	// Create base template
	tmpl := template.New(s.getTemplateName(config.ContentPath)).Funcs(mergedFuncMap)
	
	// Parse layout first if provided
	if config.LayoutPath != "" {
		var err error
		tmpl, err = tmpl.ParseFiles(config.LayoutPath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse layout %s: %w", config.LayoutPath, err)
		}
	}
	
	// Parse partials
	if len(config.PartialPaths) > 0 {
		var err error
		tmpl, err = tmpl.ParseFiles(config.PartialPaths...)
		if err != nil {
			return nil, fmt.Errorf("failed to parse partials: %w", err)
		}
	}
	
	// Parse main content
	if config.ContentPath != "" {
		if s.isFilePath(config.ContentPath) {
			var err error
			tmpl, err = tmpl.ParseFiles(config.ContentPath)
			if err != nil {
				return nil, fmt.Errorf("failed to parse content file %s: %w", config.ContentPath, err)
			}
		} else {
			// Inline content
			var err error
			tmpl, err = tmpl.Parse(config.ContentPath)
			if err != nil {
				return nil, fmt.Errorf("failed to parse inline content: %w", err)
			}
		}
	}
	
	return tmpl, nil
}

// buildCacheKey builds a cache key for the template configuration
func (s *DefaultTemplateService) buildCacheKey(config TemplateConfig) string {
	var parts []string
	
	parts = append(parts, config.RouteID)
	parts = append(parts, config.ContentPath)
	parts = append(parts, config.LayoutPath)
	parts = append(parts, strings.Join(config.PartialPaths, ","))
	parts = append(parts, strings.Join(config.Extensions, ","))
	parts = append(parts, config.LayoutContentName)
	
	return strings.Join(parts, "|")
}

// getTemplateName extracts a template name from a file path
func (s *DefaultTemplateService) getTemplateName(path string) string {
	if path == "" {
		return "main"
	}
	
	if s.isFilePath(path) {
		return filepath.Base(path)
	}
	
	// For inline content, use a default name
	return "inline"
}

// isFilePath determines if a string is a file path or inline content
func (s *DefaultTemplateService) isFilePath(content string) bool {
	// If it contains HTML tags or template directives, it's inline content
	if strings.Contains(content, "<") || strings.Contains(content, "{{") {
		return false
	}
	
	// If it contains path separators, it's likely a file path
	if strings.Contains(content, "/") || strings.Contains(content, "\\") {
		return true
	}
	
	// If it has a file extension and no spaces, it's likely a file path
	if strings.Contains(content, ".") && !strings.Contains(content, " ") {
		// Check if it looks like a filename (has extension at the end)
		lastDot := strings.LastIndex(content, ".")
		if lastDot > 0 && lastDot < len(content)-1 {
			// Check if the part after the last dot looks like an extension (no spaces, reasonable length)
			extension := content[lastDot+1:]
			if len(extension) <= 10 && !strings.Contains(extension, " ") {
				return true
			}
		}
	}
	
	return false
}

// mergeFuncMaps merges multiple function maps, with later maps taking precedence
func (s *DefaultTemplateService) mergeFuncMaps(maps ...template.FuncMap) template.FuncMap {
	result := make(template.FuncMap)
	
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	
	return result
}

// GoTemplateEngine is an implementation of TemplateEngine using Go's html/template
type GoTemplateEngine struct {
	funcMap template.FuncMap
}

// NewGoTemplateEngine creates a new Go template engine
func NewGoTemplateEngine() *GoTemplateEngine {
	return &GoTemplateEngine{
		funcMap: make(template.FuncMap),
	}
}

// ParseFiles parses template files into a template handle
func (e *GoTemplateEngine) ParseFiles(files ...string) (TemplateHandle, error) {
	tmpl := template.New(filepath.Base(files[0])).Funcs(e.funcMap)
	tmpl, err := tmpl.ParseFiles(files...)
	if err != nil {
		return nil, err
	}
	return &GoTemplateHandle{template: tmpl}, nil
}

// ParseContent parses template content string into a template handle
func (e *GoTemplateEngine) ParseContent(content string) (TemplateHandle, error) {
	tmpl := template.New("content").Funcs(e.funcMap)
	tmpl, err := tmpl.Parse(content)
	if err != nil {
		return nil, err
	}
	return &GoTemplateHandle{template: tmpl}, nil
}

// Execute executes a template with data and returns the result
func (e *GoTemplateEngine) Execute(tmpl TemplateHandle, data interface{}) ([]byte, error) {
	return tmpl.Execute(data)
}

// AddFuncMap adds function map to the template engine
func (e *GoTemplateEngine) AddFuncMap(funcMap template.FuncMap) {
	for k, v := range funcMap {
		e.funcMap[k] = v
	}
}

// GoTemplateHandle implements TemplateHandle for Go templates
type GoTemplateHandle struct {
	template *template.Template
}

// Execute executes the template with data
func (h *GoTemplateHandle) Execute(data interface{}) ([]byte, error) {
	var buf strings.Builder
	err := h.template.Execute(&buf, data)
	if err != nil {
		return nil, err
	}
	return []byte(buf.String()), nil
}

// Clone creates a copy of the template
func (h *GoTemplateHandle) Clone() (TemplateHandle, error) {
	cloned, err := h.template.Clone()
	if err != nil {
		return nil, err
	}
	return &GoTemplateHandle{template: cloned}, nil
}

// InMemoryTemplateCache is an in-memory implementation of TemplateCache
type InMemoryTemplateCache struct {
	cache *cache.Cache
	stats TemplateCacheStats
	mutex sync.RWMutex
}

// NewInMemoryTemplateCache creates a new in-memory template cache
func NewInMemoryTemplateCache(defaultExpiration, cleanupInterval time.Duration) *InMemoryTemplateCache {
	return &InMemoryTemplateCache{
		cache: cache.New(defaultExpiration, cleanupInterval),
		stats: TemplateCacheStats{},
		mutex: sync.RWMutex{},
	}
}

// Get retrieves a template from cache
func (c *InMemoryTemplateCache) Get(key string) (TemplateHandle, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	if value, found := c.cache.Get(key); found {
		c.stats.Hits++
		c.updateHitRatio()
		if tmpl, ok := value.(TemplateHandle); ok {
			return tmpl, true
		}
	}
	
	c.stats.Misses++
	c.updateHitRatio()
	return nil, false
}

// Set stores a template in cache
func (c *InMemoryTemplateCache) Set(key string, template TemplateHandle) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.cache.Set(key, template, cache.DefaultExpiration)
	c.stats.Entries = c.cache.ItemCount()
}

// Delete removes a template from cache
func (c *InMemoryTemplateCache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.cache.Delete(key)
	c.stats.Entries = c.cache.ItemCount()
}

// Clear clears all cached templates
func (c *InMemoryTemplateCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.cache.Flush()
	c.stats.Entries = 0
}

// Stats returns cache statistics
func (c *InMemoryTemplateCache) Stats() TemplateCacheStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.stats
}

// updateHitRatio updates the hit ratio calculation
func (c *InMemoryTemplateCache) updateHitRatio() {
	total := c.stats.Hits + c.stats.Misses
	if total > 0 {
		c.stats.HitRatio = float64(c.stats.Hits) / float64(total)
	}
}
