package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

func main() {
	dirFlag := flag.String("d", ".", "Directory to watch for changes")
	flag.Parse()
	dirToWatch := *dirFlag

	// Check if the directory exists
	if _, err := os.Stat(dirToWatch); os.IsNotExist(err) {
		log.Fatalf("Directory %s does not exist\n", dirToWatch)
	}

	// Create a new watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Start watching the directory
	err = watcher.Add(dirToWatch)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Watching directory: %s\n", dirToWatch)

	// Process file system events
	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				relativePath, _ := filepath.Rel(dirToWatch, event.Name)
				log.Printf("File changed: %s\n", relativePath)
			}
		case err := <-watcher.Errors:
			log.Printf("Watcher error: %s\n", err)
		}
	}
}
