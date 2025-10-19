# Exercise 1: URL Downloader CLI Tool

## 🎯 Objective
Create a command-line tool that downloads files from URLs with progress indicators and error handling.

## 📋 Main Focus Areas
- **Command-line arguments** (`flag` package)
- **HTTP requests** (`net/http` package)
- **File operations** (`os` and `io` packages)
- **Progress tracking** (progress bars)
- **Error handling**

## 🔧 What You'll Build
A CLI tool that:
- Downloads files from HTTP/HTTPS URLs
- Shows download progress
- Handles different output locations
- Provides useful error messages
- Supports optional flags for different behaviors

## 📝 Instructions

### Step 1: Basic Structure
Create a `main.go` file that:
1. Accepts a URL as a command-line argument
2. Makes an HTTP GET request to that URL
3. Saves the response body to a file
4. Handles basic errors

### Step 2: Command-line Flags
Add support for these flags:
- `-o, --output`: Output filename (default: derived from URL)
- `-q, --quiet`: Suppress progress output
- `-t, --timeout`: Request timeout in seconds (default: 30)

### Step 3: Progress Indicator
Implement a progress bar that shows:
- Download progress (percentage)
- Download speed
- Estimated time remaining
- Total bytes downloaded

### Step 4: Enhanced Features
Add these features:
- Resume interrupted downloads (if server supports Range requests)
- Check file existence before downloading
- Verify download integrity with checksum if available
- Support for custom User-Agent header

## 🚀 Getting Started

```bash
# Navigate to exercise directory
cd 01-url-downloader

# Create your main.go file
touch main.go

# Run your solution
go run main.go https://example.com/file.zip
```

## 💡 Implementation Tips

### HTTP Request
```go
resp, err := http.Get(url)
if err != nil {
    log.Fatalf("Failed to download: %v", err)
}
defer resp.Body.Close()
```

### File Creation
```go
file, err := os.Create(filename)
if err != nil {
    log.Fatalf("Failed to create file: %v", err)
}
defer file.Close()
```

### Progress Tracking
Track bytes copied and calculate progress percentage:
```go
copied, err := io.Copy(file, resp.Body)
progress := float64(copied) / float64(resp.ContentLength) * 100
```

## 🧪 Test Cases
Try downloading different types of files:
- Small text files
- Large binary files
- Files from different domains
- Files that don't exist (error handling)

## 📚 Go Concepts Covered
- `flag` package for command-line parsing
- `net/http` for HTTP requests
- `os` package for file operations
- `io` package for I/O operations
- Error handling patterns
- Goroutines (for concurrent progress updates)
- Channel communication

## 🔗 Useful Resources
- [Go flag package documentation](https://pkg.go.dev/flag)
- [Go net/http package](https://pkg.go.dev/net/http)
- [Go io package](https://pkg.go.dev/io)

## ✅ Success Criteria
Your tool should be able to:
- [ ] Download files from valid URLs
- [ ] Show meaningful progress information
- [ ] Handle network errors gracefully
- [ ] Support custom output filenames
- [ ] Work with both HTTP and HTTPS URLs
- [ ] Provide helpful error messages

When you're ready, check out the [solution](./solution/main.go) to see a complete implementation.