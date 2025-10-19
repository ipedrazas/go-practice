package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type OrganizeMethod string

const (
	ByType  OrganizeMethod = "type"
	BySize  OrganizeMethod = "size"
	ByDate  OrganizeMethod = "date"
)

type FileInfo struct {
	Path     string
	Info     os.FileInfo
	Category string
}

type Organizer struct {
	Directory string
	Method    OrganizeMethod
	Recursive bool
	DryRun    bool
	Force     bool
	Verbose   bool
	Stats     map[string]int
}

func main() {
	var (
		directory = flag.String("d", ".", "Directory to organize")
		method    = flag.String("b", "type", "Organization method (type, size, date)")
		recursive = flag.Bool("r", false, "Process subdirectories recursively")
		dryRun    = flag.Bool("n", false, "Dry run - show what would be done")
		force     = flag.Bool("f", false, "Force overwrite existing files")
		verbose   = flag.Bool("v", false, "Verbose output")
		help      = flag.Bool("h", false, "Show help")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Organize files in a directory by type, size, or date.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nOrganization Methods:\n")
		fmt.Fprintf(os.Stderr, "  type  - Group files by extension (Images, Documents, etc.)\n")
		fmt.Fprintf(os.Stderr, "  size  - Group by file size (Small, Medium, Large)\n")
		fmt.Fprintf(os.Stderr, "  date  - Group by modification date (Today, This Week, etc.)\n")
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -d Downloads --dry-run\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -b size -r ~/Desktop\n", os.Args[0])
	}

	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	organizer := &Organizer{
		Directory: *directory,
		Method:    OrganizeMethod(*method),
		Recursive: *recursive,
		DryRun:    *dryRun,
		Force:     *force,
		Verbose:   *verbose,
		Stats:     make(map[string]int),
	}

	if err := organizer.Organize(); err != nil {
		log.Fatalf("Organization failed: %v", err)
	}

	organizer.PrintSummary()
}

func (o *Organizer) Organize() error {
	// Check if directory exists
	info, err := os.Stat(o.Directory)
	if err != nil {
		return fmt.Errorf("cannot access directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", o.Directory)
	}

	if o.DryRun {
		fmt.Printf("DRY RUN: No files will be moved\n\n")
	}

	fmt.Printf("Organizing files in: %s\n", o.Directory)
	fmt.Printf("Method: %s\n", o.Method)
	if o.Recursive {
		fmt.Printf("Mode: Recursive\n")
	}
	fmt.Println()

	// Scan directory
	files, err := o.scanDirectory()
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	if len(files) == 0 {
		fmt.Println("No files found to organize.")
		return nil
	}

	fmt.Printf("Found %d files to organize\n\n", len(files))

	// Organize files
	for _, file := range files {
		if err := o.organizeFile(file); err != nil {
			log.Printf("Failed to organize %s: %v", file.Path, err)
		}
	}

	return nil
}

func (o *Organizer) scanDirectory() ([]FileInfo, error) {
	var files []FileInfo

	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip hidden files
		if strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}

		// Skip the organizer executable itself
		if filepath.Base(path) == filepath.Base(os.Args[0]) {
			return nil
		}

		// Get file category based on organization method
		category := o.getFileCategory(path, info)

		files = append(files, FileInfo{
			Path:     path,
			Info:     info,
			Category: category,
		})

		return nil
	}

	if o.Recursive {
		err = filepath.Walk(o.Directory, walkFunc)
	} else {
		entries, err := os.ReadDir(o.Directory)
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			path := filepath.Join(o.Directory, entry.Name())
			info, err := entry.Info()
			if err != nil {
				continue
			}

			if err := walkFunc(path, info, nil); err != nil {
				continue
			}
		}
	}

	return files, err
}

func (o *Organizer) getFileCategory(path string, info os.FileInfo) string {
	switch o.Method {
	case ByType:
		return o.getTypeCategory(path)
	case BySize:
		return o.getSizeCategory(info.Size())
	case ByDate:
		return o.getDateCategory(info.ModTime())
	default:
		return "Other"
	}
}

