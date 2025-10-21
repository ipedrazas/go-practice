# Exercise 10: File Watcher

## 🎯 Objective
Create a file system watcher that monitors directories for changes and triggers actions when files are created, modified, or deleted.

## 📋 Main Focus Areas
- **File system monitoring** (polling and state tracking)
- **Time-based operations** (`time` package, tickers)
- **File state comparison** (modification times, checksums)
- **Event-driven programming** (channels, signal handling)
- **Concurrent operations** (goroutines for monitoring)
- **Pattern matching** (file filters and exclusions)

## 🔧 What You'll Build
A file watcher that:
- Monitors directories for file changes in real-time
- Detects file creation, modification, and deletion
- Supports recursive directory watching
- Filters files by pattern (glob matching)
- Executes commands when changes are detected
- Provides detailed change logs
- Handles graceful shutdown

## 📝 Instructions

### Step 1: Basic File State Tracking
Create a tool that:
1. Takes a directory path as input
2. Records initial state of all files (name, size, mod time)
3. Periodically checks for changes
4. Reports any detected changes

### Step 2: Command-line Interface
Add support for these flags:
- `-d, --directory`: Directory to watch (default: current directory)
- `-r, --recursive`: Watch subdirectories recursively
- `-i, --interval`: Polling interval in seconds (default: 2)
- `-p, --pattern`: File pattern to watch (e.g., "*.go", "*.txt")
- `-e, --exclude`: Patterns to exclude (e.g., ".git", "node_modules")
- `-c, --command`: Command to run when changes detected
- `-v, --verbose`: Show detailed output

### Step 3: Event Detection
Implement detection for:
- **Created**: New files appear in watched directory
- **Modified**: Existing files change (size or mod time)
- **Deleted**: Files are removed from watched directory
- **Renamed**: Files are moved (detected as delete + create)

### Step 4: Advanced Features
Add these capabilities:
- Debouncing (ignore rapid successive changes)
- Command execution with file path variables
- Event history/log
- JSON output format for events
- Signal handling (Ctrl+C for clean shutdown)
- Multiple directory watching
- Checksum-based change detection (optional)

## 🚀 Getting Started

```bash
# Navigate to exercise directory
cd 10-file-watcher

# Create your main.go file
touch main.go

# Watch current directory
go run main.go

# Watch specific directory
go run main.go -d /path/to/directory

# Watch with pattern filter
go run main.go -d . -p "*.go"

# Execute command on changes
go run main.go -d ./src -c "go test ./..."

# Recursive watch with exclusions
go run main.go -d . -r -e ".git" -e "node_modules"
```

## 💡 Implementation Tips

### File State Structure
```go
type FileState struct {
    Path    string
    Size    int64
    ModTime time.Time
    IsDir   bool
}

type FileEvent struct {
    Type      string // "created", "modified", "deleted"
    Path      string
    Timestamp time.Time
}
```

### Basic Polling Loop
```go
func watchDirectory(path string, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    currentState := scanDirectory(path)

    for range ticker.C {
        newState := scanDirectory(path)
        events := compareStates(currentState, newState)

        for _, event := range events {
            handleEvent(event)
        }

        currentState = newState
    }
}
```

### State Comparison
```go
func compareStates(old, new map[string]FileState) []FileEvent {
    var events []FileEvent

    // Check for new or modified files
    for path, newInfo := range new {
        if oldInfo, exists := old[path]; !exists {
            events = append(events, FileEvent{
                Type: "created",
                Path: path,
                Timestamp: time.Now(),
            })
        } else if newInfo.ModTime.After(oldInfo.ModTime) ||
                  newInfo.Size != oldInfo.Size {
            events = append(events, FileEvent{
                Type: "modified",
                Path: path,
                Timestamp: time.Now(),
            })
        }
    }

    // Check for deleted files
    for path := range old {
        if _, exists := new[path]; !exists {
            events = append(events, FileEvent{
                Type: "deleted",
                Path: path,
                Timestamp: time.Now(),
            })
        }
    }

    return events
}
```

### Pattern Matching
```go
func matchesPattern(filename, pattern string) bool {
    matched, err := filepath.Match(pattern, filepath.Base(filename))
    if err != nil {
        return false
    }
    return matched
}
```

