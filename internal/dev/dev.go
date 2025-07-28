package dev

import (
	"net/http"
	"os"
	"path/filepath"
)

// SetupAlpinePluginServer sets up a static file server for the Alpine.js fir plugin
// during development. This allows templates to use /cdn.js locally instead of unpkg CDN.
//
// The function automatically detects the correct path to the Alpine.js plugin file
// by checking common locations relative to the current working directory.
func SetupAlpinePluginServer() {
	// Find the Alpine.js plugin file
	pluginPath := filepath.Join("alpinejs-plugin", "dist", "cdn.js")

	// Check if the plugin file exists
	if _, err := os.Stat(pluginPath); err != nil {
		// If not found in current directory, try going up one level (for examples)
		pluginPath = filepath.Join("..", pluginPath)
		if _, err := os.Stat(pluginPath); err != nil {
			// Try going up two levels (for deeply nested examples)
			pluginPath = filepath.Join("..", "..", "alpinejs-plugin", "dist", "cdn.js")
			if _, err := os.Stat(pluginPath); err != nil {
				// Plugin file not found, skip setup silently
				return
			}
		}
	}

	http.HandleFunc("/cdn.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		http.ServeFile(w, r, pluginPath)
	})
}
