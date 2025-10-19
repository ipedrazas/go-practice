package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type DirectoryInfo struct {
	Path           string           `json:"path"`
	Size           int64            `json:"size_bytes"`
	FormattedSize  string           `json:"size_formatted"`
	FileCount      int64            `json:"file_count"`
	DirCount       int64            `json:"dir_count"`
	LargestFiles   []FileInfo       `json:"largest_files,omitempty"`
	FileTypes      map[string]int64 `json:"file_types,omitempty"`
	Subdirectories []*DirectoryInfo `json:"subdirectories,omitempty"`
	LastModified   time.Time        `json:"last_modified"`
}

type FileInfo struct {
	Path          string    `json:"path"`
	Size          int64     `json:"size"`
	FormattedSize string    `json:"size_formatted"`
	LastModified  time.Time `json:"last_modified"`
}

type ScanOptions struct {
	Directory      string
	SortBy         string
	Limit          int
	HumanReadable  bool
	ShowFiles      bool
	MaxDepth       int
	ExcludePattern []string
	Verbose        bool
}

func main() {
	var (
		directory = flag.String("d", ".", "Directory to analyze")
		sortBy    = flag.String("s", "size", "Sort by (size, name, files, modified)")
		limit     = flag.Int("l", 20, "Number of results to show")
		human     = flag.Bool("h", true, "Human-readable output")
		files     = flag.Bool("f", false, "Show individual files")
		depth     = flag.Int("depth", -1, "Maximum depth to scan (-1 for unlimited)")
		exclude   = flag.String("x", "", "Comma-separated patterns to exclude")
		verbose   = flag.Bool("v", false, "Verbose output")
		help      = flag.Bool("help", false, "Show help")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Analyze disk usage by directory and file.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -d /home/user\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -l 10 -s size\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -f -depth 2\n", os.Args[0])
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
		for i, pattern := range excludePatterns {
			excludePatterns[i] = strings.TrimSpace(pattern)
		}
	}

	options := ScanOptions{
		Directory:      *directory,
		SortBy:         *sortBy,
		Limit:          *limit,
		HumanReadable:  *human,
		ShowFiles:      *files,
		MaxDepth:       *depth,
		ExcludePattern: excludePatterns,
		Verbose:        *verbose,
	}

	analyzer := &DirectoryAnalyzer{Options: options}
	if err := analyzer.Analyze(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

type DirectoryAnalyzer struct {
	Options ScanOptions
}

func (da *DirectoryAnalyzer) Analyze() error {
	// Check if directory exists
	info, err := os.Stat(da.Options.Directory)
	if err != nil {
		return fmt.Errorf("cannot access directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", da.Options.Directory)
	}

	fmt.Printf("Analyzing: %s\n", da.Options.Directory)
	startTime := time.Now()

	// Scan directory
	rootInfo, err := da.scanDirectory(da.Options.Directory, 0)
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	duration := time.Since(startTime)
	fmt.Printf("Scan completed in %v\n\n", duration)

	// Display results
	da.displayResults(rootInfo)

	return nil
}

func (da *DirectoryAnalyzer) scanDirectory(path string, depth int) (*DirectoryInfo, error) {
	if da.Options.MaxDepth >= 0 && depth > da.Options.MaxDepth {
		return nil, nil
	}

	dirInfo := &DirectoryInfo{
		Path:      path,
		FileTypes: make(map[string]int64),
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		if da.Options.Verbose {
			fmt.Printf("Warning: Cannot read directory %s: %v\n", path, err)
		}
		return nil, err
	}

	for _, entry := range entries {
		// Skip excluded patterns
		if da.shouldExclude(entry.Name()) {
			continue
		}

		fullPath := filepath.Join(path, entry.Name())

		if entry.IsDir() {
			// Recursively scan subdirectory
			subInfo, err := da.scanDirectory(fullPath, depth+1)
			if err != nil {
				continue // Skip directories we can't read
			}
			if subInfo != nil {
				dirInfo.Subdirectories = append(dirInfo.Subdirectories, subInfo)
				dirInfo.Size += subInfo.Size
				dirInfo.FileCount += subInfo.FileCount
				dirInfo.DirCount += subInfo.DirCount + 1
			}
		} else {
			// Process file
			fileInfo, err := entry.Info()
			if err != nil {
				continue
			}

			fileSize := fileInfo.Size()
			dirInfo.Size += fileSize
			dirInfo.FileCount++

			// Track file types
			ext := strings.ToLower(filepath.Ext(fileInfo.Name()))
			if ext == "" {
				ext = "no extension"
			}
			dirInfo.FileTypes[ext]++

			// Track largest files
			if da.Options.ShowFiles {
				fileInfo := FileInfo{
					Path:          fullPath,
					Size:          fileSize,
					FormattedSize: formatBytes(fileSize, da.Options.HumanReadable),
					LastModified:  fileInfo.ModTime(),
				}
				dirInfo.LargestFiles = append(dirInfo.LargestFiles, fileInfo)
			}

			// Update last modified time
			if fileInfo.ModTime().After(dirInfo.LastModified) {
				dirInfo.LastModified = fileInfo.ModTime()
			}
		}
	}

	// Sort largest files
	if len(dirInfo.LargestFiles) > 0 {
		sort.Slice(dirInfo.LargestFiles, func(i, j int) bool {
			return dirInfo.LargestFiles[i].Size > dirInfo.LargestFiles[j].Size
		})

		// Keep only top files
		if len(dirInfo.LargestFiles) > 10 {
			dirInfo.LargestFiles = dirInfo.LargestFiles[:10]
		}
	}

	// Format size
	dirInfo.FormattedSize = formatBytes(dirInfo.Size, da.Options.HumanReadable)

	return dirInfo, nil
}

func (da *DirectoryAnalyzer) shouldExclude(name string) bool {
	for _, pattern := range da.Options.ExcludePattern {
		if strings.Contains(name, pattern) {
			return true
		}
	}

	// Skip hidden files and directories by default
	if strings.HasPrefix(name, ".") {
		return true
	}

	return false
}

func (da *DirectoryAnalyzer) displayResults(rootInfo *DirectoryInfo) {
	// Show summary
	fmt.Printf("Summary for: %s\n", rootInfo.Path)
	fmt.Printf("Total Size: %s\n", rootInfo.FormattedSize)
	fmt.Printf("Files: %d\n", rootInfo.FileCount)
	fmt.Printf("Directories: %d\n", rootInfo.DirCount)
	fmt.Printf("Last Modified: %s\n\n", rootInfo.LastModified.Format("2006-01-02 15:04:05"))

	// Show file type distribution
	if len(rootInfo.FileTypes) > 0 {
		fmt.Printf("File Types:\n")
		da.displayFileTypes(rootInfo.FileTypes)
		fmt.Printf("\n")
	}

	// Show largest files
	if len(rootInfo.LargestFiles) > 0 {
		fmt.Printf("Largest Files:\n")
		da.displayLargestFiles(rootInfo.LargestFiles)
		fmt.Printf("\n")
	}

	// Show subdirectories
	if len(rootInfo.Subdirectories) > 0 {
		da.displaySubdirectories(rootInfo.Subdirectories)
	}
}

func (da *DirectoryAnalyzer) displayFileTypes(fileTypes map[string]int64) {
	// Convert to slice for sorting
	type typeInfo struct {
		ext   string
		count int64
	}

	var types []typeInfo
	for ext, count := range fileTypes {
		types = append(types, typeInfo{ext, count})
	}

	// Sort by count
	sort.Slice(types, func(i, j int) bool {
		return types[i].count > types[j].count
	})

	// Display top 10
	limit := len(types)
	if limit > 10 {
		limit = 10
	}

	for i := 0; i < limit; i++ {
		fmt.Printf("  %-15s: %d files\n", types[i].ext, types[i].count)
	}
}

func (da *DirectoryAnalyzer) displayLargestFiles(files []FileInfo) {
	limit := len(files)
	if limit > da.Options.Limit {
		limit = da.Options.Limit
	}

	for i := 0; i < limit; i++ {
		file := files[i]
		fmt.Printf("  %s %s\n", file.FormattedSize, file.Path)
	}
}

func (da *DirectoryAnalyzer) displaySubdirectories(directories []*DirectoryInfo) {
	// Sort directories based on options
	switch da.Options.SortBy {
	case "size":
		sort.Slice(directories, func(i, j int) bool {
			return directories[i].Size > directories[j].Size
		})
	case "name":
		sort.Slice(directories, func(i, j int) bool {
			return directories[i].Path < directories[j].Path
		})
	case "files":
		sort.Slice(directories, func(i, j int) bool {
			return directories[i].FileCount > directories[j].FileCount
		})
	case "modified":
		sort.Slice(directories, func(i, j int) bool {
			return directories[i].LastModified.After(directories[j].LastModified)
		})
	}

	fmt.Printf("Subdirectories (sorted by %s):\n", da.Options.SortBy)
	fmt.Printf("%-12s %-8s %-20s %s\n", "Size", "Files", "Modified", "Path")
	fmt.Printf("%-12s %-8s %-20s %s\n", "----", "-----", "--------", "----")

	limit := len(directories)
	if limit > da.Options.Limit {
		limit = da.Options.Limit
	}

	for i := 0; i < limit; i++ {
		dir := directories[i]
		modified := dir.LastModified.Format("2006-01-02 15:04")
		if dir.LastModified.IsZero() {
			modified = "unknown"
		}

		relPath, _ := filepath.Rel(da.Options.Directory, dir.Path)
		fmt.Printf("%-12s %-8d %-20s %s\n",
			dir.FormattedSize,
			dir.FileCount,
			modified,
			relPath)
	}

	if len(directories) > da.Options.Limit {
		fmt.Printf("\n... and %d more directories\n", len(directories)-da.Options.Limit)
	}
}

func formatBytes(bytes int64, humanReadable bool) string {
	if !humanReadable {
		return fmt.Sprintf("%d", bytes)
	}

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