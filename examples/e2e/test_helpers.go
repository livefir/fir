package e2e

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

// SetupStaticFileServer adds a handler to serve the Alpine.js plugin file locally
// This solves the Docker networking issue where Docker Chrome can't access localhost:8000
func SetupStaticFileServer(mux *http.ServeMux) error {
	// Find the path to the built Alpine.js plugin
	pluginPath := filepath.Join("..", "..", "alpinejs-plugin", "dist", "cdn.js")

	// Check if the file exists
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		return fmt.Errorf("alpine.js plugin not found at %s. Please run 'npm run build' in the alpinejs-plugin directory", pluginPath)
	}

	// Serve the plugin at /cdn.js to match the template expectation when using relative URLs
	mux.HandleFunc("/cdn.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		http.ServeFile(w, r, pluginPath)
	})

	return nil
}
