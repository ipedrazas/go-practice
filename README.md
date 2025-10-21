# Go Practice: Real-World CLI Exercises

Welcome to Go Practice! This collection of hands-on exercises focuses on building practical CLI tools using Go's standard library. Each exercise is designed to teach you real-world skills while creating useful tools you might actually use.

## 🎯 Learning Philosophy

Instead of abstract examples and toy functions, these exercises build complete, useful CLI applications. You'll learn Go by solving real problems and creating tools that have genuine utility.

## 📚 Exercise Syllabus

| Exercise | Focus | Description |
|----------|-------|-------------|
| [1. URL Downloader CLI Tool](./01-url-downloader/) | Command-line arguments, HTTP requests, File operations, Progress tracking, Error handling | Create a command-line tool that downloads files from URLs with progress indicators and error handling. |
| [2. File Organizer Utility](./02-file-organizer/) | File system operations, Directory traversal, File metadata, Pattern matching, String operations | Create a CLI tool that organizes files in a directory based on their type, size, or custom rules. |
| [3. Log Analyzer Tool](./03-log-analyzer/) | File reading and processing, Regular expressions, Text parsing and manipulation, Date/time handling, Data aggregation and statistics | Create a CLI tool that parses and analyzes log files to extract useful information, patterns, and statistics. |
| [4. JSON Configuration Validator](./04-json-validator/) | JSON parsing and manipulation, Struct definitions and tags, Validation and error handling, File I/O operations, Schema validation patterns | Create a CLI tool that validates JSON configuration files against predefined schemas and rules. |
| [5. Port Scanner Utility](./05-port-scanner/) | Network programming, Concurrency, Timeout handling, Parallel processing, Command-line argument parsing | Create a network port scanner that checks for open ports on target hosts using TCP connections. |
| [6. Directory Size Analyzer](./06-dir-sizer/) | File system navigation, Recursive directory traversal, Human-readable formatting, Data sorting and aggregation, Memory-efficient processing | Create a tool that analyzes disk usage by directory, showing which directories consume the most space. |
| [7. Simple Web Server with Templates](./07-web-server/) | HTTP server programming, HTML templates, Static file serving, HTTP routing and handlers, Request handling and form processing | Create a web server that serves dynamic content using HTML templates and static files. |
| [8. Exercise Index Generator Tool](./08-index-generator/) | Directory traversal and analysis, Markdown generation, File parsing and content extraction, Template-based content generation, Metadata extraction from structured content | Create a tool that automatically generates the main README.md index by scanning all exercise directories and extracting their metadata. |
| [9. Testing Fundamentals and TDD](./09-testing-fundamentals/) | Go testing framework, Test-Driven Development (TDD) methodology, Unit testing and test organization, Mocking and test doubles, Benchmark testing, Table-driven tests, Test coverage and reporting | Learn Go's testing framework by building a simple library with comprehensive tests using Test-Driven Development (TDD). |
| [10. File Watcher](./10-file-watcher/) | File system monitoring, Time-based operations, Event-driven programming, Concurrent operations, Signal handling, Pattern matching | Create a file system watcher that monitors directories for changes and triggers actions when files are created, modified, or deleted. |
## 🚀 Getting Started

Each exercise follows this structure:
- **Folder**: Contains the complete exercise in its own directory
- **Instructions**: Step-by-step guide to build the tool
- **Focus Areas**: Key Go concepts and standard library packages you'll learn
- **Solution**: Complete implementation for reference

## 📖 How to Use These Exercises

1. **Follow the Order**: Exercises build on previous concepts
2. **Read the Instructions First**: Understand what you're building
3. **Try It Yourself**: Write the code before looking at solutions
4. **Experiment**: Modify the tools to add your own features
5. **Build on Them**: Combine tools or extend them for new use cases

## 🛠 Prerequisites

- Go 1.19 or later installed
- Basic understanding of programming concepts
- Text editor or IDE
- Terminal/command line
- (Optional) [Task](https://taskfile.dev/) for automated builds and validation

### 🚀 Quick Start with Task

If you have Task installed, you can quickly validate all exercises:

```bash
# Install Task (if needed)
brew install go-task  # macOS
curl -sL https://taskfile.dev/install.sh | sh  # Linux

# Validate all exercises build correctly
task validate

# Run tests (currently testing exercise only)
task test

# Get help with all commands
task --list
```


## 📝 Tips for Success

- **Read Error Messages**: Go's error messages are helpful
- **Use the Standard Library**: Avoid external packages unless specified
- **Test as You Go**: Run your code frequently to catch issues early
- **Read the Docs**: When stuck, check the Go documentation for the relevant package

## 🤝 Contributing

Found a bug? Want to add an exercise? Contributions are welcome!

---

Ready to start? Jump into [Exercise 1: URL Downloader](./01-url-downloader/) and begin your Go practice journey!

---
*Index generated on 2025-10-20 • 10 exercises*