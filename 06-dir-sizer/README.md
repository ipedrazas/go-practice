# Exercise 6: Directory Size Analyzer

## 🎯 Objective
Create a tool that analyzes disk usage by directory, showing which directories consume the most space.

## 📋 Main Focus Areas
- **File system navigation** (`os`, `filepath` packages)
- **Recursive directory traversal**
- **Human-readable formatting**
- **Data sorting and aggregation**
- **Memory-efficient processing**

## 🔧 What You'll Build
A directory analyzer that:
- Calculates sizes of directories and subdirectories
- Shows largest files and directories
- Provides human-readable size formatting
- Supports different sorting options
- Handles symbolic links and file system errors
- Generates visual representations of disk usage

## 📝 Instructions

### Step 1: Basic Directory Scanning
Create a tool that:
1. Traverses directories recursively
2. Calculates total size for each directory
3. Handles file system errors gracefully
4. Shows basic size information

### Step 2: Command-line Interface
Add support for these flags:
- `-d, --directory`: Directory to analyze (default: current)
- `-s, --sort`: Sort by (size, name, files)
- `-l, --limit`: Number of results to show
- `-h, --human`: Human-readable output
- `-f, --files`: Show individual files
- `-x, --exclude`: Patterns to exclude

### Step 3: Analysis Features
Implement these analyses:
- Directory size breakdown
- Largest files listing
- File count by directory
- File type distribution
- Depth analysis
- Duplicate file detection

### Step 4: Advanced Features
Add these capabilities:
- Progress indicators for large scans
- Caching for repeated scans
- Output in different formats
- Visual tree representation
- Integration with system commands (du)

## 🚀 Getting Started

```bash
# Navigate to exercise directory
cd 06-dir-sizer

# Create your main.go file
touch main.go

# Analyze current directory
go run main.go

# Analyze specific directory
go run main.go -d /path/to/directory

# Show top 10 largest items
go run main.go -l 10 -s size
```

## 💡 Implementation Tips

### Directory Traversal
```go
func scanDirectory(path string) (*DirectoryInfo, error) {
    info := &DirectoryInfo{
        Path: path,
    }

    entries, err := os.ReadDir(path)
    if err != nil {
        return nil, err
    }

    for _, entry := range entries {
        if entry.IsDir() {
            subInfo, err := scanDirectory(filepath.Join(path, entry.Name()))
            if err != nil {
                continue // Skip directories we can't read
            }
            info.Subdirectories = append(info.Subdirectories, subInfo)
            info.Size += subInfo.Size
        } else {
            fileInfo, _ := entry.Info()
            info.Size += fileInfo.Size()
            info.FileCount++
        }
    }

    return info, nil
}
```

### Human-readable Sizes
```go
func formatBytes(bytes int64) string {
    const unit = 1024
    if bytes < unit {
        return fmt.Sprintf("%d B", bytes)
    }
    div, exp := int64(unit), 0
    for n := bytes / unit; n >= unit; n /= unit {
        div *= unit
        exp++
    }
    return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
```

## 🧪 Test Cases
Test with different directory structures:
- Deep nesting
- Many small files
- Few large files
- Mixed file types
- Permission issues
- Symbolic links

## 📚 Go Concepts Covered
- `filepath.Walk` for directory traversal
- File information and metadata
- Recursive algorithms
- Error handling in file operations
- Data aggregation and sorting
- Memory management for large datasets

## ✅ Success Criteria
Your tool should be able to:
- [ ] Calculate directory sizes accurately
- [ ] Handle very large directory structures
- [ ] Show results in human-readable format
- [ ] Sort results by different criteria
- [ ] Handle file system errors gracefully
- [ ] Provide useful disk usage insights

When you're ready, check out the [solution](./solution/main.go) to see a complete implementation.