package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// FileState represents the state of a file at a point in time
type FileState struct {
	Path    string
	Size    int64
	ModTime time.Time
	IsDir   bool
}

// FileEvent represents a file system event
type FileEvent struct {
	Type      string // "created", "modified", "deleted"
	Path      string
	Timestamp time.Time
}

// Config holds the watcher configuration
type Config struct {
	Directory string
	Recursive bool
	Interval  time.Duration
	Pattern   string
	Exclude   []string
	Command   string
	Verbose   bool
}

// Watcher monitors a directory for changes
type Watcher struct {
	config       Config
	currentState map[string]FileState
	events       []FileEvent
	stopChan     chan struct{}
}

func main() {
	// Define command-line flags
	var (
		directory = flag.String("d", ".", "Directory to watch")
		recursive = flag.Bool("r", false, "Watch subdirectories recursively")
		interval  = flag.Int("i", 2, "Polling interval in seconds")
		pattern   = flag.String("p", "*", "File pattern to watch (e.g., *.go)")
		exclude   = flag.String("e", "", "Comma-separated patterns to exclude")
		command   = flag.String("c", "", "Command to run when changes detected")
		verbose   = flag.Bool("v", false, "Show detailed output")
		help      = flag.Bool("h", false, "Show help")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Monitor directories for file changes and trigger actions.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -d .\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -d ./src -p \"*.go\" -c \"go test ./...\"\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -d . -r -e \".git,node_modules\" -v\n", os.Args[0])
	}
	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	// Parse exclude patterns
	var excludePatterns []string
	if *exclude != "" {
		excludePatterns = strings.Split(*exclude, ",")
		for i := range excludePatterns {
			excludePatterns[i] = strings.TrimSpace(excludePatterns[i])
		}
	}

	// Create configuration
	config := Config{
		Directory: *directory,
		Recursive: *recursive,
		Interval:  time.Duration(*interval) * time.Second,
		Pattern:   *pattern,
		Exclude:   excludePatterns,
		Command:   *command,
		Verbose:   *verbose,
	}

	// Validate directory
	if _, err := os.Stat(config.Directory); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Directory does not exist: %s\n", config.Directory)
		os.Exit(1)
	}

	// Create and start watcher
	watcher := NewWatcher(config)

	fmt.Printf("Watching: %s\n", config.Directory)
	fmt.Printf("Pattern: %s\n", config.Pattern)
	if config.Recursive {
		fmt.Printf("Mode: Recursive\n")
	}
	if len(config.Exclude) > 0 {
		fmt.Printf("Excluding: %s\n", strings.Join(config.Exclude, ", "))
	}
	if config.Command != "" {
		fmt.Printf("Command: %s\n", config.Command)
	}
	fmt.Printf("Interval: %v\n", config.Interval)
	fmt.Println("\nPress Ctrl+C to stop watching...")
	fmt.Println(strings.Repeat("-", 50))

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start watching
	go watcher.Start()

	// Wait for interrupt
	<-sigChan
	fmt.Println("\n\nShutting down gracefully...")
	watcher.Stop()

	// Show statistics
	fmt.Printf("\nTotal events detected: %d\n", len(watcher.events))
}

// NewWatcher creates a new file watcher
func NewWatcher(config Config) *Watcher {
	return &Watcher{
		config:       config,
		currentState: make(map[string]FileState),
		events:       make([]FileEvent, 0),
		stopChan:     make(chan struct{}),
	}
}

// Start begins monitoring the directory
func (w *Watcher) Start() {
	// Initial scan
	w.currentState = w.scanDirectory()

	if w.config.Verbose {
		fmt.Printf("Initial scan: %d files tracked\n", len(w.currentState))
	}

	// Create ticker for periodic checks
	ticker := time.NewTicker(w.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.checkForChanges()
		case <-w.stopChan:
			return
		}
	}
}

// Stop stops the watcher
func (w *Watcher) Stop() {
	close(w.stopChan)
}

