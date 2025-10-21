# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This is a Go learning repository containing 9 hands-on CLI exercises designed to teach practical Go programming through building real-world tools. Each exercise is self-contained in its own numbered directory with a README and solution.

## Project Structure

```
go-practice/
├── 01-url-downloader/          # HTTP requests, file I/O, progress tracking
├── 02-file-organizer/          # File system operations, directory traversal
├── 03-log-analyzer/            # Text parsing, regex, data aggregation
├── 04-json-validator/          # JSON parsing, validation, struct tags
├── 05-port-scanner/            # Network programming, concurrency, goroutines
├── 06-dir-sizer/               # Recursive traversal, disk usage analysis
├── 07-web-server/              # HTTP server, templates, routing
├── 08-index-generator/         # Markdown generation, file parsing
├── 09-testing-fundamentals/    # Go testing, TDD, benchmarks, table-driven tests
│   └── solution/
│       ├── main.go             # CLI interface
│       ├── go.mod              # Module definition
│       └── password/           # Package with validators and tests
└── 10-file-watcher/            # File system monitoring, event-driven programming
```

### Exercise Structure Pattern

Each exercise follows a consistent structure:
- `README.md` - Exercise instructions and learning objectives
- `solution/main.go` - Complete reference implementation
- Most exercises are standalone single-file programs
- Exercise 09 has a full Go module with packages and comprehensive tests

### Special Notes on Exercise 09

Exercise 09 (`09-testing-fundamentals`) is unique:
- Only exercise with a proper `go.mod` file and package structure
- Contains intentionally failing tests for TDD practice
- Has a `password` package with validators and comprehensive test suite
- Mix of passing and failing tests to demonstrate real-world debugging
- CLI interface in `main.go` for interactive testing

## Build and Development Commands

### Using Task (Recommended)