### Signal Handling
```go
func setupSignalHandler() chan os.Signal {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    return sigChan
}

func main() {
    sigChan := setupSignalHandler()

    // Start watching in goroutine
    go watchDirectory(".", 2*time.Second)

    // Wait for interrupt signal
    <-sigChan
    fmt.Println("\nShutting down gracefully...")
}
```

### Command Execution
```go
func executeCommand(command string, filePath string) error {
    // Replace placeholders
    cmd := strings.ReplaceAll(command, "{file}", filePath)

    // Execute command
    parts := strings.Fields(cmd)
    exec := exec.Command(parts[0], parts[1:]...)
    exec.Stdout = os.Stdout
    exec.Stderr = os.Stderr

    return exec.Run()
}
```

## 🧪 Test Cases
Test different scenarios:
- Create files in watched directory
- Modify existing files (edit content)
- Delete files
- Rename files
- Create subdirectories (with -r flag)
- Mass changes (multiple files at once)
- Pattern matching (only watch specific file types)
- Rapid successive changes (test debouncing)

## 📚 Go Concepts Covered
- `time.Ticker` for periodic execution
- `os.Stat` and file metadata
- `filepath.Walk` for recursive scanning
- `filepath.Match` for glob patterns
- Goroutines for concurrent monitoring
- Channels for event communication
- `os/signal` for handling interrupts
- `os/exec` for running external commands
- Map-based state tracking
- Graceful shutdown patterns

## 🔗 Useful Resources
- [Go time package](https://pkg.go.dev/time)
- [Go os package](https://pkg.go.dev/os)
- [Go filepath package](https://pkg.go.dev/path/filepath)
- [Go signal handling](https://pkg.go.dev/os/signal)
- [Go exec package](https://pkg.go.dev/os/exec)

## ✅ Success Criteria
Your tool should be able to:
- [ ] Monitor a directory for file changes
- [ ] Detect created, modified, and deleted files
- [ ] Support recursive directory watching
- [ ] Filter files by pattern (glob matching)
- [ ] Execute commands when changes are detected
- [ ] Handle graceful shutdown on Ctrl+C
- [ ] Provide clear output about detected changes
- [ ] Handle file system errors gracefully

## 🎁 Bonus Features
If you want an extra challenge:
- Checksum-based change detection (detect content changes even if mod time unchanged)
- Debouncing (delay action until changes stop)
- Watch multiple directories simultaneously
- Configuration file support (JSON/YAML)
- Rate limiting (prevent command spam)
- Event history with timestamps
- Different output formats (JSON, structured logs)
- Integration with external tools (rsync, git, build systems)
- Web dashboard showing real-time changes
- File content diffing (show what changed in file)

## 💡 Real-World Use Cases

This type of tool is commonly used for:
- **Auto-reloading development servers** (restart server when code changes)
- **Build automation** (recompile on source changes)
- **Backup triggers** (backup files when modified)
- **Log monitoring** (process new log entries)
- **Sync tools** (keep directories in sync)
- **Test runners** (run tests on code changes)

## 🔍 Implementation Approaches

### Approach 1: Polling (What We'll Build)
**Pros:**
- Simple to implement
- Works on all platforms
- No external dependencies
- Full control over checking logic

**Cons:**
- Uses more CPU (constant checking)
- Delay between change and detection
- Not instant notification

### Approach 2: OS-Native Events (Bonus)
For production use, you might use platform-specific APIs:
- **Linux**: `inotify`
- **macOS**: `FSEvents`
- **Windows**: `ReadDirectoryChangesW`

Go library: `github.com/fsnotify/fsnotify` (wraps native APIs)

For this exercise, we'll implement polling to avoid external dependencies and focus on learning core Go concepts.

## 📝 Example Usage Scenarios

```bash
# Watch Go files and run tests on changes
go run main.go -d . -p "*.go" -c "go test ./..."

# Watch docs and regenerate HTML
go run main.go -d ./docs -p "*.md" -c "make docs"

# Watch config files with verbose output
go run main.go -d /etc/myapp -v

# Recursive watch excluding build artifacts
go run main.go -d ./project -r -e "build" -e "dist" -e ".git"

# Monitor logs directory
go run main.go -d /var/log/myapp -p "*.log" -v
```

When you're ready, check out the [solution](./solution/main.go) to see a complete implementation.
