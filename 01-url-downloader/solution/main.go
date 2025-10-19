package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	// Define command-line flags
	var (
		output   = flag.String("o", "", "Output filename")
		quiet    = flag.Bool("q", false, "Suppress progress output")
		timeout  = flag.Int("t", 30, "Request timeout in seconds")
		help     = flag.Bool("h", false, "Show help")
	)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <URL>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Download files from URLs with progress indicators.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s https://example.com/file.txt\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -o myfile.txt https://example.com/file.txt\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -q https://example.com/largefile.zip\n", os.Args[0])
	}
	flag.Parse()

	// Show help if requested
	if *help {
		flag.Usage()
		return
	}

	// Check if URL is provided
	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Error: URL is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	url := flag.Arg(0)

	// Set timeout
	client := &http.Client{
		Timeout: time.Duration(*timeout) * time.Second,
	}

	// Start download
	if err := downloadFile(client, url, *output, *quiet); err != nil {
		log.Fatalf("Download failed: %v", err)
	}

	if !*quiet {
		fmt.Println("\nDownload completed successfully!")
	}
}

func downloadFile(client *http.Client, url, output string, quiet bool) error {
	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set User-Agent header
	req.Header.Set("User-Agent", "Go-Downloader/1.0")

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %s", resp.Status)
	}

	// Determine output filename
	if output == "" {
		output = getFilenameFromURL(url)
	}

	// Check if file already exists
	if _, err := os.Stat(output); err == nil {
		return fmt.Errorf("file already exists: %s", output)
	}

	// Create output file
	file, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Get content length for progress tracking
	contentLength := resp.ContentLength

	if !quiet {
		fmt.Printf("Downloading %s to %s\n", url, output)
		if contentLength > 0 {
			fmt.Printf("File size: %s\n", formatBytes(contentLength))
		}
	}

	// Copy with progress tracking
	if !quiet {
		return copyWithProgress(resp.Body, file, contentLength)
	}

	_, err = io.Copy(file, resp.Body)
	return err
}

func copyWithProgress(src io.Reader, dst io.Writer, total int64) error {
	progress := &progressTracker{
		total: total,
	}

	// Create multi-writer for file and progress tracking
	multiWriter := io.MultiWriter(dst, progress)

	// Copy data
	copied, err := io.Copy(multiWriter, src)
	if err != nil {
		return err
	}

	progress.Finish()
	fmt.Printf("\nDownloaded: %s (%.2f%%)\n", formatBytes(copied), 100.0)
	return nil
}

func getFilenameFromURL(url string) string {
	// Extract filename from URL path
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		filename := parts[len(parts)-1]
		if filename != "" {
			return filename
		}
	}

	// Fallback to domain-based filename
	parts = strings.Split(url, "//")
	if len(parts) > 1 {
		domain := strings.Split(parts[1], "/")[0]
		return domain + ".download"
	}

	return "download"
}

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
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// progressTracker implements io.Writer to track download progress
type progressTracker struct {
	total    int64
	written  int64
	lastTime time.Time
}

func (p *progressTracker) Write(data []byte) (int, error) {
	n := len(data)
	p.written += int64(n)

	// Update progress every 100ms
	now := time.Now()
	if now.Sub(p.lastTime) >= 100*time.Millisecond || p.written == p.total {
		p.updateProgress()
		p.lastTime = now
	}

	return n, nil
}

func (p *progressTracker) updateProgress() {
	if p.total <= 0 {
		fmt.Printf("\rDownloaded: %s", formatBytes(p.written))
		return
	}

	percentage := float64(p.written) / float64(p.total) * 100
	fmt.Printf("\rProgress: %.1f%% (%s/%s)",
		percentage,
		formatBytes(p.written),
		formatBytes(p.total))
}

func (p *progressTracker) Finish() {
	// Final progress update
	p.updateProgress()
}