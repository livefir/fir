package templateengine

import (
	"html/template"
	"io"
	"sync"
)

// GoTemplate wraps Go's html/template.Template to implement our Template interface.
type GoTemplate struct {
	tmpl *template.Template
}

// NewGoTemplate creates a new GoTemplate wrapper.
func NewGoTemplate(tmpl *template.Template) *GoTemplate {
	return &GoTemplate{tmpl: tmpl}
}

// Execute implements Template interface.
func (gt *GoTemplate) Execute(wr io.Writer, data interface{}) error {
	if gt.tmpl == nil {
		return ErrInvalidTemplate
	}
	return gt.tmpl.Execute(wr, data)
}

// ExecuteTemplate implements Template interface.
func (gt *GoTemplate) ExecuteTemplate(wr io.Writer, name string, data interface{}) error {
	if gt.tmpl == nil {
		return ErrInvalidTemplate
	}
	return gt.tmpl.ExecuteTemplate(wr, name, data)
}

// Name implements Template interface.
func (gt *GoTemplate) Name() string {
	if gt.tmpl == nil {
		return ""
	}
	return gt.tmpl.Name()
}

// Templates implements Template interface.
func (gt *GoTemplate) Templates() []Template {
	if gt.tmpl == nil {
		return nil
	}

	templates := gt.tmpl.Templates()
	result := make([]Template, len(templates))
	for i, t := range templates {
		result[i] = &GoTemplate{tmpl: t}
	}
	return result
}

// Clone implements Template interface.
func (gt *GoTemplate) Clone() (Template, error) {
	if gt.tmpl == nil {
		return nil, ErrInvalidTemplate
	}

	cloned, err := gt.tmpl.Clone()
	if err != nil {
		return nil, err
	}

	return &GoTemplate{tmpl: cloned}, nil
}

// Funcs implements Template interface.
func (gt *GoTemplate) Funcs(funcMap template.FuncMap) Template {
	if gt.tmpl == nil {
		return gt
	}

	return &GoTemplate{tmpl: gt.tmpl.Funcs(funcMap)}
}

// Lookup implements Template interface.
func (gt *GoTemplate) Lookup(name string) Template {
	if gt.tmpl == nil {
		return nil
	}

	found := gt.tmpl.Lookup(name)
	if found == nil {
		return nil
	}

	return &GoTemplate{tmpl: found}
}

// GetUnderlyingTemplate returns the underlying html/template.Template.
// This is useful for integration with existing code that expects *template.Template.
func (gt *GoTemplate) GetUnderlyingTemplate() *template.Template {
	return gt.tmpl
}

// InMemoryTemplateCache provides a simple in-memory cache implementation.
type InMemoryTemplateCache struct {
	cache map[string]Template
	mutex sync.RWMutex
}

// NewInMemoryTemplateCache creates a new in-memory template cache.
func NewInMemoryTemplateCache() *InMemoryTemplateCache {
	return &InMemoryTemplateCache{
		cache: make(map[string]Template),
	}
}

// Set implements TemplateCache interface.
func (c *InMemoryTemplateCache) Set(key string, template Template) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache[key] = template
	return nil
}

// Get implements TemplateCache interface.
func (c *InMemoryTemplateCache) Get(key string) (Template, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	template, exists := c.cache[key]
	return template, exists
}

// Delete implements TemplateCache interface.
func (c *InMemoryTemplateCache) Delete(key string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, exists := c.cache[key]; exists {
		delete(c.cache, key)
		return true
	}
	return false
}

// Clear implements TemplateCache interface.
func (c *InMemoryTemplateCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache = make(map[string]Template)
}

// Size implements TemplateCache interface.
func (c *InMemoryTemplateCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.cache)
}

// Keys implements TemplateCache interface.
func (c *InMemoryTemplateCache) Keys() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	keys := make([]string, 0, len(c.cache))
	for k := range c.cache {
		keys = append(keys, k)
	}
	return keys
}
