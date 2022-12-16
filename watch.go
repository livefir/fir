package fir

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/adnaan/fir/internal/dom"
	"github.com/fsnotify/fsnotify"
	"golang.org/x/exp/slices"
)

// defaultWatchExtensions is an array of default extensions to watch for changes.
var defaultWatchExtensions = []string{".gohtml", ".gotmpl", ".html", ".tmpl"}

const devReloadChannel = "dev_reload"

func watchTemplates(wc *controller) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	done := make(chan bool)
	patcher := dom.NewPatcher()
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
					wc.pubsub.Publish(context.Background(), devReloadChannel, patcher.Reload().Patchset()...)
					time.Sleep(1000 * time.Millisecond)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	// watch extensions
	filepath.WalkDir(wc.publicDir, func(path string, d fs.DirEntry, err error) error {
		if d != nil && !d.IsDir() {
			if slices.Contains(wc.watchExts, filepath.Ext(path)) {
				if strings.Contains(path, "node_modules") {
					return nil
				}
				log.Println("watching =>", path)
				return watcher.Add(path)
			}
		}
		return nil
	})

	<-done
}
