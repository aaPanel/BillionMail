package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

var (
	isDev     = false
	ROOT_PATH = GetCwdPath()
)

// FileWatcher structure
type FileWatcher struct {
	watcher     *fsnotify.Watcher // File system watcher
	rootPath    string            // Root directory to monitor
	scriptPath  string            // Script path to execute
	extensions  []string          // File extensions to monitor
	mu          sync.Mutex        // Mutex to prevent concurrent script execution
	isExecuting bool              // Flag indicating if script is executing
	debounceMap map[string]int64  // Debounce map, recording last modification time of each file
	debounceMs  int64             // Debounce time interval (milliseconds)
	logger      *log.Logger       // Logger
	delay       time.Duration     // Delay before executing the script
	timer       *time.Timer       // Timer for debouncing
}

// NewFileWatcher creates a new file watcher
func NewFileWatcher(rootPath string, scriptPath string, extensions []string) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	return &FileWatcher{
		watcher:     watcher,
		rootPath:    rootPath,
		scriptPath:  scriptPath,
		extensions:  extensions,
		debounceMap: make(map[string]int64),
		debounceMs:  500, // Default debounce time: 500 milliseconds
		logger:      log.New(os.Stdout, "[FileWatcher] ", log.LstdFlags),
		delay:       time.Second,
		timer:       time.NewTimer(time.Second),
	}, nil
}

// Start begins monitoring file changes
func (fw *FileWatcher) Start(ctx context.Context) error {
	// Add all directories to the watcher
	if err := fw.addDirsRecursively(fw.rootPath); err != nil {
		return err
	}

	fw.logger.Printf("Started monitoring directory: %s", fw.rootPath)
	fw.logger.Printf("Script to execute: %s", fw.scriptPath)
	fw.logger.Printf("Monitoring file types: %v", fw.extensions)

	// Start the monitoring loop
	go fw.watchLoop(ctx)

	return nil
}

// Stop stops monitoring
func (fw *FileWatcher) Stop() {
	fw.watcher.Close()
}

// addDirsRecursively recursively adds directories to the watcher
func (fw *FileWatcher) addDirsRecursively(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and files
		if strings.HasPrefix(filepath.Base(path), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip vendor and node_modules directories
		if info.IsDir() && (filepath.Base(path) == "vendor" || filepath.Base(path) == "node_modules") {
			return filepath.SkipDir
		}

		// Only watch directory changes
		if info.IsDir() {
			if err := fw.watcher.Add(path); err != nil {
				return fmt.Errorf("failed to add directory %s to watcher: %w", path, err)
			}
		}
		return nil
	})
}

// watchLoop monitoring loop
func (fw *FileWatcher) watchLoop(ctx context.Context) {
	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}

			// Only focus on write and create events
			if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
				continue
			}

			// If a new directory is created, add it to the watcher
			if fi, err := os.Stat(event.Name); err == nil && fi.IsDir() {
				fw.watcher.Add(event.Name)
				continue
			}

			// Check if file extension matches
			if !fw.hasMatchingExtension(event.Name) {
				continue
			}

			// Apply debounce logic
			if fw.shouldDebounce(event.Name) {
				continue
			}

			fw.timer.Reset(fw.delay)

			// Execute script
			fw.logger.Printf("Detected file change: %s", event.Name)
			go fw.executeScript(ctx)

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			fw.logger.Printf("Watcher error: %v", err)
		}
	}
}

// hasMatchingExtension checks if file has a matching extension
func (fw *FileWatcher) hasMatchingExtension(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	for _, e := range fw.extensions {
		if ext == e {
			return true
		}
	}
	return false
}

// shouldDebounce checks if debouncing should be applied (avoid multiple executions in short time)
func (fw *FileWatcher) shouldDebounce(filename string) bool {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	now := time.Now().UnixMilli()
	lastTime, exists := fw.debounceMap[filename]

	// Update last modification time
	fw.debounceMap[filename] = now

	// If it's the first modification or beyond debounce time, don't debounce
	if !exists || now-lastTime > fw.debounceMs {
		return false
	}

	return true
}

// executeScript executes the script
func (fw *FileWatcher) executeScript(ctx context.Context) {
	<-fw.timer.C

	fw.mu.Lock()
	if fw.isExecuting {
		fw.mu.Unlock()
		return
	}
	fw.isExecuting = true
	fw.mu.Unlock()

	defer func() {
		fw.mu.Lock()
		fw.isExecuting = false
		fw.mu.Unlock()
	}()

	// Check if script exists
	if _, err := os.Stat(fw.scriptPath); os.IsNotExist(err) {
		fw.logger.Printf("Script does not exist: %s", fw.scriptPath)
		return
	}

	// Create command
	cmd := exec.Command("/bin/bash", fw.scriptPath)
	cmd.Env = os.Environ()

	// Get standard output and error output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fw.logger.Printf("Failed to get standard output: %v", err)
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		fw.logger.Printf("Failed to get error output: %v", err)
		return
	}

	// Start command
	fw.logger.Printf("Executing script: %s", fw.scriptPath)
	if err := cmd.Start(); err != nil {
		fw.logger.Printf("Failed to start script: %v", err)
		return
	}

	// Real-time reading and output of standard output
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			fmt.Println("►", scanner.Text())
		}
	}()

	// Real-time reading and output of error output
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Println("⚠", scanner.Text())
		}
	}()

	// Wait for command to complete
	if err := cmd.Wait(); err != nil {
		fw.logger.Printf("Script execution failed: %v", err)
		return
	}

	fw.logger.Printf("Script execution completed")
}

// GetCwdPath Get current working directory
func GetCwdPath() string {
	if len(os.Args) == 0 {
		p, _ := os.Getwd()
		return p
	}

	// Get executable path
	exePath := filepath.Dir(os.Args[0])

	// Check if it is a development environment
	if isDev || strings.Contains(filepath.ToSlash(exePath), "/go-build") {
		isDev = true
		p, _ := os.Getwd()
		return p
	}

	// Get absolute path
	p, _ := filepath.Abs(exePath)
	return p
}

// AbsPath Get absolute path
func AbsPath(p string) string {
	// In Linux, if the path starts with "/", it is an absolute path
	if strings.HasPrefix(p, "/") {
		return p
	}

	// In Windows, if the path starts with "C:", it is an absolute path
	if len(p) > 1 && p[1] == ':' {
		return p
	}

	p, _ = filepath.Abs(filepath.Join(ROOT_PATH, p))

	return p
}

func main() {
	ctx := context.Background()
	logger := log.New(os.Stdout, "[Main] ", log.LstdFlags)

	// Arguments for root path and script path
	var rootPath, scriptPath string

	// Get command line arguments
	args := os.Args
	if len(args) < 3 {
		rootPath = "."              // Default root path
		scriptPath = "./run_dev.sh" // Default script path
	} else {
		rootPath = args[1]
		scriptPath = args[2]
	}

	// Create file watcher
	watcher, err := NewFileWatcher(AbsPath(rootPath), AbsPath(scriptPath), []string{".go", ".html"})
	if err != nil {
		logger.Fatalf("Failed to create file watcher: %v", err)
	}
	defer watcher.Stop()

	// Start monitoring
	if err := watcher.Start(ctx); err != nil {
		logger.Fatalf("Failed to start monitoring: %v", err)
	}

	logger.Printf("File watcher started, press Ctrl+C to stop")

	// Keep the main goroutine running
	select {}
}