The repository uses [Task](https://taskfile.dev/) for automation. All commands are defined in `Taskfile.yaml`.

**Common commands:**
```bash
# Build all exercises
task build

# Run tests (currently only exercise 09)
task test

# Quick validation (build check)
task validate

# Full validation (build + test)
task check

# Format all Go code
task fmt

# Run static analysis
task vet

# Clean build artifacts
task clean

# Build specific exercise
task build-exercise -- 01-url-downloader

# Test specific exercise
task test-exercise -- 09-testing-fundamentals

# Show all available tasks
task --list
```

### Manual Build Commands

**For exercises 01-08 and 10 (single-file programs):**
```bash
cd <exercise-name>/solution
go build -o <exercise-name> main.go
go run main.go [args]
```

**For exercise 09 (module with packages):**
```bash
cd 09-testing-fundamentals/solution

# Build the CLI
go build -o 09-testing-fundamentals main.go

# Run tests
go test ./password -v

# Run specific test
go test ./password -run TestIsValidLength -v

# Run benchmarks
go test ./password -bench=.

# Check test coverage
go test ./password -cover

# Interactive CLI
echo -e "test123\nexit" | go run main.go
```

## Testing Strategy

### Exercise 09 Testing Commands

Exercise 09 is specifically designed to teach Go testing:

```bash
cd 09-testing-fundamentals/solution

# Run all tests (will show some failures - this is intentional)
go test ./password -v

# Run only passing tests (good starting point)
go test ./password -run TestIsValidLength -v
go test ./password -run TestLog2 -v
go test ./password -run TestIsRepeated -v

# Run benchmarks
go test ./password -bench=BenchmarkPasswordValidation

# Run specific test with subtests
go test ./password -run TestPasswordValidator_Validate/strong_password -v

# Generate coverage report
go test ./password -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Test File Organization

- Test files follow `*_test.go` naming convention
- Tests use table-driven patterns extensively
- Includes benchmark tests for performance measurement
- Mock implementations for external dependencies (e.g., `MockBreachedService`)
- Example tests that serve as documentation

## Go Version and Dependencies

- **Go Version**: 1.19+ required (specified in `Taskfile.yaml`)
- **Module**: Only exercise 09 uses Go modules (`go.mod`)
- **Dependencies**: All exercises use only the Go standard library
- **No External Dependencies**: Intentionally avoids third-party packages

## Code Architecture Patterns

### Common Patterns Across Exercises

1. **Command-line argument parsing** using `flag` package
2. **Error handling** with descriptive error wrapping
3. **Progress indicators** for long-running operations
4. **Dry-run modes** for safe testing
5. **Graceful error handling** with meaningful messages

### Exercise-Specific Patterns

**01-url-downloader:**
- Custom `io.Writer` implementation for progress tracking
- Multi-writer pattern for simultaneous file write and progress
- HTTP client configuration with timeouts

**05-port-scanner:**
- Worker pool pattern for concurrent port scanning
- Channel-based job distribution
- Goroutines with configurable concurrency limits

**07-web-server:**
- HTTP handler functions
- HTML template rendering
- Static file serving with `http.FileServer`

**09-testing-fundamentals:**
- Table-driven tests for comprehensive coverage
- Subtests for logical grouping
- Benchmark functions for performance testing
- Mock implementations implementing interfaces
- Test helpers and utility functions

**10-file-watcher:**
- Polling-based file system monitoring
- State tracking with maps
- Event detection and handling
- Signal handling for graceful shutdown
- Concurrent monitoring with goroutines

## Common Development Workflows

### Adding a New Exercise

1. Create directory with format `##-exercise-name/`
2. Add `README.md` with exercise objectives and instructions
3. Create `solution/main.go` with reference implementation
4. Update main `README.md` exercise table
5. Add exercise to `Taskfile.yaml` build lists

### Working with Exercise 09

The testing exercise has a unique workflow:
1. **Understand passing tests first** - Learn from working examples
2. **Analyze failures** - Understand what the test expects
3. **Implement fixes** - Follow TDD cycle (Red-Green-Refactor)
4. **Run specific tests** - Iterate on individual test cases
5. **Check coverage** - Ensure comprehensive testing

### Code Quality Checks

Before committing changes:
```bash
task fmt      # Format code
task vet      # Static analysis
task build    # Ensure everything compiles
task test     # Run tests
```

## Key Learning Concepts by Exercise

1. **HTTP client, file I/O, progress tracking**
2. **File system operations, directory traversal**
3. **Text parsing, regex, data aggregation**
4. **JSON marshaling/unmarshaling, validation**
5. **Network programming, goroutines, channels, concurrency**
6. **Recursive algorithms, disk usage calculation**
7. **HTTP server, templating, routing**
8. **File parsing, Markdown generation, metadata extraction**
9. **Testing framework, TDD, mocks, benchmarks, table-driven tests**
10. **File system monitoring, time-based operations, event-driven programming, signal handling**

## Important Notes

- **Standard Library Focus**: All exercises intentionally use only Go's standard library to teach core concepts
- **Progressive Complexity**: Exercises are numbered by difficulty and build on previous concepts
- **Practical Tools**: Each exercise creates a genuinely useful CLI tool, not toy examples
- **TDD Exercise**: Exercise 09 intentionally has failing tests for learning purposes - this is expected
- **No Module Files**: Exercises 01-08 and 10 don't use `go.mod` files; they're simple standalone programs

## Troubleshooting

### Exercise 09 Test Failures

Many tests in exercise 09 are intentionally incomplete or failing. This is by design:
- 4 core tests pass (length validation, pattern detection, math functions)
- Several tests are left as learning opportunities for TDD practice
- Focus on understanding test structure before fixing implementation

### Build Issues

If builds fail:
1. Check Go version: `go version` (needs 1.19+)
2. Verify you're in the correct directory (should be in `solution/`)
3. For exercise 09, ensure you're using `go.mod` properly: `go mod tidy`
4. Check syntax: `go vet main.go`

### Task Not Found

If `task` command is not available:
```bash
# Install Task
brew install go-task        # macOS
# OR
curl -sL https://taskfile.dev/install.sh | sh  # Linux
```

## Repository Maintenance

This is a teaching repository, so:
- Solutions are complete and fully implemented for reference
- Code quality and clarity prioritized over clever optimizations
- Comments explain "why" not just "what"
- Each exercise is independent and can be completed in any order (though sequential is recommended)