// scanDirectory scans the directory and returns current state
func (w *Watcher) scanDirectory() map[string]FileState {
	state := make(map[string]FileState)

	var scanFunc filepath.WalkFunc = func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if w.config.Verbose {
				fmt.Printf("Warning: Cannot access %s: %v\n", path, err)
			}
			return nil // Continue walking
		}

		// Skip excluded patterns
		if w.shouldExclude(path) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// If not recursive, skip subdirectories
		if !w.config.Recursive && path != w.config.Directory && info.IsDir() {
			return filepath.SkipDir
		}

		// Skip directories (we only track files)
		if info.IsDir() {
			return nil
		}

		// Check pattern match
		if !w.matchesPattern(path) {
			return nil
		}

		state[path] = FileState{
			Path:    path,
			Size:    info.Size(),
			ModTime: info.ModTime(),
			IsDir:   info.IsDir(),
		}

		return nil
	}

	filepath.Walk(w.config.Directory, scanFunc)
	return state
}

// checkForChanges compares current state with previous state
func (w *Watcher) checkForChanges() {
	newState := w.scanDirectory()
	events := w.compareStates(w.currentState, newState)

	if len(events) > 0 {
		for _, event := range events {
			w.handleEvent(event)
			w.events = append(w.events, event)
		}

		// Execute command if configured
		if w.config.Command != "" {
			w.executeCommand(events)
		}
	}

	w.currentState = newState
}

// compareStates compares two states and returns events
func (w *Watcher) compareStates(old, new map[string]FileState) []FileEvent {
	var events []FileEvent

	// Check for new or modified files
	for path, newInfo := range new {
		if oldInfo, exists := old[path]; !exists {
			// New file
			events = append(events, FileEvent{
				Type:      "created",
				Path:      path,
				Timestamp: time.Now(),
			})
		} else if w.hasChanged(oldInfo, newInfo) {
			// Modified file
			events = append(events, FileEvent{
				Type:      "modified",
				Path:      path,
				Timestamp: time.Now(),
			})
		}
	}

	// Check for deleted files
	for path := range old {
		if _, exists := new[path]; !exists {
			events = append(events, FileEvent{
				Type:      "deleted",
				Path:      path,
				Timestamp: time.Now(),
			})
		}
	}

	return events
}

// hasChanged checks if a file has changed
func (w *Watcher) hasChanged(old, new FileState) bool {
	// Check modification time and size
	return new.ModTime.After(old.ModTime) || new.Size != old.Size
}

// handleEvent handles a file system event
func (w *Watcher) handleEvent(event FileEvent) {
	timestamp := event.Timestamp.Format("15:04:05")
	icon := w.getEventIcon(event.Type)

	fmt.Printf("[%s] %s %s %s\n", timestamp, icon, event.Type, event.Path)

	if w.config.Verbose {
		fmt.Printf("         Event type: %s\n", event.Type)
		if event.Type != "deleted" {
			if info, err := os.Stat(event.Path); err == nil {
				fmt.Printf("         Size: %d bytes\n", info.Size())
				fmt.Printf("         Modified: %s\n", info.ModTime().Format(time.RFC3339))
			}
		}
	}
}

// getEventIcon returns an icon for the event type
func (w *Watcher) getEventIcon(eventType string) string {
	switch eventType {
	case "created":
		return "[+]"
	case "modified":
		return "[~]"
	case "deleted":
		return "[-]"
	default:
		return "[?]"
	}
}

// executeCommand executes the configured command
func (w *Watcher) executeCommand(events []FileEvent) {
	if w.config.Command == "" {
		return
	}

	// Get list of changed files
	var files []string
	for _, event := range events {
		files = append(files, event.Path)
	}

	cmd := w.config.Command
	cmd = strings.ReplaceAll(cmd, "{files}", strings.Join(files, " "))

	fmt.Printf("\nExecuting: %s\n", cmd)

	// Parse and execute command
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return
	}

	execCmd := exec.Command(parts[0], parts[1:]...)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	if err := execCmd.Run(); err != nil {
		fmt.Printf("Command failed: %v\n", err)
	}

	fmt.Println(strings.Repeat("-", 50))
}

// matchesPattern checks if a file matches the pattern
func (w *Watcher) matchesPattern(path string) bool {
	matched, err := filepath.Match(w.config.Pattern, filepath.Base(path))
	if err != nil {
		return false
	}
	return matched
}

// shouldExclude checks if a path should be excluded
func (w *Watcher) shouldExclude(path string) bool {
	for _, pattern := range w.config.Exclude {
		if strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}
