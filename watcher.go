package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

var debounceDuration = 5000 * time.Millisecond

func debounce(execution func(), duration time.Duration) func() {
	var timer *time.Timer
	var mu sync.Mutex

	return func() {
		mu.Lock()
		defer mu.Unlock()

		if timer != nil {
			timer.Stop()
		}

		timer = time.AfterFunc(duration, execution)
	}
}

func main() {
	dirFlag := flag.String("d", ".", "Directory to watch for changes")
	cmdFlag := flag.String("c", "", "Shell command to run on file changes")
	flag.Parse()
	dirToWatch := *dirFlag
	commandToRun := *cmdFlag

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

	executeCommand := func() {
		if commandToRun != "" {
			var cmd *exec.Cmd
			if runtime.GOOS == "windows" {
				cmd = exec.Command("powershell.exe", "-Command", commandToRun)
			} else {
				cmd = exec.Command("sh", "-c", commandToRun)
			}
			cmd.Dir = dirToWatch
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			err := cmd.Run()
			if err != nil {
				log.Printf("Command execution failed: %s\n", err)
			} else {
				log.Printf("Command executed successfully\n")
			}
		}
	}

	debouncedExecuteCommand := debounce(executeCommand, debounceDuration)

	// Process file system events
	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				relativePath, _ := filepath.Rel(dirToWatch, event.Name)
				log.Printf("File changed: %s\n", relativePath)
				debouncedExecuteCommand()
			}
		case err := <-watcher.Errors:
			log.Printf("Watcher error: %s\n", err)
		}
	}
}
