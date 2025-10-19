package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type LogEntry struct {
	IP        string
	Timestamp time.Time
	Method    string
	URL       string
	Protocol  string
	Status    int
	Size      int64
	UserAgent string
	Referer   string
}

type Stats struct {
	TotalRequests    int
	TotalBytes       int64
	StatusCodes      map[int]int
	TopIPs           map[string]int
	TopPages         map[string]int
	TopUserAgents    map[string]int
	RequestsPerHour  map[string]int
	RequestsPerDay   map[string]int
	ErrorEntries     []LogEntry
	InvalidLines     int
	ParseErrors      int
}

type OutputFormat string

const (
	TextFormat OutputFormat = "text"
	JSONFormat OutputFormat = "json"
	CSVFormat  OutputFormat = "csv"
)

func main() {
	var (
		file     = flag.String("f", "", "Log file to analyze (required)")
		pattern  = flag.String("p", "", "Custom regex pattern for parsing")
		start    = flag.String("s", "", "Start time (RFC3339 format)")
		end      = flag.String("e", "", "End time (RFC3339 format)")
		output   = flag.String("o", "text", "Output format (text, json, csv)")
		top      = flag.Int("t", 10, "Number of top results to show")
		verbose  = flag.Bool("v", false, "Verbose output")
		help     = flag.Bool("h", false, "Show help")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Analyze web server log files and generate statistics.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -f access.log\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -f access.log -t 20 -o json\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -f access.log -s 2023-10-01T00:00:00Z -e 2023-10-02T00:00:00Z\n", os.Args[0])
	}

	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	if *file == "" {
		fmt.Fprintf(os.Stderr, "Error: Log file is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Parse time filters
	var startTime, endTime time.Time
	var err error

	if *start != "" {
		startTime, err = time.Parse(time.RFC3339, *start)
		if err != nil {
			log.Fatalf("Invalid start time format: %v", err)
		}
	}

	if *end != "" {
		endTime, err = time.Parse(time.RFC3339, *end)
		if err != nil {
			log.Fatalf("Invalid end time format: %v", err)
		}
	}

	// Parse output format
	format := OutputFormat(*output)
	switch format {
	case TextFormat, JSONFormat, CSVFormat:
		// Valid format
	default:
		log.Fatalf("Invalid output format: %s (use text, json, or csv)", *output)
	}

	// Analyze log file
	analyzer := &LogAnalyzer{
		FilePath:     *file,
		CustomPattern: *pattern,
		StartTime:    startTime,
		EndTime:      endTime,
		TopCount:     *top,
		Verbose:      *verbose,
	}

	stats, err := analyzer.Analyze()
	if err != nil {
		log.Fatalf("Analysis failed: %v", err)
	}

	// Output results
	if err := outputResults(stats, format, *top); err != nil {
		log.Fatalf("Failed to output results: %v", err)
	}
}

type LogAnalyzer struct {
	FilePath      string
	CustomPattern string
	StartTime     time.Time
	EndTime       time.Time
	TopCount      int
	Verbose       bool
}

func (la *LogAnalyzer) Analyze() (*Stats, error) {
	file, err := os.Open(la.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	stats := &Stats{
		StatusCodes:     make(map[int]int),
		TopIPs:          make(map[string]int),
		TopPages:        make(map[string]int),
		TopUserAgents:   make(map[string]int),
		RequestsPerHour: make(map[string]int),
		RequestsPerDay:  make(map[string]int),
	}

	// Use custom pattern if provided, otherwise default to common log format
	var logPattern *regexp.Regexp
	if la.CustomPattern != "" {
		logPattern = regexp.MustCompile(la.CustomPattern)
		if la.Verbose {
			fmt.Printf("Using custom regex pattern: %s\n", la.CustomPattern)
		}
	} else {
		// Common Log Format + Extended Log Format
		logPattern = regexp.MustCompile(`^(\S+) \S+ \S+ \[([\w:/]+\s[+\-]\d{4})\] "(\S+) (\S+) (\S+)" (\d{3}) (\d+|-)(?: "([^"]*)" "([^"]*)")?`)
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		entry, err := la.parseLogLine(line, logPattern, lineNum)
		if err != nil {
			if la.Verbose {
				log.Printf("Line %d: %v", lineNum, err)
			}
			stats.ParseErrors++
			continue
		}

		// Apply time filter
		if !la.isWithinTimeRange(entry.Timestamp) {
			continue
		}

		la.updateStats(entry, stats)
		stats.TotalRequests++
		stats.TotalBytes += entry.Size

		// Collect error entries
		if entry.Status >= 400 {
			stats.ErrorEntries = append(stats.ErrorEntries, entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	if la.Verbose {
		fmt.Printf("Processed %d lines, %d valid entries, %d errors\n",
			lineNum, stats.TotalRequests, stats.ParseErrors)
	}

	return stats, nil
}

func (la *LogAnalyzer) parseLogLine(line string, pattern *regexp.Regexp, lineNum int) (*LogEntry, error) {
	matches := pattern.FindStringSubmatch(line)
	if len(matches) < 9 {
		return nil, fmt.Errorf("line doesn't match expected format")
	}

	// Parse IP
	ip := matches[1]

	// Parse timestamp
	timestampStr := matches[2]
	timestamp, err := time.Parse("02/Jan/2006:15:04:05 -0700", timestampStr)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp format: %w", err)
	}

	// Parse request components
	method := matches[3]
	url := matches[4]
	protocol := matches[5]

	// Parse status code
	status, err := strconv.Atoi(matches[6])
	if err != nil {
		return nil, fmt.Errorf("invalid status code: %w", err)
	}

	// Parse size
	var size int64
	if matches[7] != "-" {
		size, err = strconv.ParseInt(matches[7], 10, 64)
		if err != nil {
			size = 0 // Ignore size parsing errors
		}
	}

	// Parse referer and user agent (optional fields)
	var referer, userAgent string
	if len(matches) > 8 {
		referer = matches[8]
	}
	if len(matches) > 9 {
		userAgent = matches[9]
	}

	return &LogEntry{
		IP:        ip,
		Timestamp: timestamp,
		Method:    method,
		URL:       url,
		Protocol:  protocol,
		Status:    status,
		Size:      size,
		Referer:   referer,
		UserAgent: userAgent,
	}, nil
}

func (la *LogAnalyzer) isWithinTimeRange(timestamp time.Time) bool {
	if !la.StartTime.IsZero() && timestamp.Before(la.StartTime) {
		return false
	}
	if !la.EndTime.IsZero() && timestamp.After(la.EndTime) {
		return false
	}
	return true
}

func (la *LogAnalyzer) updateStats(entry *LogEntry, stats *Stats) {
	// Status codes
	stats.StatusCodes[entry.Status]++

	// Top IPs
	stats.TopIPs[entry.IP]++

	// Top pages (ignore query parameters for grouping)
	url := entry.URL
	if idx := strings.Index(url, "?"); idx > 0 {
		url = url[:idx]
	}
	stats.TopPages[url]++

	// User agents
	if entry.UserAgent != "" {
		stats.TopUserAgents[entry.UserAgent]++
	}

	// Requests per hour
	hourKey := entry.Timestamp.Format("2006-01-02 15:00")
	stats.RequestsPerHour[hourKey]++

	// Requests per day
	dayKey := entry.Timestamp.Format("2006-01-02")
	stats.RequestsPerDay[dayKey]++
}

func outputResults(stats *Stats, format OutputFormat, topCount int) error {
	switch format {
	case TextFormat:
		return outputTextResults(stats, topCount)
	case JSONFormat:
		return outputJSONResults(stats)
	case CSVFormat:
		return outputCSVResults(stats)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func outputTextResults(stats *Stats, topCount int) error {
	fmt.Printf("Log Analysis Results\n")
	fmt.Printf("===================\n\n")

	fmt.Printf("Summary:\n")
	fmt.Printf("  Total Requests: %d\n", stats.TotalRequests)
	fmt.Printf("  Total Bytes: %s\n", formatBytes(stats.TotalBytes))
	fmt.Printf("  Parse Errors: %d\n", stats.ParseErrors)
	fmt.Printf("\n")

	// Status code distribution
	fmt.Printf("Status Code Distribution:\n")
	printTopMap(stats.StatusCodes, topCount, "Status", "Count")
	fmt.Printf("\n")

	// Top IPs
	fmt.Printf("Top IP Addresses:\n")
	printTopMap(stats.TopIPs, topCount, "IP", "Requests")
	fmt.Printf("\n")

	// Top pages
	fmt.Printf("Top Pages:\n")
	printTopMap(stats.TopPages, topCount, "Page", "Requests")
	fmt.Printf("\n")

	// Top user agents
	fmt.Printf("Top User Agents:\n")
	printTopMapString(stats.TopUserAgents, topCount, "User Agent", "Requests")
	fmt.Printf("\n")

	// Hourly requests (last 24 hours)
	fmt.Printf("Requests per Hour (last 24 hours):\n")
	hours := getSortedKeys(stats.RequestsPerHour)
	start := len(hours) - 24
	if start < 0 {
		start = 0
	}
	for i := start; i < len(hours); i++ {
		hour := hours[i]
		fmt.Printf("  %s: %d\n", hour, stats.RequestsPerHour[hour])
	}
	fmt.Printf("\n")

	// Error analysis
	if len(stats.ErrorEntries) > 0 {
		fmt.Printf("Error Analysis:\n")
		fmt.Printf("  Total Errors: %d\n", len(stats.ErrorEntries))

		// Count errors by status code
		errorStats := make(map[int]int)
		for _, entry := range stats.ErrorEntries {
			errorStats[entry.Status]++
		}

		fmt.Printf("  Error Breakdown:\n")
		for status, count := range errorStats {
			fmt.Printf("    %d: %d\n", status, count)
		}
		fmt.Printf("\n")
	}

	return nil
}

func outputJSONResults(stats *Stats) error {
	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

func outputCSVResults(stats *Stats) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write header
	header := []string{"Metric", "Value"}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write summary statistics
	records := [][]string{
		{"Total Requests", strconv.Itoa(stats.TotalRequests)},
		{"Total Bytes", strconv.FormatInt(stats.TotalBytes, 10)},
		{"Parse Errors", strconv.Itoa(stats.ParseErrors)},
	}

	for _, record := range records {
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

func printTopMap(m map[string]int, top int, keyLabel, valueLabel string) {
	sorted := getSortedMapByValue(m, top)
	for _, item := range sorted {
		fmt.Printf("  %s: %d %s\n", item.Key, item.Value, valueLabel)
	}
}

func printTopMapString(m map[string]int, top int, keyLabel, valueLabel string) {
	sorted := getSortedMapByValue(m, top)
	for _, item := range sorted {
		// Truncate long strings
		key := item.Key
		if len(key) > 80 {
			key = key[:77] + "..."
		}
		fmt.Printf("  %s: %d %s\n", key, item.Value, valueLabel)
	}
}

type MapItem struct {
	Key   string
	Value int
}

func getSortedMapByValue(m map[string]int, top int) []MapItem {
	var items []MapItem
	for k, v := range m {
		items = append(items, MapItem{k, v})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Value > items[j].Value
	})

	if len(items) > top {
		items = items[:top]
	}

	return items
}

func getSortedKeys(m map[string]int) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	return keys
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