package templateengine

import (
	"embed"
	"html/template"
)

// TemplateConfig holds all configuration needed for template loading and parsing.
// This centralizes template configuration and makes it easier to manage different
// template setups across the framework.
type TemplateConfig struct {
	// Template paths and content
	LayoutPath       string // Path to layout template file
	ContentPath      string // Path to content template file
	ErrorLayoutPath  string // Path to error layout template file
	ErrorContentPath string // Path to error content template file

	// Inline content (alternative to file paths)
	LayoutContent        string // Inline layout template content
	ContentTemplate      string // Inline content template
	ErrorLayoutContent   string // Inline error layout content
	ErrorContentTemplate string // Inline error content template

	// Template organization
	LayoutContentName      string   // Name for layout content template (default: "content")
	ErrorLayoutContentName string   // Name for error layout content template
	Partials               []string // Paths to partial template files
	Extensions             []string // File extensions to search for templates

	// Directory and file system configuration
	PublicDir string    // Root directory for template files
	EmbedFS   *embed.FS // Embedded file system for templates

	// Function maps and template configuration
	FuncMap    template.FuncMap                     // Custom template functions
	ReadFile   func(string) (string, []byte, error) // Custom file reading function
	ExistFile  func(string) bool                    // Custom file existence check function
	GetFuncMap func() template.FuncMap              // Function to get dynamic function map
	AddFunc    func(string, interface{})            // Function to add individual functions

	// Caching and performance
	DisableTemplateCache bool // Whether to disable template caching
	DisableMinification  bool // Whether to disable HTML minification

	// Template parsing options
	MissingKeyBehavior string // How to handle missing keys in templates ("zero", "error", "invalid")
	LeftDelim          string // Left template delimiter (default: "{{")
	RightDelim         string // Right template delimiter (default: "}}")

	// Development and debugging
	DevMode    bool // Enable development mode features
	DebugMode  bool // Enable debug logging and error details
	WatchFiles bool // Enable file watching for auto-reload in development
}

// DefaultTemplateConfig returns a TemplateConfig with sensible defaults.
func DefaultTemplateConfig() TemplateConfig {
	return TemplateConfig{
		LayoutContentName:      "content",
		ErrorLayoutContentName: "content",
		Extensions:             []string{".html", ".tmpl", ".gohtml"},
		MissingKeyBehavior:     "zero",
		LeftDelim:              "{{",
		RightDelim:             "}}",
		DevMode:                false,
		DebugMode:              false,
		WatchFiles:             false,
		DisableTemplateCache:   false,
		DisableMinification:    false,
	}
}

// WithLayout sets the layout path and returns the updated config.
func (c TemplateConfig) WithLayout(layoutPath string) TemplateConfig {
	c.LayoutPath = layoutPath
	return c
}

// WithContent sets the content path and returns the updated config.
func (c TemplateConfig) WithContent(contentPath string) TemplateConfig {
	c.ContentPath = contentPath
	return c
}

// WithErrorLayout sets the error layout path and returns the updated config.
func (c TemplateConfig) WithErrorLayout(errorLayoutPath string) TemplateConfig {
	c.ErrorLayoutPath = errorLayoutPath
	return c
}

// WithErrorContent sets the error content path and returns the updated config.
func (c TemplateConfig) WithErrorContent(errorContentPath string) TemplateConfig {
	c.ErrorContentPath = errorContentPath
	return c
}

// WithPartials sets the partial template paths and returns the updated config.
func (c TemplateConfig) WithPartials(partials ...string) TemplateConfig {
	c.Partials = partials
	return c
}

// WithPublicDir sets the public directory and returns the updated config.
func (c TemplateConfig) WithPublicDir(publicDir string) TemplateConfig {
	c.PublicDir = publicDir
	return c
}

// WithFuncMap sets the function map and returns the updated config.
func (c TemplateConfig) WithFuncMap(funcMap template.FuncMap) TemplateConfig {
	c.FuncMap = funcMap
	return c
}

// WithEmbedFS sets the embedded file system and returns the updated config.
func (c TemplateConfig) WithEmbedFS(embedFS *embed.FS) TemplateConfig {
	c.EmbedFS = embedFS
	return c
}

// WithDevMode enables or disables development mode and returns the updated config.
func (c TemplateConfig) WithDevMode(devMode bool) TemplateConfig {
	c.DevMode = devMode
	return c
}

// WithDebugMode enables or disables debug mode and returns the updated config.
func (c TemplateConfig) WithDebugMode(debugMode bool) TemplateConfig {
	c.DebugMode = debugMode
	return c
}

// WithCaching enables or disables template caching and returns the updated config.
func (c TemplateConfig) WithCaching(enableCache bool) TemplateConfig {
	c.DisableTemplateCache = !enableCache
	return c
}

// Validate checks if the template configuration is valid and returns an error if not.
func (c *TemplateConfig) Validate() error {
	// Basic validation - can be extended as needed
	if c.PublicDir == "" && c.EmbedFS == nil &&
		c.LayoutContent == "" && c.ContentTemplate == "" &&
		c.LayoutPath == "" && c.ContentPath == "" {
		// Allow empty config for string-based templates
	}

	// Validate delimiters
	if c.LeftDelim == "" {
		c.LeftDelim = "{{"
	}
	if c.RightDelim == "" {
		c.RightDelim = "}}"
	}

	// Validate missing key behavior
	validMissingKeyBehaviors := []string{"zero", "error", "invalid"}
	if c.MissingKeyBehavior != "" {
		valid := false
		for _, behavior := range validMissingKeyBehaviors {
			if c.MissingKeyBehavior == behavior {
				valid = true
				break
			}
		}
		if !valid {
			c.MissingKeyBehavior = "zero" // Default fallback
		}
	}

	return nil
}

// IsEmpty returns true if the config has no meaningful template configuration.
func (c TemplateConfig) IsEmpty() bool {
	return c.LayoutPath == "" && c.ContentPath == "" &&
		c.LayoutContent == "" && c.ContentTemplate == "" &&
		len(c.Partials) == 0
}

// HasLayout returns true if the config specifies a layout template.
func (c TemplateConfig) HasLayout() bool {
	return c.LayoutPath != "" || c.LayoutContent != ""
}

// HasContent returns true if the config specifies content template.
func (c TemplateConfig) HasContent() bool {
	return c.ContentPath != "" || c.ContentTemplate != ""
}

// HasErrorTemplates returns true if the config specifies error templates.
func (c TemplateConfig) HasErrorTemplates() bool {
	return c.ErrorLayoutPath != "" || c.ErrorContentPath != "" ||
		c.ErrorLayoutContent != "" || c.ErrorContentTemplate != ""
}

// Clone creates a deep copy of the template configuration.
func (c TemplateConfig) Clone() TemplateConfig {
	clone := c

	// Clone slices
	if c.Partials != nil {
		clone.Partials = make([]string, len(c.Partials))
		copy(clone.Partials, c.Partials)
	}

	if c.Extensions != nil {
		clone.Extensions = make([]string, len(c.Extensions))
		copy(clone.Extensions, c.Extensions)
	}

	// Clone function map
	if c.FuncMap != nil {
		clone.FuncMap = make(template.FuncMap)
		for k, v := range c.FuncMap {
			clone.FuncMap[k] = v
		}
	}

	return clone
}
