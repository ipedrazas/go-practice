# Exercise 2: File Organizer Utility

## 🎯 Objective
Create a CLI tool that organizes files in a directory based on their type, size, or custom rules.

## 📋 Main Focus Areas
- **File system operations** (`os` and `filepath` packages)
- **Directory traversal** (`filepath.Walk`)
- **File metadata** (file info, permissions, timestamps)
- **Pattern matching** (`path/filepath` glob patterns)
- **String operations** (file extension detection)

## 🔧 What You'll Build
A file organizer that:
- Sorts files into subdirectories by type (images, documents, etc.)
- Can organize by file size or date modified
- Supports custom organization rules
- Provides dry-run mode to preview changes
- Handles file name conflicts
- Generates summary reports

## 📝 Instructions

### Step 1: Basic File Type Organization
Create a tool that:
1. Scans a directory for files
2. Categorizes files by extension
3. Moves files to appropriate subdirectories
4. Common categories: Images, Documents, Videos, Audio, Archives, Code

### Step 2: Command-line Interface
Add support for these flags:
- `-d, --directory`: Directory to organize (default: current directory)
- `-r, --recursive`: Process subdirectories
- `-n, --dry-run`: Show what would be done without making changes
- `-b, --by`: Organization method (type, size, date)
- `-f, --force`: Overwrite existing files

### Step 3: Organization Methods
Implement different organization strategies:
- **By Type**: Group files by extension
- **By Size**: Small (<1MB), Medium (1MB-10MB), Large (>10MB)
- **By Date**: Today, This Week, This Month, Older

### Step 4: Advanced Features
Add these capabilities:
- Custom mapping rules (config file)
- Duplicate detection
- File name sanitization
- Progress reporting for large directories
- Undo functionality (move log)

## 🚀 Getting Started

```bash
# Navigate to exercise directory
cd 02-file-organizer

# Create your main.go file
touch main.go

# Test with a sample directory
mkdir -p test-files
# Add various file types to test-files/

# Run your solution
go run main.go -d test-files --dry-run
```

## 💡 Implementation Tips

### File Extension Detection
```go
func getFileCategory(filename string) string {
    ext := strings.ToLower(filepath.Ext(filename))
    switch ext {
    case ".jpg", ".png", ".gif":
        return "Images"
    case ".pdf", ".doc", ".txt":
        return "Documents"
    default:
        return "Other"
    }
}
```

### Directory Traversal
```go
filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
    if err != nil {
        return err
    }
    if !info.IsDir() {
        // Process file
    }
    return nil
})
```

### File Operations
```go
err := os.Rename(oldPath, newPath)
if err != nil {
    log.Printf("Failed to move %s: %v", path, err)
}
```

## 🧪 Test Cases
Create test scenarios:
- Mixed file types in one directory
- Nested directory structures
- Files with similar names
- Large numbers of files
- Files with special characters in names

## 📚 Go Concepts Covered
- `os` package for file system operations
- `filepath` package for path manipulation
- `filepath.Walk` for directory traversal
- File metadata and permissions
- String manipulation and pattern matching
- Error handling for file operations

## 🔗 Useful Resources
- [Go filepath package](https://pkg.go.dev/path/filepath)
- [Go os package](https://pkg.go.dev/os)
- [File I/O in Go](https://gobyexample.com/reading-files)

## ✅ Success Criteria
Your tool should be able to:
- [ ] Organize files by type into appropriate subdirectories
- [ ] Handle different organization methods (type, size, date)
- [ ] Provide dry-run functionality
- [ ] Handle file name conflicts gracefully
- [ ] Process directories recursively when requested
- [ ] Generate useful progress and summary information

## 🎁 Bonus Features
If you want an extra challenge:
- Configuration file support (JSON/YAML)
- Duplicate file detection
- File content analysis (MIME type detection)
- Integration with file system watching
- GUI interface using a web framework

When you're ready, check out the [solution](./solution/main.go) to see a complete implementation.