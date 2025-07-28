package dev

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/livefir/fir/internal/logger"
	"github.com/livefir/fir/pubsub"
)

// DefaultWatchExtensions is an array of default extensions to watch for changes.
var DefaultWatchExtensions = []string{".gohtml", ".gotmpl", ".html", ".tmpl"}

const DevReloadChannel = "dev_reload"

// stringPtr returns a pointer to the given string
func stringPtr(s string) *string {
	return &s
}

// Controller interface for type compatibility
type Controller interface {
	GetPubsub() pubsub.Adapter
	GetPublicDir() string
	GetWatchExts() []string
}

func WatchTemplates(wc Controller) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	done := make(chan bool)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write ||
					event.Op&fsnotify.Remove == fsnotify.Remove ||
					event.Op&fsnotify.Create == fsnotify.Create {
					fmt.Printf("[watcher]==> file changed: %v, reloading ... \n", event.Name)
					wc.GetPubsub().Publish(context.Background(), DevReloadChannel, pubsub.Event{ID: stringPtr("reload")})
					time.Sleep(1000 * time.Millisecond)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logger.Errorf("error: %v", err)
			}
		}
	}()

	// watch extensions
	filepath.WalkDir(wc.GetPublicDir(), func(path string, d fs.DirEntry, err error) error {
		if d != nil && !d.IsDir() {
			if slices.Contains(wc.GetWatchExts(), filepath.Ext(path)) {
				if strings.Contains(path, "node_modules") {
					return nil
				}
				fmt.Println("watching =>", path)
				return watcher.Add(path)
			}
		}
		return nil
	})

	<-done
}