func (o *Organizer) getTypeCategory(path string) string {
	ext := strings.ToLower(filepath.Ext(path))

	// Image files
	imageExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
		".bmp": true, ".tiff": true, ".webp": true, ".svg": true,
	}
	if imageExts[ext] {
		return "Images"
	}

	// Document files
	docExts := map[string]bool{
		".pdf": true, ".doc": true, ".docx": true, ".txt": true,
		".rtf": true, ".odt": true, ".pages": true,
	}
	if docExts[ext] {
		return "Documents"
	}

	// Video files
	videoExts := map[string]bool{
		".mp4": true, ".avi": true, ".mkv": true, ".mov": true,
		".wmv": true, ".flv": true, ".webm": true, ".m4v": true,
	}
	if videoExts[ext] {
		return "Videos"
	}

	// Audio files
	audioExts := map[string]bool{
		".mp3": true, ".wav": true, ".flac": true, ".aac": true,
		".ogg": true, ".wma": true, ".m4a": true,
	}
	if audioExts[ext] {
		return "Audio"
	}

	// Archive files
	archiveExts := map[string]bool{
		".zip": true, ".rar": true, ".7z": true, ".tar": true,
		".gz": true, ".bz2": true, ".xz": true,
	}
	if archiveExts[ext] {
		return "Archives"
	}

	// Code files
	codeExts := map[string]bool{
		".go": true, ".js": true, ".py": true, ".java": true,
		".cpp": true, ".c": true, ".h": true, ".html": true,
		".css": true, ".php": true, ".rb": true, ".swift": true,
		".rs": true, ".ts": true, ".jsx": true, ".tsx": true,
	}
	if codeExts[ext] {
		return "Code"
	}

	return "Other"
}

func (o *Organizer) getSizeCategory(size int64) string {
	const (
		MB = 1024 * 1024
		KB = 1024
	)

	switch {
	case size < 100*KB:
		return "Small"
	case size < 10*MB:
		return "Medium"
	default:
		return "Large"
	}
}

func (o *Organizer) getDateCategory(modTime time.Time) string {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	thisWeek := today.AddDate(0, 0, -7)
	thisMonth := today.AddDate(0, -1, 0)

	switch {
	case modTime.After(today):
		return "Today"
	case modTime.After(thisWeek):
		return "This Week"
	case modTime.After(thisMonth):
		return "This Month"
	default:
		return "Older"
	}
}

func (o *Organizer) organizeFile(file FileInfo) error {
	// Create target directory
	targetDir := filepath.Join(o.Directory, file.Category)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
	}

	// Determine target file path
	targetPath := filepath.Join(targetDir, filepath.Base(file.Path))

	// Handle file name conflicts
	if _, err := os.Stat(targetPath); err == nil {
		if !o.Force {
			targetPath = o.getUniqueFilename(targetPath)
		}
	}

	// Show action
	relPath, _ := filepath.Rel(o.Directory, file.Path)
	targetRelPath, _ := filepath.Rel(o.Directory, targetPath)

	if o.DryRun {
		fmt.Printf("Would move: %s -> %s\n", relPath, targetRelPath)
	} else {
		if o.Verbose {
			fmt.Printf("Moving: %s -> %s\n", relPath, targetRelPath)
		}
		if err := os.Rename(file.Path, targetPath); err != nil {
			return fmt.Errorf("failed to move file: %w", err)
		}
	}

	// Update statistics
	o.Stats[file.Category]++

	return nil
}

func (o *Organizer) getUniqueFilename(path string) string {
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)
	counter := 1

	for {
		newPath := fmt.Sprintf("%s (%d)%s", base, counter, ext)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
		counter++
	}
}

func (o *Organizer) PrintSummary() {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("Organization Summary")
	fmt.Println(strings.Repeat("=", 50))

	if len(o.Stats) == 0 {
		fmt.Println("No files were organized.")
		return
	}

	// Sort categories for consistent output
	var categories []string
	for category := range o.Stats {
		categories = append(categories, category)
	}
	sort.Strings(categories)

	total := 0
	for _, category := range categories {
		count := o.Stats[category]
		total += count
		fmt.Printf("%-12s: %d files\n", category, count)
	}

	fmt.Printf("\nTotal files organized: %d\n", total)

	if o.DryRun {
		fmt.Println("\nThis was a dry run. No files were actually moved.")
		fmt.Println("Run without --dry-run to organize the files.")
	}
}